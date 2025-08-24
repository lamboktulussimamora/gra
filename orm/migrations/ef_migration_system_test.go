package migrations

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// openTestSQLite opens an in-memory SQLite DB for tests.
func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestEFMigrationManager_EnsureSchema_AddUpdateRollback(t *testing.T) {
	db := openTestSQLite(t)

	em := NewEFMigrationManager(db, nil)
	// Should detect SQLite driver automatically.
	if err := em.EnsureSchema(); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	// Verify schema tables exist
	for _, tbl := range []string{em.migrationTable, em.historyTable, em.snapshotTable} {
		var cnt int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&cnt)
		if err != nil || cnt == 0 {
			t.Fatalf("expected table %s to exist, err=%v cnt=%d", tbl, err, cnt)
		}
	}

	// Add a migration and apply it
	up := "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);"
	down := "DROP TABLE IF EXISTS users;"
	mig := em.AddMigration("CreateUsers", "create users table", up, down)

	if err := em.UpdateDatabase(); err != nil {
		t.Fatalf("UpdateDatabase failed: %v", err)
	}

	// users table should exist now
	var usersCnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&usersCnt); err != nil || usersCnt == 0 {
		t.Fatalf("expected users table created, err=%v cnt=%d", err, usersCnt)
	}

	// History should contain an applied record
	hist, err := em.GetMigrationHistory()
	if err != nil {
		t.Fatalf("GetMigrationHistory failed: %v", err)
	}
	if len(hist.Applied) == 0 {
		t.Fatalf("expected at least one applied migration")
	}

	// Clear in-memory pending list since UpdateDatabase doesn't mutate it
	em.pendingMigrations = nil
	hasPending, err := em.HasPendingMigrations()
	if err != nil {
		t.Fatalf("HasPendingMigrations failed: %v", err)
	}
	if hasPending {
		t.Fatalf("did not expect pending migrations after apply")
	}

	// Rollback the migration (use internal helper to drop this migration)
	if err := em.rollbackMigration(*mig); err != nil {
		t.Fatalf("rollbackMigration failed: %v", err)
	}

	// users table should be gone
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&usersCnt); err != nil {
		t.Fatalf("query users table existence failed: %v", err)
	}
	if usersCnt != 0 {
		t.Fatalf("expected users table dropped after rollback")
	}
}

func TestEFMigrationManager_ConvertQueryPlaceholders(t *testing.T) {
	db := openTestSQLite(t)
	em := NewEFMigrationManager(db, nil)
	if err := em.EnsureSchema(); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	// Force driver to PostgreSQL to validate conversion logic
	em.driver = PostgreSQL
	in := "INSERT INTO t(a,b,c) VALUES (?,?,?) WHERE d=?"
	out := em.ConvertQueryPlaceholders(in)
	if out != "INSERT INTO t(a,b,c) VALUES ($1,$2,$3) WHERE d=$4" {
		t.Fatalf("unexpected conversion: %q", out)
	}

	// For SQLite, query should remain unchanged
	em.driver = SQLite
	if got := em.ConvertQueryPlaceholders(in); got != in {
		t.Fatalf("expected unchanged for SQLite, got %q", got)
	}
}

func TestEFMigrationManager_CreateAutoMigrations_GeneratesSQL(t *testing.T) {
	db := openTestSQLite(t)
	em := NewEFMigrationManager(db, nil)
	if err := em.EnsureSchema(); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	type Order struct{}
	type Item struct{}
	if err := em.CreateAutoMigrations([]interface{}{Order{}, &Item{}}, "Auto"); err != nil {
		t.Fatalf("CreateAutoMigrations failed: %v", err)
	}

	// Pending migrations should include one with generated SQL
	pending, err := em.GetPendingMigrations()
	if err != nil {
		t.Fatalf("GetPendingMigrations failed: %v", err)
	}
	if len(pending) == 0 {
		t.Fatalf("expected pending migration present")
	}
	if !strings.Contains(pending[0].UpSQL, "CREATE TABLE IF NOT EXISTS order") {
		t.Fatalf("expected generated up sql for orders, got: %s", pending[0].UpSQL)
	}
}

func TestEFMigrationManager_findTargetMigrationIndex(t *testing.T) {
	db := openTestSQLite(t)
	em := NewEFMigrationManager(db, nil)

	ms := []Migration{
		{ID: "1_a", Name: "a", Version: time.Now().Unix()},
		{ID: "2_b", Name: "b", Version: time.Now().Unix()},
	}
	if idx := em.findTargetMigrationIndex(ms, "2_b"); idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
	if idx := em.findTargetMigrationIndex(ms, "b"); idx != 1 {
		t.Fatalf("expected index 1 by name, got %d", idx)
	}
	if idx := em.findTargetMigrationIndex(ms, "zzz"); idx != -1 {
		t.Fatalf("expected -1 for unknown, got %d", idx)
	}
}

func TestSimpleMigrator_EndToEnd_SQLite(t *testing.T) {
	db := openTestSQLite(t)
	sm := NewSimpleMigrator(db, SQLite, t.TempDir())

	// Define a simple model and register it
	type Product struct {
		ID   int    `db:"id,primary_key"`
		Name string `db:"name" sql:"size:50"`
	}
	sm.DbSet(Product{})

	// Create initial migration file
	mig, err := sm.CreateInitialMigration("init")
	if err != nil {
		t.Fatalf("CreateInitialMigration failed: %v", err)
	}
	if mig == nil || len(mig.UpSQL) == 0 || !strings.Contains(mig.UpSQL[0], "CREATE TABLE") {
		t.Fatalf("unexpected migration content: %#v", mig)
	}

	// File should be created in migrationsDir
	if _, err := filepath.Glob(filepath.Join(sm.migrationsDir, "*_init.sql")); err != nil {
		t.Fatalf("expected migration file created: %v", err)
	}

	// Apply migration SQL
	if err := sm.ApplyMigration(mig); err != nil {
		t.Fatalf("ApplyMigration failed: %v", err)
	}

	// TableExists should return true
	exists, err := sm.TableExists("products")
	if err != nil || !exists {
		t.Fatalf("expected products table to exist, err=%v exists=%v", err, exists)
	}

	// Status should report up-to-date after table is created
	status, err := sm.GetMigrationStatus()
	if err != nil {
		t.Fatalf("GetMigrationStatus failed: %v", err)
	}
	if status.HasPendingChanges {
		t.Fatalf("did not expect pending changes after apply: %#v", status)
	}
}

// TestEFMigrationManager_recordMigrationResult ensures insert and conflict-update paths are covered.
func TestEFMigrationManager_recordMigrationResult_SQLite(t *testing.T) {
	const (
		stateFailed  = "failed"
		stateApplied = "applied"
	)
	db := openTestSQLite(t)
	em := NewEFMigrationManager(db, nil)
	if err := em.EnsureSchema(); err != nil {
		t.Fatalf("EnsureSchema failed: %v", err)
	}

	// Prepare a migration
	mig := &Migration{ID: "1_init", Name: "init", Version: time.Now().Unix(), Description: "", UpSQL: "--", DownSQL: "--"}

	// Insert as failed first
	em.recordMigrationResult(*mig, MigrationStateFailed, 123, "boom")

	var state string
	var execMs int
	var errMsg string
	// Verify inserted
	err := db.QueryRow("SELECT state, execution_time_ms, COALESCE(error_message,'') FROM "+em.historyTable+" WHERE migration_id=?", mig.ID).Scan(&state, &execMs, &errMsg)
	if err != nil {
		t.Fatalf("query history failed: %v", err)
	}
	if state != stateFailed || execMs != 123 || errMsg != "boom" {
		t.Fatalf("unexpected row: state=%s execMs=%d err=%q", state, execMs, errMsg)
	}

	// Update same migration to applied via conflict update
	em.recordMigrationResult(*mig, MigrationStateApplied, 5, "")

	err = db.QueryRow("SELECT state, execution_time_ms, COALESCE(error_message,'') FROM "+em.historyTable+" WHERE migration_id=?", mig.ID).Scan(&state, &execMs, &errMsg)
	if err != nil {
		t.Fatalf("requery history failed: %v", err)
	}
	if state != stateApplied || execMs != 5 || errMsg != "" {
		t.Fatalf("unexpected updated row: state=%s execMs=%d err=%q", state, execMs, errMsg)
	}
}

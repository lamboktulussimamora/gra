package migrations

import (
	"database/sql"
	"log"
	"testing"
)

// fakeDB wraps a real in-memory SQLite connection for manager construction without external deps
func newTestManager(t *testing.T) *EFMigrationManager {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	cfg := DefaultEFMigrationConfig()
	cfg.Logger = log.Default()
	m := NewEFMigrationManager(db, cfg)
	// Ensure schema so placeholder conversion can be exercised safely
	if err := m.EnsureSchema(); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	return m
}

func Test_convertQueryPlaceholders_SQLite_NoChange(t *testing.T) {
	m := newTestManager(t)
	m.driver = SQLite // force driver
	in := "SELECT * FROM t WHERE a=? AND b=?"
	out := m.convertQueryPlaceholders(in)
	if out != in {
		t.Fatalf("placeholders should remain '?': %s", out)
	}
}

func Test_getAutoIncrementSQL_ByDriver(t *testing.T) {
	m := newTestManager(t)
	m.driver = SQLite
	if got := m.getAutoIncrementSQL(); got != "INTEGER PRIMARY KEY AUTOINCREMENT" {
		t.Fatalf("sqlite autoinc mismatch: %s", got)
	}
	m.driver = PostgreSQL
	if got := m.getAutoIncrementSQL(); got != sqlTypeSerialPK {
		t.Fatalf("postgres autoinc mismatch: %s", got)
	}
}

func Test_RemoveLastPendingMigration_Basic(t *testing.T) {
	m := newTestManager(t)
	// Add two pending migrations with increasing time-based versions
	a := m.AddMigration("alpha", "", "CREATE TABLE a(id INT);", "DROP TABLE a;")
	b := m.AddMigration("beta", "", "CREATE TABLE b(id INT);", "DROP TABLE b;")

	removed, err := m.RemoveLastPendingMigration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed.ID != b.ID {
		t.Fatalf("expected last pending to be %s, got %s", b.ID, removed.ID)
	}

	// Next removal should remove 'a'
	removed2, err := m.RemoveLastPendingMigration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed2.ID != a.ID {
		t.Fatalf("expected next pending to be %s, got %s", a.ID, removed2.ID)
	}

	// No more pending
	if _, err := m.RemoveLastPendingMigration(); err == nil {
		t.Fatalf("expected error when no pending migrations remain")
	}
}

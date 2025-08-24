package migrations

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// helper to open in-memory sqlite for this file
func openSQLiteTB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	// Ensure a single shared connection; with :memory: each connection is a separate DB
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// Exercise processStructFieldsWithError + handleEmbeddedStructWithError recursion and error propagation
func TestAutoMigrator_processStructFieldsWithError_Propagates(t *testing.T) {
	type inner struct {
		Name string `db:"name"`
	}
	type outer struct {
		ID int `db:"id"`
		inner
	}

	am := &AutoMigrator{}
	var seen []string
	wantErr := errors.New("boom")
	err := am.processStructFieldsWithError(reflect.TypeOf(outer{}), func(field reflect.StructField, dbTag string) error {
		seen = append(seen, dbTag)
		if dbTag == "name" { // ensure error when visiting embedded field
			return wantErr
		}
		return nil
	})
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected propagated error, got %v", err)
	}
	joined := strings.Join(seen, ",")
	if !strings.Contains(joined, "id") || !strings.Contains(joined, "name") {
		t.Fatalf("expected to visit both id and name, got %q", joined)
	}
}

// Validate createIndexes builds both normal and unique indexes on SQLite
func TestAutoMigrator_createIndexes_SQLite(t *testing.T) {
	db := openSQLiteTB(t)
	am := &AutoMigrator{db: db, logger: func(string, ...interface{}) {}}

	// Create table schema first
	_, err := db.Exec("CREATE TABLE am_idx (id INTEGER PRIMARY KEY, name TEXT, email TEXT);")
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	// Model describing desired indexes
	type IndexModel struct {
		ID    int    `db:"id"`
		Name  string `db:"name" index:"true"`
		Email string `db:"email" uniqueIndex:"true"`
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := am.createIndexes(tx, "am_idx", reflect.TypeOf(IndexModel{})); err != nil {
		t.Fatalf("createIndexes failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	// Verify indexes created in sqlite_master
	var idxCnt, uidxCnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_am_idx_name'").Scan(&idxCnt); err != nil {
		t.Fatalf("query idx failed: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='uidx_am_idx_email'").Scan(&uidxCnt); err != nil {
		t.Fatalf("query uidx failed: %v", err)
	}
	if idxCnt != 1 || uidxCnt != 1 {
		t.Fatalf("expected both indexes created, idx=%d uidx=%d", idxCnt, uidxCnt)
	}
}

// Exercise SQLite table info scanning via getCurrentTableColumns
func TestAutoMigrator_getCurrentTableColumns_SQLite(t *testing.T) {
	db := openSQLiteTB(t)
	am := &AutoMigrator{db: db, logger: func(string, ...interface{}) {}}

	_, err := db.Exec("CREATE TABLE t_cols (id INTEGER NOT NULL PRIMARY KEY, name TEXT DEFAULT 'x', active INTEGER);")
	if err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	cols, err := am.getCurrentTableColumns("t_cols")
	if err != nil {
		t.Fatalf("getCurrentTableColumns failed: %v", err)
	}
	if _, ok := cols["id"]; !ok {
		t.Fatalf("expected id in columns: %#v", cols)
	}
	if got, ok := cols["name"]; !ok || !strings.Contains(got, "type:TEXT") || !strings.Contains(got, ",nullable:YES") || !strings.Contains(strings.ToLower(got), "default:") {
		t.Fatalf("unexpected name column info: %q", got)
	}
}

// Cover the error path where beginning a transaction fails and file permission change is attempted
func TestAutoMigrator_createTable_BeginTxFail_ChmodError(t *testing.T) {
	db := openSQLiteTB(t)
	am := &AutoMigrator{db: db, logger: func(string, ...interface{}) {}}

	// Close DB to force Begin() error
	_ = db.Close()

	// Minimal model type
	type M struct {
		ID int `db:"id"`
	}
	err := am.createTable("t_fail", reflect.TypeOf(M{}), "mig_fail", "checksum")
	if err == nil || !strings.Contains(err.Error(), "failed to set migration file permissions") {
		t.Fatalf("expected chmod error when Begin fails, got: %v", err)
	}
}

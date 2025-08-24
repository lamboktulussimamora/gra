package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDatabaseInspector_SQLite_SchemaDiscovery(t *testing.T) {
	// Use a shared in-memory database so multiple connections see the same DB
	// and avoid deadlocks from nested queries: database/sql may open more than one
	// connection while iterating rows; with shared cache they all see the same DB.
	db, err := sql.Open("sqlite3", "file:dbinspect1?mode=memory&cache=shared&_busy_timeout=5000")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Allow multiple connections for nested queries; keep it modest.
	db.SetMaxOpenConns(5)
	defer func() { _ = db.Close() }()

	// create a simple table and index
	_, err = db.Exec(`CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT
    );`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = db.Exec(`CREATE UNIQUE INDEX ix_users_name ON users(name);`)
	if err != nil {
		t.Fatalf("create index: %v", err)
	}

	di := NewDatabaseInspector(db, SQLite)
	schema, err := di.GetCurrentSchema()
	if err != nil {
		t.Fatalf("GetCurrentSchema: %v", err)
	}
	tbl, ok := schema["users"]
	if !ok {
		t.Fatalf("expected users table present")
	}
	if _, ok := tbl.Columns["id"]; !ok {
		t.Fatalf("expected id column present")
	}
	if _, ok := tbl.Columns["name"]; !ok {
		t.Fatalf("expected name column present")
	}
	if len(tbl.PrimaryKeys) == 0 {
		t.Fatalf("expected primary keys detected")
	}
	// index should be discovered (not autoindex)
	found := false
	for n := range tbl.Indexes {
		if n == "ix_users_name" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ix_users_name index present, got: %+v", tbl.Indexes)
	}
}

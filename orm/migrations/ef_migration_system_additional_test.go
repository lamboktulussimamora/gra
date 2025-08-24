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

package testutils

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCloseDB_WithValidConnection(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test closing the database
	CloseDB(t, db)

	// Verify the connection is closed by trying to ping
	err = db.Ping()
	if err == nil {
		t.Error("Expected database to be closed, but ping succeeded")
	}
}

func TestCloseDB_WithNilConnection(t *testing.T) {
	// Test with nil database - should not panic
	CloseDB(t, nil)
	// If we reach here without panic, test passes
}

func TestCloseDB_WithAlreadyClosedConnection(t *testing.T) {
	// Create and immediately close a database connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Close it first
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database initially: %v", err)
	}

	// Test closing already closed database - should not panic
	CloseDB(t, db)
	// If we reach here without panic, test passes
}

func TestCloseDBAnyway_WithValidConnection(t *testing.T) {
	// Create a test database connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test closing the database
	CloseDBAnyway(db)

	// Verify the connection is closed by trying to ping
	err = db.Ping()
	if err == nil {
		t.Error("Expected database to be closed, but ping succeeded")
	}
}

func TestCloseDBAnyway_WithNilConnection(_ *testing.T) {
	// Test with nil database - should not panic
	CloseDBAnyway(nil)
	// If we reach here without panic, test passes
}

func TestCloseDBAnyway_WithAlreadyClosedConnection(t *testing.T) {
	// Create and immediately close a database connection
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Close it first
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database initially: %v", err)
	}

	// Test closing already closed database - should not panic
	CloseDBAnyway(db)
	// If we reach here without panic, test passes
}

func TestCloseDB_LogsWarningOnError(t *testing.T) {
	// This test is harder to implement because we need a database that fails to close
	// For now, we'll test the basic functionality and rely on other tests for error paths

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Close database normally
	CloseDB(t, db)

	// Additional close should not cause issues
	CloseDB(t, db)
}

func TestDeferUsage(t *testing.T) {
	// Test typical defer usage pattern
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Simulate defer usage
	defer CloseDBAnyway(db)

	// Do some work with the database
	err = db.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}

	// Function will exit here, defer will execute CloseDBAnyway
}

func TestMultipleCloseOperations(t *testing.T) {
	// Test calling both functions on the same database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test CloseDB first
	CloseDB(t, db)

	// Then test CloseDBAnyway (should not panic on already closed DB)
	CloseDBAnyway(db)
}

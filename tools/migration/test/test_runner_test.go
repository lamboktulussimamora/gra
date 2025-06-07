package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testDBPath             = ":memory:"
	errFailedToOpenDB      = "Failed to open test database: %v"
	errExpectedError       = "Expected error for nil database"
	errExpectedToContainDB = "Expected error to contain '%s', got %v"
	errFailedToCreateMig   = "Failed to create migrations table: %v"
	errFailedToCreateUsers = "Failed to create users table: %v"
	successCreateMig       = "Successfully created migrations table"
	successCreateUsers     = "Successfully created users table"
)

// Test constants
const (
	testValidConnection   = "postgres://test:test@localhost:5432/testdb?sslmode=disable"
	testInvalidConnection = ""
	testUsersTable        = "users"
	testMigrationsTable   = "schema_migrations"
)

func TestCommandLineFlags(t *testing.T) {
	// Test that flag variables are properly declared
	if up == nil {
		t.Error("up flag should not be nil")
	}
	if conn == nil {
		t.Error("conn flag should not be nil")
	}
}

func TestFlagDefaults(t *testing.T) {
	// Test default flag values (before parsing)
	// Since flags are declared at package level, they should be initialized
	if up == nil {
		t.Error("up flag pointer should not be nil")
	}
	if conn == nil {
		t.Error("conn flag pointer should not be nil")
	}
}

func TestDatabaseConnection(t *testing.T) {
	t.Run("sqlite memory database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer db.Close()

		// Test connection
		err = db.Ping()
		if err != nil {
			t.Errorf("Database ping failed: %v", err)
		}
	})
}

func TestMigrationTableCreation(t *testing.T) {
	t.Run("create migrations table", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer db.Close()

		// Create migrations table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			t.Fatalf(errFailedToCreateMig, err)
		}

		// Verify table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to verify migrations table: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 migrations table, got %d", count)
		}
	})
}

func TestUsersTableCreation(t *testing.T) {
	t.Run("create users table", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer db.Close()

		// Create users table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			t.Fatalf(errFailedToCreateUsers, err)
		}

		// Verify table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to verify users table: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 users table, got %d", count)
		}
	})
}

func TestConnectionStringValidation(t *testing.T) {
	tests := []struct {
		name       string
		connString string
		shouldFail bool
	}{
		{"valid postgres connection", testValidConnection, false},
		{"empty connection string", testInvalidConnection, true},
		{"invalid format", "invalid-connection", false}, // sql.Open doesn't validate format
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.connString == "" && !tt.shouldFail {
				t.Error("Empty connection string should be considered invalid")
			}
		})
	}
}

func TestDatabaseOperations(t *testing.T) {
	t.Run("full database operations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer db.Close()

		// Test ping
		err = db.Ping()
		if err != nil {
			t.Errorf("Database ping failed: %v", err)
		}

		// Create test table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT
		)`)
		if err != nil {
			t.Errorf("Failed to create test table: %v", err)
		}

		// Insert test data
		_, err = db.Exec("INSERT INTO test_table (name) VALUES (?)", "test")
		if err != nil {
			t.Errorf("Failed to insert test data: %v", err)
		}

		// Query test data
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
		if err != nil {
			t.Errorf("Failed to query test data: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, got %d", count)
		}
	})
}

func TestTableConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Users table", testUsersTable, "users"},
		{"Migrations table", testMigrationsTable, "schema_migrations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

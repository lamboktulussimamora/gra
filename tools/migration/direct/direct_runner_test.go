package main

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testDBName             = ":memory:"
	testMsg                = "test message"
	errFailedToOpen        = "Failed to open test database: %v"
	errNilDatabaseMsg      = "nil database"
	errExpectedForNil      = "Expected error for nil database"
	errExpectedToContain   = "Expected error to contain '%s', got %v"
	errFailedToEnsure      = "Failed to ensure migration table: %v"
	errFailedToCreate      = "Failed to create test database: %v"
	errExpectedWithNil     = "Expected error with nil database, but got none"
	errFailedToCreateTable = "Failed to create migration table: %v"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"Users table", tableUsers, "users"},
		{"Products table", tableProducts, "products"},
		{"Categories table", tableCategories, "categories"},
		{"Schema migrations table", tableSchemaMigrations, "schema_migrations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.constant)
			}
		})
	}
}

func TestEnsureMigrationTable(t *testing.T) {
	t.Run("successful table creation", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Errorf("ensureMigrationTable failed: %v", err)
		}

		// Verify table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table existence: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 table, got %d", count)
		}
	})

	t.Run(errNilDatabaseMsg, func(t *testing.T) {
		err := ensureMigrationTable(nil)
		if err == nil {
			t.Error(errExpectedForNil)
		}
		if !strings.Contains(err.Error(), errNilDB) {
			t.Errorf(errExpectedToContain, errNilDB, err)
		}
	})
}

func TestShowStatus(t *testing.T) {
	t.Run("show status with empty database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf(errFailedToEnsure, err)
		}

		err = showStatus(db)
		if err != nil {
			t.Errorf("showStatus failed: %v", err)
		}
	})

	t.Run(errNilDatabaseMsg, func(t *testing.T) {
		err := showStatus(nil)
		if err == nil {
			t.Error(errExpectedForNil)
		}
		if !strings.Contains(err.Error(), errNilDB) {
			t.Errorf(errExpectedToContain, errNilDB, err)
		}
	})
}

func TestMigrateUp(t *testing.T) {
	t.Run("migrate up with clean database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf(errFailedToEnsure, err)
		}

		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp failed: %v", err)
		}
	})

	t.Run(errNilDatabaseMsg, func(t *testing.T) {
		err := migrateUp(nil)
		if err == nil {
			t.Error(errExpectedForNil)
		}
		if !strings.Contains(err.Error(), errNilDB) {
			t.Errorf(errExpectedToContain, errNilDB, err)
		}
	})
}

func TestCloseDBWithWarn(t *testing.T) {
	t.Run("close valid database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}

		// Should not panic
		closeDBWithWarn(db)
	})

	t.Run("close nil database", func(t *testing.T) {
		// Should not panic
		closeDBWithWarn(nil)
	})
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

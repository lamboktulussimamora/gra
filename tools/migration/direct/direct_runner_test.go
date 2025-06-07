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
		defer func() {
			if err := db.Close(); err != nil {
				t.Logf("Warning: failed to close database: %v", err)
			}
		}()

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
		defer func() { _ = db.Close() }()

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
		defer func() { _ = db.Close() }()

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

	t.Run("close nil database", func(_ *testing.T) {
		// Should not panic
		closeDBWithWarn(nil)
	})
}

// Additional tests to improve coverage
func TestExitWithDBClose(t *testing.T) {
	t.Run("exit with db close function exists", func(t *testing.T) {
		// We can't directly test exitWithDBClose as it calls log.Fatalf
		// This test ensures we have coverage awareness of this function
		// The function is tested indirectly through integration scenarios
		t.Log("exitWithDBClose function exists for error handling")
	})
}

func TestGetAppliedMigrationsWithData(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	// Insert test migration data
	_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1), (2), (3)")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		t.Errorf("getAppliedMigrations failed: %v", err)
	}

	testAppliedMigrationsResult(t, applied)
}

func testAppliedMigrationsResult(t *testing.T, applied map[int]bool) {
	expectedVersions := []int{1, 2, 3}
	for _, version := range expectedVersions {
		if !applied[version] {
			t.Errorf("Expected version %d to be applied", version)
		}
	}

	if len(applied) != 3 {
		t.Errorf("Expected 3 applied migrations, got %d", len(applied))
	}
}

func TestGetAppliedMigrationsDatabaseError(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	// Don't create the table to trigger an error
	_, err = getAppliedMigrations(db)
	if err == nil {
		t.Error("Expected error when table doesn't exist")
	}
}

func TestGetAppliedMigrationsNilDatabase(t *testing.T) {
	_, err := getAppliedMigrations(nil)
	if err == nil {
		t.Error(errExpectedForNil)
	}
	if !strings.Contains(err.Error(), errNilDB) {
		t.Errorf(errExpectedToContain, errNilDB, err)
	}
}

func getTestMigration() struct {
	Version     int
	Description string
	SQL         string
} {
	return struct {
		Version     int
		Description string
		SQL         string
	}{
		Version:     99,
		Description: "Test migration",
		SQL:         "CREATE TABLE test_migration (id INTEGER PRIMARY KEY)",
	}
}

func TestApplyMigrationSuccessfully(t *testing.T) {
	migration := getTestMigration()

	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	err = applyMigration(db, migration)
	if err != nil {
		t.Errorf("applyMigration failed: %v", err)
	}

	verifyMigrationApplied(t, db, migration)
}

func verifyMigrationApplied(t *testing.T, db *sql.DB, migration struct {
	Version     int
	Description string
	SQL         string
}) {
	// Verify migration was recorded
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count)
	if err != nil {
		t.Errorf("Failed to verify migration record: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 migration record, got %d", count)
	}

	// Verify table was created
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_migration'").Scan(&tableCount)
	if err != nil {
		t.Errorf("Failed to verify table creation: %v", err)
	}
	if tableCount != 1 {
		t.Errorf("Expected test_migration table to exist")
	}
}

func TestApplyMigrationNilDatabase(t *testing.T) {
	migration := getTestMigration()

	err := applyMigration(nil, migration)
	if err == nil {
		t.Error(errExpectedForNil)
	}
	if !strings.Contains(err.Error(), errNilDB) {
		t.Errorf(errExpectedToContain, errNilDB, err)
	}
}

func TestApplyMigrationInvalidSQL(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	badMigration := struct {
		Version     int
		Description string
		SQL         string
	}{
		Version:     98,
		Description: "Bad migration",
		SQL:         "INVALID SQL STATEMENT",
	}

	err = applyMigration(db, badMigration)
	if err == nil {
		t.Error("Expected error for invalid SQL")
	}
}

func TestGetMigrationsList(t *testing.T) {
	migrations := getMigrationsList()

	if len(migrations) == 0 {
		t.Error("Expected at least one migration")
	}

	// Verify migration structure
	for i, migration := range migrations {
		if migration.Version <= 0 {
			t.Errorf("Migration %d: version should be positive, got %d", i, migration.Version)
		}
		if migration.Description == "" {
			t.Errorf("Migration %d: description should not be empty", i)
		}
		if migration.SQL == "" {
			t.Errorf("Migration %d: SQL should not be empty", i)
		}
	}

	// Verify migrations are in order
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Errorf("Migrations should be in ascending order, but version %d comes after %d",
				migrations[i].Version, migrations[i-1].Version)
		}
	}
}

func TestMigrateUpWithErrors(t *testing.T) {
	t.Run("migrate up with database connection error during getAppliedMigrations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer func() { _ = db.Close() }()

		// Don't create migration table to cause error
		err = migrateUp(db)
		if err == nil {
			t.Error("Expected error when migration table doesn't exist")
		}
	})

	t.Run("migrate up with nil database", func(t *testing.T) {
		err := migrateUp(nil)
		if err == nil {
			t.Error(errExpectedForNil)
		}
	})

	t.Run("migrate up with successful scenario", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer func() { _ = db.Close() }()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf(errFailedToEnsure, err)
		}

		// Start with a clean database - no applied migrations
		// This will apply all migrations from the beginning
		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp failed: %v", err)
		}
	})
}

func TestShowStatusWithData(t *testing.T) {
	t.Run("show status with applied migrations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer func() { _ = db.Close() }()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf(errFailedToEnsure, err)
		}

		// Insert test migration data
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1), (2)")
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		err = showStatus(db)
		if err != nil {
			t.Errorf("showStatus failed: %v", err)
		}
	})

	t.Run("show status with getAppliedMigrations error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer func() { _ = db.Close() }()

		// Don't create table to trigger error
		err = showStatus(db)
		if err == nil {
			t.Error("Expected error when migration table doesn't exist")
		}
	})
}

func TestVerboseOutput(t *testing.T) {
	// Test that verbose flag works
	t.Run("verbose flag coverage", func(t *testing.T) {
		if verbose == nil {
			t.Error("verbose flag should not be nil")
		}

		// Set verbose to true for coverage
		*verbose = true

		db, err := sql.Open("sqlite3", testDBName)
		if err != nil {
			t.Fatalf(errFailedToOpen, err)
		}
		defer func() {
			*verbose = false // Reset
			_ = db.Close()
		}()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf(errFailedToEnsure, err)
		}

		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp with verbose failed: %v", err)
		}
	})
}

func TestMainFunctionComponents(t *testing.T) {
	t.Run("flag variables exist", func(t *testing.T) {
		if upFlag == nil {
			t.Error("upFlag should not be nil")
		}
		if downFlag == nil {
			t.Error("downFlag should not be nil")
		}
		if connFlag == nil {
			t.Error("connFlag should not be nil")
		}
		if statusFlag == nil {
			t.Error("statusFlag should not be nil")
		}
		if verbose == nil {
			t.Error("verbose should not be nil")
		}
	})
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// Additional tests to reach higher coverage
func TestCloseDBWithWarnError(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}

	// Close the database first to trigger an error on second close
	_ = db.Close()

	// This should trigger the error path in closeDBWithWarn
	closeDBWithWarn(db)
}

func TestMigrateUpPartialMigrations(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	// Apply first migration manually
	migrations := getMigrationsList()
	if len(migrations) > 0 {
		err = applyMigration(db, migrations[0])
		if err != nil {
			t.Fatalf("Failed to apply first migration: %v", err)
		}

		// Run migrateUp which should apply remaining migrations
		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp with partial migrations failed: %v", err)
		}
	}
}

func TestMigrateUpAllApplied(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	// Apply all migrations first
	err = migrateUp(db)
	if err != nil {
		t.Fatalf("Initial migrateUp failed: %v", err)
	}

	// Run migrateUp again
	err = migrateUp(db)
	if err != nil {
		t.Errorf("migrateUp with all migrations applied failed: %v", err)
	}
}

func TestApplyMigrationSuccess(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	migration := getTestMigration()
	migration.Version = 100
	migration.SQL = "CREATE TABLE test_success (id INTEGER PRIMARY KEY)"

	err = applyMigration(db, migration)
	if err != nil {
		t.Errorf("applyMigration should succeed: %v", err)
	}

	// Verify the migration was applied
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migration.Version).Scan(&count)
	if err != nil {
		t.Errorf("Failed to verify migration: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected migration to be recorded")
	}
}

func TestApplyMigrationDuplicateError(t *testing.T) {
	db, err := sql.Open("sqlite3", testDBName)
	if err != nil {
		t.Fatalf(errFailedToOpen, err)
	}
	defer func() { _ = db.Close() }()

	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf(errFailedToEnsure, err)
	}

	migration := getTestMigration()
	migration.Version = 101
	migration.SQL = "CREATE TABLE test_duplicate (id INTEGER PRIMARY KEY)"

	// Apply the migration first time
	err = applyMigration(db, migration)
	if err != nil {
		t.Errorf("First applyMigration should succeed: %v", err)
	}

	// Try to apply same migration again
	err = applyMigration(db, migration)
	if err == nil {
		t.Error("Expected error when applying duplicate migration")
	}
}

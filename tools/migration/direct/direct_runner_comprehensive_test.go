package main

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestCloseDBWithWarnComprehensive tests the closeDBWithWarn function with various scenarios
func TestCloseDBWithWarnComprehensive(t *testing.T) {
	t.Run("close valid database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		// Should not panic or cause issues
		closeDBWithWarn(db)
	})

	t.Run("close already closed database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		// Close first time
		err = db.Close()
		if err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}

		// Should handle already closed database gracefully
		closeDBWithWarn(db)
	})

	t.Run("close nil database", func(t *testing.T) {
		// Should handle nil database gracefully
		closeDBWithWarn(nil)
	})
}

// TestApplyMigrationComprehensive tests applyMigration function with various scenarios
func TestApplyMigrationComprehensive(t *testing.T) {
	t.Run("apply migration with transaction rollback on SQL error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Create migration with invalid SQL
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Invalid migration",
			SQL:         "CREATE TABLE invalid syntax here",
		}

		err = applyMigration(db, migration)
		if err == nil {
			t.Error("Expected error for invalid SQL, but got none")
		}

		// Verify no partial changes were made
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM "+tableSchemaMigrations+" WHERE version = ?", migration.Version).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check migration record: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected no migration record after failed migration, found %d", count)
		}
	})

	t.Run("apply migration with transaction rollback on record error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Drop the migration table to cause record error
		_, err = db.Exec("DROP TABLE " + tableSchemaMigrations)
		if err != nil {
			t.Fatalf("Failed to drop migration table: %v", err)
		}

		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Test migration",
			SQL:         "CREATE TABLE test_table (id INTEGER)",
		}

		err = applyMigration(db, migration)
		if err == nil {
			t.Error("Expected error when migration table doesn't exist, but got none")
		}
	})

	t.Run("apply migration with nil database", func(t *testing.T) {
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Test migration",
			SQL:         "CREATE TABLE test_table (id INTEGER)",
		}

		err := applyMigration(nil, migration)
		if err == nil {
			t.Error("Expected error for nil database, but got none")
		}
		if !strings.Contains(err.Error(), errNilDB) {
			t.Errorf("Expected error to contain '%s', got: %v", errNilDB, err)
		}
	})

	t.Run("successful migration application", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Create test table",
			SQL:         "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)",
		}

		err = applyMigration(db, migration)
		if err != nil {
			t.Errorf("Failed to apply valid migration: %v", err)
		}

		// Verify table was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table existence: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected test table to exist, found %d tables", count)
		}

		// Verify migration was recorded
		err = db.QueryRow("SELECT COUNT(*) FROM "+tableSchemaMigrations+" WHERE version = ?", migration.Version).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check migration record: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected migration record to exist, found %d records", count)
		}
	})
}

// TestGetAppliedMigrationsComprehensive tests getAppliedMigrations with edge cases
func TestGetAppliedMigrationsComprehensive(t *testing.T) {
	t.Run("get migrations with table scan error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		// Create migration table with wrong schema to cause scan error
		_, err = db.Exec("CREATE TABLE " + tableSchemaMigrations + " (version TEXT)")
		if err != nil {
			t.Fatalf("Failed to create migration table: %v", err)
		}

		// Insert non-integer version to cause scan error
		_, err = db.Exec("INSERT INTO "+tableSchemaMigrations+" (version) VALUES (?)", "not-a-number")
		if err != nil {
			t.Fatalf("Failed to insert invalid version: %v", err)
		}

		applied, err := getAppliedMigrations(db)
		if err == nil {
			t.Error("Expected error when scanning invalid version, but got none")
		}
		if applied != nil {
			t.Error("Expected nil applied migrations on error")
		}
	})

	t.Run("get migrations with query execution error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		// Don't create the migration table to cause query error
		applied, err := getAppliedMigrations(db)
		if err == nil {
			t.Error("Expected error when migration table doesn't exist, but got none")
		}
		if applied != nil {
			t.Error("Expected nil applied migrations on error")
		}
	})

	t.Run("get migrations with successful query", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Insert some test migrations
		_, err = db.Exec("INSERT INTO "+tableSchemaMigrations+" (version) VALUES (?)", 1)
		if err != nil {
			t.Fatalf("Failed to insert migration: %v", err)
		}

		_, err = db.Exec("INSERT INTO "+tableSchemaMigrations+" (version) VALUES (?)", 3)
		if err != nil {
			t.Fatalf("Failed to insert migration: %v", err)
		}

		applied, err := getAppliedMigrations(db)
		if err != nil {
			t.Errorf("Failed to get applied migrations: %v", err)
		}
		if len(applied) != 2 {
			t.Errorf("Expected 2 applied migrations, got %d", len(applied))
		}
		if !applied[1] || !applied[3] {
			t.Errorf("Expected migrations 1 and 3 to be applied, got %v", applied)
		}
	})
}

// TestMigrateUpComprehensive tests migrateUp function with various scenarios
func TestMigrateUpComprehensive(t *testing.T) {
	t.Run("migrate up with no pending migrations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Apply all existing migrations first
		migrations := getMigrationsList()
		for _, migration := range migrations {
			err = applyMigration(db, migration)
			if err != nil {
				t.Fatalf("Failed to apply migration %d: %v", migration.Version, err)
			}
		}

		// Now try to migrate up again - should be no-op
		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp should handle no pending migrations: %v", err)
		}
	})

	t.Run("migrate up with some pending migrations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Apply only first migration
		migrations := getMigrationsList()
		if len(migrations) > 0 {
			err = applyMigration(db, migrations[0])
			if err != nil {
				t.Fatalf("Failed to apply first migration: %v", err)
			}
		}

		// Now migrate up should apply remaining migrations
		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp failed: %v", err)
		}

		// Verify all migrations were applied
		applied, err := getAppliedMigrations(db)
		if err != nil {
			t.Errorf("Failed to get applied migrations: %v", err)
		}
		if len(applied) != len(migrations) {
			t.Errorf("Expected %d applied migrations, got %d", len(migrations), len(applied))
		}
	})
}

// TestShowStatusComprehensive tests showStatus function with various scenarios
func TestShowStatusComprehensive(t *testing.T) {
	t.Run("show status with applied migrations error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		// Don't create migration table to cause getAppliedMigrations error
		err = showStatus(db)
		if err == nil {
			t.Error("Expected error when migration table doesn't exist")
		}
	})

	t.Run("show status with successful data", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Apply some migrations
		migrations := getMigrationsList()
		if len(migrations) > 1 {
			err = applyMigration(db, migrations[0])
			if err != nil {
				t.Fatalf("Failed to apply migration: %v", err)
			}
		}

		err = showStatus(db)
		if err != nil {
			t.Errorf("showStatus failed: %v", err)
		}
	})
}

// TestVerboseLoggingComprehensive tests verbose flag functionality
func TestVerboseLoggingComprehensive(t *testing.T) {
	t.Run("verbose flag affects output", func(t *testing.T) {
		// Store original verbose state
		originalVerbose := *verbose
		defer func() { *verbose = originalVerbose }()

		// Test with verbose enabled
		*verbose = true

		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Test verbose migration",
			SQL:         "CREATE TABLE test_verbose (id INTEGER)",
		}

		// This should include verbose output (we can't easily capture it in tests)
		err = applyMigration(db, migration)
		if err != nil {
			t.Errorf("Failed to apply migration with verbose logging: %v", err)
		}

		// Test with verbose disabled
		*verbose = false

		migration.Version = 2
		migration.Description = "Test silent migration"
		migration.SQL = "CREATE TABLE test_silent (id INTEGER)"

		err = applyMigration(db, migration)
		if err != nil {
			t.Errorf("Failed to apply migration without verbose logging: %v", err)
		}
	})
}

// TestDatabaseStateManagementComprehensive tests database state management
func TestDatabaseStateManagementComprehensive(t *testing.T) {
	t.Run("database state after successful operations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		// Ensure migration table
		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Check database state
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to count tables: %v", err)
		}
		if count < 1 {
			t.Errorf("Expected at least 1 table (migration table), got %d", count)
		}

		// Apply a migration
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "State test migration",
			SQL:         "CREATE TABLE state_test (id INTEGER, name TEXT)",
		}

		err = applyMigration(db, migration)
		if err != nil {
			t.Errorf("Failed to apply migration: %v", err)
		}

		// Verify final state
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to count tables after migration: %v", err)
		}
		if count < 2 {
			t.Errorf("Expected at least 2 tables after migration, got %d", count)
		}
	})

	t.Run("transaction rollback handling", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Test concurrent transaction handling
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Concurrent test migration",
			SQL:         "CREATE TABLE concurrent_test (id INTEGER PRIMARY KEY)",
		}

		// Apply migration successfully
		err = applyMigration(db, migration)
		if err != nil {
			t.Errorf("Failed to apply migration: %v", err)
		}

		// Try to apply same migration again (should fail due to duplicate version)
		err = applyMigration(db, migration)
		if err == nil {
			t.Error("Expected error when applying duplicate migration version")
		}
	})
}

// TestGetMigrationsListComprehensive tests getMigrationsList function
func TestGetMigrationsListComprehensive(t *testing.T) {
	t.Run("get predefined migrations list", func(t *testing.T) {
		migrations := getMigrationsList()

		if len(migrations) == 0 {
			t.Error("Expected at least one migration in the list")
		}

		// Verify structure of first migration
		if len(migrations) > 0 {
			first := migrations[0]
			if first.Version <= 0 {
				t.Error("Expected positive version number")
			}
			if first.Description == "" {
				t.Error("Expected non-empty description")
			}
			if first.SQL == "" {
				t.Error("Expected non-empty SQL")
			}
		}

		// Verify migrations are ordered by version
		for i := 1; i < len(migrations); i++ {
			if migrations[i].Version <= migrations[i-1].Version {
				t.Errorf("Migrations should be ordered by version, got %d after %d",
					migrations[i].Version, migrations[i-1].Version)
			}
		}
	})

	t.Run("verify migration SQL validity", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer closeDBWithWarn(db)

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		migrations := getMigrationsList()

		// Test migrations sequentially since they may depend on each other
		for _, migration := range migrations {
			err = applyMigration(db, migration)
			if err != nil {
				t.Errorf("Migration %d has invalid SQL: %v", migration.Version, err)
			}
		}
	})
}

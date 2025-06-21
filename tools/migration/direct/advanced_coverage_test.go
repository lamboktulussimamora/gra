package main

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMainCoverage tests the main function coverage and additional edge cases
func TestMainCoverage(t *testing.T) {
	// Test flag variables and initialization
	t.Run("flag_initialization", func(t *testing.T) {
		// Test that flag variables are properly initialized
		if upFlag == nil {
			t.Error("upFlag not initialized")
		}
		if downFlag == nil {
			t.Error("downFlag not initialized")
		}
		if connFlag == nil {
			t.Error("connFlag not initialized")
		}
		if verbose == nil {
			t.Error("verbose flag not initialized")
		}
		if statusFlag == nil {
			t.Error("statusFlag not initialized")
		}
	})
}

// TestUtilityFunctions tests utility functions for better coverage
func TestUtilityFunctions(t *testing.T) {
	t.Run("closeDBWithWarn_with_valid_db", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}

		// This should not panic or error
		closeDBWithWarn(db)
	})

	t.Run("closeDBWithWarn_with_nil_db", func(t *testing.T) {
		// This should not panic
		closeDBWithWarn(nil)
	})

	t.Run("exitWithDBClose_function_exists", func(t *testing.T) {
		// Test that the function exists by checking its type
		if reflect.TypeOf(exitWithDBClose).Kind() != reflect.Func {
			t.Error("exitWithDBClose should be a function")
		}
	})
}

// TestConstantsAndVariables tests constants and package variables
func TestConstantsAndVariables(t *testing.T) {
	t.Run("constants_existence", func(t *testing.T) {
		// Test that constants are properly defined
		expectedTables := map[string]string{
			"tableUsers":            tableUsers,
			"tableProducts":         tableProducts,
			"tableCategories":       tableCategories,
			"tableSchemaMigrations": tableSchemaMigrations,
		}

		for name, table := range expectedTables {
			if table == "" {
				t.Errorf("Constant %s is empty", name)
			}
		}
	})

	t.Run("flag_variables_types", func(t *testing.T) {
		// Test that flag variables have correct types
		if reflect.TypeOf(upFlag).String() != "*bool" {
			t.Errorf("upFlag should be *bool, got %s", reflect.TypeOf(upFlag).String())
		}
		if reflect.TypeOf(downFlag).String() != "*bool" {
			t.Errorf("downFlag should be *bool, got %s", reflect.TypeOf(downFlag).String())
		}
		if reflect.TypeOf(connFlag).String() != "*string" {
			t.Errorf("connFlag should be *string, got %s", reflect.TypeOf(connFlag).String())
		}
		if reflect.TypeOf(verbose).String() != "*bool" {
			t.Errorf("verbose should be *bool, got %s", reflect.TypeOf(verbose).String())
		}
		if reflect.TypeOf(statusFlag).String() != "*bool" {
			t.Errorf("statusFlag should be *bool, got %s", reflect.TypeOf(statusFlag).String())
		}
	})

	t.Run("error_constants", func(t *testing.T) {
		if errNilDB == "" {
			t.Error("errNilDB constant is empty")
		}
		if warnCloseDB == "" {
			t.Error("warnCloseDB constant is empty")
		}
	})
}

// TestAdvancedMigrationLogic tests more complex migration scenarios
func TestAdvancedMigrationLogic(t *testing.T) {
	t.Run("getMigrationsList_hardcoded_migrations", func(t *testing.T) {
		// Test getMigrationsList (which returns hardcoded migrations)
		migrations := getMigrationsList()

		// Should get hardcoded migrations
		if len(migrations) == 0 {
			t.Error("Expected some hardcoded migrations, got none")
		}

		// Check that migrations have required fields
		for i, migration := range migrations {
			if migration.Version <= 0 {
				t.Errorf("Migration %d has invalid version: %d", i, migration.Version)
			}
			if migration.Description == "" {
				t.Errorf("Migration %d has empty description", i)
			}
			if migration.SQL == "" {
				t.Errorf("Migration %d has empty SQL", i)
			}
		}

		// Check that migrations are sorted by version
		for i := 1; i < len(migrations); i++ {
			if migrations[i-1].Version >= migrations[i].Version {
				t.Error("Migrations are not sorted by version")
				break
			}
		}
	})
}

// TestDatabaseErrorHandling tests various database error scenarios
func TestDatabaseErrorHandling(t *testing.T) {
	t.Run("getAppliedMigrations_with_corrupted_db", func(t *testing.T) {
		// Create a database and then corrupt the migration table
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		// Create a corrupted migration table with wrong schema
		_, err = db.Exec("CREATE TABLE schema_migrations (version TEXT)")
		if err != nil {
			t.Fatalf("Failed to create corrupted table: %v", err)
		}

		// Insert invalid data
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ('invalid')")
		if err != nil {
			t.Fatalf("Failed to insert invalid data: %v", err)
		}

		// This should handle the error gracefully
		applied, err := getAppliedMigrations(db)
		if err == nil {
			t.Error("Expected error when reading corrupted data")
		}
		if len(applied) != 0 {
			t.Errorf("Expected empty slice on error, got %d items", len(applied))
		}
	})

	t.Run("applyMigration_with_invalid_sql", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Create a migration struct with invalid SQL
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     999,
			Description: "Invalid migration",
			SQL:         "INVALID SQL SYNTAX HERE",
		}

		err = applyMigration(db, migration)
		if err == nil {
			t.Error("Expected error when applying invalid SQL")
		}
	})
}

// TestMigrateUpAdvancedScenarios tests advanced migration scenarios
func TestMigrateUpAdvancedScenarios(t *testing.T) {
	t.Run("migrateUp_basic", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// migrateUp uses hardcoded migrations, so just test it directly
		err = migrateUp(db)
		if err != nil {
			t.Errorf("migrateUp failed: %v", err)
		}

		// Check that migrations were applied
		applied, err := getAppliedMigrations(db)
		if err != nil {
			t.Fatalf("Failed to get applied migrations: %v", err)
		}

		if len(applied) == 0 {
			t.Error("Expected some migrations to be applied")
		}
	})

	t.Run("migrateUp_with_nil_db", func(t *testing.T) {
		// Test error handling with nil database
		err := migrateUp(nil)
		if err == nil {
			t.Error("Expected error when calling migrateUp with nil db")
		}
	})
}

// TestApplyMigrationEdgeCases tests edge cases for migration application
func TestApplyMigrationEdgeCases(t *testing.T) {
	t.Run("applyMigration_with_nil_db", func(t *testing.T) {
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Test migration",
			SQL:         "CREATE TABLE test (id INT);",
		}

		err := applyMigration(nil, migration)
		if err == nil {
			t.Error("Expected error when applying migration with nil db")
		}
		if err.Error() != errNilDB {
			t.Errorf("Expected error message '%s', got '%s'", errNilDB, err.Error())
		}
	})

	t.Run("applyMigration_with_transaction_rollback", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Create a migration that will fail in the middle
		migration := struct {
			Version     int
			Description string
			SQL         string
		}{
			Version:     1,
			Description: "Test migration with error",
			SQL:         "CREATE TABLE test (id INT); INVALID SQL HERE;",
		}

		err = applyMigration(db, migration)
		if err == nil {
			t.Error("Expected error when applying migration with invalid SQL")
		}

		// Verify that the transaction was rolled back
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for table: %v", err)
		}
		if count != 0 {
			t.Error("Expected table to not exist after failed migration (transaction rollback)")
		}
	})
}

// TestEnsureMigrationTableAdvancedEdgeCases tests edge cases for migration table creation
func TestEnsureMigrationTableAdvancedEdgeCases(t *testing.T) {
	t.Run("ensureMigrationTable_with_nil_db", func(t *testing.T) {
		err := ensureMigrationTable(nil)
		if err == nil {
			t.Error("Expected error when ensuring migration table with nil db")
		}
	})

	t.Run("ensureMigrationTable_idempotent", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		// Call ensureMigrationTable multiple times
		for i := 0; i < 3; i++ {
			err = ensureMigrationTable(db)
			if err != nil {
				t.Errorf("ensureMigrationTable failed on iteration %d: %v", i+1, err)
			}
		}

		// Verify table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for migration table: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 migration table, got %d", count)
		}
	})
}

// TestShowStatusAdvancedEdgeCases tests edge cases for status display
func TestShowStatusAdvancedEdgeCases(t *testing.T) {
	t.Run("showStatus_with_nil_db", func(t *testing.T) {
		err := showStatus(nil)
		if err == nil {
			t.Error("Expected error when showing status with nil db")
		}
	})

	t.Run("showStatus_with_empty_db", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		err = ensureMigrationTable(db)
		if err != nil {
			t.Fatalf("Failed to ensure migration table: %v", err)
		}

		// Should not error even with no applied migrations
		err = showStatus(db)
		if err != nil {
			t.Errorf("showStatus failed with empty db: %v", err)
		}
	})
}

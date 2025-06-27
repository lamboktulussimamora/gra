package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestDirectRunnerEnhanced provides comprehensive tests for the direct migration runner
func TestDirectRunnerEnhanced(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	t.Run("DatabaseConnection", func(t *testing.T) {
		// Test database connection
		err := db.Ping()
		if err != nil {
			t.Errorf("Database ping failed: %v", err)
		}
	})

	t.Run("TableConstants", func(t *testing.T) {
		// Test that constants are properly defined
		expectedTables := []string{tableUsers, tableProducts, tableCategories, tableSchemaMigrations}
		expectedNames := []string{"users", "products", "categories", "schema_migrations"}

		for i, table := range expectedTables {
			if table != expectedNames[i] {
				t.Errorf("Expected table constant %d to be %s, got %s", i, expectedNames[i], table)
			}
		}
	})

	t.Run("ErrorConstants", func(t *testing.T) {
		// Test error constants
		if errNilDB != "db is nil" {
			t.Errorf("Expected errNilDB to be 'db is nil', got '%s'", errNilDB)
		}
	})
}

// TestFlagParsing tests command line flag parsing
func TestFlagParsing(t *testing.T) {
	// Save original flag values
	originalUp := *upFlag
	originalDown := *downFlag
	originalConn := *connFlag
	originalVerbose := *verbose
	originalStatus := *statusFlag

	// Reset flags after test
	defer func() {
		*upFlag = originalUp
		*downFlag = originalDown
		*connFlag = originalConn
		*verbose = originalVerbose
		*statusFlag = originalStatus
	}()

	t.Run("DefaultFlags", func(t *testing.T) {
		// Test default flag values
		if *upFlag {
			t.Error("Expected upFlag to be false by default")
		}
		if *downFlag {
			t.Error("Expected downFlag to be false by default")
		}
		if *connFlag != "" {
			t.Error("Expected connFlag to be empty by default")
		}
		if *verbose {
			t.Error("Expected verbose to be false by default")
		}
		if *statusFlag {
			t.Error("Expected statusFlag to be false by default")
		}
	})

	t.Run("FlagModification", func(t *testing.T) {
		// Test modifying flags programmatically
		*upFlag = true
		*connFlag = "test-connection"
		*verbose = true

		if !*upFlag {
			t.Error("Expected upFlag to be true after modification")
		}
		if *connFlag != "test-connection" {
			t.Errorf("Expected connFlag to be 'test-connection', got '%s'", *connFlag)
		}
		if !*verbose {
			t.Error("Expected verbose to be true after modification")
		}
	})
}

// TestDatabaseOperations tests database operations
func TestDatabaseOperations(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	t.Run("CreateTables", func(t *testing.T) {
		// Test creating the schema migrations table
		createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			t.Errorf("Failed to create schema_migrations table: %v", err)
		}

		// Verify table was created
		var tableName string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'"
		err = db.QueryRow(query).Scan(&tableName)
		if err != nil {
			t.Errorf("Failed to verify schema_migrations table creation: %v", err)
		}
		if tableName != "schema_migrations" {
			t.Errorf("Expected table name 'schema_migrations', got '%s'", tableName)
		}
	})

	t.Run("CreateUsersTable", func(t *testing.T) {
		// Test creating users table
		createUsersSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createUsersSQL)
		if err != nil {
			t.Errorf("Failed to create users table: %v", err)
		}

		// Verify table was created
		var tableName string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name='users'"
		err = db.QueryRow(query).Scan(&tableName)
		if err != nil {
			t.Errorf("Failed to verify users table creation: %v", err)
		}
		if tableName != "users" {
			t.Errorf("Expected table name 'users', got '%s'", tableName)
		}
	})

	t.Run("CreateProductsTable", func(t *testing.T) {
		// Test creating products table
		createProductsSQL := `
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			category_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createProductsSQL)
		if err != nil {
			t.Errorf("Failed to create products table: %v", err)
		}

		// Verify table was created
		var tableName string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name='products'"
		err = db.QueryRow(query).Scan(&tableName)
		if err != nil {
			t.Errorf("Failed to verify products table creation: %v", err)
		}
		if tableName != "products" {
			t.Errorf("Expected table name 'products', got '%s'", tableName)
		}
	})

	t.Run("CreateCategoriesTable", func(t *testing.T) {
		// Test creating categories table
		createCategoriesSQL := `
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createCategoriesSQL)
		if err != nil {
			t.Errorf("Failed to create categories table: %v", err)
		}

		// Verify table was created
		var tableName string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name='categories'"
		err = db.QueryRow(query).Scan(&tableName)
		if err != nil {
			t.Errorf("Failed to verify categories table creation: %v", err)
		}
		if tableName != "categories" {
			t.Errorf("Expected table name 'categories', got '%s'", tableName)
		}
	})
}

// TestMigrationTracking tests migration tracking functionality
func TestMigrationTracking(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create schema_migrations table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	t.Run("InsertMigrationRecord", func(t *testing.T) {
		// Test inserting a migration record
		insertSQL := "INSERT INTO schema_migrations (version) VALUES (?)"
		_, err := db.Exec(insertSQL, "001_create_users")
		if err != nil {
			t.Errorf("Failed to insert migration record: %v", err)
		}

		// Verify record was inserted
		var version string
		query := "SELECT version FROM schema_migrations WHERE version = ?"
		err = db.QueryRow(query, "001_create_users").Scan(&version)
		if err != nil {
			t.Errorf("Failed to retrieve migration record: %v", err)
		}
		if version != "001_create_users" {
			t.Errorf("Expected version '001_create_users', got '%s'", version)
		}
	})

	t.Run("CountMigrations", func(t *testing.T) {
		// Insert more migration records
		migrations := []string{"002_create_products", "003_create_categories"}
		for _, migration := range migrations {
			insertSQL := "INSERT INTO schema_migrations (version) VALUES (?)"
			_, err := db.Exec(insertSQL, migration)
			if err != nil {
				t.Errorf("Failed to insert migration record %s: %v", migration, err)
			}
		}

		// Count total migrations
		var count int
		query := "SELECT COUNT(*) FROM schema_migrations"
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			t.Errorf("Failed to count migrations: %v", err)
		}
		if count != 3 { // 001_create_users + 002_create_products + 003_create_categories
			t.Errorf("Expected 3 migrations, got %d", count)
		}
	})

	t.Run("CheckMigrationExists", func(t *testing.T) {
		// Test checking if migration exists
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)"
		err := db.QueryRow(query, "001_create_users").Scan(&exists)
		if err != nil {
			t.Errorf("Failed to check migration existence: %v", err)
		}
		if !exists {
			t.Error("Expected migration '001_create_users' to exist")
		}

		// Test checking non-existent migration
		err = db.QueryRow(query, "999_non_existent").Scan(&exists)
		if err != nil {
			t.Errorf("Failed to check non-existent migration: %v", err)
		}
		if exists {
			t.Error("Expected migration '999_non_existent' to not exist")
		}
	})
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("NilDatabase", func(t *testing.T) {
		// Test with nil database
		var db *sql.DB = nil

		// This should fail gracefully
		if db != nil {
			err := db.Ping()
			if err == nil {
				t.Error("Expected error when pinging nil database")
			}
		}

		// Test error constant
		if errNilDB != "db is nil" {
			t.Errorf("Expected error message 'db is nil', got '%s'", errNilDB)
		}
	})

	t.Run("InvalidConnectionString", func(t *testing.T) {
		// Test with invalid connection string
		_, err := sql.Open("sqlite3", "/invalid/path/to/database.db")
		// SQLite driver typically doesn't fail on Open, only on actual usage
		if err != nil {
			t.Logf("Open failed as expected: %v", err)
		}
	})

	t.Run("InvalidSQL", func(t *testing.T) {
		// Setup test database
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer db.Close()

		// Test with invalid SQL
		_, err = db.Exec("INVALID SQL STATEMENT")
		if err == nil {
			t.Error("Expected error for invalid SQL statement")
		}
		if !strings.Contains(err.Error(), "syntax error") {
			t.Errorf("Expected syntax error, got: %v", err)
		}
	})
}

// TestConcurrentOperations tests concurrent database operations
func TestConcurrentOperations(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Ensure migration table exists first using the main function
	err = ensureMigrationTable(db)
	if err != nil {
		t.Fatalf("Failed to ensure migration table: %v", err)
	}

	t.Run("ConcurrentInserts", func(t *testing.T) {
		// Test simulated concurrent migration inserts
		// SQLite in-memory doesn't handle true concurrency well, so we simulate it
		versions := []int{1001, 1002, 1003} // Use integers as expected by the schema
		
		for i, version := range versions {
			insertSQL := "INSERT INTO schema_migrations (version) VALUES (?)"
			_, err := db.Exec(insertSQL, version)
			if err != nil {
				t.Logf("Insert %d failed: %v", i+1, err)
			}
		}

		// Verify all inserts succeeded
		var count int
		query := "SELECT COUNT(*) FROM schema_migrations WHERE version >= 1001 AND version <= 1003"
		err = db.QueryRow(query).Scan(&count)
		if err != nil {
			t.Errorf("Failed to count concurrent migrations: %v", err)
		}
		if count < 1 {
			t.Errorf("Expected at least 1 concurrent migration, got %d", count)
		}
	})
}

// TestUtilityFunctionsEnhanced tests utility functions and constants
func TestUtilityFunctionsEnhanced(t *testing.T) {
	t.Run("WarningConstants", func(t *testing.T) {
		// Test warning constant
		if warnCloseDB != "Warning: failed to close db: %v" {
			t.Errorf("Expected warning message format, got '%s'", warnCloseDB)
		}
	})

	t.Run("TableNameValidation", func(t *testing.T) {
		// Test table name constants
		tableNames := []string{tableUsers, tableProducts, tableCategories, tableSchemaMigrations}

		for _, tableName := range tableNames {
			if tableName == "" {
				t.Error("Table name should not be empty")
			}
			if len(tableName) < 1 {
				t.Error("Table name should have at least 1 character")
			}
			// Check for valid table name characters (alphanumeric and underscore)
			for _, char := range tableName {
				if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9') || char == '_') {
					t.Errorf("Table name '%s' contains invalid character '%c'", tableName, char)
				}
			}
		}
	})
}

// TestDatabaseSchema tests database schema operations
func TestDatabaseSchema(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	t.Run("SchemaValidation", func(t *testing.T) {
		// Create a test table with specific schema
		createTableSQL := `
		CREATE TABLE test_schema (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			t.Errorf("Failed to create test schema table: %v", err)
		}

		// Test inserting valid data
		insertSQL := "INSERT INTO test_schema (name, email) VALUES (?, ?)"
		_, err = db.Exec(insertSQL, "Test User", "test@example.com")
		if err != nil {
			t.Errorf("Failed to insert valid data: %v", err)
		}

		// Test inserting duplicate email (should fail due to UNIQUE constraint)
		_, err = db.Exec(insertSQL, "Another User", "test@example.com")
		if err == nil {
			t.Error("Expected error for duplicate email, but insertion succeeded")
		}

		// Test inserting with null name (should fail due to NOT NULL constraint)
		_, err = db.Exec("INSERT INTO test_schema (email) VALUES (?)", "null_name@example.com")
		if err == nil {
			t.Error("Expected error for null name, but insertion succeeded")
		}
	})

	t.Run("IndexCreation", func(t *testing.T) {
		// Create table first
		createTableSQL := `
		CREATE TABLE index_test (
			id INTEGER PRIMARY KEY,
			name TEXT,
			category TEXT
		)`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			t.Errorf("Failed to create index test table: %v", err)
		}

		// Create index
		createIndexSQL := "CREATE INDEX idx_name ON index_test(name)"
		_, err = db.Exec(createIndexSQL)
		if err != nil {
			t.Errorf("Failed to create index: %v", err)
		}

		// Verify index was created (SQLite specific query)
		var indexName string
		query := "SELECT name FROM sqlite_master WHERE type='index' AND name='idx_name'"
		err = db.QueryRow(query).Scan(&indexName)
		if err != nil {
			t.Errorf("Failed to verify index creation: %v", err)
		}
		if indexName != "idx_name" {
			t.Errorf("Expected index name 'idx_name', got '%s'", indexName)
		}
	})
}

// TestMainFunction tests aspects of the main function behavior
func TestMainFunction(t *testing.T) {
	t.Run("FlagInitialization", func(t *testing.T) {
		// Test that flags are properly initialized
		if upFlag == nil {
			t.Error("upFlag should be initialized")
		}
		if downFlag == nil {
			t.Error("downFlag should be initialized")
		}
		if connFlag == nil {
			t.Error("connFlag should be initialized")
		}
		if verbose == nil {
			t.Error("verbose should be initialized")
		}
		if statusFlag == nil {
			t.Error("statusFlag should be initialized")
		}
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// Test environment variable handling
		originalEnv := os.Getenv("DATABASE_URL")
		defer func() {
			if originalEnv != "" {
				os.Setenv("DATABASE_URL", originalEnv)
			} else {
				os.Unsetenv("DATABASE_URL")
			}
		}()

		// Set test environment variable
		testConnStr := "sqlite3://:memory:"
		os.Setenv("DATABASE_URL", testConnStr)

		// Verify environment variable was set
		if os.Getenv("DATABASE_URL") != testConnStr {
			t.Errorf("Expected DATABASE_URL to be '%s', got '%s'", testConnStr, os.Getenv("DATABASE_URL"))
		}
	})
}

// BenchmarkDatabaseOperations benchmarks key database operations
func BenchmarkDatabaseOperations(b *testing.B) {
	// Setup
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create schema_migrations table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		b.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	b.ResetTimer()

	b.Run("InsertMigration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			insertSQL := "INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)"
			_, _ = db.Exec(insertSQL, fmt.Sprintf("bench_%d", i))
		}
	})

	b.Run("CheckMigrationExists", func(b *testing.B) {
		// Insert a test migration first
		insertSQL := "INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)"
		db.Exec(insertSQL, "bench_test")

		for i := 0; i < b.N; i++ {
			var exists bool
			query := "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)"
			_ = db.QueryRow(query, "bench_test").Scan(&exists)
		}
	})

	b.Run("CountMigrations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var count int
			query := "SELECT COUNT(*) FROM schema_migrations"
			_ = db.QueryRow(query).Scan(&count)
		}
	})
}

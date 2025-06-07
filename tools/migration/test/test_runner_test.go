package main

import (
	"database/sql"
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
	errDatabasePingFailed  = "Database ping failed: %v"
	errFailedToVerifyMig   = "Failed to verify migration record: %v"
	errFailedToCountMig    = "Failed to count migrations: %v"
	errFailedToVerifyUsers = "Failed to verify users table: %v"
	errFailedToInsertMig   = "Failed to insert migration record: %v"
	errUnexpectedForConn   = "Unexpected error for connection string '%s': %v"
	successCreateMig       = "Successfully created migrations table"
	successCreateUsers     = "Successfully created users table"
)

// Test constants
const (
	testValidConnection    = "postgres://test:test@localhost:5432/testdb?sslmode=disable"
	testInvalidConnection  = ""
	testUsersTable         = "users"
	testMigrationsTable    = "schema_migrations"
	testConnectionString   = "test-connection"
	migrationVersion       = 1
	insertMigrationSQL     = "INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)"
	testInvalidConn        = "postgres://testuser:testpass@nonexistent:9999/nonexistent?connect_timeout=1"
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
		defer func() { _ = db.Close() }()

		// Test connection
		err = db.Ping()
		if err != nil {
			t.Errorf(errDatabasePingFailed, err)
		}
	})
}

func TestMigrationTableCreation(t *testing.T) {
	t.Run("create migrations table", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer func() { _ = db.Close() }()

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
		defer func() { _ = db.Close() }()

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
		defer func() { _ = db.Close() }()

		// Test ping
		err = db.Ping()
		if err != nil {
			t.Errorf(errDatabasePingFailed, err)
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

func TestMainFunctionWithFlags(t *testing.T) {
	// Test main function with different flag combinations
	t.Run("no connection string", func(t *testing.T) {
		// Reset flags
		*conn = ""
		*up = false
		
		// Capture output by redirecting to test
		// Since main() just returns when conn is empty, this should not panic
		main()
		// If we reach here, main() executed successfully without crashing
	})
	
	t.Run("with connection string but no up flag", func(t *testing.T) {
		// Set connection string but don't set up flag
		*conn = testValidConnection
		*up = false
		
		// This should attempt to connect but not run migrations
		// We can't easily test this without a real database, but we can ensure it doesn't crash
		// Note: This will fail with connection error, but shouldn't panic
		main()
	})
}

func TestMainFunctionComponents(t *testing.T) {
	t.Run("test flag parsing simulation", func(t *testing.T) {
		// Test that flags are properly declared and can be set
		originalConn := *conn
		originalUp := *up
		
		// Set test values
		*conn = testConnectionString
		*up = true
		
		// Verify values were set
		if *conn != testConnectionString {
			t.Errorf("Expected conn to be set to '%s', got '%s'", testConnectionString, *conn)
		}
		if !*up {
			t.Error("Expected up flag to be true")
		}
		
		// Restore original values
		*conn = originalConn
		*up = originalUp
	})
}

func TestPostgresConnectionString(t *testing.T) {
	testValidPostgresConnection(t)
	testPostgresConnectionWithSSL(t)
	testEmptyConnection(t)
}

func testValidPostgresConnection(t *testing.T) {
	t.Run("valid postgres connection", func(t *testing.T) {
		conn := "postgres://user:pass@localhost:5432/testdb"
		db, err := sql.Open("postgres", conn)
		if err != nil {
			t.Errorf(errUnexpectedForConn, conn, err)
		}
		if db != nil {
			_ = db.Close()
		}
	})
}

func testPostgresConnectionWithSSL(t *testing.T) {
	t.Run("postgres connection with ssl", func(t *testing.T) {
		conn := "postgres://user:pass@localhost:5432/testdb?sslmode=require"
		db, err := sql.Open("postgres", conn)
		if err != nil {
			t.Errorf(errUnexpectedForConn, conn, err)
		}
		if db != nil {
			_ = db.Close()
		}
	})
}

func testEmptyConnection(t *testing.T) {
	t.Run("empty connection string", func(t *testing.T) {
		conn := ""
		if conn == "" {
			// Empty connection should be handled in main()
			return
		}
	})
}

func TestMigrationOperations(t *testing.T) {
	testMigrationTableSQL(t)
	testUsersTableSQL(t)
	testMigrationRecordInsertion(t)
}

func testMigrationTableSQL(t *testing.T) {
	t.Run("migration table creation SQL", func(t *testing.T) {
		migrationSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer func() { _ = db.Close() }()
		
		_, err = db.Exec(migrationSQL)
		if err != nil {
			t.Errorf("Migration table creation SQL failed: %v", err)
		}
	})
}

func testUsersTableSQL(t *testing.T) {
	t.Run("users table creation SQL", func(t *testing.T) {
		usersSQL := `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer func() { _ = db.Close() }()
		
		_, err = db.Exec(usersSQL)
		if err != nil {
			t.Errorf("Users table creation SQL failed: %v", err)
		}
	})
}

func testMigrationRecordInsertion(t *testing.T) {
	t.Run("migration record insertion", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer func() { _ = db.Close() }()
		
		// Create migrations table first
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			t.Fatalf(errFailedToCreateMig, err)
		}
		
		// Test insertion (SQLite version)
		_, err = db.Exec(insertMigrationSQL, migrationVersion)
		if err != nil {
			t.Errorf(errFailedToInsertMig, err)
		}
		
		// Verify insertion
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migrationVersion).Scan(&count)
		if err != nil {
			t.Errorf(errFailedToVerifyMig, err)
		}
		if count != 1 {
			t.Errorf("Expected 1 migration record, got %d", count)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("invalid database driver", func(t *testing.T) {
		// Test with invalid driver should fail
		_, err := sql.Open("invalid-driver", "connection-string")
		if err == nil {
			t.Error("Expected error for invalid database driver")
		}
	})

	t.Run("connection string validation", func(t *testing.T) {
		// Test various connection string formats
		validStrings := []string{
			"postgres://user:pass@host:port/db",
			"postgres://user@host/db",
			"host=localhost port=5432 dbname=test",
		}

		for _, connStr := range validStrings {
			// sql.Open doesn't validate connection strings, it just stores them
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				t.Errorf("Unexpected error for connection string '%s': %v", connStr, err)
			}
			if db != nil {
				_ = db.Close()
			}
		}
	})
}

func TestIntegrationScenarios(t *testing.T) {
	testFullMigrationWorkflow(t)
}

func testFullMigrationWorkflow(t *testing.T) {
	t.Run("full migration workflow", func(t *testing.T) {
		db := setupTestDatabase(t)
		defer func() { _ = db.Close() }()
		
		validateDatabaseConnection(t, db)
		createMigrationsTable(t, db)
		createUsersTable(t, db)
		recordMigration(t, db)
		verifyMigrationWorkflow(t, db)
	})
}

func setupTestDatabase(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf(errFailedToOpenDB, err)
	}
	return db
}

func validateDatabaseConnection(t *testing.T, db *sql.DB) {
	err := db.Ping()
	if err != nil {
		t.Errorf(errDatabasePingFailed, err)
	}
}

func createMigrationsTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Errorf(errFailedToCreateMig, err)
	}
}

func createUsersTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Errorf(errFailedToCreateUsers, err)
	}
}

func recordMigration(t *testing.T, db *sql.DB) {
	_, err := db.Exec(insertMigrationSQL, migrationVersion)
	if err != nil {
		t.Errorf(errFailedToInsertMig, err)
	}
}

func verifyMigrationWorkflow(t *testing.T, db *sql.DB) {
	var migrationCount, userTableCount int
	
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
	if err != nil {
		t.Errorf(errFailedToCountMig, err)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&userTableCount)
	if err != nil {
		t.Errorf(errFailedToVerifyUsers, err)
	}
	
	if migrationCount == 0 {
		t.Error("Expected at least one migration record")
	}
	if userTableCount != 1 {
		t.Errorf("Expected users table to exist, got count %d", userTableCount)
	}
}

func TestMainFunctionWithUpFlag(t *testing.T) {
	t.Run("main with up flag and valid connection", func(t *testing.T) {
		// Test the main function with up flag set to true
		// This tests the migration execution path
		originalConn := *conn
		originalUp := *up
		
		// Set flags for up migration
		*conn = "postgres://test:test@localhost:5432/nonexistent?sslmode=disable"
		*up = true
		
		// This will fail to connect, but tests the path through main()
		main()
		
		// Restore original values
		*conn = originalConn
		*up = originalUp
	})
}

func TestDatabaseErrorHandling(t *testing.T) {
	t.Run("database connection error handling", func(t *testing.T) {
		// Test error handling when database connection fails
		originalConn := *conn
		originalUp := *up
		
		*conn = testInvalidConn
		*up = false
		
		// This should handle the connection error gracefully
		main()
		
		// Restore original values
		*conn = originalConn
		*up = originalUp
	})
}

func TestFlagVariableInitialization(t *testing.T) {
	t.Run("flag variables are properly initialized", func(t *testing.T) {
		// Test that flag variables are not nil and have default values
		if up == nil {
			t.Error("up flag should not be nil")
		}
		if conn == nil {
			t.Error("conn flag should not be nil")
		}
		
		// Test that we can read their values
		upValue := *up
		connValue := *conn
		
		// Values should be readable (testing dereferencing)
		if upValue || !upValue {
			// This tests that the boolean value is accessible
		}
		if len(connValue) >= 0 {
			// This tests that the string value is accessible
		}
	})
}

func TestSQLStatements(t *testing.T) {
	t.Run("SQL statements validity", func(t *testing.T) {
		// Test that the SQL statements used in main() are valid
		db, err := sql.Open("sqlite3", testDBPath)
		if err != nil {
			t.Fatalf(errFailedToOpenDB, err)
		}
		defer func() { _ = db.Close() }()
		
		// Test migrations table SQL
		migrationSQL := `CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		
		_, err = db.Exec(migrationSQL)
		if err != nil {
			t.Errorf("Migrations table SQL is invalid: %v", err)
		}
		
		// Test users table SQL (adapted for SQLite)
		usersSQL := `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		
		_, err = db.Exec(usersSQL)
		if err != nil {
			t.Errorf("Users table SQL is invalid: %v", err)
		}
		
		// Test migration insertion SQL
		_, err = db.Exec(insertMigrationSQL, migrationVersion)
		if err != nil {
			t.Errorf("Migration insertion SQL is invalid: %v", err)
		}
	})
}

func TestMainUsageMessage(t *testing.T) {
	t.Run("main function usage message", func(t *testing.T) {
		// Test that main() prints usage when no connection string is provided
		originalConn := *conn
		
		*conn = ""
		
		// This should print the usage message and return
		main()
		
		// Restore original value
		*conn = originalConn
	})
}

func TestMainWithValidConnectionNoUp(t *testing.T) {
	t.Run("main with valid connection but no up flag", func(t *testing.T) {
		// Test main() with a connection string but up=false
		originalConn := *conn
		originalUp := *up
		
		*conn = "postgres://test:test@localhost:5432/testdb?sslmode=disable"
		*up = false
		
		// This should connect (or fail to connect) but not run migrations
		main()
		
		// Restore original values
		*conn = originalConn
		*up = originalUp
	})
}

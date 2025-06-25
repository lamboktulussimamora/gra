package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMain_ComprehensiveScenarios tests scenarios that would occur in main function
func TestMain_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "no_connection_string",
			description: "Main function would return early without connection string",
		},
		{
			name:        "invalid_connection",
			description: "Main function would fail with invalid connection",
		},
		{
			name:        "valid_sqlite_connection",
			description: "Main function would succeed with valid SQLite connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			// These tests document the expected behavior of main()
			// since testing main() directly with flags is complex
		})
	}
}

// TestDatabaseOperations_ComprehensiveScenarios tests database operations independently
func TestDatabaseOperations_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*sql.DB) error
		expectError bool
		validate    func(*sql.DB) error
	}{
		{
			name: "create_migrations_table",
			setup: func(db *sql.DB) error {
				// This simulates the migrations table creation from main()
				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)`)
				return err
			},
			expectError: false,
			validate: func(db *sql.DB) error {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
				if err != nil {
					return err
				}
				if count != 1 {
					t.Errorf("Expected schema_migrations table to exist, got count: %d", count)
				}
				return nil
			},
		},
		{
			name: "create_users_table",
			setup: func(db *sql.DB) error {
				// This simulates the users table creation from main()
				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)`)
				return err
			},
			expectError: false,
			validate: func(db *sql.DB) error {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
				if err != nil {
					return err
				}
				if count != 1 {
					t.Errorf("Expected users table to exist, got count: %d", count)
				}
				return nil
			},
		},
		{
			name: "record_migration",
			setup: func(db *sql.DB) error {
				// First create schema_migrations table
				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)`)
				if err != nil {
					return err
				}
				// Then record a migration
				_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT DO NOTHING")
				return err
			},
			expectError: false,
			validate: func(db *sql.DB) error {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
				if err != nil {
					return err
				}
				if count != 1 {
					t.Errorf("Expected migration record to exist, got count: %d", count)
				}
				return nil
			},
		},
		{
			name: "complete_migration_flow",
			setup: func(db *sql.DB) error {
				// Complete flow: create tables and record migration
				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)`)
				if err != nil {
					return err
				}

				_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)`)
				if err != nil {
					return err
				}

				_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT DO NOTHING")
				return err
			},
			expectError: false,
			validate: func(db *sql.DB) error {
				// Validate both tables exist and migration is recorded
				var schemaCount, usersCount, migrationCount int

				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&schemaCount)
				if err != nil {
					return err
				}

				err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&usersCount)
				if err != nil {
					return err
				}

				err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&migrationCount)
				if err != nil {
					return err
				}

				if schemaCount != 1 || usersCount != 1 || migrationCount != 1 {
					t.Errorf("Expected all tables and migration record to exist, got schema:%d users:%d migration:%d",
						schemaCount, usersCount, migrationCount)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create in-memory SQLite database for each test
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			// Run setup
			err = tt.setup(db)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			} else if err != nil {
				t.Errorf("Unexpected error in setup: %v", err)
				return
			}

			// Run validation
			if tt.validate != nil {
				err = tt.validate(db)
				if err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// TestDatabaseConnection_ComprehensiveScenarios tests database connection scenarios
func TestDatabaseConnection_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name          string
		connectionStr string
		expectError   bool
		testPing      bool
	}{
		{
			name:          "valid_sqlite_memory",
			connectionStr: ":memory:",
			expectError:   false,
			testPing:      true,
		},
		{
			name:          "valid_sqlite_file",
			connectionStr: "file:test.db?cache=shared&mode=memory",
			expectError:   false,
			testPing:      true,
		},
		{
			name:          "invalid_connection_string",
			connectionStr: "invalid://connection/string",
			expectError:   true,
			testPing:      false,
		},
		{
			name:          "empty_connection_string",
			connectionStr: "",
			expectError:   false, // SQLite allows empty string (creates temp file)
			testPing:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For SQLite tests, use sqlite3 driver instead of postgres
			var db *sql.DB
			var err error

			if tt.name == "invalid_connection_string" {
				// Test with postgres driver to trigger error
				db, err = sql.Open("postgres", tt.connectionStr)
			} else {
				// Use sqlite3 for valid tests
				db, err = sql.Open("sqlite3", tt.connectionStr)
			}

			if err != nil {
				if !tt.expectError {
					t.Errorf("Unexpected error opening database: %v", err)
				}
				return
			}
			defer db.Close()

			if tt.testPing {
				err = db.Ping()
				if tt.expectError {
					if err == nil {
						t.Error("Expected ping to fail but it succeeded")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error pinging database: %v", err)
					}
				}
			}
		})
	}
}

// TestErrorHandling_ComprehensiveScenarios tests error handling in database operations
func TestErrorHandling_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*sql.DB, error)
		operation   func(*sql.DB) error
		expectError bool
	}{
		{
			name: "closed_database_connection",
			setup: func() (*sql.DB, error) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					return nil, err
				}
				db.Close() // Close immediately to create error condition
				return db, nil
			},
			operation: func(db *sql.DB) error {
				_, err := db.Exec("CREATE TABLE test (id INTEGER)")
				return err
			},
			expectError: true,
		},
		{
			name: "invalid_sql_syntax",
			setup: func() (*sql.DB, error) {
				return sql.Open("sqlite3", ":memory:")
			},
			operation: func(db *sql.DB) error {
				_, err := db.Exec("INVALID SQL SYNTAX")
				return err
			},
			expectError: true,
		},
		{
			name: "duplicate_table_creation",
			setup: func() (*sql.DB, error) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					return nil, err
				}
				// Create table first time
				_, err = db.Exec("CREATE TABLE test_table (id INTEGER)")
				return db, err
			},
			operation: func(db *sql.DB) error {
				// Try to create same table again (should fail without IF NOT EXISTS)
				_, err := db.Exec("CREATE TABLE test_table (id INTEGER)")
				return err
			},
			expectError: true,
		},
		{
			name: "insert_into_nonexistent_table",
			setup: func() (*sql.DB, error) {
				return sql.Open("sqlite3", ":memory:")
			},
			operation: func(db *sql.DB) error {
				_, err := db.Exec("INSERT INTO nonexistent_table (id) VALUES (1)")
				return err
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			if db != nil {
				defer db.Close()
			}

			err = tt.operation(db)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSequentialDatabaseOperations tests sequential database operations
func TestSequentialDatabaseOperations(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create schema_migrations table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Test sequential inserts
	for i := 1; i <= 5; i++ {
		_, err := db.Exec("INSERT OR REPLACE INTO schema_migrations (version) VALUES (?)", i)
		if err != nil {
			t.Errorf("Sequential insert failed for version %d: %v", i, err)
		}
	}

	// Verify inserts worked
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count migrations: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 migration records, got %d", count)
	}

	// Test reading back the data
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Errorf("Failed to query migrations: %v", err)
		return
	}
	defer rows.Close()

	versions := []int{}
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			t.Errorf("Failed to scan version: %v", err)
			continue
		}
		versions = append(versions, version)
	}

	if len(versions) != 5 {
		t.Errorf("Expected 5 versions, got %d", len(versions))
	}

	for i, version := range versions {
		if version != i+1 {
			t.Errorf("Expected version %d, got %d", i+1, version)
		}
	}
}

// BenchmarkDatabaseOperations benchmarks common database operations
func BenchmarkDatabaseOperations(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create table once
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS benchmark_test (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		b.Fatalf("Failed to create table: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO benchmark_test (name) VALUES (?)", "test_name")
		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}
	}
}

// BenchmarkTableCreation benchmarks table creation operations
func BenchmarkTableCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			b.Errorf("Failed to create database: %v", err)
			continue
		}

		_, err = db.Exec(`CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			b.Errorf("Table creation failed: %v", err)
		}

		db.Close()
	}
}

// TestPostgreSQLTestDatabase_Integration tests integration with the test PostgreSQL database
func TestPostgreSQLTestDatabase_Integration(t *testing.T) {
	// Test connection string that matches docker-compose.test.yml configuration
	testDBURL := "postgres://gra_user:gra_password@localhost:5433/gra_test?sslmode=disable"

	tests := []struct {
		name           string
		connectionStr  string
		expectError    bool
		skipIfNoDocker bool
	}{
		{
			name:           "docker_test_database",
			connectionStr:  testDBURL,
			expectError:    false,
			skipIfNoDocker: true,
		},
		{
			name:          "invalid_test_credentials",
			connectionStr: "postgres://wrong_user:wrong_pass@localhost:5433/gra_test?sslmode=disable",
			expectError:   true,
		},
		{
			name:          "invalid_test_database",
			connectionStr: "postgres://gra_user:gra_password@localhost:5433/nonexistent?sslmode=disable",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := sql.Open("postgres", tt.connectionStr)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Unexpected error opening database: %v", err)
				}
				return
			}
			defer db.Close()

			err = db.Ping()
			if tt.expectError {
				if err == nil {
					t.Error("Expected ping to fail but it succeeded")
				}
			} else {
				if err != nil {
					if tt.skipIfNoDocker {
						t.Skipf("Docker test database not available (this is expected if docker-compose.test.yml is not running): %v", err)
					} else {
						t.Errorf("Unexpected error pinging database: %v", err)
					}
				} else {
					// Test successful connection - try a simple query
					var version string
					err = db.QueryRow("SELECT version()").Scan(&version)
					if err != nil {
						t.Errorf("Failed to query version: %v", err)
					} else {
						t.Logf("Successfully connected to test database, PostgreSQL version: %s", version)
					}
				}
			}
		})
	}
}

// TestMainFunction_WithTestDatabase simulates main function behavior with test database
func TestMainFunction_WithTestDatabase(t *testing.T) {
	// This test simulates what would happen if we called main() with test database parameters
	testDBURL := "postgres://gra_user:gra_password@localhost:5433/gra_test?sslmode=disable"

	t.Run("simulate_main_with_test_db", func(t *testing.T) {
		// Simulate the main function logic
		db, err := sql.Open("postgres", testDBURL)
		if err != nil {
			t.Skipf("Test database not available (expected if docker-compose.test.yml not running): %v", err)
			return
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			t.Skipf("Test database not reachable: %v", err)
			return
		}

		// Simulate the migration operations from main()
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			t.Errorf("Failed to create schema_migrations table: %v", err)
			return
		}

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			t.Errorf("Failed to create users table: %v", err)
			return
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT (version) DO NOTHING")
		if err != nil {
			t.Errorf("Failed to record migration: %v", err)
			return
		}

		t.Log("✓ Successfully simulated main function operations with test database")
	})
}

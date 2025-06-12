package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// TestDatabaseConfig holds database configuration for testing
type TestDatabaseConfig struct {
	Driver    string
	DSN       string
	UseDocker bool
}

// GetTestDatabaseConfig returns the appropriate database configuration for testing
func GetTestDatabaseConfig() TestDatabaseConfig {
	// Check if PostgreSQL testing is requested via environment variable
	if postgresURL := os.Getenv("TEST_DATABASE_URL"); postgresURL != "" {
		return TestDatabaseConfig{
			Driver:    "postgres",
			DSN:       postgresURL,
			UseDocker: true,
		}
	}

	// Default to SQLite for normal testing
	return TestDatabaseConfig{
		Driver:    "sqlite3",
		DSN:       ":memory:",
		UseDocker: false,
	}
}

// SetupTestDatabase creates a test database connection based on configuration
func SetupTestDatabase(t *testing.T) (*sql.DB, func()) {
	config := GetTestDatabaseConfig()

	var db *sql.DB
	var err error
	var cleanup func()

	switch config.Driver {
	case "postgres":
		db, cleanup, err = setupPostgreSQLTest(t, config.DSN)
	case "sqlite3":
		db, cleanup, err = setupSQLiteTest(t)
	default:
		t.Fatalf("Unsupported database driver: %s", config.Driver)
	}

	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		t.Fatalf("Failed to setup test database: %v", err)
	}

	return db, cleanup
}

// setupPostgreSQLTest sets up PostgreSQL for testing
func setupPostgreSQLTest(t *testing.T, dsn string) (*sql.DB, func(), error) {
	t.Logf("Setting up PostgreSQL test database: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Create a unique schema for this test to avoid conflicts
	testSchema := fmt.Sprintf("test_schema_%d", os.Getpid())

	// Create schema
	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", testSchema))
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create test schema: %w", err)
	}

	// Set search path to use test schema
	_, err = db.Exec(fmt.Sprintf("SET search_path TO %s", testSchema))
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to set search path: %w", err)
	}

	cleanup := func() {
		// Drop test schema before closing connection
		_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema))
		if err != nil {
			t.Logf("Warning: failed to drop test schema: %v", err)
		}
		db.Close()
	}

	return db, cleanup, nil
}

// setupSQLiteTest sets up SQLite for testing
func setupSQLiteTest(t *testing.T) (*sql.DB, func(), error) {
	// Create temp directory for SQLite database
	tmpDir, err := os.MkdirTemp("", "migration_test_")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, nil, fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		return nil, nil, fmt.Errorf("failed to ping SQLite: %w", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup, nil
}

// SetupAutoMigrationTestWithDB creates an AutoMigrator with the provided database
func SetupAutoMigrationTestWithDB(t *testing.T, db *sql.DB) *AutoMigrator {
	// Create enhanced db context
	ctx := &dbcontext.EnhancedDbContext{} // Simplified for testing

	// Create auto migrator
	migrator := NewAutoMigrator(ctx, db)

	return migrator
}

// SetupAutoMigrationTestMultiDB sets up auto migration test with appropriate database
func SetupAutoMigrationTestMultiDB(t *testing.T) (*AutoMigrator, *sql.DB, func()) {
	db, cleanup := SetupTestDatabase(t)
	migrator := SetupAutoMigrationTestWithDB(t, db)

	return migrator, db, cleanup
}

// PostgreSQLIntegrationTest runs a test function with PostgreSQL if available
func PostgreSQLIntegrationTest(t *testing.T, testFunc func(*testing.T, *sql.DB)) {
	if os.Getenv("TEST_WITH_POSTGRES") == "" {
		t.Skip("Skipping PostgreSQL integration test (set TEST_WITH_POSTGRES=true to enable)")
	}

	config := GetTestDatabaseConfig()
	if config.Driver != "postgres" {
		t.Skip("PostgreSQL not configured for testing")
	}

	db, cleanup := SetupTestDatabase(t)
	defer cleanup()

	testFunc(t, db)
}

// DatabaseDriverSpecificTest runs tests for specific database drivers
func DatabaseDriverSpecificTest(t *testing.T, testFunc func(*testing.T, string, *sql.DB)) {
	// Test with SQLite
	t.Run("SQLite", func(t *testing.T) {
		db, cleanup, err := setupSQLiteTest(t)
		if err != nil {
			t.Fatalf("Failed to setup SQLite test: %v", err)
		}
		defer cleanup()
		testFunc(t, "sqlite3", db)
	})

	// Test with PostgreSQL if available
	if os.Getenv("TEST_WITH_POSTGRES") != "" {
		t.Run("PostgreSQL", func(t *testing.T) {
			if config := GetTestDatabaseConfig(); config.Driver == "postgres" {
				db, cleanup, err := setupPostgreSQLTest(t, config.DSN)
				if err != nil {
					t.Skipf("PostgreSQL not available: %v", err)
				}
				defer cleanup()
				testFunc(t, "postgres", db)
			}
		})
	}
}

// CheckTableExists checks if a table exists in a database-agnostic way
func CheckTableExists(db *sql.DB, tableName string) (bool, error) {
	// Try PostgreSQL first
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
			 WHERE table_schema = current_schema() AND table_name = $1`
	err := db.QueryRow(query, tableName).Scan(&count)
	if err == nil {
		return count > 0, nil
	}

	// If PostgreSQL query fails, try SQLite
	query = "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	err = db.QueryRow(query, tableName).Scan(&count)
	if err == nil {
		return count > 0, nil
	}

	// If both fail, try MySQL
	query = `SELECT COUNT(*) FROM information_schema.tables 
			 WHERE table_schema = DATABASE() AND table_name = ?`
	err = db.QueryRow(query, tableName).Scan(&count)
	if err == nil {
		return count > 0, nil
	}

	return false, fmt.Errorf("unable to check table existence with any supported database driver")
}

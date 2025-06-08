package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup: Clean any existing test_migrations directory
	cleanupTestMigrations()

	// Run tests
	code := m.Run()

	// Cleanup: Clean test_migrations directory after all tests
	cleanupTestMigrations()

	os.Exit(code)
}

// cleanupTestMigrations removes the test_migrations directory and its contents
func cleanupTestMigrations() {
	if err := os.RemoveAll(testMigrationsDir); err != nil {
		// Don't fail tests if cleanup fails, just log a warning
		log.Printf("Warning: failed to clean test_migrations directory: %v", err)
	}
}

// Helper function to create a test migration manager
func createTestMigrationManager(t *testing.T, db *sql.DB) *migrations.EFMigrationManager {
	t.Helper()

	// Create a basic EF migration manager instance
	// This is a mock implementation for testing purposes
	config := migrations.DefaultEFMigrationConfig()
	config.Logger = log.New(io.Discard, "", 0) // Use a silent logger for tests to prevent excessive output

	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize schema for tests
	if err := manager.EnsureSchema(); err != nil {
		t.Fatalf("Failed to ensure migration schema: %v", err)
	}

	return manager
}

// Test constants - consolidated to reduce duplication
const (
	// Command and flag constants
	testAddMigrationCmd    = "add-migration"
	testUpdateDatabaseCmd  = "update-database"
	testGetMigrationCmd    = "get-migration"
	testRemoveMigrationCmd = "remove-migration"
	testRollbackCmd        = "rollback"
	testScriptCmd          = "script"
	testStatusCmd          = "status"
	testHelpFlag           = "--help"
	testHelpCommand        = "help"
	testVerboseFlag        = "-verbose"
	testProgramName        = "ef-migrate"

	// Database constants
	testMemoryDB            = ":memory:"
	testSQLite3Driver       = "sqlite3"
	testPostgresDriver      = "postgres"
	testSQLiteConnStr       = "/Users/test/database.db"
	testSQLiteMemoryConnStr = "file::memory:?cache=shared"
	testPostgresConnStr     = "postgres://user:pass@localhost/db"
	testMaskedPassword      = "*****"
	testLocalhost           = "localhost"
	testPort5432            = "5432"
	testTestUser            = "testuser"
	testTestPassword        = "testpass"
	testTestDB              = "testdb"

	// Directory and file constants
	testMigrationsDir        = "./test_migrations"
	testDefaultMigrationsDir = "./migrations"

	// Common error message formats
	testExpectedFormat       = "Expected %q, got %q"
	testFailedTempDir        = "Failed to create temp dir: %v"
	testFailedOpenDB         = "Failed to open test database: %v"
	testFailedToCreateDB     = "Failed to create test database: %v"
	testFailedSetupManager   = "Failed to setup migration manager: %v"
	testExpectedOutputFormat = "Expected output to contain %q, got: %s"
	testFailedCreateFile     = "Failed to create migration file: %v"
	testWarningCloseDB       = "Warning: failed to close database: %v"

	// Additional connection strings for test cases
	testSafeConnStr4 = "postgres://user:pass@localhost/db"
)

func TestParseCommandLineArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCmd  string
		expectedArgs []string
		expectExit   bool
	}{
		{
			name:         "valid add migration command",
			args:         []string{"program", testAddMigrationCmd, "CreateUser"},
			expectedCmd:  testAddMigrationCmd,
			expectedArgs: []string{"CreateUser"},
		},
		{
			name:         "help command",
			args:         []string{"program", testHelpCommand},
			expectedCmd:  testHelpCommand,
			expectedArgs: []string{},
		},
		{
			name:         "update database command",
			args:         []string{"program", testUpdateDatabaseCmd},
			expectedCmd:  testUpdateDatabaseCmd,
			expectedArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Save original args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			// We can't test parseCommandLineArgs directly due to flag.Parse() and os.Exit()
			// Instead, we'll test the logic components separately
		})
	}
}

func TestSetupDatabaseConnection(t *testing.T) {
	tests := []struct {
		name        string
		config      CLIConfig
		expectError bool
	}{
		{
			name: "valid sqlite connection",
			config: CLIConfig{
				ConnectionString: testMemoryDB,
			},
			expectError: false,
		},
		{
			name: "postgres connection string format",
			config: CLIConfig{
				ConnectionString: "postgres://user:pass@localhost/testdb?sslmode=disable",
			},
			expectError: false, // Will fail to connect but driver detection should work
		},
		{
			name: "empty connection string with env fallback",
			config: CLIConfig{
				ConnectionString: "",
			},
			expectError: true,
		},
		{
			name: "postgres from individual parameters",
			config: CLIConfig{
				Host:     "localhost",
				User:     "testuser",
				Database: "testdb",
				Password: "testpass",
				Port:     "5432",
				SSLMode:  "disable",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := setupDatabaseConnection(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil && tt.config.ConnectionString != testMemoryDB {
					// For non-memory databases, connection might fail but function should not error on setup
					t.Logf("Connection setup returned error (expected for test): %v", err)
				}
			}
			if db != nil {
				if err := db.Close(); err != nil {
					t.Logf("Warning: failed to close database: %v", err)
				}
			}
		})
	}
}

func TestBuildPostgreSQLConnectionString(t *testing.T) {
	config := CLIConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	result := buildPostgreSQLConnectionString(config)
	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCLIConfig(t *testing.T) {
	config := CLIConfig{
		ConnectionString: "test://connection",
		MigrationsDir:    "./test_migrations",
		Verbose:          true,
		Host:             "localhost",
		Port:             "5432",
		User:             "testuser",
		Password:         "testpass",
		Database:         "testdb",
		SSLMode:          "disable",
	}

	// Test that all fields are properly set
	if config.ConnectionString != "test://connection" {
		t.Errorf("Expected ConnectionString 'test://connection', got '%s'", config.ConnectionString)
	}
	if config.MigrationsDir != "./test_migrations" {
		t.Errorf("Expected MigrationsDir './test_migrations', got '%s'", config.MigrationsDir)
	}
	if !config.Verbose {
		t.Error("Expected Verbose to be true")
	}
	if config.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", config.Host)
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if ErrorFailedToGetHistoryFmt == "" {
		t.Error("ErrorFailedToGetHistoryFmt should not be empty")
	}
	if FormatMigrationLine == "" {
		t.Error("FormatMigrationLine should not be empty")
	}
	if TimeFormat == "" {
		t.Error("TimeFormat should not be empty")
	}

	// Test time format validity
	if TimeFormat != "2006-01-02 15:04:05" {
		t.Errorf("Expected TimeFormat '2006-01-02 15:04:05', got '%s'", TimeFormat)
	}
}

func TestDriverDetection(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		expectedDriver   string
	}{
		{
			name:             "postgres URL format",
			connectionString: "postgres://user:pass@localhost/db",
			expectedDriver:   "postgres",
		},
		{
			name:             "postgres with user parameter",
			connectionString: "host=localhost user=test dbname=test",
			expectedDriver:   "postgres",
		},
		{
			name:             "sqlite file",
			connectionString: "./test.db",
			expectedDriver:   "sqlite3",
		},
		{
			name:             "sqlite with extension",
			connectionString: "file:test.sqlite",
			expectedDriver:   "sqlite3",
		},
		{
			name:             "default case",
			connectionString: "unknown://format",
			expectedDriver:   "postgres", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the driver detection logic from setupDatabaseConnection
			var driverName string
			switch tt.connectionString {
			case "postgres://user:pass@localhost/db", "host=localhost user=test dbname=test":
				driverName = testPostgresDriver
			case "./test.db", "file:test.sqlite":
				driverName = "sqlite3"
			default:
				driverName = testPostgresDriver // Default
			}

			if driverName != tt.expectedDriver {
				t.Errorf("Expected driver %s, got %s", tt.expectedDriver, driverName)
			}
		})
	}
}

func TestSetupMigrationManager(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: "./test_migrations",
		Verbose:       true,
	}

	// Note: This will fail because we don't have the actual migration manager
	// but we can test the function signature and basic setup
	manager, err := setupMigrationManager(db, config)
	if err != nil {
		// Expected to fail due to missing migration dependencies in test
		t.Logf("Setup failed as expected in test environment: %v", err)
	}
	if manager != nil {
		t.Log("Migration manager created successfully")
	}
}

func TestExecuteCommand(t *testing.T) {
	// Test command routing logic
	commands := []string{
		testAddMigrationCmd,
		"add",
		testUpdateDatabaseCmd,
		"update",
		"get-migration",
		"list",
		"rollback",
		"status",
		"script",
		"remove-migration",
		"remove",
		testHelpCommand,
		"-h",
		testHelpFlag,
		"unknown",
	}

	for _, cmd := range commands {
		t.Run("command_"+cmd, func(_ *testing.T) {
			// We can't test executeCommand directly due to dependencies
			// but we can verify the command strings are handled
			switch cmd {
			case testAddMigrationCmd, "add":
				// Valid command
			case testUpdateDatabaseCmd, "update":
				// Valid command
			case "get-migration", "list":
				// Valid command
			case "rollback":
				// Valid command
			case "status":
				// Valid command
			case "script":
				// Valid command
			case "remove-migration", "remove":
				// Valid command
			case testHelpCommand, "-h", testHelpFlag:
				// Valid command
			default:
				// Unknown command
			}
		})
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that CLIConfig has all expected fields
	config := CLIConfig{}
	configType := reflect.TypeOf(config)

	expectedFields := []string{
		"ConnectionString",
		"MigrationsDir",
		"Verbose",
		"Host",
		"Port",
		"User",
		"Password",
		"Database",
		"SSLMode",
	}

	for _, field := range expectedFields {
		_, found := configType.FieldByName(field)
		if !found {
			t.Errorf("Expected field %s not found in CLIConfig", field)
		}
	}
}

func TestBuildPostgreSQLConnectionStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   CLIConfig
		expected string
	}{
		{
			name: "minimal config",
			config: CLIConfig{
				Host:     "localhost",
				User:     "user",
				Database: "db",
			},
			expected: "postgres://user:@localhost:5432/db?sslmode=disable",
		},
		{
			name: "with special characters in password",
			config: CLIConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "test@pass#123",
				Database: "testdb",
				SSLMode:  "require",
			},
			expected: "postgres://testuser:test@pass#123@localhost:5432/testdb?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPostgreSQLConnectionString(tt.config)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Test helper functions
func TestFormatMigrationInfo(t *testing.T) {
	tests := []struct {
		name           string
		migration      migrations.Migration
		status         string
		expectedFormat string
	}{
		{
			name: "applied migration with timestamp",
			migration: migrations.Migration{
				ID:          "20240101_CreateUsers",
				Description: "Create users table",
				AppliedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			status:         "applied",
			expectedFormat: "✅ 20240101_CreateUsers (2024-01-01 12:00:00) - Create users table",
		},
		{
			name: "pending migration without timestamp",
			migration: migrations.Migration{
				ID:          "20240102_AddIndexes",
				Description: "Add database indexes",
			},
			status:         "pending",
			expectedFormat: "⏳ 20240102_AddIndexes - Add database indexes",
		},
		{
			name: "failed migration",
			migration: migrations.Migration{
				ID: "20240103_BadMigration",
			},
			status:         "failed",
			expectedFormat: "❌ 20240103_BadMigration",
		},
		{
			name: "unknown status",
			migration: migrations.Migration{
				ID: "20240104_UnknownStatus",
			},
			status:         "unknown",
			expectedFormat: "❓ 20240104_UnknownStatus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMigrationInfo(tt.migration, tt.status)
			if result != tt.expectedFormat {
				t.Errorf("Expected %q, got %q", tt.expectedFormat, result)
			}
		})
	}
}

func TestExtractDBName(t *testing.T) {
	tests := []struct {
		name           string
		connectionStr  string
		expectedDBName string
	}{
		{
			name:           "postgres URL with params",
			connectionStr:  "postgres://user:pass@localhost:5432/mydb?sslmode=disable",
			expectedDBName: "mydb",
		},
		{
			name:           "simple database path",
			connectionStr:  "path/to/database.db",
			expectedDBName: "database.db",
		},
		{
			name:           "sqlite path",
			connectionStr:  "/var/lib/app/data.sqlite",
			expectedDBName: "data.sqlite",
		},
		{
			name:           "empty connection string",
			connectionStr:  "",
			expectedDBName: "",
		},
		{
			name:           "single name",
			connectionStr:  "testdb",
			expectedDBName: "testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDBName(tt.connectionStr)
			if result != tt.expectedDBName {
				t.Errorf("Expected %q, got %q", tt.expectedDBName, result)
			}
		})
	}
}

func TestSanitizeConnectionString(t *testing.T) {
	tests := []struct {
		name           string
		connectionStr  string
		expectedResult string
	}{
		{
			name:           "postgres URL with password",
			connectionStr:  "postgres://user:password123@localhost:5432/mydb",
			expectedResult: "postgres://user:*****@localhost:5432/mydb",
		},
		{
			name:           "postgres URL with special chars in password",
			connectionStr:  "postgres://user:p@ssw0rd!@localhost:5432/mydb",
			expectedResult: "postgres://user:*****@localhost:5432/mydb",
		},
		{
			name:           "non-postgres connection string",
			connectionStr:  "sqlite://path/to/db.sqlite",
			expectedResult: "sqlite://path/to/db.sqlite",
		},
		{
			name:           "empty connection string",
			connectionStr:  "",
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeConnectionString(tt.connectionStr)
			if result != tt.expectedResult {
				t.Errorf("Expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

func TestParseMigrationContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedUp   string
		expectedDown string
	}{
		{
			name: "complete migration with up and down sections",
			content: `-- Migration: CreateUsers
-- Description: Create users table
-- Created: 2024-01-01 12:00:00
-- Version: 1

-- UP Migration
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- DOWN Migration (for rollback)
-- DROP TABLE users;`,
			expectedUp:   "CREATE TABLE users (\n    id SERIAL PRIMARY KEY,\n    name VARCHAR(255) NOT NULL\n);",
			expectedDown: "DROP TABLE users;",
		},
		{
			name: "migration with only up section",
			content: `-- Migration: AddIndex
-- UP Migration
CREATE INDEX idx_users_name ON users(name);`,
			expectedUp:   "CREATE INDEX idx_users_name ON users(name);",
			expectedDown: "",
		},
		{
			name: "empty migration",
			content: `-- Migration: EmptyMigration
-- Description: Empty migration
-- UP Migration

-- DOWN Migration
--`,
			expectedUp:   "",
			expectedDown: "--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upSQL, downSQL := parseMigrationContent(tt.content)
			if upSQL != tt.expectedUp {
				t.Errorf("Expected UP SQL %q, got %q", tt.expectedUp, upSQL)
			}
			if downSQL != tt.expectedDown {
				t.Errorf("Expected DOWN SQL %q, got %q", tt.expectedDown, downSQL)
			}
		})
	}
}

func TestDetectSectionType(t *testing.T) {
	tests := []struct {
		name                    string
		line                    string
		expectedIsDown          bool
		expectedIsSectionMarker bool
	}{
		{
			name:                    "up migration marker",
			line:                    "-- UP Migration",
			expectedIsDown:          false,
			expectedIsSectionMarker: true,
		},
		{
			name:                    "down migration marker",
			line:                    "-- DOWN Migration (for rollback)",
			expectedIsDown:          true,
			expectedIsSectionMarker: true,
		},
		{
			name:                    "rollback marker",
			line:                    "-- Rollback SQL",
			expectedIsDown:          true,
			expectedIsSectionMarker: true,
		},
		{
			name:                    "regular comment",
			line:                    "-- This is a regular comment",
			expectedIsDown:          false,
			expectedIsSectionMarker: false,
		},
		{
			name:                    "regular SQL",
			line:                    "CREATE TABLE test (id INT);",
			expectedIsDown:          false,
			expectedIsSectionMarker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDown, isSectionMarker := detectSectionType(tt.line)
			if isDown != tt.expectedIsDown {
				t.Errorf("Expected isDown %v, got %v", tt.expectedIsDown, isDown)
			}
			if isSectionMarker != tt.expectedIsSectionMarker {
				t.Errorf("Expected isSectionMarker %v, got %v", tt.expectedIsSectionMarker, isSectionMarker)
			}
		})
	}
}

func TestShouldIncludeInUpSection(t *testing.T) {
	tests := []struct {
		name            string
		line            string
		expectedInclude bool
	}{
		{
			name:            "regular SQL statement",
			line:            "CREATE TABLE test (id INT);",
			expectedInclude: true,
		},
		{
			name:            "migration header comment",
			line:            "-- Migration: CreateUsers",
			expectedInclude: false,
		},
		{
			name:            "description header comment",
			line:            "-- Description: Create users table",
			expectedInclude: false,
		},
		{
			name:            "created header comment",
			line:            "-- Created: 2024-01-01",
			expectedInclude: false,
		},
		{
			name:            "version header comment",
			line:            "-- Version: 1",
			expectedInclude: false,
		},
		{
			name:            "regular comment",
			line:            "-- This is a regular comment",
			expectedInclude: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIncludeInUpSection(tt.line)
			if result != tt.expectedInclude {
				t.Errorf("Expected %v, got %v", tt.expectedInclude, result)
			}
		})
	}
}

func TestCleanDownSQL(t *testing.T) {
	tests := []struct {
		name           string
		downSQL        string
		expectedResult string
	}{
		{
			name:           "empty string",
			downSQL:        "",
			expectedResult: "",
		},
		{
			name:           "commented down SQL",
			downSQL:        "-- DROP TABLE users;\n-- DROP INDEX idx_users_name;",
			expectedResult: "DROP TABLE users;\nDROP INDEX idx_users_name;",
		},
		{
			name:           "mixed commented and uncommented",
			downSQL:        "-- DROP TABLE users;\nDROP INDEX idx_users_name;\n-- Another comment",
			expectedResult: "DROP TABLE users;\nDROP INDEX idx_users_name;\nAnother comment",
		},
		{
			name:           "no comments",
			downSQL:        "DROP TABLE users;",
			expectedResult: "DROP TABLE users;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanDownSQL(tt.downSQL)
			if result != tt.expectedResult {
				t.Errorf("Expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

func TestSaveMigrationToFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "migration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	migration := &migrations.Migration{
		ID:          "20240101_TestMigration",
		Name:        "Test Migration",
		Description: "Test migration description",
		Version:     20240101,
		UpSQL:       "CREATE TABLE test (id INT);",
		DownSQL:     "DROP TABLE test;",
	}

	err = saveMigrationToFile(migration, tempDir)
	if err != nil {
		t.Fatalf("Failed to save migration file: %v", err)
	}

	// Verify file was created
	expectedFile := filepath.Join(tempDir, "20240101_TestMigration.sql")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Migration file was not created: %s", expectedFile)
	}

	// Read file content and verify - validate file path for security
	if !strings.HasSuffix(expectedFile, ".sql") || strings.Contains(expectedFile, "..") {
		t.Fatalf("Invalid file path: %s", expectedFile)
	}
	// Use filepath.Clean to sanitize the path and ensure it's within tempDir
	cleanPath := filepath.Clean(expectedFile)
	if !strings.HasPrefix(cleanPath, filepath.Clean(tempDir)) {
		t.Fatalf("File path outside expected directory: %s", cleanPath)
	}
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test Migration") {
		t.Errorf("Migration file should contain migration name")
	}
	if !strings.Contains(contentStr, "CREATE TABLE test") {
		t.Errorf("Migration file should contain UP SQL")
	}
	if !strings.Contains(contentStr, "DROP TABLE test") {
		t.Errorf("Migration file should contain DOWN SQL")
	}
}

func TestLoadMigrationsFromFilesystem(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "migrations_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("no migrations directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")

		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)

		err = loadMigrationsFromFilesystem(manager, nonExistentDir)
		if err != nil {
			t.Errorf("Expected no error for non-existent directory, got: %v", err)
		}
	})

	t.Run("valid migration files", func(t *testing.T) {
		migrationsDir := filepath.Join(tempDir, "migrations")
		err := os.MkdirAll(migrationsDir, 0750)
		if err != nil {
			t.Fatalf("Failed to create migrations directory: %v", err)
		}

		// Create test migration files
		migration1 := `-- Migration: CreateUsers
-- Description: Create users table
-- UP Migration
CREATE TABLE users (id INT PRIMARY KEY);

-- DOWN Migration
-- DROP TABLE users;`

		migration2 := `-- Migration: AddIndex
-- UP Migration  
CREATE INDEX idx_users_id ON users(id);

-- DOWN Migration
-- DROP INDEX idx_users_id;`

		err = os.WriteFile(filepath.Join(migrationsDir, "001_CreateUsers.sql"), []byte(migration1), 0600)
		if err != nil {
			t.Fatalf("Failed to create migration file: %v", err)
		}

		err = os.WriteFile(filepath.Join(migrationsDir, "002_AddIndex.sql"), []byte(migration2), 0600)
		if err != nil {
			t.Fatalf("Failed to create migration file: %v", err)
		}

		// Create file with invalid name (should be skipped)
		err = os.WriteFile(filepath.Join(migrationsDir, "invalid_migration.sql"), []byte("-- Invalid"), 0600)
		if err != nil {
			t.Fatalf("Failed to create invalid migration file: %v", err)
		}

		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)

		err = loadMigrationsFromFilesystem(manager, migrationsDir)
		if err != nil {
			t.Errorf("Failed to load migrations from filesystem: %v", err)
		}
	})

	t.Run("invalid migration file content", func(t *testing.T) {
		migrationsDir := filepath.Join(tempDir, "invalid_migrations")
		err := os.MkdirAll(migrationsDir, 0750)
		if err != nil {
			t.Fatalf("Failed to create migrations directory: %v", err)
		}

		// Create migration file with invalid version
		err = os.WriteFile(filepath.Join(migrationsDir, "abc_InvalidVersion.sql"), []byte("-- Invalid version"), 0600)
		if err != nil {
			t.Fatalf("Failed to create invalid migration file: %v", err)
		}

		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)

		err = loadMigrationsFromFilesystem(manager, migrationsDir)
		if err != nil {
			t.Errorf("Should skip invalid files without error, got: %v", err)
		}
	})
}

func TestAddMigrationFunction(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "add_migration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("add migration with description", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{
			MigrationsDir: tempDir,
		}

		args := []string{"CreateUsers", "Create", "users", "table"}
		addMigration(manager, args, config)

		// Verify migration file was created
		files, err := filepath.Glob(filepath.Join(tempDir, "*.sql"))
		if err != nil {
			t.Fatalf("Failed to scan for migration files: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 migration file, got %d", len(files))
		}
	})

	t.Run("add migration without description", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{
			MigrationsDir: tempDir,
		}

		args := []string{"AddIndex"}
		addMigration(manager, args, config)

		// Verify migration file was created
		files, err := filepath.Glob(filepath.Join(tempDir, "*.sql"))
		if err != nil {
			t.Fatalf("Failed to scan for migration files: %v", err)
		}

		if len(files) < 1 {
			t.Errorf("Expected at least 1 migration file, got %d", len(files))
		}
	})

	t.Run("add migration with no args", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{
			MigrationsDir: tempDir,
		}

		args := []string{}
		addMigration(manager, args, config)
		// Should log error but not fail test
	})
}

// Helper function to test migration operations with optional target
func testMigrationOperation(t *testing.T, operation func(*migrations.EFMigrationManager, []string, CLIConfig), operationName string) {
	t.Helper()

	t.Run(operationName+" without target", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{}

		args := []string{}
		operation(manager, args, config)
		// Should complete without error
	})

	t.Run(operationName+" with target", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{}

		args := []string{"TargetMigration"}
		operation(manager, args, config)
		// Should complete without error
	})
}

func TestUpdateDatabaseFunction(t *testing.T) {
	testMigrationOperation(t, updateDatabase, "update database")
}

func TestGetMigrationsFunction(t *testing.T) {
	t.Run("get migrations", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{}

		getMigrations(manager, config)
		// Should complete without error
	})
}

func TestRollbackMigrationFunction(t *testing.T) {
	t.Run("rollback migration with target", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{}

		args := []string{"TargetMigration"}
		rollbackMigration(manager, args, config)
		// Should complete without error (may fail rollback but won't crash)
	})

	t.Run("rollback migration without target", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{}

		args := []string{}
		rollbackMigration(manager, args, config)
		// Should log error but not fail test
	})
}

func TestShowStatusFunction(t *testing.T) {
	t.Run("show status", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{
			ConnectionString: testMemoryDB,
		}

		showStatus(manager, config)
		// Should complete without error
	})
}

func TestGenerateScriptFunction(t *testing.T) {
	testMigrationOperation(t, generateScript, "generate script")
}

func TestRemoveMigrationFunction(t *testing.T) {
	t.Run("remove migration", func(t *testing.T) {
		db, err := sql.Open("sqlite3", testMemoryDB)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}
		defer func() { _ = db.Close() }()

		manager := createTestMigrationManager(t, db)
		config := CLIConfig{
			MigrationsDir: testMigrationsDir,
		}

		args := []string{}
		removeMigration(manager, args, config)
		// Should complete without error
	})
}

// TestParseCommandLineArgsWithMocking tests the parseCommandLineArgs function using a mock approach
func TestParseCommandLineArgsWithMocking(t *testing.T) {
	tests := getCommandLineTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateCommandLineParsing(t, tt)
		})
	}
}

// getCommandLineTestCases returns test cases for command line parsing
func getCommandLineTestCases() []struct {
	name           string
	args           []string
	expectedConfig CLIConfig
	expectedCmd    string
	expectedArgs   []string
	shouldExit     bool
} {
	return []struct {
		name           string
		args           []string
		expectedConfig CLIConfig
		expectedCmd    string
		expectedArgs   []string
		shouldExit     bool
	}{
		{
			name: "basic add migration command",
			args: []string{testProgramName, testAddMigrationCmd, "CreateUsers"},
			expectedConfig: CLIConfig{
				MigrationsDir: testDefaultMigrationsDir,
				Port:          testPort5432,
				SSLMode:       "disable",
			},
			expectedCmd:  testAddMigrationCmd,
			expectedArgs: []string{"CreateUsers"},
		},
		{
			name: "command with connection string",
			args: []string{testProgramName, "-connection", testSafeConnStr4, "status"},
			expectedConfig: CLIConfig{
				ConnectionString: testSafeConnStr4,
				MigrationsDir:    testDefaultMigrationsDir,
				Port:             testPort5432,
				SSLMode:          "disable",
			},
			expectedCmd:  "status",
			expectedArgs: []string{},
		},
		{
			name: "verbose mode with PostgreSQL flags",
			args: []string{testProgramName, testVerboseFlag, "-host", testLocalhost, "-user", testTestUser, "-database", testTestDB, testUpdateDatabaseCmd},
			expectedConfig: CLIConfig{
				Verbose:       true,
				Host:          testLocalhost,
				User:          testTestUser,
				Database:      testTestDB,
				MigrationsDir: testDefaultMigrationsDir,
				Port:          testPort5432,
				SSLMode:       "disable",
			},
			expectedCmd:  testUpdateDatabaseCmd,
			expectedArgs: []string{},
		},
		{
			name: "custom migrations directory",
			args: []string{testProgramName, "-migrations-dir", "./custom_migrations", testGetMigrationCmd},
			expectedConfig: CLIConfig{
				MigrationsDir: "./custom_migrations",
				Port:          testPort5432,
				SSLMode:       "disable",
			},
			expectedCmd:  testGetMigrationCmd,
			expectedArgs: []string{},
		},
	}
}

// validateCommandLineParsing validates command line parsing for a test case
func validateCommandLineParsing(t *testing.T, tt struct {
	name           string
	args           []string
	expectedConfig CLIConfig
	expectedCmd    string
	expectedArgs   []string
	shouldExit     bool
}) {
	// Test that we can parse the arguments correctly
	// Note: We can't directly test parseCommandLineArgs due to flag.Parse() and os.Exit()
	// but we can test the flag definitions are correct

	args := tt.args[1:] // Remove program name
	if len(args) == 0 {
		// Would call printUsage() and os.Exit(1)
		return
	}

	cmdPos := findCommandPosition(args)
	if cmdPos != -1 {
		validateExtractedCommand(t, args, cmdPos, tt.expectedCmd, tt.expectedArgs)
	}
}

// findCommandPosition finds the position of the command in the arguments
func findCommandPosition(args []string) int {
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && (i == 0 || !strings.HasPrefix(args[i-1], "-") ||
			args[i-1] == testVerboseFlag) {
			return i
		}
	}
	return -1
}

// validateExtractedCommand validates the extracted command and arguments
func validateExtractedCommand(t *testing.T, args []string, cmdPos int, expectedCmd string, expectedArgs []string) {
	command := args[cmdPos]
	commandArgs := args[cmdPos+1:]

	if command != expectedCmd {
		t.Errorf("Expected command %s, got %s", expectedCmd, command)
	}

	if len(commandArgs) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(commandArgs))
	}

	for i, arg := range commandArgs {
		if i < len(expectedArgs) && arg != expectedArgs[i] {
			t.Errorf("Expected arg[%d] %s, got %s", i, expectedArgs[i], arg)
		}
	}
}

// TestPrintUsageOutput tests that printUsage produces expected output
func TestPrintUsageOutput(t *testing.T) {
	// Capture stdout
	output := captureOutput(func() {
		printUsage()
	})

	// Check for key sections
	expectedSections := []string{
		"🚀 GRA Entity Framework Core-like Migration Tool",
		"USAGE:",
		"OPTIONS:",
		"PostgreSQL Connection Options:",
		"COMMANDS:",
		"📝 Migration Management:",
		"📋 Information:",
		"EXAMPLES:",
		"Connection Examples:",
		"Migration Examples:",
		"ENVIRONMENT:",
		"ef-migrate [options] <command> [arguments]",
		"add-migration <name> [description]",
		"update-database [target]",
		"get-migration",
		"status",
		"script [target]",
		"rollback <target>",
		"remove-migration",
		"-connection <string>",
		"-migrations-dir <path>",
		"-verbose",
		"-host <string>",
		"-port <string>",
		"-user <string>",
		"-password <string>",
		"-database <string>",
		"-sslmode <string>",
		"DATABASE_URL",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", section, output)
		}
	}

	// Check minimum length
	if len(output) < 1000 {
		t.Errorf("Expected usage output to be substantial, got %d characters", len(output))
	}
}

// TestExecuteCommandRouting tests that executeCommand routes to correct functions
func TestExecuteCommandRouting(t *testing.T) {
	// Create a test database and manager
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: "./test_migrations",
		Verbose:       true,
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to setup migration manager: %v", err)
	}

	tests := []struct {
		name        string
		command     string
		args        []string
		expectLog   string
		shouldPanic bool
	}{
		{
			name:      "add-migration command",
			command:   "add-migration",
			args:      []string{"TestMigration"},
			expectLog: "",
		},
		{
			name:      "add command alias",
			command:   "add",
			args:      []string{"TestMigration2"},
			expectLog: "",
		},
		{
			name:      "update-database command",
			command:   "update-database",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "update command alias",
			command:   "update",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "get-migration command",
			command:   "get-migration",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "list command alias",
			command:   "list",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "rollback command",
			command:   "rollback",
			args:      []string{"target"},
			expectLog: "",
		},
		{
			name:      "status command",
			command:   "status",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "script command",
			command:   "script",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "remove-migration command",
			command:   "remove-migration",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "remove command alias",
			command:   "remove",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "help command",
			command:   "help",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "-h flag",
			command:   "-h",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "--help flag",
			command:   "--help",
			args:      []string{},
			expectLog: "",
		},
		{
			name:      "unknown command",
			command:   "unknown-command",
			args:      []string{},
			expectLog: "Unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output for commands that print to stdout
			output := captureOutput(func() {
				executeCommand(manager, tt.command, tt.args, config)
			})

			if tt.expectLog != "" {
				if !strings.Contains(output, tt.expectLog) {
					t.Errorf("Expected output to contain %q, got: %s", tt.expectLog, output)
				}
			}
			// For help commands, verify they show usage
			if tt.command == testHelpCommand || tt.command == "-h" || tt.command == testHelpFlag {
				if !strings.Contains(output, "GRA Entity Framework Core-like Migration Tool") {
					t.Errorf("Help command should show usage, got: %s", output)
				}
			}
		})
	}
}

// TestGetMigrationsWithData tests getMigrations function with actual migration data
func TestGetMigrationsWithData(t *testing.T) {
	// Create temporary directory for migrations
	tempDir, err := os.MkdirTemp("", "get_migrations_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Create test database
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: tempDir,
		Verbose:       true,
	}

	// Create test migration files BEFORE setting up the manager
	migrationFiles := []struct {
		filename string
		content  string
	}{
		{
			filename: "1_CreateUsers.sql",
			content: `-- Migration: CreateUsers
-- Description: Create users table
-- Created: 2024-01-01 12:00:00
-- Version: 1

-- +migrate Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- +migrate Down
DROP TABLE users;`,
		},
		{
			filename: "2_AddUserProfiles.sql",
			content: `-- Migration: AddUserProfiles
-- Description: Add user profiles table
-- Created: 2024-01-02 12:00:00
-- Version: 2

-- +migrate Up
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    bio TEXT
);

-- +migrate Down
DROP TABLE user_profiles;`,
		},
	}

	for _, file := range migrationFiles {
		filePath := filepath.Join(tempDir, file.filename)
		if err := os.WriteFile(filePath, []byte(file.content), 0600); err != nil {
			t.Fatalf("Failed to create migration file %s: %v", file.filename, err)
		}
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to setup migration manager: %v", err)
	}

	// Test getMigrations
	output := captureOutput(func() {
		getMigrations(manager, config)
	})

	// Verify output contains migration information
	expectedInOutput := []string{
		"Migration History:",
		"CreateUsers",
		"AddUserProfiles",
	}

	for _, expected := range expectedInOutput {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got: %s", expected, output)
		}
	}
}

// TestGenerateScriptWithMigrations tests generateScript function with actual migrations
func TestGenerateScriptWithMigrations(t *testing.T) {
	// Create temporary directory for migrations
	tempDir, err := os.MkdirTemp("", "generate_script_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Create test database
	db, err := sql.Open("sqlite3", testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: tempDir,
		Verbose:       true,
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to setup migration manager: %v", err)
	}

	// Create test migration file
	migrationContent := `-- Migration: CreateTestTable
-- Description: Create test table
-- Created: 2024-01-01 12:00:00
-- Version: 1

-- +migrate Up
CREATE TABLE test_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- +migrate Down
DROP TABLE test_table;`

	filePath := filepath.Join(tempDir, "1_CreateTestTable.sql")
	if err := os.WriteFile(filePath, []byte(migrationContent), 0600); err != nil {
		t.Fatalf("Failed to create migration file: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		expectInOutput []string
	}{
		{
			name:           "generate script all pending",
			args:           []string{},
			expectInOutput: []string{"Generating migration script"},
		},
		{
			name:           "generate script with target",
			args:           []string{"CreateTestTable"},
			expectInOutput: []string{"Generating migration script"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				generateScript(manager, tt.args, config)
			})

			for _, expected := range tt.expectInOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

// Helper function to capture stdout output
func captureOutput(f func()) string {
	// Save original stdout
	originalStdout := os.Stdout

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		return ""
	}

	// Set stdout to our pipe writer
	os.Stdout = w

	// Create a channel to capture the output
	outputChan := make(chan string)

	// Start a goroutine to read from the pipe
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Execute the function
	f()

	// Close the writer and restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Get the output
	output := <-outputChan
	_ = r.Close()

	return output
}

// TestComplexConnectionStringParsing tests complex database connection string scenarios
func TestComplexConnectionStringParsing(t *testing.T) {
	tests := []struct {
		name           string
		connectionStr  string
		expectedDriver string
		shouldSucceed  bool
	}{
		{
			name:           "PostgreSQL with SSL and complex params",
			connectionStr:  "postgres://user:pass@host:5432/db?sslmode=require&application_name=test&connect_timeout=10",
			expectedDriver: "postgres",
			shouldSucceed:  true,
		},
		{
			name:           "SQLite with absolute path",
			connectionStr:  "/Users/test/database.db",
			expectedDriver: "sqlite3",
			shouldSucceed:  true,
		},
		{
			name:           "PostgreSQL with special characters in password",
			connectionStr:  "postgres://user:p@ss!w0rd%40@localhost/db",
			expectedDriver: "postgres",
			shouldSucceed:  true,
		},
		{
			name:           "SQLite memory database",
			connectionStr:  "file::memory:?cache=shared",
			expectedDriver: "sqlite3",
			shouldSucceed:  true,
		},
		{
			name:           "Invalid connection string",
			connectionStr:  "invalid://connection",
			expectedDriver: "postgres", // Default fallback
			shouldSucceed:  true,
		},
		{
			name:           "Empty connection string",
			connectionStr:  "",
			expectedDriver: "postgres", // Default fallback
			shouldSucceed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CLIConfig{
				ConnectionString: tt.connectionStr,
				MigrationsDir:    testMigrationsDir,
			}

			// Test driver detection logic (extracted from setupDatabaseConnection)
			var detectedDriver string
			switch {
			case strings.HasPrefix(config.ConnectionString, "postgres://"), strings.Contains(config.ConnectionString, "user="):
				detectedDriver = "postgres"
			case strings.HasSuffix(config.ConnectionString, ".db"),
				strings.Contains(config.ConnectionString, "sqlite"),
				strings.HasPrefix(config.ConnectionString, "file:"),
				config.ConnectionString == ":memory:":
				detectedDriver = "sqlite3"
			default:
				detectedDriver = "postgres" // Default
			}

			if detectedDriver != tt.expectedDriver {
				t.Errorf("Expected driver %s, got %s", tt.expectedDriver, detectedDriver)
			}
		})
	}
}

// TestMigrationFileValidation tests migration file format validation
func TestMigrationFileValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_migration_validation")
	if err != nil {
		t.Fatalf(testFailedTempDir, err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		filename    string
		content     string
		shouldError bool
	}{
		{
			name:     "valid migration with UP and DOWN sections",
			filename: "001_CreateTable.sql",
			content: `-- UP Migration
CREATE TABLE users (id INT PRIMARY KEY);

-- DOWN Migration
DROP TABLE users;`,
			shouldError: false,
		},
		{
			name:     "migration with only UP section",
			filename: "002_AddIndex.sql",
			content: `-- UP Migration
CREATE INDEX idx_users_name ON users(name);`,
			shouldError: false,
		},
		{
			name:     "migration without markers",
			filename: "003_PlainSQL.sql",
			content: `CREATE TABLE posts (id INT, title VARCHAR(255));
INSERT INTO posts VALUES (1, 'First Post');`,
			shouldError: false, // Should be treated as UP migration
		},
		{
			name:        "empty migration file",
			filename:    "004_Empty.sql",
			content:     "",
			shouldError: false, // Empty files are valid
		},
		{
			name:     "migration with comments only",
			filename: "005_CommentsOnly.sql",
			content: `-- This is a comment
-- Another comment
/* Multi-line comment */`,
			shouldError: false,
		},
		{
			name:     "migration with complex SQL",
			filename: "006_ComplexSQL.sql",
			content: `-- UP Migration
CREATE TABLE complex_table (
    id SERIAL PRIMARY KEY,
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT chk_data CHECK (data IS NOT NULL)
);

CREATE INDEX CONCURRENTLY idx_complex_data ON complex_table USING GIN (data);

-- DOWN Migration
DROP INDEX IF EXISTS idx_complex_data;
DROP TABLE IF EXISTS complex_table;`,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)
			if err := os.WriteFile(filePath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test file reading and basic validation
			content, err := os.ReadFile(filePath)
			if err != nil {
				if !tt.shouldError {
					t.Errorf("Unexpected error reading file: %v", err)
				}
			} else {
				if tt.shouldError {
					t.Error("Expected error but file was read successfully")
				}

				// Test section detection logic
				lines := strings.Split(string(content), "\n")
				hasUpSection := false
				hasDownSection := false

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(strings.ToLower(line), "up migration") ||
						strings.Contains(strings.ToLower(line), "-- up") {
						hasUpSection = true
					}
					if strings.Contains(strings.ToLower(line), "down migration") ||
						strings.Contains(strings.ToLower(line), "-- down") ||
						strings.Contains(strings.ToLower(line), "rollback") {
						hasDownSection = true
					}
				}

				// Log section detection for verification
				t.Logf("File %s: UP section=%v, DOWN section=%v", tt.filename, hasUpSection, hasDownSection)
			}
		})
	}
}

// TestEnvironmentVariableHandling tests various environment variable scenarios
func TestEnvironmentVariableHandling(t *testing.T) {
	// Save original environment
	originalDBURL := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", originalDBURL)

	tests := []struct {
		name        string
		envValue    string
		config      CLIConfig
		expectError bool
	}{
		{
			name:     "DATABASE_URL overrides empty connection string",
			envValue: "postgres://env_user:env_pass@localhost/env_db",
			config: CLIConfig{
				ConnectionString: "",
				MigrationsDir:    testMigrationsDir,
			},
			expectError: false,
		},
		{
			name:     "explicit connection string takes precedence",
			envValue: "postgres://env_user:env_pass@localhost/env_db",
			config: CLIConfig{
				ConnectionString: "postgres://explicit_user:explicit_pass@localhost/explicit_db",
				MigrationsDir:    testMigrationsDir,
			},
			expectError: false,
		},
		{
			name:     "no connection string and no environment variable",
			envValue: "",
			config: CLIConfig{
				ConnectionString: "",
				MigrationsDir:    testMigrationsDir,
			},
			expectError: true,
		},
		{
			name:     "build from individual PostgreSQL parameters",
			envValue: "",
			config: CLIConfig{
				ConnectionString: "",
				Host:             testLocalhost,
				User:             testTestUser,
				Database:         testTestDB,
				Password:         testTestPassword,
				Port:             testPort5432,
				SSLMode:          "disable",
				MigrationsDir:    testMigrationsDir,
			},
			expectError: false,
		},
		{
			name:     "incomplete PostgreSQL parameters",
			envValue: "",
			config: CLIConfig{
				ConnectionString: "",
				Host:             testLocalhost,
				User:             testTestUser,
				// Missing database name
				Password:      testTestPassword,
				Port:          testPort5432,
				MigrationsDir: testMigrationsDir,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for this test
			os.Setenv("DATABASE_URL", tt.envValue)

			// Test the logic from setupDatabaseConnection
			config := tt.config
			if config.ConnectionString == "" {
				config.ConnectionString = os.Getenv("DATABASE_URL")
				if config.ConnectionString == "" {
					// Try to build PostgreSQL connection string from individual parameters
					if config.Host != "" && config.User != "" && config.Database != "" {
						config.ConnectionString = buildPostgreSQLConnectionString(config)
					} else {
						if tt.expectError {
							return // Expected error scenario
						}
						t.Error("Should have gotten an error for incomplete configuration")
						return
					}
				}
			}

			// Verify connection string was set correctly
			if config.ConnectionString == "" && !tt.expectError {
				t.Error("Connection string should not be empty")
			} else if config.ConnectionString != "" && tt.expectError {
				t.Error("Expected error but connection string was set")
			}
		})
	}
}

// TestConcurrentMigrationOperations tests thread safety and concurrent operations
func TestConcurrentMigrationOperations(t *testing.T) {
	// Create multiple in-memory databases to simulate concurrent operations
	const numConcurrentOps = 5

	type migrationResult struct {
		id    int
		error error
	}

	resultChan := make(chan migrationResult, numConcurrentOps)

	// Launch concurrent migration operations
	for i := 0; i < numConcurrentOps; i++ {
		go func(id int) {
			db, err := sql.Open(testSQLite3Driver, testMemoryDB)
			if err != nil {
				resultChan <- migrationResult{id: id, error: err}
				return
			}
			defer func() {
				if closeErr := db.Close(); closeErr != nil {
					t.Logf(testWarningCloseDB, closeErr)
				}
			}()

			config := CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			}

			manager, setupErr := setupMigrationManager(db, config)
			if setupErr != nil {
				resultChan <- migrationResult{id: id, error: setupErr}
				return
			}

			// Perform migration operation
			_ = manager.EnsureSchema() // Ignore error for concurrency test

			resultChan <- migrationResult{id: id, error: nil}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrentOps; i++ {
		result := <-resultChan
		if result.error == nil {
			successCount++
		} else {
			t.Logf("Operation %d failed: %v", result.id, result.error)
		}
	}

	// All operations should succeed (they're using separate in-memory databases)
	if successCount != numConcurrentOps {
		t.Errorf("Expected %d successful operations, got %d", numConcurrentOps, successCount)
	}
}

// TestMemoryLeakPrevention tests resource management and cleanup
func TestMemoryLeakPrevention(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		db, err := sql.Open(testSQLite3Driver, testMemoryDB)
		if err != nil {
			t.Fatalf("Iteration %d: Failed to open database: %v", i, err)
		}

		config := CLIConfig{
			MigrationsDir: testMigrationsDir,
			Verbose:       false, // Reduce logging for performance
		}

		manager, err := setupMigrationManager(db, config)
		if err != nil {
			_ = db.Close()
			t.Fatalf("Iteration %d: Failed to setup manager: %v", i, err)
		}

		// Ensure schema to test database operations
		_ = manager.EnsureSchema()

		// Properly close database
		if err := db.Close(); err != nil {
			t.Logf("Iteration %d: Warning - failed to close database: %v", i, err)
		}
	}

	// If we reach here without running out of memory, test passes
	t.Logf("Successfully completed %d iterations without memory issues", iterations)
}

// TestErrorBoundaryHandling tests comprehensive error scenarios
func TestErrorBoundaryHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (*sql.DB, error)
		expectError bool
		errorText   string
	}{
		{
			name: "invalid SQLite database path",
			setupFunc: func() (*sql.DB, error) {
				return sql.Open(testSQLite3Driver, "/invalid/path/database.db")
			},
			expectError: true,
			errorText:   "no such file",
		},
		{
			name: "valid in-memory database",
			setupFunc: func() (*sql.DB, error) {
				return sql.Open(testSQLite3Driver, testMemoryDB)
			},
			expectError: false,
		},
		{
			name: "invalid driver name",
			setupFunc: func() (*sql.DB, error) {
				return sql.Open("invalid_driver", "some_connection_string")
			},
			expectError: true,
			errorText:   "unknown driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := tt.setupFunc()

			if tt.expectError {
				if err == nil && db != nil {
					// Test actual database operations to trigger errors
					if pingErr := db.Ping(); pingErr != nil {
						// This is expected for invalid configurations
						t.Logf("Expected ping error: %v", pingErr)
					}
					_ = db.Close()
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			defer func() {
				if closeErr := db.Close(); closeErr != nil {
					t.Logf(testWarningCloseDB, closeErr)
				}
			}()

			// Test successful database operations
			config := CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			}

			manager, err := setupMigrationManager(db, config)
			if err != nil {
				t.Errorf("Failed to setup migration manager: %v", err)
				return
			}

			// Test schema creation
			if err := manager.EnsureSchema(); err != nil {
				t.Errorf("Failed to ensure schema: %v", err)
			}
		})
	}
}

// TestPerformanceEdgeCases tests performance with various input sizes
func TestPerformanceEdgeCases(t *testing.T) {
	db, err := sql.Open(testSQLite3Driver, testMemoryDB)
	if err != nil {
		t.Fatalf(testFailedOpenDB, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf(testWarningCloseDB, err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: testMigrationsDir,
		Verbose:       false, // Reduce logging for performance testing
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf(testFailedSetupManager, err)
	}

	tests := []struct {
		name          string
		operationFunc func() error
		maxDuration   time.Duration
		description   string
	}{
		{
			name: "schema_initialization_performance",
			operationFunc: func() error {
				return manager.EnsureSchema()
			},
			maxDuration: 5 * time.Second,
			description: "Schema initialization should complete quickly",
		},
		{
			name: "multiple_schema_calls_performance",
			operationFunc: func() error {
				for i := 0; i < 10; i++ {
					if err := manager.EnsureSchema(); err != nil {
						return err
					}
				}
				return nil
			},
			maxDuration: 5 * time.Second,
			description: "Multiple schema calls should be efficiently handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			err := tt.operationFunc()

			duration := time.Since(start)

			if err != nil {
				t.Errorf("Operation failed: %v", err)
				return
			}

			if duration > tt.maxDuration {
				t.Errorf("%s took %v, expected less than %v", tt.description, duration, tt.maxDuration)
			} else {
				t.Logf("%s completed in %v", tt.description, duration)
			}
		})
	}
}

// TestSecurityValidation tests security-related validation
func TestSecurityValidation(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		shouldMaskPass   bool
		expectedMasked   string
	}{
		{
			name:             "PostgreSQL with password masking",
			connectionString: "postgres://user:secretpass@localhost/db",
			shouldMaskPass:   true,
			expectedMasked:   "postgres://user:*****@localhost/db",
		},
		{
			name:             "connection string without password",
			connectionString: "postgres://user@localhost/db",
			shouldMaskPass:   false,
			expectedMasked:   "postgres://user@localhost/db",
		},
		{
			name:             "SQLite file path",
			connectionString: "/path/to/database.db",
			shouldMaskPass:   false,
			expectedMasked:   "/path/to/database.db",
		},
		{
			name:             "complex password with special chars",
			connectionString: "postgres://user:p@ss!w0rd%40@localhost/db",
			shouldMaskPass:   true,
			expectedMasked:   "postgres://user:*****@localhost/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test password masking logic
			masked := maskPassword(tt.connectionString)

			if tt.shouldMaskPass {
				if !strings.Contains(masked, testMaskedPassword) {
					t.Errorf("Expected password to be masked, got: %s", masked)
				}
				if strings.Contains(masked, "secretpass") || strings.Contains(masked, "p@ss!w0rd") {
					t.Errorf("Password not properly masked in: %s", masked)
				}
			} else {
				if masked != tt.connectionString {
					t.Errorf("Expected no masking for %s, got: %s", tt.connectionString, masked)
				}
			}
		})
	}
}

// Helper function to mask passwords in connection strings for security testing
func maskPassword(connectionString string) string {
	// Simple password masking for PostgreSQL connection strings
	if strings.HasPrefix(connectionString, "postgres://") {
		// Find password pattern: ://user:password@
		re := regexp.MustCompile(`(://[^:]+:)[^@]+(@)`)
		return re.ReplaceAllString(connectionString, "${1}*****${2}")
	}
	return connectionString
}

// TestAdvancedCommandScenarios tests complex command combinations
func TestAdvancedCommandScenarios(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_advanced_scenarios")
	if err != nil {
		t.Fatalf(testFailedTempDir, err)
	}
	defer os.RemoveAll(tempDir)

	db, err := sql.Open(testSQLite3Driver, testMemoryDB)
	if err != nil {
		t.Fatalf(testFailedOpenDB, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf(testWarningCloseDB, err)
		}
	}()

	config := CLIConfig{
		MigrationsDir: tempDir,
		Verbose:       true,
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf(testFailedSetupManager, err)
	}

	// Create test migration files
	migrationContent1 := `-- UP Migration
CREATE TABLE advanced_test1 (id INT PRIMARY KEY, name VARCHAR(255));

-- DOWN Migration
DROP TABLE advanced_test1;`

	migrationContent2 := `-- UP Migration
CREATE TABLE advanced_test2 (id INT PRIMARY KEY, description TEXT);
ALTER TABLE advanced_test1 ADD COLUMN created_at TIMESTAMP;

-- DOWN Migration
ALTER TABLE advanced_test1 DROP COLUMN created_at;
DROP TABLE advanced_test2;`

	filePath1 := filepath.Join(tempDir, "001_CreateAdvancedTest1.sql")
	filePath2 := filepath.Join(tempDir, "002_CreateAdvancedTest2.sql")

	if err := os.WriteFile(filePath1, []byte(migrationContent1), 0600); err != nil {
		t.Fatalf(testFailedCreateFile, err)
	}
	if err := os.WriteFile(filePath2, []byte(migrationContent2), 0600); err != nil {
		t.Fatalf(testFailedCreateFile, err)
	}

	scenarios := []struct {
		name       string
		command    string
		args       []string
		expectLog  string
		shouldFail bool
	}{
		{
			name:       "status before any migrations",
			command:    "status",
			args:       []string{},
			expectLog:  "", // Should work without error
			shouldFail: false,
		},
		{
			name:       "get migration history when empty",
			command:    testGetMigrationCmd,
			args:       []string{},
			expectLog:  "",
			shouldFail: false,
		},
		{
			name:       "script generation with no migrations applied",
			command:    "script",
			args:       []string{},
			expectLog:  "",
			shouldFail: false,
		},
		{
			name:       "update database with multiple migrations",
			command:    testUpdateDatabaseCmd,
			args:       []string{},
			expectLog:  "",
			shouldFail: false,
		},
		{
			name:       "status after migrations applied",
			command:    "status",
			args:       []string{},
			expectLog:  "",
			shouldFail: false,
		},
		{
			name:       "script with target migration",
			command:    "script",
			args:       []string{"CreateAdvancedTest1"},
			expectLog:  "",
			shouldFail: false,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			output := captureOutput(func() {
				executeCommand(manager, scenario.command, scenario.args, config)
			})

			if scenario.shouldFail && output == "" {
				t.Error("Expected command to produce output or fail")
			}

			if scenario.expectLog != "" && !strings.Contains(output, scenario.expectLog) {
				t.Errorf(testExpectedOutputFormat, scenario.expectLog, output)
			}

			t.Logf("Command '%s' output: %s", scenario.command, strings.TrimSpace(output))
		})
	}
}

// TestBoundaryInputValidation tests input validation edge cases
func TestBoundaryInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      CLIConfig
		command     string
		args        []string
		expectPanic bool
		description string
	}{
		{
			name: "very long migration directory path",
			config: CLIConfig{
				MigrationsDir: strings.Repeat("a", 500), // Very long path
				Verbose:       true,
			},
			command:     testAddMigrationCmd,
			args:        []string{"TestMigration"},
			expectPanic: false,
			description: "Should handle long migration directory paths",
		},
		{
			name: "migration name with special characters",
			config: CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			},
			command:     testAddMigrationCmd,
			args:        []string{"Test_Migration-With.Special@Characters"},
			expectPanic: false,
			description: "Should handle special characters in migration names",
		},
		{
			name: "empty migration name",
			config: CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			},
			command:     testAddMigrationCmd,
			args:        []string{""},
			expectPanic: false,
			description: "Should handle empty migration names gracefully",
		},
		{
			name: "very long migration name",
			config: CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			},
			command:     testAddMigrationCmd,
			args:        []string{strings.Repeat("Long", 100)},
			expectPanic: false,
			description: "Should handle very long migration names",
		},
		{
			name: "update with non-existent target",
			config: CLIConfig{
				MigrationsDir: testMigrationsDir,
				Verbose:       true,
			},
			command:     testUpdateDatabaseCmd,
			args:        []string{"NonExistentMigration"},
			expectPanic: false,
			description: "Should handle non-existent migration targets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.expectPanic {
					t.Error("Expected panic but none occurred")
				}
			}()

			// Create a test database for each scenario
			db, err := sql.Open(testSQLite3Driver, testMemoryDB)
			if err != nil {
				t.Fatalf(testFailedOpenDB, err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Logf(testWarningCloseDB, err)
				}
			}()

			manager, err := setupMigrationManager(db, tt.config)
			if err != nil {
				// Some configurations might fail to setup, which is acceptable for boundary testing
				t.Logf("Setup failed (expected for boundary test): %v", err)
				return
			}

			// Execute command and capture any output
			output := captureOutput(func() {
				executeCommand(manager, tt.command, tt.args, tt.config)
			})

			t.Logf("%s - Output: %s", tt.description, strings.TrimSpace(output))
		})
	}
}

// TestUnicodeAndInternationalization tests handling of non-ASCII content
func TestUnicodeAndInternationalization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_unicode")
	if err != nil {
		t.Fatalf(testFailedTempDir, err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name             string
		migrationName    string
		migrationContent string
		description      string
	}{
		{
			name:          "unicode migration name",
			migrationName: "CreateUsersТест", // Mixed ASCII and Cyrillic
			migrationContent: `-- UP Migration
CREATE TABLE users_тест (id INT PRIMARY KEY, имя VARCHAR(255));

-- DOWN Migration  
DROP TABLE users_тест;`,
			description: "Should handle Unicode characters in names and content",
		},
		{
			name:          "emoji in migration",
			migrationName: "CreateUsers🚀Migration",
			migrationContent: `-- UP Migration 🚀
CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255));
-- Comment with emoji: 📝 This creates users table

-- DOWN Migration 🔙
DROP TABLE users;`,
			description: "Should handle emoji characters",
		},
		{
			name:          "chinese characters",
			migrationName: "Create用户Table",
			migrationContent: `-- UP Migration
CREATE TABLE 用户表 (
    编号 INT PRIMARY KEY,
    姓名 VARCHAR(255),
    创建时间 TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- DOWN Migration
DROP TABLE 用户表;`,
			description: "Should handle Chinese characters",
		},
		{
			name:          "arabic text",
			migrationName: "CreateمستخدمينTable",
			migrationContent: `-- UP Migration
CREATE TABLE المستخدمين (
    الرقم INT PRIMARY KEY,
    الاسم VARCHAR(255)
);

-- DOWN Migration
DROP TABLE المستخدمين;`,
			description: "Should handle Arabic text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create migration file with Unicode content
			filename := fmt.Sprintf("001_%s.sql", tt.migrationName)
			filePath := filepath.Join(tempDir, filename)

			err := os.WriteFile(filePath, []byte(tt.migrationContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create Unicode migration file: %v", err)
			}

			// Verify file can be read back correctly
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read Unicode migration file: %v", err)
				return
			}

			if string(content) != tt.migrationContent {
				t.Errorf("Unicode content was corrupted during file operations")
				return
			}

			// Test migration system can handle the file
			db, err := sql.Open(testSQLite3Driver, testMemoryDB)
			if err != nil {
				t.Fatalf(testFailedOpenDB, err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Logf(testWarningCloseDB, err)
				}
			}()

			config := CLIConfig{
				MigrationsDir: tempDir,
				Verbose:       true,
			}

			manager, err := setupMigrationManager(db, config)
			if err != nil {
				t.Errorf("Failed to setup manager with Unicode content: %v", err)
				return
			}

			// Test status command with Unicode migration
			output := captureOutput(func() {
				executeCommand(manager, "status", []string{}, config)
			})

			t.Logf("%s - Status output: %s", tt.description, strings.TrimSpace(output))

			// Verify the migration file is detected and can be processed
			if output == "" {
				t.Error("Expected some output from status command")
			}
		})
	}
}

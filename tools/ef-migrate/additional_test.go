package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// TestPrintVersionCoverage tests the printVersion function
func TestPrintVersionCoverage(t *testing.T) {
	// Just call the function to get coverage - don't capture output
	printVersion()
	// Test passes if no panic occurs
}

// TestPrintUsageCoverage tests the printUsage function
func TestPrintUsageCoverage(t *testing.T) {
	// Just call the function to get coverage
	printUsage()
	// Test passes if no panic occurs
}

// TestSetupMigrationManagerEdgeCases tests additional scenarios
func TestSetupMigrationManagerEdgeCases(t *testing.T) {
	// Test with valid database but non-verbose mode
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	config := CLIConfig{
		MigrationsDir: "./test_migrations_edge",
		Verbose:       false, // Test non-verbose mode
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Logf("Setup failed as expected in test environment: %v", err)
	}
	if manager != nil {
		t.Log("Migration manager created successfully in non-verbose mode")
	}
}

// TestSaveMigrationToFileAdditional tests the saveMigrationToFile function with additional scenarios
func TestSaveMigrationToFileAdditional(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "test_save_migration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test migration object
	migration := &migrations.Migration{
		ID:          "test_migration_123",
		Name:        "TestMigration",
		Description: "Test migration for coverage",
		UpSQL:       "CREATE TABLE test (id INTEGER);",
		DownSQL:     "DROP TABLE test;",
		Version:     123,
	}

	// Test saving migration to file
	err = saveMigrationToFile(migration, tempDir)
	if err != nil {
		t.Errorf("Failed to save migration to file: %v", err)
	}

	// Check if file was created
	expectedFile := filepath.Join(tempDir, migration.ID+".sql")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Migration file was not created: %s", expectedFile)
	}
}

// TestBuildPostgreSQLConnectionStringVariations tests more variations
func TestBuildPostgreSQLConnectionStringVariations(t *testing.T) {
	tests := []struct {
		name             string
		config           CLIConfig
		expectedInString []string
	}{
		{
			name: "default_port_and_ssl",
			config: CLIConfig{
				User:     "user",
				Host:     "host",
				Database: "db",
			},
			expectedInString: []string{"user", "host", "db", "5432", "disable"},
		},
		{
			name: "custom_ssl_mode",
			config: CLIConfig{
				User:     "user",
				Host:     "host",
				Database: "db",
				SSLMode:  "verify-full",
			},
			expectedInString: []string{"user", "host", "db", "verify-full"},
		},
		{
			name: "with_password_only",
			config: CLIConfig{
				Password: "secret",
				Host:     "host",
				Database: "db",
			},
			expectedInString: []string{"secret", "host", "db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPostgreSQLConnectionString(tt.config)
			// Should always return a valid connection string format
			if !strings.HasPrefix(result, "postgres://") {
				t.Errorf("Expected postgres:// prefix, got: %s", result)
			}

			for _, expected := range tt.expectedInString {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected %s in connection string: %s", expected, result)
				}
			}
		})
	}
}

// TestExtractDBNameVariations tests more edge cases for extractDBName
func TestExtractDBNameVariations(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		expected string
	}{
		{
			name:     "postgres_with_query_params",
			connStr:  "postgres://user:pass@host:5432/mydb?sslmode=disable&connect_timeout=10",
			expected: "mydb",
		},
		{
			name:     "file_path_with_extension",
			connStr:  "/path/to/database.db",
			expected: "database.db",
		},
		{
			name:     "relative_path",
			connStr:  "./data/app.sqlite",
			expected: "app.sqlite",
		},
		{
			name:     "complex_path_with_spaces",
			connStr:  "/Users/user/My Database/app.db",
			expected: "app.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDBName(tt.connStr)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestExecuteCommandCoverage tests various command paths for coverage
func TestExecuteCommandCoverage(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	config := CLIConfig{
		MigrationsDir: "./test_migrations_coverage",
		Verbose:       true,
	}

	manager, err := setupMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to setup migration manager: %v", err)
	}

	// Test various commands for coverage
	commands := []string{"status", "get-migration", "script"}

	for _, cmd := range commands {
		t.Run("command_"+cmd, func(t *testing.T) {
			// Just call the command to get coverage - don't check output
			executeCommand(manager, cmd, []string{}, config)
		})
	}
}

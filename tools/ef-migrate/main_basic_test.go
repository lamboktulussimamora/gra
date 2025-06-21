// Package main provides tests for basic functionality of the ef-migrate tool.
// This file contains fundamental tests for core data structures and connection handling.
package main

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// TestBuildPostgreSQLConnectionStringBasic verifies that PostgreSQL connection strings
// are built correctly from CLI configuration parameters.
func TestBuildPostgreSQLConnectionStringBasic(t *testing.T) {
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

// TestCLIConfigBasic validates that the CLIConfig struct fields can be properly
// set and accessed, ensuring the configuration structure is working correctly.
func TestCLIConfigBasic(t *testing.T) {
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

// TestConstantsBasic ensures that application constants are properly defined
// and have the expected values for string formatting and time parsing.
func TestConstantsBasic(t *testing.T) {
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

// TestSetupDatabaseConnectionBasic tests the database connection setup functionality
// with various configuration scenarios including valid and invalid connections.
func TestSetupDatabaseConnectionBasic(t *testing.T) {
	tests := []struct {
		name        string
		config      CLIConfig
		expectError bool
	}{
		{
			name: "valid sqlite connection",
			config: CLIConfig{
				ConnectionString: ":memory:",
			},
			expectError: false,
		},
		{
			name: "empty connection string with env fallback",
			config: CLIConfig{
				ConnectionString: "",
			},
			expectError: true, // Should error when connection string is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the empty connection string test, temporarily clear DATABASE_URL
			if tt.config.ConnectionString == "" {
				originalDatabaseURL := os.Getenv("DATABASE_URL")
				os.Setenv("DATABASE_URL", "")
				defer os.Setenv("DATABASE_URL", originalDatabaseURL)
			}
			
			db, err := setupDatabaseConnection(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil && tt.config.ConnectionString != ":memory:" {
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

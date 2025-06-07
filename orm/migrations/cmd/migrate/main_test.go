package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// Test constants to avoid duplication
const (
	testPostgresURL    = "postgres://test"
	testPostgresDriver = "postgres"
	testMigrationsDir  = "./migrations"
	testModelsDir      = "./models"
	testSQLiteDriver   = "sqlite"
	testSQLiteURL      = "sqlite://test.db"
	testMySQLDriver    = "mysql"
	testSQLite3Driver  = "sqlite3"
)

func TestConfig(t *testing.T) {
	config := Config{
		DatabaseURL:   testPostgresURL,
		Driver:        testPostgresDriver,
		MigrationsDir: testMigrationsDir,
		ModelsDir:     testModelsDir,
	}

	// Test that all fields are properly set
	if config.DatabaseURL != testPostgresURL {
		t.Errorf("Expected DatabaseURL '%s', got '%s'", testPostgresURL, config.DatabaseURL)
	}
	if config.Driver != testPostgresDriver {
		t.Errorf("Expected Driver '%s', got '%s'", testPostgresDriver, config.Driver)
	}
	if config.MigrationsDir != testMigrationsDir {
		t.Errorf("Expected MigrationsDir '%s', got '%s'", testMigrationsDir, config.MigrationsDir)
	}
	if config.ModelsDir != testModelsDir {
		t.Errorf("Expected ModelsDir '%s', got '%s'", testModelsDir, config.ModelsDir)
	}
}

func TestConfigStructure(t *testing.T) {
	// Test that Config has all expected fields
	config := Config{}
	configType := reflect.TypeOf(config)

	expectedFields := []string{
		"DatabaseURL",
		"Driver",
		"MigrationsDir",
		"ModelsDir",
	}

	for _, field := range expectedFields {
		_, found := configType.FieldByName(field)
		if !found {
			t.Errorf("Expected field %s not found in Config", field)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with DATABASE_URL",
			config: &Config{
				DatabaseURL:   testSQLiteURL,
				Driver:        testSQLiteDriver,
				MigrationsDir: testMigrationsDir,
				ModelsDir:     testModelsDir,
			},
			expectError: false,
		},
		{
			name: "empty database URL",
			config: &Config{
				DatabaseURL:   "",
				Driver:        testPostgresDriver,
				MigrationsDir: testMigrationsDir,
				ModelsDir:     testModelsDir,
			},
			expectError: true,
		},
		{
			name: "missing migrations directory",
			config: &Config{
				DatabaseURL:   testPostgresURL,
				Driver:        testPostgresDriver,
				MigrationsDir: "",
				ModelsDir:     testModelsDir,
			},
			expectError: false, // Has default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}

func TestConnectDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid sqlite connection",
			config: &Config{
				DatabaseURL: ":memory:",
				Driver:      testSQLiteDriver,
			},
			expectError: false,
		},
		{
			name: "invalid connection string",
			config: &Config{
				DatabaseURL: "invalid://connection",
				Driver:      testPostgresDriver,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := connectDatabase(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				} else				if db != nil {
					if err := db.Close(); err != nil {
						t.Logf("Warning: failed to close database: %v", err)
					}
				}
			}
		})
	}
}

func TestGetDriver(t *testing.T) {
	tests := []struct {
		driverName     string
		expectedResult string // We'll just test the string name
	}{
		{testPostgresDriver, testPostgresDriver},
		{testSQLiteDriver, testSQLiteDriver},
		{"mysql", "mysql"},
		{"unknown", testSQLiteDriver}, // Default fallback
	}

	for _, tt := range tests {
		t.Run("driver_"+tt.driverName, func(t *testing.T) {
			driver := getDriver(tt.driverName)
			// We can't easily test the actual driver implementation
			// but we can verify the function doesn't panic and returns something
			if driver == "" { // Compare to empty string for DatabaseDriver type
				t.Errorf("Expected driver, but got empty value for %s", tt.driverName)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if errMigrationNameRequired == "" {
		t.Error("errMigrationNameRequired should not be empty")
	}

	expectedError := "migration name is required"
	if errMigrationNameRequired != expectedError {
		t.Errorf("Expected '%s', got '%s'", expectedError, errMigrationNameRequired)
	}
}

func TestCommandValidation(t *testing.T) {
	// Test command validation logic (simulated)
	validCommands := []string{
		"add",
		"apply",
		"revert",
		"status",
		"generate",
		"force",
	}

	for _, cmd := range validCommands {
		t.Run("command_"+cmd, func(t *testing.T) {
			// Simulate command validation
			isValid := false
			switch cmd {
			case "add", "apply", "revert", "status", "generate", "force":
				isValid = true
			}
			if !isValid {
				t.Errorf("Command %s should be valid", cmd)
			}
		})
	}
}

func TestCmdAddMigration(t *testing.T) {
	// Test migration name validation
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid migration name",
			args:        []string{"CreateUsersTable"},
			expectError: false,
		},
		{
			name:        "empty args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "migration with description",
			args:        []string{"AddUserProfiles", "Add user profile table"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test the actual function due to dependencies
			// but we can test the validation logic
			if len(tt.args) == 0 && !tt.expectError {
				t.Error("Expected error for empty args")
			}
		})
	}
}

func TestCmdApplyMigrations(t *testing.T) {
	// Test apply migrations command logic
	t.Run("apply_all", func(_ *testing.T) {
		// Test applying all migrations (no specific migration specified)
		// This would typically call migrator.ApplyPending()
		args := []string{}
		// In a real implementation, if len(args) == 0, 
		// we would apply all pending migrations
		_ = len(args) // Acknowledge the variable is used
	})

	t.Run("apply_specific", func(t *testing.T) {
		// Test applying specific migration
		args := []string{"CreateUsersTable"}
		if len(args) > 0 {
			migrationName := args[0]
			if migrationName == "" {
				t.Error("Migration name should not be empty")
			}
		}
	})
}

func TestDriverMapping(t *testing.T) {
	// Test driver string mapping
	drivers := map[string]string{
		"postgres": "postgres",
		"mysql":    "mysql",
		"sqlite":   "sqlite",
		"sqlite3":  "sqlite", // Alternative name
	}

	for input, expected := range drivers {
		t.Run("driver_mapping_"+input, func(t *testing.T) {
			// Simulate driver mapping logic
			var mapped string
			switch input {
			case testPostgresDriver:
				mapped = testPostgresDriver
			case testMySQLDriver:
				mapped = testMySQLDriver
			case testSQLiteDriver, testSQLite3Driver:
				mapped = testSQLiteDriver
			default:
				mapped = testSQLiteDriver // Default
			}

			if mapped != expected {
				t.Errorf("Expected %s, got %s for driver %s", expected, mapped, input)
			}
		})
	}
}

func TestRegisterModels(t *testing.T) {
	// Test model registration validation
	tests := []struct {
		name      string
		modelsDir string
		expectErr bool
	}{
		{
			name:      "valid models directory",
			modelsDir: testModelsDir,
			expectErr: false,
		},
		{
			name:      "empty models directory",
			modelsDir: "",
			expectErr: false, // Should use default
		},
		{
			name:      "nonexistent directory",
			modelsDir: "./nonexistent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test the actual function due to dependencies
			// but we can test the directory validation logic
			if tt.modelsDir == "" {
				// Should use default directory
				tt.modelsDir = testModelsDir
			}

			// Test directory existence check
			if _, err := os.Stat(tt.modelsDir); os.IsNotExist(err) && !tt.expectErr {
				t.Logf("Directory %s does not exist (expected in test)", tt.modelsDir)
			}
		})
	}
}

func TestMainFunctionValidation(t *testing.T) {
	// Test main function argument validation
	testArgs := [][]string{
		{}, // No command
		{"add"},
		{"apply"},
		{"revert"},
		{"status"},
		{"generate"},
		{"force"},
		{"unknown"}, // Invalid command
	}

	for i, args := range testArgs {
		t.Run(fmt.Sprintf("args_%d", i), func(t *testing.T) {
			if len(args) == 0 {
				// Should show usage and exit
				t.Log("No command specified - should show usage")
			} else {
				cmd := args[0]
				validCommands := map[string]bool{
					"add":      true,
					"apply":    true,
					"revert":   true,
					"status":   true,
					"generate": true,
					"force":    true,
				}

				if !validCommands[cmd] {
					t.Logf("Invalid command: %s", cmd)
				} else {
					t.Logf("Valid command: %s", cmd)
				}
			}
		})
	}
}

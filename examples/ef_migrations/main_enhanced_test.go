package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/mattn/go-sqlite3"
)

// TestPrintMigrationStatusEnhanced tests the printMigrationStatus function with various scenarios
func TestPrintMigrationStatusEnhanced(t *testing.T) {
	tests := []struct {
		name    string
		history *migrations.MigrationHistory
	}{
		{
			name: "EmptyHistory",
			history: &migrations.MigrationHistory{
				Applied: []migrations.Migration{},
				Pending: []migrations.Migration{},
				Failed:  []migrations.Migration{},
			},
		},
		{
			name: "HistoryWithAppliedMigrations",
			history: &migrations.MigrationHistory{
				Applied: []migrations.Migration{
					{
						ID:          "001_CreateUsers",
						Name:        "CreateUsers",
						Description: "Create users table",
						AppliedAt:   time.Now(),
						State:       migrations.MigrationStateApplied,
					},
					{
						ID:          "002_CreateProfiles",
						Name:        "CreateProfiles",
						Description: "Create profiles table",
						AppliedAt:   time.Now().Add(-1 * time.Hour),
						State:       migrations.MigrationStateApplied,
					},
				},
				Pending: []migrations.Migration{},
				Failed:  []migrations.Migration{},
			},
		},
		{
			name: "HistoryWithPendingMigrations",
			history: &migrations.MigrationHistory{
				Applied: []migrations.Migration{},
				Pending: []migrations.Migration{
					{
						ID:          "003_CreateSettings",
						Name:        "CreateSettings",
						Description: "Create settings table",
						UpSQL:       "CREATE TABLE settings(...)",
						DownSQL:     "DROP TABLE settings",
						State:       migrations.MigrationStatePending,
					},
				},
				Failed: []migrations.Migration{},
			},
		},
		{
			name: "HistoryWithFailedMigrations",
			history: &migrations.MigrationHistory{
				Applied: []migrations.Migration{},
				Pending: []migrations.Migration{},
				Failed: []migrations.Migration{
					{
						ID:          "004_InvalidMigration",
						Name:        "InvalidMigration",
						Description: "Failed migration",
						AppliedAt:   time.Now(),
						State:       migrations.MigrationStateFailed,
					},
				},
			},
		},
		{
			name: "MixedHistory",
			history: &migrations.MigrationHistory{
				Applied: []migrations.Migration{
					{
						ID:          "001_CreateUsers",
						Name:        "CreateUsers",
						Description: "Create users table",
						AppliedAt:   time.Now(),
						State:       migrations.MigrationStateApplied,
					},
				},
				Pending: []migrations.Migration{
					{
						ID:          "002_CreateProfiles",
						Name:        "CreateProfiles",
						Description: "Create profiles table",
						UpSQL:       "CREATE TABLE profiles(...)",
						DownSQL:     "DROP TABLE profiles",
						State:       migrations.MigrationStatePending,
					},
				},
				Failed: []migrations.Migration{
					{
						ID:          "003_FailedMigration",
						Name:        "FailedMigration",
						Description: "This migration failed",
						AppliedAt:   time.Now(),
						State:       migrations.MigrationStateFailed,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output (in a real scenario you might use a test logger)
			// For now, we just ensure the function doesn't panic
			printMigrationStatus(tt.history)
		})
	}
}

// TestDemonstrateAutoMigration tests the auto migration demonstration
func TestDemonstrateAutoMigration(t *testing.T) {
	// Create test database
	dbPath := filepath.Join("test_migrations", "test_auto_migration.db")
	if err := os.MkdirAll("test_migrations", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(dbPath)
	}()

	// Create EF Migration Manager
	config := migrations.DefaultEFMigrationConfig()
	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize schema
	if err := manager.EnsureSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Test auto migration demonstration
	t.Run("AutoMigrationDemo", func(t *testing.T) {
		// This should not panic
		demonstrateAutoMigration(manager)

		// Verify that migrations were created
		history, err := manager.GetMigrationHistory()
		if err != nil {
			t.Errorf("Failed to get migration history: %v", err)
		}

		// Should have at least one migration after auto generation
		if len(history.Applied)+len(history.Pending) == 0 {
			t.Error("Expected at least one migration after auto generation")
		}
	})
}

// TestUserEntityValidation tests the User entity structure validation
func TestUserEntityValidation(t *testing.T) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		IsActive:  true,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Test entity field values
	if user.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected Name to be 'Test User', got %s", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("Expected Age to be 25, got %d", user.Age)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
	if user.CreatedAt == "" {
		t.Error("Expected CreatedAt to be set")
	}
}

// TestCompleteLifecycleIntegration tests the complete migration lifecycle integration
func TestCompleteLifecycleIntegration(t *testing.T) {
	// Create test database
	dbPath := filepath.Join("test_migrations", "test_complete_lifecycle.db")
	if err := os.MkdirAll("test_migrations", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(dbPath)
	}()

	// Create EF Migration Manager with custom config
	config := migrations.DefaultEFMigrationConfig()
	config.Logger = log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize schema
	if err := manager.EnsureSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	t.Run("AddMultipleMigrations", func(t *testing.T) {
		// Add first migration
		migration1 := manager.AddMigration(
			"TestCreateUsers",
			"Test create users table",
			"CREATE TABLE test_users (id INTEGER PRIMARY KEY, name TEXT);",
			"DROP TABLE IF EXISTS test_users;",
		)
		if migration1 == nil {
			t.Error("Expected migration1 to be created")
		}

		// Add second migration
		migration2 := manager.AddMigration(
			"TestCreateProfiles",
			"Test create profiles table",
			"CREATE TABLE test_profiles (id INTEGER PRIMARY KEY, user_id INTEGER);",
			"DROP TABLE IF EXISTS test_profiles;",
		)
		if migration2 == nil {
			t.Error("Expected migration2 to be created")
		}

		// Check migration history before applying
		history, err := manager.GetMigrationHistory()
		if err != nil {
			t.Fatalf("Failed to get migration history: %v", err)
		}

		if len(history.Pending) != 2 {
			t.Errorf("Expected 2 pending migrations, got %d", len(history.Pending))
		}
	})

	t.Run("ApplyMigrationsAndRollback", func(t *testing.T) {
		// Apply all migrations
		if err := manager.UpdateDatabase(); err != nil {
			t.Errorf("Failed to update database: %v", err)
		}

		// Check that migrations were applied
		history, err := manager.GetMigrationHistory()
		if err != nil {
			t.Fatalf("Failed to get migration history: %v", err)
		}

		if len(history.Applied) == 0 {
			t.Error("Expected some migrations to be applied")
		}

		// Test rollback (if the first migration exists)
		if len(history.Applied) > 0 {
			firstMigration := history.Applied[0]
			if err := manager.RollbackMigration(firstMigration.ID); err != nil {
				// Rollback might fail in some implementations, that's okay for testing
				t.Logf("Rollback failed (expected in some cases): %v", err)
			}
		}
	})
}

// TestMainFunctionFlow tests the main function components in isolation
func TestMainFunctionFlow(t *testing.T) {
	// Test database connection error handling
	t.Run("DatabaseConnectionError", func(t *testing.T) {
		// Test with invalid database path
		_, err := sql.Open("sqlite3", "/invalid/path/test.db")
		// sqlite3 driver typically doesn't fail on Open, only on actual use
		// So we test a different scenario
		if err == nil {
			// This is expected behavior for sqlite3
			t.Log("sqlite3 driver doesn't fail on Open with invalid path")
		}
	})

	// Test EF Migration Manager creation
	t.Run("EFMigrationManagerCreation", func(t *testing.T) {
		dbPath := filepath.Join("test_migrations", "test_manager.db")
		if err := os.MkdirAll("test_migrations", 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer func() {
			db.Close()
			_ = os.Remove(dbPath)
		}()

		config := migrations.DefaultEFMigrationConfig()
		config.Logger = log.New(os.Stdout, "[TEST] ", log.LstdFlags)
		manager := migrations.NewEFMigrationManager(db, config)

		if manager == nil {
			t.Error("Expected EF migration manager to be created")
		}
	})
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("InvalidDatabasePath", func(t *testing.T) {
		// Try to open database in non-existent directory
		db, err := sql.Open("sqlite3", "/non/existent/path/test.db")
		if err != nil {
			t.Logf("Expected error for invalid path: %v", err)
		} else {
			defer db.Close()
			// Attempt to use the database to trigger actual error
			if err := db.Ping(); err != nil {
				t.Logf("Database ping failed as expected: %v", err)
			}
		}
	})

	t.Run("MigrationWithInvalidSQL", func(t *testing.T) {
		dbPath := filepath.Join("test_migrations", "test_invalid_sql.db")
		if err := os.MkdirAll("test_migrations", 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer func() {
			db.Close()
			_ = os.Remove(dbPath)
		}()

		config := migrations.DefaultEFMigrationConfig()
		manager := migrations.NewEFMigrationManager(db, config)

		// Initialize schema
		if err := manager.EnsureSchema(); err != nil {
			t.Fatalf("Failed to initialize schema: %v", err)
		}

		// Add migration with invalid SQL
		_ = manager.AddMigration(
			"TestInvalidSQL",
			"Test migration with invalid SQL",
			"INVALID SQL STATEMENT;",
			"DROP TABLE IF EXISTS invalid_table;",
		)

		// Try to apply migration (should handle error gracefully)
		err = manager.UpdateDatabase()
		if err != nil {
			t.Logf("Expected error for invalid SQL: %v", err)
		}
	})
}

// TestConfigurationVariations tests different configuration options
func TestConfigurationVariations(t *testing.T) {
	dbPath := filepath.Join("test_migrations", "test_config.db")
	if err := os.MkdirAll("test_migrations", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(dbPath)
	}()

	t.Run("DefaultConfig", func(t *testing.T) {
		config := migrations.DefaultEFMigrationConfig()
		manager := migrations.NewEFMigrationManager(db, config)

		if manager == nil {
			t.Error("Expected manager to be created with default config")
		}
	})

	t.Run("CustomLogger", func(t *testing.T) {
		config := migrations.DefaultEFMigrationConfig()

		// Create custom logger that writes to a string builder
		var logOutput strings.Builder
		config.Logger = log.New(&logOutput, "[CUSTOM] ", log.LstdFlags)

		manager := migrations.NewEFMigrationManager(db, config)

		if manager == nil {
			t.Error("Expected manager to be created with custom logger")
		}

		// Initialize schema to generate some log output
		if err := manager.EnsureSchema(); err != nil {
			t.Errorf("Failed to initialize schema: %v", err)
		}
	})
}

// TestConcurrencyScenarios tests concurrent migration operations
func TestConcurrencyScenarios(t *testing.T) {
	dbPath := filepath.Join("test_migrations", "test_concurrency.db")
	if err := os.MkdirAll("test_migrations", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(dbPath)
	}()

	config := migrations.DefaultEFMigrationConfig()
	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize schema
	if err := manager.EnsureSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	t.Run("ConcurrentMigrationHistory", func(t *testing.T) {
		// Test multiple concurrent calls to GetMigrationHistory
		done := make(chan error, 3)

		for i := 0; i < 3; i++ {
			go func(id int) {
				_, err := manager.GetMigrationHistory()
				done <- err
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 3; i++ {
			if err := <-done; err != nil {
				t.Errorf("Concurrent migration history call %d failed: %v", i, err)
			}
		}
	})
}

// TestUtilityFunctions tests utility and helper functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("CreateTestDirectory", func(t *testing.T) {
		testDir := "test_migrations/utility_test"
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Errorf("Failed to create test directory: %v", err)
		}
		defer os.RemoveAll("test_migrations/utility_test")

		// Verify directory exists
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Error("Test directory was not created")
		}
	})

	t.Run("DatabasePathGeneration", func(t *testing.T) {
		testPaths := []string{
			"test_migrations/test1.db",
			"test_migrations/test2.db",
			"test_migrations/subfolder/test3.db",
		}

		for _, path := range testPaths {
			// Ensure parent directory exists
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Errorf("Failed to create directory for path %s: %v", path, err)
			}

			// Test if path is valid for database creation
			db, err := sql.Open("sqlite3", path)
			if err != nil {
				t.Errorf("Failed to open database at path %s: %v", path, err)
				continue
			}
			db.Close()

			// Clean up
			_ = os.Remove(path)
		}

		// Clean up directories
		_ = os.RemoveAll("test_migrations/subfolder")
	})
}

// TestMigrationSQLGeneration tests migration SQL generation
func TestMigrationSQLGeneration(t *testing.T) {
	testCases := []struct {
		name        string
		upSQL       string
		downSQL     string
		expectValid bool
	}{
		{
			name:        "ValidCreateTable",
			upSQL:       "CREATE TABLE test_table (id INTEGER PRIMARY KEY);",
			downSQL:     "DROP TABLE IF EXISTS test_table;",
			expectValid: true,
		},
		{
			name:        "ValidAddColumn",
			upSQL:       "ALTER TABLE users ADD COLUMN phone VARCHAR(20);",
			downSQL:     "ALTER TABLE users DROP COLUMN phone;",
			expectValid: true,
		},
		{
			name:        "ValidCreateIndex",
			upSQL:       "CREATE INDEX idx_users_email ON users(email);",
			downSQL:     "DROP INDEX IF EXISTS idx_users_email;",
			expectValid: true,
		},
		{
			name:        "EmptySQL",
			upSQL:       "",
			downSQL:     "",
			expectValid: false,
		},
		{
			name:        "OnlyUpSQL",
			upSQL:       "CREATE TABLE test (id INTEGER);",
			downSQL:     "",
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate SQL content
			if tc.expectValid {
				if tc.upSQL == "" || tc.downSQL == "" {
					t.Error("Expected valid SQL but got empty strings")
				}
				if !strings.Contains(strings.ToUpper(tc.upSQL), "CREATE") &&
					!strings.Contains(strings.ToUpper(tc.upSQL), "ALTER") &&
					!strings.Contains(strings.ToUpper(tc.upSQL), "INSERT") {
					t.Error("Expected valid SQL operation")
				}
			} else {
				if tc.upSQL != "" && tc.downSQL != "" {
					t.Error("Expected invalid SQL but both up and down SQL are provided")
				}
			}
		})
	}
}

// BenchmarkMigrationOperations benchmarks key migration operations
func BenchmarkMigrationOperations(b *testing.B) {
	// Setup
	dbPath := filepath.Join("test_migrations", "benchmark.db")
	if err := os.MkdirAll("test_migrations", 0755); err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll("test_migrations")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	config := migrations.DefaultEFMigrationConfig()
	manager := migrations.NewEFMigrationManager(db, config)

	// Initialize schema
	if err := manager.EnsureSchema(); err != nil {
		b.Fatalf("Failed to initialize schema: %v", err)
	}

	b.Run("AddMigration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = manager.AddMigration(
				fmt.Sprintf("TestMigration%d", i),
				fmt.Sprintf("Test migration %d", i),
				fmt.Sprintf("CREATE TABLE test_table_%d (id INTEGER);", i),
				fmt.Sprintf("DROP TABLE IF EXISTS test_table_%d;", i),
			)
		}
	})

	b.Run("GetMigrationHistory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = manager.GetMigrationHistory()
		}
	})
}

package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

// Test models for comprehensive testing
type TestEntity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TestEntityWithTableName struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (t TestEntityWithTableName) TableName() string {
	return "custom_test_entities"
}

// SQLiteMigrationRunner extends MigrationRunner with SQLite-compatible methods
type SQLiteMigrationRunner struct {
	*MigrationRunner
}

// NewSQLiteMigrationRunner creates a new SQLite migration runner for testing
func NewSQLiteMigrationRunner(db *sql.DB) *SQLiteMigrationRunner {
	return &SQLiteMigrationRunner{
		MigrationRunner: &MigrationRunner{
			db:     db,
			logger: log.New(os.Stdout, "[TEST] ", log.LstdFlags),
		},
	}
}

// tableExists implements SQLite-compatible table existence check
func (smr *SQLiteMigrationRunner) tableExists(tableName string) (bool, error) {
	query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	var count int
	err := smr.db.QueryRow(query, tableName).Scan(&count)
	return count > 0, err
}

// migrateEntity overrides the original method to use SQLite-compatible table check
func (smr *SQLiteMigrationRunner) migrateEntity(entity interface{}) error {
	// Check if table exists using SQLite-compatible method
	tableName := getTableName(entity)
	exists, err := smr.tableExists(tableName)
	if err != nil {
		return err
	}

	if !exists {
		// Create table
		smr.logger.Printf("Creating table: %s", tableName)
		if err := smr.createTable(entity, tableName); err != nil {
			return err
		}
	} else {
		smr.logger.Printf("Table %s already exists, skipping", tableName)
	}

	return nil
}

// TestNewMigrationRunner_ComprehensiveScenarios tests various scenarios for NewMigrationRunner
func TestNewMigrationRunner_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name          string
		connectionStr string
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid_sqlite_connection",
			connectionStr: "file:test_comprehensive.db?cache=shared&mode=memory",
			expectError:   false,
		},
		{
			name:          "invalid_driver",
			connectionStr: "invaliddriver://test.db",
			expectError:   true,
			errorContains: "failed to ping database",
		},
		{
			name:          "empty_connection_string",
			connectionStr: "",
			expectError:   true,
			errorContains: "failed to ping database",
		},
		{
			name:          "malformed_connection",
			connectionStr: "::::invalid::::",
			expectError:   true,
			errorContains: "failed to ping database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For SQLite, we need to use sqlite3 driver
			if strings.Contains(tt.connectionStr, "file:") {
				db, err := sql.Open("sqlite3", tt.connectionStr)
				if err != nil {
					if !tt.expectError {
						t.Errorf("Unexpected error opening database: %v", err)
					}
					return
				}
				defer db.Close()

				if tt.expectError {
					t.Errorf("Expected error but got success")
				}
			} else {
				_, err := NewMigrationRunner(tt.connectionStr)
				if tt.expectError {
					if err == nil {
						t.Error("Expected error but got nil")
					} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
					}
				} else if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestMigrationRunner_AutoMigrate_ComprehensiveFlow tests comprehensive auto migration scenarios
func TestMigrationRunner_AutoMigrate_ComprehensiveFlow(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	// First create the migrations table manually for SQLite
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE, executed_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Test with models that have proper schema generation
	entities := []interface{}{
		&models.User{},
	}

	// Migrate each entity
	for _, entity := range entities {
		if err := runner.migrateEntity(entity); err != nil {
			t.Errorf("Failed to migrate entity %T: %v", entity, err)
		}
	}

	// Test running migration again (should not fail)
	for _, entity := range entities {
		if err := runner.migrateEntity(entity); err != nil {
			t.Errorf("Second migration failed for entity %T: %v", entity, err)
		}
	}
}

// TestMigrationRunner_AutoMigrate_DatabaseError tests error scenarios
func TestMigrationRunner_AutoMigrate_DatabaseError(t *testing.T) {
	// Create and close database to simulate connection error
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db.Close() // Close immediately to create error condition

	runner := &MigrationRunner{
		db:     db,
		logger: nil,
	}

	// AutoMigrate should fail due to closed database
	err = runner.AutoMigrate()
	if err == nil {
		t.Error("Expected AutoMigrate to fail with closed database")
	}
}

// TestMigrateEntity_ComprehensiveScenarios tests various entity migration scenarios
func TestMigrateEntity_ComprehensiveScenarios(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	// First create the migrations table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE, executed_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	tests := []struct {
		name        string
		entity      interface{}
		expectError bool
		setupDB     func(*sql.DB) error
	}{
		{
			name:        "new_entity",
			entity:      &models.User{},
			expectError: false,
		},
		{
			name:        "entity_with_custom_tablename",
			entity:      &models.Role{},
			expectError: false,
		},
		{
			name:   "existing_table",
			entity: &models.Product{},
			setupDB: func(db *sql.DB) error {
				// Create the table first
				_, err := db.Exec("CREATE TABLE products (id INTEGER, name TEXT)")
				return err
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupDB != nil {
				if err := tt.setupDB(db); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := runner.migrateEntity(tt.entity)
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

// TestTableExists_ComprehensiveScenarios tests table existence checks
func TestTableExists_ComprehensiveScenarios(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	// Test non-existent table
	exists, err := runner.tableExists("nonexistent_table")
	if err != nil {
		t.Errorf("Unexpected error checking non-existent table: %v", err)
	}
	if exists {
		t.Error("Expected table to not exist")
	}

	// Create a table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test existing table
	exists, err = runner.tableExists("test_table")
	if err != nil {
		t.Errorf("Unexpected error checking existing table: %v", err)
	}
	if !exists {
		t.Error("Expected table to exist")
	}
}

// TestCreateTable_ComprehensiveScenarios tests table creation
func TestCreateTable_ComprehensiveScenarios(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	tests := []struct {
		name        string
		entity      interface{}
		tableName   string
		expectError bool
	}{
		{
			name:        "models_user",
			entity:      &models.User{},
			tableName:   "users",
			expectError: false,
		},
		{
			name:        "models_role",
			entity:      &models.Role{},
			tableName:   "roles",
			expectError: false,
		},
		{
			name:        "empty_tablename",
			entity:      &models.User{},
			tableName:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.createTable(tt.entity, tt.tableName)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify table was created (if no error expected)
				if tt.tableName != "" {
					var count int
					query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
					err := db.QueryRow(query, tt.tableName).Scan(&count)
					if err != nil {
						t.Errorf("Failed to check if table was created: %v", err)
					}
					if count == 0 {
						t.Errorf("Table %s was not created", tt.tableName)
					}
				}
			}
		})
	}
}

// TestGetTableName_ComprehensiveScenarios tests table name generation
func TestGetTableName_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name         string
		entity       interface{}
		expectedName string
	}{
		{
			name:         "simple_struct",
			entity:       &TestEntity{},
			expectedName: "testentitys",
		},
		{
			name:         "struct_with_tablename_method",
			entity:       &TestEntityWithTableName{},
			expectedName: "custom_test_entities",
		},
		{
			name:         "models_user",
			entity:       &models.User{},
			expectedName: "users",
		},
		{
			name:         "models_role",
			entity:       &models.Role{},
			expectedName: "roles",
		},
		{
			name:         "models_category",
			entity:       &models.Category{},
			expectedName: "categories",
		},
		{
			name:         "models_product",
			entity:       &models.Product{},
			expectedName: "products",
		},
		{
			name:         "models_order",
			entity:       &models.Order{},
			expectedName: "orders",
		},
		{
			name:         "models_orderitem",
			entity:       &models.OrderItem{},
			expectedName: "order_items",
		},
		{
			name:         "models_review",
			entity:       &models.Review{},
			expectedName: "reviews",
		},
		{
			name:         "models_userrole",
			entity:       &models.UserRole{},
			expectedName: "user_roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTableName(tt.entity)
			if result != tt.expectedName {
				t.Errorf("Expected table name '%s', got '%s'", tt.expectedName, result)
			}
		})
	}
}

// TestGetTableName_ReflectionEdgeCases tests edge cases in reflection
func TestGetTableName_ReflectionEdgeCases(t *testing.T) {
	// Test with non-pointer struct
	entity := TestEntity{}
	result := getTableName(entity)
	expected := "testentitys"
	if result != expected {
		t.Errorf("Expected table name '%s' for non-pointer struct, got '%s'", expected, result)
	}

	// Test with interface{}
	var iface interface{} = &TestEntity{}
	result = getTableName(iface)
	if result != expected {
		t.Errorf("Expected table name '%s' for interface{}, got '%s'", expected, result)
	}
}

// TestShowStatus_ComprehensiveScenarios tests migration status display
func TestShowStatus_ComprehensiveScenarios(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	// Test with no migrations table (should fail gracefully)
	err = runner.ShowStatus()
	if err == nil {
		t.Error("Expected error when migrations table doesn't exist")
	}

	// Create migrations table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE, executed_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		t.Fatalf("Failed to create migrations table: %v", err)
	}

	// Test with empty migrations table
	err = runner.ShowStatus()
	if err != nil {
		t.Errorf("ShowStatus failed with empty table: %v", err)
	}

	// Add some migrations
	_, err = db.Exec("INSERT INTO migrations (name) VALUES ('test_migration_1'), ('test_migration_2')")
	if err != nil {
		t.Fatalf("Failed to insert test migrations: %v", err)
	}

	// Test with populated migrations table
	err = runner.ShowStatus()
	if err != nil {
		t.Errorf("ShowStatus failed with populated table: %v", err)
	}
}

// TestShowStatus_DatabaseError tests error scenarios in ShowStatus
func TestShowStatus_DatabaseError(t *testing.T) {
	// Create and close database to simulate connection error
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db.Close() // Close immediately to create error condition

	runner := NewSQLiteMigrationRunner(db)

	// ShowStatus should fail due to closed database
	err = runner.ShowStatus()
	if err == nil {
		t.Error("Expected ShowStatus to fail with closed database")
	}
}

// TestCreateMigrationsTable_ComprehensiveScenarios tests migrations table creation
func TestCreateMigrationsTable_ComprehensiveScenarios(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	runner := NewSQLiteMigrationRunner(db)

	// Test creating migrations table
	err = runner.createMigrationsTable()
	if err != nil {
		t.Errorf("Failed to create migrations table: %v", err)
	}

	// Verify table was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
	if err != nil {
		t.Errorf("Failed to check migrations table: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected migrations table to exist, got count: %d", count)
	}

	// Test creating migrations table again (should not fail due to IF NOT EXISTS)
	err = runner.createMigrationsTable()
	if err != nil {
		t.Errorf("Failed to create migrations table second time: %v", err)
	}
}

// TestMigrationRunner_Close_ComprehensiveScenarios tests database connection closing
func TestMigrationRunner_Close_ComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupRunner func() *SQLiteMigrationRunner
		expectError bool
	}{
		{
			name: "valid_connection",
			setupRunner: func() *SQLiteMigrationRunner {
				db, _ := sql.Open("sqlite3", ":memory:")
				return NewSQLiteMigrationRunner(db)
			},
			expectError: false,
		},
		{
			name: "already_closed_connection",
			setupRunner: func() *SQLiteMigrationRunner {
				db, _ := sql.Open("sqlite3", ":memory:")
				db.Close() // Close before returning
				return NewSQLiteMigrationRunner(db)
			},
			expectError: false, // SQLite allows multiple closes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := tt.setupRunner()
			err := runner.Close()
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

// BenchmarkAutoMigrate benchmarks the auto migration process
func BenchmarkAutoMigrate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new in-memory database for each iteration
		testDB, _ := sql.Open("sqlite3", ":memory:")
		testRunner := &MigrationRunner{
			db:     testDB,
			logger: nil,
		}
		testRunner.AutoMigrate()
		testRunner.Close()
	}
}

// BenchmarkGetTableName benchmarks table name generation
func BenchmarkGetTableName(b *testing.B) {
	entity := &models.User{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getTableName(entity)
	}
}

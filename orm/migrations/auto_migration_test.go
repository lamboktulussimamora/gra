package migrations

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Test error message constants
const (
	errFailedToCreateMigrationsTable = "Failed to create migrations table: %v"
	errFailedToCheckTableExistence   = "Failed to check table existence: %v"
)

// setupAutoMigrationTest uses the new multi-database test setup
func setupAutoMigrationTest(t *testing.T) (*AutoMigrator, *sql.DB, func()) {
	return SetupAutoMigrationTestMultiDB(t)
}

func TestNewAutoMigrator(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	if migrator == nil {
		t.Fatal("Expected auto migrator to be created")
	}
	if migrator.db != db {
		t.Error("Expected database to be set correctly")
	}
	if migrator.ctx == nil {
		t.Error("Expected context to be set")
	}
	if migrator.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestSetLogger(t *testing.T) {
	migrator, _, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	called := false
	customLogger := func(format string, args ...interface{}) {
		called = true
	}

	migrator.SetLogger(customLogger)
	migrator.logger("test message")

	if !called {
		t.Error("Expected custom logger to be called")
	}
}

func TestCreateMigrationsTable(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	err := migrator.createMigrationsTable()
	if err != nil {
		t.Fatalf(errFailedToCreateMigrationsTable, err)
	}

	// Verify table was created using database-agnostic check
	exists, err := CheckTableExists(db, "__migrations")
	if err != nil {
		t.Fatalf(errFailedToCheckTableExistence, err)
	}
	if !exists {
		t.Error("Expected __migrations table to be created")
	}
}

func TestMigrateModels(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Test with single model
	err := migrator.MigrateModels(&AutoTestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Verify migrations table was created
	exists, err := CheckTableExists(db, "__migrations")
	if err != nil {
		t.Fatalf("Failed to check migrations table: %v", err)
	}
	if !exists {
		t.Error("Expected __migrations table to be created")
	}

	// Verify model table was created
	exists, err = CheckTableExists(db, "auto_test_user")
	if err != nil {
		t.Fatalf(errFailedToCheckTableExistence, err)
	}
	if !exists {
		t.Error("Expected auto_test_user table to be created")
	}

	// Test with multiple models
	err = migrator.MigrateModels(&AutoTestUser{}, &AutoTestProduct{})
	if err != nil {
		t.Fatalf("Failed to migrate multiple models: %v", err)
	}

	// Verify second model table was created
	exists, err = CheckTableExists(db, "auto_test_product")
	if err != nil {
		t.Fatalf(errFailedToCheckTableExistence, err)
	}
	if !exists {
		t.Error("Expected auto_test_product table to be created")
	}
}

func TestMigrateModel(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Create migrations table first
	err := migrator.createMigrationsTable()
	if err != nil {
		t.Fatalf(errFailedToCreateMigrationsTable, err)
	}

	// Test migrating a model
	err = migrator.migrateModel(&AutoTestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate model: %v", err)
	}

	// Verify table was created using database-agnostic check
	exists, err := CheckTableExists(db, "auto_test_user")
	if err != nil {
		t.Fatalf(errFailedToCheckTableExistence, err)
	}
	if !exists {
		t.Error("Expected auto_test_user table to be created")
	}
}

func TestGetTableName(t *testing.T) {
	migrator, _, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	tests := []struct {
		model    interface{}
		expected string
	}{
		{&AutoTestUser{}, "auto_test_user"},
		{&AutoTestProduct{}, "auto_test_product"},
	}

	for _, test := range tests {
		tableName := migrator.getTableName(test.model)
		if tableName != test.expected {
			t.Errorf("Expected table name %s, got %s", test.expected, tableName)
		}
	}
}

func TestGetCurrentTableColumns(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Create a test table first
	createTableSQL := `
		CREATE TABLE get_columns_test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	columns, err := migrator.getCurrentTableColumns("get_columns_test_table")
	if err != nil {
		t.Fatalf("Failed to get columns: %v", err)
	}

	expectedColumns := []string{"id", "name", "email", "created_at"}
	if len(columns) < len(expectedColumns) {
		t.Fatalf("Expected at least %d columns, got %d", len(expectedColumns), len(columns))
	}

	for _, expected := range expectedColumns {
		found := false
		for columnName := range columns {
			if columnName == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column %s not found", expected)
		}
	}
}

func TestCreateIndexes(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Begin a transaction for testing
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create a test table first
	createTableSQL := `
		CREATE TABLE create_indexes_test_table (
			id INTEGER PRIMARY KEY,
			email TEXT,
			name TEXT,
			created_at TIMESTAMP
		)
	`
	_, err = tx.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test model type for creating indexes (using AutoTestUser as template)
	userType := reflect.TypeOf(AutoTestUser{})

	// Test createIndexes method with transaction
	err = migrator.createIndexes(tx, "create_indexes_test_table", userType)
	if err != nil {
		t.Fatalf("Failed to create indexes: %v", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Note: Index verification would be database-specific and complex
	// The fact that no error occurred indicates success
}

func TestModelFieldMapping(t *testing.T) {
	migrator, _, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Test with AutoTestUser
	user := &AutoTestUser{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	tableName := migrator.getTableName(user)
	if tableName != "auto_test_user" {
		t.Errorf("Expected table name 'auto_test_user', got '%s'", tableName)
	}

	// Test with AutoTestProduct
	product := &AutoTestProduct{
		ID:          1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CreatedAt:   time.Now(),
	}

	tableName = migrator.getTableName(product)
	if tableName != "auto_test_product" {
		t.Errorf("Expected table name 'auto_test_product', got '%s'", tableName)
	}
}

func TestMigrationIdempotency(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Run migration multiple times - should not fail
	for i := 0; i < 3; i++ {
		err := migrator.MigrateModels(&AutoTestUser{})
		if err != nil {
			t.Fatalf("Migration failed on iteration %d: %v", i+1, err)
		}

		// Verify table still exists
		exists, err := CheckTableExists(db, "auto_test_user")
		if err != nil {
			t.Fatalf("Failed to check table existence on iteration %d: %v", i+1, err)
		}
		if !exists {
			t.Errorf("Table should exist after iteration %d", i+1)
		}
	}
}

// TestMultiDatabaseCompatibility tests that the same migration works on different databases
func TestMultiDatabaseCompatibility(t *testing.T) {
	DatabaseDriverSpecificTest(t, func(t *testing.T, driver string, db *sql.DB) {
		t.Logf("Testing with database driver: %s", driver)

		migrator := SetupAutoMigrationTestWithDB(t, db)

		// Test migration
		err := migrator.MigrateModels(&AutoTestUser{}, &AutoTestProduct{})
		if err != nil {
			t.Fatalf("Failed to migrate models with %s: %v", driver, err)
		}

		// Verify tables were created
		tables := []string{"auto_test_user", "auto_test_product", "__migrations"}
		for _, table := range tables {
			exists, err := CheckTableExists(db, table)
			if err != nil {
				t.Fatalf("Failed to check table %s with %s: %v", table, driver, err)
			}
			if !exists {
				t.Errorf("Table %s should exist with %s", table, driver)
			}
		}
	})
}

func TestAutoMigrationErrorHandling(t *testing.T) {
	// Test with nil database - this should panic, so we need to recover
	ctx := &dbcontext.EnhancedDbContext{}
	migrator := NewAutoMigrator(ctx, nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil database")
		}
	}()

	// This should panic
	migrator.MigrateModels(&AutoTestUser{})
}

func TestAutoMigrationInvalidModel(t *testing.T) {
	// Test with invalid model
	migrator, _, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Test with nil model - this should panic, so we need to recover
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil model")
		}
	}()

	// This should panic
	migrator.MigrateModels(nil)
}

func TestLargeDatasetMigration(t *testing.T) {
	migrator, db, cleanup := setupAutoMigrationTest(t)
	defer cleanup()

	// Create many models to test performance
	models := make([]interface{}, 0, 10)
	for i := 0; i < 10; i++ {
		models = append(models, &AutoTestUser{})
		models = append(models, &AutoTestProduct{})
	}

	start := time.Now()
	err := migrator.MigrateModels(models...)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to migrate large dataset: %v", err)
	}

	t.Logf("Large dataset migration completed in %v", duration)

	// Verify core tables exist (duplicates should be handled gracefully)
	tables := []string{"auto_test_user", "auto_test_product", "__migrations"}
	for _, table := range tables {
		exists, err := CheckTableExists(db, table)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("Table %s should exist after large dataset migration", table)
		}
	}
}

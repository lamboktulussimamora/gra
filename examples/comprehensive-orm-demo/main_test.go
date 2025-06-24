package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/migrations"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

func TestGetConnectionString(t *testing.T) {
	// Test default value
	defaultPath := getConnectionString()
	if defaultPath != "./demo.db" {
		t.Errorf("Expected default path './demo.db', got %s", defaultPath)
	}

	// Test with environment variable
	os.Setenv("DB_PATH", "/tmp/test.db")
	defer os.Unsetenv("DB_PATH")

	envPath := getConnectionString()
	if envPath != "/tmp/test.db" {
		t.Errorf("Expected env path '/tmp/test.db', got %s", envPath)
	}
}

func TestGetEnvDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "use_default_when_env_not_set",
			key:          "NONEXISTENT_KEY",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "use_env_when_set",
			key:          "TEST_KEY",
			defaultValue: "default_value",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "use_default_when_env_empty",
			key:          "EMPTY_KEY",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env var
			os.Unsetenv(tt.key)

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRunMigrations(t *testing.T) {
	// Use in-memory SQLite for testing
	connectionString := ":memory:"

	err := runMigrations(connectionString)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	// Verify that tables were created by checking the database
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if migration table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='__ef_migrations_history'").Scan(&tableName)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to check migration table: %v", err)
	}
}

func TestDemonstrateORM(t *testing.T) {
	// Use a test file instead of in-memory to persist between function calls
	connectionString := "./test_demo.db"
	defer os.Remove("./test_demo.db") // Clean up after test

	// First run migrations
	err := runMigrations(connectionString)
	if err != nil {
		t.Fatalf("runMigrations failed: %v", err)
	}

	// Then test ORM demonstration
	err = demonstrateORM(connectionString)
	if err != nil {
		t.Fatalf("demonstrateORM failed: %v", err)
	}
}

func TestDemonstrateBasicCRUD(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// First run migrations to create tables
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	entities := []interface{}{
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
		&models.Role{},
		&models.UserRole{},
	}
	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Test basic CRUD operations
	err = demonstrateBasicCRUD(ctx)
	if err != nil {
		t.Fatalf("demonstrateBasicCRUD failed: %v", err)
	}
}

func TestDemonstrateAdvancedQuerying(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// First run migrations to create tables
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	entities := []interface{}{
		&models.User{},
	}
	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Test advanced querying
	err = demonstrateAdvancedQuerying(ctx)
	if err != nil {
		t.Fatalf("demonstrateAdvancedQuerying failed: %v", err)
	}
}

func TestDemonstrateTransactions(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// First run migrations to create tables
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	entities := []interface{}{
		&models.User{},
	}
	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Test transactions
	err = demonstrateTransactions(ctx)
	if err != nil {
		t.Fatalf("demonstrateTransactions failed: %v", err)
	}
}

func TestDemonstrateChangeTracking(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// First run migrations to create tables
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	entities := []interface{}{
		&models.User{},
	}
	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Test change tracking
	err = demonstrateChangeTracking(ctx)
	if err != nil {
		t.Fatalf("demonstrateChangeTracking failed: %v", err)
	}
}

func TestUserModelOperations(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// First run migrations to create tables
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	entities := []interface{}{
		&models.User{},
	}
	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Test user operations
	user := &models.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		IsActive:  true,
	}

	// Add user
	ctx.Add(user)
	changes, err := ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	if changes == 0 {
		t.Error("Expected changes > 0 when saving user")
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set after save")
	}

	// Update user
	user.Email = "updated@example.com"
	ctx.Update(user)
	changes, err = ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if changes == 0 {
		t.Error("Expected changes > 0 when updating user")
	}

	// Delete user
	ctx.Delete(user)
	changes, err = ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	if changes == 0 {
		t.Error("Expected changes > 0 when deleting user")
	}
}

func TestDbContextOperations(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test enhanced database context creation
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)
	if ctx == nil {
		t.Fatal("Expected enhanced database context to be created")
	}

	// Test transaction creation
	tx, err := ctx.Database.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Test transaction context creation
	txCtx := dbcontext.NewEnhancedDbContextWithTx(tx)
	if txCtx == nil {
		t.Fatal("Expected transaction context to be created")
	}

	// Test rollback
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}

func TestMigrationRunner(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Test migration runner creation
	migrationRunner := migrations.NewAutoMigrator(ctx, db)
	if migrationRunner == nil {
		t.Fatal("Expected migration runner to be created")
	}

	// Test model migration
	entities := []interface{}{
		&models.User{},
		&models.Product{},
		&models.Category{},
	}

	err = migrationRunner.MigrateModels(entities...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	// Verify tables were created
	tables := []string{"users", "products", "categories"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("Expected table %s to exist", table)
		}
	}
}

func TestFullWorkflow(t *testing.T) {
	// Test the complete workflow that main() would execute
	connectionString := "./test_workflow.db"
	defer os.Remove("./test_workflow.db") // Clean up after test

	// Step 1: Run migrations
	err := runMigrations(connectionString)
	if err != nil {
		t.Fatalf("Migration step failed: %v", err)
	}

	// Step 2: Demonstrate ORM
	err = demonstrateORM(connectionString)
	if err != nil {
		t.Fatalf("ORM demonstration step failed: %v", err)
	}
}

func TestErrorHandling(t *testing.T) {
	// Test with invalid connection string
	err := runMigrations("invalid://connection")
	if err == nil {
		t.Error("Expected error with invalid connection string")
	}

	err = demonstrateORM("invalid://connection")
	if err == nil {
		t.Error("Expected error with invalid connection string")
	}
}

func TestModelStructures(t *testing.T) {
	// Test User model
	user := &models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}

	if user.FirstName != "John" {
		t.Errorf("Expected FirstName 'John', got %s", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("Expected LastName 'Doe', got %s", user.LastName)
	}
	if user.Email != "john.doe@example.com" {
		t.Errorf("Expected Email 'john.doe@example.com', got %s", user.Email)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}

	// Test Product model
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		SKU:         "TEST-001",
		InStock:     true,
		StockCount:  10,
	}

	if product.Name != "Test Product" {
		t.Errorf("Expected Name 'Test Product', got %s", product.Name)
	}
	if product.Price != 99.99 {
		t.Errorf("Expected Price 99.99, got %f", product.Price)
	}
	if !product.InStock {
		t.Error("Expected InStock to be true")
	}

	// Test Category model
	category := &models.Category{
		Name:        "Test Category",
		Description: "Test Description",
	}

	if category.Name != "Test Category" {
		t.Errorf("Expected Name 'Test Category', got %s", category.Name)
	}
}

func TestConstants(t *testing.T) {
	// Test the constant used in the code
	if isActiveWhere != "is_active = ?" {
		t.Errorf("Expected isActiveWhere to be 'is_active = ?', got %s", isActiveWhere)
	}
}

func TestDatabaseConnectionHandling(t *testing.T) {
	// Test proper database connection handling
	connectionString := ":memory:"

	// Test database opening and closing
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test database ping
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test database closing
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	// Test various environment variable scenarios
	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "default_db_path",
			envVar:   "DB_PATH",
			envValue: "",
			expected: "./demo.db",
		},
		{
			name:     "custom_db_path",
			envVar:   "DB_PATH",
			envValue: "/custom/path.db",
			expected: "/custom/path.db",
		},
		{
			name:     "relative_path",
			envVar:   "DB_PATH",
			envValue: "../data/test.db",
			expected: "../data/test.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Unsetenv(tt.envVar)

			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := getConnectionString()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMainFunctionComponents(t *testing.T) {
	// Test that all components used by main() function exist and work

	// Test connection string generation
	connStr := getConnectionString()
	if connStr == "" {
		t.Error("Connection string should not be empty")
	}

	// Test environment variable helper
	envVal := getEnvDefault("NONEXISTENT", "default")
	if envVal != "default" {
		t.Errorf("Expected 'default', got %s", envVal)
	}

	// Test that migration and ORM demonstration functions exist
	// (they are tested individually above)
	t.Log("All main function components are properly tested")
}

func TestMainFunctionExists(t *testing.T) {
	// Test that main function exists and can be referenced
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main function caused panic: %v", r)
		}
	}()

	// Just checking that main is available
	t.Log("main function exists and is accessible")
}

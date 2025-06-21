package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test models for SimpleMigrator tests
type SimpleMigratorTestUser struct {
	ID        uint      `db:"id" migration:"type:integer;primary_key:true;auto_increment:true"`
	Name      string    `db:"name" migration:"type:varchar(100);not_null:true"`
	Email     string    `db:"email" migration:"type:varchar(255);unique:true"`
	Age       *int      `db:"age" migration:"type:integer"`
	CreatedAt time.Time `db:"created_at" migration:"type:datetime;default:CURRENT_TIMESTAMP"`
}

func (SimpleMigratorTestUser) TableName() string {
	return "simple_test_users"
}

type SimpleMigratorTestProfile struct {
	ID     uint   `db:"id" migration:"type:integer;primary_key:true;auto_increment:true"`
	UserID uint   `db:"user_id" migration:"type:integer;not_null:true"`
	Bio    string `db:"bio" migration:"type:text"`
}

func (SimpleMigratorTestProfile) TableName() string {
	return "simple_test_profiles"
}

func setupSimpleMigratorTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "simple_test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestNewSimpleMigrator(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, SQLite, "./migrations")

	if migrator == nil {
		t.Fatal("Expected migrator to be created, got nil")
	}

	if migrator.db != db {
		t.Error("Expected database to be set correctly")
	}

	if migrator.driver != SQLite {
		t.Errorf("Expected driver to be SQLite, got %v", migrator.driver)
	}

	if migrator.registry == nil {
		t.Error("Expected registry to be initialized")
	}

	if migrator.migrationsDir != "./migrations" {
		t.Errorf("Expected migrations directory to be './migrations', got %s", migrator.migrationsDir)
	}
}

func TestDbSet(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, SQLite, "./migrations")

	// Register a model
	migrator.DbSet(&SimpleMigratorTestUser{})

	// Verify model is registered
	models := migrator.GetRegisteredModels()
	if len(models) != 1 {
		t.Errorf("Expected 1 registered model, got %d", len(models))
	}

	if _, exists := models["simple_test_users"]; !exists {
		t.Error("Expected SimpleMigratorTestUser model to be registered")
	}
}

func TestGetRegisteredModels(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, SQLite, "./migrations")

	// Initially should be empty
	models := migrator.GetRegisteredModels()
	if len(models) != 0 {
		t.Errorf("Expected 0 registered models initially, got %d", len(models))
	}

	// Register multiple models
	migrator.DbSet(&SimpleMigratorTestUser{})
	migrator.DbSet(&SimpleMigratorTestProfile{})

	// Verify both models are registered
	models = migrator.GetRegisteredModels()
	if len(models) != 2 {
		t.Errorf("Expected 2 registered models, got %d", len(models))
	}

	expectedTables := []string{"simple_test_users", "simple_test_profiles"}
	for _, table := range expectedTables {
		if _, exists := models[table]; !exists {
			t.Errorf("Expected table %s to be registered", table)
		}
	}
}

func TestTableExists(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, SQLite, "./migrations")

	// Test with non-existent table
	exists, err := migrator.TableExists("non_existent_table")
	if err != nil {
		t.Fatalf("Unexpected error checking table existence: %v", err)
	}
	if exists {
		t.Error("Expected table 'non_existent_table' to not exist")
	}

	// Create a test table
	_, err = db.Exec(`CREATE TABLE test_table (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test with existing table
	exists, err = migrator.TableExists("test_table")
	if err != nil {
		t.Fatalf("Unexpected error checking table existence: %v", err)
	}
	if !exists {
		t.Error("Expected table 'test_table' to exist")
	}
}

func TestTableExistsUnsupportedDriver(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, DatabaseDriver("unsupported"), "./migrations")

	_, err := migrator.TableExists("any_table")
	if err == nil {
		t.Error("Expected error for unsupported database driver")
	}

	expectedError := "unsupported database driver: unsupported"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGenerateCreateTableSQL(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	migrator := NewSimpleMigrator(db, SQLite, "./migrations")

	// Register model and get snapshot
	migrator.DbSet(&SimpleMigratorTestUser{})
	models := migrator.GetRegisteredModels()
	snapshot := models["simple_test_users"]

	if snapshot == nil {
		t.Fatal("Expected to get model snapshot for simple_test_users")
	}

	// Generate SQL
	sql := migrator.GenerateCreateTableSQL(snapshot)

	// Basic checks
	if sql == "" {
		t.Error("Expected non-empty SQL")
	}

	// Should contain table name
	if !strings.Contains(sql, "simple_test_users") {
		t.Error("Expected SQL to contain table name 'simple_test_users'")
	}

	// Should contain CREATE TABLE
	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("Expected SQL to contain 'CREATE TABLE'")
	}

	// Should contain PRIMARY KEY
	if !strings.Contains(sql, "PRIMARY KEY") {
		t.Error("Expected SQL to contain 'PRIMARY KEY'")
	}

	// Should contain some expected columns
	expectedColumns := []string{"id", "name", "email"}
	for _, col := range expectedColumns {
		if !strings.Contains(sql, col) {
			t.Errorf("Expected SQL to contain column '%s'", col)
		}
	}
}

// TestCreateInitialMigration tests the creation of initial migration files
func TestCreateInitialMigration(t *testing.T) {
	db, cleanup := setupSimpleMigratorTestDB(t)
	defer cleanup()

	// Create temporary migrations directory
	tempDir := t.TempDir()
	migrator := NewSimpleMigrator(db, SQLite, tempDir)

	// Register models
	migrator.DbSet(&SimpleMigratorTestUser{})
	migrator.DbSet(&SimpleMigratorTestProfile{})

	// Create initial migration
	_, err := migrator.CreateInitialMigration("InitialMigration")
	if err != nil {
		t.Fatalf("Failed to create initial migration: %v", err)
	}

	// Check if migration file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected migration file to be created")
	}

	// Verify file name pattern
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), "InitialMigration") && strings.Contains(file.Name(), ".sql") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find migration file with 'InitialMigration' in the name")
	}
}

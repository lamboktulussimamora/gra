package migrations

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	_ "github.com/mattn/go-sqlite3"
)

// Test model structures for migration testing (with unique names)
type MigrationTestUser struct {
	ID       int    `db:"id" gorm:"primary_key" json:"id"`
	Name     string `db:"name" gorm:"size:255;not null" json:"name"`
	Email    string `db:"email" gorm:"size:255;unique;not null" json:"email"`
	Age      int    `db:"age" json:"age"`
	IsActive bool   `db:"is_active" gorm:"default:true" json:"is_active"`
}

type MigrationTestProduct struct {
	ID          int     `db:"id" gorm:"primary_key" json:"id"`
	Name        string  `db:"name" gorm:"size:255;not null" json:"name"`
	Price       float64 `db:"price" gorm:"type:decimal(10,2);not null" json:"price"`
	Description string  `db:"description" gorm:"type:text" json:"description"`
	UserID      int     `db:"user_id" gorm:"not null" json:"user_id"`
}

type MigrationTestCategory struct {
	ID          int    `db:"id" gorm:"primary_key" json:"id"`
	Name        string `db:"name" gorm:"size:255;unique;not null" json:"name"`
	Description string `db:"description" gorm:"type:text" json:"description"`
}

// EmbeddedProfile test for embedded structs
type MigrationEmbeddedProfile struct {
	Bio     string `db:"bio" json:"bio"`
	Website string `db:"website" json:"website"`
}

type MigrationTestUserWithEmbedded struct {
	ID                       int    `db:"id" gorm:"primary_key" json:"id"`
	Name                     string `db:"name" gorm:"size:255;not null" json:"name"`
	MigrationEmbeddedProfile        // Embedded struct (anonymous)
}

// TestAutoMigratorComprehensive tests AutoMigrator with comprehensive scenarios
func TestAutoMigratorComprehensive(t *testing.T) {
	t.Run("new auto migrator creation", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		if migrator == nil {
			t.Error("Expected non-nil migrator")
			return
		}
		if migrator.ctx != ctx {
			t.Error("Expected context to be set correctly")
		}
		if migrator.db != db {
			t.Error("Expected database to be set correctly")
		}
	})
}

// TestSetLoggerComprehensive tests custom logger functionality
func TestSetLoggerComprehensive(t *testing.T) {
	t.Run("custom logger setting", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		var loggedMessages []string
		customLogger := func(format string, args ...interface{}) {
			loggedMessages = append(loggedMessages, format)
		}

		migrator.SetLogger(customLogger)

		// Test that custom logger is called
		migrator.logger("test message")
		if len(loggedMessages) != 1 || loggedMessages[0] != "test message" {
			t.Errorf("Expected custom logger to be called with 'test message', got %v", loggedMessages)
		}
	})
}

// TestCreateDatabaseComprehensive tests database creation functionality
func TestCreateDatabaseComprehensive(t *testing.T) {
	t.Run("create database with sqlite (always succeeds)", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.CreateDatabase("test_db")
		// SQLite doesn't support CREATE DATABASE, but should not error
		if err != nil {
			// Check if error is about unsupported syntax (acceptable for SQLite)
			if !strings.Contains(err.Error(), "syntax error") && !strings.Contains(err.Error(), "near \"IF\"") {
				t.Errorf("Unexpected error creating database: %v", err)
			}
		}
	})
}

// TestDropDatabaseComprehensive tests database dropping functionality
func TestDropDatabaseComprehensive(t *testing.T) {
	t.Run("drop database with sqlite (not supported)", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.DropDatabase("test_db")
		// SQLite doesn't support DROP DATABASE, expect error
		if err == nil {
			t.Error("Expected error when dropping database in SQLite")
		}
	})
}

// TestCreateMigrationsTableComprehensive tests migrations table creation
func TestCreateMigrationsTableComprehensive(t *testing.T) {
	t.Run("create migrations table successfully", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.createMigrationsTable()
		if err != nil {
			t.Errorf("Failed to create migrations table: %v", err)
		}

		// Verify table exists
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='__migrations'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to check migrations table existence: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected migrations table to exist, found %d tables", count)
		}
	})

	t.Run("create migrations table with database error", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		db.Close() // Close database to cause error

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.createMigrationsTable()
		if err == nil {
			t.Error("Expected error when creating migrations table with closed database")
		}
	})
}

// TestMigrateModelsComprehensive tests comprehensive model migration scenarios
func TestMigrateModelsComprehensive(t *testing.T) {
	t.Run("migrate single model successfully", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.MigrateModels(MigrationTestUser{})
		if err != nil {
			t.Errorf("Failed to migrate MigrationTestUser model: %v", err)
		}

		// Verify table was created with correct name
		var tables []string
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			t.Errorf("Failed to query tables: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				t.Errorf("Failed to scan table name: %v", err)
			}
			tables = append(tables, tableName)
		}

		// Check if a table was created (exact name might vary based on implementation)
		if len(tables) < 2 { // migrations table + our table
			t.Errorf("Expected at least 2 tables (migrations + model), found: %v", tables)
		}
	})

	t.Run("migrate multiple models successfully", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.MigrateModels(MigrationTestUser{}, MigrationTestProduct{}, MigrationTestCategory{})
		if err != nil {
			t.Errorf("Failed to migrate multiple models: %v", err)
		}

		// Verify tables were created
		var tables []string
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			t.Errorf("Failed to query tables: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				t.Errorf("Failed to scan table name: %v", err)
			}
			tables = append(tables, tableName)
		}

		// Should have migrations table + 3 model tables
		if len(tables) < 4 {
			t.Errorf("Expected at least 4 tables (migrations + 3 models), found: %v", tables)
		}
	})

	t.Run("migrate model with embedded struct", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.MigrateModels(MigrationTestUserWithEmbedded{})
		if err != nil {
			t.Errorf("Failed to migrate model with embedded struct: %v", err)
		}

		// Verify table was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migration_test_user_with_embedded'").Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table existence: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected migration_test_user_with_embedded table to exist, found %d tables", count)
		}
	})

	t.Run("migrate model with closed database", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		db.Close() // Close database to cause error

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		err = migrator.MigrateModels(MigrationTestUser{})
		if err == nil {
			t.Error("Expected error when migrating with closed database")
		}
	})
}

// TestGetTableNameComprehensive tests table name generation
func TestGetTableNameComprehensive(t *testing.T) {
	t.Run("get table name from struct", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test with different struct types - check that we get a non-empty name
		testModels := []interface{}{
			MigrationTestUser{},
			MigrationTestProduct{},
			MigrationTestCategory{},
		}

		for _, model := range testModels {
			tableName := migrator.getTableName(reflect.TypeOf(model))
			if tableName == "" {
				t.Errorf("Expected non-empty table name for %T", model)
			}
		}
	})

	t.Run("get table name with pointer", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test with pointer to struct
		user := &MigrationTestUser{}
		tableName := migrator.getTableName(reflect.TypeOf(user))
		if tableName == "" {
			t.Error("Expected non-empty table name for pointer to MigrationTestUser")
		}
	})
}

// TestToSnakeCaseComprehensive tests snake case conversion
func TestToSnakeCaseComprehensive(t *testing.T) {
	t.Run("convert various strings to snake case", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test basic snake case conversion (implementation might vary)
		testCases := []struct {
			input string
		}{
			{"TestUser"},
			{"TestProduct"},
			{"UserProfile"},
			{"ID"},
			{"test"},
			{"Test"},
			{"testUser"},
		}

		for _, tc := range testCases {
			result := migrator.toSnakeCase(tc.input)
			if result == "" {
				t.Errorf("Expected non-empty result for toSnakeCase(%s)", tc.input)
			}
			// Just check that it's a reasonable conversion (lowercase)
			if strings.ToLower(result) != result {
				t.Errorf("Expected lowercase result for toSnakeCase(%s), got %s", tc.input, result)
			}
		}
	})
}

// TestGenerateCreateTableSQLComprehensive tests SQL generation
func TestGenerateCreateTableSQLComprehensive(t *testing.T) {
	t.Run("generate create table SQL for various models", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test SQL generation for MigrationTestUser
		sql := migrator.generateCreateTableSQL("migration_test_users", reflect.TypeOf(MigrationTestUser{}))
		if sql == "" {
			t.Error("Expected non-empty SQL for MigrationTestUser")
		}
		if !strings.Contains(strings.ToUpper(sql), "CREATE TABLE") {
			t.Errorf("Expected SQL to contain CREATE TABLE, got: %s", sql)
		}
		if !strings.Contains(sql, "id") {
			t.Errorf("Expected SQL to contain id column, got: %s", sql)
		}
		if !strings.Contains(sql, "name") {
			t.Errorf("Expected SQL to contain name column, got: %s", sql)
		}
		if !strings.Contains(sql, "email") {
			t.Errorf("Expected SQL to contain email column, got: %s", sql)
		}

		// Test SQL generation for MigrationTestProduct
		sql = migrator.generateCreateTableSQL("migration_test_products", reflect.TypeOf(MigrationTestProduct{}))
		if sql == "" {
			t.Error("Expected non-empty SQL for MigrationTestProduct")
		}
		if !strings.Contains(strings.ToUpper(sql), "CREATE TABLE") {
			t.Errorf("Expected SQL to contain CREATE TABLE, got: %s", sql)
		}
	})
}

// TestGenerateColumnDefinitionComprehensive tests column definition generation
func TestGenerateColumnDefinitionComprehensive(t *testing.T) {
	t.Run("generate column definitions for various field types", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test different field types
		userType := reflect.TypeOf(MigrationTestUser{})

		// Test each field
		for i := 0; i < userType.NumField(); i++ {
			field := userType.Field(i)
			columnDef := migrator.generateColumnDefinition(field, "")

			if columnDef == "" {
				t.Errorf("Expected non-empty column definition for field %s", field.Name)
			}

			// Check that column definition contains some expected content
			if !strings.Contains(columnDef, strings.ToLower(field.Name)) && !strings.Contains(columnDef, migrator.toSnakeCase(field.Name)) {
				t.Errorf("Expected column definition to contain field name or snake_case version for %s, got: %s", field.Name, columnDef)
			}
		}
	})
}

// TestCalculateChecksumComprehensive tests checksum calculation
func TestCalculateChecksumComprehensive(t *testing.T) {
	t.Run("calculate checksum for model schemas", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Calculate checksum for MigrationTestUser
		schema := migrator.generateTableSchema(reflect.TypeOf(MigrationTestUser{}))
		checksum := migrator.calculateChecksum(schema)

		if checksum == "" {
			t.Error("Expected non-empty checksum")
		}

		// Verify same schema produces same checksum
		checksum2 := migrator.calculateChecksum(schema)
		if checksum != checksum2 {
			t.Error("Expected same schema to produce same checksum")
		}

		// Verify different schema produces different checksum
		schema2 := migrator.generateTableSchema(reflect.TypeOf(MigrationTestProduct{}))
		checksum3 := migrator.calculateChecksum(schema2)
		if checksum == checksum3 {
			t.Error("Expected different schemas to produce different checksums")
		}
	})
}

// TestGenerateTableSchemaComprehensive tests table schema generation
func TestGenerateTableSchemaComprehensive(t *testing.T) {
	t.Run("generate table schema for various models", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test schema generation for different models
		models := []interface{}{
			MigrationTestUser{},
			MigrationTestProduct{},
			MigrationTestCategory{},
		}

		for _, model := range models {
			schema := migrator.generateTableSchema(reflect.TypeOf(model))
			if schema == "" {
				t.Errorf("Expected non-empty schema for %T", model)
			}

			// Schema should contain some field information
			modelType := reflect.TypeOf(model)
			hasFields := false
			for i := 0; i < modelType.NumField(); i++ {
				field := modelType.Field(i)
				if field.Tag.Get("db") != "" {
					hasFields = true
					break
				}
			}
			if hasFields && !strings.Contains(schema, "id") {
				t.Errorf("Expected schema to contain at least some field information for %T, got: %s", model, schema)
			}
		}
	})
}

// TestProcessStructFieldsComprehensive tests struct field processing
func TestProcessStructFieldsComprehensive(t *testing.T) {
	t.Run("process struct fields for various models", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test processing struct fields
		userType := reflect.TypeOf(MigrationTestUser{})
		var fieldNames []string

		migrator.processStructFields(userType, func(field reflect.StructField, dbTag string) {
			fieldNames = append(fieldNames, dbTag)
		})

		if len(fieldNames) == 0 {
			t.Error("Expected non-empty fields list")
		}

		// Should have fields for all struct fields with db tags
		expectedFields := []string{"id", "name", "email", "age", "is_active"}
		if len(fieldNames) != len(expectedFields) {
			t.Errorf("Expected %d fields, got %d", len(expectedFields), len(fieldNames))
		}

		// Check that all expected fields are present
		for _, expectedField := range expectedFields {
			found := false
			for _, fieldName := range fieldNames {
				if fieldName == expectedField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected field %s to be present in processed fields", expectedField)
			}
		}
	})
}

// TestIsEmbeddedStructComprehensive tests embedded struct detection
func TestIsEmbeddedStructComprehensive(t *testing.T) {
	t.Run("detect embedded structs correctly", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Test with regular struct (should not be embedded)
		userType := reflect.TypeOf(MigrationTestUser{})
		for i := 0; i < userType.NumField(); i++ {
			field := userType.Field(i)
			if migrator.isEmbeddedStruct(field) {
				t.Errorf("Field %s should not be detected as embedded struct", field.Name)
			}
		}

		// Test with embedded struct
		embeddedType := reflect.TypeOf(MigrationTestUserWithEmbedded{})
		foundEmbedded := false
		for i := 0; i < embeddedType.NumField(); i++ {
			field := embeddedType.Field(i)
			if field.Anonymous && migrator.isEmbeddedStruct(field) {
				foundEmbedded = true
				break
			}
		}
		if !foundEmbedded {
			t.Error("Expected to find embedded struct in MigrationTestUserWithEmbedded")
		}
	})
}

// TestCreateIndexesComprehensive tests index creation
func TestCreateIndexesComprehensive(t *testing.T) {
	t.Run("create indexes with transaction", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		ctx := &dbcontext.EnhancedDbContext{}
		migrator := NewAutoMigrator(ctx, db)

		// Create a test table first
		_, err = db.Exec(`CREATE TABLE test_users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL
		)`)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		// Start a transaction for index creation
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		// Test index creation
		err = migrator.createIndexes(tx, "test_users", reflect.TypeOf(MigrationTestUser{}))
		if err != nil {
			t.Errorf("Failed to create indexes: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Errorf("Failed to commit transaction: %v", err)
		}
	})
}

package schema

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test entity
type TestUser struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

type TestProduct struct {
	ID          int64   `db:"id" migration:"primary_key,auto_increment"`
	Name        string  `db:"name" migration:"not_null,max_length:200"`
	Price       float64 `db:"price" migration:"not_null"`
	CategoryID  int64   `db:"category_id" migration:"foreign_key:categories.id"`
	Description string  `db:"description" migration:"type:TEXT"`
}

func TestDetectDatabaseDriver(t *testing.T) {
	// Test with SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close database: %v", closeErr)
		}
	}()

	driver := DetectDatabaseDriver(db)
	if driver != SQLite {
		t.Errorf("Expected SQLite driver, got %v", driver)
	}

	// Test with nil database
	driver = DetectDatabaseDriver(nil)
	if driver != PostgreSQL {
		t.Errorf("Expected PostgreSQL as default, got %v", driver)
	}
}

func TestDetectDatabaseDriverFromConnectionString(t *testing.T) {
	tests := []struct {
		driver   string
		expected DatabaseDriver
	}{
		{"postgres", PostgreSQL},
		{"postgresql", PostgreSQL},
		{"sqlite3", SQLite},
		{"sqlite", SQLite},
		{"mysql", MySQL},
		{"unknown", PostgreSQL}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			result := DetectDatabaseDriverFromConnectionString(tt.driver)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGenerateCreateTableSQL(t *testing.T) {
	user := &TestUser{}
	sql := GenerateCreateTableSQL(user, "users")

	if sql == "" {
		t.Error("Generated SQL should not be empty")
	}
	if !strings.Contains(sql, "CREATE TABLE") {
		t.Error("SQL should contain CREATE TABLE")
	}
	if !strings.Contains(sql, "users") {
		t.Error("SQL should contain table name 'users'")
	}
}

func TestGenerateCreateTableSQLForDriver(t *testing.T) {
	user := &TestUser{}

	// Test PostgreSQL
	pgSQL := GenerateCreateTableSQLForDriver(user, "users", PostgreSQL)
	if !strings.Contains(pgSQL, "CREATE TABLE") {
		t.Error("PostgreSQL SQL should contain CREATE TABLE")
	}
	if !strings.Contains(pgSQL, "SERIAL") {
		t.Error("PostgreSQL SQL should contain SERIAL for auto increment")
	}

	// Test SQLite
	sqliteSQL := GenerateCreateTableSQLForDriver(user, "users", SQLite)
	if !strings.Contains(sqliteSQL, "CREATE TABLE") {
		t.Error("SQLite SQL should contain CREATE TABLE")
	}
	if !strings.Contains(sqliteSQL, "AUTOINCREMENT") {
		t.Error("SQLite SQL should contain AUTOINCREMENT")
	}

	// Test MySQL
	mysqlSQL := GenerateCreateTableSQLForDriver(user, "users", MySQL)
	if !strings.Contains(mysqlSQL, "CREATE TABLE") {
		t.Error("MySQL SQL should contain CREATE TABLE")
	}
	if !strings.Contains(mysqlSQL, "AUTO_INCREMENT") {
		t.Error("MySQL SQL should contain AUTO_INCREMENT")
	}
}

func TestParseFieldToColumnForDriver(t *testing.T) {
	// Get the ID field from TestUser
	userType := reflect.TypeOf(TestUser{})
	idField, found := userType.FieldByName("ID")
	if !found {
		t.Fatal("ID field not found")
	}

	// Test PostgreSQL
	pgColumn := ParseFieldToColumnForDriver(idField, PostgreSQL)
	if !strings.Contains(pgColumn, "id") {
		t.Error("PostgreSQL column should contain field name")
	}
	if !strings.Contains(pgColumn, "SERIAL") {
		t.Error("PostgreSQL column should contain SERIAL")
	}
	if !strings.Contains(pgColumn, "PRIMARY KEY") {
		t.Error("PostgreSQL column should contain PRIMARY KEY")
	}

	// Test SQLite
	sqliteColumn := ParseFieldToColumnForDriver(idField, SQLite)
	if !strings.Contains(sqliteColumn, "id") {
		t.Error("SQLite column should contain field name")
	}
	if !strings.Contains(sqliteColumn, "INTEGER") {
		t.Error("SQLite column should contain INTEGER")
	}

	// Test MySQL
	mysqlColumn := ParseFieldToColumnForDriver(idField, MySQL)
	if !strings.Contains(mysqlColumn, "id") {
		t.Error("MySQL column should contain field name")
	}
	if !strings.Contains(mysqlColumn, "BIGINT") {
		t.Error("MySQL column should contain BIGINT")
	}
}

func TestParseFieldToColumnForDriverStringField(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	nameField, found := userType.FieldByName("Name")
	if !found {
		t.Fatal("Name field not found")
	}

	// Test PostgreSQL
	pgColumn := ParseFieldToColumnForDriver(nameField, PostgreSQL)
	if !strings.Contains(pgColumn, "name") {
		t.Error("PostgreSQL column should contain field name")
	}
	if !strings.Contains(pgColumn, "VARCHAR(100)") {
		t.Error("PostgreSQL column should contain VARCHAR with length")
	}
	if !strings.Contains(pgColumn, "NOT NULL") {
		t.Error("PostgreSQL column should contain NOT NULL")
	}
}

func TestParseFieldToColumnForDriverBoolField(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	activeField, found := userType.FieldByName("IsActive")
	if !found {
		t.Fatal("IsActive field not found")
	}

	// Test PostgreSQL
	pgColumn := ParseFieldToColumnForDriver(activeField, PostgreSQL)
	if !strings.Contains(pgColumn, "is_active") {
		t.Error("PostgreSQL column should contain field name")
	}
	if !strings.Contains(pgColumn, "BOOLEAN") {
		t.Error("PostgreSQL column should contain BOOLEAN")
	}
	if !strings.Contains(pgColumn, "DEFAULT true") {
		t.Error("PostgreSQL column should contain DEFAULT true")
	}
}

func TestParseFieldToColumnForDriverTimeField(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	createdField, found := userType.FieldByName("CreatedAt")
	if !found {
		t.Fatal("CreatedAt field not found")
	}

	// Test PostgreSQL
	pgColumn := ParseFieldToColumnForDriver(createdField, PostgreSQL)
	if !strings.Contains(pgColumn, "created_at") {
		t.Error("PostgreSQL column should contain field name")
	}
	if !strings.Contains(pgColumn, "TIMESTAMP") {
		t.Error("PostgreSQL column should contain TIMESTAMP")
	}
}

func TestGenerateDropTableSQL(t *testing.T) {
	sql := GenerateDropTableSQL("users")
	expected := "DROP TABLE IF EXISTS users CASCADE;"
	if sql != expected {
		t.Errorf("Expected '%s', got '%s'", expected, sql)
	}
}

func TestGenerateIndexSQL(t *testing.T) {
	// Test regular index
	sql := GenerateIndexSQL("users", "idx_users_email", []string{"email"}, false)
	if !strings.Contains(sql, "CREATE INDEX") {
		t.Error("SQL should contain CREATE INDEX")
	}
	if !strings.Contains(sql, "idx_users_email") {
		t.Error("SQL should contain index name")
	}
	if !strings.Contains(sql, "users") {
		t.Error("SQL should contain table name")
	}
	if !strings.Contains(sql, "email") {
		t.Error("SQL should contain column name")
	}

	// Test unique index
	uniqueSQL := GenerateIndexSQL("users", "idx_users_email_unique", []string{"email"}, true)
	if !strings.Contains(uniqueSQL, "CREATE UNIQUE INDEX") {
		t.Error("SQL should contain CREATE UNIQUE INDEX")
	}

	// Test multi-column index
	multiSQL := GenerateIndexSQL("users", "idx_users_name_email", []string{"name", "email"}, false)
	if !strings.Contains(multiSQL, "name, email") {
		t.Error("SQL should contain both column names")
	}
}

func TestGenerateForeignKeySQL(t *testing.T) {
	sql := GenerateForeignKeySQL("orders", "user_id", "users", "id")

	if !strings.Contains(sql, "ALTER TABLE orders") {
		t.Error("SQL should contain ALTER TABLE with correct table name")
	}
	if !strings.Contains(sql, "ADD CONSTRAINT") {
		t.Error("SQL should contain ADD CONSTRAINT")
	}
	if !strings.Contains(sql, "fk_orders_user_id") {
		t.Error("SQL should contain constraint name")
	}
	if !strings.Contains(sql, "FOREIGN KEY (user_id)") {
		t.Error("SQL should contain foreign key column")
	}
	if !strings.Contains(sql, "REFERENCES users(id)") {
		t.Error("SQL should contain reference table and column")
	}
}

func TestCollectColumnsForDriver(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	columns := collectColumnsForDriver(userType, PostgreSQL)

	if len(columns) == 0 {
		t.Error("Should generate at least one column")
	}

	// Check that some expected columns are present
	columnsStr := strings.Join(columns, " ")
	if !strings.Contains(columnsStr, "id") {
		t.Error("Should contain id column")
	}
	if !strings.Contains(columnsStr, "name") {
		t.Error("Should contain name column")
	}
	if !strings.Contains(columnsStr, "email") {
		t.Error("Should contain email column")
	}
}

func TestProcessFieldForDriver(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})
	nameField, found := userType.FieldByName("Name")
	if !found {
		t.Fatal("Name field not found")
	}

	columns := processFieldForDriver(nameField, PostgreSQL)
	if len(columns) == 0 {
		t.Error("Should return at least one column definition")
	}

	columnsStr := strings.Join(columns, " ")
	if !strings.Contains(columnsStr, "name") {
		t.Error("Should contain field name")
	}
}

func TestGetEmbeddedType(t *testing.T) {
	// Test with a simple type
	stringType := reflect.TypeOf("")
	result := getEmbeddedType(stringType)
	if result != stringType {
		t.Error("Should return the same type for non-pointer")
	}

	// Test with pointer type
	ptrType := reflect.TypeOf(&TestUser{})
	result = getEmbeddedType(ptrType)
	expectedType := reflect.TypeOf(TestUser{})
	if result != expectedType {
		t.Error("Should return the underlying type for pointer")
	}
}

func TestExtractSQLValue(t *testing.T) {
	tests := []struct {
		sqlTag   string
		key      string
		expected string
	}{
		{"type:VARCHAR(255),not_null", "type", "VARCHAR(255)"},
		{"max_length:100,not_null", "max_length", "100"},
		{"default:true,not_null", "default", "true"},
		{"not_null,unique", "unique", ""},
		{"primary_key", "primary_key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := extractSQLValue(tt.sqlTag, tt.key)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDatabaseDriverConstants(t *testing.T) {
	if PostgreSQL != "postgres" {
		t.Errorf("PostgreSQL constant should be 'postgres', got '%s'", PostgreSQL)
	}
	if SQLite != "sqlite3" {
		t.Errorf("SQLite constant should be 'sqlite3', got '%s'", SQLite)
	}
	if MySQL != "mysql" {
		t.Errorf("MySQL constant should be 'mysql', got '%s'", MySQL)
	}
}

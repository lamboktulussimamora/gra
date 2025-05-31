package schema

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// DatabaseDriver represents the type of database driver
type DatabaseDriver string

const (
	// PostgreSQL driver
	PostgreSQL DatabaseDriver = "postgres"
	// SQLite driver
	SQLite DatabaseDriver = "sqlite3"
	// MySQL driver
	MySQL DatabaseDriver = "mysql"
)

// DetectDatabaseDriver attempts to detect the database driver type from a *sql.DB instance
func DetectDatabaseDriver(db *sql.DB) DatabaseDriver {
	// Try to get the driver name through reflection
	if db != nil {
		// Use a test query approach to detect database type
		// PostgreSQL specific query
		if _, err := db.Query("SELECT version()"); err == nil {
			// Try PostgreSQL-specific syntax
			if _, err := db.Query("SELECT 1::integer"); err == nil {
				return PostgreSQL
			}
		}

		// SQLite specific query
		if _, err := db.Query("SELECT sqlite_version()"); err == nil {
			return SQLite
		}

		// MySQL specific query
		if _, err := db.Query("SELECT VERSION()"); err == nil {
			return MySQL
		}
	}

	// Default to PostgreSQL if detection fails
	return PostgreSQL
}

// DetectDatabaseDriverFromConnectionString detects database type from connection string
func DetectDatabaseDriverFromConnectionString(driverName string) DatabaseDriver {
	switch strings.ToLower(driverName) {
	case "postgres", "postgresql":
		return PostgreSQL
	case "sqlite3", "sqlite":
		return SQLite
	case "mysql":
		return MySQL
	default:
		return PostgreSQL // Default fallback
	}
}

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          func(db *sql.DB) error
	Down        func(db *sql.DB) error
}

// ColumnDefinition represents a database column
type ColumnDefinition struct {
	Name         string
	Type         string
	IsPrimaryKey bool
	IsUnique     bool
	IsNullable   bool
	DefaultValue *string
	IsForeignKey bool
	References   *ForeignKeyReference
}

// ForeignKeyReference represents a foreign key reference
type ForeignKeyReference struct {
	Table  string
	Column string
}

// TableDefinition represents a database table
type TableDefinition struct {
	Name    string
	Columns []ColumnDefinition
	Indexes []IndexDefinition
}

// IndexDefinition represents a database index
type IndexDefinition struct {
	Name     string
	Columns  []string
	IsUnique bool
}

// GenerateCreateTableSQL generates CREATE TABLE SQL from a struct
func GenerateCreateTableSQL(entity interface{}, tableName string) string {
	return GenerateCreateTableSQLForDriver(entity, tableName, PostgreSQL)
}

// GenerateCreateTableSQLForDriver generates CREATE TABLE SQL from a struct for a specific database driver
func GenerateCreateTableSQLForDriver(entity interface{}, tableName string, driver DatabaseDriver) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []string
	var constraints []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip embedded structs (like BaseEntity)
		if field.Anonymous {
			// Recursively process embedded struct fields
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}

			for j := 0; j < embeddedType.NumField(); j++ {
				embeddedField := embeddedType.Field(j)
				if columnDef := ParseFieldToColumnForDriver(embeddedField, driver); columnDef != "" {
					columns = append(columns, columnDef)
				}
			}
			continue
		}

		// Skip navigation properties (slices and pointers to other structs)
		if isNavigationProperty(field) {
			continue
		}

		if columnDef := ParseFieldToColumnForDriver(field, driver); columnDef != "" {
			columns = append(columns, columnDef)
		}
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s", tableName, strings.Join(columns, ",\n  "))

	if len(constraints) > 0 {
		sql += ",\n  " + strings.Join(constraints, ",\n  ")
	}

	sql += "\n);"

	return sql
}

// parseFieldToColumn converts a struct field to a SQL column definition
func parseFieldToColumn(field reflect.StructField) string {
	return ParseFieldToColumnForDriver(field, PostgreSQL)
}

// ParseFieldToColumnForDriver converts a struct field to a SQL column definition for a specific database driver
func ParseFieldToColumnForDriver(field reflect.StructField, driver DatabaseDriver) string {
	dbTag := field.Tag.Get("db")
	if dbTag == "" || dbTag == "-" {
		return ""
	}

	sqlTag := field.Tag.Get("sql")
	columnName := dbTag

	// Determine SQL type based on Go type and database driver
	sqlType := goTypeToSQLTypeForDriver(field.Type, driver)

	// Parse SQL tags for additional column properties
	var parts []string
	parts = append(parts, fmt.Sprintf("%s %s", columnName, sqlType))

	if strings.Contains(sqlTag, "primary_key") {
		parts = append(parts, "PRIMARY KEY")
	}

	if strings.Contains(sqlTag, "auto_increment") {
		// Handle auto-increment based on database driver
		switch driver {
		case PostgreSQL:
			// PostgreSQL uses SERIAL for auto-increment
			if strings.Contains(sqlType, "INTEGER") || strings.Contains(sqlType, "BIGINT") {
				if strings.Contains(sqlType, "BIGINT") {
					parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "BIGSERIAL", 1)
				} else {
					parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "SERIAL", 1)
				}
			}
		case SQLite:
			// SQLite uses INTEGER PRIMARY KEY AUTOINCREMENT
			if strings.Contains(sqlTag, "primary_key") && strings.Contains(sqlType, "INTEGER") {
				parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "INTEGER", 1)
				// Add AUTOINCREMENT if not already present
				if !strings.Contains(strings.Join(parts, " "), "AUTOINCREMENT") {
					parts = append(parts, "AUTOINCREMENT")
				}
			}
		case MySQL:
			// MySQL uses AUTO_INCREMENT
			if strings.Contains(sqlType, "INTEGER") || strings.Contains(sqlType, "BIGINT") {
				parts = append(parts, "AUTO_INCREMENT")
			}
		}
	}

	if strings.Contains(sqlTag, "not_null") {
		parts = append(parts, "NOT NULL")
	}

	if strings.Contains(sqlTag, "unique") {
		parts = append(parts, "UNIQUE")
	}

	// Handle default values
	if defaultMatch := extractSQLValue(sqlTag, "default"); defaultMatch != "" {
		if defaultMatch == "CURRENT_TIMESTAMP" {
			parts = append(parts, "DEFAULT CURRENT_TIMESTAMP")
		} else if defaultMatch != "null" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultMatch))
		}
	}

	return strings.Join(parts, " ")
}

// goTypeToSQLType converts Go types to PostgreSQL types
func goTypeToSQLType(t reflect.Type) string {
	return goTypeToSQLTypeForDriver(t, PostgreSQL)
}

// goTypeToSQLTypeForDriver converts Go types to SQL types for a specific database driver
func goTypeToSQLTypeForDriver(t reflect.Type, driver DatabaseDriver) string {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch driver {
	case PostgreSQL:
		return goTypeToPostgreSQLType(t)
	case SQLite:
		return goTypeToSQLiteType(t)
	case MySQL:
		return goTypeToMySQLType(t)
	default:
		return goTypeToPostgreSQLType(t)
	}
}

// goTypeToPostgreSQLType converts Go types to PostgreSQL types
func goTypeToPostgreSQLType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "VARCHAR(255)"
	case reflect.Int, reflect.Int32:
		return "INTEGER"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Float32:
		return "REAL"
	case reflect.Float64:
		return "DOUBLE PRECISION"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "TIMESTAMP"
		}
		return "TEXT"
	}
}

// goTypeToSQLiteType converts Go types to SQLite types
func goTypeToSQLiteType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "INTEGER" // SQLite uses INTEGER for boolean (0/1)
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "DATETIME"
		}
		return "TEXT"
	}
}

// goTypeToMySQLType converts Go types to MySQL types
func goTypeToMySQLType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "VARCHAR(255)"
	case reflect.Int, reflect.Int32:
		return "INT"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Float32:
		return "FLOAT"
	case reflect.Float64:
		return "DOUBLE"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "DATETIME"
		}
		return "TEXT"
	}
}

// isNavigationProperty checks if a field is a navigation property
func isNavigationProperty(field reflect.StructField) bool {
	t := field.Type

	// Skip slices (one-to-many relationships)
	if t.Kind() == reflect.Slice {
		return true
	}

	// Skip pointers to structs that don't have db tags (foreign key relationships)
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		if elem.Kind() == reflect.Struct && field.Tag.Get("db") == "" {
			return true
		}
	}

	// Skip structs without db tags
	if t.Kind() == reflect.Struct && field.Tag.Get("db") == "" && t != reflect.TypeOf(time.Time{}) {
		return true
	}

	return false
}

// extractSQLValue extracts a value from SQL tag
func extractSQLValue(sqlTag, key string) string {
	parts := strings.Split(sqlTag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, key+":") {
			value := strings.TrimPrefix(part, key+":")
			return strings.Trim(value, "'\"")
		}
	}
	return ""
}

// GenerateDropTableSQL generates DROP TABLE SQL
func GenerateDropTableSQL(tableName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", tableName)
}

// GenerateIndexSQL generates CREATE INDEX SQL
func GenerateIndexSQL(tableName, indexName string, columns []string, unique bool) string {
	uniqueKeyword := ""
	if unique {
		uniqueKeyword = "UNIQUE "
	}

	return fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s (%s);",
		uniqueKeyword, indexName, tableName, strings.Join(columns, ", "))
}

// GenerateForeignKeySQL generates ALTER TABLE SQL for foreign keys
func GenerateForeignKeySQL(tableName, columnName, refTable, refColumn string) string {
	constraintName := fmt.Sprintf("fk_%s_%s", tableName, columnName)
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);",
		tableName, constraintName, columnName, refTable, refColumn)
}

// Package schema provides database schema utilities and driver detection for the ORM layer.
// It includes functions for generating SQL for table creation, index creation, and foreign key constraints.
// The package also supports automatic detection of database drivers (PostgreSQL, SQLite, MySQL)
// and provides migration support through struct tags.
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

	sqlTypeInteger = "INTEGER"
	sqlTypeText    = "TEXT"
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

	columns := collectColumnsForDriver(t, driver)
	constraints := collectConstraintsForDriver(t, driver)

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s", tableName, strings.Join(columns, ",\n  "))
	if len(constraints) > 0 {
		sql += ",\n  " + strings.Join(constraints, ",\n  ")
	}
	sql += "\n);"
	return sql
}

// collectColumnsForDriver recursively collects column definitions for a struct type
func collectColumnsForDriver(t reflect.Type, driver DatabaseDriver) []string {
	var columns []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columns = append(columns, processFieldForDriver(field, driver)...) // returns []string
	}
	return columns
}

// processFieldForDriver processes a struct field for column definitions
func processFieldForDriver(field reflect.StructField, driver DatabaseDriver) []string {
	if field.Anonymous {
		return collectColumnsForDriver(getEmbeddedType(field.Type), driver)
	}
	if isNavigationProperty(field) {
		return nil
	}
	if columnDef := ParseFieldToColumnForDriver(field, driver); columnDef != "" {
		return []string{columnDef}
	}
	return nil
}

// getEmbeddedType returns the underlying type for an embedded field
func getEmbeddedType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// collectConstraintsForDriver returns an empty slice (no constraints extracted)
func collectConstraintsForDriver(_ reflect.Type, _ DatabaseDriver) []string {
	// Constraint extraction not implemented
	return nil
}

// ParseFieldToColumnForDriver converts a struct field to a SQL column definition for a specific database driver
func ParseFieldToColumnForDriver(field reflect.StructField, driver DatabaseDriver) string {
	dbTag := field.Tag.Get("db")
	if dbTag == "" || dbTag == "-" {
		return ""
	}

	columnInfo := extractColumnInfo(field, driver)
	return buildColumnDefinition(columnInfo)
}

// columnInfo holds information about a database column
type columnInfo struct {
	name         string
	sqlType      string
	isPrimaryKey bool
	isAutoIncr   bool
	isNotNull    bool
	isUnique     bool
	defaultValue string
	driver       DatabaseDriver
	sqlTag       string
	migrationTag string
}

// extractColumnInfo extracts column information from a struct field
func extractColumnInfo(field reflect.StructField, driver DatabaseDriver) columnInfo {
	sqlTag := field.Tag.Get("sql")
	migrationTag := field.Tag.Get("migration")
	columnName := field.Tag.Get("db")

	sqlType := goTypeToSQLTypeForDriver(field.Type, driver)
	if migrationTag != "" {
		if typeMatch := extractSQLValue(migrationTag, "type"); typeMatch != "" {
			sqlType = typeMatch
		}
	}

	return columnInfo{
		name:         columnName,
		sqlType:      sqlType,
		isPrimaryKey: hasTagAttr(sqlTag, migrationTag, "primary_key"),
		isAutoIncr:   hasTagAttr(sqlTag, migrationTag, "auto_increment"),
		isNotNull:    hasTagAttr(sqlTag, migrationTag, "not_null"),
		isUnique:     hasTagAttr(sqlTag, migrationTag, "unique"),
		defaultValue: extractDefaultValue(sqlTag, migrationTag),
		driver:       driver,
		sqlTag:       sqlTag,
		migrationTag: migrationTag,
	}
}

// buildColumnDefinition builds the complete column definition string
func buildColumnDefinition(info columnInfo) string {
	parts := []string{fmt.Sprintf("%s %s", info.name, info.sqlType)}

	if info.isPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	if info.isAutoIncr {
		parts = handleAutoIncrementForColumn(parts, info)
	}

	if info.isNotNull {
		parts = append(parts, "NOT NULL")
	}

	if info.isUnique {
		parts = append(parts, "UNIQUE")
	}

	if info.defaultValue != "" {
		parts = appendDefaultValue(parts, info.defaultValue)
	}

	return strings.Join(parts, " ")
}

// extractDefaultValue extracts default value from tags
func extractDefaultValue(sqlTag, migrationTag string) string {
	if sqlTag != "" {
		if defaultMatch := extractSQLValue(sqlTag, "default"); defaultMatch != "" {
			return defaultMatch
		}
	}
	if migrationTag != "" {
		return extractSQLValue(migrationTag, "default")
	}
	return ""
}

// handleAutoIncrementForColumn handles auto-increment for a specific column
func handleAutoIncrementForColumn(parts []string, info columnInfo) []string {
	switch info.driver {
	case PostgreSQL:
		return handleAutoIncrementPostgres(parts, info.sqlType)
	case SQLite:
		return handleAutoIncrementSQLite(parts, info.sqlType, info.sqlTag, info.migrationTag)
	case MySQL:
		return handleAutoIncrementMySQL(parts, info.sqlType)
	default:
		return parts
	}
}

// appendDefaultValue appends default value to column parts
func appendDefaultValue(parts []string, defaultValue string) []string {
	if defaultValue == "CURRENT_TIMESTAMP" {
		return append(parts, "DEFAULT CURRENT_TIMESTAMP")
	}
	if defaultValue != "null" {
		return append(parts, fmt.Sprintf("DEFAULT %s", defaultValue))
	}
	return parts
}

// hasTagAttr checks if either sqlTag or migrationTag contains the attribute
func hasTagAttr(sqlTag, migrationTag, attr string) bool {
	return strings.Contains(sqlTag, attr) || strings.Contains(migrationTag, attr)
}

func handleAutoIncrementPostgres(parts []string, sqlType string) []string {
	if strings.Contains(sqlType, "INTEGER") || strings.Contains(sqlType, "BIGINT") {
		if strings.Contains(sqlType, "BIGINT") {
			parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "BIGSERIAL", 1)
		} else {
			parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "SERIAL", 1)
		}
	}
	return parts
}

func handleAutoIncrementSQLite(parts []string, sqlType, sqlTag, migrationTag string) []string {
	if hasTagAttr(sqlTag, migrationTag, "primary_key") && strings.Contains(sqlType, "INTEGER") {
		parts[len(parts)-1] = strings.Replace(parts[len(parts)-1], sqlType, "INTEGER", 1)
		if !strings.Contains(strings.Join(parts, " "), "AUTOINCREMENT") {
			parts = append(parts, "AUTOINCREMENT")
		}
	}
	return parts
}

func handleAutoIncrementMySQL(parts []string, sqlType string) []string {
	if strings.Contains(sqlType, "INTEGER") || strings.Contains(sqlType, "BIGINT") {
		parts = append(parts, "AUTO_INCREMENT")
	}
	return parts
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
		return sqlTypeInteger
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
		return sqlTypeText
	}
}

// goTypeToSQLiteType converts Go types to SQLite types
func goTypeToSQLiteType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return sqlTypeText
	case reflect.Int, reflect.Int32, reflect.Int64:
		return sqlTypeInteger
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "INTEGER" // SQLite uses INTEGER for boolean (0/1)
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "DATETIME"
		}
		return sqlTypeText
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
		return sqlTypeText
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

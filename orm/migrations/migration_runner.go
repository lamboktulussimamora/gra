package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/lib/pq"
)

// SQL and error message constants for migration runner
const (
	migrationsTableCreateSQL = `
	CREATE TABLE IF NOT EXISTS migrations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	sqlSelectMigrationCount = "SELECT COUNT(*) FROM migrations WHERE name = $1"
	sqlInsertMigration      = "INSERT INTO migrations (name) VALUES ($1)"
	sqlSelectMigrations     = "SELECT name, executed_at FROM migrations ORDER BY executed_at"

	errCreateMigrationsTable = "failed to create migrations table: %w"
	errCheckMigrationStatus  = "failed to check migration status: %w"
	errCreateTable           = "failed to create table %s: %w"
	errRecordMigration       = "failed to record migration: %w"
	errQueryMigrations       = "failed to query migrations: %w"
	errScanMigrationRow      = "failed to scan migration row: %w"

	msgMigrationsTableReady   = "✓ Migrations table ready"
	msgTableAlreadyExists     = "✓ Table %s already exists, skipping"
	msgCreatedTable           = "✓ Created table: %s"
	msgMigrationStatus        = "Migration Status:"
	msgMigrationStatusDivider = "================"
)

// MigrationRunner handles automatic database migrations
type MigrationRunner struct {
	db     *sql.DB
	logger *log.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(connectionString string) (*MigrationRunner, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MigrationRunner{
		db:     db,
		logger: log.Default(),
	}, nil
}

// Close closes the database connection
func (mr *MigrationRunner) Close() error {
	return mr.db.Close()
}

// AutoMigrate automatically creates or updates database schema based on entity models
func (mr *MigrationRunner) AutoMigrate() error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all entity types to migrate in dependency order
	entities := []interface{}{
		&models.Category{},
		&models.User{},
		&models.Product{},
	}

	for _, entity := range entities {
		if err := mr.migrateEntity(entity); err != nil {
			return fmt.Errorf("failed to migrate entity %T: %w", entity, err)
		}
	}

	mr.logger.Println("✓ All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) createMigrationsTable() error {
	_, err := mr.db.Exec(migrationsTableCreateSQL)
	if err != nil {
		return fmt.Errorf(errCreateMigrationsTable, err)
	}

	mr.logger.Println(msgMigrationsTableReady)
	return nil
}

// migrateEntity creates or updates table for an entity
func (mr *MigrationRunner) migrateEntity(entity interface{}) error {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	tableName := mr.getTableName(entityType.Name())
	migrationName := fmt.Sprintf("create_table_%s", tableName)

	// Check if migration already executed
	var count int
	err := mr.db.QueryRow(sqlSelectMigrationCount, migrationName).Scan(&count)
	if err != nil {
		return fmt.Errorf(errCheckMigrationStatus, err)
	}

	if count > 0 {
		mr.logger.Printf(msgTableAlreadyExists, tableName)
		return nil
	}

	// Generate CREATE TABLE statement
	createSQL := mr.generateCreateTableSQL(tableName, entityType)

	// Execute the migration
	_, err = mr.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf(errCreateTable, tableName, err)
	}

	// Record the migration
	_, err = mr.db.Exec(sqlInsertMigration, migrationName)
	if err != nil {
		return fmt.Errorf(errRecordMigration, err)
	}

	mr.logger.Printf(msgCreatedTable, tableName)
	return nil
}

// generateCreateTableSQL generates SQL for creating a table based on struct
func (mr *MigrationRunner) generateCreateTableSQL(tableName string, entityType reflect.Type) string {
	var columns []string

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columnDef := mr.generateColumnDefinition(field, dbTag)
		if columnDef != "" {
			columns = append(columns, columnDef)
		}
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		tableName, strings.Join(columns, ",\n  "))
}

// Helper for SQL type mapping
func sqlTypeForField(fieldType reflect.Type, dbTag string, field reflect.StructField) (string, bool) {
	isNullable := false
	if fieldType.Kind() == reflect.Ptr {
		isNullable = true
		fieldType = fieldType.Elem()
	}

	switch fieldType.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		if dbTag == "id" {
			return "SERIAL PRIMARY KEY", isNullable
		}
		return "INTEGER", isNullable
	case reflect.String:
		maxLength := field.Tag.Get("maxlength")
		if maxLength != "" {
			return fmt.Sprintf("VARCHAR(%s)", maxLength), isNullable
		}
		return "TEXT", isNullable
	case reflect.Float32, reflect.Float64:
		return "DECIMAL(10,2)", isNullable
	case reflect.Bool:
		return "BOOLEAN", isNullable
	case reflect.Struct:
		if fieldType.String() == "time.Time" {
			return "TIMESTAMP", isNullable
		}
		return "", isNullable // Skip unknown struct types
	default:
		return "", isNullable // Skip unsupported types
	}
}

// Helper for NOT NULL constraint
func addNotNullConstraint(sqlType, dbTag string, isNullable bool) string {
	if !isNullable && dbTag != "id" {
		return sqlType + " NOT NULL"
	}
	return sqlType
}

// Helper for default timestamp
func addDefaultTimestamp(sqlType, fieldTypeStr, dbTag string) string {
	if fieldTypeStr == "time.Time" && (dbTag == "created_at" || dbTag == "updated_at") {
		return sqlType + " DEFAULT CURRENT_TIMESTAMP"
	}
	return sqlType
}

func (mr *MigrationRunner) generateColumnDefinition(field reflect.StructField, dbTag string) string {
	fieldType := field.Type

	sqlType, isNullable := sqlTypeForField(fieldType, dbTag, field)
	if sqlType == "" {
		return ""
	}

	sqlType = addNotNullConstraint(sqlType, dbTag, isNullable)
	sqlType = addDefaultTimestamp(sqlType, fieldType.String(), dbTag)

	return fmt.Sprintf("%s %s", dbTag, sqlType)
}

// getTableName converts struct name to table name
func (mr *MigrationRunner) getTableName(structName string) string {
	// Convert CamelCase to snake_case and pluralize
	var result strings.Builder

	for i, r := range structName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String()) + "s"
}

// GetMigrationStatus returns the status of all migrations
func (mr *MigrationRunner) GetMigrationStatus() error {
	rows, err := mr.db.Query(sqlSelectMigrations)
	if err != nil {
		return fmt.Errorf(errQueryMigrations, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			mr.logger.Printf("Warning: Failed to close rows: %v", closeErr)
		}
	}()

	mr.logger.Println(msgMigrationStatus)
	mr.logger.Println(msgMigrationStatusDivider)

	for rows.Next() {
		var name string
		var executedAt time.Time

		if err := rows.Scan(&name, &executedAt); err != nil {
			return fmt.Errorf(errScanMigrationRow, err)
		}
		mr.logger.Printf("✓ %s (executed: %s)", name, executedAt.Format("2006-01-02 15:04:05"))
	}

	return rows.Err()
}

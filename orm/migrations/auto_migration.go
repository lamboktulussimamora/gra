// Package migrations provides database schema auto-migration functionality
package migrations

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/schema"
)

// AutoMigrator provides EF Core-style automatic database migrations
type AutoMigrator struct {
	ctx    *dbcontext.EnhancedDbContext
	db     *sql.DB
	logger func(string, ...interface{})
}

// NewAutoMigrator creates a new auto migrator
func NewAutoMigrator(ctx *dbcontext.EnhancedDbContext, db *sql.DB) *AutoMigrator {
	return &AutoMigrator{
		ctx:    ctx,
		db:     db,
		logger: func(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) },
	}
}

// SetLogger sets a custom logger function
func (am *AutoMigrator) SetLogger(logger func(string, ...interface{})) {
	am.logger = logger
}

// MigrateModels automatically creates/updates database schema for entity models
func (am *AutoMigrator) MigrateModels(models ...interface{}) error {
	// Create migrations table if it doesn't exist
	if err := am.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Migrate each model
	for _, model := range models {
		if err := am.migrateModel(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	am.logger("✓ All model migrations completed successfully")
	return nil
}

// CreateDatabase creates the database if it doesn't exist (PostgreSQL)
func (am *AutoMigrator) CreateDatabase(dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
	_, err := am.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	am.logger("✓ Database %s created or already exists", dbName)
	return nil
}

// DropDatabase drops the database (use with caution)
func (am *AutoMigrator) DropDatabase(dbName string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := am.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	am.logger("✓ Database %s dropped", dbName)
	return nil
}

// createMigrationsTable creates the __migrations tracking table
func (am *AutoMigrator) createMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS __migrations (
		id SERIAL PRIMARY KEY,
		migration_name VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		checksum VARCHAR(255)
	)`

	_, err := am.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create __migrations table: %w", err)
	}

	am.logger("✓ Migrations tracking table ready")
	return nil
}

// migrateModel creates or updates table for a model
func (am *AutoMigrator) migrateModel(model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	tableName := am.getTableName(model)
	migrationName := fmt.Sprintf("create_table_%s", tableName)

	// Generate table schema
	schema := am.generateTableSchema(modelType)
	checksum := am.calculateChecksum(schema)

	// Check if migration already applied with same checksum
	var existingChecksum string
	err := am.db.QueryRow("SELECT checksum FROM __migrations WHERE migration_name = $1", migrationName).Scan(&existingChecksum)

	if err == nil {
		// Migration exists
		if existingChecksum == checksum {
			am.logger("✓ Table %s is up to date", tableName)
			return nil
		} else {
			// Schema changed, need to update
			am.logger("⚠ Table %s schema changed, updating...", tableName)
			return am.updateTableSchema(tableName, modelType, migrationName, checksum)
		}
	} else if err == sql.ErrNoRows {
		// Migration doesn't exist, create table
		return am.createTable(tableName, modelType, migrationName, checksum)
	} else {
		return fmt.Errorf("failed to check migration status: %w", err)
	}
}

// createTable creates a new table
func (am *AutoMigrator) createTable(tableName string, modelType reflect.Type, migrationName, checksum string) error {
	createSQL := am.generateCreateTableSQL(tableName, modelType)

	// Start transaction
	tx, err := am.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create table
	_, err = tx.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	// Create indexes
	if err := am.createIndexes(tx, tableName, modelType); err != nil {
		return fmt.Errorf("failed to create indexes for %s: %w", tableName, err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO __migrations (migration_name, checksum) VALUES ($1, $2)", migrationName, checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	am.logger("✓ Created table: %s", tableName)
	return nil
}

// updateTableSchema updates an existing table schema
func (am *AutoMigrator) updateTableSchema(tableName string, modelType reflect.Type, migrationName, checksum string) error {
	// Get current table structure
	currentColumns, err := am.getCurrentTableColumns(tableName)
	if err != nil {
		return fmt.Errorf("failed to get current table columns: %w", err)
	}

	// Generate new structure
	newColumns := am.getModelColumns(modelType)

	// Start transaction
	tx, err := am.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Add new columns
	for colName, colDef := range newColumns {
		if _, exists := currentColumns[colName]; !exists {
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, colDef)
			_, err = tx.Exec(alterSQL)
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", colName, err)
			}
			am.logger("✓ Added column %s to table %s", colName, tableName)
		}
	}

	// Update migration record
	_, err = tx.Exec("UPDATE __migrations SET checksum = $1, applied_at = CURRENT_TIMESTAMP WHERE migration_name = $2", checksum, migrationName)
	if err != nil {
		return fmt.Errorf("failed to update migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update transaction: %w", err)
	}

	am.logger("✓ Updated table: %s", tableName)
	return nil
}

// processStructFields recursively processes all struct fields including embedded ones
func (am *AutoMigrator) processStructFields(modelType reflect.Type, fieldHandler func(field reflect.StructField, dbTag string)) {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check if this is an embedded struct
		if field.Anonymous {
			// This is an embedded struct, process its fields recursively
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				am.processStructFields(fieldType, fieldHandler)
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Call the handler for this field
		fieldHandler(field, dbTag)
	}
}

// processStructFieldsWithError recursively processes all struct fields including embedded ones with error handling
func (am *AutoMigrator) processStructFieldsWithError(modelType reflect.Type, fieldHandler func(field reflect.StructField, dbTag string) error) error {
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check if this is an embedded struct
		if field.Anonymous {
			// This is an embedded struct, process its fields recursively
			fieldType := field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				if err := am.processStructFieldsWithError(fieldType, fieldHandler); err != nil {
					return err
				}
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// Call the handler for this field
		if err := fieldHandler(field, dbTag); err != nil {
			return err
		}
	}
	return nil
}

// generateCreateTableSQL generates CREATE TABLE SQL using database-aware schema generation
func (am *AutoMigrator) generateCreateTableSQL(tableName string, modelType reflect.Type) string {
	// Detect database driver
	driver := schema.DetectDatabaseDriver(am.db)

	// Create a model instance to pass to the schema generator
	modelPtr := reflect.New(modelType)
	model := modelPtr.Interface()

	// Use the database-aware schema generation
	createSQL := schema.GenerateCreateTableSQLForDriver(model, tableName, driver)
	return createSQL
}

// generateCreateTableSQLLegacy generates CREATE TABLE SQL (legacy method for fallback)
func (am *AutoMigrator) generateCreateTableSQLLegacy(tableName string, modelType reflect.Type) string {
	var columns []string
	var constraints []string

	// Process all fields including embedded structs
	am.processStructFields(modelType, func(field reflect.StructField, dbTag string) {
		columnDef := am.generateColumnDefinition(field, dbTag)
		if columnDef != "" {
			columns = append(columns, columnDef)
		}

		// Handle foreign key constraints
		if fkTag := field.Tag.Get("fk"); fkTag != "" {
			constraint := am.generateForeignKeyConstraint(tableName, dbTag, fkTag)
			if constraint != "" {
				constraints = append(constraints, constraint)
			}
		}
	})

	// Combine columns and constraints
	allDefinitions := append(columns, constraints...)

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		tableName, strings.Join(allDefinitions, ",\n  "))
}

// generateColumnDefinition generates SQL column definition using database-aware schema generation
func (am *AutoMigrator) generateColumnDefinition(field reflect.StructField, dbTag string) string {
	// Detect database driver
	driver := schema.DetectDatabaseDriver(am.db)

	// Use the database-aware column parsing from schema package
	// The schema package reads the db tag from the field, so we use the field directly
	return schema.ParseFieldToColumnForDriver(field, driver)
}

// generateForeignKeyConstraint generates foreign key constraint
func (am *AutoMigrator) generateForeignKeyConstraint(tableName, columnName, fkTag string) string {
	// Parse fk tag: "table.column" or "table"
	parts := strings.Split(fkTag, ".")
	refTable := parts[0]
	refColumn := "id"
	if len(parts) > 1 {
		refColumn = parts[1]
	}

	constraintName := fmt.Sprintf("fk_%s_%s", tableName, columnName)
	return fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)",
		constraintName, columnName, refTable, refColumn)
}

// createIndexes creates indexes based on struct tags
func (am *AutoMigrator) createIndexes(tx *sql.Tx, tableName string, modelType reflect.Type) error {
	return am.processStructFieldsWithError(modelType, func(field reflect.StructField, dbTag string) error {
		// Create index if specified
		if field.Tag.Get("index") == "true" {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, dbTag)
			indexSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, dbTag)
			_, err := tx.Exec(indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create index %s: %w", indexName, err)
			}
		}

		// Create unique index if specified
		if field.Tag.Get("uniqueIndex") == "true" {
			indexName := fmt.Sprintf("uidx_%s_%s", tableName, dbTag)
			indexSQL := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, dbTag)
			_, err := tx.Exec(indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create unique index %s: %w", indexName, err)
			}
		}
		return nil
	})
}

// Helper functions

// getTableName gets table name from model
func (am *AutoMigrator) getTableName(model interface{}) string {
	// Check if model has TableName method
	if tn, ok := model.(interface{ TableName() string }); ok {
		tableName := tn.TableName()
		return tableName
	}

	// Use the same logic as dbcontext for consistency
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeName := t.Name()
	snakeCaseName := am.toSnakeCase(typeName)
	return snakeCaseName
}

// toSnakeCase converts CamelCase to snake_case (same as in dbcontext)
func (am *AutoMigrator) toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result.WriteRune('_')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// generateTableSchema generates a complete table schema for checksum calculation
func (am *AutoMigrator) generateTableSchema(modelType reflect.Type) string {
	var parts []string

	am.processStructFields(modelType, func(field reflect.StructField, dbTag string) {
		columnDef := am.generateColumnDefinition(field, dbTag)
		if columnDef != "" {
			parts = append(parts, columnDef)
		}
	})

	return strings.Join(parts, "|")
}

// calculateChecksum calculates a simple checksum for schema comparison
func (am *AutoMigrator) calculateChecksum(schema string) string {
	// Simple hash function (in production, use a proper hash like SHA256)
	hash := 0
	for _, char := range schema {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// getCurrentTableColumns gets current table column information
func (am *AutoMigrator) getCurrentTableColumns(tableName string) (map[string]string, error) {
	driver := schema.DetectDatabaseDriver(am.db)

	var query string
	var args []interface{}

	switch driver {
	case schema.PostgreSQL:
		query = `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = $1
			ORDER BY ordinal_position`
		args = []interface{}{tableName}

	case schema.SQLite:
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		args = []interface{}{}

	case schema.MySQL:
		query = `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = ? AND table_schema = DATABASE()
			ORDER BY ordinal_position`
		args = []interface{}{tableName}

	default:
		return nil, fmt.Errorf("unsupported database driver: %v", driver)
	}

	rows, err := am.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]string)

	if driver == schema.SQLite {
		// SQLite PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull int
			var defaultValue sql.NullString
			var pk int

			if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
				return nil, err
			}

			nullable := "YES"
			if notNull == 1 {
				nullable = "NO"
			}

			colInfo := fmt.Sprintf("type:%s,nullable:%s", dataType, nullable)
			if defaultValue.Valid {
				colInfo += fmt.Sprintf(",default:%s", defaultValue.String)
			}
			columns[name] = colInfo
		}
	} else {
		// PostgreSQL and MySQL use information_schema
		for rows.Next() {
			var colName, dataType, isNullable string
			var columnDefault sql.NullString

			err := rows.Scan(&colName, &dataType, &isNullable, &columnDefault)
			if err != nil {
				return nil, err
			}

			colInfo := fmt.Sprintf("type:%s,nullable:%s", dataType, isNullable)
			if columnDefault.Valid {
				colInfo += fmt.Sprintf(",default:%s", columnDefault.String)
			}
			columns[colName] = colInfo
		}
	}

	return columns, rows.Err()
}

// getModelColumns gets column definitions from model
func (am *AutoMigrator) getModelColumns(modelType reflect.Type) map[string]string {
	columns := make(map[string]string)

	am.processStructFields(modelType, func(field reflect.StructField, dbTag string) {
		columnDef := am.generateColumnDefinition(field, dbTag)
		if columnDef != "" {
			columns[dbTag] = columnDef
		}
	})

	return columns
}

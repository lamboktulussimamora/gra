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

// SQL and error message constants for auto migration
const (
	dbErrCreateMigrationsTable = "failed to create __migrations table: %w"

	dbErrBeginTx               = "failed to begin transaction: %w"
	dbErrCreateTable           = "failed to create table %s: %w"
	dbErrCreateIndexes         = "failed to create indexes for %s: %w"
	dbErrRecordMigration       = "failed to record migration: %w"
	dbErrCommitMigration       = "failed to commit migration transaction: %w"
	dbErrGetCurrentColumns     = "failed to get current table columns: %w"
	dbErrAddColumn             = "failed to add column %s: %w"
	dbErrUpdateMigrationRecord = "failed to update migration record: %w"
	dbErrCommitUpdate          = "failed to commit update transaction: %w"
	dbWarnRollback             = "Warning: Failed to rollback transaction: %v"
	dbWarnCloseRows            = "Warning: Failed to close rows: %v"
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
		return fmt.Errorf(dbErrCreateMigrationsTable, err)
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
		return fmt.Errorf(dbErrBeginTx, err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			if rollbackErr != sql.ErrTxDone {
				am.logger(dbWarnRollback, rollbackErr)
			}
		}
	}()

	// Create table
	_, err = tx.Exec(createSQL)
	if err != nil {
		return fmt.Errorf(dbErrCreateTable, tableName, err)
	}

	// Create indexes
	if err := am.createIndexes(tx, tableName, modelType); err != nil {
		return fmt.Errorf(dbErrCreateIndexes, tableName, err)
	}

	// Record migration
	_, err = tx.Exec("INSERT INTO __migrations (migration_name, checksum) VALUES ($1, $2)", migrationName, checksum)
	if err != nil {
		return fmt.Errorf(dbErrRecordMigration, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf(dbErrCommitMigration, err)
	}

	am.logger("✓ Created table: %s", tableName)
	return nil
}

// updateTableSchema updates an existing table schema
func (am *AutoMigrator) updateTableSchema(tableName string, modelType reflect.Type, migrationName, checksum string) error {
	// Get current table structure
	currentColumns, err := am.getCurrentTableColumns(tableName)
	if err != nil {
		return fmt.Errorf(dbErrGetCurrentColumns, err)
	}

	// Generate new structure
	newColumns := am.getModelColumns(modelType)

	// Start transaction
	tx, err := am.db.Begin()
	if err != nil {
		return fmt.Errorf(dbErrBeginTx, err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			if rollbackErr != sql.ErrTxDone {
				am.logger(dbWarnRollback, rollbackErr)
			}
		}
	}()

	// Add new columns
	for colName, colDef := range newColumns {
		if _, exists := currentColumns[colName]; !exists {
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, colDef)
			_, err = tx.Exec(alterSQL)
			if err != nil {
				return fmt.Errorf(dbErrAddColumn, colName, err)
			}
			am.logger("✓ Added column %s to table %s", colName, tableName)
		}
	}

	// Update migration record
	_, err = tx.Exec("UPDATE __migrations SET checksum = $1, applied_at = CURRENT_TIMESTAMP WHERE migration_name = $2", checksum, migrationName)
	if err != nil {
		return fmt.Errorf(dbErrUpdateMigrationRecord, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf(dbErrCommitUpdate, err)
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

		if am.isEmbeddedStruct(field) {
			if err := am.handleEmbeddedStructWithError(field, fieldHandler); err != nil {
				return err
			}
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		if err := fieldHandler(field, dbTag); err != nil {
			return err
		}
	}
	return nil
}

// isEmbeddedStruct checks if a struct field is an embedded struct
func (am *AutoMigrator) isEmbeddedStruct(field reflect.StructField) bool {
	if !field.Anonymous {
		return false
	}
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	return fieldType.Kind() == reflect.Struct
}

// handleEmbeddedStructWithError processes embedded struct fields recursively with error handling
func (am *AutoMigrator) handleEmbeddedStructWithError(field reflect.StructField, fieldHandler func(field reflect.StructField, dbTag string) error) error {
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() == reflect.Struct {
		return am.processStructFieldsWithError(fieldType, fieldHandler)
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

// generateColumnDefinition generates SQL column definition using database-aware schema generation
func (am *AutoMigrator) generateColumnDefinition(field reflect.StructField, dbTag string) string {
	// Detect database driver
	driver := schema.DetectDatabaseDriver(am.db)

	// Use the database-aware column parsing from schema package
	// The schema package reads the db tag from the field, so we use the field directly
	return schema.ParseFieldToColumnForDriver(field, driver)
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

	query, args, err := am.getTableColumnsQuery(driver, tableName)
	if err != nil {
		return nil, err
	}

	rows, err := am.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			am.logger(dbWarnCloseRows, closeErr)
		}
	}()

	columns := make(map[string]string)

	switch driver {
	case schema.SQLite:
		return am.scanSQLiteTableInfo(rows, columns)
	case schema.PostgreSQL, schema.MySQL:
		return am.scanInformationSchemaColumns(rows, columns)
	default:
		return nil, fmt.Errorf("unsupported database driver: %v", driver)
	}
}

// getTableColumnsQuery returns the query and args for fetching table columns based on driver
func (am *AutoMigrator) getTableColumnsQuery(driver schema.DatabaseDriver, tableName string) (string, []interface{}, error) {
	switch driver {
	case schema.PostgreSQL:
		return `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = $1
			ORDER BY ordinal_position`, []interface{}{tableName}, nil
	case schema.SQLite:
		return fmt.Sprintf("PRAGMA table_info(%s)", tableName), []interface{}{}, nil
	case schema.MySQL:
		return `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = ? AND table_schema = DATABASE()
			ORDER BY ordinal_position`, []interface{}{tableName}, nil
	default:
		return "", nil, fmt.Errorf("unsupported database driver: %v", driver)
	}
}

// scanSQLiteTableInfo scans SQLite PRAGMA table_info results into columns map
func (am *AutoMigrator) scanSQLiteTableInfo(rows *sql.Rows, columns map[string]string) (map[string]string, error) {
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
	return columns, rows.Err()
}

// scanInformationSchemaColumns scans PostgreSQL/MySQL information_schema results into columns map
func (am *AutoMigrator) scanInformationSchemaColumns(rows *sql.Rows, columns map[string]string) (map[string]string, error) {
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

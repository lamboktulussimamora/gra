package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// DatabaseInspector reads current database schema state
type DatabaseInspector struct {
	db     *sql.DB
	driver DatabaseDriver
}

// NewDatabaseInspector creates a new database inspector
func NewDatabaseInspector(db *sql.DB, driver DatabaseDriver) *DatabaseInspector {
	return &DatabaseInspector{
		db:     db,
		driver: driver,
	}
}

// GetCurrentSchema reads the current database schema and returns table snapshots
func (di *DatabaseInspector) GetCurrentSchema() (map[string]*TableSchema, error) {
	switch di.driver {
	case PostgreSQL:
		return di.getPostgreSQLSchema()
	case MySQL:
		return di.getMySQLSchema()
	case SQLite:
		return di.getSQLiteSchema()
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", di.driver)
	}
}

// TableSchema represents the current state of a table in the database
type TableSchema struct {
	Name        string
	Columns     map[string]*DatabaseColumnInfo
	PrimaryKeys []string
	Indexes     map[string]*IndexInfo
	Constraints map[string]*ConstraintInfo
}

// DatabaseColumnInfo represents a column as it exists in the database
type DatabaseColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	MaxLength    *int
	Precision    *int
	Scale        *int
	IsIdentity   bool
	IsGenerated  bool
}

// getPostgreSQLSchema reads schema from PostgreSQL
func (di *DatabaseInspector) getPostgreSQLSchema() (map[string]*TableSchema, error) {
	tables := make(map[string]*TableSchema)

	// Get all tables in the current schema
	tableRows, err := di.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer tableRows.Close()

	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		table := &TableSchema{
			Name:        tableName,
			Columns:     make(map[string]*DatabaseColumnInfo),
			PrimaryKeys: []string{},
			Indexes:     make(map[string]*IndexInfo),
			Constraints: make(map[string]*ConstraintInfo),
		}

		// Get columns for this table
		if err := di.getPostgreSQLColumns(table); err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		// Get primary keys
		if err := di.getPostgreSQLPrimaryKeys(table); err != nil {
			return nil, fmt.Errorf("failed to get primary keys for table %s: %w", tableName, err)
		}

		// Get indexes
		if err := di.getPostgreSQLIndexes(table); err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", tableName, err)
		}

		// Get constraints
		if err := di.getPostgreSQLConstraints(table); err != nil {
			return nil, fmt.Errorf("failed to get constraints for table %s: %w", tableName, err)
		}

		tables[tableName] = table
	}

	return tables, nil
}

// getPostgreSQLColumns reads column information for a table
func (di *DatabaseInspector) getPostgreSQLColumns(table *TableSchema) error {
	rows, err := di.db.Query(`
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			is_identity,
			is_generated
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = $1
		ORDER BY ordinal_position
	`, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			columnName   string
			dataType     string
			isNullable   string
			defaultValue sql.NullString
			maxLength    sql.NullInt64
			precision    sql.NullInt64
			scale        sql.NullInt64
			isIdentity   string
			isGenerated  string
		)

		if err := rows.Scan(
			&columnName, &dataType, &isNullable, &defaultValue,
			&maxLength, &precision, &scale, &isIdentity, &isGenerated,
		); err != nil {
			return err
		}

		column := &DatabaseColumnInfo{
			Name:        columnName,
			DataType:    dataType,
			IsNullable:  isNullable == "YES",
			IsIdentity:  isIdentity == "YES",
			IsGenerated: isGenerated != "NEVER",
		}

		if defaultValue.Valid {
			column.DefaultValue = &defaultValue.String
		}
		if maxLength.Valid {
			length := int(maxLength.Int64)
			column.MaxLength = &length
		}
		if precision.Valid {
			prec := int(precision.Int64)
			column.Precision = &prec
		}
		if scale.Valid {
			sc := int(scale.Int64)
			column.Scale = &sc
		}

		table.Columns[columnName] = column
	}

	return nil
}

// getPostgreSQLPrimaryKeys reads primary key information
func (di *DatabaseInspector) getPostgreSQLPrimaryKeys(table *TableSchema) error {
	rows, err := di.db.Query(`
		SELECT column_name
		FROM information_schema.key_column_usage 
		WHERE table_schema = 'public' 
		AND table_name = $1
		AND constraint_name IN (
			SELECT constraint_name 
			FROM information_schema.table_constraints 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			AND constraint_type = 'PRIMARY KEY'
		)
		ORDER BY ordinal_position
	`, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return err
		}
		table.PrimaryKeys = append(table.PrimaryKeys, columnName)
	}

	return nil
}

// getPostgreSQLIndexes reads index information
func (di *DatabaseInspector) getPostgreSQLIndexes(table *TableSchema) error {
	rows, err := di.db.Query(`
		SELECT 
			i.indexname,
			i.indexdef,
			ix.indisunique
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		JOIN pg_index ix ON ix.indexrelid = (
			SELECT oid FROM pg_class WHERE relname = i.indexname
		)
		WHERE i.schemaname = 'public' 
		AND i.tablename = $1
		AND i.indexname NOT LIKE '%_pkey'
	`, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			indexName string
			indexDef  string
			isUnique  bool
		)

		if err := rows.Scan(&indexName, &indexDef, &isUnique); err != nil {
			return err
		}

		// Parse column names from index definition
		columns := di.parsePostgreSQLIndexColumns(indexDef)

		table.Indexes[indexName] = &IndexInfo{
			Name:     indexName,
			Columns:  columns,
			IsUnique: isUnique,
		}
	}

	return nil
}

// parsePostgreSQLIndexColumns extracts column names from PostgreSQL index definition
func (di *DatabaseInspector) parsePostgreSQLIndexColumns(indexDef string) []string {
	// Simple parsing for common cases
	// More sophisticated parsing would be needed for complex expressions
	start := strings.Index(indexDef, "(")
	end := strings.LastIndex(indexDef, ")")
	if start == -1 || end == -1 || start >= end {
		return []string{}
	}

	columnPart := indexDef[start+1 : end]
	columns := strings.Split(columnPart, ",")

	result := make([]string, 0, len(columns))
	for _, col := range columns {
		col = strings.TrimSpace(col)
		// Remove any ordering or function calls for simple column names
		if parts := strings.Fields(col); len(parts) > 0 {
			result = append(result, parts[0])
		}
	}

	return result
}

// getPostgreSQLConstraints reads constraint information
func (di *DatabaseInspector) getPostgreSQLConstraints(table *TableSchema) error {
	rows, err := di.db.Query(`
		SELECT 
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		LEFT JOIN information_schema.constraint_column_usage ccu 
			ON tc.constraint_name = ccu.constraint_name
		WHERE tc.table_schema = 'public' 
		AND tc.table_name = $1
		AND tc.constraint_type IN ('FOREIGN KEY', 'UNIQUE', 'CHECK')
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	constraintMap := make(map[string]*ConstraintInfo)

	for rows.Next() {
		var (
			constraintName    string
			constraintType    string
			columnName        sql.NullString
			foreignTableName  sql.NullString
			foreignColumnName sql.NullString
		)

		if err := rows.Scan(
			&constraintName, &constraintType, &columnName,
			&foreignTableName, &foreignColumnName,
		); err != nil {
			return err
		}

		constraint, exists := constraintMap[constraintName]
		if !exists {
			constraint = &ConstraintInfo{
				Name: constraintName,
				Type: constraintType,
			}
			constraintMap[constraintName] = constraint
		}

		if columnName.Valid {
			constraint.Columns = append(constraint.Columns, columnName.String)
		}

		if constraintType == "FOREIGN KEY" && foreignTableName.Valid && foreignColumnName.Valid {
			constraint.ReferencedTable = foreignTableName.String
			constraint.ReferencedColumns = append(constraint.ReferencedColumns, foreignColumnName.String)
		}
	}

	// Sort columns for each constraint to ensure consistent ordering
	for _, constraint := range constraintMap {
		sort.Strings(constraint.Columns)
		sort.Strings(constraint.ReferencedColumns)
	}

	table.Constraints = constraintMap
	return nil
}

// getMySQLSchema reads schema from MySQL
func (di *DatabaseInspector) getMySQLSchema() (map[string]*TableSchema, error) {
	// Implementation for MySQL would go here
	// Similar structure to PostgreSQL but with MySQL-specific queries
	return nil, fmt.Errorf("MySQL schema inspection not yet implemented")
}

// getSQLiteSchema reads schema from SQLite
func (di *DatabaseInspector) getSQLiteSchema() (map[string]*TableSchema, error) {
	tables := make(map[string]*TableSchema)

	// Get all tables (excluding sqlite_* system tables)
	tableRows, err := di.db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer tableRows.Close()

	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		table := &TableSchema{
			Name:        tableName,
			Columns:     make(map[string]*DatabaseColumnInfo),
			PrimaryKeys: []string{},
			Indexes:     make(map[string]*IndexInfo),
			Constraints: make(map[string]*ConstraintInfo),
		}

		// Get columns for this table
		if err := di.getSQLiteColumns(table); err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		// Get indexes
		if err := di.getSQLiteIndexes(table); err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", tableName, err)
		}

		tables[tableName] = table
	}

	return tables, nil
}

// getSQLiteColumns reads column information for a SQLite table
func (di *DatabaseInspector) getSQLiteColumns(table *TableSchema) error {
	// Use PRAGMA table_info to get column information
	rows, err := di.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table.Name))
	if err != nil {
		return fmt.Errorf("failed to get column info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}

		column := &DatabaseColumnInfo{
			Name:       name,
			DataType:   dataType,
			IsNullable: notNull == 0,
			IsIdentity: false, // SQLite doesn't have separate identity concept
		}

		if defaultValue.Valid {
			column.DefaultValue = &defaultValue.String
		}

		// Parse data type for length, precision, scale
		di.parseSQLiteDataType(column, dataType)

		table.Columns[name] = column

		// If this is a primary key column, add it to the primary keys list
		if pk == 1 {
			table.PrimaryKeys = append(table.PrimaryKeys, name)
		}
	}

	// Sort primary keys by ordinal position
	sort.Strings(table.PrimaryKeys)
	return nil
}

// getSQLiteIndexes reads index information for a SQLite table
func (di *DatabaseInspector) getSQLiteIndexes(table *TableSchema) error {
	// Get index list for the table
	rows, err := di.db.Query(fmt.Sprintf("PRAGMA index_list(%s)", table.Name))
	if err != nil {
		return fmt.Errorf("failed to get index list: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name, unique, origin string
		var partial int

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return fmt.Errorf("failed to scan index info: %w", err)
		}

		// Skip auto-created indexes for primary keys and unique constraints
		if strings.HasPrefix(name, "sqlite_autoindex_") {
			continue
		}

		index := &IndexInfo{
			Name:   name,
			Unique: unique == "1",
			Type:   "btree", // SQLite primarily uses btree indexes
		}

		// Get index columns
		colRows, err := di.db.Query(fmt.Sprintf("PRAGMA index_info(%s)", name))
		if err != nil {
			return fmt.Errorf("failed to get index columns: %w", err)
		}

		var columns []string
		for colRows.Next() {
			var seqno, cid int
			var colName string
			if err := colRows.Scan(&seqno, &cid, &colName); err != nil {
				colRows.Close()
				return fmt.Errorf("failed to scan index column: %w", err)
			}
			columns = append(columns, colName)
		}
		colRows.Close()

		index.Columns = columns
		table.Indexes[name] = index
	}

	return nil
}

// parseSQLiteDataType parses SQLite data type to extract length, precision, scale
func (di *DatabaseInspector) parseSQLiteDataType(column *DatabaseColumnInfo, dataType string) {
	// SQLite data types can be like VARCHAR(255), DECIMAL(10,2), etc.
	upperType := strings.ToUpper(dataType)

	// Extract length for VARCHAR, CHAR, etc.
	if strings.Contains(upperType, "VARCHAR") || strings.Contains(upperType, "CHAR") {
		if start := strings.Index(upperType, "("); start != -1 {
			if end := strings.Index(upperType[start:], ")"); end != -1 {
				lengthStr := upperType[start+1 : start+end]
				if length := di.parseIntValue(lengthStr); length > 0 {
					column.MaxLength = &length
				}
			}
		}
	}

	// Extract precision and scale for DECIMAL, NUMERIC
	if strings.Contains(upperType, "DECIMAL") || strings.Contains(upperType, "NUMERIC") {
		if start := strings.Index(upperType, "("); start != -1 {
			if end := strings.Index(upperType[start:], ")"); end != -1 {
				params := upperType[start+1 : start+end]
				parts := strings.Split(params, ",")
				if len(parts) >= 1 {
					if precision := di.parseIntValue(strings.TrimSpace(parts[0])); precision > 0 {
						column.Precision = &precision
					}
				}
				if len(parts) >= 2 {
					if scale := di.parseIntValue(strings.TrimSpace(parts[1])); scale >= 0 {
						column.Scale = &scale
					}
				}
			}
		}
	}
}

// parseIntValue safely parses an integer value
func (di *DatabaseInspector) parseIntValue(s string) int {
	if s == "" {
		return 0
	}
	// Simple integer parsing without importing strconv
	var result int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0 // Invalid character
		}
	}
	return result
}

// CompareWithModelSnapshot compares database schema with model snapshots and returns migration changes
func (di *DatabaseInspector) CompareWithModelSnapshot(dbSchema map[string]*TableSchema, modelSnapshots map[string]*ModelSnapshot) ([]MigrationChange, error) {
	var changes []MigrationChange

	fmt.Printf("DEBUG CompareWithModelSnapshot: dbSchema has %d tables, modelSnapshots has %d models\n", len(dbSchema), len(modelSnapshots))

	// Track which tables exist in both database and models
	processedTables := make(map[string]bool)

	// Check for new tables (exist in model but not in database)
	for modelName, snapshot := range modelSnapshots {
		tableName := snapshot.TableName
		processedTables[tableName] = true

		fmt.Printf("DEBUG: Processing model %s -> table %s\n", modelName, tableName)

		if _, exists := dbSchema[tableName]; !exists {
			// Table doesn't exist in database - create it
			fmt.Printf("DEBUG: Table %s does not exist in database, creating CreateTable change\n", tableName)
			changes = append(changes, MigrationChange{
				Type:      CreateTable,
				TableName: tableName,
				ModelName: modelName,
				NewValue:  snapshot,
			})
		} else {
			// Table exists - check for column changes
			fmt.Printf("DEBUG: Table %s exists, checking for column changes\n", tableName)
			columnChanges := di.compareTableColumns(dbSchema[tableName], snapshot)
			changes = append(changes, columnChanges...)
		}
	}

	// Check for tables to drop (exist in database but not in models)
	for tableName, tableSchema := range dbSchema {
		if di.isSystemTable(tableName) {
			fmt.Printf("DEBUG: Skipping system table %s\n", tableName)
			continue
		}

		if !processedTables[tableName] {
			fmt.Printf("DEBUG: Table %s exists in database but not in models, creating DropTable change\n", tableName)
			changes = append(changes, MigrationChange{
				Type:      DropTable,
				TableName: tableName,
				OldValue:  tableSchema,
			})
		}
	}

	fmt.Printf("DEBUG CompareWithModelSnapshot: Generated %d changes\n", len(changes))
	for i, change := range changes {
		fmt.Printf("DEBUG: Change %d: %s %s.%s\n", i, change.Type, change.TableName, change.ColumnName)
	}

	return changes, nil
}

// compareTableColumns compares columns between database table and model snapshot
func (di *DatabaseInspector) compareTableColumns(dbTable *TableSchema, modelSnapshot *ModelSnapshot) []MigrationChange {
	var changes []MigrationChange

	// Track which columns exist in both database and model
	processedColumns := make(map[string]bool)

	// Check for new columns (exist in model but not in database)
	for columnName, modelColumn := range modelSnapshot.Columns {
		processedColumns[columnName] = true

		if dbColumn, exists := dbTable.Columns[columnName]; !exists {
			// Column doesn't exist in database - add it
			fmt.Printf("DEBUG: Column %s.%s does not exist in database, creating AddColumn change\n", dbTable.Name, columnName)
			changes = append(changes, MigrationChange{
				Type:       AddColumn,
				TableName:  dbTable.Name,
				ColumnName: columnName,
				NewColumn:  modelColumn,
			})
		} else {
			// Column exists - check if it has changed
			if di.hasColumnChanged(modelColumn, dbColumn) {
				fmt.Printf("DEBUG: Column %s.%s has changed, creating AlterColumn change\n", dbTable.Name, columnName)
				changes = append(changes, MigrationChange{
					Type:       AlterColumn,
					TableName:  dbTable.Name,
					ColumnName: columnName,
					OldColumn:  di.convertDatabaseColumnToColumnInfo(dbColumn),
					NewColumn:  modelColumn,
				})
			}
		}
	}

	// Check for columns to drop (exist in database but not in model)
	for columnName, dbColumn := range dbTable.Columns {
		if !processedColumns[columnName] {
			fmt.Printf("DEBUG: Column %s.%s exists in database but not in model, creating DropColumn change\n", dbTable.Name, columnName)
			changes = append(changes, MigrationChange{
				Type:       DropColumn,
				TableName:  dbTable.Name,
				ColumnName: columnName,
				OldColumn:  di.convertDatabaseColumnToColumnInfo(dbColumn),
			})
		}
	}

	return changes
}

// compareTableStructure compares a model snapshot with database table structure
func (di *DatabaseInspector) compareTableStructure(
	snapshot *ModelSnapshot,
	dbTable *TableSchema,
	modelName string,
) []MigrationChange {
	var changes []MigrationChange

	// Find columns to add
	for columnName, columnInfo := range snapshot.Columns {
		if _, exists := dbTable.Columns[columnName]; !exists {
			changes = append(changes, MigrationChange{
				Type:       AddColumn,
				TableName:  snapshot.TableName,
				ModelName:  modelName,
				ColumnName: columnName,
				NewValue:   columnInfo,
			})
		}
	}

	// Find columns to drop
	for columnName, dbColumn := range dbTable.Columns {
		if _, exists := snapshot.Columns[columnName]; !exists {
			changes = append(changes, MigrationChange{
				Type:       DropColumn,
				TableName:  snapshot.TableName,
				ColumnName: columnName,
				OldValue:   dbColumn,
			})
		}
	}

	// Find columns to modify
	for columnName, columnInfo := range snapshot.Columns {
		if dbColumn, exists := dbTable.Columns[columnName]; exists {
			if di.hasColumnChanged(columnInfo, dbColumn) {
				changes = append(changes, MigrationChange{
					Type:       AlterColumn,
					TableName:  snapshot.TableName,
					ModelName:  modelName,
					ColumnName: columnName,
					OldValue:   dbColumn,
					NewValue:   columnInfo,
				})
			}
		}
	}

	// Compare indexes
	indexChanges := di.compareIndexes(snapshot, dbTable, modelName)
	changes = append(changes, indexChanges...)

	return changes
}

// hasColumnChanged checks if a column definition has changed
func (di *DatabaseInspector) hasColumnChanged(modelColumn *ColumnInfo, dbColumn *DatabaseColumnInfo) bool {
	// Debug: Log column comparison
	fmt.Printf("DEBUG: Comparing column %s:\n", dbColumn.Name)
	fmt.Printf("DEBUG:   Model: DataType=%s, IsNullable=%t, DefaultValue=%v\n",
		modelColumn.DataType, modelColumn.IsNullable, modelColumn.DefaultValue)
	fmt.Printf("DEBUG:   DB: DataType=%s, IsNullable=%t, DefaultValue=%v\n",
		dbColumn.DataType, dbColumn.IsNullable, dbColumn.DefaultValue)

	// Compare data types (normalize for comparison)
	if !di.isDataTypeCompatible(modelColumn.DataType, dbColumn.DataType) {
		fmt.Printf("DEBUG:   -> Data type mismatch: %s vs %s\n", modelColumn.DataType, dbColumn.DataType)
		return true
	}

	// Compare nullable
	if modelColumn.IsNullable != dbColumn.IsNullable {
		fmt.Printf("DEBUG:   -> Nullable mismatch: %t vs %t\n", modelColumn.IsNullable, dbColumn.IsNullable)
		return true
	}

	// Compare default values
	if (modelColumn.DefaultValue == nil) != (dbColumn.DefaultValue == nil) {
		fmt.Printf("DEBUG:   -> Default value existence mismatch\n")
		return true
	}
	if modelColumn.DefaultValue != nil && dbColumn.DefaultValue != nil &&
		*modelColumn.DefaultValue != *dbColumn.DefaultValue {
		fmt.Printf("DEBUG:   -> Default value content mismatch: %s vs %s\n",
			*modelColumn.DefaultValue, *dbColumn.DefaultValue)
		return true
	}

	// Compare length constraints
	if (modelColumn.MaxLength == nil) != (dbColumn.MaxLength == nil) {
		fmt.Printf("DEBUG:   -> Max length existence mismatch\n")
		return true
	}
	if modelColumn.MaxLength != nil && dbColumn.MaxLength != nil &&
		*modelColumn.MaxLength != *dbColumn.MaxLength {
		fmt.Printf("DEBUG:   -> Max length value mismatch: %d vs %d\n",
			*modelColumn.MaxLength, *dbColumn.MaxLength)
		return true
	}

	fmt.Printf("DEBUG:   -> No changes detected\n")
	return false
}

// isDataTypeCompatible checks if model and database data types are compatible
func (di *DatabaseInspector) isDataTypeCompatible(modelType, dbType string) bool {
	// Normalize types for comparison
	modelType = strings.ToUpper(strings.TrimSpace(modelType))
	dbType = strings.ToUpper(strings.TrimSpace(dbType))

	// Direct match
	if modelType == dbType {
		return true
	}

	// Common type mappings
	typeMap := map[string][]string{
		"VARCHAR":   {"CHARACTER VARYING", "TEXT"},
		"TEXT":      {"CHARACTER VARYING", "VARCHAR"},
		"INTEGER":   {"INT", "INT4", "SERIAL"},
		"BIGINT":    {"INT8", "BIGSERIAL"},
		"BOOLEAN":   {"BOOL"},
		"TIMESTAMP": {"TIMESTAMPTZ", "TIMESTAMP WITH TIME ZONE"},
		"DECIMAL":   {"NUMERIC"},
	}

	if alternatives, exists := typeMap[modelType]; exists {
		for _, alt := range alternatives {
			if strings.HasPrefix(dbType, alt) {
				return true
			}
		}
	}

	if alternatives, exists := typeMap[dbType]; exists {
		for _, alt := range alternatives {
			if strings.HasPrefix(modelType, alt) {
				return true
			}
		}
	}

	return false
}

// compareIndexes compares indexes between model and database
func (di *DatabaseInspector) compareIndexes(
	snapshot *ModelSnapshot,
	dbTable *TableSchema,
	modelName string,
) []MigrationChange {
	var changes []MigrationChange

	// Find indexes to create
	for indexName, indexInfo := range snapshot.Indexes {
		if _, exists := dbTable.Indexes[indexName]; !exists {
			changes = append(changes, MigrationChange{
				Type:      CreateIndex,
				TableName: snapshot.TableName,
				ModelName: modelName,
				IndexName: indexName,
				NewValue:  indexInfo,
			})
		}
	}

	// Find indexes to drop
	for indexName, dbIndex := range dbTable.Indexes {
		if _, exists := snapshot.Indexes[indexName]; !exists {
			changes = append(changes, MigrationChange{
				Type:      DropIndex,
				TableName: snapshot.TableName,
				IndexName: indexName,
				OldValue:  dbIndex,
			})
		}
	}

	return changes
}

// isSystemTable checks if a table is a system table that should be excluded from migrations
func (di *DatabaseInspector) isSystemTable(tableName string) bool {
	systemTables := []string{
		"__migration_history",
		"__ef_migrations_history",     // EF migration system table
		"__ef_migration_history",      // EF migration detailed history table
		"__model_snapshot",            // EF migration model snapshot table
		"schema_migrations",           // Common Rails/Laravel naming
		"flyway_schema_history",       // Flyway
		"liquibase_databasechangelog", // Liquibase
		"migration_versions",          // Some frameworks
	}

	for _, systemTable := range systemTables {
		if tableName == systemTable {
			return true
		}
	}

	// Also check for SQLite system tables
	if strings.HasPrefix(tableName, "sqlite_") {
		return true
	}

	return false
}

// convertDatabaseColumnToColumnInfo converts DatabaseColumnInfo to ColumnInfo
func (di *DatabaseInspector) convertDatabaseColumnToColumnInfo(dbColumn *DatabaseColumnInfo) *ColumnInfo {
	return &ColumnInfo{
		Name:         dbColumn.Name,
		DataType:     dbColumn.DataType,
		SQLType:      dbColumn.DataType, // Use same as DataType for database columns
		IsNullable:   dbColumn.IsNullable,
		DefaultValue: dbColumn.DefaultValue,
		MaxLength:    dbColumn.MaxLength,
		Precision:    dbColumn.Precision,
		Scale:        dbColumn.Scale,
		IsIdentity:   dbColumn.IsIdentity,
	}
}

package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/lamboktulussimamora/gra/logger"
)

// Logger message constants
const (
	logFailedToCloseRows      = "Failed to close rows: %v"
	logFailedToCloseTableRows = "Failed to close tableRows: %v"
	logFailedToCloseColRows   = "Failed to close colRows: %v"
)

// DatabaseInspector reads current database schema state
type DatabaseInspector struct {
	db     *sql.DB
	driver DatabaseDriver
	logger *logger.Logger
}

// NewDatabaseInspector creates a new database inspector
func NewDatabaseInspector(db *sql.DB, driver DatabaseDriver) *DatabaseInspector {
	return &DatabaseInspector{
		db:     db,
		driver: driver,
		logger: logger.New("DB_INSPECTOR"),
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
	defer func() {
		if closeErr := tableRows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseTableRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := tableRows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseTableRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseRows, closeErr)
		}
	}()

	for rows.Next() {
		if err := di.processSQLiteIndexRow(rows, table); err != nil {
			return err
		}
	}

	return nil
}

// processSQLiteIndexRow processes a single index row from SQLite PRAGMA index_list
func (di *DatabaseInspector) processSQLiteIndexRow(rows *sql.Rows, table *TableSchema) error {
	var seq int
	var name, unique, origin string
	var partial int

	if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
		return fmt.Errorf("failed to scan index info: %w", err)
	}

	// Skip auto-created indexes for primary keys and unique constraints
	if strings.HasPrefix(name, "sqlite_autoindex_") {
		return nil
	}

	index := &IndexInfo{
		Name:   name,
		Unique: unique == "1",
		Type:   "btree", // SQLite primarily uses btree indexes
	}

	// Get index columns
	columns, err := di.getSQLiteIndexColumns(name)
	if err != nil {
		return err
	}

	index.Columns = columns
	table.Indexes[name] = index
	return nil
}

// getSQLiteIndexColumns retrieves the columns for a specific SQLite index
func (di *DatabaseInspector) getSQLiteIndexColumns(indexName string) ([]string, error) {
	colRows, err := di.db.Query(fmt.Sprintf("PRAGMA index_info(%s)", indexName))
	if err != nil {
		return nil, fmt.Errorf("failed to get index columns: %w", err)
	}
	defer func() {
		if closeErr := colRows.Close(); closeErr != nil {
			di.logger.Warnf(logFailedToCloseColRows, closeErr)
		}
	}()

	var columns []string
	for colRows.Next() {
		var seqno, cid int
		var colName string
		if err := colRows.Scan(&seqno, &cid, &colName); err != nil {
			return nil, fmt.Errorf("failed to scan index column: %w", err)
		}
		columns = append(columns, colName)
	}

	return columns, nil
}

// parseSQLiteDataType parses SQLite data type to extract length, precision, scale
func (di *DatabaseInspector) parseSQLiteDataType(column *DatabaseColumnInfo, dataType string) {
	upperType := strings.ToUpper(dataType)

	// Handle different data type categories
	if di.isCharacterType(upperType) {
		di.parseCharacterTypeLength(column, upperType)
	} else if di.isNumericType(upperType) {
		di.parseNumericTypePrecision(column, upperType)
	}
}

// isCharacterType checks if the data type is a character type
func (di *DatabaseInspector) isCharacterType(upperType string) bool {
	return strings.Contains(upperType, "VARCHAR") || strings.Contains(upperType, "CHAR")
}

// isNumericType checks if the data type is a numeric type
func (di *DatabaseInspector) isNumericType(upperType string) bool {
	return strings.Contains(upperType, "DECIMAL") || strings.Contains(upperType, "NUMERIC")
}

// parseCharacterTypeLength extracts length for VARCHAR, CHAR, etc.
func (di *DatabaseInspector) parseCharacterTypeLength(column *DatabaseColumnInfo, upperType string) {
	params := di.extractTypeParameters(upperType)
	if params != "" {
		if length := di.parseIntValue(params); length > 0 {
			column.MaxLength = &length
		}
	}
}

// parseNumericTypePrecision extracts precision and scale for DECIMAL, NUMERIC
func (di *DatabaseInspector) parseNumericTypePrecision(column *DatabaseColumnInfo, upperType string) {
	params := di.extractTypeParameters(upperType)
	if params == "" {
		return
	}

	parts := strings.Split(params, ",")
	di.setPrecisionFromParts(column, parts)
	di.setScaleFromParts(column, parts)
}

// extractTypeParameters extracts parameters from type definition like VARCHAR(255) -> "255"
func (di *DatabaseInspector) extractTypeParameters(upperType string) string {
	start := strings.Index(upperType, "(")
	if start == -1 {
		return ""
	}

	end := strings.Index(upperType[start:], ")")
	if end == -1 {
		return ""
	}

	return upperType[start+1 : start+end]
}

// setPrecisionFromParts sets precision from the first part of parameters
func (di *DatabaseInspector) setPrecisionFromParts(column *DatabaseColumnInfo, parts []string) {
	if len(parts) >= 1 {
		if precision := di.parseIntValue(strings.TrimSpace(parts[0])); precision > 0 {
			column.Precision = &precision
		}
	}
}

// setScaleFromParts sets scale from the second part of parameters
func (di *DatabaseInspector) setScaleFromParts(column *DatabaseColumnInfo, parts []string) {
	if len(parts) >= 2 {
		if scale := di.parseIntValue(strings.TrimSpace(parts[1])); scale >= 0 {
			column.Scale = &scale
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
	processedTables := make(map[string]bool)

	// Process model snapshots for create/alter operations
	modelChanges := di.processModelSnapshots(dbSchema, modelSnapshots, processedTables)
	changes = append(changes, modelChanges...)

	// Process database tables for drop operations
	dropChanges := di.processDroppedTables(dbSchema, processedTables)
	changes = append(changes, dropChanges...)

	// Log final results
	di.logComparisonResults(changes)

	return changes, nil
}

// processModelSnapshots handles table creation and column changes for model snapshots
func (di *DatabaseInspector) processModelSnapshots(dbSchema map[string]*TableSchema, modelSnapshots map[string]*ModelSnapshot, processedTables map[string]bool) []MigrationChange {
	var changes []MigrationChange

	for modelName, snapshot := range modelSnapshots {
		tableName := snapshot.TableName
		processedTables[tableName] = true

		di.logger.Debugf("Processing model %s -> table %s", modelName, tableName)

		if _, exists := dbSchema[tableName]; !exists {
			// Table doesn't exist in database - create it
			di.logger.Debugf("Table %s does not exist in database, creating CreateTable change", tableName)
			changes = append(changes, MigrationChange{
				Type:      CreateTable,
				TableName: tableName,
				ModelName: modelName,
				NewValue:  snapshot,
			})
		} else {
			// Table exists - check for column changes
			di.logger.Debugf("Table %s exists, checking for column changes", tableName)
			columnChanges := di.compareTableColumns(dbSchema[tableName], snapshot)
			changes = append(changes, columnChanges...)
		}
	}

	return changes
}

// processDroppedTables handles tables that exist in database but not in models
func (di *DatabaseInspector) processDroppedTables(dbSchema map[string]*TableSchema, processedTables map[string]bool) []MigrationChange {
	var changes []MigrationChange

	for tableName, tableSchema := range dbSchema {
		if di.isSystemTable(tableName) {
			di.logger.Debugf("Skipping system table %s", tableName)
			continue
		}

		if !processedTables[tableName] {
			di.logger.Debugf("Table %s exists in database but not in models, creating DropTable change", tableName)
			changes = append(changes, MigrationChange{
				Type:      DropTable,
				TableName: tableName,
				OldValue:  tableSchema,
			})
		}
	}

	return changes
}

// logComparisonResults logs the final comparison results
func (di *DatabaseInspector) logComparisonResults(changes []MigrationChange) {
	di.logger.Debugf("CompareWithModelSnapshot: Generated %d changes", len(changes))
	for i, change := range changes {
		di.logger.Debugf("Change %d: %s %s.%s", i, change.Type, change.TableName, change.ColumnName)
	}
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
			di.logger.Debugf("Column %s.%s does not exist in database, creating AddColumn change", dbTable.Name, columnName)
			changes = append(changes, MigrationChange{
				Type:       AddColumn,
				TableName:  dbTable.Name,
				ColumnName: columnName,
				NewColumn:  modelColumn,
			})
		} else if di.hasColumnChanged(modelColumn, dbColumn) {
			// Column exists - check if it has changed
			di.logger.Debugf("Column %s.%s has changed, creating AlterColumn change", dbTable.Name, columnName)
			changes = append(changes, MigrationChange{
				Type:       AlterColumn,
				TableName:  dbTable.Name,
				ColumnName: columnName,
				OldColumn:  di.convertDatabaseColumnToColumnInfo(dbColumn),
				NewColumn:  modelColumn,
			})
		}
	}

	// Check for columns to drop (exist in database but not in model)
	for columnName, dbColumn := range dbTable.Columns {
		if !processedColumns[columnName] {
			di.logger.Debugf("Column %s.%s exists in database but not in model, creating DropColumn change", dbTable.Name, columnName)
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

// hasColumnChanged checks if a column definition has changed
func (di *DatabaseInspector) hasColumnChanged(modelColumn *ColumnInfo, dbColumn *DatabaseColumnInfo) bool {
	// Debug: Log column comparison
	di.logger.Debugf("Comparing column %s:", dbColumn.Name)
	di.logger.Debugf("  Model: DataType=%s, IsNullable=%t, DefaultValue=%v",
		modelColumn.DataType, modelColumn.IsNullable, modelColumn.DefaultValue)
	di.logger.Debugf("  DB: DataType=%s, IsNullable=%t, DefaultValue=%v",
		dbColumn.DataType, dbColumn.IsNullable, dbColumn.DefaultValue)

	// Compare data types (normalize for comparison)
	if !di.isDataTypeCompatible(modelColumn.DataType, dbColumn.DataType) {
		di.logger.Debugf("  -> Data type mismatch: %s vs %s", modelColumn.DataType, dbColumn.DataType)
		return true
	}

	// Compare nullable
	if modelColumn.IsNullable != dbColumn.IsNullable {
		di.logger.Debugf("  -> Nullable mismatch: %t vs %t", modelColumn.IsNullable, dbColumn.IsNullable)
		return true
	}

	// Compare default values
	if (modelColumn.DefaultValue == nil) != (dbColumn.DefaultValue == nil) {
		di.logger.Debugf("  -> Default value existence mismatch")
		return true
	}
	if modelColumn.DefaultValue != nil && dbColumn.DefaultValue != nil &&
		*modelColumn.DefaultValue != *dbColumn.DefaultValue {
		di.logger.Debugf("  -> Default value content mismatch: %s vs %s",
			*modelColumn.DefaultValue, *dbColumn.DefaultValue)
		return true
	}

	// Compare length constraints
	if (modelColumn.MaxLength == nil) != (dbColumn.MaxLength == nil) {
		di.logger.Debugf("  -> Max length existence mismatch")
		return true
	}
	if modelColumn.MaxLength != nil && dbColumn.MaxLength != nil &&
		*modelColumn.MaxLength != *dbColumn.MaxLength {
		di.logger.Debugf("  -> Max length value mismatch: %d vs %d",
			*modelColumn.MaxLength, *dbColumn.MaxLength)
		return true
	}

	di.logger.Debugf("  -> No changes detected")
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

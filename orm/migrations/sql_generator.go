package migrations

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SQLGenerator generates SQL migration scripts from migration changes
type SQLGenerator struct {
	driver DatabaseDriver
}

// NewSQLGenerator creates a new SQL generator for the specified database driver
func NewSQLGenerator(driver DatabaseDriver) *SQLGenerator {
	return &SQLGenerator{
		driver: driver,
	}
}

// GenerateMigrationSQL generates SQL scripts for a migration plan
func (sg *SQLGenerator) GenerateMigrationSQL(plan *MigrationPlan) (*MigrationSQL, error) {
	if len(plan.Changes) == 0 {
		return &MigrationSQL{
			UpScript:   "-- No changes detected\n",
			DownScript: "-- No changes to revert\n",
		}, nil
	}

	upScript, err := sg.generateUpScript(plan.Changes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate up script: %w", err)
	}

	downScript, err := sg.generateDownScript(plan.Changes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate down script: %w", err)
	}

	return &MigrationSQL{
		UpScript:   upScript,
		DownScript: downScript,
		Metadata: MigrationMetadata{
			Timestamp:      time.Now(),
			Checksum:       plan.PlanChecksum,
			HasDestructive: plan.HasDestructive,
			RequiresReview: plan.RequiresReview,
			ChangeCount:    len(plan.Changes),
		},
	}, nil
}

// MigrationSQL contains the generated SQL scripts
type MigrationSQL struct {
	UpScript   string
	DownScript string
	Metadata   MigrationMetadata
}

// MigrationMetadata contains metadata about the migration
type MigrationMetadata struct {
	Timestamp      time.Time
	Checksum       string
	HasDestructive bool
	RequiresReview bool
	ChangeCount    int
}

// generateUpScript generates the up migration script
func (sg *SQLGenerator) generateUpScript(changes []MigrationChange) (string, error) {
	statements := make([]string, 0, len(changes))
	var comments []string

	// Add header comment
	comments = append(comments, "-- Migration Up Script")
	comments = append(comments, fmt.Sprintf("-- Generated at: %s", time.Now().Format(time.RFC3339)))
	comments = append(comments, fmt.Sprintf("-- Changes: %d", len(changes)))
	comments = append(comments, "")

	// Group changes by type for better organization
	groupedChanges := sg.groupChangesByType(changes)

	// Process changes in order
	for _, changeType := range []ChangeType{CreateTable, AddColumn, AlterColumn, CreateIndex, DropIndex, DropColumn, DropTable} {
		if changeList, exists := groupedChanges[changeType]; exists {
			typeComment := fmt.Sprintf("-- %s (%d)", sg.getChangeTypeDescription(changeType), len(changeList))
			comments = append(comments, typeComment)

			for _, change := range changeList {
				sql, err := sg.generateChangeSQL(change, true)
				if err != nil {
					return "", fmt.Errorf("failed to generate SQL for change %+v: %w", change, err)
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
			comments = append(comments, "")
		}
	}

	// Combine comments and statements
	script := strings.Join(comments, "\n")
	if len(statements) > 0 {
		script += "\n" + strings.Join(statements, "\n\n") + "\n"
	}

	return script, nil
}

// generateDownScript generates the down migration script
func (sg *SQLGenerator) generateDownScript(changes []MigrationChange) (string, error) {
	statements := make([]string, 0, len(changes))
	var comments []string

	// Add header comment
	comments = append(comments, "-- Migration Down Script")
	comments = append(comments, fmt.Sprintf("-- Generated at: %s", time.Now().Format(time.RFC3339)))
	comments = append(comments, "-- Reverses changes from up script")
	comments = append(comments, "")

	// Reverse the order and invert operations
	reversedChanges := sg.reverseChanges(changes)

	// Group reversed changes
	groupedChanges := sg.groupChangesByType(reversedChanges)

	// Process reversed changes
	for _, changeType := range []ChangeType{DropIndex, DropColumn, DropTable, CreateIndex, AlterColumn, AddColumn, CreateTable} {
		if changeList, exists := groupedChanges[changeType]; exists {
			typeComment := fmt.Sprintf("-- %s (%d)", sg.getChangeTypeDescription(changeType), len(changeList))
			comments = append(comments, typeComment)

			for _, change := range changeList {
				sql, err := sg.generateChangeSQL(change, false)
				if err != nil {
					return "", fmt.Errorf("failed to generate reverse SQL for change %+v: %w", change, err)
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
			comments = append(comments, "")
		}
	}

	// Combine comments and statements
	script := strings.Join(comments, "\n")
	if len(statements) > 0 {
		script += "\n" + strings.Join(statements, "\n\n") + "\n"
	}

	return script, nil
}

// groupChangesByType groups changes by their type
func (sg *SQLGenerator) groupChangesByType(changes []MigrationChange) map[ChangeType][]MigrationChange {
	grouped := make(map[ChangeType][]MigrationChange)

	for _, change := range changes {
		grouped[change.Type] = append(grouped[change.Type], change)
	}

	// Sort changes within each group
	for changeType := range grouped {
		sort.Slice(grouped[changeType], func(i, j int) bool {
			return sg.compareChangesForSQL(grouped[changeType][i], grouped[changeType][j])
		})
	}

	return grouped
}

// compareChangesForSQL provides ordering for changes within SQL generation
func (sg *SQLGenerator) compareChangesForSQL(a, b MigrationChange) bool {
	// Sort by table name first
	if a.TableName != b.TableName {
		return a.TableName < b.TableName
	}

	// Then by column/index name
	if a.ColumnName != b.ColumnName {
		return a.ColumnName < b.ColumnName
	}

	return a.IndexName < b.IndexName
}

// getChangeTypeDescription returns a human-readable description for change types
func (sg *SQLGenerator) getChangeTypeDescription(changeType ChangeType) string {
	descriptions := map[ChangeType]string{
		CreateTable: "Create Tables",
		DropTable:   "Drop Tables",
		AddColumn:   "Add Columns",
		DropColumn:  "Drop Columns",
		AlterColumn: "Alter Columns",
		CreateIndex: "Create Indexes",
		DropIndex:   "Drop Indexes",
	}

	if desc, exists := descriptions[changeType]; exists {
		return desc
	}
	return string(changeType)
}

// reverseChanges creates reversed changes for down script
func (sg *SQLGenerator) reverseChanges(changes []MigrationChange) []MigrationChange {
	reversed := make([]MigrationChange, 0, len(changes))

	// Process in reverse order
	for i := len(changes) - 1; i >= 0; i-- {
		change := changes[i]
		reversedChange := sg.reverseChange(change)
		if reversedChange != nil {
			reversed = append(reversed, *reversedChange)
		}
	}

	return reversed
}

// reverseChange creates the reverse of a single change
func (sg *SQLGenerator) reverseChange(change MigrationChange) *MigrationChange {
	switch change.Type {
	case CreateTable:
		return &MigrationChange{
			Type:      DropTable,
			TableName: change.TableName,
			ModelName: change.ModelName,
			OldValue:  change.NewValue,
		}
	case DropTable:
		return &MigrationChange{
			Type:      CreateTable,
			TableName: change.TableName,
			ModelName: change.ModelName,
			NewValue:  change.OldValue,
		}
	case AddColumn:
		return &MigrationChange{
			Type:       DropColumn,
			TableName:  change.TableName,
			ColumnName: change.ColumnName,
			OldValue:   change.NewValue,
		}
	case DropColumn:
		return &MigrationChange{
			Type:       AddColumn,
			TableName:  change.TableName,
			ColumnName: change.ColumnName,
			NewValue:   change.OldValue,
		}
	case AlterColumn:
		return &MigrationChange{
			Type:       AlterColumn,
			TableName:  change.TableName,
			ColumnName: change.ColumnName,
			OldValue:   change.NewValue,
			NewValue:   change.OldValue,
		}
	case CreateIndex:
		return &MigrationChange{
			Type:      DropIndex,
			TableName: change.TableName,
			IndexName: change.IndexName,
			OldValue:  change.NewValue,
		}
	case DropIndex:
		return &MigrationChange{
			Type:      CreateIndex,
			TableName: change.TableName,
			IndexName: change.IndexName,
			NewValue:  change.OldValue,
		}
	default:
		return nil // Unsupported change type
	}
}

// generateChangeSQL generates SQL for a specific change
func (sg *SQLGenerator) generateChangeSQL(change MigrationChange, isUp bool) (string, error) {
	switch change.Type {
	case CreateTable:
		return sg.generateCreateTableSQL(change)
	case DropTable:
		return sg.generateDropTableSQL(change)
	case AddColumn:
		return sg.generateAddColumnSQL(change)
	case DropColumn:
		return sg.generateDropColumnSQL(change)
	case AlterColumn:
		return sg.generateAlterColumnSQL(change)
	case CreateIndex:
		return sg.generateCreateIndexSQL(change)
	case DropIndex:
		return sg.generateDropIndexSQL(change)
	default:
		return "", fmt.Errorf("unsupported change type: %s", change.Type)
	}
}

// generateCreateTableSQL generates CREATE TABLE statement
func (sg *SQLGenerator) generateCreateTableSQL(change MigrationChange) (string, error) {
	snapshot, ok := change.NewValue.(*ModelSnapshot)
	if !ok {
		// Debug: check what type we actually have
		if change.NewValue == nil {
			return "", fmt.Errorf("invalid value type for CreateTable: NewValue is nil")
		}
		return "", fmt.Errorf("invalid value type for CreateTable: expected *ModelSnapshot, got %T", change.NewValue)
	}

	var statements []string
	var columnDefs []string
	var primaryKeys []string

	// Sort columns for consistent output
	columnNames := make([]string, 0, len(snapshot.Columns))
	for name := range snapshot.Columns {
		columnNames = append(columnNames, name)
	}
	sort.Strings(columnNames)

	for _, columnName := range columnNames {
		column := snapshot.Columns[columnName]
		columnDef := sg.generateColumnDefinition(column)
		columnDefs = append(columnDefs, fmt.Sprintf("    %s %s", columnName, columnDef))

		// For SQLite, skip adding to primaryKeys if it's an identity column (already has inline PRIMARY KEY)
		if column.IsPrimaryKey {
			if !(sg.driver == SQLite && column.IsIdentity) {
				primaryKeys = append(primaryKeys, columnName)
			}
		}
	}

	// Add primary key constraint (only if we have primary keys that don't already have inline PRIMARY KEY)
	if len(primaryKeys) > 0 {
		pkConstraint := fmt.Sprintf("    PRIMARY KEY (%s)", strings.Join(primaryKeys, ", "))
		columnDefs = append(columnDefs, pkConstraint)
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE %s (\n%s\n);",
		sg.quoteIdentifier(snapshot.TableName),
		strings.Join(columnDefs, ",\n"))

	statements = append(statements, createTableSQL)

	// Create indexes
	for indexName, index := range snapshot.Indexes {
		indexSQL := sg.generateCreateIndexStatement(snapshot.TableName, indexName, &index)
		statements = append(statements, indexSQL)
	}

	// Add foreign key constraints
	for constraintName, constraint := range snapshot.Constraints {
		if constraint.Type == foreignKeyConstraintType {
			fkSQL := sg.generateAddForeignKeySQL(snapshot.TableName, constraintName, constraint)
			statements = append(statements, fkSQL)
		}
	}

	return strings.Join(statements, "\n\n"), nil
}

// generateColumnDefinition generates column definition SQL
func (sg *SQLGenerator) generateColumnDefinition(column *ColumnInfo) string {
	var parts []string

	// Debug: log column info
	fmt.Printf("DEBUG: Column info: Name=%s, Type=%s, SQLType=%s, DataType=%s\n",
		column.Name, column.Type, column.SQLType, column.DataType)

	// Data type - prefer SQLType over DataType
	var dataType string
	if column.SQLType != "" {
		dataType = column.SQLType
	} else if column.DataType != "" {
		dataType = sg.mapDataType(column.DataType)
	} else {
		dataType = sg.mapDataType(column.Type)
	}

	if column.MaxLength != nil && sg.supportsLength(dataType) {
		dataType = fmt.Sprintf("%s(%d)", dataType, *column.MaxLength)
	} else if column.Precision != nil && column.Scale != nil {
		dataType = fmt.Sprintf("%s(%d,%d)", dataType, *column.Precision, *column.Scale)
	}
	parts = append(parts, dataType)

	// Nullable
	if !column.IsNullable {
		parts = append(parts, "NOT NULL")
	}

	// Default value
	if column.DefaultValue != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", *column.DefaultValue))
	}

	// Auto increment
	if column.IsIdentity && sg.driver == PostgreSQL {
		// For PostgreSQL, use SERIAL or BIGSERIAL
		if strings.ToUpper(column.DataType) == "BIGINT" {
			parts[0] = "BIGSERIAL"
		} else {
			parts[0] = "SERIAL"
		}
	} else if column.IsIdentity && sg.driver == MySQL {
		parts = append(parts, "AUTO_INCREMENT")
	} else if column.IsIdentity && sg.driver == SQLite {
		// For SQLite, identity primary key columns should be INTEGER PRIMARY KEY AUTOINCREMENT
		if column.IsPrimaryKey {
			parts[0] = "INTEGER"
			parts = append(parts, "PRIMARY KEY")
			parts = append(parts, "AUTOINCREMENT")
		}
	}

	return strings.Join(parts, " ")
}

// mapDataType maps Go/generic types to database-specific types
func (sg *SQLGenerator) mapDataType(dataType string) string {
	switch sg.driver {
	case PostgreSQL:
		return sg.mapPostgreSQLType(dataType)
	case MySQL:
		return sg.mapMySQLType(dataType)
	case SQLite:
		return sg.mapSQLiteType(dataType)
	default:
		return dataType
	}
}

// mapPostgreSQLType maps types for PostgreSQL
func (sg *SQLGenerator) mapPostgreSQLType(dataType string) string {
	typeMap := map[string]string{
		"STRING":  "VARCHAR",
		"TEXT":    "TEXT",
		"INT":     "INTEGER",
		"INT64":   "BIGINT",
		"FLOAT64": "DOUBLE PRECISION",
		"BOOL":    "BOOLEAN",
		"TIME":    "TIMESTAMP",
		"DECIMAL": "DECIMAL",
		"BYTES":   "BYTEA",
	}

	if mapped, exists := typeMap[strings.ToUpper(dataType)]; exists {
		return mapped
	}
	return dataType
}

// mapMySQLType maps types for MySQL
func (sg *SQLGenerator) mapMySQLType(dataType string) string {
	typeMap := map[string]string{
		"STRING":  "VARCHAR",
		"TEXT":    "TEXT",
		"INT":     "INT",
		"INT64":   "BIGINT",
		"FLOAT64": "DOUBLE",
		"BOOL":    "BOOLEAN",
		"TIME":    "TIMESTAMP",
		"DECIMAL": "DECIMAL",
		"BYTES":   "BLOB",
	}

	if mapped, exists := typeMap[strings.ToUpper(dataType)]; exists {
		return mapped
	}
	return dataType
}

// mapSQLiteType maps types for SQLite
func (sg *SQLGenerator) mapSQLiteType(dataType string) string {
	typeMap := map[string]string{
		"STRING":  "TEXT",
		"TEXT":    "TEXT",
		"INT":     "INTEGER",
		"INT64":   "INTEGER",
		"FLOAT64": "REAL",
		"BOOL":    "INTEGER",
		"TIME":    "TEXT",
		"DECIMAL": "REAL",
		"BYTES":   "BLOB",
	}

	if mapped, exists := typeMap[strings.ToUpper(dataType)]; exists {
		return mapped
	}
	return dataType
}

// supportsLength checks if a data type supports length specification
func (sg *SQLGenerator) supportsLength(dataType string) bool {
	lengthTypes := map[string]bool{
		"VARCHAR": true,
		"CHAR":    true,
		"STRING":  true,
	}
	return lengthTypes[strings.ToUpper(dataType)]
}

// generateDropTableSQL generates DROP TABLE statement
func (sg *SQLGenerator) generateDropTableSQL(change MigrationChange) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", sg.quoteIdentifier(change.TableName)), nil
}

// generateAddColumnSQL generates ADD COLUMN statement
func (sg *SQLGenerator) generateAddColumnSQL(change MigrationChange) (string, error) {
	column, ok := change.NewValue.(*ColumnInfo)
	if !ok {
		return "", fmt.Errorf("invalid value type for AddColumn: expected *ColumnInfo")
	}

	columnDef := sg.generateColumnDefinition(column)
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;",
		sg.quoteIdentifier(change.TableName),
		sg.quoteIdentifier(change.ColumnName),
		columnDef), nil
}

// generateDropColumnSQL generates DROP COLUMN statement
func (sg *SQLGenerator) generateDropColumnSQL(change MigrationChange) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;",
		sg.quoteIdentifier(change.TableName),
		sg.quoteIdentifier(change.ColumnName)), nil
}

// generateAlterColumnSQL generates ALTER COLUMN statement
func (sg *SQLGenerator) generateAlterColumnSQL(change MigrationChange) (string, error) {
	newColumn, ok := change.NewValue.(*ColumnInfo)
	if !ok {
		return "", fmt.Errorf("invalid value type for AlterColumn: expected *ColumnInfo")
	}

	// PostgreSQL and MySQL have different syntax for altering columns
	switch sg.driver {
	case PostgreSQL:
		return sg.generatePostgreSQLAlterColumn(change.TableName, change.ColumnName, newColumn)
	case MySQL:
		return sg.generateMySQLAlterColumn(change.TableName, change.ColumnName, newColumn)
	case SQLite:
		return "", fmt.Errorf("SQLite does not support ALTER COLUMN directly")
	default:
		return "", fmt.Errorf("unsupported driver for ALTER COLUMN: %s", sg.driver)
	}
}

// generatePostgreSQLAlterColumn generates PostgreSQL-specific ALTER COLUMN
func (sg *SQLGenerator) generatePostgreSQLAlterColumn(tableName, columnName string, column *ColumnInfo) (string, error) {
	var statements []string
	statements = make([]string, 0, 2)

	// Alter data type
	dataType := sg.mapDataType(column.DataType)
	if column.MaxLength != nil && sg.supportsLength(column.DataType) {
		dataType = fmt.Sprintf("%s(%d)", dataType, *column.MaxLength)
	}

	statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;",
		sg.quoteIdentifier(tableName),
		sg.quoteIdentifier(columnName),
		dataType))

	// Alter nullable
	if column.IsNullable {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;",
			sg.quoteIdentifier(tableName),
			sg.quoteIdentifier(columnName)))
	} else {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;",
			sg.quoteIdentifier(tableName),
			sg.quoteIdentifier(columnName)))
	}

	// Alter default
	if column.DefaultValue != nil {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;",
			sg.quoteIdentifier(tableName),
			sg.quoteIdentifier(columnName),
			*column.DefaultValue))
	} else {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;",
			sg.quoteIdentifier(tableName),
			sg.quoteIdentifier(columnName)))
	}

	return strings.Join(statements, "\n"), nil
}

// generateMySQLAlterColumn generates MySQL-specific ALTER COLUMN
func (sg *SQLGenerator) generateMySQLAlterColumn(tableName, columnName string, column *ColumnInfo) (string, error) {
	columnDef := sg.generateColumnDefinition(column)
	return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s;",
		sg.quoteIdentifier(tableName),
		sg.quoteIdentifier(columnName),
		columnDef), nil
}

// generateCreateIndexSQL generates CREATE INDEX statement
func (sg *SQLGenerator) generateCreateIndexSQL(change MigrationChange) (string, error) {
	index, ok := change.NewValue.(*IndexInfo)
	if !ok {
		return "", fmt.Errorf("invalid value type for CreateIndex: expected *IndexInfo")
	}

	return sg.generateCreateIndexStatement(change.TableName, change.IndexName, index), nil
}

// generateCreateIndexStatement generates CREATE INDEX statement
func (sg *SQLGenerator) generateCreateIndexStatement(tableName, indexName string, index *IndexInfo) string {
	uniqueClause := ""
	if index.IsUnique {
		uniqueClause = "UNIQUE "
	}

	columns := make([]string, len(index.Columns))
	for i, col := range index.Columns {
		columns[i] = sg.quoteIdentifier(col)
	}

	return fmt.Sprintf("CREATE %sINDEX %s ON %s (%s);",
		uniqueClause,
		sg.quoteIdentifier(indexName),
		sg.quoteIdentifier(tableName),
		strings.Join(columns, ", "))
}

// generateDropIndexSQL generates DROP INDEX statement
func (sg *SQLGenerator) generateDropIndexSQL(change MigrationChange) (string, error) {
	switch sg.driver {
	case PostgreSQL:
		return fmt.Sprintf("DROP INDEX IF EXISTS %s;", sg.quoteIdentifier(change.IndexName)), nil
	case MySQL:
		return fmt.Sprintf("DROP INDEX %s ON %s;",
			sg.quoteIdentifier(change.IndexName),
			sg.quoteIdentifier(change.TableName)), nil
	case SQLite:
		return fmt.Sprintf("DROP INDEX IF EXISTS %s;", sg.quoteIdentifier(change.IndexName)), nil
	default:
		return "", fmt.Errorf("unsupported driver for DROP INDEX: %s", sg.driver)
	}
}

// generateAddForeignKeySQL generates ADD FOREIGN KEY constraint
func (sg *SQLGenerator) generateAddForeignKeySQL(tableName, constraintName string, constraint *ConstraintInfo) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s);",
		sg.quoteIdentifier(tableName),
		sg.quoteIdentifier(constraintName),
		strings.Join(sg.quoteIdentifiers(constraint.Columns), ", "),
		sg.quoteIdentifier(constraint.ReferencedTable),
		strings.Join(sg.quoteIdentifiers(constraint.ReferencedColumns), ", "))
}

// quoteIdentifier quotes an identifier for the target database
func (sg *SQLGenerator) quoteIdentifier(identifier string) string {
	switch sg.driver {
	case PostgreSQL:
		return fmt.Sprintf(`"%s"`, identifier)
	case MySQL:
		return fmt.Sprintf("`%s`", identifier)
	case SQLite:
		return fmt.Sprintf(`"%s"`, identifier)
	default:
		return identifier
	}
}

// quoteIdentifiers quotes multiple identifiers
func (sg *SQLGenerator) quoteIdentifiers(identifiers []string) []string {
	quoted := make([]string, len(identifiers))
	for i, id := range identifiers {
		quoted[i] = sg.quoteIdentifier(id)
	}
	return quoted
}

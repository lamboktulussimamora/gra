package migrations

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Common SQL type and tag constants for model registry
const (
	indexTrueValue   = "true"
	sqlTypeBigInt    = "BIGINT"
	sqlTypeBigSerial = "BIGSERIAL"
	sqlTypeSerial    = "SERIAL"
	sqlTypeReal      = "REAL"
	foreignKeyTag    = "foreign_key:"
)

// NewModelRegistry creates a new model registry
func NewModelRegistry(driver DatabaseDriver) *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelSnapshot),
		driver: driver,
	}
}

// RegisterModel registers a model in the registry
func (mr *ModelRegistry) RegisterModel(model interface{}) {
	snapshot := mr.createModelSnapshot(model)
	mr.models[snapshot.TableName] = &snapshot
}

// GetModels returns all registered models
func (mr *ModelRegistry) GetModels() map[string]*ModelSnapshot {
	return mr.models
}

// createModelSnapshot creates a snapshot of a model's schema
func (mr *ModelRegistry) createModelSnapshot(model interface{}) ModelSnapshot {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	tableName := mr.getTableName(model)
	columns := make(map[string]*ColumnInfo)
	indexes := make(map[string]IndexInfo)
	constraints := make(map[string]*ConstraintInfo)

	// Process struct fields recursively
	mr.processStructFields(modelType, "", func(field reflect.StructField, dbName string, prefix string) {
		if dbName == "" || dbName == "-" {
			return // Skip fields without db tags or explicitly excluded
		}

		columnInfo := mr.createColumnInfo(field, dbName)
		columns[dbName] = &columnInfo

		// Extract indexes from field tags
		mr.extractIndexInfo(field, dbName, tableName, indexes)

		// Extract constraints from field tags
		mr.extractConstraintInfo(field, dbName, tableName, constraints)
	})

	snapshot := ModelSnapshot{
		TableName:   tableName,
		ModelType:   modelType,
		Columns:     columns,
		Indexes:     indexes,
		Constraints: constraints,
	}

	snapshot.Checksum = mr.calculateSnapshotChecksum(snapshot)
	return snapshot
}

// processStructFields recursively processes struct fields including embedded ones.
// The 'prefix' parameter is used for nested/embedded structs. (revive: parameter is used)
func (mr *ModelRegistry) processStructFields(structType reflect.Type, prefix string, callback func(reflect.StructField, string, string)) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs
		if field.Anonymous {
			if field.Type.Kind() == reflect.Struct {
				mr.processStructFields(field.Type, "", callback)
			} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				mr.processStructFields(field.Type.Elem(), "", callback)
			}
			continue
		}

		// Get database column name
		dbName := mr.getDBColumnName(field)
		if prefix != "" {
			dbName = prefix + "_" + dbName
		}

		callback(field, dbName, prefix)
	}
}

// createColumnInfo creates column information from a struct field
func (mr *ModelRegistry) createColumnInfo(field reflect.StructField, dbName string) ColumnInfo {
	fieldType := field.Type
	isNullable := false

	// Handle pointer types (nullable)
	if fieldType.Kind() == reflect.Ptr {
		isNullable = true
		fieldType = fieldType.Elem()
	}

	columnInfo := ColumnInfo{
		Name:         dbName,
		Type:         fieldType.String(),
		SQLType:      mr.getSQLType(field, fieldType),
		DataType:     mr.getSQLType(field, fieldType), // Set DataType to same as SQLType for comparison
		Nullable:     isNullable,
		IsNullable:   isNullable, // Set both fields for compatibility
		IsPrimaryKey: mr.isPrimaryKey(field),
		IsUnique:     mr.isUnique(field),
		IsIdentity:   mr.isAutoIncrement(field),
		IsForeignKey: mr.isForeignKey(field),
		Size:         mr.getSize(field),
		Precision:    mr.getPrecision(field),
		Scale:        mr.getScale(field),
	}

	// Set MaxLength from Size if Size > 0
	if columnInfo.Size > 0 {
		columnInfo.MaxLength = &columnInfo.Size
	}

	// Extract default value
	if defaultVal := field.Tag.Get("default"); defaultVal != "" {
		columnInfo.Default = &defaultVal
		columnInfo.DefaultValue = &defaultVal // Set both fields for compatibility
	}

	// Extract foreign key information
	if columnInfo.IsForeignKey {
		columnInfo.References = mr.getForeignKeyInfo(field)
	}

	return columnInfo
}

// extractIndexInfo extracts index information from field tags
func (mr *ModelRegistry) extractIndexInfo(field reflect.StructField, dbName, tableName string, indexes map[string]IndexInfo) {
	// Regular index
	if indexName := field.Tag.Get("index"); indexName != "" {
		if indexName == indexTrueValue {
			indexName = fmt.Sprintf("idx_%s_%s", tableName, dbName)
		}
		indexes[indexName] = IndexInfo{
			Name:    indexName,
			Columns: []string{dbName},
			Unique:  false,
			Type:    "btree",
		}
	}

	// Unique index
	if uniqueIndex := field.Tag.Get("uniqueIndex"); uniqueIndex != "" {
		if uniqueIndex == indexTrueValue {
			uniqueIndex = fmt.Sprintf("uidx_%s_%s", tableName, dbName)
		}
		indexes[uniqueIndex] = IndexInfo{
			Name:    uniqueIndex,
			Columns: []string{dbName},
			Unique:  true,
			Type:    "btree",
		}
	}
}

// extractConstraintInfo extracts constraint information from field tags
func (mr *ModelRegistry) extractConstraintInfo(field reflect.StructField, dbName, tableName string, constraints map[string]*ConstraintInfo) {
	// Check constraint
	if check := field.Tag.Get("check"); check != "" {
		constraintName := fmt.Sprintf("chk_%s_%s", tableName, dbName)
		constraints[constraintName] = &ConstraintInfo{
			Name: constraintName,
			Type: "CHECK",
			SQL:  fmt.Sprintf("CHECK (%s)", check),
		}
	}

	// Foreign key constraint
	if mr.isForeignKey(field) {
		fkInfo := mr.getForeignKeyInfo(field)
		if fkInfo != nil {
			constraintName := fmt.Sprintf("fk_%s_%s", tableName, dbName)
			constraints[constraintName] = &ConstraintInfo{
				Name: constraintName,
				Type: "FOREIGN_KEY",
				SQL:  fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", dbName, fkInfo.Table, fkInfo.Column),
			}
		}
	}
}

// Helper methods for field analysis
func (mr *ModelRegistry) getTableName(model interface{}) string {
	// Check if model implements TableNamer interface
	if tn, ok := model.(interface{ TableName() string }); ok {
		return tn.TableName()
	}

	// Use reflection to get type name
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	name := strings.ToLower(modelType.Name())

	// Remove common suffixes
	name = strings.TrimSuffix(name, "entity")
	name = strings.TrimSuffix(name, "model")

	// Pluralize (simple approach)
	return mr.pluralize(name)
}

func (mr *ModelRegistry) pluralize(word string) string {
	if strings.HasSuffix(word, "y") {
		return strings.TrimSuffix(word, "y") + "ies"
	}
	if strings.HasSuffix(word, "s") {
		return word + "es"
	}
	return word + "s"
}

func (mr *ModelRegistry) getDBColumnName(field reflect.StructField) string {
	if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
		// Extract just the column name (before any comma)
		parts := strings.Split(tag, ",")
		return parts[0]
	}
	return mr.toSnakeCase(field.Name)
}

func (mr *ModelRegistry) toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// Helper to extract explicit SQL type from struct tags (split for complexity)
func getExplicitSQLType(field reflect.StructField) (string, bool) {
	if sqlType, ok := getExplicitSQLTypeFromMigrationTag(field); ok {
		return sqlType, true
	}
	if sqlType, ok := getExplicitSQLTypeFromSQLTag(field); ok {
		return sqlType, true
	}
	return "", false
}

func getExplicitSQLTypeFromMigrationTag(field reflect.StructField) (string, bool) {
	migrationTag := field.Tag.Get("migration")
	if migrationTag == "" || !strings.Contains(migrationTag, "type:") {
		return "", false
	}
	for _, part := range strings.Split(migrationTag, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "type:") {
			return strings.TrimPrefix(part, "type:"), true
		}
	}
	return "", false
}

func getExplicitSQLTypeFromSQLTag(field reflect.StructField) (string, bool) {
	sqlTag := field.Tag.Get("sql")
	if sqlTag == "" || !strings.Contains(sqlTag, "type:") {
		return "", false
	}
	for _, part := range strings.Split(sqlTag, ";") {
		if strings.HasPrefix(part, "type:") {
			return strings.TrimPrefix(part, "type:"), true
		}
	}
	return "", false
}

func (mr *ModelRegistry) getSQLType(field reflect.StructField, fieldType reflect.Type) string {
	if sqlType, ok := getExplicitSQLType(field); ok {
		return sqlType
	}

	size := mr.getSize(field)
	switch fieldType.Kind() {
	case reflect.Bool:
		return mr.getBooleanType()
	case reflect.Int, reflect.Int32:
		if mr.isPrimaryKey(field) {
			return mr.getAutoIncrementType(false)
		}
		return mr.getIntegerType()
	case reflect.Int64:
		if mr.isPrimaryKey(field) {
			return mr.getAutoIncrementType(true)
		}
		return mr.getBigIntType()
	case reflect.Float32:
		return mr.getRealType()
	case reflect.Float64:
		precision := mr.getPrecision(field)
		scale := mr.getScale(field)
		if precision != nil && scale != nil {
			return fmt.Sprintf("DECIMAL(%d,%d)", *precision, *scale)
		}
		return mr.getDoubleType()
	case reflect.String:
		if size > 0 {
			return fmt.Sprintf("VARCHAR(%d)", size)
		}
		if isTextType(field) {
			return sqlTypeText
		}
		return "VARCHAR(255)"
	default:
		if fieldType.String() == "time.Time" {
			return "TIMESTAMP"
		}
		return sqlTypeText
	}
}

func isTextType(field reflect.StructField) bool {
	migrationTag := field.Tag.Get("migration")
	if strings.Contains(migrationTag, "type:TEXT") {
		return true
	}
	sqlTag := field.Tag.Get("sql")
	return strings.Contains(sqlTag, "type:TEXT")
}

func (mr *ModelRegistry) isPrimaryKey(field reflect.StructField) bool {
	// Check db tag
	if tag := field.Tag.Get("db"); tag != "" {
		if strings.Contains(tag, "primary_key") {
			return true
		}
	}
	// Check sql tag
	if tag := field.Tag.Get("sql"); tag != "" {
		if strings.Contains(tag, "primary_key") {
			return true
		}
	}
	// Check migration tag
	if tag := field.Tag.Get("migration"); tag != "" {
		if strings.Contains(tag, "primary_key") {
			return true
		}
	}
	// Default check for ID field
	return strings.ToLower(field.Name) == "id"
}

func (mr *ModelRegistry) isUnique(field reflect.StructField) bool {
	if tag := field.Tag.Get("db"); tag != "" && strings.Contains(tag, "unique") {
		return true
	}
	if tag := field.Tag.Get("sql"); tag != "" && strings.Contains(tag, "unique") {
		return true
	}
	return false
}

func (mr *ModelRegistry) isForeignKey(field reflect.StructField) bool {
	if tag := field.Tag.Get("sql"); tag != "" {
		return strings.Contains(tag, foreignKeyTag)
	}
	// Convention: fields ending with _id are foreign keys
	return strings.HasSuffix(strings.ToLower(field.Name), "id") && strings.ToLower(field.Name) != "id"
}

func (mr *ModelRegistry) getForeignKeyInfo(field reflect.StructField) *ForeignKeyInfo {
	tag := field.Tag.Get("sql")
	if tag == "" {
		return nil
	}
	for _, part := range strings.Split(tag, ";") {
		if !strings.HasPrefix(part, foreignKeyTag) {
			continue
		}
		fk := strings.TrimPrefix(part, foreignKeyTag)
		if !strings.Contains(fk, "(") || !strings.Contains(fk, ")") {
			continue
		}
		parsed := parseForeignKey(fk)
		if parsed != nil {
			return parsed
		}
	}
	return nil
}

// Helper to parse foreign key string in format table(column)
func parseForeignKey(fk string) *ForeignKeyInfo {
	parts := strings.Split(fk, "(")
	if len(parts) != 2 {
		return nil
	}
	table := parts[0]
	column := strings.TrimSuffix(parts[1], ")")
	return &ForeignKeyInfo{Table: table, Column: column}
}

func (mr *ModelRegistry) getSize(field reflect.StructField) int {
	// Check migration tags for max_length
	if size := mr.getSizeFromMigrationTag(field); size > 0 {
		return size
	}

	// Check sql tags for size
	if size := mr.getSizeFromSQLTag(field); size > 0 {
		return size
	}

	return 0
}

// getSizeFromMigrationTag extracts max_length from migration tag
func (mr *ModelRegistry) getSizeFromMigrationTag(field reflect.StructField) int {
	tag := field.Tag.Get("migration")
	if tag == "" {
		return 0
	}

	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max_length:") {
			var size int
			if _, err := fmt.Sscanf(part, "max_length:%d", &size); err == nil {
				return size
			}
		}
	}
	return 0
}

// getSizeFromSQLTag extracts size from sql tag
func (mr *ModelRegistry) getSizeFromSQLTag(field reflect.StructField) int {
	tag := field.Tag.Get("sql")
	if tag == "" {
		return 0
	}

	for _, part := range strings.Split(tag, ";") {
		if strings.HasPrefix(part, "size:") {
			var size int
			if _, err := fmt.Sscanf(part, "size:%d", &size); err == nil {
				return size
			}
		}
	}
	return 0
}

func (mr *ModelRegistry) isAutoIncrement(field reflect.StructField) bool {
	// Check migration tags
	if tag := field.Tag.Get("migration"); tag != "" {
		return strings.Contains(tag, "auto_increment")
	}
	// Check sql tags
	if tag := field.Tag.Get("sql"); tag != "" {
		return strings.Contains(tag, "auto_increment")
	}
	// Check db tags
	if tag := field.Tag.Get("db"); tag != "" {
		return strings.Contains(tag, "auto_increment")
	}
	// Convention: primary key integer fields are auto increment
	return mr.isPrimaryKey(field) && (field.Type.Kind() == reflect.Int || field.Type.Kind() == reflect.Int64)
}

// getPrecision extracts precision from field tags
func (mr *ModelRegistry) getPrecision(field reflect.StructField) *int {
	if tag := field.Tag.Get("sql"); tag != "" {
		for _, part := range strings.Split(tag, ";") {
			if strings.HasPrefix(part, "precision:") {
				var precision int
				if _, err := fmt.Sscanf(part, "precision:%d", &precision); err == nil {
					return &precision
				}
			}
		}
	}
	return nil
}

// getScale extracts scale from field tags
func (mr *ModelRegistry) getScale(field reflect.StructField) *int {
	if tag := field.Tag.Get("sql"); tag != "" {
		for _, part := range strings.Split(tag, ";") {
			if strings.HasPrefix(part, "scale:") {
				var scale int
				if _, err := fmt.Sscanf(part, "scale:%d", &scale); err == nil {
					return &scale
				}
			}
		}
	}
	return nil
}

// calculateSnapshotChecksum calculates a checksum for the model snapshot
func (mr *ModelRegistry) calculateSnapshotChecksum(snapshot ModelSnapshot) string {
	parts := make([]string, 0, 1+len(snapshot.Columns)+len(snapshot.Indexes))

	// Add table name
	parts = append(parts, fmt.Sprintf("table:%s", snapshot.TableName))

	// Add columns in sorted order
	columnNames := make([]string, 0, len(snapshot.Columns))
	for name := range snapshot.Columns {
		columnNames = append(columnNames, name)
	}
	sort.Strings(columnNames)

	for _, name := range columnNames {
		col := snapshot.Columns[name]
		colStr := fmt.Sprintf("col:%s:%s:%t:%t:%t",
			col.Name, col.SQLType, col.Nullable, col.IsPrimaryKey, col.IsUnique)
		if col.Default != nil {
			colStr += fmt.Sprintf(":%s", *col.Default)
		}
		parts = append(parts, colStr)
	}

	// Add indexes
	indexNames := make([]string, 0, len(snapshot.Indexes))
	for name := range snapshot.Indexes {
		indexNames = append(indexNames, name)
	}
	sort.Strings(indexNames)

	for _, name := range indexNames {
		idx := snapshot.Indexes[name]
		parts = append(parts, fmt.Sprintf("idx:%s:%s:%t", idx.Name, strings.Join(idx.Columns, ","), idx.Unique))
	}

	// Calculate SHA256 hash
	data := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Database-specific type mapping methods
func (mr *ModelRegistry) getBooleanType() string {
	switch mr.driver {
	case SQLite:
		return sqlTypeInteger // SQLite uses INTEGER for boolean (0/1)
	case MySQL:
		return "TINYINT(1)"
	case PostgreSQL:
		return "BOOLEAN"
	default:
		return "BOOLEAN"
	}
}

func (mr *ModelRegistry) getAutoIncrementType(isBigInt bool) string {
	switch mr.driver {
	case SQLite:
		return sqlTypeInteger // SQLite uses INTEGER with AUTOINCREMENT
	case MySQL:
		if isBigInt {
			return sqlTypeBigInt
		}
		return "INT"
	case PostgreSQL:
		fallthrough
	default:
		if isBigInt {
			return sqlTypeBigSerial
		}
		return sqlTypeSerial
	}
}

func (mr *ModelRegistry) getIntegerType() string {
	switch mr.driver {
	case SQLite:
		return sqlTypeInteger
	case MySQL:
		return "INT"
	case PostgreSQL:
		return sqlTypeInteger
	default:
		return sqlTypeInteger
	}
}

func (mr *ModelRegistry) getBigIntType() string {
	switch mr.driver {
	case SQLite:
		return sqlTypeInteger // SQLite uses INTEGER for all integer types
	case MySQL:
		return sqlTypeBigInt
	case PostgreSQL:
		return sqlTypeBigInt
	default:
		return sqlTypeBigInt
	}
}

func (mr *ModelRegistry) getRealType() string {
	switch mr.driver {
	case SQLite:
		return sqlTypeReal
	case MySQL:
		return "FLOAT"
	case PostgreSQL:
		return sqlTypeReal
	default:
		return sqlTypeReal
	}
}

func (mr *ModelRegistry) getDoubleType() string {
	switch mr.driver {
	case SQLite:
		return "REAL" // SQLite uses REAL for all floating point
	case MySQL:
		return "DOUBLE"
	case PostgreSQL:
		return "DOUBLE PRECISION"
	default:
		return "DOUBLE PRECISION"
	}
}

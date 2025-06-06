package dbcontext

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Error messages constants
const (
	errDatabaseConnectionNil = "database connection is nil"
)

// EFContext is a simple Entity Framework Core-inspired ORM
type EFContext struct {
	db *sql.DB
}

// NewEFContext creates a new EF-style context
func NewEFContext(db *sql.DB) *EFContext {
	return &EFContext{db: db}
}

// EntityInterface represents a database entity that must have an ID field
type EntityInterface interface {
	GetID() interface{}
	SetID(interface{})
}

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	ID        uint      `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// GetID returns the ID of the BaseEntity.
func (b *BaseEntity) GetID() interface{} {
	return b.ID
}

// SetID sets the ID of the BaseEntity.
func (b *BaseEntity) SetID(id interface{}) {
	if idVal, ok := id.(uint); ok {
		b.ID = idVal
	}
}

// Add adds an entity to the context (Entity Framework style)
func (ctx *EFContext) Add(entity EntityInterface) error {
	if ctx.db == nil {
		return errors.New(errDatabaseConnectionNil)
	}
	return ctx.insert(entity)
}

// Update updates an entity in the context
func (ctx *EFContext) Update(entity EntityInterface) error {
	if ctx.db == nil {
		return errors.New(errDatabaseConnectionNil)
	}
	return ctx.update(entity)
}

// Remove removes an entity from the context
func (ctx *EFContext) Remove(entity EntityInterface) error {
	if ctx.db == nil {
		return errors.New(errDatabaseConnectionNil)
	}
	return ctx.delete(entity)
}

// Find finds an entity by ID
func (ctx *EFContext) Find(entity EntityInterface, id interface{}) error {
	if ctx.db == nil {
		return errors.New(errDatabaseConnectionNil)
	}
	return ctx.findByID(entity, id)
}

// SaveChanges commits all changes (currently no-op since we're doing immediate operations)
func (ctx *EFContext) SaveChanges() error {
	return nil
}

// ExtractFieldsForDebug extracts fields for debugging purposes
func (ctx *EFContext) ExtractFieldsForDebug(entity EntityInterface) ([]string, []interface{}) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var columns []string
	var values []interface{}
	var placeholders []string
	placeholderNum := 1

	ctx.processStructFields(v, &columns, &values, &placeholders, &placeholderNum, "insert")
	return columns, values
}

// insert inserts a new entity into the database
func (ctx *EFContext) insert(entity EntityInterface) error {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	tableName := ctx.getTableNameFromType(v.Type())

	// Extract fields for insert (excluding ID)
	columns, values, placeholders := ctx.extractFieldsForInsert(v)

	if len(columns) == 0 {
		return errors.New("no fields to insert")
	}

	// Set timestamps
	ctx.setTimestamps(v, true)

	// #nosec G201 -- Table and columns are controlled by ORM, not user input
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	var id interface{}
	err := ctx.db.QueryRow(query, values...).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	entity.SetID(id)
	return nil
}

// processStructFields recursively processes struct fields for INSERT/UPDATE
func (ctx *EFContext) processStructFields(v reflect.Value, columns *[]string, values *[]interface{}, placeholders *[]string, placeholderNum *int, operation string) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !ctx.shouldProcessField(field, fieldValue, operation) {
			continue
		}

		ctx.addFieldToQuery(field, fieldValue, columns, values, placeholders, placeholderNum)
	}
}

// shouldProcessField determines if a field should be processed based on operation and field properties
func (ctx *EFContext) shouldProcessField(field reflect.StructField, fieldValue reflect.Value, operation string) bool {
	// Skip unexported fields
	if !fieldValue.CanInterface() {
		return false
	}

	// Handle embedded structs (like BaseEntity) - these are processed recursively
	if field.Anonymous && field.Type.Kind() == reflect.Struct {
		return true
	}

	// Skip ID field for inserts and updates
	if ctx.isIDField(field) {
		return false
	}

	// Skip fields marked to be ignored
	if ctx.shouldSkipField(field) {
		return false
	}

	return true
}

// isIDField checks if a field is an ID field
func (ctx *EFContext) isIDField(field reflect.StructField) bool {
	return field.Name == "ID" || strings.ToLower(field.Name) == "id"
}

// addFieldToQuery adds field data to the query building arrays
func (ctx *EFContext) addFieldToQuery(field reflect.StructField, fieldValue reflect.Value, columns *[]string, values *[]interface{}, placeholders *[]string, placeholderNum *int) {
	// Handle embedded structs recursively
	if field.Anonymous && field.Type.Kind() == reflect.Struct {
		ctx.processStructFields(fieldValue, columns, values, placeholders, placeholderNum, "insert")
		return
	}

	// Get column name and add to query
	columnName := ctx.getColumnNameFromField(field)
	*columns = append(*columns, columnName)
	*values = append(*values, fieldValue.Interface())
	*placeholders = append(*placeholders, "$"+strconv.Itoa(*placeholderNum))
	*placeholderNum++
}

// extractFieldsForInsert extracts fields for INSERT (excludes ID, includes timestamps)
func (ctx *EFContext) extractFieldsForInsert(v reflect.Value) ([]string, []interface{}, []string) {
	var columns []string
	var values []interface{}
	var placeholders []string
	placeholderNum := 1

	ctx.processStructFields(v, &columns, &values, &placeholders, &placeholderNum, "insert")
	return columns, values, placeholders
}

// setTimestamps sets created_at and updated_at timestamps
func (ctx *EFContext) setTimestamps(v reflect.Value, isInsert bool) {
	now := time.Now()

	// Set CreatedAt for inserts
	if isInsert {
		if createdField := v.FieldByName("CreatedAt"); createdField.IsValid() && createdField.CanSet() {
			createdField.Set(reflect.ValueOf(now))
		}
	}

	// Always set UpdatedAt
	if updatedField := v.FieldByName("UpdatedAt"); updatedField.IsValid() && updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(now))
	}
}

// getTableNameFromType gets the table name from struct type
func (ctx *EFContext) getTableNameFromType(t reflect.Type) string {
	name := t.Name()
	// Convert to snake_case and pluralize
	return ctx.toSnakeCaseEF(name) + "s"
}

// getColumnNameFromField gets the column name from struct field
func (ctx *EFContext) getColumnNameFromField(field reflect.StructField) string {
	// Check for db tag first
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		return dbTag
	}

	// Check for json tag
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		return jsonTag
	}

	// Convert field name to snake_case
	return ctx.toSnakeCaseEF(field.Name)
}

// shouldSkipField determines if a field should be skipped
func (ctx *EFContext) shouldSkipField(field reflect.StructField) bool {
	// Skip fields with db:"-" tag
	if dbTag := field.Tag.Get("db"); dbTag == "-" {
		return true
	}

	// Skip fields with json:"-" tag
	if jsonTag := field.Tag.Get("json"); jsonTag == "-" {
		return true
	}

	return false
}

// toSnakeCaseEF converts camelCase to snake_case
func (ctx *EFContext) toSnakeCaseEF(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// Helper methods for other operations (simplified for now)
func (ctx *EFContext) update(_ EntityInterface) error {
	return errors.New("update not yet implemented")
}

func (ctx *EFContext) delete(_ EntityInterface) error {
	return errors.New("delete not yet implemented")
}

func (ctx *EFContext) findByID(_ EntityInterface, _ interface{}) error {
	return errors.New("findByID not yet implemented")
}

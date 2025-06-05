package dbcontext

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// detectDatabaseDriver detects the database driver type
func detectDatabaseDriver(db *sql.DB) string {
	// Test queries to detect database type
	if _, err := db.Query("SELECT 1::integer"); err == nil {
		return "postgres"
	}
	if _, err := db.Query("SELECT sqlite_version()"); err == nil {
		return "sqlite3"
	}
	if _, err := db.Query("SELECT VERSION()"); err == nil {
		return "mysql"
	}
	// Default to sqlite3 if detection fails
	return "sqlite3"
}

// convertQueryPlaceholders converts query placeholders based on database driver
func convertQueryPlaceholders(query string, driver string) string {
	if driver != "postgres" {
		return query // SQLite and MySQL use ? placeholders
	}

	// Convert ? placeholders to $1, $2, $3 for PostgreSQL
	count := 0
	result := ""
	for _, char := range query {
		if char == '?' {
			count++
			result += fmt.Sprintf("$%d", count)
		} else {
			result += string(char)
		}
	}
	return result
}

// EntityState represents the state of an entity in the change tracker
type EntityState int

const (
	EntityStateUnchanged EntityState = iota
	EntityStateAdded
	EntityStateModified
	EntityStateDeleted
)

// String returns the string representation of EntityState
func (s EntityState) String() string {
	switch s {
	case EntityStateUnchanged:
		return "Unchanged"
	case EntityStateAdded:
		return "Added"
	case EntityStateModified:
		return "Modified"
	case EntityStateDeleted:
		return "Deleted"
	default:
		return "Unknown"
	}
}

// ChangeTracker manages entity states and changes
type ChangeTracker struct {
	entities map[interface{}]EntityState
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		entities: make(map[interface{}]EntityState),
	}
}

// GetEntityState returns the current state of an entity
func (ct *ChangeTracker) GetEntityState(entity interface{}) EntityState {
	if state, exists := ct.entities[entity]; exists {
		return state
	}
	return EntityStateUnchanged
}

// SetEntityState sets the state of an entity
func (ct *ChangeTracker) SetEntityState(entity interface{}, state EntityState) {
	ct.entities[entity] = state
}

// TrackEntity adds an entity to tracking with specified state
func (ct *ChangeTracker) TrackEntity(entity interface{}, state EntityState) {
	ct.entities[entity] = state
}

// Database provides transaction support
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new Database instance
func NewDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

// Begin starts a new transaction
func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// EnhancedDbContext provides Entity Framework Core-like functionality
type EnhancedDbContext struct {
	db            *sql.DB
	tx            *sql.Tx
	ChangeTracker *ChangeTracker
	Database      *Database
	driver        string
}

// NewEnhancedDbContext creates a new enhanced database context
func NewEnhancedDbContext(connectionString string) (*EnhancedDbContext, error) {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	driver := detectDatabaseDriver(db)

	return &EnhancedDbContext{
		db:            db,
		ChangeTracker: NewChangeTracker(),
		Database:      NewDatabase(db),
		driver:        driver,
	}, nil
}

// NewEnhancedDbContextWithDB creates a new enhanced database context with existing DB
func NewEnhancedDbContextWithDB(db *sql.DB) *EnhancedDbContext {
	driver := detectDatabaseDriver(db)
	return &EnhancedDbContext{
		db:            db,
		ChangeTracker: NewChangeTracker(),
		Database:      NewDatabase(db),
		driver:        driver,
	}
}

// NewEnhancedDbContextWithTx creates a new enhanced database context with transaction
func NewEnhancedDbContextWithTx(tx *sql.Tx) *EnhancedDbContext {
	// Note: for transactions, we can't easily detect the driver type
	// so we default to sqlite3. In practice, this constructor is used
	// within an existing context that already has the driver detected.
	return &EnhancedDbContext{
		tx:            tx,
		ChangeTracker: NewChangeTracker(),
		driver:        "sqlite3", // default, should be set by parent context
	}
}

// Add marks an entity for insertion
func (ctx *EnhancedDbContext) Add(entity interface{}) {
	ctx.ChangeTracker.SetEntityState(entity, EntityStateAdded)
}

// Update marks an entity for update
func (ctx *EnhancedDbContext) Update(entity interface{}) {
	ctx.ChangeTracker.SetEntityState(entity, EntityStateModified)
}

// Delete marks an entity for deletion
func (ctx *EnhancedDbContext) Delete(entity interface{}) {
	ctx.ChangeTracker.SetEntityState(entity, EntityStateDeleted)
}

// SaveChanges persists all pending changes to the database
func (ctx *EnhancedDbContext) SaveChanges() (int, error) {
	affected := 0

	for entity, state := range ctx.ChangeTracker.entities {
		switch state {
		case EntityStateAdded:
			err := ctx.insertEntity(entity)
			if err != nil {
				return affected, err
			}
			ctx.ChangeTracker.SetEntityState(entity, EntityStateUnchanged)
			affected++

		case EntityStateModified:
			err := ctx.updateEntity(entity)
			if err != nil {
				return affected, err
			}
			ctx.ChangeTracker.SetEntityState(entity, EntityStateUnchanged)
			affected++

		case EntityStateDeleted:
			err := ctx.deleteEntity(entity)
			if err != nil {
				return affected, err
			}
			delete(ctx.ChangeTracker.entities, entity)
			affected++
		}
	}

	return affected, nil
}

// insertEntity inserts a new entity into the database
func (ctx *EnhancedDbContext) insertEntity(entity interface{}) error {
	// Set timestamps before inserting
	setTimestamps(entity, true) // true = create timestamps

	tableName := getTableName(entity)
	columns, values, placeholders := getInsertData(entity, ctx.driver)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	var err error
	var result sql.Result

	if ctx.tx != nil {
		result, err = ctx.tx.Exec(query, values...)
	} else {
		result, err = ctx.db.Exec(query, values...)
	}

	if err != nil {
		return err
	}

	// Set the ID if it's an auto-increment field
	if id, err := result.LastInsertId(); err == nil && id > 0 {
		setIDField(entity, id)
	}

	return nil
}

// updateEntity updates an existing entity in the database
func (ctx *EnhancedDbContext) updateEntity(entity interface{}) error {
	// Set UpdatedAt timestamp before updating
	setTimestamps(entity, false) // false = update timestamp only

	tableName := getTableName(entity)
	setPairs, values, idValue := getUpdateData(entity, ctx.driver)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?",
		tableName, strings.Join(setPairs, ", "))

	// Convert placeholders for PostgreSQL
	query = convertQueryPlaceholders(query, ctx.driver)

	values = append(values, idValue)

	if ctx.tx != nil {
		_, err := ctx.tx.Exec(query, values...)
		return err
	} else {
		_, err := ctx.db.Exec(query, values...)
		return err
	}
}

// deleteEntity removes an entity from the database
func (ctx *EnhancedDbContext) deleteEntity(entity interface{}) error {
	tableName := getTableName(entity)
	idValue := getIDValue(entity)

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	// Convert placeholders for PostgreSQL
	query = convertQueryPlaceholders(query, ctx.driver)

	// Debug output
	fmt.Printf("DEBUG DELETE: tableName=%s, idValue=%v, query=%s\n", tableName, idValue, query)

	if ctx.tx != nil {
		result, err := ctx.tx.Exec(query, idValue)
		if err == nil {
			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("DEBUG DELETE TX: rowsAffected=%d\n", rowsAffected)
		}
		return err
	} else {
		result, err := ctx.db.Exec(query, idValue)
		if err == nil {
			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("DEBUG DELETE DB: rowsAffected=%d\n", rowsAffected)
		}
		return err
	}
}

// EnhancedDbSet provides LINQ-style querying capabilities
type EnhancedDbSet[T any] struct {
	ctx         *EnhancedDbContext
	tableName   string
	whereClause string
	whereArgs   []interface{}
	orderClause string
	limitValue  int
	offsetValue int
	noTracking  bool
}

// NewEnhancedDbSet creates a new enhanced database set
func NewEnhancedDbSet[T any](ctx *EnhancedDbContext) *EnhancedDbSet[T] {
	var entity T
	tableName := getTableName(&entity)
	return &EnhancedDbSet[T]{
		ctx:       ctx,
		tableName: tableName,
	}
}

// Where adds a WHERE clause to the query
func (set *EnhancedDbSet[T]) Where(condition string, args ...interface{}) *EnhancedDbSet[T] {
	newSet := *set

	// Convert placeholders for PostgreSQL
	condition = set.adjustPlaceholdersForCondition(condition)

	if newSet.whereClause != "" {
		newSet.whereClause += " AND " + condition
	} else {
		newSet.whereClause = condition
	}
	newSet.whereArgs = append(newSet.whereArgs, args...)
	return &newSet
}

// adjustPlaceholdersForCondition converts ? placeholders to appropriate format
func (set *EnhancedDbSet[T]) adjustPlaceholdersForCondition(condition string) string {
	if set.ctx.driver != "postgres" {
		return condition
	}

	// Convert ? to $N starting from the next available position
	count := len(set.whereArgs)
	result := ""
	for _, char := range condition {
		if char == '?' {
			count++
			result += fmt.Sprintf("$%d", count)
		} else {
			result += string(char)
		}
	}
	return result
}

// WhereLike adds a WHERE LIKE clause to the query
func (set *EnhancedDbSet[T]) WhereLike(column string, pattern string) *EnhancedDbSet[T] {
	return set.Where(column+" LIKE ?", pattern)
}

// WhereIn adds a WHERE IN clause to the query
func (set *EnhancedDbSet[T]) WhereIn(column string, values []interface{}) *EnhancedDbSet[T] {
	if len(values) == 0 {
		return set
	}

	newSet := *set
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	condition = newSet.adjustPlaceholdersForCondition(condition)

	if newSet.whereClause != "" {
		newSet.whereClause += " AND " + condition
	} else {
		newSet.whereClause = condition
	}
	newSet.whereArgs = append(newSet.whereArgs, values...)
	return &newSet
}

// WhereOr adds an OR WHERE clause to the query
func (set *EnhancedDbSet[T]) WhereOr(condition string, args ...interface{}) *EnhancedDbSet[T] {
	newSet := *set
	if newSet.whereClause != "" {
		newSet.whereClause += " OR (" + condition + ")"
	} else {
		newSet.whereClause = condition
	}
	newSet.whereArgs = append(newSet.whereArgs, args...)
	return &newSet
}

// OrderBy adds an ORDER BY clause to the query
func (set *EnhancedDbSet[T]) OrderBy(column string) *EnhancedDbSet[T] {
	newSet := *set
	newSet.orderClause = column
	return &newSet
}

// OrderByDescending adds an ORDER BY DESC clause to the query
func (set *EnhancedDbSet[T]) OrderByDescending(column string) *EnhancedDbSet[T] {
	newSet := *set
	newSet.orderClause = column + " DESC"
	return &newSet
}

// Take limits the number of results
func (set *EnhancedDbSet[T]) Take(count int) *EnhancedDbSet[T] {
	newSet := *set
	newSet.limitValue = count
	return &newSet
}

// Skip skips a number of results
func (set *EnhancedDbSet[T]) Skip(count int) *EnhancedDbSet[T] {
	newSet := *set
	newSet.offsetValue = count
	return &newSet
}

// AsNoTracking disables change tracking for the query
func (set *EnhancedDbSet[T]) AsNoTracking() *EnhancedDbSet[T] {
	newSet := *set
	newSet.noTracking = true
	return &newSet
}

// ToList executes the query and returns all results
func (set *EnhancedDbSet[T]) ToList() ([]*T, error) {
	query := set.buildQuery()

	var rows *sql.Rows
	var err error

	if set.ctx.tx != nil {
		rows, err = set.ctx.tx.Query(query, set.whereArgs...)
	} else {
		rows, err = set.ctx.db.Query(query, set.whereArgs...)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Note: this is logged but doesn't affect the return value since we're in a defer
			log.Printf("Warning: Failed to close rows: %v", closeErr)
		}
	}()

	var results []*T
	for rows.Next() {
		entity := new(T)
		err := scanEntity(rows, entity)
		if err != nil {
			return nil, err
		}

		if !set.noTracking {
			set.ctx.ChangeTracker.TrackEntity(entity, EntityStateUnchanged)
		}

		results = append(results, entity)
	}

	return results, rows.Err()
}

// FirstOrDefault returns the first result or nil if none found
func (set *EnhancedDbSet[T]) FirstOrDefault() (*T, error) {
	results, err := set.Take(1).ToList()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Count returns the number of entities matching the query
func (set *EnhancedDbSet[T]) Count() (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", set.tableName)
	if set.whereClause != "" {
		query += " WHERE " + set.whereClause
	}

	var count int
	var err error

	if set.ctx.tx != nil {
		err = set.ctx.tx.QueryRow(query, set.whereArgs...).Scan(&count)
	} else {
		err = set.ctx.db.QueryRow(query, set.whereArgs...).Scan(&count)
	}

	return count, err
}

// Any checks if any records match the query
func (set *EnhancedDbSet[T]) Any() (bool, error) {
	count, err := set.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Find finds an entity by its primary key
func (set *EnhancedDbSet[T]) Find(id interface{}) (*T, error) {
	return set.Where("id = ?", id).FirstOrDefault()
}

// First returns the first result (errors if no results)
func (set *EnhancedDbSet[T]) First() (*T, error) {
	results, err := set.Take(1).ToList()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	return results[0], nil
}

// Single returns a single result (errors if 0 or >1 results)
func (set *EnhancedDbSet[T]) Single() (*T, error) {
	results, err := set.Take(2).ToList()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("multiple results found, expected single result")
	}
	return results[0], nil
}

// buildQuery constructs the SQL query string
func (set *EnhancedDbSet[T]) buildQuery() string {
	query := fmt.Sprintf("SELECT * FROM %s", set.tableName)

	if set.whereClause != "" {
		query += " WHERE " + set.whereClause
	}

	if set.orderClause != "" {
		query += " ORDER BY " + set.orderClause
	}

	if set.limitValue > 0 {
		query += fmt.Sprintf(" LIMIT %d", set.limitValue)
	}

	if set.offsetValue > 0 {
		query += fmt.Sprintf(" OFFSET %d", set.offsetValue)
	}

	return query
}

// Helper functions

// getTableName extracts table name from entity type
func getTableName(entity interface{}) string {
	// Check if entity has TableName method
	if tn, ok := entity.(interface{ TableName() string }); ok {
		return tn.TableName()
	}

	// Fall back to struct name converted to snake_case
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return toSnakeCase(t.Name())
}

// getInsertData extracts columns, values, and placeholders for INSERT
func getInsertData(entity interface{}, driver string) ([]string, []interface{}, []string) {
	return getFieldData(entity, true, driver) // true = exclude ID for INSERT
}

// getFieldData extracts field data recursively, handling embedded structs
func getFieldData(entity interface{}, excludeID bool, driver string) ([]string, []interface{}, []string) {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	var columns []string
	var values []interface{}
	var placeholders []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Create a pointer to the embedded struct value
			embeddedPtr := reflect.New(field.Type)
			embeddedPtr.Elem().Set(value)

			// Recursively get fields from embedded struct
			embeddedCols, embeddedVals, embeddedPlaceholders := getFieldData(embeddedPtr.Interface(), excludeID, driver)
			columns = append(columns, embeddedCols...)
			values = append(values, embeddedVals...)
			placeholders = append(placeholders, embeddedPlaceholders...)
			continue
		}

		// Skip ID field for auto-increment if requested
		if excludeID && strings.ToLower(field.Name) == "id" {
			continue
		}

		// Skip fields with sql:"-" tag or db:"-" tag
		if dbTag := field.Tag.Get("db"); dbTag == "-" {
			continue
		}
		if sqlTag := field.Tag.Get("sql"); sqlTag == "-" {
			continue
		}

		// Get column name from db tag or convert field name
		columnName := field.Tag.Get("db")
		if columnName == "" {
			columnName = toSnakeCase(field.Name)
		}

		columns = append(columns, columnName)
		values = append(values, value.Interface())

		// Generate appropriate placeholder based on driver
		if driver == "postgres" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(placeholders)+1))
		} else {
			placeholders = append(placeholders, "?")
		}
	}

	return columns, values, placeholders
}

// getUpdateData extracts SET clauses and values for UPDATE
func getUpdateData(entity interface{}, driver string) ([]string, []interface{}, interface{}) {
	columns, values, _ := getFieldData(entity, false, driver) // false = include all fields

	var setPairs []string
	var updateValues []interface{}
	var idValue interface{}

	for i, col := range columns {
		if strings.ToLower(col) == "id" {
			idValue = values[i]
			continue
		}
		if driver == "postgres" {
			setPairs = append(setPairs, fmt.Sprintf("%s = $%d", col, len(updateValues)+1))
		} else {
			setPairs = append(setPairs, col+" = ?")
		}
		updateValues = append(updateValues, values[i])
	}

	return setPairs, updateValues, idValue
}

// getIDValue extracts the ID value from an entity, including embedded structs
func getIDValue(entity interface{}) interface{} {
	return findFieldValue(entity, "ID")
}

// setIDField sets the ID field of an entity, including embedded structs
func setIDField(entity interface{}, id int64) {
	setEntityIDValue(entity, "ID", id)
}

// findFieldValue recursively finds a field value in struct and embedded structs
func findFieldValue(entity interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Check if this is the field we're looking for
		if field.Name == fieldName {
			return value.Interface()
		}

		// Check embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedPtr := reflect.New(field.Type)
			embeddedPtr.Elem().Set(value)
			if result := findFieldValue(embeddedPtr.Interface(), fieldName); result != nil {
				return result
			}
		}
	}
	return nil
}

// setEntityIDValue recursively sets a field value in struct and embedded structs
func setEntityIDValue(entity interface{}, fieldName string, value int64) {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Check if this is the field we're looking for
		if field.Name == fieldName && fieldValue.CanSet() {
			switch fieldValue.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				fieldValue.SetInt(value)
			case reflect.Uint, reflect.Uint32, reflect.Uint64:
				fieldValue.SetUint(uint64(value))
			}
			return
		}

		// Check embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct && fieldValue.CanSet() {
			embeddedPtr := reflect.New(field.Type)
			embeddedPtr.Elem().Set(fieldValue)
			setEntityIDValue(embeddedPtr.Interface(), fieldName, value)
			fieldValue.Set(embeddedPtr.Elem())
		}
	}
}

// setTimestamps sets CreatedAt and UpdatedAt timestamps on an entity
func setTimestamps(entity interface{}, isCreate bool) {
	now := time.Now()

	if isCreate {
		setTimestampField(entity, "CreatedAt", now)
	}
	setTimestampField(entity, "UpdatedAt", now)
}

// setTimestampField recursively sets a timestamp field in struct and embedded structs
func setTimestampField(entity interface{}, fieldName string, value time.Time) {
	v := reflect.ValueOf(entity).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Check if this is the field we're looking for
		if field.Name == fieldName && fieldValue.CanSet() {
			if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
				fieldValue.Set(reflect.ValueOf(value))
			}
			return
		}

		// Check embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct && fieldValue.CanSet() {
			embeddedPtr := reflect.New(field.Type)
			embeddedPtr.Elem().Set(fieldValue)
			setTimestampField(embeddedPtr.Interface(), fieldName, value)
			fieldValue.Set(embeddedPtr.Elem())
		}
	}
}

// scanEntity scans database row into entity
func scanEntity(rows *sql.Rows, entity interface{}) error {
	v := reflect.ValueOf(entity).Elem()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Create slice of interface{} to hold column values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return err
	}

	// Map columns to struct fields
	for i, column := range columns {
		fieldName := toCamelCase(column)
		field := v.FieldByName(fieldName)

		if !field.IsValid() || !field.CanSet() {
			continue
		}

		value := values[i]
		if value == nil {
			continue
		}

		err := setFieldValue(field, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// setFieldValue sets a field value with type conversion
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else if bytes, ok := value.([]byte); ok {
			field.SetString(string(bytes))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := value.(int64); ok {
			field.SetInt(num)
		} else if str, ok := value.(string); ok {
			if num, err := strconv.ParseInt(str, 10, 64); err == nil {
				field.SetInt(num)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := value.(int64); ok && num >= 0 {
			field.SetUint(uint64(num))
		} else if str, ok := value.(string); ok {
			if num, err := strconv.ParseUint(str, 10, 64); err == nil {
				field.SetUint(num)
			}
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := value.(float64); ok {
			field.SetFloat(num)
		} else if str, ok := value.(string); ok {
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				field.SetFloat(num)
			}
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		} else if num, ok := value.(int64); ok {
			field.SetBool(num != 0)
		}
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			if str, ok := value.(string); ok {
				if t, err := time.Parse("2006-01-02 15:04:05", str); err == nil {
					field.Set(reflect.ValueOf(t))
				}
			}
		}
	}

	return nil
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(str string) string {
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

// toCamelCase converts snake_case to CamelCase
func toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

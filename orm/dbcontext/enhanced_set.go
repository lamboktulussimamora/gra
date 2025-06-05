// Package dbcontext provides LINQ-style query operations for entities
package dbcontext

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// WhereClause represents a WHERE condition
type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
	Logic    string // AND, OR
}

// OrderClause represents an ORDER BY condition
type OrderClause struct {
	Column string
	Desc   bool
}

// JoinClause represents a JOIN operation
type JoinClause struct {
	Type       string // INNER, LEFT, RIGHT, FULL
	Table      string
	Condition  string
	TableAlias string
}

// QueryBuilder provides LINQ-style query building
type QueryBuilder struct {
	ctx          *EnhancedDbContext
	tableName    string
	entityType   reflect.Type
	whereClauses []WhereClause
	orderClauses []OrderClause
	joinClauses  []JoinClause
	selectFields []string
	limit        int
	offset       int
	distinct     bool
	groupBy      []string
	having       []WhereClause
}

// EnhancedSet provides LINQ-style operations for a specific entity type
type EnhancedSet[T any] struct {
	builder *QueryBuilder
}

// NewEnhancedSet creates a new enhanced set for the given entity type
func NewEnhancedSet[T any](ctx *EnhancedDbContext) *EnhancedSet[T] {
	var entity T
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	tableName := getTableNameFromType(entityType)

	builder := &QueryBuilder{
		ctx:        ctx,
		tableName:  tableName,
		entityType: entityType,
		limit:      -1,
		offset:     -1,
	}

	return &EnhancedSet[T]{
		builder: builder,
	}
}

// Where adds a WHERE clause to the query
func (es *EnhancedSet[T]) Where(column string, operator string, value interface{}) *EnhancedSet[T] {
	es.builder.whereClauses = append(es.builder.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		Logic:    "AND",
	})
	return es
}

// WhereOr adds an OR WHERE clause to the query
func (es *EnhancedSet[T]) WhereOr(column string, operator string, value interface{}) *EnhancedSet[T] {
	es.builder.whereClauses = append(es.builder.whereClauses, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		Logic:    "OR",
	})
	return es
}

// WhereIn adds a WHERE IN clause to the query
func (es *EnhancedSet[T]) WhereIn(column string, values []interface{}) *EnhancedSet[T] {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}

	es.builder.whereClauses = append(es.builder.whereClauses, WhereClause{
		Column:   column,
		Operator: "IN (" + strings.Join(placeholders, ", ") + ")",
		Value:    values,
		Logic:    "AND",
	})
	return es
}

// WhereLike adds a WHERE LIKE clause to the query
func (es *EnhancedSet[T]) WhereLike(column string, pattern string) *EnhancedSet[T] {
	return es.Where(column, "LIKE", pattern)
}

// WhereNull adds a WHERE IS NULL clause to the query
func (es *EnhancedSet[T]) WhereNull(column string) *EnhancedSet[T] {
	es.builder.whereClauses = append(es.builder.whereClauses, WhereClause{
		Column:   column,
		Operator: "IS NULL",
		Value:    nil,
		Logic:    "AND",
	})
	return es
}

// WhereNotNull adds a WHERE IS NOT NULL clause to the query
func (es *EnhancedSet[T]) WhereNotNull(column string) *EnhancedSet[T] {
	es.builder.whereClauses = append(es.builder.whereClauses, WhereClause{
		Column:   column,
		Operator: "IS NOT NULL",
		Value:    nil,
		Logic:    "AND",
	})
	return es
}

// OrderBy adds an ORDER BY clause to the query
func (es *EnhancedSet[T]) OrderBy(column string) *EnhancedSet[T] {
	es.builder.orderClauses = append(es.builder.orderClauses, OrderClause{
		Column: column,
		Desc:   false,
	})
	return es
}

// OrderByDesc adds an ORDER BY DESC clause to the query
func (es *EnhancedSet[T]) OrderByDesc(column string) *EnhancedSet[T] {
	es.builder.orderClauses = append(es.builder.orderClauses, OrderClause{
		Column: column,
		Desc:   true,
	})
	return es
}

// Take limits the number of results
func (es *EnhancedSet[T]) Take(count int) *EnhancedSet[T] {
	es.builder.limit = count
	return es
}

// Skip skips the specified number of results
func (es *EnhancedSet[T]) Skip(count int) *EnhancedSet[T] {
	es.builder.offset = count
	return es
}

// Select specifies which fields to select
func (es *EnhancedSet[T]) Select(fields ...string) *EnhancedSet[T] {
	es.builder.selectFields = fields
	return es
}

// Distinct adds DISTINCT to the query
func (es *EnhancedSet[T]) Distinct() *EnhancedSet[T] {
	es.builder.distinct = true
	return es
}

// GroupBy adds GROUP BY clause to the query
func (es *EnhancedSet[T]) GroupBy(columns ...string) *EnhancedSet[T] {
	es.builder.groupBy = columns
	return es
}

// Having adds HAVING clause to the query
func (es *EnhancedSet[T]) Having(column string, operator string, value interface{}) *EnhancedSet[T] {
	es.builder.having = append(es.builder.having, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
		Logic:    "AND",
	})
	return es
}

// InnerJoin adds an INNER JOIN clause
func (es *EnhancedSet[T]) InnerJoin(table string, condition string) *EnhancedSet[T] {
	es.builder.joinClauses = append(es.builder.joinClauses, JoinClause{
		Type:      "INNER",
		Table:     table,
		Condition: condition,
	})
	return es
}

// LeftJoin adds a LEFT JOIN clause
func (es *EnhancedSet[T]) LeftJoin(table string, condition string) *EnhancedSet[T] {
	es.builder.joinClauses = append(es.builder.joinClauses, JoinClause{
		Type:      "LEFT",
		Table:     table,
		Condition: condition,
	})
	return es
}

// RightJoin adds a RIGHT JOIN clause
func (es *EnhancedSet[T]) RightJoin(table string, condition string) *EnhancedSet[T] {
	es.builder.joinClauses = append(es.builder.joinClauses, JoinClause{
		Type:      "RIGHT",
		Table:     table,
		Condition: condition,
	})
	return es
}

// ToList executes the query and returns all results
func (es *EnhancedSet[T]) ToList() ([]T, error) {
	query, args := es.builder.buildSelectQuery()

	var db *sql.DB
	if es.builder.ctx.tx != nil {
		// Use transaction if available
		rows, err := es.builder.ctx.tx.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log but don't affect return value
				fmt.Printf("Warning: Failed to close rows: %v\n", closeErr)
			}
		}()

		return es.scanRows(rows)
	} else {
		// Use regular database connection
		db = es.builder.ctx.Database.db
		rows, err := db.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				// Log but don't affect return value
				fmt.Printf("Warning: Failed to close rows: %v\n", closeErr)
			}
		}()

		return es.scanRows(rows)
	}
}

// First executes the query and returns the first result
func (es *EnhancedSet[T]) First() (T, error) {
	es.builder.limit = 1
	results, err := es.ToList()

	var zero T
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("no results found")
	}

	return results[0], nil
}

// FirstOrDefault executes the query and returns the first result or default value
func (es *EnhancedSet[T]) FirstOrDefault() (T, error) {
	es.builder.limit = 1
	results, err := es.ToList()

	var zero T
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, nil
	}

	return results[0], nil
}

// Single executes the query and returns a single result (errors if 0 or >1 results)
func (es *EnhancedSet[T]) Single() (T, error) {
	results, err := es.ToList()

	var zero T
	if err != nil {
		return zero, err
	}

	if len(results) == 0 {
		return zero, fmt.Errorf("no results found")
	}

	if len(results) > 1 {
		return zero, fmt.Errorf("multiple results found, expected single result")
	}

	return results[0], nil
}

// Count returns the count of records matching the query
func (es *EnhancedSet[T]) Count() (int64, error) {
	// Create a copy of the builder for count query
	countBuilder := &QueryBuilder{
		ctx:          es.builder.ctx,
		tableName:    es.builder.tableName,
		entityType:   es.builder.entityType,
		whereClauses: es.builder.whereClauses,
		joinClauses:  es.builder.joinClauses,
		groupBy:      es.builder.groupBy,
		having:       es.builder.having,
		selectFields: []string{"COUNT(*)"},
	}

	query, args := countBuilder.buildSelectQuery()

	var count int64
	var err error

	if es.builder.ctx.tx != nil {
		err = es.builder.ctx.tx.QueryRow(query, args...).Scan(&count)
	} else {
		err = es.builder.ctx.Database.db.QueryRow(query, args...).Scan(&count)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Any returns true if any records match the query
func (es *EnhancedSet[T]) Any() (bool, error) {
	count, err := es.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Find finds an entity by its primary key
func (es *EnhancedSet[T]) Find(id interface{}) (T, error) {
	return es.Where("id", "=", id).First()
}

// scanRows scans database rows into entities
func (es *EnhancedSet[T]) scanRows(rows *sql.Rows) ([]T, error) {
	var results []T

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	for rows.Next() {
		entity := reflect.New(es.builder.entityType).Interface()
		valuePtrs := make([]interface{}, len(columns))

		// Map columns to struct fields
		entityVal := reflect.ValueOf(entity).Elem()
		for i, col := range columns {
			field := es.findFieldByDbTag(entityVal, col)
			if field.IsValid() && field.CanSet() {
				valuePtrs[i] = field.Addr().Interface()
			} else {
				var temp interface{}
				valuePtrs[i] = &temp
			}
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to T type
		if convertedEntity, ok := entity.(T); ok {
			results = append(results, convertedEntity)
		} else {
			// Handle pointer types
			if entityPtr := reflect.ValueOf(entity); entityPtr.Kind() == reflect.Ptr {
				if convertedEntity, ok := entityPtr.Elem().Interface().(T); ok {
					results = append(results, convertedEntity)
				}
			}
		}
	}

	return results, nil
}

// findFieldByDbTag finds a struct field by its db tag
func (es *EnhancedSet[T]) findFieldByDbTag(val reflect.Value, dbTag string) reflect.Value {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("db") == dbTag {
			return val.Field(i)
		}
	}
	return reflect.Value{}
}

// buildSelectQuery builds the complete SELECT query
func (qb *QueryBuilder) buildSelectQuery() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// SELECT clause
	query.WriteString("SELECT ")
	if qb.distinct {
		query.WriteString("DISTINCT ")
	}

	if len(qb.selectFields) > 0 {
		query.WriteString(strings.Join(qb.selectFields, ", "))
	} else {
		query.WriteString("*")
	}

	// FROM clause
	query.WriteString(" FROM ")
	query.WriteString(qb.tableName)

	// JOIN clauses
	for _, join := range qb.joinClauses {
		query.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition))
	}

	// WHERE clause
	if len(qb.whereClauses) > 0 {
		query.WriteString(" WHERE ")
		for i, where := range qb.whereClauses {
			if i > 0 {
				query.WriteString(" ")
				query.WriteString(where.Logic)
				query.WriteString(" ")
			}

			query.WriteString(where.Column)
			query.WriteString(" ")
			query.WriteString(where.Operator)

			if where.Value != nil {
				if where.Operator == "IN" || strings.Contains(where.Operator, "IN (") {
					// Handle IN clause with multiple values
					if values, ok := where.Value.([]interface{}); ok {
						args = append(args, values...)
					}
				} else {
					query.WriteString(" ?")
					args = append(args, where.Value)
				}
			}
		}
	}

	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY ")
		query.WriteString(strings.Join(qb.groupBy, ", "))
	}

	// HAVING clause
	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		for i, having := range qb.having {
			if i > 0 {
				query.WriteString(" ")
				query.WriteString(having.Logic)
				query.WriteString(" ")
			}

			query.WriteString(having.Column)
			query.WriteString(" ")
			query.WriteString(having.Operator)

			if having.Value != nil {
				query.WriteString(" ?")
				args = append(args, having.Value)
			}
		}
	}

	// ORDER BY clause
	if len(qb.orderClauses) > 0 {
		query.WriteString(" ORDER BY ")
		var orderParts []string
		for _, order := range qb.orderClauses {
			orderPart := order.Column
			if order.Desc {
				orderPart += " DESC"
			}
			orderParts = append(orderParts, orderPart)
		}
		query.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT clause
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET clause
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), args
}

// getTableNameFromType gets the table name from a reflect.Type
func getTableNameFromType(entityType reflect.Type) string {
	// Check if type has TableName method
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// Try to create an instance and check for TableName method
	if entityType.Kind() == reflect.Struct {
		instance := reflect.New(entityType).Interface()
		if tn, ok := instance.(interface{ TableName() string }); ok {
			return tn.TableName()
		}
	}

	// Default to struct name in lowercase with 's' suffix
	typeName := entityType.Name()
	return strings.ToLower(typeName) + "s"
}

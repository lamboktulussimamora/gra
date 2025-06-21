package dbcontext

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestUser2 for enhanced set testing to avoid conflicts
type TestUser2 struct {
	ID         int64  `db:"id"`
	Name       string `db:"name"`
	Email      string `db:"email"`
	IsActive   bool   `db:"is_active"`
	Age        int    `db:"age"`
	Department string `db:"department"`
}

func (TestUser2) TableName() string {
	return "test_users_2"
}

// TestProject for join testing
type TestProject struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	UserID int64  `db:"user_id"`
}

func (TestProject) TableName() string {
	return "test_projects"
}

func setupEnhancedSetTestDB(t *testing.T) (*sql.DB, *EnhancedDbContext) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test table
	createTableQuery := `
		CREATE TABLE test_users_2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			is_active BOOLEAN DEFAULT 1,
			age INTEGER,
			department TEXT
		)
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create projects table for join testing
	createProjectsQuery := `
		CREATE TABLE test_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			user_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES test_users_2(id)
		)
	`
	if _, err := db.Exec(createProjectsQuery); err != nil {
		t.Fatalf("Failed to create projects table: %v", err)
	}

	// Insert test data
	testData := []TestUser2{
		{Name: "Alice Johnson", Email: "alice@test.com", IsActive: true, Age: 25, Department: "Engineering"},
		{Name: "Bob Smith", Email: "bob@test.com", IsActive: true, Age: 30, Department: "Marketing"},
		{Name: "Charlie Brown", Email: "charlie@test.com", IsActive: false, Age: 35, Department: "Engineering"},
		{Name: "Diana Prince", Email: "diana@test.com", IsActive: true, Age: 28, Department: "HR"},
		{Name: "Eve Adams", Email: "eve@test.com", IsActive: false, Age: 32, Department: "Marketing"},
	}

	for _, user := range testData {
		_, err := db.Exec("INSERT INTO test_users_2 (name, email, is_active, age, department) VALUES (?, ?, ?, ?, ?)",
			user.Name, user.Email, user.IsActive, user.Age, user.Department)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Insert project data
	projects := []TestProject{
		{Name: "Project Alpha", UserID: 1},
		{Name: "Project Beta", UserID: 2},
		{Name: "Project Gamma", UserID: 1},
	}

	for _, project := range projects {
		_, err := db.Exec("INSERT INTO test_projects (name, user_id) VALUES (?, ?)",
			project.Name, project.UserID)
		if err != nil {
			t.Fatalf("Failed to insert project data: %v", err)
		}
	}

	ctx := NewEnhancedDbContextWithDB(db)

	return db, ctx
}

func TestNewEnhancedSet(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)
	if set == nil {
		t.Fatal("NewEnhancedSet should not return nil")
	}

	if set.builder == nil {
		t.Fatal("QueryBuilder should not be nil")
	}

	if set.builder.tableName != "test_users_2" {
		t.Errorf("Expected table name 'test_users_2', got '%s'", set.builder.tableName)
	}
}

func TestEnhancedSetWhereNull(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test WhereNull
	result := set.WhereNull("department")
	if result == nil {
		t.Error("WhereNull should return the set for chaining")
	}

	// Verify the where clause was added
	if len(set.builder.whereClauses) != 1 {
		t.Errorf("Expected 1 where clause, got %d", len(set.builder.whereClauses))
	}

	clause := set.builder.whereClauses[0]
	if clause.Column != "department" {
		t.Errorf("Expected column 'department', got '%s'", clause.Column)
	}
	if clause.Operator != "IS NULL" {
		t.Errorf("Expected operator 'IS NULL', got '%s'", clause.Operator)
	}
	if clause.Value != nil {
		t.Errorf("Expected nil value, got %v", clause.Value)
	}
}

func TestEnhancedSetWhereNotNull(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test WhereNotNull
	result := set.WhereNotNull("email")
	if result == nil {
		t.Error("WhereNotNull should return the set for chaining")
	}

	// Verify the where clause was added
	if len(set.builder.whereClauses) != 1 {
		t.Errorf("Expected 1 where clause, got %d", len(set.builder.whereClauses))
	}

	clause := set.builder.whereClauses[0]
	if clause.Column != "email" {
		t.Errorf("Expected column 'email', got '%s'", clause.Column)
	}
	if clause.Operator != "IS NOT NULL" {
		t.Errorf("Expected operator 'IS NOT NULL', got '%s'", clause.Operator)
	}
}

func TestEnhancedSetOrderByDesc(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test OrderByDesc
	result := set.OrderByDesc("age")
	if result == nil {
		t.Error("OrderByDesc should return the set for chaining")
	}

	// Verify the order clause was added
	if len(set.builder.orderClauses) != 1 {
		t.Errorf("Expected 1 order clause, got %d", len(set.builder.orderClauses))
	}

	clause := set.builder.orderClauses[0]
	if clause.Column != "age" {
		t.Errorf("Expected column 'age', got '%s'", clause.Column)
	}
	if !clause.Desc {
		t.Error("Expected desc to be true")
	}
}

func TestEnhancedSetSelectFields(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test Select with specific fields
	result := set.Select("name", "email", "age")
	if result == nil {
		t.Error("Select should return the set for chaining")
	}

	// Verify the select fields were set
	expected := []string{"name", "email", "age"}
	if len(set.builder.selectFields) != len(expected) {
		t.Errorf("Expected %d select fields, got %d", len(expected), len(set.builder.selectFields))
	}

	for i, field := range expected {
		if set.builder.selectFields[i] != field {
			t.Errorf("Expected field '%s' at index %d, got '%s'", field, i, set.builder.selectFields[i])
		}
	}
}

func TestEnhancedSetDistinct(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test Distinct
	result := set.Distinct()
	if result == nil {
		t.Error("Distinct should return the set for chaining")
	}

	// Verify distinct flag was set
	if !set.builder.distinct {
		t.Error("Expected distinct to be true")
	}
}

func TestEnhancedSetGroupBy(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test GroupBy
	result := set.GroupBy("department", "is_active")
	if result == nil {
		t.Error("GroupBy should return the set for chaining")
	}

	// Verify group by clauses were set
	expected := []string{"department", "is_active"}
	if len(set.builder.groupBy) != len(expected) {
		t.Errorf("Expected %d group by fields, got %d", len(expected), len(set.builder.groupBy))
	}

	for i, field := range expected {
		if set.builder.groupBy[i] != field {
			t.Errorf("Expected group by field '%s' at index %d, got '%s'", field, i, set.builder.groupBy[i])
		}
	}
}

func TestEnhancedSetHaving(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test Having
	result := set.Having("COUNT(*)", ">", 5)
	if result == nil {
		t.Error("Having should return the set for chaining")
	}

	// Verify the having clause was added
	if len(set.builder.having) != 1 {
		t.Errorf("Expected 1 having clause, got %d", len(set.builder.having))
	}

	clause := set.builder.having[0]
	if clause.Column != "COUNT(*)" {
		t.Errorf("Expected column 'COUNT(*)', got '%s'", clause.Column)
	}
	if clause.Operator != ">" {
		t.Errorf("Expected operator '>', got '%s'", clause.Operator)
	}
	if clause.Value != 5 {
		t.Errorf("Expected value 5, got %v", clause.Value)
	}
}

func TestEnhancedSetJoins(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test InnerJoin
	set.InnerJoin("test_projects", "test_projects.user_id = test_users_2.id")
	if len(set.builder.joinClauses) != 1 {
		t.Errorf("Expected 1 join clause, got %d", len(set.builder.joinClauses))
	}

	join := set.builder.joinClauses[0]
	if join.Type != "INNER" {
		t.Errorf("Expected join type 'INNER', got '%s'", join.Type)
	}
	if join.Table != "test_projects" {
		t.Errorf("Expected table 'test_projects', got '%s'", join.Table)
	}

	// Test LeftJoin
	set.LeftJoin("other_table", "other_table.id = test_users_2.id")
	if len(set.builder.joinClauses) != 2 {
		t.Errorf("Expected 2 join clauses, got %d", len(set.builder.joinClauses))
	}

	leftJoin := set.builder.joinClauses[1]
	if leftJoin.Type != "LEFT" {
		t.Errorf("Expected join type 'LEFT', got '%s'", leftJoin.Type)
	}

	// Test RightJoin
	set.RightJoin("right_table", "right_table.id = test_users_2.id")
	if len(set.builder.joinClauses) != 3 {
		t.Errorf("Expected 3 join clauses, got %d", len(set.builder.joinClauses))
	}

	rightJoin := set.builder.joinClauses[2]
	if rightJoin.Type != "RIGHT" {
		t.Errorf("Expected join type 'RIGHT', got '%s'", rightJoin.Type)
	}
}

func TestEnhancedSetSingle(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test Single with exactly one result
	user, err := set.Where("name", "=", "Alice Johnson").Single()
	if err != nil {
		t.Errorf("Single should not return error for one result: %v", err)
	}
	if user.Name != "Alice Johnson" {
		t.Errorf("Expected name 'Alice Johnson', got '%s'", user.Name)
	}

	// Test Single with no results
	set2 := NewEnhancedSet[TestUser2](ctx)
	_, err = set2.Where("name", "=", "Nonexistent").Single()
	if err == nil {
		t.Error("Single should return error when no results found")
	}

	// Test Single with multiple results
	set3 := NewEnhancedSet[TestUser2](ctx)
	_, err = set3.Where("is_active", "=", true).Single()
	if err == nil {
		t.Error("Single should return error when multiple results found")
	}
}

func TestEnhancedSetComplexQuery(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test complex query with multiple clauses
	users, err := set.
		Where("is_active", "=", true).
		WhereOr("age", ">", 30).
		OrderBy("name").
		Take(10).
		Skip(0).
		ToList()

	if err != nil {
		t.Errorf("Complex query should not return error: %v", err)
	}

	if len(users) == 0 {
		t.Error("Complex query should return some results")
	}

	// Verify results are sorted by name
	for i := 1; i < len(users); i++ {
		if users[i-1].Name > users[i].Name {
			t.Error("Results should be sorted by name")
		}
	}
}

func TestEnhancedSetQueryBuilding(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Build a complex query to test query building
	set.Select("name", "email").
		Where("is_active", "=", true).
		WhereIn("department", []interface{}{"Engineering", "Marketing"}).
		OrderBy("name").
		OrderByDesc("age").
		Take(5).
		Skip(1).
		Distinct()

	query, args := set.builder.buildSelectQuery()

	// Verify query contains expected parts
	expectedParts := []string{
		"SELECT DISTINCT name, email",
		"FROM test_users_2",
		"WHERE is_active = ?",
		"AND department IN (?, ?)",
		"ORDER BY name, age DESC",
		"LIMIT 5",
		"OFFSET 1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(query, part) {
			t.Errorf("Query should contain '%s', got: %s", part, query)
		}
	}

	// Verify args
	expectedArgs := 3 // is_active + 2 department values
	if len(args) != expectedArgs {
		t.Errorf("Expected %d args, got %d", expectedArgs, len(args))
	}
}

func TestGetTableNameFromType(t *testing.T) {
	// Test with struct that has TableName method
	user := TestUser2{}
	userType := reflect.TypeOf(user)
	tableName := getTableNameFromType(userType)
	if tableName != "test_users_2" {
		t.Errorf("Expected 'test_users_2', got '%s'", tableName)
	}

	// Test with struct without TableName method
	type SimpleStruct struct {
		ID   int
		Name string
	}
	simple := SimpleStruct{}
	simpleType := reflect.TypeOf(simple)
	simpleTableName := getTableNameFromType(simpleType)
	if simpleTableName != "simplestructs" {
		t.Errorf("Expected 'simplestructs', got '%s'", simpleTableName)
	}

	// Test with pointer type
	ptrType := reflect.TypeOf(&user)
	ptrTableName := getTableNameFromType(ptrType)
	if ptrTableName != "test_users_2" {
		t.Errorf("Expected 'test_users_2' for pointer type, got '%s'", ptrTableName)
	}
}

func TestEnhancedSetErrorCases(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	// Close the database to simulate connection errors
	db.Close()

	set := NewEnhancedSet[TestUser2](ctx)

	// Test ToList with closed database
	_, err := set.ToList()
	if err == nil {
		t.Error("ToList should return error with closed database")
	}

	// Test Count with closed database
	_, err = set.Count()
	if err == nil {
		t.Error("Count should return error with closed database")
	}

	// Test Any with closed database
	_, err = set.Any()
	if err == nil {
		t.Error("Any should return error with closed database")
	}

	// Test First with closed database
	_, err = set.First()
	if err == nil {
		t.Error("First should return error with closed database")
	}
}

func TestEnhancedSetWithTransaction(t *testing.T) {
	db, ctx := setupEnhancedSetTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := ctx.Database.db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create context with transaction
	txCtx := NewEnhancedDbContextWithTx(tx)

	set := NewEnhancedSet[TestUser2](txCtx)

	// Test query within transaction
	users, err := set.Where("is_active", "=", true).ToList()
	if err != nil {
		t.Errorf("Query within transaction should not return error: %v", err)
	}

	if len(users) == 0 {
		t.Error("Query within transaction should return results")
	}

	// Test count within transaction
	count, err := set.Count()
	if err != nil {
		t.Errorf("Count within transaction should not return error: %v", err)
	}

	if count <= 0 {
		t.Error("Count within transaction should return positive value")
	}
}

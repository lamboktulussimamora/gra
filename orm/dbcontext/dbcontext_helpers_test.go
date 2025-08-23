package dbcontext

import (
	"reflect"
	"testing"
	"time"
)

// Test types
type AuditFields struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type userEntity struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
	AuditFields
}

type customTable struct{ X int }

func (customTable) TableName() string { return "custom_table" }

func TestToSnakeAndCamelCase(t *testing.T) {
	if got := toSnakeCase("UserProfile"); got != "user_profile" {
		t.Fatalf("toSnakeCase: expected user_profile, got %s", got)
	}
	if got := toCamelCase("user_profile"); got != "UserProfile" {
		t.Fatalf("toCamelCase: expected UserProfile, got %s", got)
	}
}

func TestGetPlaceholder(t *testing.T) {
	if p := getPlaceholder("postgres", 0); p != "$1" {
		t.Fatalf("expected $1, got %s", p)
	}
	if p := getPlaceholder("sqlite3", 0); p != "?" {
		t.Fatalf("expected ?, got %s", p)
	}
	if p := getPlaceholder("mysql", 3); p != "?" {
		t.Fatalf("expected ?, got %s", p)
	}
}

func TestGetTableName(t *testing.T) {
	u := &userEntity{}
	if tn := getTableName(u); tn != "user_entity" {
		t.Fatalf("expected user_entity, got %s", tn)
	}
	c := &customTable{}
	if tn := getTableName(c); tn != "custom_table" {
		t.Fatalf("expected custom_table, got %s", tn)
	}
}

func TestGetFieldData_InsertExcludeID_Embedded(t *testing.T) {
	type Inner struct {
		Hidden string `sql:"-"`
		Code   string `db:"code"`
	}
	type sample struct {
		ID   int64
		Name string
		Inner
	}
	s := &sample{ID: 10, Name: "A", Inner: Inner{Hidden: "X", Code: "C"}}
	cols, vals, ph := getFieldData(s, true, "postgres")
	// ID excluded, Hidden excluded by sql:"-", Name and code included
	if !reflect.DeepEqual(cols, []string{"name", "code"}) {
		t.Fatalf("columns mismatch: %v", cols)
	}
	if len(vals) != 2 || vals[0] != "A" || vals[1] != "C" {
		t.Fatalf("values mismatch: %v", vals)
	}
	if !reflect.DeepEqual(ph, []string{"$1", "$2"}) {
		t.Fatalf("placeholders mismatch: %v", ph)
	}
}

func TestGetUpdateData_WithID_PostgresAndMySQL(t *testing.T) {
	s := &userEntity{ID: 42, Name: "N", Age: 7}
	setPG, valsPG, idPG := getUpdateData(s, "postgres")
	if idPG != int64(42) {
		t.Fatalf("expected id 42, got %v", idPG)
	}
	if !reflect.DeepEqual(setPG, []string{"name = $1", "age = $2", "created_at = $3", "updated_at = $4"}) {
		t.Fatalf("setPG mismatch: %v", setPG)
	}
	if len(valsPG) != 4 {
		t.Fatalf("valsPG len: %d", len(valsPG))
	}

	setMY, valsMY, idMY := getUpdateData(s, "mysql")
	if idMY != int64(42) {
		t.Fatalf("expected id 42, got %v", idMY)
	}
	// order should match columns sequence except id
	if !reflect.DeepEqual(setMY, []string{"name = ?", "age = ?", "created_at = ?", "updated_at = ?"}) {
		t.Fatalf("setMY mismatch: %v", setMY)
	}
	if len(valsMY) != 4 {
		t.Fatalf("valsMY len: %d", len(valsMY))
	}
}

func TestIDHelpers_WithEmbedding(t *testing.T) {
	s := &userEntity{ID: 0}
	if v := getIDValue(s); v != int64(0) {
		t.Fatalf("expected 0, got %v", v)
	}
	setIDField(s, 99)
	if v := getIDValue(s); v != int64(99) {
		t.Fatalf("expected 99, got %v", v)
	}
}

func TestSetTimestamps_CreateAndUpdate(t *testing.T) {
	s := &userEntity{}
	if !s.CreatedAt.IsZero() || !s.UpdatedAt.IsZero() {
		t.Fatalf("timestamps should start zero")
	}
	setTimestamps(s, true)
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not set on create")
	}
	firstCreated := s.CreatedAt
	time.Sleep(time.Millisecond)
	setTimestamps(s, false)
	if !s.CreatedAt.Equal(firstCreated) {
		t.Fatalf("CreatedAt should not change on update")
	}
	if s.UpdatedAt.Equal(firstCreated) {
		t.Fatalf("UpdatedAt should change on update")
	}
}

func TestEnhancedDbSet_AdjustPlaceholdersForCondition(t *testing.T) {
	ctx := &EnhancedDbContext{driver: driverPostgres}
	set := &EnhancedDbSet[userEntity]{ctx: ctx, whereArgs: []interface{}{1}}
	out := set.adjustPlaceholdersForCondition("name = ? AND age > ?")
	if out != "name = $2 AND age > $3" {
		t.Fatalf("unexpected adjusted condition: %s", out)
	}
}

func TestEnhancedDbSet_BuildQuery_AllClauses(t *testing.T) {
	set := &EnhancedDbSet[userEntity]{
		tableName:   "users",
		whereClause: "age > 18",
		orderClause: "name DESC",
		limitValue:  10,
		offsetValue: 5,
	}
	q := set.buildQuery()
	exp := "SELECT * FROM users WHERE age > 18 ORDER BY name DESC LIMIT 10 OFFSET 5"
	if q != exp {
		t.Fatalf("expected %q, got %q", exp, q)
	}
}

func TestEnhancedDbSet_Chaining_WithPlaceholders(t *testing.T) {
	// sqlite/mysql style
	set1 := &EnhancedDbSet[userEntity]{tableName: "users", ctx: &EnhancedDbContext{driver: "sqlite3"}}
	q1 := set1.Where("name = ?", "A").WhereLike("name", "A%").WhereIn("age", []interface{}{1, 2}).WhereOr("id = ?", 1).OrderByDescending("name").Take(3).Skip(1).buildQuery()
	exp1 := "SELECT * FROM users WHERE name = ? AND name LIKE ? AND age IN (?, ?) OR (id = ?) ORDER BY name DESC LIMIT 3 OFFSET 1"
	if q1 != exp1 {
		t.Fatalf("exp1: %q, got %q", exp1, q1)
	}

	// postgres style
	set2 := &EnhancedDbSet[userEntity]{tableName: "users", ctx: &EnhancedDbContext{driver: driverPostgres}}
	q2 := set2.Where("name = ?", "A").WhereLike("name", "A%").WhereIn("age", []interface{}{1, 2}).OrderBy("name").Take(2).buildQuery()
	exp2 := "SELECT * FROM users WHERE name = $1 AND name LIKE $2 AND age IN ($3, $4) ORDER BY name LIMIT 2"
	if q2 != exp2 {
		t.Fatalf("exp2: %q, got %q", exp2, q2)
	}
}

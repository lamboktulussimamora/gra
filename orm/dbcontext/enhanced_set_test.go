package dbcontext

import (
	"reflect"
	"strings"
	"testing"
)

// Test entity with explicit table name
type esUser struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Email     string  `db:"email"`
	Role      string  `db:"role"`
	DeletedAt *string `db:"deleted_at"`
	UpdatedAt *string `db:"updated_at"`
}

func (esUser) TableName() string { return "es_users" }

// Test entity without explicit TableName to verify default naming (lower + s)
type Widget struct {
	ID int
}

func TestEnhancedSet_buildSelectQuery_FullChain(t *testing.T) {
	// No DB operations are performed; we only exercise the query builder
	ctx := &EnhancedDbContext{}
	set := NewEnhancedSet[esUser](ctx)

	// Chain a variety of operations
	set = set.
		Where("name", "=", "Alice").
		Where("age", ">", 21).
		WhereIn("role", []interface{}{"admin", "user"}).
		WhereLike("email", "%@example.com").
		WhereNull("deleted_at").
		WhereNotNull("updated_at").
		OrderByDesc("created_at").
		OrderBy("id").
		Select("id", "name").
		Distinct().
		InnerJoin("roles r", "r.id = es_users.role_id").
		LeftJoin("departments d", "d.id = es_users.dept_id").
		GroupBy("role").
		Having("COUNT(id)", ">", 1).
		Skip(5).
		Take(10)

	query, args := set.builder.buildSelectQuery()

	expected := "SELECT DISTINCT id, name FROM es_users INNER JOIN roles r ON r.id = es_users.role_id LEFT JOIN departments d ON d.id = es_users.dept_id WHERE name = = ? age > ? role IN (?, ?) email LIKE ? deleted_at IS NULL updated_at IS NOT NULL GROUP BY role HAVING COUNT(id) > ? ORDER BY created_at DESC, id LIMIT 10 OFFSET 5"

	// We won't assert strict equality on the entire string to keep test resilient to spacing;
	// instead, verify key fragments and ordering.
	mustContainInOrder := []string{
		"SELECT DISTINCT id, name",
		"FROM es_users",
		"INNER JOIN roles r ON r.id = es_users.role_id",
		"LEFT JOIN departments d ON d.id = es_users.dept_id",
		"WHERE name = ?",
		"AND age > ?",
		"AND role IN (", // parentheses with placeholders
		")",
		"AND email LIKE ?",
		"AND deleted_at IS NULL",
		"AND updated_at IS NOT NULL",
		"GROUP BY role",
		"HAVING COUNT(id) > ?",
		"ORDER BY created_at DESC, id",
		"LIMIT 10",
		"OFFSET 5",
	}

	// Normalize whitespace to single spaces for robust contains checks
	normalized := strings.Join(strings.Fields(query), " ")
	lastPos := 0
	for _, frag := range mustContainInOrder {
		idx := strings.Index(normalized[lastPos:], frag)
		if idx < 0 {
			t.Fatalf("expected query to contain fragment in order: %q\nquery: %s", frag, query)
		}
		lastPos += idx + len(frag)
	}

	// Args order: name, age, role values (2), like pattern, having value
	if len(args) != 6 {
		t.Fatalf("expected 6 args, got %d: %#v", len(args), args)
	}
	if args[0] != "Alice" || args[1] != 21 || args[2] != "admin" || args[3] != "user" || args[4] != "%@example.com" || args[5] != 1 {
		t.Fatalf("unexpected args: %#v", args)
	}
	_ = expected // keep linter happy about unused variable in case of future equality check
}

func TestEnhancedSet_WhereOr(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedSet[esUser](ctx)
	set = set.Where("role", "=", "user").WhereOr("name", "=", "Bob")

	query, args := set.builder.buildSelectQuery()

	// Expect WHERE role = ? OR name = ? (order matters based on call sequence)
	normalized := strings.Join(strings.Fields(query), " ")
	if !strings.Contains(normalized, "WHERE role = ? OR name = ?") {
		t.Fatalf("unexpected WHERE clause: %s", query)
	}
	if len(args) != 2 || args[0] != "user" || args[1] != "Bob" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestEnhancedSet_JoinsOnly(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedSet[esUser](ctx)
	set = set.InnerJoin("roles r", "r.id = es_users.role_id").RightJoin("teams t", "t.id = es_users.team_id")
	query, _ := set.builder.buildSelectQuery()
	normalized := strings.Join(strings.Fields(query), " ")
	if !strings.Contains(normalized, "FROM es_users INNER JOIN roles r ON r.id = es_users.role_id RIGHT JOIN teams t ON t.id = es_users.team_id") {
		t.Fatalf("unexpected JOIN sequence: %s", query)
	}
}

func TestEnhancedSet_DefaultTableName(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedSet[Widget](ctx)
	if set.builder.tableName != "widgets" {
		t.Fatalf("expected default table name 'widgets', got %q", set.builder.tableName)
	}
}

func TestEnhancedSet_findFieldByDbTag(t *testing.T) {
	ctx := &EnhancedDbContext{}
	set := NewEnhancedSet[esUser](ctx)
	val := reflect.ValueOf(&esUser{}).Elem()
	f := set.findFieldByDbTag(val, "email")
	if !f.IsValid() {
		t.Fatalf("expected to find field by db tag 'email'")
	}
}

package dbcontext

import (
	"reflect"
	"testing"
)

// Test the LINQ-style QueryBuilder via EnhancedSet by inspecting the built SQL and args.
func TestQueryBuilder_BuildSelectQuery_AllClauses(t *testing.T) {
	type user struct{}

	ctx := &EnhancedDbContext{} // ctx not used by buildSelectQuery
	es := NewEnhancedSet[user](ctx)

	// Chain various clauses to exercise most builder branches
	es = es.
		Where("age", ">", 21).
		WhereOr("status", "=", "active").
		WhereIn("id", []interface{}{1, 2, 3}).
		GroupBy("status").
		Having("COUNT(*)", ">", 1).
		OrderBy("created_at").
		Distinct().
		Select("id", "name").
		Take(10).
		Skip(5)

	// Access the internal builder (same package) and build the query
	sql, args := es.builder.buildSelectQuery()

	// Basic shape assertions
	wantContains := []string{
		"SELECT DISTINCT id, name FROM users",
		"WHERE age > ? OR status = ? AND id IN (?, ?, ?)",
		"GROUP BY status",
		"HAVING COUNT(*) > ?",
		"ORDER BY created_at",
		"LIMIT 10",
		"OFFSET 5",
	}
	for _, part := range wantContains {
		if !contains(sql, part) {
			t.Fatalf("expected SQL to contain %q, got: %s", part, sql)
		}
	}

	// Args order: where (age, status, in...), then having (count value)
	wantArgs := []interface{}{21, "active", 1, 2, 3, 1}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch: want %v, got %v", wantArgs, args)
	}
}

// tiny helper to avoid importing strings in multiple tests
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0)))
}

// simple indexOf without exposing strings package to avoid naming conflicts; acceptable for tests
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

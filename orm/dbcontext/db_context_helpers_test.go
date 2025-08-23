package dbcontext

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestToSnakeAndCamelCase_DBContextHelpers(t *testing.T) {
	if got := toSnakeCase("UserID"); got != "user_i_d" { // current simple behavior
		t.Fatalf("toSnakeCase unexpected: %q", got)
	}
	if got := toCamelCase("created_at"); got != "CreatedAt" {
		t.Fatalf("toCamelCase unexpected: %q", got)
	}
}

func TestGetPlaceholder_Drivers(t *testing.T) {
	if p := getPlaceholder("postgres", 0); p != "$1" {
		t.Fatalf("expected $1, got %q", p)
	}
	if p := getPlaceholder("sqlite3", 3); p != "?" {
		t.Fatalf("expected ?, got %q", p)
	}
}

func TestSetFieldValue_AllKinds(t *testing.T) {
	type S struct {
		A string
		I int64
		U uint64
		F float64
		B bool
		T time.Time
	}
	var s S
	v := reflect.ValueOf(&s).Elem()

	// string
	if err := setFieldValue(v.FieldByName("A"), []byte("hello")); err != nil {
		t.Fatal(err)
	}
	// int
	if err := setFieldValue(v.FieldByName("I"), int64(42)); err != nil {
		t.Fatal(err)
	}
	// uint (from string)
	if err := setFieldValue(v.FieldByName("U"), "7"); err != nil {
		t.Fatal(err)
	}
	// float
	if err := setFieldValue(v.FieldByName("F"), 3.14); err != nil {
		t.Fatal(err)
	}
	// bool (from int64)
	if err := setFieldValue(v.FieldByName("B"), int64(1)); err != nil {
		t.Fatal(err)
	}
	// time (from string)
	if err := setFieldValue(v.FieldByName("T"), "2006-01-02 15:04:05"); err != nil {
		t.Fatal(err)
	}

	if s.A != "hello" || s.I != 42 || s.U != 7 || s.F == 0 || !s.B || s.T.IsZero() {
		t.Fatalf("unexpected assigned values: %+v", s)
	}
}

// Tiny smoke to exercise convertQueryPlaceholders for postgres vs others using a fake context
func TestConvertQueryPlaceholders(t *testing.T) {
	q := "SELECT * FROM t WHERE a=? AND b=?"
	if got := convertQueryPlaceholders(q, "postgres"); got != "SELECT * FROM t WHERE a=$1 AND b=$2" {
		t.Fatalf("unexpected converted: %q", got)
	}
	if got := convertQueryPlaceholders(q, "sqlite3"); got != q {
		t.Fatalf("sqlite should keep '?', got %q", got)
	}
}

// Ensure buildQuery respects order/limit/offset concatenation (no DB calls)
func TestEnhancedDbSet_buildQuery_Composition(t *testing.T) {
	type temp struct{ ID int }
	ctx := &EnhancedDbContext{db: &sql.DB{}}
	set := NewEnhancedDbSet[temp](ctx).
		Where("name = ?", "a").
		OrderBy("id").
		Take(5).
		Skip(2)
	q := set.buildQuery()
	if want := "SELECT * FROM temp WHERE name = ? ORDER BY id LIMIT 5 OFFSET 2"; q != want {
		t.Fatalf("unexpected query: %q", q)
	}
}

package dbcontext

import (
	"reflect"
	"testing"
	"time"
)

func TestToSnakeCase_EdgeAcronyms(t *testing.T) {
	if got := toSnakeCase("HTTPServer"); got != "h_t_t_p_server" { // current simple logic
		t.Fatalf("toSnakeCase edge: expected h_t_t_p_server, got %s", got)
	}
}

func TestToCamelCase_Empty(t *testing.T) {
	if got := toCamelCase(""); got != "" {
		t.Fatalf("toCamelCase empty: expected empty, got %s", got)
	}
}

func TestSetTimeField_WithValidAndInvalid(t *testing.T) {
	type T struct{ When time.Time }
	var v T
	rv := reflect.ValueOf(&v).Elem().FieldByName("When")

	setTimeField(rv, "2024-01-02 03:04:05")
	if v.When.Year() != 2024 || v.When.Month() != 1 || v.When.Day() != 2 {
		t.Fatalf("expected parsed time set, got %v", v.When)
	}

	prev := v.When
	setTimeField(rv, "invalid") // should not change
	if !v.When.Equal(prev) {
		t.Fatalf("invalid time should not change field, got %v vs %v", v.When, prev)
	}
}

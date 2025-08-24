package main

import "testing"

func TestTraverse_ObjectPath(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": 42.0,
			},
		},
	}
	v, ok := traverse(data, []string{"a", "b", "c"})
	if !ok {
		t.Fatalf("expected path to exist")
	}
	if v.(float64) != 42.0 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestTraverse_ArrayIndex(t *testing.T) {
	data := map[string]any{
		"arr": []any{"x", 7.0, true},
	}
	v0, ok0 := traverse(data, []string{"arr", "0"})
	if !ok0 || v0.(string) != "x" {
		t.Fatalf("expected arr[0] == 'x', got %v, ok=%v", v0, ok0)
	}
	v1, ok1 := traverse(data, []string{"arr", "1"})
	if !ok1 || v1.(float64) != 7.0 {
		t.Fatalf("expected arr[1] == 7.0, got %v, ok=%v", v1, ok1)
	}
	v2, ok2 := traverse(data, []string{"arr", "2"})
	if !ok2 || v2.(bool) != true {
		t.Fatalf("expected arr[2] == true, got %v, ok=%v", v2, ok2)
	}
}

func TestTraverse_InvalidPath(t *testing.T) {
	data := map[string]any{"k": 1}
	if _, ok := traverse(data, []string{"missing"}); ok {
		t.Fatalf("expected missing path to return ok=false")
	}
	if _, ok := traverse([]any{"a"}, []string{"notAnIndex"}); ok {
		t.Fatalf("expected invalid index to return ok=false")
	}
}

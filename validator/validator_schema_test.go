package validator

import "testing"

// Tests for Schema-based validation (DB-free) to increase coverage

func TestSchema_String_Number_Array_Validation(t *testing.T) {
	s := NewSchema().
		AddField("name", SchemaField{
			Type:      "string",
			Required:  true,
			MinLength: 3,
			MaxLength: 5,
			Pattern:   "^[A-Z][a-z]+$",
			Enum:      []string{"Alice", "Bob"},
		}).
		AddField("age", SchemaField{
			Type: "number",
			Min:  1,
			Max:  10,
		}).
		AddField("tags", SchemaField{
			Type:      "array",
			MinLength: 2,
		})

	// Happy path
	ok := map[string]any{
		"name": "Alice",
		"age":  5, // int acceptable
		"tags": []any{"a", "b"},
	}
	if errs := s.Validate(ok); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	// Missing required name
	missingName := map[string]any{
		"age":  3,
		"tags": []any{"a", "b"},
	}
	errs := s.Validate(missingName)
	if len(errs) != 1 || errs[0].Field != "name" {
		t.Fatalf("expected one error for name is required, got %v", errs)
	}

	// Type mismatch for name
	wrongType := map[string]any{
		"name": 123, // not a string
		"age":  3,
		"tags": []any{"a", "b"},
	}
	errs = s.Validate(wrongType)
	if len(errs) != 1 || errs[0].Field != "name" || errs[0].Message != "name must be a string" {
		t.Fatalf("expected type error for name string, got %v", errs)
	}

	// Pattern mismatch
	badPattern := map[string]any{
		"name": "alice", // fails pattern (starts lowercase) and enum
		"age":  3.5,     // float acceptable
		"tags": []any{"a", "b"},
	}
	errs = s.Validate(badPattern)
	if len(errs) == 0 {
		t.Fatalf("expected errors for name pattern/enum, got none")
	}

	// Array length below minimum
	arrTooShort := map[string]any{
		"name": "Bob",
		"age":  2,
		"tags": []any{"only"},
	}
	errs = s.Validate(arrTooShort)
	if len(errs) != 1 || errs[0].Field != "tags" {
		t.Fatalf("expected one error for tags min length, got %v", errs)
	}
}

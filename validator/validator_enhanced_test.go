package validator

import (
	"testing"
)

// TestValidatorEnhancedCoverage provides additional tests to improve code coverage
func TestValidatorEnhancedCoverage(t *testing.T) {
	t.Run("validateEmail with non-string field", func(t *testing.T) {
		v := New()
		// Test struct with non-string email field
		type TestStruct struct {
			Email int `validate:"email"`
		}

		data := TestStruct{Email: 123}
		v.Validate(data)

		// Should not generate errors for non-string fields
		if v.HasErrors() {
			t.Errorf("Expected no errors for non-string email field, got: %v", v.errors)
		}
	})

	t.Run("validateEmail with empty string", func(t *testing.T) {
		v := New()
		type TestStruct struct {
			Email string `validate:"email"`
		}

		data := TestStruct{Email: ""}
		v.Validate(data)

		// Should not generate errors for empty email (handled by required rule)
		if v.HasErrors() {
			t.Errorf("Expected no errors for empty email, got: %v", v.errors)
		}
	})

	t.Run("validateRegexp with non-string field", func(t *testing.T) {
		v := New()
		type TestStruct struct {
			Value int `validate:"regexp:^[0-9]+$"`
		}

		data := TestStruct{Value: 123}
		v.Validate(data)

		// Should not generate errors for non-string fields
		if v.HasErrors() {
			t.Errorf("Expected no errors for non-string regexp field, got: %v", v.errors)
		}
	})

	t.Run("HasBatchErrors with empty batch", func(t *testing.T) {
		v := New()
		results := make([]BatchResult, 0)
		hasErrors := v.HasBatchErrors(results)

		if hasErrors {
			t.Error("Expected no errors for empty batch")
		}
	})

	t.Run("HasBatchErrors with mixed results", func(t *testing.T) {
		v := New()
		results := []BatchResult{
			{Index: 0, Errors: []ValidationError{}},
			{Index: 1, Errors: []ValidationError{{Field: "test", Message: "error"}}},
		}
		hasErrors := v.HasBatchErrors(results)

		if !hasErrors {
			t.Error("Expected errors when batch contains errors")
		}
	})

	t.Run("addAnchorsIfNeeded behavior through regex validation", func(t *testing.T) {
		// Test the pattern anchor addition logic indirectly through regex validation
		v := New()
		type TestStruct struct {
			Username string `validate:"regexp:[a-z0-9_]{3,16}"`
		}

		// This should work because addAnchorsIfNeeded should add anchors
		data := TestStruct{Username: "validuser123"}
		v.Validate(data)

		if v.HasErrors() {
			t.Errorf("Expected no errors for valid username with anchor-needing pattern, got: %v", v.errors)
		}
	})

	t.Run("schema validation with basic types", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("name", SchemaField{Type: "string", Required: true})
		schema.AddField("age", SchemaField{Type: "number", Min: 0, Max: 150})

		// Test valid data
		validData := map[string]interface{}{
			"name": "John Doe",
			"age":  25,
		}

		result := schema.Validate(validData)
		if len(result) > 0 {
			t.Errorf("Expected no errors for valid data, got: %v", result)
		}

		// Test missing required field
		invalidData := map[string]interface{}{
			"age": 25,
		}

		result = schema.Validate(invalidData)
		if len(result) == 0 {
			t.Error("Expected error for missing required field")
		}
	})

	t.Run("FilterInvalid with mixed batch results", func(t *testing.T) {
		v := New()
		results := []BatchResult{
			{Index: 0, Errors: []ValidationError{}},
			{Index: 1, Errors: []ValidationError{{Field: "test", Message: "error"}}},
			{Index: 2, Errors: []ValidationError{}},
			{Index: 3, Errors: []ValidationError{{Field: "another", Message: "another error"}}},
		}

		invalidResults := v.FilterInvalid(results)

		if len(invalidResults) != 2 {
			t.Errorf("Expected 2 invalid results, got %d", len(invalidResults))
		}

		// Check that the correct indices were filtered
		if invalidResults[0].Index != 1 || invalidResults[1].Index != 3 {
			t.Errorf("Expected indices 1 and 3, got %d and %d", invalidResults[0].Index, invalidResults[1].Index)
		}
	})
}

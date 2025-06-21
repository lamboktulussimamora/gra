package validator

import (
	"reflect"
	"testing"
)

func TestGetCompiledRegexp(t *testing.T) {
	// Test valid pattern
	pattern := "^[a-zA-Z0-9_]+$"
	regex, err := getCompiledRegexp(pattern)
	if err != nil {
		t.Errorf("Expected no error for valid pattern, got: %v", err)
	}
	if regex == nil {
		t.Error("Expected regex to be non-nil")
	}

	// Test pattern from cache (second call)
	regex2, err2 := getCompiledRegexp(pattern)
	if err2 != nil {
		t.Errorf("Expected no error for cached pattern, got: %v", err2)
	}
	if regex2 != regex {
		t.Error("Expected cached regex to be the same instance")
	}

	// Test invalid pattern
	invalidPattern := "[unclosed"
	_, err = getCompiledRegexp(invalidPattern)
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}

	// Test empty pattern
	emptyPattern := ""
	regex3, err3 := getCompiledRegexp(emptyPattern)
	if err3 != nil {
		t.Errorf("Expected no error for empty pattern, got: %v", err3)
	}
	if regex3 == nil {
		t.Error("Expected regex to be non-nil for empty pattern")
	}
}

func TestParseRulesWithRange(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		validateTag string
		expected    []string
	}{
		{
			name:        "Range only",
			validateTag: "range=1,10",
			expected:    []string{"range=1,10"},
		},
		{
			name:        "Range with other rules before",
			validateTag: "required,range=1,10",
			expected:    []string{"required", "range=1,10"},
		},
		{
			name:        "Range with other rules after",
			validateTag: "range=1,10,email",
			expected:    []string{"range=1,10", "email"},
		},
		{
			name:        "Range with rules before and after",
			validateTag: "required,range=1,10,email",
			expected:    []string{"required", "range=1,10", "email"},
		},
		{
			name:        "Multiple rules with range in middle",
			validateTag: "required,min=5,range=1,100,max=50,email",
			expected:    []string{"required", "min=5", "range=1,100", "max=50", "email"},
		},
		{
			name:        "No range rule (fallback)",
			validateTag: "required,email,min=5",
			expected:    []string{"required", "email", "min=5"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.parseRulesWithRange(test.validateTag)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d rules, got %d", len(test.expected), len(result))
			}
			for i, expected := range test.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected rule %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestParseRulesWithRegexp(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		validateTag string
		expected    []string
	}{
		{
			name:        "Regexp only",
			validateTag: "regexp=^[a-zA-Z]+$",
			expected:    []string{"regexp=^[a-zA-Z]+$"},
		},
		{
			name:        "Regexp with other rules before",
			validateTag: "required,regexp=^[a-zA-Z]+$",
			expected:    []string{"required", "regexp=^[a-zA-Z]+$"},
		},
		{
			name:        "Regexp with other rules after",
			validateTag: "regexp=^[a-zA-Z]+$,min=3",
			expected:    []string{"regexp=^[a-zA-Z]+$", "min=3"},
		},
		{
			name:        "Complex regexp with special characters",
			validateTag: "regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2}$",
			expected:    []string{"regexp=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2}$"},
		},
		{
			name:        "Regexp with rules before and after",
			validateTag: "required,regexp=^[0-9]{10}$,min=10",
			expected:    []string{"required", "regexp=^[0-9]{10}$", "min=10"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.parseRulesWithRegexp(test.validateTag)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d rules, got %d", len(test.expected), len(result))
			}
			for i, expected := range test.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected rule %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestGetFieldName(t *testing.T) {
	validator := New()

	tests := []struct {
		name     string
		prefix   string
		tag      string
		expected string
	}{
		{
			name:     "No prefix",
			prefix:   "",
			tag:      "username",
			expected: "username",
		},
		{
			name:     "With prefix",
			prefix:   "user",
			tag:      "name",
			expected: "user.name",
		},
		{
			name:     "Tag with comma options",
			prefix:   "",
			tag:      "email,omitempty",
			expected: "email",
		},
		{
			name:     "Prefix and tag with options",
			prefix:   "profile",
			tag:      "age,omitempty",
			expected: "profile.age",
		},
		{
			name:     "Deep nesting",
			prefix:   "user.profile",
			tag:      "settings",
			expected: "user.profile.settings",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.getFieldName(test.prefix, test.tag)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestAddError(t *testing.T) {
	validator := New()

	// Test with default message
	validator.addError("field1", "default message", "")
	if len(validator.errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(validator.errors))
	}
	if validator.errors[0].Field != "field1" {
		t.Errorf("Expected field 'field1', got '%s'", validator.errors[0].Field)
	}
	if validator.errors[0].Message != "default message" {
		t.Errorf("Expected message 'default message', got '%s'", validator.errors[0].Message)
	}

	// Test with custom message
	validator.addError("field2", "default message", "custom message")
	if len(validator.errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(validator.errors))
	}
	if validator.errors[1].Message != "custom message" {
		t.Errorf("Expected message 'custom message', got '%s'", validator.errors[1].Message)
	}

	// Test HasErrors
	if !validator.HasErrors() {
		t.Error("Expected HasErrors to be true")
	}
}

func TestValidateSliceOfStructs(t *testing.T) {
	validator := New()

	type NestedStruct struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"min=0"`
	}

	type TestStruct struct {
		Items []NestedStruct `json:"items" validate:"required"`
	}

	// Valid data
	validData := TestStruct{
		Items: []NestedStruct{
			{Name: "John", Age: 25},
			{Name: "Jane", Age: 30},
		},
	}

	errors := validator.Validate(validData)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for valid data, got %d errors", len(errors))
		for _, err := range errors {
			t.Logf("Unexpected error: %s - %s", err.Field, err.Message)
		}
	}

	// Invalid data
	invalidData := TestStruct{
		Items: []NestedStruct{
			{Name: "", Age: 25},     // Missing required name
			{Name: "Jane", Age: -5}, // Invalid age
		},
	}

	validator = New() // Reset validator
	errors = validator.Validate(invalidData)

	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors for invalid data, got %d errors", len(errors))
		for _, err := range errors {
			t.Logf("Error: %s - %s", err.Field, err.Message)
		}
	}

	// Check that errors reference the correct nested fields
	errorFields := make(map[string]bool)
	for _, err := range errors {
		errorFields[err.Field] = true
	}

	expectedFields := []string{"items[0].name", "items[1].age"}
	for _, expectedField := range expectedFields {
		if !errorFields[expectedField] {
			t.Errorf("Expected error for field '%s'", expectedField)
		}
	}
}

func TestProcessFieldEdgeCases(t *testing.T) {
	validator := New()

	// Test with non-struct field (should apply regular validation)
	val := reflect.ValueOf("test")
	validator.processField(val, "testField", "required")

	// Should not add any errors since the field has a value
	if validator.HasErrors() {
		t.Error("Expected no errors for non-empty string field")
	}

	// Test with empty string (should add error)
	validator = New()
	val = reflect.ValueOf("")
	validator.processField(val, "testField", "required")

	if !validator.HasErrors() {
		t.Error("Expected error for empty required string field")
	}
}

func TestConstants(t *testing.T) {
	// Test pattern constants
	if UsernamePattern != "^[a-zA-Z0-9_]{3,20}$" {
		t.Errorf("Expected UsernamePattern to be '^[a-zA-Z0-9_]{3,20}$', got '%s'", UsernamePattern)
	}

	if LowercaseUsernamePattern != "[a-z0-9_]{3,16}" {
		t.Errorf("Expected LowercaseUsernamePattern to be '[a-z0-9_]{3,16}', got '%s'", LowercaseUsernamePattern)
	}

	if PhoneNumberPattern != "[0-9]{10}" {
		t.Errorf("Expected PhoneNumberPattern to be '[0-9]{10}', got '%s'", PhoneNumberPattern)
	}

	// Test prefix constants
	if UsernamePatternPrefix != "^[a-zA-Z0-9_]{3" {
		t.Errorf("Expected UsernamePatternPrefix to be '^[a-zA-Z0-9_]{3', got '%s'", UsernamePatternPrefix)
	}

	// Test rule name constants
	if RuleRequired != "required" {
		t.Errorf("Expected RuleRequired to be 'required', got '%s'", RuleRequired)
	}

	if RuleEmail != "email" {
		t.Errorf("Expected RuleEmail to be 'email', got '%s'", RuleEmail)
	}

	// Test error message constants
	if InvalidRangeMsg != "Invalid range values for %s" {
		t.Errorf("Expected InvalidRangeMsg to be 'Invalid range values for %%s', got '%s'", InvalidRangeMsg)
	}
}

func TestEmailRegex(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"test+tag@example.org",
		"user_name@example-domain.com",
		"123@domain.com",
	}

	invalidEmails := []string{
		"invalid-email",
		"@domain.com",
		"user@",
		"user@domain",
		"",
	}

	for _, email := range validEmails {
		if !EmailRegex.MatchString(email) {
			t.Errorf("Expected '%s' to be valid email", email)
		}
	}

	for _, email := range invalidEmails {
		if EmailRegex.MatchString(email) {
			t.Errorf("Expected '%s' to be invalid email", email)
		}
	}
}

func TestValidationErrorStruct(t *testing.T) {
	err := ValidationError{
		Field:   "username",
		Message: "Field is required",
	}

	if err.Field != "username" {
		t.Errorf("Expected field 'username', got '%s'", err.Field)
	}

	if err.Message != "Field is required" {
		t.Errorf("Expected message 'Field is required', got '%s'", err.Message)
	}
}

func TestRegexpCacheConcurrency(t *testing.T) {
	// Test concurrent access to regexp cache
	const numGoroutines = 10
	const pattern = "^test[0-9]+$"

	done := make(chan bool, numGoroutines)

	// Launch multiple goroutines that compile the same pattern
	for i := 0; i < numGoroutines; i++ {
		go func() {
			regex, err := getCompiledRegexp(pattern)
			if err != nil {
				t.Errorf("Unexpected error in goroutine: %v", err)
			}
			if regex == nil {
				t.Error("Expected regex to be non-nil in goroutine")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestParseValidationRulesEdgeCases(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		validateTag string
		expected    []string
	}{
		{
			name:        "Empty validation tag",
			validateTag: "",
			expected:    []string{},
		},
		{
			name:        "Only commas",
			validateTag: ",,,,",
			expected:    []string{},
		},
		{
			name:        "Single rule with commas",
			validateTag: "required,",
			expected:    []string{"required"},
		},
		{
			name:        "Mixed empty and valid rules",
			validateTag: "required,,email,,min=5",
			expected:    []string{"required", "email", "min=5"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validator.parseValidationRules(test.validateTag)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d rules, got %d", len(test.expected), len(result))
			}
			for i, expected := range test.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected rule %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

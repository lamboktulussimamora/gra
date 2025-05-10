package validator

import (
	"testing"
)

// TestNew ensures the validator creates a new instance correctly
func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
	if v.errors == nil {
		t.Fatal("New() should initialize errors slice")
	}
	if len(v.errors) != 0 {
		t.Fatalf("New validator should have 0 errors, got %d", len(v.errors))
	}
}

// Test structs
type TestUser struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Age      int     `json:"age" validate:"min=18,max=120"`
	Password string  `json:"password" validate:"min=6,max=100"`
	Balance  float64 `json:"balance" validate:"min=0"`
}

type TestNestedStruct struct {
	Title string   `json:"title" validate:"required"`
	User  TestUser `json:"user" validate:"required"`
}

type TestNoValidation struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

// TestRequiredValidation tests the required validation rule
func TestRequiredValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		expectedValid bool
	}{
		{
			name: "Valid User",
			input: TestUser{
				Name:     "John Doe",
				Email:    "john@example.com",
				Age:      30,
				Password: "password123",
			},
			expectedValid: true,
		},
		{
			name: "Missing Name",
			input: TestUser{
				Email:    "john@example.com",
				Age:      30,
				Password: "password123",
			},
			expectedValid: false,
		},
		{
			name: "Missing Email",
			input: TestUser{
				Name:     "John Doe",
				Age:      30,
				Password: "password123",
			},
			expectedValid: false,
		},
		{
			name:          "No Validation",
			input:         TestNoValidation{},
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			errors := v.Validate(tt.input)

			if tt.expectedValid && len(errors) > 0 {
				t.Errorf("Expected valid, but got %d errors: %v", len(errors), errors)
			}

			if !tt.expectedValid && len(errors) == 0 {
				t.Errorf("Expected invalid, but got no errors")
			}
		})
	}
}

// TestRequiredEdgeCases tests edge cases for the required validation rule
func TestRequiredEdgeCases(t *testing.T) {
	type AllTypesStruct struct {
		String  string            `json:"string" validate:"required"`
		Int     int               `json:"int" validate:"required"`
		Int8    int8              `json:"int8" validate:"required"`
		Int16   int16             `json:"int16" validate:"required"`
		Int32   int32             `json:"int32" validate:"required"`
		Int64   int64             `json:"int64" validate:"required"`
		Uint    uint              `json:"uint" validate:"required"`
		Uint8   uint8             `json:"uint8" validate:"required"`
		Uint16  uint16            `json:"uint16" validate:"required"`
		Uint32  uint32            `json:"uint32" validate:"required"`
		Uint64  uint64            `json:"uint64" validate:"required"`
		Float32 float32           `json:"float32" validate:"required"`
		Float64 float64           `json:"float64" validate:"required"`
		Bool    bool              `json:"bool" validate:"required"`
		Ptr     *string           `json:"ptr" validate:"required"`
		Slice   []int             `json:"slice" validate:"required"`
		Map     map[string]string `json:"map" validate:"required"`
	}

	// Create a valid struct with all required fields
	str := "test"
	validStruct := AllTypesStruct{
		String:  "test",
		Int:     1,
		Int8:    1,
		Int16:   1,
		Int32:   1,
		Int64:   1,
		Uint:    1,
		Uint8:   1,
		Uint16:  1,
		Uint32:  1,
		Uint64:  1,
		Float32: 1.0,
		Float64: 1.0,
		Bool:    true,
		Ptr:     &str,
		Slice:   []int{1},
		Map:     map[string]string{"key": "value"},
	}

	v := New()
	errors := v.Validate(validStruct)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid struct, got %d errors: %v", len(errors), errors)
	}

	// Now test each field with an invalid value
	testCases := []struct {
		name      string
		modifyFn  func(*AllTypesStruct)
		fieldName string
	}{
		{
			name:      "Empty String",
			modifyFn:  func(s *AllTypesStruct) { s.String = "" },
			fieldName: "string",
		},
		{
			name:      "Zero Int",
			modifyFn:  func(s *AllTypesStruct) { s.Int = 0 },
			fieldName: "int",
		},
		{
			name:      "Zero Int8",
			modifyFn:  func(s *AllTypesStruct) { s.Int8 = 0 },
			fieldName: "int8",
		},
		{
			name:      "Zero Int16",
			modifyFn:  func(s *AllTypesStruct) { s.Int16 = 0 },
			fieldName: "int16",
		},
		{
			name:      "Zero Int32",
			modifyFn:  func(s *AllTypesStruct) { s.Int32 = 0 },
			fieldName: "int32",
		},
		{
			name:      "Zero Int64",
			modifyFn:  func(s *AllTypesStruct) { s.Int64 = 0 },
			fieldName: "int64",
		},
		{
			name:      "Zero Uint",
			modifyFn:  func(s *AllTypesStruct) { s.Uint = 0 },
			fieldName: "uint",
		},
		{
			name:      "Zero Uint8",
			modifyFn:  func(s *AllTypesStruct) { s.Uint8 = 0 },
			fieldName: "uint8",
		},
		{
			name:      "Zero Uint16",
			modifyFn:  func(s *AllTypesStruct) { s.Uint16 = 0 },
			fieldName: "uint16",
		},
		{
			name:      "Zero Uint32",
			modifyFn:  func(s *AllTypesStruct) { s.Uint32 = 0 },
			fieldName: "uint32",
		},
		{
			name:      "Zero Uint64",
			modifyFn:  func(s *AllTypesStruct) { s.Uint64 = 0 },
			fieldName: "uint64",
		},
		{
			name:      "Zero Float32",
			modifyFn:  func(s *AllTypesStruct) { s.Float32 = 0 },
			fieldName: "float32",
		},
		{
			name:      "Zero Float64",
			modifyFn:  func(s *AllTypesStruct) { s.Float64 = 0 },
			fieldName: "float64",
		},
		{
			name:      "False Bool",
			modifyFn:  func(s *AllTypesStruct) { s.Bool = false },
			fieldName: "bool",
		},
		{
			name:      "Nil Pointer",
			modifyFn:  func(s *AllTypesStruct) { s.Ptr = nil },
			fieldName: "ptr",
		},
		{
			name:      "Nil Slice",
			modifyFn:  func(s *AllTypesStruct) { s.Slice = nil },
			fieldName: "slice",
		},
		{
			name:      "Nil Map",
			modifyFn:  func(s *AllTypesStruct) { s.Map = nil },
			fieldName: "map",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			invalidStruct := validStruct // Create a copy
			tc.modifyFn(&invalidStruct)  // Modify one field

			v := New()
			errors := v.Validate(invalidStruct)

			// Should have exactly 1 error
			if len(errors) != 1 {
				t.Errorf("Expected 1 error for %s, got %d errors: %v", tc.name, len(errors), errors)
				return
			}

			// Error should be for the expected field
			if errors[0].Field != tc.fieldName {
				t.Errorf("Expected error for field '%s', got error for '%s'", tc.fieldName, errors[0].Field)
			}
		})
	}
}

// TestEmailValidation tests the email validation rule
func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"Valid Email", "user@example.com", true},
		{"Valid Email With Dots", "user.name@example.com", true},
		{"Valid Email With Plus", "user+tag@example.com", true},
		{"Valid Email With Dashes", "user-name@example-site.com", true},
		{"Missing @", "userexample.com", false},
		{"Missing Domain", "user@", false},
		{"Missing TLD", "user@example", false},
		{"Invalid Characters", "user@ex ample.com", false},
		{"Empty String", "", true}, // Empty is valid as it's handled by required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			user := TestUser{
				Name:     "Test User",
				Email:    tt.email,
				Age:      30,
				Password: "password123",
			}

			errors := v.Validate(user)

			// If email is empty, it should be caught by 'required' validation
			if tt.email == "" {
				hasEmailError := false
				for _, err := range errors {
					if err.Field == "email" && err.Message == "email is required" {
						hasEmailError = true
					}
				}

				if !hasEmailError {
					t.Errorf("Expected 'required' error for empty email")
				}
				return
			}

			// Check for email validation errors
			hasEmailFormatError := false
			for _, err := range errors {
				if err.Field == "email" && err.Message == "email must be a valid email address" {
					hasEmailFormatError = true
				}
			}

			if tt.valid && hasEmailFormatError {
				t.Errorf("Email %q should be valid, but validation failed", tt.email)
			}

			if !tt.valid && !hasEmailFormatError {
				t.Errorf("Email %q should be invalid, but validation passed", tt.email)
			}
		})
	}
}

// TestMinValidation tests the min validation rule
func TestMinValidation(t *testing.T) {
	tests := []struct {
		name       string
		age        int
		password   string
		balance    float64
		shouldPass bool
	}{
		{"Valid Values", 20, "password", 10.5, true},
		{"Age Below Min", 17, "password", 10.5, false},
		{"Password Below Min", 20, "pass", 10.5, false},
		{"Balance Below Min", 20, "password", -1.5, false},
		{"All Below Min", 17, "pw", -1.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			user := TestUser{
				Name:     "Test User",
				Email:    "test@example.com",
				Age:      tt.age,
				Password: tt.password,
				Balance:  tt.balance,
			}

			errors := v.Validate(user)

			// Check age validation
			hasAgeError := false
			hasPasswordError := false
			hasBalanceError := false

			for _, err := range errors {
				if err.Field == "age" && err.Message == "age must be at least 18" {
					hasAgeError = true
				}
				if err.Field == "password" && err.Message == "password must be at least 6 characters" {
					hasPasswordError = true
				}
				if err.Field == "balance" && err.Message == "balance must be at least 0.000000" {
					hasBalanceError = true
				}
			}

			if tt.age < 18 && !hasAgeError {
				t.Errorf("Expected age validation error for age %d", tt.age)
			}

			if tt.age >= 18 && hasAgeError {
				t.Errorf("Unexpected age validation error for age %d", tt.age)
			}

			if len(tt.password) < 6 && !hasPasswordError {
				t.Errorf("Expected password validation error for password %q", tt.password)
			}

			if len(tt.password) >= 6 && hasPasswordError {
				t.Errorf("Unexpected password validation error for password %q", tt.password)
			}

			if tt.balance < 0 && !hasBalanceError {
				t.Errorf("Expected balance validation error for balance %f", tt.balance)
			}

			if tt.balance >= 0 && hasBalanceError {
				t.Errorf("Unexpected balance validation error for balance %f", tt.balance)
			}
		})
	}
}

// TestMaxValidation tests the max validation rule
func TestMaxValidation(t *testing.T) {
	tests := []struct {
		name       string
		age        int
		password   string
		shouldPass bool
	}{
		{"Valid Values", 100, "password-thats-acceptable", true},
		{"Age Above Max", 121, "password", false},
		{"Password Above Max", 100, string(make([]byte, 101)), false},
		{"All Above Max", 121, string(make([]byte, 101)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			user := TestUser{
				Name:     "Test User",
				Email:    "test@example.com",
				Age:      tt.age,
				Password: tt.password,
			}

			errors := v.Validate(user)

			// Check validation errors
			hasAgeError := false
			hasPasswordError := false

			for _, err := range errors {
				if err.Field == "age" && err.Message == "age must be at most 120" {
					hasAgeError = true
				}
				if err.Field == "password" && err.Message == "password must be at most 100 characters" {
					hasPasswordError = true
				}
			}

			if tt.age > 120 && !hasAgeError {
				t.Errorf("Expected age validation error for age %d", tt.age)
			}

			if tt.age <= 120 && hasAgeError {
				t.Errorf("Unexpected age validation error for age %d", tt.age)
			}

			if len(tt.password) > 100 && !hasPasswordError {
				t.Errorf("Expected password validation error for password length %d", len(tt.password))
			}

			if len(tt.password) <= 100 && hasPasswordError {
				t.Errorf("Unexpected password validation error for password length %d", len(tt.password))
			}
		})
	}
}

// TestMaxValidationEdgeCases tests edge cases for max validation
func TestMaxValidationEdgeCases(t *testing.T) {
	type MaxValidationStruct struct {
		StringField string  `json:"stringField" validate:"max=5"`
		IntField    int     `json:"intField" validate:"max=10"`
		Int8Field   int8    `json:"int8Field" validate:"max=10"`
		Int16Field  int16   `json:"int16Field" validate:"max=10"`
		Int32Field  int32   `json:"int32Field" validate:"max=10"`
		Int64Field  int64   `json:"int64Field" validate:"max=10"`
		UintField   uint    `json:"uintField" validate:"max=10"`
		Uint8Field  uint8   `json:"uint8Field" validate:"max=10"`
		Uint16Field uint16  `json:"uint16Field" validate:"max=10"`
		Uint32Field uint32  `json:"uint32Field" validate:"max=10"`
		Uint64Field uint64  `json:"uint64Field" validate:"max=10"`
		FloatField  float64 `json:"floatField" validate:"max=10.5"`
	}

	// Create a valid struct with all fields at or below max
	validStruct := MaxValidationStruct{
		StringField: "12345", // Exactly 5 chars
		IntField:    10,      // Exactly 10
		Int8Field:   10,
		Int16Field:  10,
		Int32Field:  10,
		Int64Field:  10,
		UintField:   10,
		Uint8Field:  10,
		Uint16Field: 10,
		Uint32Field: 10,
		Uint64Field: 10,
		FloatField:  10.5, // Exactly 10.5
	}

	v := New()
	errors := v.Validate(validStruct)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid struct, got %d errors: %v", len(errors), errors)
	}

	// Now test each field with a value exceeding the max
	testCases := []struct {
		name      string
		modifyFn  func(*MaxValidationStruct)
		fieldName string
	}{
		{
			name:      "String Too Long",
			modifyFn:  func(s *MaxValidationStruct) { s.StringField = "123456" }, // 6 chars
			fieldName: "stringField",
		},
		{
			name:      "Int Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.IntField = 11 },
			fieldName: "intField",
		},
		{
			name:      "Int8 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Int8Field = 11 },
			fieldName: "int8Field",
		},
		{
			name:      "Int16 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Int16Field = 11 },
			fieldName: "int16Field",
		},
		{
			name:      "Int32 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Int32Field = 11 },
			fieldName: "int32Field",
		},
		{
			name:      "Int64 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Int64Field = 11 },
			fieldName: "int64Field",
		},
		{
			name:      "Uint Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.UintField = 11 },
			fieldName: "uintField",
		},
		{
			name:      "Uint8 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Uint8Field = 11 },
			fieldName: "uint8Field",
		},
		{
			name:      "Uint16 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Uint16Field = 11 },
			fieldName: "uint16Field",
		},
		{
			name:      "Uint32 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Uint32Field = 11 },
			fieldName: "uint32Field",
		},
		{
			name:      "Uint64 Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.Uint64Field = 11 },
			fieldName: "uint64Field",
		},
		{
			name:      "Float Too Large",
			modifyFn:  func(s *MaxValidationStruct) { s.FloatField = 10.6 },
			fieldName: "floatField",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			invalidStruct := validStruct // Create a copy
			tc.modifyFn(&invalidStruct)  // Modify one field

			v := New()
			errors := v.Validate(invalidStruct)

			// Should have exactly 1 error
			if len(errors) != 1 {
				t.Errorf("Expected 1 error for %s, got %d errors: %v", tc.name, len(errors), errors)
				return
			}

			// Error should be for the expected field
			if errors[0].Field != tc.fieldName {
				t.Errorf("Expected error for field '%s', got error for '%s'", tc.fieldName, errors[0].Field)
			}
		})
	}
}

// TestNestedStructValidation tests validation on nested structs
func TestNestedStructValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         TestNestedStruct
		expectedValid bool
	}{
		{
			name: "Valid Nested Struct",
			input: TestNestedStruct{
				Title: "Test Title",
				User: TestUser{
					Name:     "John Doe",
					Email:    "john@example.com",
					Age:      30,
					Password: "password123",
				},
			},
			expectedValid: true,
		},
		{
			name: "Missing Title",
			input: TestNestedStruct{
				User: TestUser{
					Name:     "John Doe",
					Email:    "john@example.com",
					Age:      30,
					Password: "password123",
				},
			},
			expectedValid: false,
		},
		{
			name: "Invalid Nested User",
			input: TestNestedStruct{
				Title: "Test Title",
				User: TestUser{
					Name:     "", // Missing required field
					Email:    "john@example.com",
					Age:      30,
					Password: "password123",
				},
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			errors := v.Validate(tt.input)

			if tt.expectedValid && len(errors) > 0 {
				t.Errorf("Expected valid, but got %d errors: %v", len(errors), errors)
			}

			if !tt.expectedValid && len(errors) == 0 {
				t.Errorf("Expected invalid, but got no errors")
			}
		})
	}
}

// TestHasErrors tests the HasErrors function
func TestHasErrors(t *testing.T) {
	v := New()
	if v.HasErrors() {
		t.Error("New validator should not have errors")
	}

	// Add an error
	v.errors = append(v.errors, ValidationError{
		Field:   "test",
		Message: "test error",
	})

	if !v.HasErrors() {
		t.Error("Validator should have errors")
	}
}

// TestArrayOfNestedStructs tests validation of arrays of nested structs
func TestArrayOfNestedStructs(t *testing.T) {
	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		Country string `json:"country" validate:"required"`
		ZipCode string `json:"zipCode" validate:"required"`
	}

	type Product struct {
		ID          int     `json:"id" validate:"min=1"`
		Name        string  `json:"name" validate:"required"`
		Price       float64 `json:"price" validate:"min=0.01"`
		Description string  `json:"description" validate:"max=1000"`
	}

	type Order struct {
		ID                int       `json:"id" validate:"min=1"`
		Products          []Product `json:"products" validate:"required"`
		ShippingAddresses []Address `json:"shippingAddresses" validate:"required"`
	}

	v := New()

	// Valid case - everything correctly populated
	validOrder := Order{
		ID: 101,
		Products: []Product{
			{ID: 1, Name: "Product 1", Price: 19.99, Description: "Description 1"},
			{ID: 2, Name: "Product 2", Price: 29.99, Description: "Description 2"},
		},
		ShippingAddresses: []Address{
			{Street: "123 Main St", City: "New York", Country: "USA", ZipCode: "10001"},
		},
	}

	errors := v.Validate(validOrder)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors for valid order, got %d", len(errors))
	}

	// Invalid case 1 - invalid product in array
	invalidProductOrder := Order{
		ID: 102,
		Products: []Product{
			{ID: 1, Name: "Product 1", Price: 19.99, Description: "Description 1"},
			{ID: 2, Name: "", Price: 29.99, Description: "Description 2"}, // Missing name
		},
		ShippingAddresses: []Address{
			{Street: "123 Main St", City: "New York", Country: "USA", ZipCode: "10001"},
		},
	}

	errors = v.Validate(invalidProductOrder)
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error for invalid product name, got %d", len(errors))
	}

	// Invalid case 2 - invalid address in array
	invalidAddressOrder := Order{
		ID: 103,
		Products: []Product{
			{ID: 1, Name: "Product 1", Price: 19.99, Description: "Description 1"},
		},
		ShippingAddresses: []Address{
			{Street: "123 Main St", City: "", Country: "USA", ZipCode: "10001"}, // Missing city
		},
	}

	errors = v.Validate(invalidAddressOrder)
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error for invalid address city, got %d", len(errors))
	}

	// Invalid case 3 - multiple validation errors
	multipleErrorsOrder := Order{
		ID: 104,
		Products: []Product{
			{ID: 0, Name: "", Price: -1, Description: "Description"}, // Multiple errors
		},
		ShippingAddresses: []Address{
			{Street: "", City: "", Country: "", ZipCode: ""}, // All fields missing
		},
	}

	errors = v.Validate(multipleErrorsOrder)
	if len(errors) < 6 { // At least 6 errors (ID, Name, Price, all 4 address fields)
		t.Errorf("Expected at least 6 validation errors, got %d", len(errors))
	}
}

// TestNonStructValidation tests validation of non-struct types
func TestNonStructValidation(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"String Value", "hello"},
		{"Integer Value", 123},
		{"Float Value", 45.67},
		{"Boolean Value", true},
		{"Nil Value", nil},
		{"Map Value", map[string]string{"key": "value"}},
		{"Array Value", [3]int{1, 2, 3}},
		{"Slice Value", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			errors := v.Validate(tt.input)

			// Non-struct values should not produce validation errors
			if len(errors) > 0 {
				t.Errorf("Expected no errors for non-struct value %v, got %d errors: %v",
					tt.input, len(errors), errors)
			}
		})
	}
}

// TestPointerStructValidation tests validation of pointer to struct
func TestPointerStructValidation(t *testing.T) {
	// Valid case - pointer to struct with all fields valid
	validUser := &TestUser{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Password: "password123",
		Balance:  100.0,
	}

	// Invalid case - pointer to struct with missing required field
	invalidUser := &TestUser{
		Email:    "john@example.com",
		Age:      30,
		Password: "password123",
		Balance:  100.0,
		// Name is missing
	}

	t.Run("Valid Pointer to Struct", func(t *testing.T) {
		v := New()
		errors := v.Validate(validUser)

		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid pointer to struct, got %d errors: %v",
				len(errors), errors)
		}
	})

	t.Run("Invalid Pointer to Struct", func(t *testing.T) {
		v := New()
		errors := v.Validate(invalidUser)

		hasNameError := false
		for _, err := range errors {
			if err.Field == "name" && err.Message == "name is required" {
				hasNameError = true
				break
			}
		}

		if !hasNameError {
			t.Errorf("Expected 'name is required' error for invalid pointer to struct")
		}
	})

	t.Run("Nil Pointer to Struct", func(t *testing.T) {
		v := New()
		var nilUser *TestUser = nil
		errors := v.Validate(nilUser)

		// Nil pointers should not produce validation errors
		if len(errors) > 0 {
			t.Errorf("Expected no errors for nil pointer, got %d errors", len(errors))
		}
	})
}

// TestEmbeddedStructValidation tests validation of structs with embedded fields
func TestEmbeddedStructValidation(t *testing.T) {
	// Create a struct with embedded fields
	type BaseInfo struct {
		ID        int    `json:"id" validate:"min=1"`
		CreatedBy string `json:"createdBy" validate:"required"`
	}

	type ProductWithEmbedded struct {
		BaseInfo          // Embedded struct
		Name     string   `json:"name" validate:"required"`
		Price    float64  `json:"price" validate:"min=0.01"`
		Tags     []string `json:"tags"` // No validation
	}

	t.Run("Valid Product with Embedded Struct", func(t *testing.T) {
		// Valid case
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID:        123,
				CreatedBy: "Admin",
			},
			Name:  "Test Product",
			Price: 29.99,
			Tags:  []string{"test", "product"},
		}

		v := New()
		errors := v.Validate(product)

		if len(errors) > 0 {
			t.Errorf("Expected no errors, got %d errors: %v", len(errors), errors)
		}
	})

	t.Run("Missing Field in Embedded Struct", func(t *testing.T) {
		// Invalid case - missing field in embedded struct
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID: 123,
				// CreatedBy is missing
			},
			Name:  "Test Product",
			Price: 29.99,
			Tags:  []string{"test", "product"},
		}

		v := New()
		errors := v.Validate(product)

		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d errors: %v", len(errors), errors)
		}

		hasCreatedByError := false
		for _, err := range errors {
			if err.Field == "createdBy" {
				hasCreatedByError = true
				break
			}
		}

		if !hasCreatedByError {
			t.Error("Expected validation error for 'createdBy', but none found")
		}
	})

	t.Run("Missing Field in Main Struct", func(t *testing.T) {
		// Invalid case - missing field in main struct
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID:        123,
				CreatedBy: "Admin",
			},
			// Name is missing
			Price: 29.99,
			Tags:  []string{"test", "product"},
		}

		v := New()
		errors := v.Validate(product)

		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d errors: %v", len(errors), errors)
		}

		hasNameError := false
		for _, err := range errors {
			if err.Field == "name" {
				hasNameError = true
				break
			}
		}

		if !hasNameError {
			t.Error("Expected validation error for 'name', but none found")
		}
	})
}

// TestNoJSONTag tests that fields without JSON tags are skipped in validation
func TestNoJSONTag(t *testing.T) {
	type NoTagStruct struct {
		Name  string `validate:"required"`           // No JSON tag
		Age   int    `json:"-" validate:"min=18"`    // JSON ignored
		Email string `json:"email" validate:"email"` // Normal field
	}

	// Test with invalid fields that should be skipped
	invalidStruct := NoTagStruct{
		// Name is missing but has no JSON tag
		// Age is below minimum but has json:"-"
		Age:   16,
		Email: "valid@example.com",
	}

	v := New()
	errors := v.Validate(invalidStruct)

	// Should only validate the email field, which is valid
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}

	// Now let's make the email invalid to confirm it's being validated
	invalidStruct.Email = "not-an-email"
	errors = v.Validate(invalidStruct)

	// Should have 1 error for email
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error for email, got %d", len(errors))
	}

	// Check that the error is for email
	if len(errors) > 0 && errors[0].Field != "email" {
		t.Errorf("Expected error for field 'email', got '%s'", errors[0].Field)
	}
}

// TestSimpleSliceValidation tests validation of slices with non-struct elements
func TestSimpleSliceValidation(t *testing.T) {
	// Currently, the validator does not directly validate array/slice elements
	// unless they are structs. This test confirms current behavior.

	type SliceStruct struct {
		Names []string `json:"names" validate:"required"`
		Ages  []int    `json:"ages" validate:"required"`
	}

	// Valid case with non-empty slices
	validSlices := SliceStruct{
		Names: []string{"Alice", "Bob"},
		Ages:  []int{25, 30},
	}

	// Invalid case with empty slice (the slice itself is present but empty)
	invalidSlices := SliceStruct{
		Names: []string{},
		Ages:  []int{},
	}

	// Test valid slices
	v := New()
	errors := v.Validate(validSlices)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid slices, got %d: %v", len(errors), errors)
	}

	// Test with empty slices
	// Currently, empty slices pass the required check as the slice itself is not nil
	errors = v.Validate(invalidSlices)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for empty slices (current behavior), got %d: %v", len(errors), errors)
	}

	// Test with nil slices - this should fail the required check
	nilSlices := SliceStruct{
		Names: nil,
		Ages:  nil,
	}

	errors = v.Validate(nilSlices)

	// Should have 2 errors, one for each nil slice
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors for nil slices, got %d", len(errors))
	}
}

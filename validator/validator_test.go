package validator

import (
	"strings"
	"testing"
)

const (
	// Field names for tests
	fieldName        = "name"
	fieldEmail       = "email"
	fieldPrice       = "price"
	fieldUserID      = "userID"
	fieldPhone       = "phone"
	fieldAge         = "age"
	fieldPassword    = "password"
	fieldBalance     = "balance"
	fieldPostalCode  = "postalCode"
	fieldUsername    = "username"
	fieldDescription = "description"
	fieldCity        = "city"
	fieldCountry     = "country"
	fieldZipCode     = "zipCode"
	fieldCode        = "code"

	// Test values
	testUsername    = "testuser"
	testEmail       = "test@example.com"
	testUser        = "Test User"
	testPassword    = "password123"
	testAddress     = "123 Main St"
	testPostalCode  = "A1B 2C3"
	testName        = "John Doe"
	testUserEmail   = "john@example.com"
	testCity        = "New York"
	testCountry     = "USA"
	testZipCode     = "10001"
	testValidCode   = "ABC"
	testInvalidCode = "123"
	testPhoneNumber = "+1234567890"

	// Test product values
	testProductName = "Product 1"
	testDescription = "Description 1"
	testProduct2    = "Product 2"
	testDesc2       = "Description 2"

	// Test price values
	testPrice1 = 19.99
	testPrice2 = 29.99

	// Field type constants
	fieldTypeString = "string"

	// Common test validation messages
	msgValidationPass     = "Expected valid %s to pass validation, got errors: %v"
	msgValidationFail     = "Expected '%s' to fail validation, but got no errors"
	msgNoError            = "Expected no errors, got %d errors: %v"
	msgInvalidField       = "Expected error for field '%s', got error for '%s'"
	msgFieldNoError       = "Expected error for field %s, but none was reported"
	msgErrorCount         = "Expected %d errors, got %d: %v"
	msgFieldNotFoundError = "Expected error for field '%s', but not found in errors: %v"

	// Additional validation test messages
	msgNoErrorsValidRange  = "Expected no errors for valid range, got: %v"
	msgErrorOutsideRange   = "Expected error for value outside range"
	msgErrorInvalidEnum    = "Expected error for invalid enum value"
	msgExpectedRequiredErr = "Expected 'is required' error message"
	msgExpectedMinLenErr   = "Expected minimum length error"
	msgExpectedMaxLenErr   = "Expected maximum length error"
	msgExpectedEnumErr     = "Expected enum error message"
	msgExpectedRangeErr    = "Expected range error message, got: %s"
	msgNoErrorsEdgeValues  = "Expected no errors for edge values, got: %v"
)

// TestNew ensures the validator creates a new instance correctly
func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
		return
	}
	if v.errors == nil {
		t.Fatal("New() should initialize errors slice")
		return
	}
	if len(v.errors) != 0 {
		t.Fatalf("New validator should have 0 errors, got %d", len(v.errors))
		return
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
		input         any
		expectedValid bool
	}{
		{
			name: "Valid User",
			input: TestUser{
				Name:     testName,
				Email:    testUserEmail,
				Age:      30,
				Password: "password123",
			},
			expectedValid: true,
		},
		{
			name: "Missing Name",
			input: TestUser{
				Email:    testUserEmail,
				Age:      30,
				Password: "password123",
			},
			expectedValid: false,
		},
		{
			name: "Missing Email",
			input: TestUser{
				Name:     testName,
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
				t.Errorf(msgInvalidField, tc.fieldName, errors[0].Field)
			}
		})
	}
}

// Constants for validation error messages
const (
	msgEmailFormatError = "email must be a valid email address"
	msgEmailRequired    = "email is required"
)

// checkEmailValidationError finds email validation errors in the error list
func checkEmailValidationError(errors []ValidationError, errorType string) bool {
	for _, err := range errors {
		if err.Field == fieldEmail && err.Message == errorType {
			return true
		}
	}
	return false
}

// validateEmailResult checks if the validation result matches expectations
func validateEmailResult(t *testing.T, email string, isValid bool, errors []ValidationError) {
	// For empty email, just check required error
	if email == "" {
		hasRequiredError := checkEmailValidationError(errors, msgEmailRequired)
		if !hasRequiredError {
			t.Errorf("Expected 'required' error for empty email")
		}
		return
	}

	// Check for format errors
	hasFormatError := checkEmailValidationError(errors, msgEmailFormatError)

	// Validate based on expected validity
	if isValid && hasFormatError {
		t.Errorf("Email %q should be valid, but validation failed", email)
	}

	if !isValid && !hasFormatError {
		t.Errorf("Email %q should be invalid, but validation passed", email)
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
				Name:     testUser,
				Email:    tt.email,
				Age:      30,
				Password: testPassword,
			}

			errors := v.Validate(user)
			validateEmailResult(t, tt.email, tt.valid, errors)
		})
	}
}

// findValidationError searches for a specific validation error
func findValidationError(errors []ValidationError, field, expectedMessage string) bool {
	for _, err := range errors {
		if err.Field == field && err.Message == expectedMessage {
			return true
		}
	}
	return false
}

// checkMinValidation performs validation checks for min rule tests
func checkMinValidation(t *testing.T, errors []ValidationError, field string, value interface{}, minValue interface{}, expectError bool) {
	var errorMsg string
	var hasError bool

	switch field {
	case fieldAge:
		errorMsg = "age must be at least 18"
	case fieldPassword:
		errorMsg = "password must be at least 6 characters"
	case fieldBalance:
		errorMsg = "balance must be at least 0"
		// Accept both 0 and 0.000000 for float formatting
		altErrorMsg := "balance must be at least 0.000000"
		hasError = findValidationError(errors, field, errorMsg) || findValidationError(errors, field, altErrorMsg)
		if expectError && !hasError {
			t.Errorf("Expected %s validation error for %v < %v", field, value, minValue)
		}
		if !expectError && hasError {
			t.Errorf("Unexpected %s validation error for %v >= %v", field, value, minValue)
		}
		return
	}

	hasError = findValidationError(errors, field, errorMsg)

	if expectError && !hasError {
		t.Errorf("Expected %s validation error for %v < %v", field, value, minValue)
	}

	if !expectError && hasError {
		t.Errorf("Unexpected %s validation error for %v >= %v", field, value, minValue)
	}
}

// TestMinValidation tests the min validation rule
func TestMinValidation(t *testing.T) {
	const minAge = 18
	const minPasswordLen = 6
	const minBalance = 0.0

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
				Name:     testUser,
				Email:    testEmail,
				Age:      tt.age,
				Password: tt.password,
				Balance:  tt.balance,
			}

			errors := v.Validate(user)

			// Test each validation field
			checkMinValidation(t, errors, fieldAge, tt.age, minAge, tt.age < minAge)
			checkMinValidation(t, errors, fieldPassword, tt.password, minPasswordLen, len(tt.password) < minPasswordLen)
			checkMinValidation(t, errors, fieldBalance, tt.balance, minBalance, tt.balance < minBalance)
		})
	}
}

// Constants for max validation messages and values
const (
	msgMaxAgeError      = "age must be at most 120"
	msgMaxPasswordError = "password must be at most 100 characters"
	maxAge              = 120
	maxPasswordLength   = 100
)

// checkMaxValidationError finds a specific max validation error in the error list
func checkMaxValidationError(errors []ValidationError, field, message string) bool {
	for _, err := range errors {
		if err.Field == field && err.Message == message {
			return true
		}
	}
	return false
}

// validateMaxAgeValue checks if the age validation result is as expected
func validateMaxAgeValue(t *testing.T, age int, hasAgeError bool) {
	if age > maxAge && !hasAgeError {
		t.Errorf("Expected age validation error for age %d", age)
	}

	if age <= maxAge && hasAgeError {
		t.Errorf("Unexpected age validation error for age %d", age)
	}
}

// validateMaxPasswordValue checks if the password validation result is as expected
func validateMaxPasswordValue(t *testing.T, password string, hasPasswordError bool) {
	if len(password) > maxPasswordLength && !hasPasswordError {
		t.Errorf("Expected password validation error for password length %d", len(password))
	}

	if len(password) <= maxPasswordLength && hasPasswordError {
		t.Errorf("Unexpected password validation error for password length %d", len(password))
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
				Name:     testName,
				Email:    testEmail,
				Age:      tt.age,
				Password: tt.password,
			}

			errors := v.Validate(user)

			// Check validation errors
			hasAgeError := checkMaxValidationError(errors, fieldAge, msgMaxAgeError)
			hasPasswordError := checkMaxValidationError(errors, fieldPassword, msgMaxPasswordError)

			// Validate the results
			validateMaxAgeValue(t, tt.age, hasAgeError)
			validateMaxPasswordValue(t, tt.password, hasPasswordError)
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
				t.Errorf(msgInvalidField, tc.fieldName, errors[0].Field)
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
					Name:     testName,
					Email:    testUserEmail,
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
					Name:     testName,
					Email:    testUserEmail,
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
					Email:    testUserEmail,
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
			{ID: 1, Name: testProductName, Price: testPrice1, Description: testDescription},
			{ID: 2, Name: testProduct2, Price: testPrice2, Description: testDesc2},
		},
		ShippingAddresses: []Address{
			{Street: testAddress, City: testCity, Country: testCountry, ZipCode: testZipCode},
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
			{ID: 1, Name: testProductName, Price: testPrice1, Description: testDescription},
			{ID: 2, Name: "", Price: testPrice2, Description: testDesc2}, // Missing name
		},
		ShippingAddresses: []Address{
			{Street: testAddress, City: testCity, Country: testCountry, ZipCode: testZipCode},
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
			{ID: 1, Name: testProductName, Price: testPrice1, Description: testDescription},
		},
		ShippingAddresses: []Address{
			{Street: testAddress, City: "", Country: testCountry, ZipCode: testZipCode}, // Missing city
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
		input any
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

// validateRegexpFields checks if the expected invalid fields have validation errors
func validateRegexpFields(t *testing.T, errors []ValidationError, invalidFields []string) {
	errorFields := make(map[string]bool)
	for _, err := range errors {
		errorFields[err.Field] = true
	}

	for _, field := range invalidFields {
		if !errorFields[field] {
			t.Errorf("Expected error for field %s, but none was reported", field)
		}
	}
}

// checkFieldHasError checks if a field has validation errors
func checkFieldHasError(errors []ValidationError, field string) bool {
	for _, err := range errors {
		if err.Field == field {
			return true
		}
	}
	return false
}

// validateSimpleUser performs validation tests on a simple User struct
func validateSimpleUser(t *testing.T) {
	type User struct {
		Username string `json:"username" validate:"regexp=[a-z0-9_]{3,16}"`
		Phone    string `json:"phone" validate:"regexp=[0-9]{10}"`
	}

	v := New()

	// Test cases for User struct validation
	testCases := []struct {
		name         string
		user         User
		shouldPass   bool
		invalidField string
	}{
		{
			name: "Valid User",
			user: User{
				Username: "valid_user123",
				Phone:    "1234567890",
			},
			shouldPass: true,
		},
		{
			name: "Invalid Username",
			user: User{
				Username: "Invalid@User", // Contains invalid characters
				Phone:    "1234567890",
			},
			shouldPass:   false,
			invalidField: fieldUsername,
		},
		{
			name: "Invalid Phone",
			user: User{
				Username: "valid_user",
				Phone:    "123", // Too short
			},
			shouldPass:   false,
			invalidField: fieldPhone,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := v.Validate(tc.user)

			if tc.shouldPass {
				if len(errors) > 0 {
					t.Errorf(msgValidationPass, tc.name, errors)
				}
				return
			}

			// Shouldn't pass - check for errors
			if len(errors) == 0 {
				t.Errorf(msgValidationFail, tc.invalidField)
				return
			}

			if !checkFieldHasError(errors, tc.invalidField) {
				t.Errorf(msgInvalidField, tc.invalidField, errors[0].Field)
			}
		})
	}
}

// runRegexpValidationTest runs the validation test for the given input and expected outcomes
func runRegexpValidationTest(t *testing.T, validator *Validator, input interface{}, expectedValid bool, invalidFields []string) {
	errors := validator.Validate(input)

	if expectedValid {
		if len(errors) > 0 {
			t.Errorf(msgValidationPass, "input", errors)
		}
		return
	}

	// Test should fail validation
	if len(errors) == 0 {
		t.Error("Expected validation errors, but got none")
		return
	}

	// Check that all expected invalid fields produced errors
	validateRegexpFields(t, errors, invalidFields)
}

// TestRegexpValidation tests the regexp validation rule
func TestRegexpValidation(t *testing.T) {
	// Define the test struct outside the loop to reduce complexity
	type RegexpTest struct {
		PhoneNumber string `json:"phoneNumber" validate:"regexp=^[0-9]{10}$"`
		PostalCode  string `json:"postalCode" validate:"regexp=^[A-Z][0-9][A-Z] [0-9][A-Z][0-9]$"`
		Username    string `json:"username" validate:"regexp=^[a-zA-Z0-9_]{3,20}$"`
	}

	// Create a reusable validator
	v := New()

	// Test case 1: All valid fields
	t.Run("All Valid", func(t *testing.T) {
		validInput := RegexpTest{
			PhoneNumber: "1234567890",
			PostalCode:  testPostalCode,
			Username:    "user_123",
		}
		runRegexpValidationTest(t, v, validInput, true, nil)
	})

	// Test case 2: Invalid phone number
	t.Run("Invalid Phone", func(t *testing.T) {
		invalidPhone := RegexpTest{
			PhoneNumber: "123456", // Too short
			PostalCode:  testPostalCode,
			Username:    "user_123",
		}
		runRegexpValidationTest(t, v, invalidPhone, false, []string{"phoneNumber"})
	})

	// Test case 3: Invalid postal code
	t.Run("Invalid Postal Code", func(t *testing.T) {
		invalidPostal := RegexpTest{
			PhoneNumber: "1234567890",
			PostalCode:  "123 456", // Wrong format
			Username:    "user_123",
		}
		runRegexpValidationTest(t, v, invalidPostal, false, []string{fieldPostalCode})
	})

	// Test case 4: Invalid username
	t.Run("Invalid Username", func(t *testing.T) {
		invalidUsername := RegexpTest{
			PhoneNumber: "1234567890",
			PostalCode:  testPostalCode,
			Username:    "user@123", // Invalid character
		}
		runRegexpValidationTest(t, v, invalidUsername, false, []string{fieldUsername})
	})

	// Test case 5: Multiple invalid fields
	t.Run("Multiple Invalid Fields", func(t *testing.T) {
		multipleInvalid := RegexpTest{
			PhoneNumber: "12345", // Too short
			PostalCode:  "123",   // Wrong format
			Username:    "u",     // Too short
		}
		runRegexpValidationTest(t, v, multipleInvalid, false, []string{"phoneNumber", "postalCode", "username"})
	})

	// Run separate test for simple user struct
	t.Run("Simple User Validation", func(t *testing.T) {
		validateSimpleUser(t)
	})
}

// TestSimpleRegexpValidation tests the regexp validation rule
func TestSimpleRegexpValidation(t *testing.T) {
	const (
		validCode   = "ABC"
		invalidCode = "123"
	)

	// Create a struct with a single field using regexp validation
	type SimpleRegexp struct {
		Code string `json:"code" validate:"regexp=^[A-Z]{3}$"` // Exactly 3 uppercase letters
	}

	v := New()

	// Test valid value
	valid := SimpleRegexp{Code: validCode}
	if errs := v.Validate(valid); len(errs) > 0 {
		t.Errorf("Expected valid code '%s' to pass validation, got errors: %v", validCode, errs)
	}

	// Test invalid value
	invalid := SimpleRegexp{Code: invalidCode}
	if errs := v.Validate(invalid); len(errs) == 0 {
		t.Errorf("Expected '%s' to fail validation, but got no errors", invalidCode)
	}
}

// TestBatchValidation tests the batch validation functionality
func TestBatchValidation(t *testing.T) {
	type Item struct {
		Name  string `json:"name" validate:"required"`
		Price int    `json:"price" validate:"min=1"`
	}

	items := []any{
		Item{Name: "Valid Item", Price: 100},
		Item{Name: "", Price: 200},        // Missing name
		Item{Name: "Low Price", Price: 0}, // Price below minimum
		Item{Name: "", Price: 0},          // Multiple errors
	}

	v := New()
	results := v.ValidateBatch(items)

	// Check if we have the correct number of results
	if len(results) != len(items) {
		t.Errorf("Expected %d results, got %d", len(items), len(results))
	}

	// First item should be valid
	if len(results[0].Errors) != 0 {
		t.Errorf("Expected item 0 to be valid, but got %d errors", len(results[0].Errors))
	}

	// Second item should have one error (name)
	if len(results[1].Errors) != 1 {
		t.Errorf("Expected item 1 to have 1 error, got %d", len(results[1].Errors))
	} else if results[1].Errors[0].Field != fieldName {
		t.Errorf("Expected error on '%s' field, got '%s'", fieldName, results[1].Errors[0].Field)
	}

	// Third item should have one error (price)
	if len(results[2].Errors) != 1 {
		t.Errorf("Expected item 2 to have 1 error, got %d", len(results[2].Errors))
	} else if results[2].Errors[0].Field != fieldPrice {
		t.Errorf("Expected error on '%s' field, got '%s'", fieldPrice, results[2].Errors[0].Field)
	}

	// Fourth item should have two errors (name and price)
	if len(results[3].Errors) != 2 {
		t.Errorf("Expected item 3 to have 2 errors, got %d", len(results[3].Errors))
	}

	// Test HasBatchErrors
	if !v.HasBatchErrors(results) {
		t.Error("Expected HasBatchErrors to return true")
	}

	// Test FilterInvalid
	invalid := v.FilterInvalid(results)
	if len(invalid) != 3 {
		t.Errorf("Expected 3 invalid results, got %d", len(invalid))
	}
}

// TestPointerStructValidation tests validation of pointer to struct
func TestPointerStructValidation(t *testing.T) {
	// Valid case - pointer to struct with all fields valid
	validUser := &TestUser{
		Name:     testName,
		Email:    testUserEmail,
		Age:      30,
		Password: "password123",
		Balance:  100.0,
	}

	// Invalid case - pointer to struct with missing required field
	invalidUser := &TestUser{
		Email:    testUserEmail,
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
		var nilUser *TestUser
		errors := v.Validate(nilUser)

		// Nil pointers should not produce validation errors
		if len(errors) > 0 {
			t.Errorf("Expected no errors for nil pointer, got %d errors", len(errors))
		}
	})
}

// checkErrorCount validates that the expected number of validation errors occurred
func checkErrorCount(t *testing.T, errors []ValidationError, expectedCount int) {
	t.Helper()
	if len(errors) != expectedCount {
		t.Errorf(msgErrorCount, expectedCount, len(errors), errors)
	}
}

// checkFieldHasValidationError checks if a specific field has a validation error
func checkFieldHasValidationError(t *testing.T, errors []ValidationError, fieldName string) {
	t.Helper()
	if !checkFieldHasError(errors, fieldName) {
		t.Errorf(msgFieldNoError, fieldName)
	}
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

	// Create a validator once for all subtests
	v := New()

	t.Run("Valid Product with Embedded Struct", func(t *testing.T) {
		// Valid case
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID:        123,
				CreatedBy: "Admin",
			},
			Name:  "Test Product",
			Price: testPrice2,
			Tags:  []string{"test", "product"},
		}

		errors := v.Validate(product)
		checkErrorCount(t, errors, 0)
	})

	t.Run("Missing Field in Embedded Struct", func(t *testing.T) {
		// Invalid case - missing field in embedded struct
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID: 123,
				// CreatedBy is missing
			},
			Name:  "Test Product",
			Price: testPrice2,
			Tags:  []string{"test", "product"},
		}

		errors := v.Validate(product)
		checkErrorCount(t, errors, 1)
		checkFieldHasValidationError(t, errors, "createdBy")
	})

	t.Run("Missing Field in Main Struct", func(t *testing.T) {
		// Invalid case - missing field in main struct
		product := ProductWithEmbedded{
			BaseInfo: BaseInfo{
				ID:        123,
				CreatedBy: "Admin",
			},
			// Name is missing
			Price: testPrice2,
			Tags:  []string{"test", "product"},
		}

		errors := v.Validate(product)
		checkErrorCount(t, errors, 1)
		checkFieldHasValidationError(t, errors, "name")
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

// TestSchemaFunctionsCoverage tests schema-related functions that had 0% coverage
func TestSchemaFunctionsCoverage(t *testing.T) {
	// Test NewSchema function
	t.Run("NewSchema", func(t *testing.T) {
		schema := NewSchema()
		if schema == nil {
			t.Error("NewSchema should return a non-nil schema")
			return
		}
		if schema.Fields == nil {
			t.Error("NewSchema should initialize Fields map")
		}
	})

	// Test AddField function
	t.Run("AddField", func(t *testing.T) {
		schema := NewSchema()
		field := SchemaField{
			Type:     "string",
			Required: true,
		}

		result := schema.AddField("testField", field)
		if result != schema {
			t.Error("AddField should return the same schema instance")
		}

		if storedField, exists := schema.Fields["testField"]; !exists {
			t.Error("AddField should add the field to the schema")
		} else if storedField.Type != fieldTypeString || !storedField.Required {
			t.Error("AddField should store the field correctly")
		}
	})

	// Test Schema.Validate function
	t.Run("Schema_Validate", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("name", SchemaField{
			Type:      "string",
			Required:  true,
			MinLength: 2,
		})
		schema.AddField("age", SchemaField{
			Type:     "number",
			Required: false,
			Min:      0,
		})

		// Test valid data
		validData := map[string]any{
			"name": "John",
			"age":  25,
		}
		errors := schema.Validate(validData)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid data, got: %v", errors)
		}

		// Test missing required field
		invalidData := map[string]any{
			"age": 25,
		}
		errors = schema.Validate(invalidData)
		if len(errors) == 0 {
			t.Error("Expected error for missing required field")
		}
	})
}

// TestParsingFunctionsCoverage tests parsing functions that had 0% coverage
func TestParsingFunctionsCoverage(t *testing.T) {
	// Test enum validation which will trigger internal parsing
	t.Run("enum_validation_parsing", func(t *testing.T) {
		type TestStruct struct {
			Status string `json:"status" validate:"enum=active,inactive,pending"`
		}

		v := New()
		// Valid enum value
		valid := TestStruct{Status: "active"}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid enum value, got: %v", errors)
		}

		// Invalid enum value
		invalid := TestStruct{Status: "unknown"}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for invalid enum value")
		}
	})

	// Test min/max validation which uses parsing internally
	t.Run("min_max_validation_parsing", func(t *testing.T) {
		type TestStruct struct {
			Age int `json:"age" validate:"min=18,max=65"`
		}

		v := New()
		// Valid value
		valid := TestStruct{Age: 30}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid age, got: %v", errors)
		}

		// Invalid value (too low)
		invalid := TestStruct{Age: 15}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for age below minimum")
		}
	})

	// Test float validation parsing
	t.Run("float_validation_parsing", func(t *testing.T) {
		type TestStruct struct {
			Price float64 `json:"price" validate:"min=0.01"`
		}

		v := New()
		// Valid value
		valid := TestStruct{Price: 15.99}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid price, got: %v", errors)
		}

		// Invalid value (too low)
		invalid := TestStruct{Price: -1.0}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for negative price")
		}
	})
}

// TestValidationRulesParsing tests parsing functions that had low coverage
func TestValidationRulesParsing(t *testing.T) {
	t.Run("parseRulesBeforeRegexp", func(t *testing.T) {
		type TestStruct struct {
			TestField string `json:"testField" validate:"required,min=3,regexp=^[a-z]+$,max=10"`
		}

		v := New()
		// This should trigger parseRulesBeforeRegexp internally
		invalid := TestStruct{TestField: "AB"} // uppercase, fails regexp
		errors := v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected validation errors")
		}
	})

	t.Run("parseRegexpAndRemainingRules", func(t *testing.T) {
		type TestStruct struct {
			TestField string `json:"testField" validate:"regexp=^[a-z]+$,min=3,max=10"`
		}

		v := New()
		// This should trigger parseRegexpAndRemainingRules internally
		invalid := TestStruct{TestField: "ABC"} // uppercase, fails regexp
		errors := v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected validation errors")
		}
	})
}

// TestSchemaValidationFunctions tests schema validation functions with 0% coverage
func TestSchemaValidationFunctions(t *testing.T) {
	t.Run("handleRequiredField", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("requiredField", SchemaField{
			Type:     "string",
			Required: true,
		})

		// Test missing required field
		data := map[string]any{}
		errors := schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for missing required field")
		}

		// Check error message
		found := false
		for _, err := range errors {
			if err.Field == "requiredField" && strings.Contains(err.Message, "is required") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'is required' error message")
		}
	})

	t.Run("processFieldValidation", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("stringField", SchemaField{
			Type:      "string",
			MinLength: 5,
		})
		schema.AddField("numberField", SchemaField{
			Type: "number",
			Min:  10,
		})

		// Test field validation
		data := map[string]any{
			"stringField": "abc", // too short
			"numberField": 5,     // too small
		}
		errors := schema.Validate(data)
		if len(errors) != 2 {
			t.Errorf("Expected 2 errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("validateArray", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("arrayField", SchemaField{
			Type: "array",
		})

		// Test array validation
		data := map[string]any{
			"arrayField": []any{"item1", "item2"},
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid array, got: %v", errors)
		}

		// Test non-array value
		data = map[string]any{
			"arrayField": "not an array",
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for non-array value")
		}
	})

	t.Run("validateType", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("stringField", SchemaField{Type: "string"})
		schema.AddField("numberField", SchemaField{Type: "number"})
		schema.AddField("boolField", SchemaField{Type: "boolean"})
		schema.AddField("objectField", SchemaField{Type: "object"})
		schema.AddField("arrayField", SchemaField{Type: "array"})

		// Test correct types
		validData := map[string]any{
			"stringField": "hello",
			"numberField": 42,
			"boolField":   true,
			"objectField": map[string]any{"key": "value"},
			"arrayField":  []any{"item"},
		}
		errors := schema.Validate(validData)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for correct types, got: %v", errors)
		}

		// Test incorrect types
		invalidData := map[string]any{
			"stringField": 123,
			"numberField": "not a number",
			"boolField":   "not a bool",
		}
		errors = schema.Validate(invalidData)
		if len(errors) != 3 {
			t.Errorf("Expected 3 type errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("validateString", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("testString", SchemaField{
			Type:      "string",
			MinLength: 3,
			MaxLength: 10,
			Pattern:   "^[a-z]+$",
			Enum:      []string{"hello", "world", "test"},
		})

		// Test valid string
		data := map[string]any{
			"testString": "hello",
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid string, got: %v", errors)
		}

		// Test invalid length
		data = map[string]any{
			"testString": "hi", // too short
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for string too short")
		}

		// Test invalid pattern
		data = map[string]any{
			"testString": "HELLO", // uppercase, fails pattern
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for invalid pattern")
		}

		// Test invalid enum
		data = map[string]any{
			"testString": "invalid", // not in enum
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for invalid enum value")
		}
	})

	t.Run("validateStringLength", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("testString", SchemaField{
			Type:      "string",
			MinLength: 5,
			MaxLength: 10,
		})

		// Test string within length bounds
		data := map[string]any{
			"testString": "hello",
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid length, got: %v", errors)
		}

		// Test string too short
		data = map[string]any{
			"testString": "hi",
		}
		errors = schema.Validate(data)
		if len(errors) == 0 || !strings.Contains(errors[0].Message, "at least 5 characters") {
			t.Error("Expected minimum length error")
		}

		// Test string too long
		data = map[string]any{
			"testString": "this is too long",
		}
		errors = schema.Validate(data)
		if len(errors) == 0 || !strings.Contains(errors[0].Message, "at most 10 characters") {
			t.Error("Expected maximum length error")
		}
	})

	t.Run("validateStringPattern", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("testString", SchemaField{
			Type:    "string",
			Pattern: "^[a-z]{3,5}$",
		})

		// Test valid pattern
		data := map[string]any{
			"testString": "abc",
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid pattern, got: %v", errors)
		}

		// Test invalid pattern
		data = map[string]any{
			"testString": "ABC123",
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for invalid pattern")
		}
	})

	t.Run("validateStringEnum", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("status", SchemaField{
			Type: "string",
			Enum: []string{"active", "inactive", "pending"},
		})

		// Test valid enum value
		data := map[string]any{
			"status": "active",
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid enum, got: %v", errors)
		}

		// Test invalid enum value
		data = map[string]any{
			"status": "unknown",
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for invalid enum value")
		}
	})

	t.Run("validateNumber", func(t *testing.T) {
		schema := NewSchema()
		schema.AddField("testNumber", SchemaField{
			Type: "number",
			Min:  10,
			Max:  100,
		})

		// Test valid number
		data := map[string]any{
			"testNumber": 50,
		}
		errors := schema.Validate(data)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid number, got: %v", errors)
		}

		// Test number below minimum
		data = map[string]any{
			"testNumber": 5,
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for number below minimum")
		}

		// Test number above maximum
		data = map[string]any{
			"testNumber": 150,
		}
		errors = schema.Validate(data)
		if len(errors) == 0 {
			t.Error("Expected error for number above maximum")
		}
	})
}

// TestEnumValidationCoverage tests enum validation with 0% coverage
func TestEnumValidationCoverage(t *testing.T) {
	type TestStruct struct {
		Status string `json:"status" validate:"enum=active,inactive,pending"`
		Role   string `json:"role" validate:"enum=admin,user,guest"`
	}

	v := New()

	// Test valid enum values
	valid := TestStruct{Status: "active", Role: "admin"}
	errors := v.Validate(valid)
	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid enum values, got: %v", errors)
	}

	// Test invalid enum value
	invalid := TestStruct{Status: "unknown", Role: "admin"}
	errors = v.Validate(invalid)
	if len(errors) == 0 {
		t.Error(msgErrorInvalidEnum)
	}

	// Test empty enum value (should pass if not required)
	empty := TestStruct{Status: "", Role: "admin"}
	errors = v.Validate(empty)
	if len(errors) > 0 {
		t.Errorf("Expected no errors for empty enum value, got: %v", errors)
	}
}

// TestValidationFunctionsCoverage tests various validation functions with 0% coverage
func TestValidationFunctionsCoverage(t *testing.T) {
	// Test enum validation coverage
	t.Run("validateEnum", func(t *testing.T) {
		type TestStruct struct {
			Status string `json:"status" validate:"enum=active,inactive,pending"`
		}

		v := New()

		// Valid enum value
		valid := TestStruct{Status: "active"}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid enum value, got: %v", errors)
		}

		// Invalid enum value
		invalid := TestStruct{Status: "unknown"}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for invalid enum value")
		}

		// Empty enum value (should pass if not required)
		empty := TestStruct{Status: ""}
		errors = v.Validate(empty)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for empty enum value, got: %v", errors)
		}
	})

	// Test min/max validation for different types
	t.Run("validateMinMax", func(t *testing.T) {
		type TestStruct struct {
			Age   int     `json:"age" validate:"min=18,max=65"`
			Price float64 `json:"price" validate:"min=0.01,max=999.99"`
			Count uint    `json:"count" validate:"min=1,max=100"`
		}

		v := New()

		// Valid values
		valid := TestStruct{Age: 30, Price: 50.00, Count: 10}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid values, got: %v", errors)
		}

		// Invalid age (below min)
		invalidAge := TestStruct{Age: 15, Price: 50.00, Count: 10}
		errors = v.Validate(invalidAge)
		if len(errors) == 0 {
			t.Error("Expected error for age below minimum")
		}

		// Invalid price (above max)
		invalidPrice := TestStruct{Age: 30, Price: 1000.00, Count: 10}
		errors = v.Validate(invalidPrice)
		if len(errors) == 0 {
			t.Error("Expected error for price above maximum")
		}
	})

	// Test regexp validation coverage
	t.Run("validateRegexp", func(t *testing.T) {
		type TestStruct struct {
			Code string `json:"code" validate:"regexp=^[A-Z]{3}$"`
		}

		v := New()

		// Valid pattern
		valid := TestStruct{Code: "ABC"}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid pattern, got: %v", errors)
		}

		// Invalid pattern
		invalid := TestStruct{Code: "123"}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for invalid pattern")
		}
	})

	// Test email validation coverage
	t.Run("validateEmail", func(t *testing.T) {
		type TestStruct struct {
			Email string `json:"email" validate:"email"`
		}

		v := New()

		// Valid email
		valid := TestStruct{Email: "test@example.com"}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid email, got: %v", errors)
		}

		// Invalid email
		invalid := TestStruct{Email: "not-an-email"}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for invalid email")
		}
	})
}

// TestRangeValidationFunctions tests the range validation functions with 0% coverage
func TestRangeValidationFunctions(t *testing.T) {
	// Test int range validation
	t.Run("validateIntRange", func(t *testing.T) {
		type TestStruct struct {
			Age    int   `json:"age" validate:"range=18,65"`
			Score  int32 `json:"score" validate:"range=-100,100"`
			Points int64 `json:"points" validate:"range=0,1000"`
		}

		v := New()

		// Valid values within range
		valid := TestStruct{Age: 25, Score: 50, Points: 500}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsValidRange, errors)
		}

		// Age below minimum
		belowMin := TestStruct{Age: 17, Score: 50, Points: 500}
		errors = v.Validate(belowMin)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}
		if !strings.Contains(errors[0].Message, "must be between 18 and 65") {
			t.Errorf(msgExpectedRangeErr, errors[0].Message)
		}

		// Age above maximum
		aboveMax := TestStruct{Age: 66, Score: 50, Points: 500}
		errors = v.Validate(aboveMax)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Score below minimum (negative range)
		negativeRange := TestStruct{Age: 25, Score: -101, Points: 500}
		errors = v.Validate(negativeRange)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Points above maximum
		largePoints := TestStruct{Age: 25, Score: 50, Points: 1001}
		errors = v.Validate(largePoints)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Test edge values (exactly at boundaries)
		edgeValues := TestStruct{Age: 18, Score: -100, Points: 1000}
		errors = v.Validate(edgeValues)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsEdgeValues, errors)
		}
	})

	// Test uint range validation
	t.Run("validateUintRange", func(t *testing.T) {
		type TestStruct struct {
			Count    uint   `json:"count" validate:"range=1,100"`
			Quantity uint32 `json:"quantity" validate:"range=0,1000"`
			Size     uint64 `json:"size" validate:"range=10,500"`
		}

		v := New()

		// Valid values within range
		valid := TestStruct{Count: 50, Quantity: 250, Size: 100}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsValidRange, errors)
		}

		// Count below minimum
		belowMin := TestStruct{Count: 0, Quantity: 250, Size: 100}
		errors = v.Validate(belowMin)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}
		if !strings.Contains(errors[0].Message, "must be between 1 and 100") {
			t.Errorf(msgExpectedRangeErr, errors[0].Message)
		}

		// Quantity above maximum
		aboveMax := TestStruct{Count: 50, Quantity: 1001, Size: 100}
		errors = v.Validate(aboveMax)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Size below minimum
		sizeBelowMin := TestStruct{Count: 50, Quantity: 250, Size: 9}
		errors = v.Validate(sizeBelowMin)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Test edge values (exactly at boundaries)
		edgeValues := TestStruct{Count: 1, Quantity: 0, Size: 500}
		errors = v.Validate(edgeValues)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsEdgeValues, errors)
		}
	})

	// Test float range validation
	t.Run("validateFloatRange", func(t *testing.T) {
		type TestStruct struct {
			Price       float64 `json:"price" validate:"range=0.01,999.99"`
			Temperature float32 `json:"temperature" validate:"range=-50.0,50.0"`
			Rating      float64 `json:"rating" validate:"range=1.0,5.0"`
		}

		v := New()

		// Valid values within range
		valid := TestStruct{Price: 100.50, Temperature: 25.5, Rating: 3.5}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsValidRange, errors)
		}

		// Price below minimum
		belowMin := TestStruct{Price: 0.005, Temperature: 25.5, Rating: 3.5}
		errors = v.Validate(belowMin)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}
		if !strings.Contains(errors[0].Message, "must be between") {
			t.Errorf(msgExpectedRangeErr, errors[0].Message)
		}

		// Temperature below minimum (negative range)
		tempBelowMin := TestStruct{Price: 100.50, Temperature: -51.0, Rating: 3.5}
		errors = v.Validate(tempBelowMin)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Rating above maximum
		ratingAboveMax := TestStruct{Price: 100.50, Temperature: 25.5, Rating: 5.1}
		errors = v.Validate(ratingAboveMax)
		if len(errors) == 0 {
			t.Error(msgErrorOutsideRange)
		}

		// Test edge values (exactly at boundaries)
		edgeValues := TestStruct{Price: 0.01, Temperature: -50.0, Rating: 5.0}
		errors = v.Validate(edgeValues)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsEdgeValues, errors)
		}
	})

	// Test validateRange function (main dispatcher)
	t.Run("validateRange", func(t *testing.T) {
		type TestStruct struct {
			IntField   int     `json:"intField" validate:"range=1,10"`
			UintField  uint    `json:"uintField" validate:"range=0,100"`
			FloatField float64 `json:"floatField" validate:"range=0.1,9.9"`
		}

		v := New()

		// Valid values for all types
		valid := TestStruct{IntField: 5, UintField: 50, FloatField: 5.5}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf(msgNoErrorsValidRange, errors)
		}

		// Test multiple validation errors
		invalid := TestStruct{IntField: 11, UintField: 101, FloatField: 10.0}
		errors = v.Validate(invalid)
		if len(errors) != 3 {
			t.Errorf("Expected 3 range validation errors, got %d: %v", len(errors), errors)
		}
	})

	// Test invalid range specifications
	t.Run("invalidRangeSpecs", func(t *testing.T) {
		type TestStruct struct {
			BadRange1 int `json:"badRange1" validate:"range=invalid"`
			BadRange2 int `json:"badRange2" validate:"range=1,2,3"`
			BadRange3 int `json:"badRange3" validate:"range=abc,def"`
		}

		v := New()

		// Test single value range (should fail)
		singleValue := TestStruct{BadRange1: 5}
		errors := v.Validate(singleValue)
		if len(errors) == 0 {
			t.Error("Expected error for invalid range specification")
		}
		if !strings.Contains(errors[0].Message, "Invalid range specification") {
			t.Errorf("Expected invalid range specification error, got: %s", errors[0].Message)
		}

		// Test too many values in range
		tooManyValues := TestStruct{BadRange2: 5}
		errors = v.Validate(tooManyValues)
		if len(errors) == 0 {
			t.Error("Expected error for invalid range specification")
		}

		// Test non-numeric range values
		nonNumeric := TestStruct{BadRange3: 5}
		errors = v.Validate(nonNumeric)
		if len(errors) == 0 {
			t.Error("Expected error for non-numeric range values")
		}
	})

	// Test range validation with whitespace in range specification
	t.Run("rangeWithWhitespace", func(t *testing.T) {
		type TestStruct struct {
			Value int `json:"value" validate:"range= 1 , 10 "`
		}

		v := New()

		// Valid value (range should be trimmed)
		valid := TestStruct{Value: 5}
		errors := v.Validate(valid)
		if len(errors) > 0 {
			t.Errorf("Expected no errors for valid range with whitespace, got: %v", errors)
		}

		// Invalid value
		invalid := TestStruct{Value: 11}
		errors = v.Validate(invalid)
		if len(errors) == 0 {
			t.Error("Expected error for value outside trimmed range")
		}
	})
}

// TestParsingHelperFunctions tests the parsing helper functions with 0% coverage
func TestParsingHelperFunctions(t *testing.T) {
	// Test parseInt function
	t.Run("parseInt", func(t *testing.T) {
		// Valid integer strings
		validCases := []struct {
			input    string
			expected int64
		}{
			{"123", 123},
			{"-456", -456},
			{"0", 0},
			{"9223372036854775807", 9223372036854775807},   // max int64
			{"-9223372036854775808", -9223372036854775808}, // min int64
		}

		for _, tc := range validCases {
			result, err := parseInt(tc.input)
			if err != nil {
				t.Errorf("parseInt(%s) returned error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("parseInt(%s) = %d, expected %d", tc.input, result, tc.expected)
			}
		}

		// Invalid integer strings
		invalidCases := []string{
			"abc",
			"12.34",
			"",
			"123abc",
			"9223372036854775808", // overflow
		}

		for _, input := range invalidCases {
			_, err := parseInt(input)
			if err == nil {
				t.Errorf("parseInt(%s) should have returned an error", input)
			}
		}
	})

	// Test parseUint function
	t.Run("parseUint", func(t *testing.T) {
		// Valid unsigned integer strings
		validCases := []struct {
			input    string
			expected uint64
		}{
			{"123", 123},
			{"0", 0},
			{"18446744073709551615", 18446744073709551615}, // max uint64
		}

		for _, tc := range validCases {
			result, err := parseUint(tc.input)
			if err != nil {
				t.Errorf("parseUint(%s) returned error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("parseUint(%s) = %d, expected %d", tc.input, result, tc.expected)
			}
		}

		// Invalid unsigned integer strings
		invalidCases := []string{
			"abc",
			"-123", // negative
			"12.34",
			"",
			"123abc",
			"18446744073709551616", // overflow
		}

		for _, input := range invalidCases {
			_, err := parseUint(input)
			if err == nil {
				t.Errorf("parseUint(%s) should have returned an error", input)
			}
		}
	})

	// Test parseFloat function
	t.Run("parseFloat", func(t *testing.T) {
		// Valid float strings
		validCases := []struct {
			input    string
			expected float64
		}{
			{"123.456", 123.456},
			{"-789.012", -789.012},
			{"0", 0.0},
			{"0.0", 0.0},
			{"1.23e10", 1.23e10},
			{"-1.23e-5", -1.23e-5},
		}

		for _, tc := range validCases {
			result, err := parseFloat(tc.input)
			if err != nil {
				t.Errorf("parseFloat(%s) returned error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("parseFloat(%s) = %g, expected %g", tc.input, result, tc.expected)
			}
		}

		// Invalid float strings
		invalidCases := []string{
			"abc",
			"",
			"123abc",
			"12.34.56",
		}

		for _, input := range invalidCases {
			_, err := parseFloat(input)
			if err == nil {
				t.Errorf("parseFloat(%s) should have returned an error", input)
			}
		}
	})
}

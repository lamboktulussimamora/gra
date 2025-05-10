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

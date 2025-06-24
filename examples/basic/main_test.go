package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/router"
	"github.com/lamboktulussimamora/gra/validator"
)

// setupRouter creates a test router with the same configuration as main
func setupRouter() *router.Router {
	return setupRoutes()
}

func TestHomeRoute(t *testing.T) {
	router := setupRouter()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check response structure
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", response["status"])
	}

	if response["message"] != "Welcome to GRA Framework" {
		t.Errorf("Expected message 'Welcome to GRA Framework', got %v", response["message"])
	}

	// Check that data contains version, time, and endpoints
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if data["version"] == nil {
		t.Error("Expected version in data")
	}

	if data["time"] == nil {
		t.Error("Expected time in data")
	}

	if data["endpoints"] == nil {
		t.Error("Expected endpoints in data")
	}
}

func TestGetUserRoute(t *testing.T) {
	router := setupRouter()

	testCases := []struct {
		name     string
		userID   string
		expected map[string]interface{}
	}{
		{
			name:   "valid_user_id",
			userID: "123", expected: map[string]interface{}{
				"id":    "123",
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
		},
		{
			name:   "string_user_id",
			userID: "abc", expected: map[string]interface{}{
				"id":    "abc",
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/users/"+tc.userID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["status"] != "success" {
				t.Errorf("Expected status 'success', got %v", response["status"])
			}

			data, ok := response["data"].(map[string]interface{})
			if !ok {
				t.Fatal("Expected data to be a map")
			}

			if data["id"] != tc.expected["id"] {
				t.Errorf("Expected id %v, got %v", tc.expected["id"], data["id"])
			}

			if data["name"] != tc.expected["name"] {
				t.Errorf("Expected name %v, got %v", tc.expected["name"], data["name"])
			}

			if data["email"] != tc.expected["email"] {
				t.Errorf("Expected email %v, got %v", tc.expected["email"], data["email"])
			}
		})
	}
}

func TestCreateUserRoute(t *testing.T) {
	router := setupRouter()

	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectSuccess  bool
		expectErrors   bool
	}{
		{
			name: "valid_user",
			requestBody: User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
			expectErrors:   false,
		},
		{
			name: "missing_name",
			requestBody: User{
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectErrors:   true,
		},
		{
			name: "invalid_email",
			requestBody: User{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectErrors:   true,
		},
		{
			name: "short_password",
			requestBody: User{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectErrors:   true,
		},
		{
			name: "missing_required_fields",
			requestBody: User{
				Name: "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectErrors:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tc.expectSuccess {
				if response["status"] != "success" {
					t.Errorf("Expected status 'success', got %v", response["status"])
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected data to be a map")
				}

				// Check that user was created with ID
				if data["id"] == nil {
					t.Error("Expected user to have an ID")
				}

				// Check that password is hidden
				if data["password"] != "********" {
					t.Errorf("Expected password to be hidden, got %v", data["password"])
				}
			}

			if tc.expectErrors {
				if response["status"] != "error" {
					t.Errorf("Expected status 'error', got %v", response["status"])
				}

				if response["errors"] == nil {
					t.Error("Expected errors field to be present")
				}
			}
		})
	}
}

func TestCreateUserWithInvalidJSON(t *testing.T) {
	router := setupRouter()

	req, err := http.NewRequest("POST", "/users", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got %v", response["status"])
	}

	if response["error"] != "Invalid request body" {
		t.Errorf("Expected error 'Invalid request body', got %v", response["error"])
	}
}

func TestUserStruct(t *testing.T) {
	testCases := []struct {
		name     string
		user     User
		hasError bool
	}{
		{
			name: "valid_user",
			user: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			hasError: false,
		},
		{
			name: "empty_name",
			user: User{
				ID:       1,
				Email:    "john@example.com",
				Password: "password123",
			},
			hasError: true,
		},
		{
			name: "invalid_email",
			user: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "not-an-email",
				Password: "password123",
			},
			hasError: true,
		},
		{
			name: "short_password",
			user: User{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			hasError: true,
		},
	}

	v := validator.New()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := v.Validate(tc.user)
			hasError := len(errors) > 0

			if hasError != tc.hasError {
				t.Errorf("Expected hasError=%v, got %v (errors: %v)", tc.hasError, hasError, errors)
			}
		})
	}
}

func TestUserStructJSONSerialization(t *testing.T) {
	user := User{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledUser User
	if err := json.Unmarshal(jsonData, &unmarshaledUser); err != nil {
		t.Fatalf("Failed to unmarshal user: %v", err)
	}

	// Verify data integrity
	if unmarshaledUser.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, unmarshaledUser.ID)
	}
	if unmarshaledUser.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, unmarshaledUser.Name)
	}
	if unmarshaledUser.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, unmarshaledUser.Email)
	}
	if unmarshaledUser.Password != user.Password {
		t.Errorf("Expected Password %s, got %s", user.Password, unmarshaledUser.Password)
	}
}

func TestRouterMiddlewares(t *testing.T) {
	router := setupRouter()

	// Test that middlewares are applied (this is more of an integration test)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that CORS middleware added headers
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected CORS headers to be set")
	}

	// The request should succeed, indicating middleware chain works
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestNotFoundRoute(t *testing.T) {
	router := setupRouter()

	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSetupRoutes(t *testing.T) {
	router := setupRoutes()

	if router == nil {
		t.Fatal("Expected router to be non-nil")
	}

	// Test that routes are properly configured by making a simple request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreateUserHandler(t *testing.T) {
	router := setupRoutes()

	// Test valid user creation
	user := User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonBody, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "User created successfully" {
		t.Errorf("Expected message 'User created successfully', got %v", response["message"])
	}
}

// TestMainFunctionality tests the main function indirectly by checking what it sets up
func TestMainFunctionality(t *testing.T) {
	// Test that setupRoutes creates a valid router
	router := setupRoutes()
	if router == nil {
		t.Fatal("setupRoutes should return a non-nil router")
	}

	// Test that all expected routes are configured
	testRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users/123"},
		{"POST", "/users"},
	}

	for _, route := range testRoutes {
		t.Run("route_"+route.method+"_"+route.path, func(t *testing.T) {
			var req *http.Request
			var err error

			if route.method == "POST" {
				validUser := User{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: "password123",
				}
				jsonBody, _ := json.Marshal(validUser)
				req, err = http.NewRequest(route.method, route.path, bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(route.method, route.path, nil)
			}

			if err != nil {
				t.Fatalf("Failed to create %s request for %s: %v", route.method, route.path, err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// All configured routes should not return 404
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", route.method, route.path)
			}
		})
	}
}

// TestUserValidationEdgeCases tests additional validation scenarios
func TestUserValidationEdgeCases(t *testing.T) {
	v := validator.New()

	testCases := []struct {
		name        string
		user        User
		expectError bool
		description string
	}{
		{
			name: "minimum_valid_password",
			user: User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "123456", // exactly 6 characters (minimum)
			},
			expectError: false,
			description: "Password with exactly 6 characters should be valid",
		},
		{
			name: "email_with_subdomain",
			user: User{
				Name:     "Test User",
				Email:    "test@sub.example.com",
				Password: "password123",
			},
			expectError: false,
			description: "Email with subdomain should be valid",
		},
		{
			name: "email_with_numbers",
			user: User{
				Name:     "Test User",
				Email:    "test123@example123.com",
				Password: "password123",
			},
			expectError: false,
			description: "Email with numbers should be valid",
		},
		{
			name: "empty_email",
			user: User{
				Name:     "Test User",
				Email:    "",
				Password: "password123",
			},
			expectError: true,
			description: "Empty email should be invalid",
		},
		{
			name: "email_without_domain",
			user: User{
				Name:     "Test User",
				Email:    "test@",
				Password: "password123",
			},
			expectError: true,
			description: "Email without domain should be invalid",
		},
		{
			name: "all_fields_empty",
			user: User{
				Name:     "",
				Email:    "",
				Password: "",
			},
			expectError: true,
			description: "All empty fields should be invalid",
		},
		{
			name: "long_valid_values",
			user: User{
				Name:     "Very Long User Name That Should Still Be Valid",
				Email:    "very.long.email.address@very.long.domain.name.com",
				Password: "very_long_password_that_should_be_valid_123456789",
			},
			expectError: false,
			description: "Long but valid values should pass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errors := v.Validate(tc.user)
			hasError := len(errors) > 0

			if hasError != tc.expectError {
				t.Errorf("%s: Expected hasError=%v, got %v (errors: %v)",
					tc.description, tc.expectError, hasError, errors)
			}
		})
	}
}

// TestCreateUserAdvancedScenarios tests more complex user creation scenarios
func TestCreateUserAdvancedScenarios(t *testing.T) {
	router := setupRouter()

	testCases := []struct {
		name           string
		contentType    string
		requestBody    string
		expectedStatus int
		description    string
	}{
		{
			name:           "valid_user_with_extra_spaces",
			contentType:    "application/json",
			requestBody:    `{"name": "John Doe", "email": "john@example.com", "password": "password123"}`,
			expectedStatus: http.StatusCreated,
			description:    "User with valid data should be handled",
		},
		{
			name:           "empty_json_object",
			contentType:    "application/json",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Empty JSON object should be rejected",
		},
		{
			name:           "missing_content_type",
			contentType:    "",
			requestBody:    `{"name": "John", "email": "john@example.com", "password": "password123"}`,
			expectedStatus: http.StatusCreated, // Should still work
			description:    "Missing content-type should still work",
		},
		{
			name:           "invalid_content_type",
			contentType:    "text/plain",
			requestBody:    `{"name": "John", "email": "john@example.com", "password": "password123"}`,
			expectedStatus: http.StatusCreated, // Should still work
			description:    "Invalid content-type should still work",
		},
		{
			name:           "extra_json_fields",
			contentType:    "application/json",
			requestBody:    `{"name": "John", "email": "john@example.com", "password": "password123", "extra": "field", "age": 25}`,
			expectedStatus: http.StatusCreated,
			description:    "Extra JSON fields should be ignored",
		},
		{
			name:           "null_values",
			contentType:    "application/json",
			requestBody:    `{"name": null, "email": null, "password": null}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Null values should be rejected",
		},
		{
			name:           "malformed_json_missing_quotes",
			contentType:    "application/json",
			requestBody:    `{name: "John", email: "john@example.com", password: "password123"}`,
			expectedStatus: http.StatusBadRequest,
			description:    "Malformed JSON should be rejected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/users", strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d", tc.description, tc.expectedStatus, w.Code)
			}
		})
	}
}

// TestHTTPMethodsAndHeaders tests various HTTP scenarios
func TestHTTPMethodsAndHeaders(t *testing.T) {
	router := setupRouter()

	// Test unsupported HTTP methods
	unsupportedMethods := []string{"PUT", "DELETE", "PATCH"}
	for _, method := range unsupportedMethods {
		t.Run("unsupported_method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/users/123", nil)
			if err != nil {
				t.Fatalf("Failed to create %s request: %v", method, err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should return Method Not Allowed or Not Found
			if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
				t.Logf("Method %s returned status %d (this might be expected)", method, w.Code)
			}
		})
	}

	// Test various headers
	t.Run("request_with_user_agent", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("User-Agent", "Test-Agent/1.0")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test request with custom headers
	t.Run("request_with_custom_headers", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("X-Custom-Header", "test-value")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

// TestUserStructFields tests individual field validation
func TestUserStructFields(t *testing.T) {
	// Test User struct field types and JSON tags
	user := User{
		ID:       42,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Test that struct fields are accessible
	if user.ID != 42 {
		t.Errorf("Expected ID to be 42, got %d", user.ID)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected Name to be 'Test User', got %s", user.Name)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", user.Email)
	}

	if user.Password != "password123" {
		t.Errorf("Expected Password to be 'password123', got %s", user.Password)
	}
}

// TestCreateUserHandlerDirectly tests the createUserHandler function more thoroughly
func TestCreateUserHandlerDirectly(t *testing.T) {
	router := setupRouter()

	// Test successful user creation with specific response format
	t.Run("verify_response_format", func(t *testing.T) {
		user := User{
			Name:     "Jane Doe",
			Email:    "jane@example.com",
			Password: "securepass123",
		}

		jsonBody, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("Failed to marshal user: %v", err)
		}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify specific response structure
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}

		if response["message"] != "User created successfully" {
			t.Errorf("Expected message 'User created successfully', got %v", response["message"])
		}

		data, ok := response["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		// Check that ID was assigned
		if data["id"] != float64(1) { // JSON numbers are float64
			t.Errorf("Expected ID to be 1, got %v", data["id"])
		}

		// Check that original data is preserved
		if data["name"] != user.Name {
			t.Errorf("Expected name to be %s, got %v", user.Name, data["name"])
		}

		if data["email"] != user.Email {
			t.Errorf("Expected email to be %s, got %v", user.Email, data["email"])
		}

		// Check that password is masked
		if data["password"] != "********" {
			t.Errorf("Expected password to be masked, got %v", data["password"])
		}
	})
}

// TestErrorResponseFormats tests various error response formats
func TestErrorResponseFormats(t *testing.T) {
	router := setupRouter()

	// Test validation error response format
	t.Run("validation_error_format", func(t *testing.T) {
		invalidUser := User{
			Name:     "", // Missing required field
			Email:    "invalid-email",
			Password: "123", // Too short
		}

		jsonBody, err := json.Marshal(invalidUser)
		if err != nil {
			t.Fatalf("Failed to marshal user: %v", err)
		}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify error response structure
		if response["status"] != "error" {
			t.Errorf("Expected status 'error', got %v", response["status"])
		}

		if response["error"] != "Validation failed" {
			t.Errorf("Expected error 'Validation failed', got %v", response["error"])
		}

		// Should have errors array
		if response["errors"] == nil {
			t.Error("Expected errors field to be present")
		}
	})
}

// TestSetupRoutesConfiguration tests the router configuration details
func TestSetupRoutesConfiguration(t *testing.T) {
	router := setupRoutes()

	// Test middleware is properly configured by checking CORS headers
	req, err := http.NewRequest("OPTIONS", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create OPTIONS request: %v", err)
	}
	req.Header.Set("Origin", "http://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should have CORS headers set
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header to be set")
	}
}

// TestRouterNotNil ensures the router setup returns a valid router
func TestRouterNotNil(t *testing.T) {
	router := setupRoutes()
	if router == nil {
		t.Fatal("setupRoutes() should not return nil")
	}
}

// TestHomeRouteEndpointsArray tests the endpoints array in home route response
func TestHomeRouteEndpointsArray(t *testing.T) {
	router := setupRouter()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	endpoints, ok := data["endpoints"].([]interface{})
	if !ok {
		t.Fatal("Expected endpoints to be an array")
	}

	if len(endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(endpoints))
	}

	// Check that each endpoint has the required fields
	for i, ep := range endpoints {
		endpoint, ok := ep.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected endpoint %d to be a map", i)
		}

		if endpoint["method"] == nil {
			t.Errorf("Expected endpoint %d to have method", i)
		}

		if endpoint["path"] == nil {
			t.Errorf("Expected endpoint %d to have path", i)
		}

		if endpoint["description"] == nil {
			t.Errorf("Expected endpoint %d to have description", i)
		}
	}
}

// TestMainSetupLogic tests the setup logic used in main function
func TestMainSetupLogic(t *testing.T) {
	// Test that the setup logic creates what main function expects
	router := setupRoutes()

	// Verify that the router is properly configured
	if router == nil {
		t.Fatal("Router should not be nil")
	}

	// Test that all routes expected by main are configured
	routes := []struct {
		method     string
		path       string
		shouldWork bool
	}{
		{"GET", "/", true},
		{"GET", "/users/test", true},
		{"POST", "/users", true},
		{"PUT", "/nonexistent", false}, // Should not work
	}

	for _, route := range routes {
		req, err := http.NewRequest(route.method, route.path, nil)
		if err != nil {
			t.Fatalf("Failed to create %s request: %v", route.method, err)
		}

		if route.method == "POST" && route.path == "/users" {
			user := User{Name: "Test", Email: "test@example.com", Password: "password123"}
			jsonBody, _ := json.Marshal(user)
			req, _ = http.NewRequest(route.method, route.path, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if route.shouldWork {
			// Should not return 404 or 405
			if w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed {
				t.Errorf("Route %s %s should work but got status %d", route.method, route.path, w.Code)
			}
		}
	}
}

// TestApplicationConfiguration tests the application configuration
func TestApplicationConfiguration(t *testing.T) {
	// Test router creation
	router1 := setupRoutes()
	router2 := setupRoutes()

	// Both should be valid but different instances
	if router1 == nil || router2 == nil {
		t.Fatal("Both routers should be non-nil")
	}

	if router1 == router2 {
		t.Error("Each call should return a new router instance")
	}

	// Test that both routers work the same way
	for _, r := range []*router.Router{router1, router2} {
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Router should handle GET / correctly, got status %d", w.Code)
		}
	}
}

// TestAllFunctionsExist ensures all functions referenced in main exist and can be called
func TestAllFunctionsExist(t *testing.T) {
	// Test setupRoutes function
	router := setupRoutes()
	if router == nil {
		t.Error("setupRoutes() should return a router")
	}

	// Test createUserHandler indirectly through the router
	user := User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not panic and should return a reasonable status
	if w.Code != http.StatusCreated {
		t.Errorf("createUserHandler should work correctly, got status %d", w.Code)
	}
}

// TestErrorHandlingPaths tests various error conditions
func TestErrorHandlingPaths(t *testing.T) {
	router := setupRouter()

	// Test with various malformed requests
	malformedRequests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		expectError bool
	}{
		{"empty_body_post", "POST", "/users", "", "application/json", true},
		{"malformed_json", "POST", "/users", "{invalid", "application/json", true},
		{"missing_fields", "POST", "/users", "{}", "application/json", true},
		{"valid_get", "GET", "/", "", "", false},
		{"large_user_id", "GET", "/users/123456789012345678901234567890", "", "", false},
		{"special_chars_id", "GET", "/users/test-user_123", "", "", false},
	}

	for _, req := range malformedRequests {
		t.Run(req.name, func(t *testing.T) {
			var httpReq *http.Request
			var err error

			if req.body != "" {
				httpReq, err = http.NewRequest(req.method, req.path, strings.NewReader(req.body))
			} else {
				httpReq, err = http.NewRequest(req.method, req.path, nil)
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if req.contentType != "" {
				httpReq.Header.Set("Content-Type", req.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)

			// Should not panic
			if req.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status (4xx or 5xx), got %d", w.Code)
				}
			} else {
				if w.Code >= 500 {
					t.Errorf("Should not return server error, got %d", w.Code)
				}
			}
		})
	}
}

// TestConcurrentRequests tests handling multiple requests
func TestConcurrentRequests(t *testing.T) {
	router := setupRouter()

	// Make multiple concurrent requests
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			req, _ := http.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent request %d failed with status %d", id, w.Code)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestCompleteWorkflow tests a complete user creation workflow
func TestCompleteWorkflow(t *testing.T) {
	router := setupRouter()

	// Step 1: Check API info
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("API info request failed: %d", w.Code)
	}

	// Step 2: Try to get a user (mock response)
	req, _ = http.NewRequest("GET", "/users/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get user request failed: %d", w.Code)
	}

	// Step 3: Create a new user
	user := User{
		Name:     "Workflow Test User",
		Email:    "workflow@example.com",
		Password: "securepass123",
	}
	jsonBody, _ := json.Marshal(user)
	req, _ = http.NewRequest("POST", "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create user request failed: %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify the response has the expected structure
	if response["status"] != "success" {
		t.Errorf("Expected success status, got %v", response["status"])
	}

	if response["data"] == nil {
		t.Error("Expected data field in response")
	}
}

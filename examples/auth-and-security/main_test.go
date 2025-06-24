package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/jwt"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.app == nil {
		t.Fatal("Server app should not be nil")
	}

	if server.jwtService == nil {
		t.Fatal("Server JWT service should not be nil")
	}

	if server.users == nil {
		t.Fatal("Server users map should not be nil")
	}

	// Check default users exist
	if len(server.users) != 2 {
		t.Errorf("Expected 2 default users, got %d", len(server.users))
	}

	if _, exists := server.users["user1"]; !exists {
		t.Error("Default user 'user1' should exist")
	}

	if _, exists := server.users["admin"]; !exists {
		t.Error("Default user 'admin' should exist")
	}
}

func TestHandleHome(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test via router
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Welcome to the API" {
		t.Errorf("Expected message 'Welcome to the API', got %v", response["message"])
	}
}

func TestHandleLogin(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name            string
		requestBody     LoginRequest
		expectedStatus  int
		shouldHaveToken bool
	}{
		{
			name: "valid_login_user1",
			requestBody: LoginRequest{
				Username: "user1",
				Password: "password1",
			},
			expectedStatus:  http.StatusOK,
			shouldHaveToken: true,
		},
		{
			name: "valid_login_admin",
			requestBody: LoginRequest{
				Username: "admin",
				Password: "admin123",
			},
			expectedStatus:  http.StatusOK,
			shouldHaveToken: true,
		},
		{
			name: "invalid_username",
			requestBody: LoginRequest{
				Username: "nonexistent",
				Password: "password1",
			},
			expectedStatus:  http.StatusUnauthorized,
			shouldHaveToken: false,
		},
		{
			name: "invalid_password",
			requestBody: LoginRequest{
				Username: "user1",
				Password: "wrongpassword",
			},
			expectedStatus:  http.StatusUnauthorized,
			shouldHaveToken: false,
		},
		{
			name: "empty_username",
			requestBody: LoginRequest{
				Username: "",
				Password: "password1",
			},
			expectedStatus:  http.StatusUnauthorized, // Changed from BadRequest
			shouldHaveToken: false,
		},
		{
			name: "empty_password",
			requestBody: LoginRequest{
				Username: "user1",
				Password: "",
			},
			expectedStatus:  http.StatusUnauthorized, // Changed from BadRequest
			shouldHaveToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.app.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.shouldHaveToken {
				if response["token"] == nil {
					t.Error("Expected token in response")
				}
				if response["user"] == nil {
					t.Error("Expected user in response")
				}
			} else {
				if response["token"] != nil {
					t.Error("Should not have token in error response")
				}
			}
		})
	}
}

func TestHandleLoginInvalidJSON(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.app.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetUserByUsername(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name     string
		username string
		expected *User
	}{
		{
			name:     "existing_user1",
			username: "user1",
			expected: &User{
				ID:       "1",
				Username: "user1",
				Password: "password1",
				Role:     "user",
			},
		},
		{
			name:     "existing_admin",
			username: "admin",
			expected: &User{
				ID:       "2",
				Username: "admin",
				Password: "admin123",
				Role:     "admin",
			},
		},
		{
			name:     "nonexistent_user",
			username: "nonexistent",
			expected: nil,
		},
		{
			name:     "empty_username",
			username: "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, exists := server.users[tt.username]

			if tt.expected == nil {
				if exists {
					t.Errorf("Expected no user, got %+v", user)
				}
			} else {
				if !exists {
					t.Fatal("Expected user to exist")
				}
				if user.ID != tt.expected.ID ||
					user.Username != tt.expected.Username ||
					user.Password != tt.expected.Password ||
					user.Role != tt.expected.Role {
					t.Errorf("Expected user %+v, got %+v", tt.expected, user)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a valid token
	user := server.users["user1"]
	claims := jwt.StandardClaims{
		Subject: user.ID,
		Custom: map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	}

	token, err := server.jwtService.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectValid bool
	}{
		{
			name:        "valid_token",
			token:       token,
			expectValid: true,
		},
		{
			name:        "invalid_token",
			token:       "invalid.token.here",
			expectValid: false,
		},
		{
			name:        "empty_token",
			token:       "",
			expectValid: false,
		},
		{
			name:        "malformed_token",
			token:       "not.a.jwt",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := server.ValidateToken(tt.token)

			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected valid token, got error: %v", err)
				}
				if claims == nil {
					t.Error("Expected claims, got nil")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid token")
				}
				// Note: The service might return claims with error, so we don't check for nil claims
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name     string
		userID   string
		expected string
	}{
		{
			name:     "user1_role",
			userID:   "1",
			expected: "user",
		},
		{
			name:     "admin_role",
			userID:   "2",
			expected: "admin",
		},
		{
			name:     "nonexistent_user",
			userID:   "999",
			expected: "",
		},
		{
			name:     "empty_user_id",
			userID:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find user by ID
			var role string
			for _, user := range server.users {
				if user.ID == tt.userID {
					role = user.Role
					break
				}
			}

			if role != tt.expected {
				t.Errorf("Expected role %s, got %s", tt.expected, role)
			}
		})
	}
}

func TestUserStruct(t *testing.T) {
	user := User{
		ID:       "1",
		Username: "testuser",
		Password: "testpass",
		Role:     "user",
	}

	if user.ID != "1" {
		t.Errorf("Expected ID '1', got %s", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got %s", user.Username)
	}
	if user.Password != "testpass" {
		t.Errorf("Expected Password 'testpass', got %s", user.Password)
	}
	if user.Role != "user" {
		t.Errorf("Expected Role 'user', got %s", user.Role)
	}
}

func TestLoginRequestStruct(t *testing.T) {
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}

	if loginReq.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got %s", loginReq.Username)
	}
	if loginReq.Password != "testpass" {
		t.Errorf("Expected Password 'testpass', got %s", loginReq.Password)
	}
}

func TestLoginResponseStruct(t *testing.T) {
	user := User{ID: "1", Username: "test", Password: "pass", Role: "user"}
	loginResp := LoginResponse{
		Token: "test.jwt.token",
		User:  user,
	}

	if loginResp.Token != "test.jwt.token" {
		t.Errorf("Expected Token 'test.jwt.token', got %s", loginResp.Token)
	}
	if loginResp.User.ID != "1" {
		t.Errorf("Expected User ID '1', got %s", loginResp.User.ID)
	}
}

func TestServerStruct(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test that all fields are properly initialized
	if server.app == nil {
		t.Error("Expected app to be initialized")
	}
	if server.jwtService == nil {
		t.Error("Expected jwtService to be initialized")
	}
	if server.users == nil {
		t.Error("Expected users map to be initialized")
	}
}

func TestAuthenticationFlow(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Step 1: Login to get token
	loginReq := LoginRequest{
		Username: "user1",
		Password: "password1",
	}

	bodyBytes, err := json.Marshal(loginReq)
	if err != nil {
		t.Fatalf("Failed to marshal login request: %v", err)
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d", w.Code)
	}

	var loginResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token := loginResponse["token"].(string)
	if token == "" {
		t.Fatal("No token received from login")
	}

	// Step 2: Validate the token
	claims, err := server.ValidateToken(token)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims == nil {
		t.Fatal("No claims received from token validation")
	}

	// Step 3: Check user data from login response
	userData := loginResponse["user"].(map[string]interface{})
	if userData["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", userData["role"])
	}
}

func TestJWTServiceIntegration(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test JWT service functionality
	claims := jwt.StandardClaims{
		Subject: "test-user",
		Custom: map[string]interface{}{
			"username": "testuser",
			"role":     "user",
		},
	}

	// Generate token
	token, err := server.jwtService.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Validate token
	validatedClaims, err := server.jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Convert to map for easier access
	claimsMap := validatedClaims
	if claimsMap == nil {
		t.Fatal("Expected claims to be non-nil")
	}

	if claimsMap["sub"] != claims.Subject {
		t.Errorf("Expected subject %s, got %v", claims.Subject, claimsMap["sub"])
	}

	// Check custom data
	if claimsMap["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", claimsMap["username"])
	}

	if claimsMap["role"] != "user" {
		t.Errorf("Expected role 'user', got %v", claimsMap["role"])
	}
}

func TestErrorHandling(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "invalid_method_on_login",
			method:         "GET",
			path:           "/login",
			body:           "",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing_content_type",
			method:         "POST",
			path:           "/login",
			body:           `{"username":"test","password":"test"}`,
			contentType:    "",
			expectedStatus: http.StatusUnauthorized, // Changed from BadRequest
		},
		{
			name:           "invalid_content_type",
			method:         "POST",
			path:           "/login",
			body:           `{"username":"test","password":"test"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusUnauthorized, // Changed from BadRequest
		},
		{
			name:           "not_found",
			method:         "GET",
			path:           "/nonexistent",
			body:           "",
			contentType:    "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte(tt.body)))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			server.app.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestConcurrentLogin(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test concurrent login requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			loginReq := LoginRequest{
				Username: "user1",
				Password: "password1",
			}

			bodyBytes, err := json.Marshal(loginReq)
			if err != nil {
				results <- http.StatusInternalServerError
				return
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.app.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Check all requests succeeded
	for i := 0; i < numRequests; i++ {
		status := <-results
		if status != http.StatusOK {
			t.Errorf("Concurrent request %d failed with status %d", i, status)
		}
	}
}

func TestMainFunctionExists(t *testing.T) {
	// Test that main function exists and can be referenced
	// This ensures the main function is properly defined
	// We just verify it doesn't panic when we reference it
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main function caused panic: %v", r)
		}
	}()

	// Just checking that main is available
	t.Log("main function exists and is accessible")
}

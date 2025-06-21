package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

func TestDetermineAllowedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       string
	}{
		{
			name:           "Wildcard with no origin",
			origin:         "",
			allowedOrigins: []string{"*"},
			expected:       "*",
		},
		{
			name:           "Exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://another.com"},
			expected:       "https://example.com",
		},
		{
			name:           "Wildcard with origin",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       "https://example.com",
		},
		{
			name:           "No match",
			origin:         "https://notallowed.com",
			allowedOrigins: []string{"https://example.com"},
			expected:       "",
		},
		{
			name:           "Empty origin, no wildcard",
			origin:         "",
			allowedOrigins: []string{"https://example.com"},
			expected:       "",
		},
		{
			name:           "Multiple allowed origins with match",
			origin:         "https://second.com",
			allowedOrigins: []string{"https://first.com", "https://second.com", "https://third.com"},
			expected:       "https://second.com",
		},
		{
			name:           "Empty allowed origins",
			origin:         "https://example.com",
			allowedOrigins: []string{},
			expected:       "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := determineAllowedOrigin(test.origin, test.allowedOrigins)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Item exists",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "Item does not exist",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "grape",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "apple",
			expected: false,
		},
		{
			name:     "Empty item",
			slice:    []string{"apple", "", "cherry"},
			item:     "",
			expected: true,
		},
		{
			name:     "Case sensitive",
			slice:    []string{"Apple", "Banana"},
			item:     "apple",
			expected: false,
		},
		{
			name:     "Single item match",
			slice:    []string{"onlyitem"},
			item:     "onlyitem",
			expected: true,
		},
		{
			name:     "Single item no match",
			slice:    []string{"onlyitem"},
			item:     "other",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := contains(test.slice, test.item)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	// Test default allow origins
	if len(config.AllowOrigins) != 1 || config.AllowOrigins[0] != "*" {
		t.Errorf("Expected default AllowOrigins to be ['*'], got %v", config.AllowOrigins)
	}

	// Test default allow methods
	expectedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}
	if len(config.AllowMethods) != len(expectedMethods) {
		t.Errorf("Expected %d default methods, got %d", len(expectedMethods), len(config.AllowMethods))
	}

	for i, method := range expectedMethods {
		if config.AllowMethods[i] != method {
			t.Errorf("Expected method '%s' at index %d, got '%s'", method, i, config.AllowMethods[i])
		}
	}

	// Test default allow headers
	expectedHeaders := []string{"Authorization", "Content-Type"}
	if len(config.AllowHeaders) != len(expectedHeaders) {
		t.Errorf("Expected %d default headers, got %d", len(expectedHeaders), len(config.AllowHeaders))
	}

	// Test default expose headers
	if len(config.ExposeHeaders) != 0 {
		t.Errorf("Expected empty ExposeHeaders by default, got %v", config.ExposeHeaders)
	}

	// Test default allow credentials
	if config.AllowCredentials != false {
		t.Errorf("Expected AllowCredentials to be false by default, got %v", config.AllowCredentials)
	}

	// Test default max age
	if config.MaxAge != 86400 {
		t.Errorf("Expected MaxAge to be 86400 by default, got %d", config.MaxAge)
	}
}

func TestCORSWithComplexConfig(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"https://frontend.com", "https://admin.com"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Authorization", "X-Custom-Header"},
		ExposeHeaders:    []string{"X-Total-Count", "X-Page-Count"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	middleware := CORSWithConfig(config)

	// Test preflight request from allowed origin
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://frontend.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	nextCalled := false
	next := func(c *context.Context) {
		nextCalled = true
	}

	middleware(next)(ctx)

	// Verify headers
	if w.Header().Get("Access-Control-Allow-Origin") != "https://frontend.com" {
		t.Errorf("Expected Allow-Origin 'https://frontend.com', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("Expected Allow-Methods 'GET, POST', got '%s'", w.Header().Get("Access-Control-Allow-Methods"))
	}

	if w.Header().Get("Access-Control-Allow-Headers") != "Authorization, X-Custom-Header" {
		t.Errorf("Expected Allow-Headers 'Authorization, X-Custom-Header', got '%s'", w.Header().Get("Access-Control-Allow-Headers"))
	}

	if w.Header().Get("Access-Control-Expose-Headers") != "X-Total-Count, X-Page-Count" {
		t.Errorf("Expected Expose-Headers 'X-Total-Count, X-Page-Count', got '%s'", w.Header().Get("Access-Control-Expose-Headers"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Allow-Credentials 'true', got '%s'", w.Header().Get("Access-Control-Allow-Credentials"))
	}

	if w.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("Expected Max-Age '3600', got '%s'", w.Header().Get("Access-Control-Max-Age"))
	}

	// Verify preflight request doesn't call next
	if nextCalled {
		t.Error("Next handler should not be called for OPTIONS request")
	}
}

func TestCORSWithCredentialsAndWildcard(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}

	middleware := CORSWithConfig(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	next := func(c *context.Context) {}
	middleware(next)(ctx)

	// Should not set credentials with wildcard origin for security
	if w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Error("Should not set Allow-Credentials with wildcard origin")
	}
}

func TestCORSWithUnallowedOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"https://allowed.com"},
	}

	middleware := CORSWithConfig(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://notallowed.com")

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	next := func(c *context.Context) {}
	middleware(next)(ctx)

	// Should not set any CORS headers for unallowed origin
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set Allow-Origin for unallowed origin")
	}
}

func TestCORSSimpleFunction(t *testing.T) {
	middleware := CORS("https://mysite.com")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://mysite.com")

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	next := func(c *context.Context) {}
	middleware(next)(ctx)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://mysite.com" {
		t.Errorf("Expected Allow-Origin 'https://mysite.com', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSWithEmptyConfig(t *testing.T) {
	config := CORSConfig{} // Empty config

	middleware := CORSWithConfig(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	next := func(c *context.Context) {}
	middleware(next)(ctx)

	// With empty config, no CORS headers should be set
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set Allow-Origin with empty AllowOrigins")
	}

	if w.Header().Get("Access-Control-Max-Age") != "" {
		t.Error("Should not set Max-Age when MaxAge is 0")
	}
}

func TestSetCORSHeadersEdgeCases(t *testing.T) {
	// Test with no origin header
	config := DefaultCORSConfig()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header set

	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	setCORSHeaders(ctx, config)

	// Should set wildcard when no origin and wildcard is allowed
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Allow-Origin '*' when no origin header, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestSetExtendedCORSHeadersEdgeCases(t *testing.T) {
	headers := make(http.Header)

	// Test with credentials but wildcard origin (should not set credentials)
	config := CORSConfig{
		AllowCredentials: true,
		MaxAge:           0, // Zero max age should not set header
	}

	setExtendedCORSHeaders(headers, config, "https://example.com", true) // hasWildcard = true

	if headers.Get("Access-Control-Allow-Credentials") != "" {
		t.Error("Should not set credentials with wildcard origin")
	}

	if headers.Get("Access-Control-Max-Age") != "" {
		t.Error("Should not set Max-Age when MaxAge is 0")
	}

	// Test with credentials and specific origin (should set credentials)
	headers = make(http.Header)
	setExtendedCORSHeaders(headers, config, "https://example.com", false) // hasWildcard = false

	if headers.Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Should set credentials with specific origin")
	}

	// Test with positive max age
	config.MaxAge = 1800
	headers = make(http.Header)
	setExtendedCORSHeaders(headers, config, "https://example.com", false)

	if headers.Get("Access-Control-Max-Age") != "1800" {
		t.Errorf("Expected Max-Age '1800', got '%s'", headers.Get("Access-Control-Max-Age"))
	}
}

func TestAuthMiddlewareEdgeCases(t *testing.T) {
	// Mock JWT service
	mockJWT := &MockJWTService{
		validateFunc: func(token string) (any, error) {
			if token == "valid-token" {
				return map[string]interface{}{"user_id": 123}, nil
			}
			return nil, fmt.Errorf("invalid token")
		},
	}

	middleware := Auth(mockJWT, "claims")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedNext   bool
	}{
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedNext:   false,
		},
		{
			name:           "Invalid header format - no space",
			authHeader:     "Bearertoken",
			expectedStatus: http.StatusUnauthorized,
			expectedNext:   false,
		},
		{
			name:           "Invalid header format - wrong prefix",
			authHeader:     "Basic dG9rZW4=",
			expectedStatus: http.StatusUnauthorized,
			expectedNext:   false,
		},
		{
			name:           "Invalid header format - too many parts",
			authHeader:     "Bearer token extra part",
			expectedStatus: http.StatusUnauthorized,
			expectedNext:   false,
		},
		{
			name:           "Valid token",
			authHeader:     "Bearer valid-token",
			expectedStatus: 0, // No error expected
			expectedNext:   true,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedNext:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}

			w := httptest.NewRecorder()
			ctx := context.New(w, req)

			nextCalled := false
			next := func(c *context.Context) {
				nextCalled = true
			}

			middleware(next)(ctx)

			if nextCalled != test.expectedNext {
				t.Errorf("Expected next called: %v, got: %v", test.expectedNext, nextCalled)
			}

			if test.expectedStatus > 0 && w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			// Test claims are set for valid token
			if test.name == "Valid token" {
				claims := ctx.Value("claims")
				if claims == nil {
					t.Error("Expected claims to be set in context")
				}
			}
		})
	}
}

// MockJWTService for testing
type MockJWTService struct {
	validateFunc func(string) (any, error)
}

func (m *MockJWTService) ValidateToken(token string) (any, error) {
	return m.validateFunc(token)
}

func TestRecoveryMiddlewareWithPanic(t *testing.T) {
	middleware := Recovery()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	next := func(c *context.Context) {
		panic("test panic")
	}

	// Should not panic
	middleware(next)(ctx)

	// Should return 500 error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check that response contains error message
	if !strings.Contains(w.Body.String(), "Internal server error") {
		t.Error("Expected error message in response body")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	middleware := Logger()

	req := httptest.NewRequest(http.MethodPost, "/test/path", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	nextCalled := false
	next := func(c *context.Context) {
		nextCalled = true
	}

	// Should not panic and should call next
	middleware(next)(ctx)

	if !nextCalled {
		t.Error("Next handler should be called")
	}
}

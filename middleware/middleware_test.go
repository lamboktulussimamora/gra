package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// MockJWTAuthenticator is a mock implementation of JWTAuthenticator
type MockJWTAuthenticator struct {
	ShouldSucceed bool
	Claims        map[string]any
}

func (m *MockJWTAuthenticator) ValidateToken(tokenString string) (any, error) {
	if !m.ShouldSucceed {
		return nil, errors.New("invalid token")
	}
	return m.Claims, nil
}

func TestAuth(t *testing.T) {
	// Create mock JWT authenticator
	claims := map[string]any{
		"userId": "123",
		"role":   "admin",
	}

	mockJWT := &MockJWTAuthenticator{
		ShouldSucceed: true,
		Claims:        claims,
	}

	// Create a handler to verify the auth middleware
	handlerCalled := false
	var capturedClaims any
	testHandler := func(c *context.Context) {
		handlerCalled = true
		capturedClaims = c.Value("user")
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"status": "success",
		})
	}

	// Create the auth middleware
	authMiddleware := Auth(mockJWT, "user")
	wrappedHandler := authMiddleware(testHandler)

	testCases := []struct {
		name           string
		authHeader     string
		shouldSucceed  bool
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer valid-token",
			shouldSucceed:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No auth header",
			authHeader:     "",
			shouldSucceed:  true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid header format",
			authHeader:     "InvalidFormat token",
			shouldSucceed:  true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			shouldSucceed:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset test variables
			handlerCalled = false
			capturedClaims = nil
			mockJWT.ShouldSucceed = tc.shouldSucceed

			// Create test request and response
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/protected", nil)

			if tc.authHeader != "" {
				r.Header.Set("Authorization", tc.authHeader)
			}

			c := context.New(w, r)

			// Execute middleware and handler
			wrappedHandler(c)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check if handler was called when expected
			shouldCallHandler := tc.expectedStatus == http.StatusOK
			if shouldCallHandler && !handlerCalled {
				t.Error("Expected handler to be called, but it wasn't")
			}

			if !shouldCallHandler && handlerCalled {
				t.Error("Expected handler not to be called, but it was")
			}

			// Verify claims were passed when handler was called
			if handlerCalled {
				if capturedClaims == nil {
					t.Error("Expected claims to be added to context, but they weren't")
				} else {
					claimsMap, ok := capturedClaims.(map[string]any)
					if !ok {
						t.Error("Claims not of expected type")
					} else if claimsMap["userId"] != claims["userId"] {
						t.Errorf("Expected userId %v, got %v", claims["userId"], claimsMap["userId"])
					}
				}
			}
		})
	}
}

func TestLogger(t *testing.T) {
	// Create a handler to verify the logger middleware
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	// Create the logger middleware
	loggerMiddleware := Logger()
	wrappedHandler := loggerMiddleware(testHandler)

	// Create test request and response
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := context.New(w, r)

	// Execute middleware and handler
	wrappedHandler(c)

	// Check if handler was called
	if !handlerCalled {
		t.Error("Expected handler to be called, but it wasn't")
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRecovery(t *testing.T) {
	testCases := []struct {
		name           string
		handler        router.HandlerFunc
		expectedStatus int
		shouldPanic    bool
	}{
		{
			name: "No panic",
			handler: func(c *context.Context) {
				c.Status(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			shouldPanic:    false,
		},
		{
			name: "With panic",
			handler: func(c *context.Context) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			shouldPanic:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the recovery middleware
			recoveryMiddleware := Recovery()
			wrappedHandler := recoveryMiddleware(tc.handler)

			// Create test request and response
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			c := context.New(w, r)

			// Execute middleware and handler
			wrappedHandler(c)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	testCases := []struct {
		name             string
		method           string
		expectedStatus   int
		handlerShouldRun bool
	}{
		{
			name:             "Normal GET request",
			method:           "GET",
			expectedStatus:   http.StatusOK,
			handlerShouldRun: true,
		},
		{
			name:             "CORS preflight request",
			method:           "OPTIONS",
			expectedStatus:   http.StatusOK,
			handlerShouldRun: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Track if handler was called
			handlerCalled := false
			testHandler := func(c *context.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}

			// Create the CORS middleware
			allowOrigin := "*"
			corsMiddleware := CORS(allowOrigin)
			wrappedHandler := corsMiddleware(testHandler)

			// Create test request and response
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "/test", nil)
			c := context.New(w, r)

			// Execute middleware and handler
			wrappedHandler(c)

			// Check if handler was called when expected
			if tc.handlerShouldRun && !handlerCalled {
				t.Error("Expected handler to be called, but it wasn't")
			}

			if !tc.handlerShouldRun && handlerCalled {
				t.Error("Expected handler not to be called, but it was")
			}

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check CORS headers
			headers := w.Header()

			if headers.Get("Access-Control-Allow-Origin") != allowOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin to be %s, got %s",
					allowOrigin, headers.Get("Access-Control-Allow-Origin"))
			}

			if headers.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
				t.Errorf("Expected Access-Control-Allow-Methods to be %s, got %s",
					"GET, POST, PUT, DELETE, OPTIONS", headers.Get("Access-Control-Allow-Methods"))
			}

			if headers.Get("Access-Control-Allow-Headers") != "Authorization, Content-Type" {
				t.Errorf("Expected Access-Control-Allow-Headers to be %s, got %s",
					"Authorization, Content-Type", headers.Get("Access-Control-Allow-Headers"))
			}
		})
	}
}

// Header name constants for testing
const (
	headerXSSProtection       = "X-XSS-Protection"
	headerContentTypeOptions  = "X-Content-Type-Options"
	headerFrameOptions        = "X-Frame-Options"
	headerReferrerPolicy      = "Referrer-Policy"
	headerCSP                 = "Content-Security-Policy"
	headerHSTS                = "Strict-Transport-Security"
	headerCrossOriginResource = "Cross-Origin-Resource-Policy"
)

func TestSecureHeaders(t *testing.T) {
	// Create a request with a method and URL
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a context
	c := context.New(w, req)

	// Create handler function
	handlerFunc := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		if _, err := c.Writer.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	// Apply secure headers middleware
	middleware := SecureHeaders()
	handler := middleware(handlerFunc)

	// Call handler
	handler(c)

	// Assert headers
	headers := w.Result().Header

	// Check X-XSS-Protection header
	if headers.Get(headerXSSProtection) != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection header to be '1; mode=block', got '%s'", headers.Get(headerXSSProtection))
	}

	// Check X-Content-Type-Options header
	if headers.Get(headerContentTypeOptions) != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options header to be 'nosniff', got '%s'", headers.Get(headerContentTypeOptions))
	}

	// Check X-Frame-Options header
	if headers.Get(headerFrameOptions) != "SAMEORIGIN" {
		t.Errorf("Expected X-Frame-Options header to be 'SAMEORIGIN', got '%s'", headers.Get(headerFrameOptions))
	}

	// Check Referrer-Policy header
	if headers.Get(headerReferrerPolicy) != "no-referrer" {
		t.Errorf("Expected Referrer-Policy header to be 'no-referrer', got '%s'", headers.Get(headerReferrerPolicy))
	}

	// Check Cross-Origin-Resource-Policy header
	if headers.Get(headerCrossOriginResource) != "same-origin" {
		t.Errorf("Expected Cross-Origin-Resource-Policy header to be 'same-origin', got '%s'", headers.Get(headerCrossOriginResource))
	}
}

func TestSecureHeadersWithConfig(t *testing.T) {
	// Create a custom config
	config := SecureHeadersConfig{
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            300,
		HSTSIncludeSubdomains: true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "same-origin",
	}

	// Create a request with a method and URL
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a context
	c := context.New(w, req)

	// Create handler function
	handlerFunc := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		if _, err := c.Writer.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	// Apply secure headers middleware with custom config
	middleware := SecureHeadersWithConfig(config)
	handler := middleware(handlerFunc)

	// Call handler
	handler(c)

	// Assert headers
	headers := w.Result().Header

	// Check X-XSS-Protection header
	if headers.Get(headerXSSProtection) != "0" {
		t.Errorf("Expected X-XSS-Protection header to be '0', got '%s'", headers.Get(headerXSSProtection))
	}

	// Check X-Frame-Options header
	if headers.Get(headerFrameOptions) != "DENY" {
		t.Errorf("Expected X-Frame-Options header to be 'DENY', got '%s'", headers.Get(headerFrameOptions))
	}

	// Check Content-Security-Policy header
	if headers.Get(headerCSP) != "default-src 'self'" {
		t.Errorf("Expected Content-Security-Policy header to be \"default-src 'self'\", got '%s'", headers.Get(headerCSP))
	}

	// Check HSTS header
	expectedHSTS := "max-age=300; includeSubDomains"
	if headers.Get(headerHSTS) != expectedHSTS {
		t.Errorf("Expected Strict-Transport-Security header to be '%s', got '%s'", expectedHSTS, headers.Get(headerHSTS))
	}
}

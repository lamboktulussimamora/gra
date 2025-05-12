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

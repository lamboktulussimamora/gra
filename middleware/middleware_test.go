package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// Test error message constants to avoid duplication
const (
	testUserID                  = "123"
	testRole                    = "admin"
	claimsKey                   = "user"
	errExpectedHandlerCalled    = "Expected handler to be called, but it wasn't"
	errExpectedHandlerNotCalled = "Expected handler not to be called, but it was"
	errStatusCodeMismatch       = "Expected status code %d, got %d"
	errClaimsNotAdded           = "Expected claims to be added to context, but they weren't"
	errClaimsWrongType          = "Claims not of expected type"
	errUserIDMismatch           = "Expected userID %v, got %v"
	errHeaderMismatch           = "Expected %s to be %s, got %s"
)

// Constants for authorization headers
const (
	bearerTokenPrefix   = "Bearer "
	validTokenHeader    = bearerTokenPrefix + "valid-token"
	invalidFormatHeader = "InvalidFormat token"
	invalidTokenHeader  = bearerTokenPrefix + "invalid-token"
)

// MockJWTAuthenticator is a mock implementation of JWTAuthenticator
type MockJWTAuthenticator struct {
	ShouldSucceed bool
	Claims        map[string]any
}

func (m *MockJWTAuthenticator) ValidateToken(_ string) (any, error) {
	// This implementation ignores the actual token string value
	// as we're only testing based on the ShouldSucceed flag
	if !m.ShouldSucceed {
		return nil, errors.New("invalid token")
	}
	return m.Claims, nil
}

// TestAuth tests the Auth middleware functionality
func TestAuth(t *testing.T) {
	// Set up test cases
	testCases := []struct {
		name           string
		authHeader     string
		shouldSucceed  bool
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     validTokenHeader,
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
			authHeader:     invalidFormatHeader,
			shouldSucceed:  true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     invalidTokenHeader,
			shouldSucceed:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	// Set up test claims
	claims := map[string]any{
		"userID": testUserID,
		"role":   testRole,
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runAuthTest(t, tc.authHeader, tc.shouldSucceed, tc.expectedStatus, claims)
		})
	}
}

// runAuthTest executes a single Auth middleware test case
func runAuthTest(t *testing.T, authHeader string, shouldSucceed bool, expectedStatus int, claims map[string]any) {
	// Create test variables
	handlerCalled := false
	var capturedClaims any

	// Create mock JWT authenticator
	mockJWT := &MockJWTAuthenticator{
		ShouldSucceed: shouldSucceed,
		Claims:        claims,
	}

	// Create test handler
	testHandler := func(c *context.Context) {
		handlerCalled = true
		capturedClaims = c.Value(claimsKey)
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"status": "success",
		})
	}

	// Create auth middleware
	authMiddleware := Auth(mockJWT, claimsKey)
	wrappedHandler := authMiddleware(testHandler)

	// Create request and response
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	// Set auth header if provided
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}

	// Execute middleware
	c := context.New(w, r)
	wrappedHandler(c)

	// Verify response status code
	if w.Code != expectedStatus {
		t.Errorf(errStatusCodeMismatch, expectedStatus, w.Code)
	}

	// Verify handler execution
	verifyHandlerExecution(t, expectedStatus, handlerCalled)

	// Verify claims if handler was called
	if handlerCalled {
		verifyClaims(t, capturedClaims, claims)
	}
}

// verifyHandlerExecution checks if the handler was called when expected
func verifyHandlerExecution(t *testing.T, expectedStatus int, handlerCalled bool) {
	t.Helper()
	shouldCallHandler := expectedStatus == http.StatusOK

	if shouldCallHandler && !handlerCalled {
		t.Error(errExpectedHandlerCalled)
	}
	if !shouldCallHandler && handlerCalled {
		t.Error(errExpectedHandlerNotCalled)
	}
}

// verifyCORSHandlerExecution checks if the handler was called when expected for CORS tests
func verifyCORSHandlerExecution(t *testing.T, handlerShouldRun bool, handlerCalled bool) {
	t.Helper()

	if handlerShouldRun && !handlerCalled {
		t.Error(errExpectedHandlerCalled)
	}
	if !handlerShouldRun && handlerCalled {
		t.Error(errExpectedHandlerNotCalled)
	}
}

// verifyClaims checks if the expected claims were passed to the context
func verifyClaims(t *testing.T, capturedClaims any, expectedClaims map[string]any) {
	t.Helper()
	if capturedClaims == nil {
		t.Error(errClaimsNotAdded)
		return
	}

	claimsMap, ok := capturedClaims.(map[string]any)
	if !ok {
		t.Error(errClaimsWrongType)
		return
	}

	if claimsMap["userID"] != expectedClaims["userID"] {
		t.Errorf(errUserIDMismatch, expectedClaims["userID"], claimsMap["userID"])
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
		t.Error(errExpectedHandlerCalled)
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf(errStatusCodeMismatch, http.StatusOK, w.Code)
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
			handler: func(_ *context.Context) {
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
				t.Errorf(errStatusCodeMismatch, tc.expectedStatus, w.Code)
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

	// Constants for CORS header validation
	const (
		allowOrigin        = "*"
		allowMethods       = "GET, POST, PUT, DELETE, OPTIONS"
		allowHeaders       = "Authorization, Content-Type"
		headerAllowOrigin  = "Access-Control-Allow-Origin"
		headerAllowMethods = "Access-Control-Allow-Methods"
		headerAllowHeaders = "Access-Control-Allow-Headers"
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Track if handler was called
			handlerCalled := false
			testHandler := func(c *context.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}

			// Create the CORS middleware
			corsMiddleware := CORS(allowOrigin)
			wrappedHandler := corsMiddleware(testHandler)

			// Create test request and response
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "/test", nil)
			c := context.New(w, r)

			// Execute middleware and handler
			wrappedHandler(c)

			// Verify handler execution
			verifyCORSHandlerExecution(t, tc.handlerShouldRun, handlerCalled)

			// Verify status code
			if w.Code != tc.expectedStatus {
				t.Errorf(errStatusCodeMismatch, tc.expectedStatus, w.Code)
			}

			// Verify CORS headers
			headers := w.Header()
			if headers.Get(headerAllowOrigin) != allowOrigin {
				t.Errorf(errHeaderMismatch,
					headerAllowOrigin, allowOrigin, headers.Get(headerAllowOrigin))
			}

			if headers.Get(headerAllowMethods) != allowMethods {
				t.Errorf(errHeaderMismatch,
					headerAllowMethods, allowMethods, headers.Get(headerAllowMethods))
			}

			if headers.Get(headerAllowHeaders) != allowHeaders {
				t.Errorf(errHeaderMismatch,
					headerAllowHeaders, allowHeaders, headers.Get(headerAllowHeaders))
			}
		})
	}
}

// Header name and value constants for testing
const (
	// Header names
	headerXSSProtection       = "X-XSS-Protection"
	headerContentTypeOptions  = "X-Content-Type-Options"
	headerFrameOptions        = "X-Frame-Options"
	headerReferrerPolicy      = "Referrer-Policy"
	headerCSP                 = "Content-Security-Policy"
	headerHSTS                = "Strict-Transport-Security"
	headerCrossOriginResource = "Cross-Origin-Resource-Policy"

	// Header default values
	valueXSSProtection       = "1; mode=block"
	valueContentTypeOptions  = "nosniff"
	valueFrameOptions        = "SAMEORIGIN"
	valueReferrerPolicy      = "no-referrer"
	valueSameOrigin          = "same-origin"
	valueCrossOriginResource = valueSameOrigin

	// Error message format for header value mismatch
	errHeaderValueMismatch = "Expected %s header to be '%s', got '%s'"
)

// verifySecureHeader checks if a security header has the expected value
func verifySecureHeader(t *testing.T, headers http.Header, headerName string, expectedValue string) {
	t.Helper()
	if headers.Get(headerName) != expectedValue {
		t.Errorf(errHeaderValueMismatch, headerName, expectedValue, headers.Get(headerName))
	}
}

func TestSecureHeaders(t *testing.T) {
	// Create a request with a method and URL
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
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

	// Check security headers with helper function
	verifySecureHeader(t, headers, headerXSSProtection, valueXSSProtection)
	verifySecureHeader(t, headers, headerContentTypeOptions, valueContentTypeOptions)
	verifySecureHeader(t, headers, headerFrameOptions, valueFrameOptions)
	verifySecureHeader(t, headers, headerReferrerPolicy, valueReferrerPolicy)
	verifySecureHeader(t, headers, headerCrossOriginResource, valueCrossOriginResource)
}

func TestSecureHeadersWithConfig(t *testing.T) {
	// Custom values for the test
	const (
		customXSSProtection         = "0"
		customXFrameOptions         = "DENY"
		customCSP                   = "default-src 'self'"
		customReferrerPolicy        = valueSameOrigin
		customHSTSMaxAge            = 300
		customHSTSMaxAgeHeaderValue = "max-age=300; includeSubDomains"
		customCrossOriginResource   = valueSameOrigin
	)

	// Create a custom config
	config := SecureHeadersConfig{
		XSSProtection:             customXSSProtection,
		ContentTypeNosniff:        valueContentTypeOptions,
		XFrameOptions:             customXFrameOptions,
		HSTSMaxAge:                customHSTSMaxAge,
		HSTSIncludeSubdomains:     true,
		ContentSecurityPolicy:     customCSP,
		ReferrerPolicy:            customReferrerPolicy,
		CrossOriginResourcePolicy: customCrossOriginResource,
	}

	// Create a request with a method and URL
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
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

	// Check security headers with custom values
	verifySecureHeader(t, headers, headerXSSProtection, customXSSProtection)
	verifySecureHeader(t, headers, headerContentTypeOptions, valueContentTypeOptions)
	verifySecureHeader(t, headers, headerFrameOptions, customXFrameOptions)
	verifySecureHeader(t, headers, headerCSP, customCSP)
	verifySecureHeader(t, headers, headerReferrerPolicy, customReferrerPolicy)
	verifySecureHeader(t, headers, headerHSTS, customHSTSMaxAgeHeaderValue)
	verifySecureHeader(t, headers, headerCrossOriginResource, valueCrossOriginResource)
}

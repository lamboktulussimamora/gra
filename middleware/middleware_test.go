package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestSecureHeaders_HSTSPreload(t *testing.T) {
	cfg := DefaultSecureHeadersConfig()
	cfg.HSTSMaxAge = 10
	cfg.HSTSIncludeSubdomains = true
	cfg.HSTSPreload = true

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	c := context.New(w, r)
	h := SecureHeadersWithConfig(cfg)(func(c *context.Context) { c.Status(http.StatusOK) })
	h(c)

	if got := w.Header().Get(headerHSTS); got != "max-age=10; includeSubDomains; preload" {
		t.Fatalf("unexpected HSTS header with preload: %q", got)
	}
}

// ---- Additional tests to improve coverage ----

func TestInMemoryStoreIncrement_BelowAndAtLimit(t *testing.T) {
	store := NewInMemoryStore()
	limit := 3
	window := 2 // seconds

	// First call: count=1, exceeded=false
	count, exceeded := store.Increment("key", limit, window)
	if exceeded {
		t.Fatalf("unexpected exceeded on first call")
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	// Second call: count=2, exceeded=false
	count, exceeded = store.Increment("key", limit, window)
	if exceeded {
		t.Fatalf("unexpected exceeded on second call")
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	// Third call hits the limit but should not mark exceeded yet (increment happens when not exceeded)
	count, exceeded = store.Increment("key", limit, window)
	if exceeded {
		t.Fatalf("unexpected exceeded on reaching limit; should exceed on next attempt")
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	// Fourth call should return exceeded=true and not increment count
	count, exceeded = store.Increment("key", limit, window)
	if !exceeded {
		t.Fatalf("expected exceeded=true on call beyond limit")
	}
	if count != 3 {
		t.Fatalf("expected count to remain 3 when exceeded, got %d", count)
	}
}

func TestInMemoryStoreIncrement_CleanupOldEntries(t *testing.T) {
	store := NewInMemoryStore()
	key := "k"
	limit := 10
	window := 1 // second

	// Seed with an old timestamp outside window and a current timestamp
	now := time.Now().Unix()
	store.data[key] = map[int64]int{
		now - 10: 5, // should be cleaned
		now:      2,
	}

	// Call increment once; this should cleanup the old bucket and increment the current
	count, exceeded := store.Increment(key, limit, window)
	if exceeded {
		t.Fatalf("did not expect to exceed limit during cleanup test")
	}
	if count != 3 {
		t.Fatalf("expected count 3 after cleanup+increment, got %d", count)
	}

	// Verify old entries cleaned
	for ts := range store.data[key] {
		if ts < now-int64(window) {
			t.Fatalf("found stale timestamp %d not cleaned", ts)
		}
	}
}

func TestRateLimitMiddleware_HeadersAndBlocking(t *testing.T) {
	store := NewInMemoryStore()
	cfg := RateLimiterConfig{
		Store:  store,
		Limit:  2,
		Window: 2,
		KeyFunc: func(_ *context.Context) string {
			return "fixed-key"
		},
		ErrorMessage: "Rate limit exceeded. Try again later.",
	}

	called := 0
	handler := func(c *context.Context) {
		called++
		c.Status(http.StatusOK)
	}

	mw := RateLimitWithConfig(cfg)
	wrapped := mw(handler)

	// First request
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w1, r1))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if got := w1.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Fatalf("expected X-RateLimit-Limit=2, got %s", got)
	}
	if got := w1.Header().Get("X-RateLimit-Remaining"); got != "1" {
		t.Fatalf("expected X-RateLimit-Remaining=1, got %s", got)
	}

	// Second request
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w2, r2))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if got := w2.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected X-RateLimit-Remaining=0, got %s", got)
	}

	// Third request should be limited
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w3, r3))
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w3.Code)
	}
	if called != 2 {
		t.Fatalf("handler should have been called exactly twice, got %d", called)
	}
}

func TestRateLimitMiddleware_ExcludeFunc(t *testing.T) {
	cfg := RateLimiterConfig{
		Store:  NewInMemoryStore(),
		Limit:  1,
		Window: 1,
		KeyFunc: func(_ *context.Context) string {
			return "ignored"
		},
		ExcludeFunc: func(_ *context.Context) bool { return true },
	}
	called := false
	handler := func(c *context.Context) {
		called = true
		c.Status(http.StatusOK)
	}
	wrapped := RateLimitWithConfig(cfg)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w, r))
	if !called {
		t.Fatalf("expected handler to be called when excluded")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// When excluded, middleware returns early and should not set rate headers
	if got := w.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Fatalf("expected no X-RateLimit-Limit when excluded, got %s", got)
	}
}

func TestRequestID_Default(t *testing.T) {
	var captured string
	handler := func(c *context.Context) {
		if v := c.Value("requestID"); v != nil {
			if s, ok := v.(string); ok {
				captured = s
			}
		}
		c.Status(http.StatusOK)
	}
	wrapped := RequestID()(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w, r))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if captured == "" {
		t.Fatalf("expected request ID stored in context")
	}
	if hdr := w.Header().Get("X-Request-ID"); hdr == "" {
		t.Fatalf("expected X-Request-ID response header to be set")
	}
}

func TestRequestID_WithConfig(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	cfg.HeaderName = "X-Custom-ReqID"
	cfg.ResponseHeader = false // don't expose header
	cfg.Generator = func() string { return "fixed-id" }

	var captured string
	handler := func(c *context.Context) {
		if v := c.Value(cfg.ContextKey); v != nil {
			if s, ok := v.(string); ok {
				captured = s
			}
		}
		c.Status(http.StatusOK)
	}

	wrapped := RequestIDWithConfig(cfg)(handler)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	wrapped(context.New(w, r))

	if captured != "fixed-id" {
		t.Fatalf("expected fixed-id in context, got %s", captured)
	}
	if hdr := w.Header().Get(cfg.HeaderName); hdr != "" {
		t.Fatalf("did not expect response header %s to be set", cfg.HeaderName)
	}
}

func TestContainsHelper(t *testing.T) {
	list := []string{"a", "b", "c"}
	if !contains(list, "a") || !contains(list, "c") {
		t.Fatalf("expected contains true for existing items")
	}
	if contains(list, "z") {
		t.Fatalf("expected contains false for non-existing item")
	}
}

func TestDetermineAllowedOrigin(t *testing.T) {
	// wildcard when no origin
	if got := determineAllowedOrigin("", []string{"*"}); got != "*" {
		t.Fatalf("expected '*', got %q", got)
	}
	// exact match
	if got := determineAllowedOrigin("http://a", []string{"http://a", "http://b"}); got != "http://a" {
		t.Fatalf("expected origin match, got %q", got)
	}
	// not allowed
	if got := determineAllowedOrigin("http://c", []string{"http://a"}); got != "" {
		t.Fatalf("expected empty string for not allowed, got %q", got)
	}
}

func TestCSPBuilderAndMiddleware(t *testing.T) {
	b := NewCSPBuilder().DefaultSrc("'self'").ScriptSrc("'self'", "cdn.example.com").UpgradeInsecureRequests()
	csp := b.Build()
	if csp == "" || !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "script-src 'self' cdn.example.com") || !strings.Contains(csp, "upgrade-insecure-requests") {
		t.Fatalf("unexpected CSP string: %s", csp)
	}

	mw := CSP(b)
	handler := mw(func(c *context.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(context.New(w, r))
	if got := w.Header().Get("Content-Security-Policy"); got != csp {
		t.Fatalf("expected CSP header %q, got %q", csp, got)
	}
}

func TestCORSWithConfig_Custom(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins:     []string{"http://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"X-Test"},
		ExposeHeaders:    []string{"X-Exposed"},
		AllowCredentials: true,
		MaxAge:           60,
	}
	mw := CORSWithConfig(cfg)
	h := mw(func(c *context.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "http://example.com")
	h(context.New(w, r))
	hdr := w.Result().Header
	if got := hdr.Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("expected allowed origin to echo request origin, got %q", got)
	}
	if got := hdr.Get("Access-Control-Allow-Methods"); got != "GET, POST" {
		t.Fatalf("unexpected allow methods: %q", got)
	}
	if got := hdr.Get("Access-Control-Allow-Headers"); got != "X-Test" {
		t.Fatalf("unexpected allow headers: %q", got)
	}
	if got := hdr.Get("Access-Control-Expose-Headers"); got != "X-Exposed" {
		t.Fatalf("unexpected expose headers: %q", got)
	}
	if got := hdr.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials true, got %q", got)
	}
	if got := hdr.Get("Access-Control-Max-Age"); got != "60" {
		t.Fatalf("unexpected max age: %q", got)
	}
}

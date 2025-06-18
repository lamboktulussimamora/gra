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

// Global test constants to avoid duplication
const (
	// Test data
	testUserID        = "123"
	testRole          = "admin"
	claimsKey         = "user"
	testKey           = "test-key"
	customRequestID   = "custom-id-123"
	existingRequestID = "existing-id"

	// CSP directives
	selfDirective = "'self'"
	noneDirective = "'none'"
	customCSP     = "default-src 'self'"

	// Origins
	testOriginExample = "https://example.com"
	allowOrigin       = "*"
	allowMethods      = "GET, POST, PUT, DELETE, OPTIONS"
	allowHeaders      = "Authorization, Content-Type"

	// Error messages
	expectedHandlerCalled       = "Expected handler to be called"
	errExpectedHandlerNotCalled = "Expected handler not to be called, but it was"
	errStatusCodeMismatch       = "Expected status code %d, got %d"
	errClaimsNotAdded           = "Expected claims to be added to context, but they weren't"
	errClaimsWrongType          = "Claims not of expected type"
	errUserIDMismatch           = "Expected userID %v, got %v"
	errHeaderMismatch           = "Expected %s to be %s, got %s"

	// Headers
	headerAllowOrigin     = "Access-Control-Allow-Origin"
	headerAllowMethods    = "Access-Control-Allow-Methods"
	headerAllowHeaders    = "Access-Control-Allow-Headers"
	headerRequestID       = "X-Request-ID"
	headerCustomRequestID = "X-Custom-Request-ID"
	headerCSP             = "Content-Security-Policy"

	// Authorization headers
	bearerTokenPrefix   = "Bearer "
	validTokenHeader    = bearerTokenPrefix + "valid-token"
	invalidFormatHeader = "InvalidFormat token"
	invalidTokenHeader  = bearerTokenPrefix + "invalid-token"

	// Security headers
	headerXSSProtection       = "X-XSS-Protection"
	headerContentTypeOptions  = "X-Content-Type-Options"
	headerFrameOptions        = "X-Frame-Options"
	headerReferrerPolicy      = "Referrer-Policy"
	headerHSTS                = "Strict-Transport-Security"
	headerCrossOriginResource = "Cross-Origin-Resource-Policy"

	// Security header values
	valueXSSProtection       = "1; mode=block"
	valueContentTypeOptions  = "nosniff"
	valueFrameOptions        = "SAMEORIGIN"
	valueReferrerPolicy      = "no-referrer"
	valueSameOrigin          = "same-origin"
	valueCrossOriginResource = valueSameOrigin

	// Error message format for header value mismatch
	errHeaderValueMismatch = "Expected %s header to be '%s', got '%s'"
)

// MockJWTAuthenticator is a mock implementation of JWTAuthenticator
type MockJWTAuthenticator struct {
	ShouldSucceed bool
	Claims        map[string]any
}

func (m *MockJWTAuthenticator) ValidateToken(_ string) (any, error) {
	if !m.ShouldSucceed {
		return nil, errors.New("invalid token")
	}
	return m.Claims, nil
}

// TestAuth tests the Auth middleware functionality
func TestAuth(t *testing.T) {
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

	claims := map[string]any{
		"userID": testUserID,
		"role":   testRole,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runAuthTest(t, tc.authHeader, tc.shouldSucceed, tc.expectedStatus, claims)
		})
	}
}

// runAuthTest executes a single Auth middleware test case
func runAuthTest(t *testing.T, authHeader string, shouldSucceed bool, expectedStatus int, claims map[string]any) {
	handlerCalled := false
	var capturedClaims any

	mockJWT := &MockJWTAuthenticator{
		ShouldSucceed: shouldSucceed,
		Claims:        claims,
	}

	testHandler := func(c *context.Context) {
		handlerCalled = true
		capturedClaims = c.Value(claimsKey)
		c.Status(http.StatusOK)
	}

	authMiddleware := Auth(mockJWT, claimsKey)
	wrappedHandler := authMiddleware(testHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/protected", nil)

	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}

	c := context.New(w, r)
	wrappedHandler(c)

	if w.Code != expectedStatus {
		t.Errorf(errStatusCodeMismatch, expectedStatus, w.Code)
	}

	verifyHandlerExecution(t, expectedStatus, handlerCalled)

	if handlerCalled {
		verifyClaims(t, capturedClaims, claims)
	}
}

// verifyHandlerExecution checks if the handler was called when expected
func verifyHandlerExecution(t *testing.T, expectedStatus int, handlerCalled bool) {
	t.Helper()
	shouldCallHandler := expectedStatus == http.StatusOK

	if shouldCallHandler && !handlerCalled {
		t.Error(expectedHandlerCalled)
	}
	if !shouldCallHandler && handlerCalled {
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
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	loggerMiddleware := Logger()
	wrappedHandler := loggerMiddleware(testHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := context.New(w, r)

	wrappedHandler(c)

	if !handlerCalled {
		t.Error(expectedHandlerCalled)
	}

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
			recoveryMiddleware := Recovery()
			wrappedHandler := recoveryMiddleware(tc.handler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			c := context.New(w, r)

			wrappedHandler(c)

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			testHandler := func(c *context.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}

			corsMiddleware := CORS(allowOrigin)
			wrappedHandler := corsMiddleware(testHandler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "/test", nil)
			c := context.New(w, r)

			wrappedHandler(c)

			verifyCORSHandlerExecution(t, tc.handlerShouldRun, handlerCalled)

			if w.Code != tc.expectedStatus {
				t.Errorf(errStatusCodeMismatch, tc.expectedStatus, w.Code)
			}

			headers := w.Header()
			if headers.Get(headerAllowOrigin) != allowOrigin {
				t.Errorf(errHeaderMismatch, headerAllowOrigin, allowOrigin, headers.Get(headerAllowOrigin))
			}

			if headers.Get(headerAllowMethods) != allowMethods {
				t.Errorf(errHeaderMismatch, headerAllowMethods, allowMethods, headers.Get(headerAllowMethods))
			}

			if headers.Get(headerAllowHeaders) != allowHeaders {
				t.Errorf(errHeaderMismatch, headerAllowHeaders, allowHeaders, headers.Get(headerAllowHeaders))
			}
		})
	}
}

// verifyCORSHandlerExecution checks if the handler was called when expected for CORS tests
func verifyCORSHandlerExecution(t *testing.T, handlerShouldRun bool, handlerCalled bool) {
	t.Helper()

	if handlerShouldRun && !handlerCalled {
		t.Error(expectedHandlerCalled)
	}
	if !handlerShouldRun && handlerCalled {
		t.Error(errExpectedHandlerNotCalled)
	}
}

// verifySecureHeader checks if a security header has the expected value
func verifySecureHeader(t *testing.T, headers http.Header, headerName string, expectedValue string) {
	t.Helper()
	if headers.Get(headerName) != expectedValue {
		t.Errorf(errHeaderValueMismatch, headerName, expectedValue, headers.Get(headerName))
	}
}

func TestSecureHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c := context.New(w, req)

	handlerFunc := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		if _, err := c.Writer.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	middleware := SecureHeaders()
	handler := middleware(handlerFunc)

	handler(c)

	headers := w.Result().Header

	verifySecureHeader(t, headers, headerXSSProtection, valueXSSProtection)
	verifySecureHeader(t, headers, headerContentTypeOptions, valueContentTypeOptions)
	verifySecureHeader(t, headers, headerFrameOptions, valueFrameOptions)
	verifySecureHeader(t, headers, headerReferrerPolicy, valueReferrerPolicy)
	verifySecureHeader(t, headers, headerCrossOriginResource, valueCrossOriginResource)
}

func TestSecureHeadersWithConfig(t *testing.T) {
	const (
		customXSSProtection         = "0"
		customXFrameOptions         = "DENY"
		customReferrerPolicy        = valueSameOrigin
		customHSTSMaxAge            = 300
		customHSTSMaxAgeHeaderValue = "max-age=300; includeSubDomains"
		customCrossOriginResource   = valueSameOrigin
	)

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

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c := context.New(w, req)

	handlerFunc := func(c *context.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		if _, err := c.Writer.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}

	middleware := SecureHeadersWithConfig(config)
	handler := middleware(handlerFunc)

	handler(c)

	headers := w.Result().Header

	verifySecureHeader(t, headers, headerXSSProtection, customXSSProtection)
	verifySecureHeader(t, headers, headerContentTypeOptions, valueContentTypeOptions)
	verifySecureHeader(t, headers, headerFrameOptions, customXFrameOptions)
	verifySecureHeader(t, headers, headerCSP, customCSP)
	verifySecureHeader(t, headers, headerReferrerPolicy, customReferrerPolicy)
	verifySecureHeader(t, headers, headerHSTS, customHSTSMaxAgeHeaderValue)
	verifySecureHeader(t, headers, headerCrossOriginResource, valueCrossOriginResource)
}

func TestRateLimit(t *testing.T) {
	runRateLimitTest(t, "Within limit", 5, 60, 3, []int{200, 200, 200})
	runRateLimitTest(t, "Exceed limit", 2, 60, 3, []int{200, 200, 429})
}

func runRateLimitTest(t *testing.T, name string, limit, windowSeconds, requestCount int, expectedStatus []int) {
	t.Run(name, func(t *testing.T) {
		middleware := RateLimit(limit, windowSeconds)
		handlerCalled := make([]bool, requestCount)

		handler := func(c *context.Context) {
			for i := range handlerCalled {
				if !handlerCalled[i] {
					handlerCalled[i] = true
					break
				}
			}
			c.Status(http.StatusOK)
		}

		wrappedHandler := middleware(handler)

		for i := 0; i < requestCount; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			ctx := context.New(w, req)

			wrappedHandler(ctx)

			if w.Code != expectedStatus[i] {
				t.Errorf("Request %d: expected status %d, got %d", i, expectedStatus[i], w.Code)
			}

			verifyRateLimitHeaders(t, w)
		}
	})
}

func verifyRateLimitHeaders(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header to be set")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header to be set")
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header to be set")
	}
}

func TestRateLimitWithConfig(t *testing.T) {
	store := NewInMemoryStore()
	config := RateLimiterConfig{
		Store:  store,
		Limit:  2,
		Window: 60,
		KeyFunc: func(_ *context.Context) string {
			return testKey
		},
		ExcludeFunc: func(c *context.Context) bool {
			return c.Request.URL.Path == "/excluded"
		},
		ErrorMessage: "Custom rate limit message",
	}

	middleware := RateLimitWithConfig(config)
	handler := func(c *context.Context) {
		c.Status(http.StatusOK)
	}
	wrappedHandler := middleware(handler)

	// Test excluded path
	req := httptest.NewRequest("GET", "/excluded", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)
	wrappedHandler(ctx)
	if w.Code != http.StatusOK {
		t.Errorf("Expected excluded path to pass, got status %d", w.Code)
	}

	// Test normal rate limiting
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		ctx := context.New(w, req)
		wrappedHandler(ctx)

		expectedStatus := http.StatusOK
		if i >= 2 {
			expectedStatus = http.StatusTooManyRequests
		}

		if w.Code != expectedStatus {
			t.Errorf("Request %d: expected status %d, got %d", i, expectedStatus, w.Code)
		}
	}
}

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	count, exceeded := store.Increment(testKey, 3, 60)
	if count != 1 || exceeded {
		t.Errorf("Expected count=1, exceeded=false, got count=%d, exceeded=%t", count, exceeded)
	}

	store.Increment(testKey, 3, 60)
	count, exceeded = store.Increment(testKey, 3, 60)
	if count != 3 || exceeded {
		t.Errorf("Expected count=3, exceeded=false, got count=%d, exceeded=%t", count, exceeded)
	}

	count, exceeded = store.Increment(testKey, 3, 60)
	if count != 3 || !exceeded {
		t.Errorf("Expected count=3, exceeded=true, got count=%d, exceeded=%t", count, exceeded)
	}
}

func TestRequestID(t *testing.T) {
	middleware := RequestID()
	handlerCalled := false

	handler := func(c *context.Context) {
		handlerCalled = true
		if c.Value("requestID") == nil {
			t.Error("Expected request ID to be in context")
		}
		c.Status(http.StatusOK)
	}

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	wrappedHandler(ctx)

	if !handlerCalled {
		t.Error(expectedHandlerCalled)
	}

	if w.Header().Get(headerRequestID) == "" {
		t.Error("Expected X-Request-ID header to be set in response")
	}
}

func TestRequestIDWithConfig(t *testing.T) {
	config := RequestIDConfig{
		Generator: func() string {
			return customRequestID
		},
		HeaderName:     headerCustomRequestID,
		ContextKey:     "customRequestID",
		ResponseHeader: true,
	}

	middleware := RequestIDWithConfig(config)
	handlerCalled := false

	handler := func(c *context.Context) {
		handlerCalled = true
		if c.Value("customRequestID") != customRequestID {
			t.Error("Expected custom request ID to be in context")
		}
		c.Status(http.StatusOK)
	}

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	wrappedHandler(ctx)

	if !handlerCalled {
		t.Error(expectedHandlerCalled)
	}

	if w.Header().Get(headerCustomRequestID) != customRequestID {
		t.Error("Expected X-Custom-Request-ID header to be set with custom value")
	}
}

func TestRequestIDExistingHeader(t *testing.T) {
	middleware := RequestID()
	handlerCalled := false

	handler := func(c *context.Context) {
		handlerCalled = true
		if c.Value("requestID") != existingRequestID {
			t.Error("Expected existing request ID to be preserved")
		}
		c.Status(http.StatusOK)
	}

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(headerRequestID, existingRequestID)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	wrappedHandler(ctx)

	if !handlerCalled {
		t.Error(expectedHandlerCalled)
	}

	if w.Header().Get(headerRequestID) != existingRequestID {
		t.Error("Expected existing X-Request-ID header to be preserved")
	}
}

func TestCSPBuilder(t *testing.T) {
	builder := NewCSPBuilder()

	csp := builder.
		DefaultSrc(selfDirective).
		ScriptSrc(selfDirective, "'unsafe-inline'").
		StyleSrc(selfDirective, "https://fonts.googleapis.com").
		ImgSrc(selfDirective, "data:", "https:").
		ConnectSrc(selfDirective).
		FontSrc(selfDirective, "https://fonts.gstatic.com").
		ObjectSrc(noneDirective).
		MediaSrc(selfDirective).
		FrameSrc(noneDirective).
		WorkerSrc(selfDirective).
		FrameAncestors(noneDirective).
		FormAction(selfDirective).
		ReportTo("csp-endpoint").
		ReportURI("/csp-report").
		UpgradeInsecureRequests().
		Build()

	if csp == "" {
		t.Error("Expected CSP string to be non-empty")
	}

	expectedDirectives := []string{
		"default-src " + selfDirective,
		"script-src " + selfDirective + " 'unsafe-inline'",
		"style-src " + selfDirective + " https://fonts.googleapis.com",
		"object-src " + noneDirective,
		"upgrade-insecure-requests",
	}

	for _, directive := range expectedDirectives {
		if !strings.Contains(csp, directive) {
			t.Errorf("Expected CSP to contain '%s', got: %s", directive, csp)
		}
	}
}

func TestCSPMiddleware(t *testing.T) {
	builder := NewCSPBuilder().
		DefaultSrc(selfDirective).
		ScriptSrc(selfDirective)

	middleware := CSP(builder)
	handlerCalled := false

	handler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	}

	wrappedHandler := middleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.New(w, req)

	wrappedHandler(ctx)

	if !handlerCalled {
		t.Error(expectedHandlerCalled)
	}

	cspHeader := w.Header().Get(headerCSP)
	if cspHeader == "" {
		t.Error("Expected Content-Security-Policy header to be set")
	}

	if !strings.Contains(cspHeader, "default-src "+selfDirective) {
		t.Error("Expected CSP header to contain default-src directive")
	}
}

// TestRateLimitingAdvanced tests advanced rate limiting scenarios
func TestRateLimitingAdvanced(t *testing.T) {
	t.Run("concurrent requests within rate limit", func(t *testing.T) {
		rateLimit := 10
		timeWindowSeconds := 1
		middleware := RateLimit(rateLimit, timeWindowSeconds)

		// Create a test handler
		handlerCalled := 0
		handler := func(c *context.Context) {
			handlerCalled++
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Make concurrent requests within the rate limit
		numRequests := 5
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:8080"
				rr := httptest.NewRecorder()
				c := &context.Context{
					Request: req,
					Writer:  rr,
				}

				wrappedHandler(c)
				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}

		// All requests should have been allowed
		if handlerCalled != numRequests {
			t.Errorf("Expected %d handler calls, got %d", numRequests, handlerCalled)
		}
	})

	t.Run("different IP addresses have separate rate limits", func(t *testing.T) {
		rateLimit := 2
		timeWindowSeconds := 1
		middleware := RateLimit(rateLimit, timeWindowSeconds)

		handler := func(c *context.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Test with two different IP addresses
		ips := []string{"192.168.1.1:8080", "192.168.1.2:8080"}

		for _, ip := range ips {
			// Each IP should be able to make rate limit number of requests
			for i := 0; i < rateLimit; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = ip
				rr := httptest.NewRecorder()
				c := &context.Context{
					Request: req,
					Writer:  rr,
				}

				wrappedHandler(c)

				if rr.Code != http.StatusOK {
					t.Errorf("Request %d from IP %s should have been allowed, got status %d", i+1, ip, rr.Code)
				}
			}

			// The next request should be rate limited
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip
			rr := httptest.NewRecorder()
			c := &context.Context{
				Request: req,
				Writer:  rr,
			}

			wrappedHandler(c)

			if rr.Code != http.StatusTooManyRequests {
				t.Errorf("Request from IP %s should have been rate limited, got status %d", ip, rr.Code)
			}
		}
	})

	t.Run("rate limit resets after time window", func(t *testing.T) {
		rateLimit := 1
		timeWindowSeconds := 1 // 1 second window
		middleware := RateLimit(rateLimit, timeWindowSeconds)

		handler := func(c *context.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Make first request - should succeed
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:8080"
		rr1 := httptest.NewRecorder()
		c1 := &context.Context{
			Request: req1,
			Writer:  rr1,
		}

		wrappedHandler(c1)

		if rr1.Code != http.StatusOK {
			t.Errorf("First request should have been allowed, got status %d", rr1.Code)
		}

		// Make second request immediately - should be rate limited
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.1:8080"
		rr2 := httptest.NewRecorder()
		c2 := &context.Context{
			Request: req2,
			Writer:  rr2,
		}

		wrappedHandler(c2)

		if rr2.Code != http.StatusTooManyRequests {
			t.Errorf("Second request should have been rate limited, got status %d", rr2.Code)
		}

		// Wait for time window to reset
		time.Sleep(1100 * time.Millisecond) // Wait slightly longer than 1 second

		// Make third request - should succeed after reset
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req3.RemoteAddr = "192.168.1.1:8080"
		rr3 := httptest.NewRecorder()
		c3 := &context.Context{
			Request: req3,
			Writer:  rr3,
		}

		wrappedHandler(c3)

		if rr3.Code != http.StatusOK {
			t.Errorf("Third request after reset should have been allowed, got status %d", rr3.Code)
		}
	})
}

// TestCORSAdvanced tests advanced CORS scenarios
func TestCORSAdvanced(t *testing.T) {
	t.Run("preflight request with custom headers", func(t *testing.T) {
		corsConfig := CORSConfig{
			AllowOrigins:  []string{"https://example.com", "https://test.com"},
			AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut},
			AllowHeaders:  []string{"Authorization", "Content-Type", "X-Custom-Header"},
			ExposeHeaders: []string{"X-Total-Count", "X-Page-Count"},
			MaxAge:        3600,
		}

		middleware := CORSWithConfig(corsConfig)

		handler := func(c *context.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Create preflight request
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		req.Header.Set("Access-Control-Request-Headers", "Authorization, X-Custom-Header")

		rr := httptest.NewRecorder()
		c := &context.Context{
			Request: req,
			Writer:  rr,
		}

		wrappedHandler(c)

		// Check preflight response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d for preflight, got %d", http.StatusOK, rr.Code)
		}

		// Verify CORS headers
		headers := map[string]string{
			"Access-Control-Allow-Origin":  "https://example.com",
			"Access-Control-Allow-Methods": strings.Join(corsConfig.AllowMethods, ", "),
			"Access-Control-Allow-Headers": strings.Join(corsConfig.AllowHeaders, ", "),
			"Access-Control-Max-Age":       "3600",
		}

		for header, expectedValue := range headers {
			actualValue := rr.Header().Get(header)
			if actualValue != expectedValue {
				t.Errorf("Expected %s to be '%s', got '%s'", header, expectedValue, actualValue)
			}
		}
	})

	t.Run("actual request with credentials", func(t *testing.T) {
		corsConfig := CORSConfig{
			AllowOrigins:     []string{"https://example.com"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost},
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			ExposeHeaders:    []string{"X-Total-Count"},
			AllowCredentials: true,
		}

		middleware := CORSWithConfig(corsConfig)

		handler := func(c *context.Context) {
			c.SetHeader("X-Total-Count", "100")
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Create actual request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Authorization", "Bearer token")

		rr := httptest.NewRecorder()
		c := &context.Context{
			Request: req,
			Writer:  rr,
		}

		wrappedHandler(c)

		// Check response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		// Verify CORS headers for actual request
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin":      "https://example.com",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Expose-Headers":    "X-Total-Count",
		}

		for header, expectedValue := range expectedHeaders {
			actualValue := rr.Header().Get(header)
			if actualValue != expectedValue {
				t.Errorf("Expected %s to be '%s', got '%s'", header, expectedValue, actualValue)
			}
		}
	})

	t.Run("rejected origin", func(t *testing.T) {
		corsConfig := CORSConfig{
			AllowOrigins: []string{"https://allowed.com"},
			AllowMethods: []string{http.MethodGet},
		}

		middleware := CORSWithConfig(corsConfig)

		handler := func(c *context.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Create request from non-allowed origin
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")

		rr := httptest.NewRecorder()
		c := &context.Context{
			Request: req,
			Writer:  rr,
		}

		wrappedHandler(c)

		// Should still process the request but without CORS headers
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		// Should not have CORS headers
		if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
			t.Errorf("Expected no Access-Control-Allow-Origin header, got '%s'", origin)
		}
	})

	t.Run("wildcard origin with credentials should fail", func(t *testing.T) {
		corsConfig := CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowCredentials: true,
		}

		middleware := CORSWithConfig(corsConfig)

		handler := func(c *context.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "success"})
		}

		wrappedHandler := middleware(handler)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://example.com")

		rr := httptest.NewRecorder()
		c := &context.Context{
			Request: req,
			Writer:  rr,
		}

		wrappedHandler(c)

		// Should not allow credentials with wildcard origin
		credentials := rr.Header().Get("Access-Control-Allow-Credentials")
		if credentials == "true" {
			t.Error("Should not allow credentials with wildcard origin")
		}
	})
}

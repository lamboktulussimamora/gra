package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
)

func TestNew(t *testing.T) {
	r := New()

	if r == nil {
		t.Fatal("New() returned nil")
	}

	if r.routes == nil {
		t.Error("Router routes not initialized")
	}

	if r.middlewares == nil {
		t.Error("Router middlewares not initialized")
	}

	if r.notFound == nil {
		t.Error("Router notFound handler not initialized")
	}

	if r.methodNotAllowed == nil {
		t.Error("Router methodNotAllowed handler not initialized")
	}
}

func TestHandleAndHTTPMethods(t *testing.T) {
	r := New()

	dummyHandler := func(c *context.Context) {}

	// Test Handle method
	r.Handle("GET", "/test", dummyHandler)

	if len(r.routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(r.routes))
	}

	if r.routes[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", r.routes[0].Method)
	}

	if r.routes[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", r.routes[0].Path)
	}

	// Test HTTP method convenience functions
	testCases := []struct {
		method      string
		addRoute    func(string, HandlerFunc)
		expectedLen int
	}{
		{"GET", r.GET, 2},
		{"POST", r.POST, 3},
		{"PUT", r.PUT, 4},
		{"DELETE", r.DELETE, 5},
	}

	for _, tc := range testCases {
		tc.addRoute("/"+tc.method, dummyHandler)

		if len(r.routes) != tc.expectedLen {
			t.Errorf("Expected %d routes after adding %s route, got %d",
				tc.expectedLen, tc.method, len(r.routes))
		}

		lastRoute := r.routes[len(r.routes)-1]
		if lastRoute.Method != tc.method {
			t.Errorf("Expected method %s, got %s", tc.method, lastRoute.Method)
		}

		if lastRoute.Path != "/"+tc.method {
			t.Errorf("Expected path /%s, got %s", tc.method, lastRoute.Path)
		}
	}
}

func TestUse(t *testing.T) {
	r := New()

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			next(c)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			next(c)
		}
	}

	// Test adding a single middleware
	r.Use(middleware1)

	if len(r.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(r.middlewares))
	}

	// Test adding multiple middlewares
	r.Use(middleware1, middleware2)

	if len(r.middlewares) != 3 {
		t.Errorf("Expected 3 middlewares, got %d", len(r.middlewares))
	}
}

func TestSetNotFound(t *testing.T) {
	r := New()

	customHandler := func(c *context.Context) {
		c.Status(http.StatusNotFound).JSON(http.StatusNotFound, map[string]string{
			"error": "Custom not found",
		})
	}

	r.SetNotFound(customHandler)

	// Verify handler was set by calling it and checking the response
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notfound", nil)
	c := context.New(w, req)

	r.notFound(c)

	if w.Code != http.StatusNotFound {
		t.Error("SetNotFound did not set the handler correctly")
	}
}

func TestSetMethodNotAllowed(t *testing.T) {
	r := New()

	customHandler := func(c *context.Context) {
		c.Status(http.StatusMethodNotAllowed).JSON(http.StatusMethodNotAllowed, map[string]string{
			"error": "Custom method not allowed",
		})
	}

	r.SetMethodNotAllowed(customHandler)

	// Verify handler was set by calling it and checking the response
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/methodnotallowed", nil)
	c := context.New(w, req)

	r.methodNotAllowed(c)

	if w.Code != http.StatusMethodNotAllowed {
		t.Error("SetMethodNotAllowed did not set the handler correctly")
	}
}

func TestPathMatch(t *testing.T) {
	testCases := []struct {
		name           string
		routePath      string
		requestPath    string
		shouldMatch    bool
		expectedParams map[string]string
	}{
		{
			name:           "Exact match",
			routePath:      "/users/profile",
			requestPath:    "/users/profile",
			shouldMatch:    true,
			expectedParams: map[string]string{},
		},
		{
			name:           "Single parameter",
			routePath:      "/users/:id",
			requestPath:    "/users/123",
			shouldMatch:    true,
			expectedParams: map[string]string{"id": "123"},
		},
		{
			name:           "Multiple parameters",
			routePath:      "/users/:id/posts/:postId",
			requestPath:    "/users/123/posts/456",
			shouldMatch:    true,
			expectedParams: map[string]string{"id": "123", "postId": "456"},
		},
		{
			name:           "No match - different segment count",
			routePath:      "/users/profile",
			requestPath:    "/users/profile/settings",
			shouldMatch:    false,
			expectedParams: nil,
		},
		{
			name:           "No match - different path",
			routePath:      "/users/profile",
			requestPath:    "/users/settings",
			shouldMatch:    false,
			expectedParams: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, params := pathMatch(tc.routePath, tc.requestPath)

			if match != tc.shouldMatch {
				t.Errorf("Expected match to be %v, got %v", tc.shouldMatch, match)
			}

			if !match {
				return
			}

			if len(params) != len(tc.expectedParams) {
				t.Errorf("Expected %d parameters, got %d", len(tc.expectedParams), len(params))
			}

			for key, expectedValue := range tc.expectedParams {
				if value, ok := params[key]; !ok || value != expectedValue {
					t.Errorf("Expected param %s to be %s, got %s", key, expectedValue, value)
				}
			}
		})
	}
}

func TestChain(t *testing.T) {
	order := []string{}

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			order = append(order, "middleware1 before")
			next(c)
			order = append(order, "middleware1 after")
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			order = append(order, "middleware2 before")
			next(c)
			order = append(order, "middleware2 after")
		}
	}

	handler := func(c *context.Context) {
		order = append(order, "handler")
	}

	chainedHandler := Chain(middleware1, middleware2)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := context.New(w, r)

	chainedHandler(c)

	expectedOrder := []string{
		"middleware1 before",
		"middleware2 before",
		"handler",
		"middleware2 after",
		"middleware1 after",
	}

	if len(order) != len(expectedOrder) {
		t.Fatalf("Expected %d items in execution order, got %d", len(expectedOrder), len(order))
	}

	for i, item := range expectedOrder {
		if order[i] != item {
			t.Errorf("Expected item %d to be %s, got %s", i, item, order[i])
		}
	}
}

func TestServeHTTP(t *testing.T) {
	r := New()

	// Track handler execution
	handlerExecuted := false
	notFoundExecuted := false
	methodNotAllowedExecuted := false

	// Setup test handlers
	testHandler := func(c *context.Context) {
		handlerExecuted = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}

	r.SetNotFound(func(c *context.Context) {
		notFoundExecuted = true
		c.Status(http.StatusNotFound).JSON(http.StatusNotFound, map[string]string{
			"error": "not found",
		})
	})

	r.SetMethodNotAllowed(func(c *context.Context) {
		methodNotAllowedExecuted = true
		c.Status(http.StatusMethodNotAllowed).JSON(http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	})

	// Register routes
	r.GET("/users/:id", testHandler)
	r.POST("/users", testHandler)

	// Test cases
	testCases := []struct {
		name            string
		method          string
		path            string
		expectedStatus  int
		expectedHandler *bool
		expectedParams  map[string]string
	}{
		{
			name:            "Exact route match",
			method:          "GET",
			path:            "/users/123",
			expectedStatus:  http.StatusOK,
			expectedHandler: &handlerExecuted,
			expectedParams:  map[string]string{"id": "123"},
		},
		{
			name:            "Route not found",
			method:          "GET",
			path:            "/unknown",
			expectedStatus:  http.StatusNotFound,
			expectedHandler: &notFoundExecuted,
		},
		{
			name:            "Method not allowed",
			method:          "PUT",
			path:            "/users/123",
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedHandler: &methodNotAllowedExecuted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			handlerExecuted = false
			notFoundExecuted = false
			methodNotAllowedExecuted = false

			// Create test request and recorder
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)

			// Execute request
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check that the expected handler was executed
			if tc.expectedHandler != nil && !*tc.expectedHandler {
				t.Error("Expected handler was not executed")
			}
		})
	}
}

func TestServeHTTPWithMiddleware(t *testing.T) {
	r := New()

	// Track middleware and handler execution
	middlewareExecuted := false
	handlerExecuted := false

	// Setup test middleware
	testMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c *context.Context) {
			middlewareExecuted = true
			c.WithValue("key", "value")
			next(c)
		}
	}

	// Setup test handler
	testHandler := func(c *context.Context) {
		handlerExecuted = true

		// Check that middleware set the context value
		value := c.Value("key")
		if value != "value" {
			t.Errorf("Expected middleware to set context value 'value', got %v", value)
		}

		c.Status(http.StatusOK)
	}

	// Register middleware and route
	r.Use(testMiddleware)
	r.GET("/test", testHandler)

	// Create test request and recorder
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Execute request
	r.ServeHTTP(w, req)

	// Check middleware and handler execution
	if !middlewareExecuted {
		t.Error("Middleware was not executed")
	}

	if !handlerExecuted {
		t.Error("Handler was not executed")
	}

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

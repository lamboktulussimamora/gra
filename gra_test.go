package gra

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	// Test routes/paths
	testPath = "/test"

	// HTTP methods
	methodGet = "GET"

	// Headers
	headerXTest = "X-Test"
	headerValue = "middleware"

	// Error messages
	errNewReturnedNil     = "New() returned nil"
	errVersionEmpty       = "Version should not be empty"
	errStatusCodeMismatch = "Expected status code %d, got %d"
	errHeaderNotSet       = "Middleware didn't set the expected header"
	errHandlerNotCalled   = "Handler was not called"
	errRunUnexpected      = "Run returned an unexpected error: %v"

	// Test payload
	testSuccessMessage = "Test message"
	testEndpointMsg    = "Success"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal(errNewReturnedNil)
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error(errVersionEmpty)
	}
}

func TestAliases(t *testing.T) {
	// Test Context alias
	w := httptest.NewRecorder()
	r := httptest.NewRequest(methodGet, testPath, nil)
	c := &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
	}

	// Test methods on the Context alias
	c.Success(http.StatusOK, testSuccessMessage, map[string]any{
		"time": time.Now(),
	})

	if w.Code != http.StatusOK {
		t.Errorf(errStatusCodeMismatch, http.StatusOK, w.Code)
	}

	// Test HandlerFunc alias
	var fn HandlerFunc = func(c *Context) {
		c.Status(http.StatusOK)
	}

	// Create a new context to test the handler
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(methodGet, testPath, nil)
	c2 := &Context{
		Writer:  w2,
		Request: r2,
		Params:  make(map[string]string),
	}

	fn(c2)

	if w2.Code != http.StatusOK {
		t.Errorf(errStatusCodeMismatch, http.StatusOK, w2.Code)
	}

	// Test Middleware alias by creating a simple middleware
	var middleware Middleware = func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			// Add a header before calling the next handler
			c.Writer.Header().Set(headerXTest, headerValue)
			next(c)
		}
	}

	wrappedFn := middleware(fn)

	// Create a new context to test the middleware
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest(methodGet, testPath, nil)
	c3 := &Context{
		Writer:  w3,
		Request: r3,
		Params:  make(map[string]string),
	}

	wrappedFn(c3)

	if w3.Code != http.StatusOK {
		t.Errorf(errStatusCodeMismatch, http.StatusOK, w3.Code)
	}

	if w3.Header().Get(headerXTest) != headerValue {
		t.Error(errHeaderNotSet)
	}
}

// TestRunWithMockServer tests the Run function with a mock server
func TestRunWithMockServer(t *testing.T) {
	// This is a bit tricky to test directly since it blocks
	// Instead, we'll verify that we can create a valid server

	r := New()

	// Configure a test route
	called := false
	r.GET(testPath, func(c *Context) {
		called = true
		c.Success(http.StatusOK, testEndpointMsg, nil)
	})

	// Create a test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest(methodGet, testPath, nil)

	// Serve the request directly instead of starting a server
	r.ServeHTTP(w, req)

	// Check the response
	if !called {
		t.Error(errHandlerNotCalled)
	}

	if w.Code != http.StatusOK {
		t.Errorf(errStatusCodeMismatch, http.StatusOK, w.Code)
	}
}

// TestRun tests the Run function
func TestRun(t *testing.T) {
	r := New()

	// Start a server in a goroutine so it doesn't block
	go func() {
		err := Run(":0", r)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf(errRunUnexpected, err)
		}
	}()

	// Create a test server that uses our router
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Configure a test route
	r.GET(testPath, func(c *Context) {
		c.Success(http.StatusOK, testEndpointMsg, nil)
	})
}

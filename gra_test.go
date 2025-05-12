package gra

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestAliases(t *testing.T) {
	// Test Context alias
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	c := &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
	}

	// Test methods on the Context alias
	c.Success(http.StatusOK, "Test message", map[string]any{
		"time": time.Now(),
	})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Test HandlerFunc alias
	var fn HandlerFunc = func(c *Context) {
		c.Status(http.StatusOK)
	}

	// Create a new context to test the handler
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/test", nil)
	c2 := &Context{
		Writer:  w2,
		Request: r2,
		Params:  make(map[string]string),
	}

	fn(c2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w2.Code)
	}

	// Test Middleware alias by creating a simple middleware
	var middleware Middleware = func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			// Add a header before calling the next handler
			c.Writer.Header().Set("X-Test", "middleware")
			next(c)
		}
	}

	wrappedFn := middleware(fn)

	// Create a new context to test the middleware
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/test", nil)
	c3 := &Context{
		Writer:  w3,
		Request: r3,
		Params:  make(map[string]string),
	}

	wrappedFn(c3)

	if w3.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w3.Code)
	}

	if w3.Header().Get("X-Test") != "middleware" {
		t.Error("Middleware didn't set the expected header")
	}
}

// TestRunWithMockServer tests the Run function with a mock server
func TestRunWithMockServer(t *testing.T) {
	// This is a bit tricky to test directly since it blocks
	// Instead, we'll verify that we can create a valid server

	r := New()

	// Configure a test route
	called := false
	r.GET("/test", func(c *Context) {
		called = true
		c.Success(http.StatusOK, "Success", nil)
	})

	// Create a test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Serve the request directly instead of starting a server
	r.ServeHTTP(w, req)

	// Check the response
	if !called {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

// TestRun tests the Run function
func TestRun(t *testing.T) {
	r := New()

	// Start a server in a goroutine so it doesn't block
	go func() {
		err := Run(":0", r)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Run returned an unexpected error: %v", err)
		}
	}()

	// Create a test server that uses our router
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Configure a test route
	r.GET("/test", func(c *Context) {
		c.Success(http.StatusOK, "Success", nil)
	})
}

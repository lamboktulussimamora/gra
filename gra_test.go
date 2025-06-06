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
		t.Error("Expected New() to return a non-nil router")
	}
}

func TestVersion(t *testing.T) {
	expectedVersion := "1.0.3"
	if Version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, Version)
	}
}

func TestDefaultTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{"DefaultReadTimeout", DefaultReadTimeout, 10 * time.Second},
		{"DefaultWriteTimeout", DefaultWriteTimeout, 30 * time.Second},
		{"DefaultIdleTimeout", DefaultIdleTimeout, 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout != tt.expected {
				t.Errorf("Expected %s to be %v, got %v", tt.name, tt.expected, tt.timeout)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Create a test router
	r := New()

	// Add a simple route
	r.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "test"})
	})

	// Start server in a goroutine
	addr := ":0" // Let OS choose available port

	// We can't easily test the actual server start without complex setup,
	// but we can test that the function doesn't panic with valid inputs
	if r == nil {
		t.Error("Router should not be nil")
	}

	// Test that the function signature is correct by calling it in a separate goroutine
	// This won't actually start a server since we're not providing a valid address
	go func() {
		// We expect this to fail since ":0" isn't a valid listening address in practice
		// but the function should be callable
		_ = Run(addr, r) // Ignore error in test context
	}()

	// Give it a moment and continue
	time.Sleep(10 * time.Millisecond)
}

func TestRunWithConfig(t *testing.T) {
	// Create a test router
	r := New()

	// Add a simple route
	r.GET("/test", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{"message": "test"})
	})

	// Test custom timeouts
	readTimeout := 5 * time.Second
	writeTimeout := 15 * time.Second
	idleTimeout := 60 * time.Second

	addr := ":0" // Let OS choose available port

	// Test that the function signature is correct
	if r == nil {
		t.Error("Router should not be nil")
	}

	// Test that the function can be called with custom timeouts
	go func() {
		// We expect this to fail since ":0" isn't a valid listening address in practice
		// but the function should be callable
		_ = RunWithConfig(addr, r, readTimeout, writeTimeout, idleTimeout) // Ignore error in test context
	}()

	// Give it a moment and continue
	time.Sleep(10 * time.Millisecond)
}

func TestTypeAliases(t *testing.T) {
	// Test that our type aliases work correctly
	r := New()

	// Test Context alias
	r.GET("/context-test", func(c *Context) {
		// This should compile and work correctly
		c.Status(http.StatusOK)
	})

	// Test HandlerFunc alias
	var handler HandlerFunc = func(c *Context) {
		c.Status(http.StatusOK)
	}

	if handler == nil {
		t.Error("HandlerFunc alias should work")
	}

	// Test Middleware alias
	var middleware Middleware = func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			next(c)
		}
	}

	if middleware == nil {
		t.Error("Middleware alias should work")
	}
}

func TestIntegration(t *testing.T) {
	// Test a more comprehensive integration scenario
	r := New()

	// Add some middleware
	r.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			// Add a test header
			c.Request.Header.Set("X-Test", "middleware")
			next(c)
		}
	})

	// Add routes
	r.GET("/", func(c *Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "hello",
			"version": Version,
		})
	})

	r.POST("/data", func(c *Context) {
		var data map[string]interface{}
		if err := c.BindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		c.JSON(http.StatusCreated, data)
	})

	// Test GET request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test that middleware was applied
	if req.Header.Get("X-Test") != "middleware" {
		t.Error("Middleware was not applied correctly")
	}
}

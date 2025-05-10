package adapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

func TestHTTPHandler(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"message": "success",
		})
	}

	// Convert to http.HandlerFunc
	httpHandlerFunc := HTTPHandler(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Call the handler
	httpHandlerFunc(w, r)

	// Verify it was called
	if !handlerCalled {
		t.Error("Handler function was not called")
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandlerAdapter_ServeHTTP(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"message": "success",
		})
	}

	// Create HandlerAdapter
	adapter := HandlerAdapter(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Call ServeHTTP
	adapter.ServeHTTP(w, r)

	// Verify the handler was called
	if !handlerCalled {
		t.Error("Handler function was not called")
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAsHTTPHandler(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"message": "success",
		})
	}

	// Convert to http.Handler
	httpHandler := AsHTTPHandler(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Call ServeHTTP
	httpHandler.ServeHTTP(w, r)

	// Verify the handler was called
	if !handlerCalled {
		t.Error("Handler function was not called")
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Verify that AsHTTPHandler returns a HandlerAdapter
	_, ok := httpHandler.(HandlerAdapter)
	if !ok {
		t.Error("AsHTTPHandler should return a HandlerAdapter")
	}
}

func TestHandlerChain(t *testing.T) {
	// Track the order of execution
	executionOrder := []string{}

	// Create middleware
	middleware := func(c *context.Context) {
		executionOrder = append(executionOrder, "middleware")
		c.WithValue("key", "value")
	}

	// Create a test router.HandlerFunc
	testHandler := func(c *context.Context) {
		executionOrder = append(executionOrder, "handler")
		
		// Check if middleware data was passed correctly
		value := c.Value("key")
		if value != "value" {
			t.Errorf("Expected middleware to set context value 'value', got %v", value)
		}
		
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			"message": "success",
		})
	}

	// Create a middleware wrapped handler
	middlewareFunc := func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			middleware(c)
			next(c)
		}
	}

	wrappedHandler := middlewareFunc(testHandler)

	// Convert to http.Handler
	httpHandler := AsHTTPHandler(wrappedHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	// Call ServeHTTP
	httpHandler.ServeHTTP(w, r)

	// Verify execution order
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 execution steps, got %d", len(executionOrder))
	}

	if len(executionOrder) >= 1 && executionOrder[0] != "middleware" {
		t.Errorf("Expected middleware to execute first, got %s", executionOrder[0])
	}

	if len(executionOrder) >= 2 && executionOrder[1] != "handler" {
		t.Errorf("Expected handler to execute second, got %s", executionOrder[1])
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

package adapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

const (
	// Test constants
	testEndpoint     = "/test"
	testMethod       = "GET"
	testMessageKey   = "message"
	testMessageValue = "success"
	testContextKey   = "key"
	testContextValue = "value"

	// Error messages
	errHandlerNotCalled = "Handler function was not called"
	errStatusCode       = "Expected status code %d, got %d"
	errExecutionSteps   = "Expected 2 execution steps, got %d"
	errExecutionOrder   = "Expected %s to execute %s, got %s"
	errMiddlewareValue  = "Expected middleware to set context value '%s', got %v"
	errWrongHandlerType = "AsHTTPHandler should return a HandlerAdapter"
)

func TestHTTPHandler(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			testMessageKey: testMessageValue,
		})
	}

	// Convert to http.HandlerFunc
	httpHandlerFunc := HTTPHandler(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(testMethod, testEndpoint, nil)

	// Call the handler
	httpHandlerFunc(w, r)

	// Verify it was called
	if !handlerCalled {
		t.Error(errHandlerNotCalled)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}
}

func TestHandlerAdapterServeHTTP(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			testMessageKey: testMessageValue,
		})
	}

	// Create HandlerAdapter
	adapter := HandlerAdapter(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(testMethod, testEndpoint, nil)

	// Call ServeHTTP
	adapter.ServeHTTP(w, r)

	// Verify the handler was called
	if !handlerCalled {
		t.Error(errHandlerNotCalled)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}
}

func TestAsHTTPHandler(t *testing.T) {
	// Create a test router.HandlerFunc
	handlerCalled := false
	testHandler := func(c *context.Context) {
		handlerCalled = true
		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			testMessageKey: testMessageValue,
		})
	}

	// Convert to http.Handler
	httpHandler := AsHTTPHandler(testHandler)

	// Create a test request and response recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest(testMethod, testEndpoint, nil)

	// Call ServeHTTP
	httpHandler.ServeHTTP(w, r)

	// Verify the handler was called
	if !handlerCalled {
		t.Error(errHandlerNotCalled)
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}

	// Verify that AsHTTPHandler returns a HandlerAdapter
	_, ok := httpHandler.(HandlerAdapter)
	if !ok {
		t.Error(errWrongHandlerType)
	}
}

func TestHandlerChain(t *testing.T) {
	// Track the order of execution
	executionOrder := []string{}

	// Create middleware
	middleware := func(c *context.Context) {
		executionOrder = append(executionOrder, "middleware")
		c.WithValue(testContextKey, testContextValue)
	}

	// Create a test router.HandlerFunc
	testHandler := func(c *context.Context) {
		executionOrder = append(executionOrder, "handler")

		// Check if middleware data was passed correctly
		value := c.Value(testContextKey)
		if value != testContextValue {
			t.Errorf(errMiddlewareValue, testContextValue, value)
		}

		c.Status(http.StatusOK).JSON(http.StatusOK, map[string]string{
			testMessageKey: testMessageValue,
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
	r := httptest.NewRequest(testMethod, testEndpoint, nil)

	// Call ServeHTTP
	httpHandler.ServeHTTP(w, r)

	// Verify execution order
	if len(executionOrder) != 2 {
		t.Errorf(errExecutionSteps, len(executionOrder))
	}

	if len(executionOrder) >= 1 && executionOrder[0] != "middleware" {
		t.Errorf(errExecutionOrder, "middleware", "first", executionOrder[0])
	}

	if len(executionOrder) >= 2 && executionOrder[1] != "handler" {
		t.Errorf(errExecutionOrder, "handler", "second", executionOrder[1])
	}

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf(errStatusCode, http.StatusOK, w.Code)
	}
}

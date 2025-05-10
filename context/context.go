// Package context provides the Context type for handling HTTP requests and responses.
package context

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// APIResponse is a standardized response structure
type APIResponse struct {
	Status  string      `json:"status"`          // "success" or "error"
	Message string      `json:"message"`         // Human-readable message
	Data    interface{} `json:"data,omitempty"`  // Optional data payload
	Error   string      `json:"error,omitempty"` // Error message if status is "error"
}

// Context wraps the HTTP request and response
// It provides helper methods for handling requests and responses
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string // For route parameters
	ctx     context.Context
}

// New creates a new Context
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		ctx:     r.Context(),
	}
}

// Status sets the HTTP status code
func (c *Context) Status(code int) *Context {
	c.Writer.WriteHeader(code)
	return c
}

// JSON sends a JSON response
func (c *Context) JSON(status int, obj interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// BindJSON binds JSON request body to a struct
func (c *Context) BindJSON(obj interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	defer c.Request.Body.Close()

	return json.Unmarshal(body, obj)
}

// Success sends a success response
func (c *Context) Success(status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func (c *Context) Error(status int, errorMsg string) {
	c.JSON(status, APIResponse{
		Status: "error",
		Error:  errorMsg,
	})
}

// GetParam gets a path parameter value
func (c *Context) GetParam(key string) string {
	return c.Params[key]
}

// GetQuery gets a query parameter value
func (c *Context) GetQuery(key string) string {
	return c.Request.URL.Query().Get(key)
}

// WithValue adds a value to the request context
func (c *Context) WithValue(key, value interface{}) *Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.Request = c.Request.WithContext(c.ctx)
	return c
}

// Value gets a value from the request context
func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

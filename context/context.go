// Package context provides the Context type for handling HTTP requests and responses.
package context

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// HTTP header constants
const (
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"

	ContentTypeJSON = "application/json"
)

// APIResponse is a standardized response structure
type APIResponse struct {
	Status  string `json:"status"`          // "success" or "error"
	Message string `json:"message"`         // Human-readable message
	Data    any    `json:"data,omitempty"`  // Optional data payload
	Error   string `json:"error,omitempty"` // Error message if status is "error"
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
func (c *Context) JSON(status int, obj any) {
	c.Writer.Header().Set(HeaderContentType, ContentTypeJSON)
	c.Writer.WriteHeader(status)
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// BindJSON binds JSON request body to a struct
func (c *Context) BindJSON(obj any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := c.Request.Body.Close(); cerr != nil {
			log.Printf("Error closing request body: %v", cerr)
		}
	}()

	return json.Unmarshal(body, obj)
}

// Success sends a success response
func (c *Context) Success(status int, message string, data any) {
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

// JSONData sends a JSON response with just the data without wrapping it in APIResponse.
// Use this when you want to return only the data payload directly, for example:
// - When you need to conform to a specific API format expected by a client
// - When you want to return an array directly in the response body
// - When integrating with systems that expect a simple JSON structure
func (c *Context) JSONData(status int, data any) {
	c.Writer.Header().Set(HeaderContentType, ContentTypeJSON)
	c.Writer.WriteHeader(status)
	if err := json.NewEncoder(c.Writer).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// WithValue adds a value to the request context
func (c *Context) WithValue(key, value any) *Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.Request = c.Request.WithContext(c.ctx)
	return c
}

// Value gets a value from the request context
func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}

// GetHeader gets a header value from the request
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a header value in the response
func (c *Context) SetHeader(key, value string) *Context {
	c.Writer.Header().Set(key, value)
	return c
}

// GetCookie gets a cookie from the request
func (c *Context) GetCookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetCookie sets a cookie in the response
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) *Context {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
	return c
}

// GetContentType gets the Content-Type header
func (c *Context) GetContentType() string {
	return c.GetHeader(HeaderContentType)
}

// Redirect redirects the request to a new URL
func (c *Context) Redirect(status int, url string) {
	http.Redirect(c.Writer, c.Request, url, status)
}

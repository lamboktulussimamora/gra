// Package gra provides a lightweight HTTP framework for building web applications.
//
// GRA is a minimalist web framework inspired by Gin, designed for building
// clean architecture applications in Go. It includes a Context object for handling
// requests and responses, a Router for URL routing, middleware support, and validation
// utilities.
package gra

import (
	"net/http"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// Version is the current version of the framework
const Version = "1.0.3"

// New creates a new router with default configuration
func New() *router.Router {
	return router.New()
}

// Run starts the HTTP server with the given router
func Run(addr string, r *router.Router) error {
	return http.ListenAndServe(addr, r)
}

// Context is an alias for context.Context
type Context = context.Context

// HandlerFunc is an alias for router.HandlerFunc
type HandlerFunc = router.HandlerFunc

// Middleware is an alias for router.Middleware
type Middleware = router.Middleware

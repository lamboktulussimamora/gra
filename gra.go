// Package gra provides a lightweight HTTP framework for building web applications.
//
// GRA is a minimalist web framework inspired by Gin, designed for building
// clean architecture applications in Go. It includes a Context object for handling
// requests and responses, a Router for URL routing, middleware support, and validation
// utilities.
package gra

import (
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// Version is the current version of the framework
const Version = "1.0.3"

// New creates a new router with default configuration
func New() *router.Router {
	return router.New()
}

// Default timeout values for the HTTP server
const (
	// DefaultReadTimeout is the maximum duration for reading the entire request
	DefaultReadTimeout = 10 * time.Second

	// DefaultWriteTimeout is the maximum duration for writing the response
	DefaultWriteTimeout = 30 * time.Second

	// DefaultIdleTimeout is the maximum duration to wait for the next request
	DefaultIdleTimeout = 120 * time.Second
)

// Run starts the HTTP server with the given router and default timeouts
func Run(addr string, r *router.Router) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	return srv.ListenAndServe()
}

// RunWithConfig starts the HTTP server with custom configuration
func RunWithConfig(addr string, r *router.Router, readTimeout, writeTimeout, idleTimeout time.Duration) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return srv.ListenAndServe()
}

// Context is an alias for context.Context
type Context = context.Context

// HandlerFunc is an alias for router.HandlerFunc
type HandlerFunc = router.HandlerFunc

// Middleware is an alias for router.Middleware
type Middleware = router.Middleware

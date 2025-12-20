// Package gra provides a lightweight HTTP framework for building web applications.
//
// GRA is a minimalist web framework inspired by Gin, designed for building
// clean architecture applications in Go. It includes a Context object for handling
// requests and responses, a Router for URL routing, middleware support, and validation
// utilities.
package gra

import (
	stdctx "context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gcontext "github.com/lamboktulussimamora/gra/context"
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

	// DefaultReadHeaderTimeout is the maximum duration for reading request headers
	DefaultReadHeaderTimeout = 5 * time.Second

	// DefaultWriteTimeout is the maximum duration for writing the response
	DefaultWriteTimeout = 30 * time.Second

	// DefaultIdleTimeout is the maximum duration to wait for the next request
	DefaultIdleTimeout = 120 * time.Second

	// DefaultMaxHeaderBytes is the maximum size of request headers
	DefaultMaxHeaderBytes = 1 << 20 // 1 MiB
)

// Run starts the HTTP server with the given router and default timeouts
func Run(addr string, r *router.Router) error {
	return NewServer(addr, r).ListenAndServe()
}

// RunWithConfig starts the HTTP server with custom configuration
func RunWithConfig(addr string, r *router.Router, readTimeout, writeTimeout, idleTimeout time.Duration) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}
	return srv.ListenAndServe()
}

// NewServer constructs an http.Server with GRA's recommended defaults.
// This is useful when you need graceful shutdown or additional server configuration.
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       DefaultReadTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		WriteTimeout:      DefaultWriteTimeout,
		IdleTimeout:       DefaultIdleTimeout,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}
}

// RunWithGracefulShutdown starts the server and blocks until SIGINT/SIGTERM.
// It then shuts down gracefully with the given timeout.
func RunWithGracefulShutdown(addr string, r *router.Router, shutdownTimeout time.Duration) error {
	srv := NewServer(addr, r)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(stdctx.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err == nil {
			return nil
		}
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := stdctx.WithTimeout(stdctx.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return err
		}
		return nil
	}
}

// Context is an alias for context.Context
type Context = gcontext.Context

// HandlerFunc is an alias for router.HandlerFunc
type HandlerFunc = router.HandlerFunc

// Middleware is an alias for router.Middleware
type Middleware = router.Middleware

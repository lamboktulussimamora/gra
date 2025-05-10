// Package adapter provides adapters for different HTTP handler types.
package adapter

import (
	"net/http"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/router"
)

// HTTPHandler converts a router.HandlerFunc to an http.HandlerFunc
func HTTPHandler(f router.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.New(w, r)
		f(ctx)
	}
}

// HandlerAdapter wraps a router.HandlerFunc to implement http.Handler
type HandlerAdapter router.HandlerFunc

// ServeHTTP implements the http.Handler interface for HandlerAdapter
func (f HandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(w, r)
	router.HandlerFunc(f)(ctx)
}

// AsHTTPHandler converts a router.HandlerFunc to http.Handler
func AsHTTPHandler(f router.HandlerFunc) http.Handler {
	return HandlerAdapter(f)
}

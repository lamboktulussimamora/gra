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

// ServeHTTP implements the http.Handler interface for HandlerFunc
func (f router.HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(w, r)
	f(ctx)
}

// Package router provides HTTP routing capabilities.
package router

import (
	"net/http"
	"strings"

	"github.com/lamboktulussimamora/gra/context"
)

// HandlerFunc defines a function that processes requests using Context
type HandlerFunc func(*context.Context)

// Middleware defines a function that runs before a request handler
type Middleware func(HandlerFunc) HandlerFunc

// Route represents a URL route and its handler
type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

// Router handles HTTP requests and routes them to the appropriate handler
type Router struct {
	routes           []Route
	middlewares      []Middleware
	notFound         HandlerFunc
	methodNotAllowed HandlerFunc
}

// New creates a new router
func New() *Router {
	return &Router{
		routes:      []Route{},
		middlewares: []Middleware{},
		notFound: func(c *context.Context) {
			c.Error(http.StatusNotFound, "Not found")
		},
		methodNotAllowed: func(c *context.Context) {
			c.Error(http.StatusMethodNotAllowed, "Method not allowed")
		},
	}
}

// Use adds middleware to the router
func (r *Router) Use(middleware ...Middleware) {
	r.middlewares = append(r.middlewares, middleware...)
}

// Handle registers a new route with the router
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

// GET registers a new GET route
func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

// POST registers a new POST route
func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT registers a new PUT route
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle(http.MethodPut, path, handler)
}

// DELETE registers a new DELETE route
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

// SetNotFound sets the not found handler
func (r *Router) SetNotFound(handler HandlerFunc) {
	r.notFound = handler
}

// SetMethodNotAllowed sets the method not allowed handler
func (r *Router) SetMethodNotAllowed(handler HandlerFunc) {
	r.methodNotAllowed = handler
}

// pathMatch checks if the request path matches a route path
// and extracts path parameters
func pathMatch(routePath, requestPath string) (bool, map[string]string) {
	routeParts := strings.Split(routePath, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(routeParts) != len(requestParts) {
		return false, nil
	}

	params := make(map[string]string)

	for i, routePart := range routeParts {
		if len(routePart) > 0 && routePart[0] == ':' {
			// This is a path parameter
			paramName := routePart[1:]
			params[paramName] = requestParts[i]
		} else if routePart != requestParts[i] {
			// Not a parameter and doesnt match
			return false, nil
		}
	}

	return true, params
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find route
	var handler HandlerFunc
	var params map[string]string

	matchedRoutes := []Route{}

	for _, route := range r.routes {
		if match, pathParams := pathMatch(route.Path, req.URL.Path); match {
			if route.Method == req.Method {
				handler = route.Handler
				params = pathParams
				break
			} else {
				matchedRoutes = append(matchedRoutes, route)
			}
		}
	}

	// If no handler was found but we matched some routes with a different method,
	// its a method not allowed
	if handler == nil && len(matchedRoutes) > 0 {
		handler = r.methodNotAllowed
	}

	// If no handler was found at all, use the not found handler
	if handler == nil {
		handler = r.notFound
	}

	// Create context
	c := context.New(w, req)
	c.Params = params

	// Apply middlewares
	if len(r.middlewares) > 0 {
		handler = Chain(r.middlewares...)(handler)
	}

	// Execute handler
	handler(c)
}

// Chain creates a chain of middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

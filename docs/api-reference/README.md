# API Reference

This section provides detailed documentation for all the components and functions available in the GRA framework.

## Table of Contents

- [Core Package](#core-package)
- [Router Package](#router-package)
- [Context Package](#context-package)
- [Middleware Package](#middleware-package)
- [JWT Package](#jwt-package)
- [Cache Package](#cache-package)
- [Validator Package](#validator-package)
- [Logger Package](#logger-package)
- [Versioning Package](#versioning-package)
- [Adapter Package](#adapter-package)

## Core Package

The core package provides the main functionality of the GRA framework.

### Functions

#### `New() *Router`

Creates a new router instance.

```go
r := gra.New()
```

#### `Run(addr string, handlers ...HandlerFunc) error`

Starts the HTTP server with the given address and optional middleware.

```go
gra.Run(":8080", r)
```

## Router Package

The router package handles HTTP request routing.

### Types

#### `Router`

Main router type that manages routes and middleware.

##### Methods

- `Use(...HandlerFunc) *Router` - Adds middleware to the router
- `Group(prefix string) *Group` - Creates a new route group with the given prefix
- `GET(path string, handlers ...HandlerFunc)` - Adds a GET route
- `POST(path string, handlers ...HandlerFunc)` - Adds a POST route
- `PUT(path string, handlers ...HandlerFunc)` - Adds a PUT route
- `DELETE(path string, handlers ...HandlerFunc)` - Adds a DELETE route
- `PATCH(path string, handlers ...HandlerFunc)` - Adds a PATCH route
- `OPTIONS(path string, handlers ...HandlerFunc)` - Adds an OPTIONS route
- `HEAD(path string, handlers ...HandlerFunc)` - Adds a HEAD route
- `Any(path string, handlers ...HandlerFunc)` - Adds a route for all HTTP methods

#### `Group`

A group of routes with a common prefix.

##### Methods

- `Use(...HandlerFunc) *Group` - Adds middleware to the group
- `Group(prefix string) *Group` - Creates a nested group
- `GET(path string, handlers ...HandlerFunc)` - Adds a GET route to the group
- `POST(path string, handlers ...HandlerFunc)` - Adds a POST route to the group
- `PUT(path string, handlers ...HandlerFunc)` - Adds a PUT route to the group
- `DELETE(path string, handlers ...HandlerFunc)` - Adds a DELETE route to the group
- `PATCH(path string, handlers ...HandlerFunc)` - Adds a PATCH route to the group
- `OPTIONS(path string, handlers ...HandlerFunc)` - Adds an OPTIONS route to the group
- `HEAD(path string, handlers ...HandlerFunc)` - Adds a HEAD route to the group
- `Any(path string, handlers ...HandlerFunc)` - Adds a route for all HTTP methods to the group

## Context Package

The context package provides HTTP request and response context.

### Types

#### `Context`

Contains request and response information.

##### Methods

- `GetParam(name string) string` - Gets a path parameter
- `GetQuery(name string, defaultValue ...string) string` - Gets a query parameter
- `GetHeader(key string) string` - Gets a request header
- `GetCookie(name string) (*http.Cookie, error)` - Gets a request cookie
- `BindJSON(obj interface{}) error` - Parses the request body as JSON
- `Set(key string, value interface{})` - Sets a value in the context
- `Get(key string) (interface{}, bool)` - Gets a value from the context
- `MustGet(key string) interface{}` - Gets a value from the context or panics
- `JSON(code int, obj interface{})` - Sends a JSON response
- `String(code int, format string, values ...interface{})` - Sends a text response
- `File(filepath string)` - Sends a file response
- `Redirect(code int, location string)` - Redirects to another location
- `Success(code int, message string, data interface{})` - Sends a standardized success response
- `Error(code int, message string, errors ...string)` - Sends a standardized error response
- `Next()` - Calls the next middleware or handler in the chain
- `Abort()` - Stops the middleware chain
- `IsAborted() bool` - Checks if the middleware chain was aborted

## Middleware Package

The middleware package provides pre-built middleware functions.

### Functions

#### `Logger() HandlerFunc`

Creates a middleware for logging request/response information.

```go
r.Use(middleware.Logger())
```

#### `Recovery() HandlerFunc`

Creates a middleware that recovers from panics.

```go
r.Use(middleware.Recovery())
```

#### `CORS(allowOrigin string) HandlerFunc`

Creates a middleware for CORS support.

```go
r.Use(middleware.CORS("*"))
```

#### `SecureHeaders() HandlerFunc`

Creates a middleware that adds security-related HTTP headers.

```go
r.Use(middleware.SecureHeaders())
```

#### `Auth(jwtService *jwt.Service, userType string) HandlerFunc`

Creates a middleware for JWT authentication.

```go
r.Use(middleware.Auth(jwtService, "user"))
```

#### `Cache(duration time.Duration) HandlerFunc`

Creates a middleware for response caching.

```go
r.Use(middleware.Cache(5 * time.Minute))
```

## JWT Package

Documentation for the JWT authentication functionality.

## Cache Package

Documentation for the caching functionality.

## Validator Package

Documentation for input validation functionality.

## Logger Package

Documentation for the logging functionality.

## Versioning Package

Documentation for API versioning functionality.

## Adapter Package

Documentation for adapting external functions to GRA handlers.

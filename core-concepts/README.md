# GRA Framework Core Concepts

This section explains the fundamental concepts and architecture of the GRA framework.

## Framework Architecture

The GRA framework is built around a few key components that work together to handle HTTP requests efficiently:

<img src="../assets/images/architecture-diagram.svg" alt="GRA Framework Architecture" width="700" />

## Table of Contents

- [Router](#router)
- [Context](#context)
- [Middleware](#middleware)
- [Request Handling](#request-handling)
- [Response Format](#response-format)
- [Error Handling](#error-handling)

## Router

The router is the core component of GRA that maps HTTP requests to handler functions. It supports:

- HTTP method-based routing (GET, POST, PUT, DELETE, etc.)
- Path parameters using the `:param` syntax
- Route grouping for organizing endpoints
- Middleware application at global, group, or route level

```go
// Create a new router
r := gra.New()

// Register routes
r.GET("/users", listUsers)
r.POST("/users", createUser)
r.GET("/users/:id", getUser)
r.PUT("/users/:id", updateUser)
r.DELETE("/users/:id", deleteUser)

// Create a route group
api := r.Group("/api")
api.GET("/products", listProducts)
```

## Context

The `Context` object encapsulates HTTP request and response information and provides methods for:

- Reading request data (headers, query parameters, path parameters, body)
- Setting and getting values in the request context
- Returning responses (JSON, text, file, etc.)
- Using standardized response formats
- Error handling

```go
func handleRequest(c *gra.Context) {
    // Get path parameter
    id := c.GetParam("id")
    
    // Get query parameter
    sortBy := c.GetQuery("sortBy", "createdAt")
    
    // Parse JSON request
    var requestData map[string]interface{}
    if err := c.BindJSON(&requestData); err != nil {
        c.Error(http.StatusBadRequest, "Invalid JSON")
        return
    }
    
    // Return standardized success response
    c.Success(http.StatusOK, "Operation successful", map[string]interface{}{
        "id": id,
        "data": requestData,
    })
}
```

## Middleware

Middleware functions provide a way to execute code before or after handler functions. They can:

- Process requests before they reach handlers
- Modify the context
- Short-circuit the request handling
- Process responses after handlers complete

```go
// Define a custom middleware
func AuthMiddleware() gra.HandlerFunc {
    return func(c *gra.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.Error(http.StatusUnauthorized, "Authorization token required")
            return
        }
        
        // Set user info in context for handlers
        c.Set("user", userInfo)
        
        // Continue to the next middleware or handler
        c.Next()
    }
}

// Use middleware
r.Use(AuthMiddleware())
```

## Request Handling

GRA uses handler functions to process requests. A handler function takes a `Context` parameter and returns no value:

```go
func handleGetUser(c *gra.Context) {
    id := c.GetParam("id")
    
    user, err := userService.FindByID(id)
    if err != nil {
        c.Error(http.StatusNotFound, "User not found")
        return
    }
    
    c.Success(http.StatusOK, "User retrieved", user)
}
```

## Response Format

GRA encourages standardized API responses for consistency:

```json
{
    "status": "success",
    "message": "Operation successful",
    "data": {
        "id": 1,
        "name": "Example"
    }
}
```

Or for errors:

```json
{
    "status": "error",
    "message": "Resource not found",
    "errors": ["The requested resource could not be found"]
}
```

## Error Handling

GRA provides built-in error handling through:

- The `c.Error()` method for returning standardized error responses
- Panic recovery middleware to prevent server crashes
- Custom error types and handlers

```go
// Basic error handling
if err != nil {
    c.Error(http.StatusInternalServerError, "Failed to process request")
    return
}

// More detailed error handling
if err == ErrNotFound {
    c.Error(http.StatusNotFound, "Resource not found")
    return
} else if err == ErrUnauthorized {
    c.Error(http.StatusUnauthorized, "Not authorized to access this resource")
    return
} else if err != nil {
    c.Error(http.StatusInternalServerError, "An unexpected error occurred")
    return
}
```

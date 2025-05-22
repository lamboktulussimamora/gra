# Router Package

The router package is a core component of the GRA framework that provides HTTP routing capabilities.

## Features

- HTTP method-based routing (GET, POST, PUT, DELETE, etc.)
- Path parameters with colon syntax (e.g., `/users/:id`)
- Route grouping for organizing routes
- Middleware support with chaining capabilities
- Custom 404 and 405 handlers

## Usage

### Basic Router Setup

```go
import (
    "github.com/lamboktulussimamora/gra/router"
    "github.com/lamboktulussimamora/gra/context"
    "net/http"
)

func main() {
    // Create a new router
    r := router.New()
    
    // Register a simple route
    r.GET("/hello", func(c *context.Context) {
        c.Success(http.StatusOK, "Hello World!", nil)
    })
    
    // Start the server
    http.ListenAndServe(":8080", r)
}
```

### Path Parameters

```go
r.GET("/users/:id", func(c *context.Context) {
    id := c.GetParam("id")
    // Use the id parameter
    c.Success(http.StatusOK, "User found", map[string]any{"id": id})
})
```

### Route Groups

```go
// Create an API group
api := r.Group("/api")

// Add routes to the group
api.GET("/users", ListUsersHandler)
api.POST("/users", CreateUserHandler)

// Create a nested group
v1 := api.Group("/v1")
v1.GET("/products", ListProductsV1Handler)
```

### Middleware

```go
// Global middleware applied to all routes
r.Use(Logger(), Recovery())

// Group-specific middleware
admin := r.Group("/admin")
admin.Use(AuthMiddleware(), RoleCheck("admin"))
admin.GET("/dashboard", AdminDashboardHandler)
```

## Best Practices

1. Organize related routes into groups
2. Use middleware for cross-cutting concerns
3. Keep handler functions small and focused
4. Follow RESTful conventions for API routes

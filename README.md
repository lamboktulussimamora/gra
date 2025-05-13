# GRA Framework

[![Test and Coverage](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml/badge.svg)](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/lamboktulussimamora/gra/badge.svg?branch=main)](https://coveralls.io/github/lamboktulussimamora/gra?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/lamboktulussimamora/gra)](https://goreportcard.com/report/github.com/lamboktulussimamora/gra)

A lightweight HTTP framework for building web applications in Go, inspired by Gin.

## Features

- Context-based request handling
- HTTP routing with path parameters
- Middleware support
- Request validation
- Standardized API responses
- Structured logging
- Clean architecture friendly

## Installation

```bash
go get github.com/lamboktulussimamora/gra
```

## Quick Start

```go
package main

import (
	"net/http"
	"github.com/lamboktulussimamora/gra"
)

func main() {
	// Create a new router
	r := gra.New()

	// Define a route
	r.GET("/hello", func(c *gra.Context) {
		c.Success(http.StatusOK, "Hello World", nil)
	})

	// Start the server
	gra.Run(":8080", r)
}
```

## Context

The `Context` provides a convenient way to handle HTTP requests and responses:

```go
// Get path parameters
id := c.GetParam("id")

// Get query parameters
name := c.GetQuery("name")

// Parse JSON request body
var user User
if err := c.BindJSON(&user); err != nil {
	c.Error(http.StatusBadRequest, "Invalid request")
	return
}

// Send JSON response
c.JSON(http.StatusOK, map[string]any{
	"message": "Success",
})

// Send standardized success response
c.Success(http.StatusOK, "User created", user)

// Send standardized error response
c.Error(http.StatusNotFound, "User not found")
```

## Router

The `Router` handles HTTP routing:

```go
r := core.New()

// Register routes
r.GET("/users", listUsers)
r.POST("/users", createUser)
r.GET("/users/:id", getUser)
r.PUT("/users/:id", updateUser)
r.DELETE("/users/:id", deleteUser)
```

## Middleware

Middleware functions can be used to add common functionality:

```go
// Use global middleware
r.Use(
	middleware.Logger(),
	middleware.Recovery(),
	middleware.CORS("*"),
)
```

## Validation

Validate request data using struct tags:

```go
type User struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"enum=user,admin,guest"`
	Phone    string `json:"phone" validate:"regexp=^[0-9]{10}$"`
}

func createUser(c *core.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request")
		return
	}

	v := validator.New()
	errors := v.Validate(user)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, map[string]any{
			"status": "error",
			"errors": errors,
		})
		return
	}

	// Process validated user...
}
```

### Validation Rules

The validator supports the following validation rules:

- `required`: Field cannot be empty/zero
- `email`: Field must be a valid email address
- `min=X`: String length or number value must be at least X
- `max=X`: String length or number value must be at most X
- `regexp=pattern`: String must match the specified regular expression pattern
- `enum=val1,val2,val3`: String must be one of the specified values
- `range=min,max`: Number must be within the specified inclusive range

### Custom Error Messages

You can specify custom error messages for validation rules:

```go
type Product struct {
    Name  string `json:"name" validate:"required|Name is mandatory"`
    Price int    `json:"price" validate:"range=1,10000|Price must be between $1 and $10,000"`
}
```

### Batch Validation

You can validate multiple objects at once:

```go
// Products to validate
products := []any{
    Product{Name: "Product 1", Price: 100},
    Product{Name: "", Price: -5},  // Invalid
    Product{Name: "Product 3", Price: 9999},
}

v := validator.New()
results := v.ValidateBatch(products)

// Check if any validation errors
if v.HasBatchErrors(results) {
    // Get only invalid results
    invalid := v.FilterInvalid(results)
    
    for _, result := range invalid {
        fmt.Printf("Item at index %d has %d errors\n", result.Index, len(result.Errors))
        for _, err := range result.Errors {
            fmt.Printf("  - Field: %s, Error: %s\n", err.Field, err.Message)
        }
    }
}
```

## Examples

See the `examples` directory for more usage examples.

## Testing

The framework has comprehensive test coverage:

- Overall test coverage: 90.2%
- Full test coverage for middleware and adapter packages
- High coverage for router (95.7%), context (90.0%), logger (87.9%), and validator (82.6%)

To run tests:

```bash
go test ./...
```

To check test coverage:

```bash
go test ./... -cover
```

## License

MIT

# GRA Framework Quick Start Guide

A minimalist HTTP framework for Go inspired by Gin, focusing on clean architecture principles.

## Installation

```bash
go get github.com/lamboktulussimamora/gra
```

## Creating Your First API

1. Create a new Go project:

```bash
mkdir myapi
cd myapi
go mod init myapi
```

2. Create a main.go file:

```go
package main

import (
	"net/http"
	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	// Create a new router
	r := gra.New()

	// Add middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS("*"))

	// Define routes
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to GRA Framework", Response{
			Message: "Hello, World!",
		})
	})

	r.GET("/hello/:name", func(c *gra.Context) {
		name := c.GetParam("name")
		c.Success(http.StatusOK, "Hello "+name, nil)
	})

	// Start the server
	gra.Run(":8080", r)
}
```

3. Run the application:

```bash
go mod tidy
go run main.go
```

4. Test your API:

```bash
# Get the welcome message
curl http://localhost:8080/

# Get a personalized greeting
curl http://localhost:8080/hello/John
```

## Implementing Clean Architecture

GRA Framework is designed to support clean architecture patterns. Here's a typical project structure:

```
myapi/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   └── user/
│   │       ├── repository.go   # Repository interface
│   │       └── user.go         # Domain entity
│   ├── interface/
│   │   ├── handler/
│   │   │   └── user_handler.go # HTTP handlers
│   │   └── repository/
│   │       └── user_repo.go    # Repository implementation
│   └── usecase/
│       └── user_usecase.go     # Business logic
├── go.mod
└── go.sum
```

## Key Features

- **Context-based handlers**: Simplified request/response handling
- **Middleware support**: Easy to add logging, auth, CORS, etc.
- **Clean validation**: Struct tag-based validation
- **Standardized responses**: Consistent API responses
- **Path parameters**: Simple URL parameter extraction
- **Query parameters**: Easy access to URL query parameters

## Need Help?

For more detailed examples and documentation, see the `examples/` directory in the repository.

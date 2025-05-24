# Getting Started with GRA Framework

This guide will help you get up and running with the GRA Framework quickly and effectively.

## Prerequisites

Before you begin, make sure you have:

- Go 1.18 or later installed
- A basic understanding of Go programming
- A code editor or IDE (like VS Code, GoLand, etc.)

## Installation

### Using Go Modules (Recommended)

The simplest way to install GRA is using Go modules:

```bash
go get github.com/lamboktulussimamora/gra@latest
```

You can also specify a particular version:

```bash
go get github.com/lamboktulussimamora/gra@v1.2.0
```

### Manual Installation

If you prefer, you can clone the repository directly:

```bash
git clone https://github.com/lamboktulussimamora/gra.git
cd gra
go install
```

## Create Your First Application

Create a file named `main.go` with the following content:

```go
package main

import (
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
)

func main() {
	// Create a new router
	r := gra.New()

	// Add middlewares
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS("*"),
	)

	// Set up routes
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to GRA Framework", map[string]any{
			"version": gra.Version,
			"time":    time.Now(),
		})
	})

	// Route with path parameter
	r.GET("/users/:id", func(c *gra.Context) {
		id := c.GetParam("id")
		c.Success(http.StatusOK, "User found", map[string]any{
			"id":   id,
			"name": "John Doe",
		})
	})

	// Start the server
	r.Run(":8080")
}
```

## Initialize Go Module

Before creating your application, initialize a Go module:

```bash
mkdir my-gra-app
cd my-gra-app
go mod init example.com/myapp
```

## Run Your Application

Run your application with:

```bash
go run main.go
```

Visit [http://localhost:8080](http://localhost:8080) in your browser to see your application in action.

## Project Structure

For small applications, a single `main.go` file might be sufficient. However, as your application grows, you'll want to organize your code better. Here's a recommended project structure for a GRA application:

```
myapp/
├── cmd/
│   └── api/
│       └── main.go         # Application entry point
├── internal/
│   ├── handlers/           # HTTP handlers
│   │   ├── users.go
│   │   └── products.go
│   ├── middleware/         # Custom middleware
│   │   └── auth.go
│   ├── models/             # Data models
│   │   ├── user.go
│   │   └── product.go
│   └── services/           # Business logic
│       ├── user_service.go
│       └── product_service.go
├── pkg/                    # Reusable packages
│   ├── config/
│   │   └── config.go
│   └── validator/
│       └── validator.go
├── api/                    # API documentation
│   └── swagger.yaml
├── go.mod
└── go.sum
```

### Explanation:

- **cmd/**: Contains the main applications
- **internal/**: Private application code
  - **handlers/**: HTTP request handlers
  - **middleware/**: Custom middleware
  - **models/**: Data models
  - **services/**: Business logic
- **pkg/**: Public libraries that can be used by external applications
- **api/**: API documentation

## Best Practices

When working with the GRA framework, consider these best practices:

1. **Group Related Routes**: Use route groups to organize related endpoints
2. **Apply Middleware Strategically**: Apply global middleware at the router level and specific middleware at the group or route level
3. **Use Standardized Response Formats**: Utilize the built-in Success and Error methods for consistent responses
4. **Validate Request Data**: Always validate incoming data with the validator package
5. **Handle Errors Gracefully**: Use appropriate status codes and error messages
6. **Include Logging**: Configure the Logger middleware for debugging and monitoring

Your server will be running at [http://localhost:8080](http://localhost:8080).

## Next Steps

- Check out the [core concepts](../core-concepts/) to learn more about the framework design
- See [examples](../examples/) for more complex usage patterns
- Learn about [middleware](../middleware/) to enhance your application's functionality

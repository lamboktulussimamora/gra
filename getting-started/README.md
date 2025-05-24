# Getting Started with GRA Framework

This guide will help you get up and running with the GRA Framework quickly.

## Installation

Install the GRA framework using Go modules:

```bash
go get github.com/lamboktulussimamora/gra
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

If you haven't already, initialize a Go module:

```bash
go mod init example.com/myapp
```

## Run Your Application

Run your application with:

```bash
go run main.go
```

Your server will be running at [http://localhost:8080](http://localhost:8080).

## Next Steps

- Check out the [core concepts](../core-concepts/) to learn more about the framework design
- See [examples](../examples/) for more complex usage patterns
- Learn about [middleware](../middleware/) to enhance your application's functionality

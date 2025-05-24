# GRA Framework Documentation

[![Test and Coverage](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml/badge.svg)](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/lamboktulussimamora/gra/badge.svg?branch=main)](https://coveralls.io/github/lamboktulussimamora/gra?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/lamboktulussimamora/gra)](https://goreportcard.com/report/github.com/lamboktulussimamora/gra)

Welcome to the GRA Framework documentation. GRA is a lightweight HTTP framework for building web applications in Go, inspired by Gin.

## Documentation Sections

- [Getting Started](getting-started/)
- [Core Concepts](core-concepts/)
- [API Reference](api-reference/)
- [Middleware](middleware/)
- [Examples](examples/)

## Features

- Context-based request handling
- HTTP routing with path parameters
- JWT authentication and authorization
- Secure HTTP headers middleware
- API versioning support
- Response caching
- Middleware support
- Request validation
- Standardized API responses
- Structured logging
- Clean architecture friendly

## Quick Start

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

    // Start the server
    r.Run(":8080")
}
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for details on how to contribute to GRA Framework.

## License

GRA Framework is open-source software licensed under the MIT license.

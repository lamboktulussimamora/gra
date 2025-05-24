# GRA Framework Examples

This section provides practical examples of using the GRA framework for different use cases.

## Table of Contents

- [Basic HTTP Server](#basic-http-server)
- [REST API with CRUD Operations](#rest-api-with-crud-operations)
- [Authentication and Security](#authentication-and-security)
- [API Versioning and Caching](#api-versioning-and-caching)
- [Middleware Usage](#middleware-usage)
- [Validation and Error Handling](#validation-and-error-handling)

## Basic HTTP Server

This example demonstrates how to create a simple HTTP server with GRA:

```go
package main

import (
	"net/http"
	"github.com/lamboktulussimamora/gra"
)

func main() {
	// Create a new router
	r := gra.New()

	// Define routes
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to GRA Framework", nil)
	})

	r.GET("/hello/:name", func(c *gra.Context) {
		name := c.GetParam("name")
		c.Success(http.StatusOK, "Hello, "+name+"!", nil)
	})

	// Start the server
	gra.Run(":8080", r)
}
```

To run this example:

```bash
go run main.go
```

Then visit `http://localhost:8080/` or `http://localhost:8080/hello/world` in your browser.

## REST API with CRUD Operations

This example demonstrates how to build a REST API with CRUD operations:

```go
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/lamboktulussimamora/gra"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	r := gra.New()
	
	// In-memory store for demo purposes
	var (
		users     = make(map[int]User)
		userMutex = &sync.Mutex{}
		nextID    = 1
	)

	// Add some sample data
	users[1] = User{ID: 1, Name: "Alice", Age: 28}
	users[2] = User{ID: 2, Name: "Bob", Age: 32}
	nextID = 3

	// GET /users - List all users
	r.GET("/users", func(c *gra.Context) {
		userMutex.Lock()
		userList := make([]User, 0, len(users))
		for _, user := range users {
			userList = append(userList, user)
		}
		userMutex.Unlock()

		c.Success(http.StatusOK, "Users retrieved", userList)
	})

	// GET /users/:id - Get a specific user
	r.GET("/users/:id", func(c *gra.Context) {
		id, err := strconv.Atoi(c.GetParam("id"))
		if err != nil {
			c.Error(http.StatusBadRequest, "Invalid user ID")
			return
		}

		userMutex.Lock()
		user, exists := users[id]
		userMutex.Unlock()

		if !exists {
			c.Error(http.StatusNotFound, "User not found")
			return
		}

		c.Success(http.StatusOK, "User retrieved", user)
	})

	// POST /users - Create a new user
	r.POST("/users", func(c *gra.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.Error(http.StatusBadRequest, "Invalid user data")
			return
		}

		userMutex.Lock()
		user.ID = nextID
		users[nextID] = user
		nextID++
		userMutex.Unlock()

		c.Success(http.StatusCreated, "User created", user)
	})

	// PUT /users/:id - Update a user
	r.PUT("/users/:id", func(c *gra.Context) {
		id, err := strconv.Atoi(c.GetParam("id"))
		if err != nil {
			c.Error(http.StatusBadRequest, "Invalid user ID")
			return
		}

		userMutex.Lock()
		_, exists := users[id]
		if !exists {
			userMutex.Unlock()
			c.Error(http.StatusNotFound, "User not found")
			return
		}

		var user User
		if err := c.BindJSON(&user); err != nil {
			userMutex.Unlock()
			c.Error(http.StatusBadRequest, "Invalid user data")
			return
		}

		user.ID = id
		users[id] = user
		userMutex.Unlock()

		c.Success(http.StatusOK, "User updated", user)
	})

	// DELETE /users/:id - Delete a user
	r.DELETE("/users/:id", func(c *gra.Context) {
		id, err := strconv.Atoi(c.GetParam("id"))
		if err != nil {
			c.Error(http.StatusBadRequest, "Invalid user ID")
			return
		}

		userMutex.Lock()
		_, exists := users[id]
		if !exists {
			userMutex.Unlock()
			c.Error(http.StatusNotFound, "User not found")
			return
		}

		delete(users, id)
		userMutex.Unlock()

		c.Success(http.StatusOK, "User deleted", nil)
	})

	fmt.Println("Starting REST API server on :8080")
	gra.Run(":8080", r)
}
```

## Authentication and Security

This example demonstrates how to implement JWT authentication and security headers:

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/jwt"
	"github.com/lamboktulussimamora/gra/middleware"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	r := gra.New()

	// Create JWT service
	jwtSecret := []byte("your-secure-secret-key")
	jwtService, err := jwt.NewServiceWithKey(jwtSecret)
	if err != nil {
		panic("Failed to create JWT service: " + err.Error())
	}

	// Apply global middleware
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS("*"),
		middleware.SecureHeaders(),
	)

	// Mock user store
	users := map[string]User{
		"admin": {Username: "admin", Password: "admin123", Role: "admin"},
		"user":  {Username: "user", Password: "user123", Role: "user"},
	}

	// Public routes
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to secure API", nil)
	})

	r.POST("/login", func(c *gra.Context) {
		var req LoginRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(http.StatusBadRequest, "Invalid login request")
			return
		}

		// Check credentials
		user, exists := users[req.Username]
		if !exists || user.Password != req.Password {
			c.Error(http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Generate JWT token
		claims := map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		}

		token, err := jwtService.GenerateToken(claims)
		if err != nil {
			c.Error(http.StatusInternalServerError, "Failed to generate token")
			return
		}

		c.Success(http.StatusOK, "Login successful", map[string]string{
			"token": token,
		})
	})

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.Auth(jwtService, ""))

	protected.GET("/profile", func(c *gra.Context) {
		// Get user from context (set by Auth middleware)
		userData, _ := c.Get("user")
		c.Success(http.StatusOK, "Profile retrieved", userData)
	})

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(jwtService, "admin"))

	admin.GET("/dashboard", func(c *gra.Context) {
		c.Success(http.StatusOK, "Admin dashboard", map[string]interface{}{
			"stats": map[string]int{
				"users":    100,
				"requests": 5000,
			},
		})
	})

	fmt.Println("Starting secure server on :8080")
	gra.Run(":8080", r)
}
```

## API Versioning and Caching

This example demonstrates API versioning and response caching:

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
	"github.com/lamboktulussimamora/gra/versioning"
)

func main() {
	r := gra.New()

	// Apply global middleware
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
	)

	// Create API group
	api := r.Group("/api")

	// Version 1 routes (default)
	v1 := versioning.New(api, versioning.Config{
		DefaultVersion: "1.0",
		HeaderName:     "API-Version",
	})

	// Apply cache middleware to v1
	v1.Use(middleware.Cache(5 * time.Minute))

	v1.GET("/products", func(c *gra.Context) {
		// This response will be cached for 5 minutes
		c.Success(http.StatusOK, "Products from API v1", []map[string]interface{}{
			{"id": 1, "name": "Product v1.1", "price": 19.99},
			{"id": 2, "name": "Product v1.2", "price": 29.99},
		})
	})

	// Version 2 routes
	v2 := versioning.New(api, versioning.Config{
		Version:    "2.0",
		HeaderName: "API-Version",
	})

	v2.GET("/products", func(c *gra.Context) {
		// Version 2 includes additional details
		c.Success(http.StatusOK, "Products from API v2", []map[string]interface{}{
			{"id": 1, "name": "Product v2.1", "price": 19.99, "stock": 150, "category": "Electronics"},
			{"id": 2, "name": "Product v2.2", "price": 29.99, "stock": 75, "category": "Books"},
		})
	})

	fmt.Println("Starting versioned API server on :8080")
	fmt.Println("Try: curl -H 'API-Version: 1.0' http://localhost:8080/api/products")
	fmt.Println("Try: curl -H 'API-Version: 2.0' http://localhost:8080/api/products")
	gra.Run(":8080", r)
}
```

## Middleware Usage

This example demonstrates how to use and create custom middleware:

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
)

// Custom middleware: request timing
func RequestTimer() gra.HandlerFunc {
	return func(c *gra.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Add response header
		c.Writer.Header().Set("X-Response-Time", duration.String())

		// Log duration
		fmt.Printf("[%s] %s - %v\n", c.Request.Method, c.Request.URL.Path, duration)
	}
}

// Custom middleware: request counter
func RequestCounter() gra.HandlerFunc {
	var count int64

	return func(c *gra.Context) {
		// Increment counter
		count++

		// Set in context
		c.Set("request_number", count)

		// Process request
		c.Next()
	}
}

func main() {
	r := gra.New()

	// Apply global middleware
	r.Use(
		middleware.Recovery(),
		middleware.Logger(),
		RequestTimer(),
		RequestCounter(),
	)

	// Define routes
	r.GET("/", func(c *gra.Context) {
		requestNum, _ := c.Get("request_number")
		c.Success(http.StatusOK, fmt.Sprintf("Request #%d", requestNum), nil)
	})

	// Group with additional middleware
	api := r.Group("/api")
	api.Use(middleware.CORS("*"))

	api.GET("/status", func(c *gra.Context) {
		// Simulate processing delay
		time.Sleep(100 * time.Millisecond)
		
		requestNum, _ := c.Get("request_number")
		c.Success(http.StatusOK, "API is running", map[string]interface{}{
			"status":         "healthy",
			"request_number": requestNum,
		})
	})

	fmt.Println("Starting server with middleware on :8080")
	gra.Run(":8080", r)
}
```

## Validation and Error Handling

This example demonstrates request validation and error handling:

```go
package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
	"github.com/lamboktulussimamora/gra/validator"
)

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=50"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"min=0"`
	Category    string  `json:"category" validate:"required,oneof=electronics clothing books food other"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=3,max=50"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	Stock       int     `json:"stock" validate:"omitempty,min=0"`
	Category    string  `json:"category" validate:"omitempty,oneof=electronics clothing books food other"`
}

func main() {
	r := gra.New()

	// Apply global middleware
	r.Use(
		middleware.Recovery(),
		middleware.Logger(),
	)

	// Custom validator
	v := validator.New()
	
	// Products API
	products := r.Group("/products")

	// Create product
	products.POST("/", func(c *gra.Context) {
		var req CreateProductRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(http.StatusBadRequest, "Invalid request body", "Request must be valid JSON")
			return
		}

		// Validate request
		if err := v.Validate(req); err != nil {
			// Convert validation errors to readable format
			validationErrors := v.FormatErrors(err)
			c.Error(http.StatusBadRequest, "Validation error", validationErrors...)
			return
		}

		// Process valid request
		// (In a real app, you would save to a database)
		c.Success(http.StatusCreated, "Product created", map[string]interface{}{
			"id":          123,
			"name":        req.Name,
			"description": req.Description,
			"price":       req.Price,
			"stock":       req.Stock,
			"category":    req.Category,
		})
	})

	// Update product
	products.PUT("/:id", func(c *gra.Context) {
		id := c.GetParam("id")
		
		// Validate ID
		if _, err := strconv.Atoi(id); err != nil {
			c.Error(http.StatusBadRequest, "Invalid product ID", "Product ID must be a number")
			return
		}

		var req UpdateProductRequest
		if err := c.BindJSON(&req); err != nil {
			c.Error(http.StatusBadRequest, "Invalid request body", "Request must be valid JSON")
			return
		}

		// Validate request
		if err := v.Validate(req); err != nil {
			validationErrors := v.FormatErrors(err)
			c.Error(http.StatusBadRequest, "Validation error", validationErrors...)
			return
		}

		// Process valid request
		// (In a real app, you would update in a database)
		c.Success(http.StatusOK, fmt.Sprintf("Product %s updated", id), map[string]interface{}{
			"id": id,
			// Include only fields that were provided
			"updated_fields": req,
		})
	})

	fmt.Println("Starting validation example server on :8080")
	gra.Run(":8080", r)
}
```

Each of these examples demonstrates different aspects of the GRA framework. You can find the complete source code in the `/examples` directory.

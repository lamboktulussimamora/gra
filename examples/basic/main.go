// Basic example using the gra framework
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
	"github.com/lamboktulussimamora/gra/router"
	"github.com/lamboktulussimamora/gra/validator"
)

// User represents a user model
// with validation tags
// to ensure required fields are present
// and email format is valid
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// setupRoutes configures all routes and middleware for the application
func setupRoutes() *router.Router {
	// Create a new router
	r := gra.New()

	// Add middlewares
	r.Use(
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS("*"),
	)

	// Home route - provides basic API information
	r.GET("/", func(c *gra.Context) {
		c.Success(http.StatusOK, "Welcome to GRA Framework", map[string]any{
			"version": gra.Version,
			"time":    time.Now(),
			"endpoints": []map[string]string{
				{"method": "GET", "path": "/", "description": "API information"},
				{"method": "GET", "path": "/users/:id", "description": "Get user by ID"},
				{"method": "POST", "path": "/users", "description": "Create new user"},
			},
		})
	})

	// Get user by ID route
	r.GET("/users/:id", func(c *gra.Context) {
		id := c.GetParam("id")

		// In a real application, you would fetch from database
		// Here we just return a mock user
		c.Success(http.StatusOK, "User found", map[string]any{
			"id":    id,
			"name":  "John Doe",
			"email": "john.doe@example.com",
		})
	})

	// Create user route
	r.POST("/users", createUserHandler)

	return r
}

// createUserHandler handles user creation with validation
func createUserHandler(c *gra.Context) {
	var user User

	// Bind JSON request body to user struct
	if err := c.BindJSON(&user); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate user input
	v := validator.New()
	errors := v.Validate(user)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, map[string]any{
			"status": "error",
			"error":  "Validation failed",
			"errors": errors,
		})
		return
	}

	// Mock creating a user (in real app, save to database)
	user.ID = 1
	user.Password = "********" // Hide password in response

	c.Success(http.StatusCreated, "User created successfully", user)
}

func main() {
	// Setup routes
	router := setupRoutes()

	// Start the server
	fmt.Println("🚀 Starting GRA Framework Basic Example")
	fmt.Println("📍 Server running at http://localhost:8080")
	fmt.Println("📖 Available endpoints:")
	fmt.Println("   GET  / - API information")
	fmt.Println("   GET  /users/:id - Get user by ID")
	fmt.Println("   POST /users - Create new user")
	fmt.Println()

	if err := gra.Run(":8080", router); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
	}
}

// Basic example using the gra framework
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/middleware"
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

	r.GET("/users/:id", func(c *gra.Context) {
		id := c.GetParam("id")
		c.Success(http.StatusOK, "User found", map[string]any{
			"id":   id,
			"name": "John Doe",
		})
	})

	r.POST("/users", func(c *gra.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.Error(http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate user
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

		// Mock creating a user
		user.ID = 1
		user.Password = "********" // Hide password

		c.Success(http.StatusCreated, "User created", user)
	})

	// Start the server
	fmt.Println("Server running at http://localhost:8080")
	gra.Run(":8080", r)
}

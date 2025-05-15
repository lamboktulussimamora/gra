// Example application demonstrating JWT authentication and secure headers
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/lamboktulussimamora/gra"
	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/jwt"
	"github.com/lamboktulussimamora/gra/middleware"
	"github.com/lamboktulussimamora/gra/router"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// LoginRequest represents the login form data
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response with token
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Server represents the API server
type Server struct {
	app        *router.Router
	jwtService *jwt.Service
	users      map[string]User // In-memory users for demo
}

func main() {
	// Create a new server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.app))
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Create JWT service with a signing key
	jwtService, err := jwt.NewServiceWithKey([]byte("my-secret-key"))
	if err != nil {
		return nil, err
	}

	// Create server
	server := &Server{
		app:        gra.New(), // gra.New() returns a *router.Router
		jwtService: jwtService,
		users:      make(map[string]User),
	}

	// Add demo users
	server.users["user1"] = User{
		ID:       "1",
		Username: "user1",
		Password: "password1",
		Role:     "user",
	}
	server.users["admin"] = User{
		ID:       "2",
		Username: "admin",
		Password: "admin123",
		Role:     "admin",
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Create a router wrapper to handle different middleware chains
	publicRouter := gra.New()
	apiRouter := gra.New()

	// Apply common middleware to all routers
	publicRouter.Use(middleware.Logger())
	publicRouter.Use(middleware.Recovery())
	publicRouter.Use(middleware.SecureHeaders())

	apiRouter.Use(middleware.Logger())
	apiRouter.Use(middleware.Recovery())
	apiRouter.Use(middleware.SecureHeaders())

	// Apply authentication middleware only to API router
	apiRouter.Use(middleware.Auth(s, "user"))

	// Public routes - no authentication required
	publicRouter.GET("/", s.handleHome)
	publicRouter.POST("/login", s.handleLogin)

	// User routes
	apiRouter.GET("/profile", s.handleGetProfile)

	// Admin routes with additional middleware
	adminRouter := gra.New()
	adminRouter.Use(middleware.Logger())
	adminRouter.Use(middleware.Recovery())
	adminRouter.Use(middleware.SecureHeaders())
	adminRouter.Use(middleware.Auth(s, "user"))
	adminRouter.Use(s.adminOnly)
	adminRouter.GET("/dashboard", s.handleAdminDashboard)

	// Direct routes for public endpoints
	s.app.GET("/", s.handleHome)
	s.app.POST("/login", s.handleLogin)
	
	// API routes with authentication
	s.app.GET("/api/profile", s.withMiddlewares(
		s.handleGetProfile,
		middleware.Logger(),
		middleware.Recovery(),
		middleware.SecureHeaders(),
		middleware.Auth(s, "user"),
	))
	
	// Admin routes with additional middleware
	s.app.GET("/api/admin/dashboard", s.withMiddlewares(
		s.handleAdminDashboard,
		middleware.Logger(),
		middleware.Recovery(),
		middleware.SecureHeaders(),
		middleware.Auth(s, "user"),
		s.adminOnly,
	))
}

// withMiddlewares applies middleware to a handler
func (s *Server) withMiddlewares(handler router.HandlerFunc, middlewares ...router.Middleware) router.HandlerFunc {
	// Apply middlewares in reverse order (last middleware runs first)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ValidateToken implements the JWTAuthenticator interface
func (s *Server) ValidateToken(tokenString string) (any, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// adminOnly middleware ensures that the user has admin role
func (s *Server) adminOnly(next router.HandlerFunc) router.HandlerFunc {
	return func(c *context.Context) {
		// Get user claims from context
		userClaims, ok := c.Value("user").(map[string]interface{})
		if !ok {
			c.Error(http.StatusUnauthorized, "Invalid user claims")
			return
		}

		// Check if user has admin role
		role, ok := userClaims["role"].(string)
		if !ok || role != "admin" {
			c.Error(http.StatusForbidden, "Admin access required")
			return
		}

		// Call the next handler
		next(c)
	}
}

// handleHome handles the home page
func (s *Server) handleHome(c *context.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to the API",
	})
}

// handleLogin handles user login
func (s *Server) handleLogin(c *context.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request")
		return
	}

	// Find the user
	user, exists := s.users[req.Username]
	if !exists || user.Password != req.Password {
		c.Error(http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Create JWT token
	claims := jwt.StandardClaims{
		Subject: user.ID,
		Custom: map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	}

	token, err := s.jwtService.GenerateToken(claims)
	if err != nil {
		c.Error(http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return token and user data
	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// handleGetProfile handles getting user profile
func (s *Server) handleGetProfile(c *context.Context) {
	// Get user claims from context
	userClaims, ok := c.Value("user").(map[string]interface{})
	if !ok {
		c.Error(http.StatusUnauthorized, "Invalid user claims")
		return
	}

	// Get user ID from claims
	userID, ok := userClaims["sub"].(string)
	if !ok {
		c.Error(http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// Find the user by ID
	var user User
	for _, u := range s.users {
		if u.ID == userID {
			user = u
			break
		}
	}

	if user.ID == "" {
		c.Error(http.StatusNotFound, "User not found")
		return
	}

	// Return user data (without password)
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// handleAdminDashboard handles admin dashboard
func (s *Server) handleAdminDashboard(c *context.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Welcome to the admin dashboard",
		"users":   s.users,
		"time":    time.Now(),
	})
}

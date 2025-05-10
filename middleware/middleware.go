// Package middleware provides common HTTP middleware components.
package middleware

import (
	"net/http"
	"strings"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/logger"
	"github.com/lamboktulussimamora/gra/router"
)

// JWTAuthenticator defines an interface for JWT token validation
type JWTAuthenticator interface {
	ValidateToken(tokenString string) (interface{}, error)
}

// Auth authenticates requests using JWT
func Auth(jwtService JWTAuthenticator, claimsKey string) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Get the Authorization header
			authHeader := c.Request.Header.Get("Authorization")
			if authHeader == "" {
				c.Error(http.StatusUnauthorized, "Authorization header is required")
				return
			}

			// Check if the header has the correct format (Bearer <token>)
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.Error(http.StatusUnauthorized, "Authorization header format must be Bearer <token>")
				return
			}

			// Extract the token
			tokenString := parts[1]

			// Validate the token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				c.Error(http.StatusUnauthorized, "Invalid token")
				return
			}

			// Add claims to context
			c.WithValue(claimsKey, claims)

			// Call the next handler
			next(c)
		}
	}
}

// Logger logs incoming requests
func Logger() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Log the request
			method := c.Request.Method
			path := c.Request.URL.Path

			// Log before handling
			log := logger.Get()
			log.Infof("Request: %s %s", method, path)

			// Call the next handler
			next(c)

			// Log after handling
			log.Infof("Completed: %s %s", method, path)
		}
	}
}

// Recovery recovers from panics
func Recovery() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			defer func() {
				if err := recover(); err != nil {
					log := logger.Get()
					log.Errorf("Panic recovered: %v", err)
					c.Error(http.StatusInternalServerError, "Internal server error")
				}
			}()

			next(c)
		}
	}
}

// CORS handles Cross-Origin Resource Sharing
func CORS(allowOrigin string) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				c.Writer.WriteHeader(http.StatusOK)
				return
			}

			next(c)
		}
	}
}

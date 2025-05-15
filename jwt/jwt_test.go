package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewService(t *testing.T) {
	t.Run("should return error when signing key is missing", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = nil // Empty signing key

		_, err := NewService(config)
		if err != ErrMissingKey {
			t.Errorf("Expected ErrMissingKey, got %v", err)
		}
	})

	t.Run("should create service with valid config", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte("test-secret-key")

		service, err := NewService(config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if service == nil {
			t.Error("Expected service to be created, got nil")
		}
	})
}

func TestNewServiceWithKey(t *testing.T) {
	t.Run("should create service with key", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte("test-secret-key"))
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if service == nil {
			t.Error("Expected service to be created, got nil")
		}
	})
}

func TestGenerateToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte("test-secret-key"))

	t.Run("should return error when subject is missing", func(t *testing.T) {
		claims := StandardClaims{
			// Missing subject
		}

		_, err := service.GenerateToken(claims)
		if err != ErrMissingSubject {
			t.Errorf("Expected ErrMissingSubject, got %v", err)
		}
	})

	t.Run("should generate valid token", func(t *testing.T) {
		claims := StandardClaims{
			Subject: "user-123",
			Custom: map[string]interface{}{
				"name":  "John Doe",
				"admin": true,
			},
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Error("Expected token to be generated, got empty string")
		}
	})

	t.Run("should include optional claims", func(t *testing.T) {
		claims := StandardClaims{
			ID:       "token-id-123",
			Subject:  "user-456",
			Audience: []string{"api", "web"},
			Custom: map[string]interface{}{
				"role": "admin",
			},
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Parse and validate token
		parsedToken, _ := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key"), nil
		})

		parsedClaims := parsedToken.Claims.(jwt.MapClaims)

		if parsedClaims["sub"] != "user-456" {
			t.Errorf("Expected sub to be 'user-456', got %v", parsedClaims["sub"])
		}
		if parsedClaims["jti"] != "token-id-123" {
			t.Errorf("Expected jti to be 'token-id-123', got %v", parsedClaims["jti"])
		}
		if parsedClaims["role"] != "admin" {
			t.Errorf("Expected role to be 'admin', got %v", parsedClaims["role"])
		}
	})
}

func TestValidateToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte("test-secret-key"))

	t.Run("should validate valid token", func(t *testing.T) {
		// Generate a token first
		claims := StandardClaims{
			Subject: "user-123",
			Custom: map[string]interface{}{
				"name": "John Doe",
				"role": "admin",
			},
		}

		token, _ := service.GenerateToken(claims)

		// Validate the token
		parsedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if parsedClaims["sub"] != "user-123" {
			t.Errorf("Expected subject to be 'user-123', got %v", parsedClaims["sub"])
		}
		if parsedClaims["name"] != "John Doe" {
			t.Errorf("Expected name to be 'John Doe', got %v", parsedClaims["name"])
		}
		if parsedClaims["role"] != "admin" {
			t.Errorf("Expected role to be 'admin', got %v", parsedClaims["role"])
		}
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		invalidToken := "invalid.token.string"
		_, err := service.ValidateToken(invalidToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("should reject token signed with different method", func(t *testing.T) {
		// Create a token with a different signing method
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "user-123",
		})

		tokenString, _ := token.SignedString([]byte("different-key"))

		_, err := service.ValidateToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("should reject expired token", func(t *testing.T) {
		// Create a token with expired claims
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		})

		tokenString, _ := expiredToken.SignedString([]byte("test-secret-key"))

		_, err := service.ValidateToken(tokenString)
		if err != ErrExpiredToken {
			t.Errorf("Expected ErrExpiredToken, got %v", err)
		}
	})
}

func TestRefreshToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte("test-secret-key"))

	t.Run("should refresh valid token", func(t *testing.T) {
		// Generate a token first
		claims := StandardClaims{
			Subject: "user-123",
			Custom: map[string]interface{}{
				"role": "admin",
			},
		}

		token, _ := service.GenerateToken(claims)

		// Refresh the token
		newToken, err := service.RefreshToken(token)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if newToken == token {
			t.Error("Expected new token to be different from original token")
		}

		// Validate the new token
		newClaims, _ := service.ValidateToken(newToken)
		if newClaims["sub"] != "user-123" {
			t.Errorf("Expected subject to be 'user-123', got %v", newClaims["sub"])
		}
		if newClaims["role"] != "admin" {
			t.Errorf("Expected role to be 'admin', got %v", newClaims["role"])
		}
	})

	t.Run("should reject invalid token for refresh", func(t *testing.T) {
		invalidToken := "invalid.token.string"
		_, err := service.RefreshToken(invalidToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("should allow refresh for expired token", func(t *testing.T) {
		// We need to manually create an expired token that we can successfully parse
		// This is a bit tricky to test properly without modifying the code
		// For a real implementation, you'd want to make the jwt.Parse function mockable
		// For now, this is left as a placeholder for this test case
	})
}

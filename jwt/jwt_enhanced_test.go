package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTEnhancedCoverage provides additional tests to improve code coverage
func TestJWTEnhancedCoverage(t *testing.T) {
	t.Run("NewServiceWithNilSigningMethod", func(t *testing.T) {
		config := Config{
			SigningKey:      []byte(testSecretKey),
			SigningMethod:   nil, // Test nil signing method
			ExpirationTime:  time.Hour,
			RefreshDuration: time.Hour * 24,
			Issuer:          "test-issuer",
		}

		service, err := NewService(config)
		if err != nil {
			t.Fatalf("Expected service to be created with nil signing method, got error: %v", err)
		}

		if service == nil {
			t.Fatal("Expected service to be created, got nil")
		}

		// Verify that default signing method was set
		if service.config.SigningMethod != jwt.SigningMethodHS256 {
			t.Errorf("Expected default signing method to be HS256, got %v", service.config.SigningMethod)
		}

		// Test that the service works with the default signing method
		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with default signing method, got error: %v", err)
		}

		_, err = service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work with default signing method, got error: %v", err)
		}
	})

	t.Run("GenerateRandomTokenIDUniqueness", func(t *testing.T) {
		// Test that generateRandomTokenID produces unique values
		// This indirectly tests the function without mocking
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := generateRandomTokenID()
			if id == "" {
				t.Error("Expected non-empty token ID")
			}
			if ids[id] {
				t.Errorf("Duplicate token ID generated: %s", id)
			}
			ids[id] = true

			// All IDs should be hex-encoded and consistent length
			if len(id) != 16 { // 8 bytes = 16 hex chars
				t.Errorf("Expected token ID to be 16 characters (hex-encoded 8 bytes), got %d: %s", len(id), id)
			}
		}
	})

	t.Run("ValidateTokenWithRegisteredClaims", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Create a token with RegisteredClaims instead of MapClaims
		// The current implementation should handle this correctly
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   testUserIDCommon,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})

		tokenString, err := token.SignedString([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		// This should work because RegisteredClaims is handled by the JWT library
		validatedClaims, err := service.ValidateToken(tokenString)
		if err != nil {
			t.Fatalf("Expected validation to work with RegisteredClaims: %v", err)
		}

		if validatedClaims["sub"] != testUserIDCommon {
			t.Errorf("Expected subject to be preserved: %v", validatedClaims["sub"])
		}
	})

	t.Run("RefreshTokenWithMissingSubject", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Create a token that will result in missing subject during refresh
		tokenWithoutSub := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp":    time.Now().Add(-time.Hour).Unix(), // Expired
			"iat":    time.Now().Unix(),
			"custom": "value",
		})

		tokenString, err := tokenWithoutSub.SignedString([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		_, err = service.RefreshToken(tokenString)
		if err != ErrMissingSubject {
			t.Errorf("Expected ErrMissingSubject for token with no subject, got: %v", err)
		}
	})

	t.Run("RefreshTokenPreservesCustomClaims", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Create a token with various custom claims
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom: map[string]interface{}{
				"role":        testRoleAdmin,
				"permissions": []string{"read", "write"},
				"level":       5,
				"active":      true,
			},
		}

		originalToken, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		refreshedToken, err := service.RefreshToken(originalToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		refreshedClaims, err := service.ValidateToken(refreshedToken)
		if err != nil {
			t.Fatalf("Failed to validate refreshed token: %v", err)
		}

		// Verify custom claims are preserved
		if refreshedClaims["role"] != testRoleAdmin {
			t.Errorf("Expected role to be preserved: %v", refreshedClaims["role"])
		}

		permissions, ok := refreshedClaims["permissions"].([]interface{})
		if !ok || len(permissions) != 2 {
			t.Errorf("Expected permissions to be preserved: %v", refreshedClaims["permissions"])
		}

		if refreshedClaims["level"].(float64) != 5 {
			t.Errorf("Expected level to be preserved: %v", refreshedClaims["level"])
		}

		if refreshedClaims["active"].(bool) != true {
			t.Errorf("Expected active to be preserved: %v", refreshedClaims["active"])
		}

		// Verify subject is preserved
		if refreshedClaims["sub"] != testUserIDCommon {
			t.Errorf("Expected subject to be preserved: %v", refreshedClaims["sub"])
		}
	})

	t.Run("TokenValidationWithSpecificErrors", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Test token with wrong signing method in header
		wrongMethodToken := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
			"sub": testUserIDCommon,
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		tokenString, err := wrongMethodToken.SignedString([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		_, err = service.ValidateToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for wrong signing method, got: %v", err)
		}

		// Test completely malformed token
		_, err = service.ValidateToken("not.a.jwt")
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for malformed token, got: %v", err)
		}

		// Test token with invalid JSON payload
		invalidJSONToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalidjson.signature"
		_, err = service.ValidateToken(invalidJSONToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for invalid JSON, got: %v", err)
		}
	})

	t.Run("RefreshTokenWithWrongSigningMethod", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Create token with different signing method
		wrongMethodToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"sub": testUserIDCommon,
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired
		})

		tokenString, err := wrongMethodToken.SignedString([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		_, err = service.RefreshToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for wrong signing method in refresh, got: %v", err)
		}
	})

	t.Run("GenerateTokenWithEmptyCustomClaims", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Test with nil custom claims
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom:  nil,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with nil custom claims: %v", err)
		}

		validatedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work: %v", err)
		}

		if validatedClaims["sub"] != testUserIDCommon {
			t.Errorf("Expected subject to be preserved: %v", validatedClaims["sub"])
		}

		// Test with empty custom claims map
		claims.Custom = make(map[string]interface{})
		token, err = service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with empty custom claims: %v", err)
		}

		_, err = service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work: %v", err)
		}
	})
}

// TestJWTServiceBoundaryConditions tests boundary conditions and edge cases
func TestJWTServiceBoundaryConditions(t *testing.T) {
	t.Run("VeryShortSigningKey", func(t *testing.T) {
		// Test with very short signing key
		shortKey := []byte("x")
		service, err := NewServiceWithKey(shortKey)
		if err != nil {
			t.Fatalf("Expected service creation to work with short key: %v", err)
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with short key: %v", err)
		}

		_, err = service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work with short key: %v", err)
		}
	})

	t.Run("VeryLongSigningKey", func(t *testing.T) {
		// Test with very long signing key (1KB)
		longKey := make([]byte, 1024)
		for i := range longKey {
			longKey[i] = byte(i % 256)
		}

		service, err := NewServiceWithKey(longKey)
		if err != nil {
			t.Fatalf("Expected service creation to work with long key: %v", err)
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with long key: %v", err)
		}

		_, err = service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work with long key: %v", err)
		}
	})

	t.Run("TokenWithVeryLongSubject", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		// Create a very long subject (1KB)
		longSubject := string(make([]byte, 1024))
		for i := range longSubject {
			longSubject = longSubject[:i] + "x" + longSubject[i+1:]
		}

		claims := StandardClaims{
			Subject: longSubject,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with long subject: %v", err)
		}

		validatedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token validation to work with long subject: %v", err)
		}

		if validatedClaims["sub"] != longSubject {
			t.Error("Expected long subject to be preserved")
		}
	})

	t.Run("TokenWithZeroExpiration", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)
		config.ExpirationTime = 0 // Zero expiration

		service, err := NewService(config)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf("Expected token generation to work with zero expiration: %v", err)
		}

		// Token should be immediately expired
		_, err = service.ValidateToken(token)
		if err != ErrExpiredToken {
			t.Errorf("Expected token to be immediately expired with zero duration, got: %v", err)
		}
	})
}

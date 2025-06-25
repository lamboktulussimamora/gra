package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTAdvancedScenarios tests more complex JWT scenarios
func TestJWTAdvancedScenarios(t *testing.T) {
	t.Run("RefreshTokenWithExpiredToken", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)
		config.ExpirationTime = time.Millisecond * 50 // Very short expiration for original token

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Generate a token
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom: map[string]interface{}{
				"role": testRoleAdmin,
			},
		}

		originalToken, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Wait for token to expire
		time.Sleep(time.Millisecond * 60)

		// Verify token is expired
		_, err = service.ValidateToken(originalToken)
		if err != ErrExpiredToken {
			t.Errorf(errMsgExpiredToken, err)
		}

		// Now update config to have longer expiration for refresh
		config.ExpirationTime = time.Hour
		service.config = config

		// Should still be able to refresh expired token
		newToken, err := service.RefreshToken(originalToken)
		if err != nil {
			t.Fatalf("Expected refresh to work with expired token, got %v", err)
		}

		if newToken == "" {
			t.Error(errTokenEmpty)
		}

		if newToken == originalToken {
			t.Error(errTokenNotDifferent)
		}

		// New token should be valid
		newClaims, err := service.ValidateToken(newToken)
		if err != nil {
			t.Fatalf("Expected new token to be valid, got %v", err)
		}

		// Should preserve original subject
		if newClaims["sub"] != testUserIDCommon {
			t.Errorf(errMsgExpectedSub, testUserIDCommon, newClaims["sub"])
		}

		// Should preserve custom claims
		if newClaims["role"] != testRoleAdmin {
			t.Errorf(errMsgExpectedRole, testRoleAdmin, newClaims["role"])
		}
	})

	t.Run("RefreshTokenWithInvalidToken", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Try to refresh invalid token
		_, err = service.RefreshToken(testInvalidToken)
		if err == nil {
			t.Error("Expected error when refreshing invalid token")
		}

		// Should not be able to refresh token signed with different key
		differentService, err := NewService(Config{
			SigningKey:      []byte(testDifferentKey),
			SigningMethod:   jwt.SigningMethodHS256,
			ExpirationTime:  time.Hour,
			RefreshDuration: time.Hour * 24,
			Issuer:          "different-issuer",
		})
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		validToken, err := service.GenerateToken(StandardClaims{
			Subject: testUserIDCommon,
		})
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		_, err = differentService.RefreshToken(validToken)
		if err == nil {
			t.Error("Expected error when refreshing token with different signing key")
		}
	})

	t.Run("TokenWithAudience", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		claims := StandardClaims{
			Subject:  testUserIDCommon,
			Audience: []string{testAPIAudience, testWebAudience},
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		validatedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Check audience claim
		aud, exists := validatedClaims["aud"]
		if !exists {
			t.Error("Expected audience claim to exist")
		}

		audSlice, ok := aud.([]interface{})
		if !ok {
			t.Error("Expected audience to be a slice")
		}

		if len(audSlice) != 2 {
			t.Errorf("Expected 2 audience values, got %d", len(audSlice))
		}

		if audSlice[0].(string) != testAPIAudience {
			t.Errorf("Expected first audience to be %s, got %s", testAPIAudience, audSlice[0])
		}

		if audSlice[1].(string) != testWebAudience {
			t.Errorf("Expected second audience to be %s, got %s", testWebAudience, audSlice[1])
		}
	})

	t.Run("TokenWithComplexCustomClaims", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		customClaims := map[string]interface{}{
			"role":        testRoleAdmin,
			"permissions": []string{"read", "write", "delete"},
			"metadata": map[string]interface{}{
				"department": "engineering",
				"level":      5,
			},
			"is_active": true,
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
			ID:      testTokenID,
			Custom:  customClaims,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		validatedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Verify custom claims
		if validatedClaims["role"] != testRoleAdmin {
			t.Errorf(errMsgExpectedRole, testRoleAdmin, validatedClaims["role"])
		}

		permissions, ok := validatedClaims["permissions"].([]interface{})
		if !ok {
			t.Error("Expected permissions to be a slice")
		}

		if len(permissions) != 3 {
			t.Errorf("Expected 3 permissions, got %d", len(permissions))
		}

		metadata, ok := validatedClaims["metadata"].(map[string]interface{})
		if !ok {
			t.Error("Expected metadata to be a map")
		}

		if metadata["department"].(string) != "engineering" {
			t.Errorf("Expected department to be 'engineering', got %s", metadata["department"])
		}

		if validatedClaims["is_active"].(bool) != true {
			t.Errorf("Expected is_active to be true, got %v", validatedClaims["is_active"])
		}
	})

	t.Run("GenerateRandomTokenID", func(t *testing.T) {
		// Generate multiple token IDs to ensure uniqueness
		ids := make(map[string]bool)
		for i := 0; i < 10; i++ {
			id := generateRandomTokenID()
			if id == "" {
				t.Error("Expected non-empty token ID")
			}
			if ids[id] {
				t.Errorf("Duplicate token ID generated: %s", id)
			}
			ids[id] = true
		}
	})

	t.Run("TokenValidationEdgeCases", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Test empty token
		_, err = service.ValidateToken("")
		if err == nil {
			t.Error("Expected error for empty token")
		}

		// Test malformed token (not enough parts)
		_, err = service.ValidateToken("invalid.token")
		if err == nil {
			t.Error("Expected error for malformed token")
		}

		// Test token with invalid signature
		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		validToken, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Modify the token to make signature invalid
		tamperedToken := validToken[:len(validToken)-5] + "XXXXX"
		_, err = service.ValidateToken(tamperedToken)
		if err == nil {
			t.Error("Expected error for token with invalid signature")
		}
	})
}

// TestJWTConfigurationVariations tests different JWT configurations
func TestJWTConfigurationVariations(t *testing.T) {
	t.Run("DifferentSigningMethods", func(t *testing.T) {
		signingMethods := []jwt.SigningMethod{
			jwt.SigningMethodHS256,
			jwt.SigningMethodHS384,
			jwt.SigningMethodHS512,
		}

		for _, method := range signingMethods {
			t.Run(method.Alg(), func(t *testing.T) {
				config := DefaultConfig()
				config.SigningKey = []byte(testSecretKey)
				config.SigningMethod = method

				service, err := NewService(config)
				if err != nil {
					t.Fatalf(errMsgNoError, err)
				}

				claims := StandardClaims{
					Subject: testUserIDCommon,
				}

				token, err := service.GenerateToken(claims)
				if err != nil {
					t.Fatalf("Failed to generate token with %s: %v", method.Alg(), err)
				}

				validatedClaims, err := service.ValidateToken(token)
				if err != nil {
					t.Fatalf("Failed to validate token with %s: %v", method.Alg(), err)
				}

				if validatedClaims["sub"] != testUserIDCommon {
					t.Errorf("Expected subject %s, got %v", testUserIDCommon, validatedClaims["sub"])
				}
			})
		}
	})

	t.Run("CustomExpiration", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)
		config.ExpirationTime = time.Second * 2

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		// Token should be valid immediately
		_, err = service.ValidateToken(token)
		if err != nil {
			t.Fatalf("Expected token to be valid immediately: %v", err)
		}

		// Wait for expiration
		time.Sleep(time.Second * 3)

		// Token should now be expired
		_, err = service.ValidateToken(token)
		if err != ErrExpiredToken {
			t.Errorf(errMsgExpiredToken, err)
		}
	})

	t.Run("CustomIssuer", func(t *testing.T) {
		customIssuer := "custom-test-issuer"
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)
		config.Issuer = customIssuer

		service, err := NewService(config)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		claims := StandardClaims{
			Subject: testUserIDCommon,
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		validatedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf(errMsgNoError, err)
		}

		if validatedClaims["iss"] != customIssuer {
			t.Errorf("Expected issuer %s, got %v", customIssuer, validatedClaims["iss"])
		}
	})
}

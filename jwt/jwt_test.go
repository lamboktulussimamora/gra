package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Test constants
const (
	testRoleAdmin    = "admin"
	testUserID       = "user-456"
	testTokenID      = "token-id-123"
	testUserIDCommon = "user-123"
	testSecretKey    = "test-secret-key"
	testName         = "John Doe"
	testDifferentKey = "different-key"
	testInvalidToken = "invalid.token.string"
	testAPIAudience  = "api"
	testWebAudience  = "web"

	// Error message templates
	errMsgNoError         = "Expected no error, got %v"
	errMsgInvalidToken    = "Expected ErrInvalidToken, got %v"
	errMsgExpectedRole    = "Expected role to be '%s', got %v"
	errMsgExpectedSub     = "Expected sub to be '%s', got %v"
	errMsgExpectedSubject = "Expected subject to be '%s', got %v"
	errMsgExpectedJTI     = "Expected jti to be '%s', got %v"
	errMsgExpectedName    = "Expected name to be '%s', got %v"
	// #nosec G101 -- This is a test error message, not a credential
	errMsgExpiredToken   = "Expected ErrExpiredToken, got %v"
	errMsgMissingKey     = "Expected ErrMissingKey, got %v"
	errMsgMissingSubject = "Expected ErrMissingSubject, got %v"
	errServiceNil        = "Expected service to be created, got nil"
	errTokenEmpty        = "Expected token to be generated, got empty string"
	errTokenNotDifferent = "Expected new token to be different from original token"
)

func TestNewService(t *testing.T) {
	t.Run("should return error when signing key is missing", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = nil // Empty signing key

		_, err := NewService(config)
		if err != ErrMissingKey {
			t.Errorf(errMsgMissingKey, err)
		}
	})

	t.Run("should create service with valid config", func(t *testing.T) {
		config := DefaultConfig()
		config.SigningKey = []byte(testSecretKey)

		service, err := NewService(config)
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}
		if service == nil {
			t.Error(errServiceNil)
		}
	})
}

func TestNewServiceWithKey(t *testing.T) {
	t.Run("should create service with key", func(t *testing.T) {
		service, err := NewServiceWithKey([]byte(testSecretKey))
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}
		if service == nil {
			t.Error(errServiceNil)
		}
	})
}

func TestGenerateToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte(testSecretKey))

	t.Run("should return error when subject is missing", func(t *testing.T) {
		claims := StandardClaims{
			// Missing subject
		}

		_, err := service.GenerateToken(claims)
		if err != ErrMissingSubject {
			t.Errorf(errMsgMissingSubject, err)
		}
	})

	t.Run("should generate valid token", func(t *testing.T) {
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom: map[string]interface{}{
				"name":  testName,
				"admin": true,
			},
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}
		if token == "" {
			t.Error(errTokenEmpty)
		}
	})

	t.Run("should include optional claims", func(t *testing.T) {
		claims := StandardClaims{
			ID:       testTokenID,
			Subject:  testUserID,
			Audience: []string{testAPIAudience, testWebAudience},
			Custom: map[string]interface{}{
				"role": testRoleAdmin,
			},
		}

		token, err := service.GenerateToken(claims)
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}

		// Parse and validate token
		parsedToken, _ := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
			return []byte(testSecretKey), nil
		})

		parsedClaims := parsedToken.Claims.(jwt.MapClaims)

		if parsedClaims["sub"] != testUserID {
			t.Errorf(errMsgExpectedSub, testUserID, parsedClaims["sub"])
		}
		if parsedClaims["jti"] != testTokenID {
			t.Errorf(errMsgExpectedJTI, testTokenID, parsedClaims["jti"])
		}
		if parsedClaims["role"] != testRoleAdmin {
			t.Errorf(errMsgExpectedRole, testRoleAdmin, parsedClaims["role"])
		}
	})
}

func TestValidateToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte(testSecretKey))

	t.Run("should validate valid token", func(t *testing.T) {
		// Generate a token first
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom: map[string]interface{}{
				"name": testName,
				"role": testRoleAdmin,
			},
		}

		token, _ := service.GenerateToken(claims)

		// Validate the token
		parsedClaims, err := service.ValidateToken(token)
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}

		if parsedClaims["sub"] != testUserIDCommon {
			t.Errorf(errMsgExpectedSubject, testUserIDCommon, parsedClaims["sub"])
		}
		if parsedClaims["name"] != testName {
			t.Errorf(errMsgExpectedName, testName, parsedClaims["name"])
		}
		if parsedClaims["role"] != testRoleAdmin {
			t.Errorf(errMsgExpectedRole, testRoleAdmin, parsedClaims["role"])
		}
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		_, err := service.ValidateToken(testInvalidToken)
		if err != ErrInvalidToken {
			t.Errorf(errMsgInvalidToken, err)
		}
	})

	t.Run("should reject token signed with different method", func(t *testing.T) {
		// Create a token with a different signing method
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": testUserIDCommon,
		})

		tokenString, _ := token.SignedString([]byte(testDifferentKey))

		_, err := service.ValidateToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf(errMsgInvalidToken, err)
		}
	})

	t.Run("should reject expired token", func(t *testing.T) {
		// Create a token with expired claims
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": testUserIDCommon,
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		})

		tokenString, _ := expiredToken.SignedString([]byte(testSecretKey))

		_, err := service.ValidateToken(tokenString)
		if err != ErrExpiredToken {
			t.Errorf(errMsgExpiredToken, err)
		}
	})
}

func TestRefreshToken(t *testing.T) {
	service, _ := NewServiceWithKey([]byte(testSecretKey))

	t.Run("should refresh valid token", func(t *testing.T) {
		// Generate a token first
		claims := StandardClaims{
			Subject: testUserIDCommon,
			Custom: map[string]interface{}{
				"role": testRoleAdmin,
			},
		}

		token, _ := service.GenerateToken(claims)

		// Refresh the token
		newToken, err := service.RefreshToken(token)
		if err != nil {
			t.Errorf(errMsgNoError, err)
		}
		if newToken == token {
			t.Error(errTokenNotDifferent)
		}

		// Validate the new token
		newClaims, _ := service.ValidateToken(newToken)
		if newClaims["sub"] != testUserIDCommon {
			t.Errorf(errMsgExpectedSubject, testUserIDCommon, newClaims["sub"])
		}
		if newClaims["role"] != testRoleAdmin {
			t.Errorf(errMsgExpectedRole, testRoleAdmin, newClaims["role"])
		}
	})

	t.Run("should reject invalid token for refresh", func(t *testing.T) {
		_, err := service.RefreshToken(testInvalidToken)
		if err != ErrInvalidToken {
			t.Errorf(errMsgInvalidToken, err)
		}
	})

	t.Run("should allow refresh for expired token", func(t *testing.T) {
		// We need to manually create an expired token that we can successfully parse
		// This is a bit tricky to test properly without modifying the code
		// For a real implementation, you'd want to make the jwt.Parse function mockable
		// For now, this is left as a placeholder for this test case
	})
}

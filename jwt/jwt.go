// Package jwt provides JWT authentication functionality for the GRA framework.
package jwt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Common error types
var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrMissingKey     = errors.New("signing key is required")
	ErrMissingSubject = errors.New("subject claim is required")
)

// Config holds JWT configuration parameters
type Config struct {
	SigningKey      []byte
	SigningMethod   jwt.SigningMethod
	ExpirationTime  time.Duration
	RefreshDuration time.Duration
	Issuer          string
}

// DefaultConfig returns the default JWT configuration
func DefaultConfig() Config {
	return Config{
		SigningMethod:   jwt.SigningMethodHS256,
		ExpirationTime:  time.Hour * 24,     // 24 hours
		RefreshDuration: time.Hour * 24 * 7, // 7 days
		Issuer:          "gra-framework",
	}
}

// Service provides JWT token generation and validation
type Service struct {
	config Config
}

// NewService creates a new JWT service with the provided config
func NewService(config Config) (*Service, error) {
	if len(config.SigningKey) == 0 {
		return nil, ErrMissingKey
	}

	// Use default signing method if not specified
	if config.SigningMethod == nil {
		config.SigningMethod = jwt.SigningMethodHS256
	}

	return &Service{
		config: config,
	}, nil
}

// NewServiceWithKey creates a new JWT service with a signing key
func NewServiceWithKey(signingKey []byte) (*Service, error) {
	config := DefaultConfig()
	config.SigningKey = signingKey
	return NewService(config)
}

// StandardClaims represents the standard JWT claims
type StandardClaims struct {
	ID        string
	Subject   string
	Audience  []string
	ExpiresAt int64
	IssuedAt  int64
	Issuer    string
	Custom    map[string]interface{}
}

// GenerateToken creates a new JWT token with the provided claims
func (s *Service) GenerateToken(claims StandardClaims) (string, error) {
	if claims.Subject == "" {
		return "", ErrMissingSubject
	}

	now := time.Now()
	expiresAt := now.Add(s.config.ExpirationTime)

	// Create JWT claims
	jwtClaims := jwt.MapClaims{
		"sub": claims.Subject,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
		"iss": s.config.Issuer,
	}

	if claims.ID != "" {
		jwtClaims["jti"] = claims.ID
	}

	if len(claims.Audience) > 0 {
		jwtClaims["aud"] = claims.Audience
	}

	// Add custom claims if any
	for k, v := range claims.Custom {
		jwtClaims[k] = v
	}

	// Create token
	token := jwt.NewWithClaims(s.config.SigningMethod, jwtClaims)

	// Sign and get the complete encoded token as a string
	return token.SignedString(s.config.SigningKey)
}

// ValidateToken validates the JWT token and returns the parsed claims
func (s *Service) ValidateToken(tokenString string) (map[string]interface{}, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != s.config.SigningMethod.Alg() {
			return nil, ErrInvalidToken
		}
		return s.config.SigningKey, nil
	})

	if err != nil {
		// Check if the error is due to token expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Validate token
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Convert to map[string]interface{}
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// RefreshToken generates a new token based on the claims in an existing token
func (s *Service) RefreshToken(tokenString string) (string, error) {
	// First validate the old token
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh for expired tokens, but not for invalid tokens
		if err != ErrExpiredToken {
			return "", err
		}
	}

	// Create a new StandardClaims object
	newClaims := StandardClaims{
		Subject: claims["sub"].(string),
		// Add some randomness to ensure new token is different
		ID:     generateRandomTokenID(),
		Custom: make(map[string]interface{}),
	}

	// Copy custom claims
	for k, v := range claims {
		if k != "exp" && k != "iat" && k != "sub" && k != "iss" && k != "jti" {
			newClaims.Custom[k] = v
		}
	}

	// Generate new token
	return s.GenerateToken(newClaims)
}

// generateRandomTokenID creates a random token ID for uniqueness
func generateRandomTokenID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().Format(time.RFC3339Nano)
	}
	return fmt.Sprintf("%x", b)
}

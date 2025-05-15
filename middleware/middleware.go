// Package middleware provides common HTTP middleware components.
package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lamboktulussimamora/gra/context"
	"github.com/lamboktulussimamora/gra/logger"
	"github.com/lamboktulussimamora/gra/router"
)

// JWTAuthenticator defines an interface for JWT token validation
type JWTAuthenticator interface {
	ValidateToken(tokenString string) (any, error)
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

// CORSConfig contains configuration options for the CORS middleware
type CORSConfig struct {
	AllowOrigins     []string // List of allowed origins (e.g. "http://example.com")
	AllowMethods     []string // List of allowed HTTP methods
	AllowHeaders     []string // List of allowed HTTP headers
	ExposeHeaders    []string // List of headers that are safe to expose
	AllowCredentials bool     // Indicates whether the request can include user credentials
	MaxAge           int      // Indicates how long the results of a preflight request can be cached (in seconds)
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS handles Cross-Origin Resource Sharing with simplified configuration
func CORS(allowOrigin string) router.Middleware {
	config := DefaultCORSConfig()
	config.AllowOrigins = []string{allowOrigin}
	return CORSWithConfig(config)
}

// CORSWithConfig handles Cross-Origin Resource Sharing with custom configuration
func CORSWithConfig(config CORSConfig) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Set CORS headers
			origin := c.GetHeader("Origin")

			// Check if the origin is allowed
			allowedOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowedOrigin = origin
					if o == "*" && origin == "" {
						allowedOrigin = "*"
					}
					break
				}
			}

			if allowedOrigin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			// Set allowed methods
			if len(config.AllowMethods) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			// Set allowed headers
			if len(config.AllowHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			// Set expose headers
			if len(config.ExposeHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			// Set allow credentials
			if config.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set max age
			if config.MaxAge > 0 {
				c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				c.Writer.WriteHeader(http.StatusOK)
				return
			}

			next(c)
		}
	}
}

// RateLimiterStore defines an interface for rate limiter storage
type RateLimiterStore interface {
	// Increment increases the counter for a key, returns the current count and if the limit is exceeded
	Increment(key string, limit int, windowSeconds int) (int, bool)
}

// InMemoryStore implements a simple in-memory store for rate limiting
type InMemoryStore struct {
	data map[string]map[int64]int
	mu   sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store for rate limiting
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]map[int64]int),
	}
}

// Increment increases the counter for a key, returns the current count and if the limit is exceeded
func (s *InMemoryStore) Increment(key string, limit int, windowSeconds int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	windowStart := now - int64(windowSeconds)

	// Initialize counts for this key if not exists
	if _, exists := s.data[key]; !exists {
		s.data[key] = make(map[int64]int)
	}

	// Clean up old entries
	for timestamp := range s.data[key] {
		if timestamp < windowStart {
			delete(s.data[key], timestamp)
		}
	}

	// Count total requests in the time window
	totalRequests := 0
	for _, count := range s.data[key] {
		totalRequests += count
	}

	// Check if limit is exceeded
	exceeded := totalRequests >= limit

	// Only increment if not exceeded
	if !exceeded {
		s.data[key][now]++
		totalRequests++
	}

	return totalRequests, exceeded
}

// RateLimiterConfig contains configuration for the rate limiter
type RateLimiterConfig struct {
	Store        RateLimiterStore              // Store for tracking request counts
	Limit        int                           // Maximum number of requests in the time window
	Window       int                           // Time window in seconds
	KeyFunc      func(*context.Context) string // Function to generate a key from the request
	ExcludeFunc  func(*context.Context) bool   // Function to exclude certain requests from rate limiting
	ErrorMessage string                        // Error message when rate limit is exceeded
}

// RateLimit creates a middleware that limits the number of requests
func RateLimit(limit int, windowSeconds int) router.Middleware {
	store := NewInMemoryStore()

	config := RateLimiterConfig{
		Store:  store,
		Limit:  limit,
		Window: windowSeconds,
		KeyFunc: func(c *context.Context) string {
			// Default to IP-based rate limiting
			return c.Request.RemoteAddr
		},
		ErrorMessage: "Rate limit exceeded. Try again later.",
	}

	return RateLimitWithConfig(config)
}

// RateLimitWithConfig creates a middleware with custom rate limiting configuration
func RateLimitWithConfig(config RateLimiterConfig) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Check if this request should be excluded from rate limiting
			if config.ExcludeFunc != nil && config.ExcludeFunc(c) {
				next(c)
				return
			}

			// Generate key for this request
			key := config.KeyFunc(c)

			// Increment counter and check if limit exceeded
			count, exceeded := config.Store.Increment(key, config.Limit, config.Window)

			// Set RateLimit headers
			c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
			c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(config.Limit-count))
			c.Writer.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Unix())+config.Window))

			if exceeded {
				c.Error(http.StatusTooManyRequests, config.ErrorMessage)
				return
			}

			next(c)
		}
	}
}

// RequestIDConfig contains configuration for the request ID middleware
type RequestIDConfig struct {
	// Generator is a function that generates a request ID
	Generator func() string
	// HeaderName is the header name for the request ID
	HeaderName string
	// ContextKey is the key used to store the request ID in the context
	ContextKey string
	// ResponseHeader determines if the request ID is included in the response headers
	ResponseHeader bool
}

// DefaultRequestIDConfig returns a default request ID configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Generator: func() string {
			// Generate a random UUID-like string
			b := make([]byte, 16)
			_, err := rand.Read(b)
			if err != nil {
				return "req-" + strconv.FormatInt(time.Now().UnixNano(), 36)
			}
			return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
		},
		HeaderName:     "X-Request-ID",
		ContextKey:     "requestID",
		ResponseHeader: true,
	}
}

// RequestID adds a unique request ID to each request
func RequestID() router.Middleware {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

// RequestIDWithConfig adds a unique request ID to each request with custom config
func RequestIDWithConfig(config RequestIDConfig) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Check if there's already a request ID in the headers
			reqID := c.GetHeader(config.HeaderName)

			// If no request ID is provided, generate one
			if reqID == "" {
				reqID = config.Generator()
			}

			// Store the request ID in the context
			c.WithValue(config.ContextKey, reqID)

			// Add the request ID to the response header if configured
			if config.ResponseHeader {
				c.SetHeader(config.HeaderName, reqID)
			}

			// Call the next handler
			next(c)
		}
	}
}

// SecureHeadersConfig holds configuration for secure headers middleware
type SecureHeadersConfig struct {
	XSSProtection             string // X-XSS-Protection header
	ContentTypeNosniff        string // X-Content-Type-Options header
	XFrameOptions             string // X-Frame-Options header
	HSTSMaxAge                int    // Strict-Transport-Security max age in seconds
	HSTSIncludeSubdomains     bool   // Strict-Transport-Security includeSubdomains flag
	HSTSPreload               bool   // Strict-Transport-Security preload flag
	ContentSecurityPolicy     string // Content-Security-Policy header
	ReferrerPolicy            string // Referrer-Policy header
	PermissionsPolicy         string // Permissions-Policy header
	CrossOriginEmbedderPolicy string // Cross-Origin-Embedder-Policy header
	CrossOriginOpenerPolicy   string // Cross-Origin-Opener-Policy header
	CrossOriginResourcePolicy string // Cross-Origin-Resource-Policy header
}

// DefaultSecureHeadersConfig returns a default configuration for secure headers
func DefaultSecureHeadersConfig() SecureHeadersConfig {
	return SecureHeadersConfig{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		HSTSMaxAge:                31536000, // 1 year
		HSTSIncludeSubdomains:     true,
		HSTSPreload:               false,
		ContentSecurityPolicy:     "", // Empty by default, should be configured by user
		ReferrerPolicy:            "no-referrer",
		PermissionsPolicy:         "",
		CrossOriginEmbedderPolicy: "",
		CrossOriginOpenerPolicy:   "",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// SecureHeaders adds security-related headers to the response
func SecureHeaders() router.Middleware {
	return SecureHeadersWithConfig(DefaultSecureHeadersConfig())
}

// SecureHeadersWithConfig adds security-related headers to the response with custom configuration
func SecureHeadersWithConfig(config SecureHeadersConfig) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *context.Context) {
			// Set security headers before processing the request
			setSecurityHeaders(c.Writer, config)

			// Call the next handler
			next(c)
		}
	}
}

// setSecurityHeaders applies all configured security headers to the response
func setSecurityHeaders(w http.ResponseWriter, config SecureHeadersConfig) {
	// Apply basic security headers
	setBasicSecurityHeaders(w, config)

	// Apply HSTS header if configured
	setHSTSHeader(w, config)

	// Apply content security headers
	setContentSecurityHeaders(w, config)

	// Apply cross-origin security headers
	setCrossOriginHeaders(w, config)
}

// setBasicSecurityHeaders applies the basic security headers
func setBasicSecurityHeaders(w http.ResponseWriter, config SecureHeadersConfig) {
	// X-XSS-Protection header
	if config.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", config.XSSProtection)
	}

	// X-Content-Type-Options header
	if config.ContentTypeNosniff != "" {
		w.Header().Set("X-Content-Type-Options", config.ContentTypeNosniff)
	}

	// X-Frame-Options header
	if config.XFrameOptions != "" {
		w.Header().Set("X-Frame-Options", config.XFrameOptions)
	}

	// Referrer-Policy header
	if config.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
	}
}

// setHSTSHeader constructs and applies the HSTS header
func setHSTSHeader(w http.ResponseWriter, config SecureHeadersConfig) {
	// Strict-Transport-Security header
	if config.HSTSMaxAge > 0 {
		hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
		if config.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if config.HSTSPreload {
			hstsValue += "; preload"
		}
		w.Header().Set("Strict-Transport-Security", hstsValue)
	}
}

// setContentSecurityHeaders applies content security related headers
func setContentSecurityHeaders(w http.ResponseWriter, config SecureHeadersConfig) {
	// Content-Security-Policy header
	if config.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
	}

	// Permissions-Policy header
	if config.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
	}
}

// setCrossOriginHeaders applies cross-origin related security headers
func setCrossOriginHeaders(w http.ResponseWriter, config SecureHeadersConfig) {
	// Cross-Origin-Embedder-Policy header
	if config.CrossOriginEmbedderPolicy != "" {
		w.Header().Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
	}

	// Cross-Origin-Opener-Policy header
	if config.CrossOriginOpenerPolicy != "" {
		w.Header().Set("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
	}

	// Cross-Origin-Resource-Policy header
	if config.CrossOriginResourcePolicy != "" {
		w.Header().Set("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
	}
}

// CSPBuilder helps to build a Content Security Policy (CSP) string
type CSPBuilder struct {
	directives map[string][]string
}

// NewCSPBuilder creates a new CSP builder with default directives
func NewCSPBuilder() *CSPBuilder {
	return &CSPBuilder{
		directives: make(map[string][]string),
	}
}

// AddDirective adds a directive with values to the CSP
func (b *CSPBuilder) AddDirective(directive string, values ...string) *CSPBuilder {
	if len(values) > 0 {
		if _, exists := b.directives[directive]; !exists {
			b.directives[directive] = []string{}
		}
		b.directives[directive] = append(b.directives[directive], values...)
	}
	return b
}

// DefaultSrc sets the default-src directive
func (b *CSPBuilder) DefaultSrc(values ...string) *CSPBuilder {
	return b.AddDirective("default-src", values...)
}

// ScriptSrc sets the script-src directive
func (b *CSPBuilder) ScriptSrc(values ...string) *CSPBuilder {
	return b.AddDirective("script-src", values...)
}

// StyleSrc sets the style-src directive
func (b *CSPBuilder) StyleSrc(values ...string) *CSPBuilder {
	return b.AddDirective("style-src", values...)
}

// ImgSrc sets the img-src directive
func (b *CSPBuilder) ImgSrc(values ...string) *CSPBuilder {
	return b.AddDirective("img-src", values...)
}

// ConnectSrc sets the connect-src directive
func (b *CSPBuilder) ConnectSrc(values ...string) *CSPBuilder {
	return b.AddDirective("connect-src", values...)
}

// FontSrc sets the font-src directive
func (b *CSPBuilder) FontSrc(values ...string) *CSPBuilder {
	return b.AddDirective("font-src", values...)
}

// ObjectSrc sets the object-src directive
func (b *CSPBuilder) ObjectSrc(values ...string) *CSPBuilder {
	return b.AddDirective("object-src", values...)
}

// MediaSrc sets the media-src directive
func (b *CSPBuilder) MediaSrc(values ...string) *CSPBuilder {
	return b.AddDirective("media-src", values...)
}

// FrameSrc sets the frame-src directive
func (b *CSPBuilder) FrameSrc(values ...string) *CSPBuilder {
	return b.AddDirective("frame-src", values...)
}

// WorkerSrc sets the worker-src directive
func (b *CSPBuilder) WorkerSrc(values ...string) *CSPBuilder {
	return b.AddDirective("worker-src", values...)
}

// FrameAncestors sets the frame-ancestors directive
func (b *CSPBuilder) FrameAncestors(values ...string) *CSPBuilder {
	return b.AddDirective("frame-ancestors", values...)
}

// FormAction sets the form-action directive
func (b *CSPBuilder) FormAction(values ...string) *CSPBuilder {
	return b.AddDirective("form-action", values...)
}

// ReportTo sets the report-to directive
func (b *CSPBuilder) ReportTo(value string) *CSPBuilder {
	return b.AddDirective("report-to", value)
}

// ReportURI sets the report-uri directive
func (b *CSPBuilder) ReportURI(value string) *CSPBuilder {
	return b.AddDirective("report-uri", value)
}

// UpgradeInsecureRequests adds the upgrade-insecure-requests directive
func (b *CSPBuilder) UpgradeInsecureRequests() *CSPBuilder {
	return b.AddDirective("upgrade-insecure-requests", "")
}

// Build builds the CSP string
func (b *CSPBuilder) Build() string {
	parts := []string{}

	for directive, values := range b.directives {
		if len(values) == 0 || (len(values) == 1 && values[0] == "") {
			// Handle directives without values (like upgrade-insecure-requests)
			parts = append(parts, directive)
		} else {
			// Handle directives with values
			part := directive + " " + strings.Join(values, " ")
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, "; ")
}

// CSP creates a middleware that sets the Content-Security-Policy header
func CSP(builder *CSPBuilder) router.Middleware {
	config := DefaultSecureHeadersConfig()
	config.ContentSecurityPolicy = builder.Build()
	return SecureHeadersWithConfig(config)
}

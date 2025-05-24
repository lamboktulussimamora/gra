# Middleware in GRA Framework

Middleware functions provide a way to execute code before or after request handling. They are useful for authentication, logging, error handling, and more.

## Table of Contents

- [Using Middleware](#using-middleware)
- [Built-in Middleware](#built-in-middleware)
  - [Logger](#logger)
  - [Recovery](#recovery)
  - [CORS](#cors)
  - [Secure Headers](#secure-headers)
  - [Authentication](#authentication)
  - [Cache](#cache)
- [Creating Custom Middleware](#creating-custom-middleware)

## Using Middleware

Middleware can be applied at the router level, group level, or route level:

### Router-Level Middleware

Applied to all routes:

```go
r := gra.New()
r.Use(
    middleware.Logger(),
    middleware.Recovery(),
    middleware.CORS("*"),
)
```

### Group-Level Middleware

Applied to all routes in a group:

```go
api := r.Group("/api")
api.Use(
    middleware.Auth(jwtService, "user"),
)
```

### Route-Level Middleware

Applied to a specific route:

```go
r.GET("/admin", middleware.Auth(jwtService, "admin"), adminHandler)
```

## Built-in Middleware

GRA comes with several pre-built middleware functions:

### Logger

Logs HTTP request and response details:

```go
r.Use(middleware.Logger())
```

Example output:
```
[GRA] 2023/06/21 - 12:34:56 | 200 | 2.345ms | 127.0.0.1 | GET /api/users
```

#### Configuration

You can customize the logger:

```go
config := middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
    TimeFormat: "2006-01-02 15:04:05",
}
r.Use(middleware.LoggerWithConfig(config))
```

### Recovery

Recovers from panics and returns a 500 response:

```go
r.Use(middleware.Recovery())
```

#### Configuration

You can customize recovery behavior:

```go
config := middleware.RecoveryConfig{
    OnPanic: func(c *gra.Context, err interface{}) {
        // Custom panic handling
        log.Printf("Panic: %v", err)
        c.Error(http.StatusInternalServerError, "Internal Server Error")
    },
}
r.Use(middleware.RecoveryWithConfig(config))
```

### CORS

Configures Cross-Origin Resource Sharing:

```go
// Allow all origins
r.Use(middleware.CORS("*"))

// Allow specific origin
r.Use(middleware.CORS("https://example.com"))

// Custom configuration
config := middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * 3600, // 12 hours
}
r.Use(middleware.CORSWithConfig(config))
```

### Secure Headers

Adds security-related HTTP headers to responses:

```go
r.Use(middleware.SecureHeaders())
```

Default headers added:
- X-XSS-Protection: 1; mode=block
- X-Content-Type-Options: nosniff
- X-Frame-Options: SAMEORIGIN
- Content-Security-Policy: default-src 'self'
- Referrer-Policy: strict-origin-when-cross-origin
- Strict-Transport-Security: max-age=31536000; includeSubDomains
- Cache-Control: no-store, must-revalidate

#### Configuration

You can customize the headers:

```go
config := middleware.SecureHeadersConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeOptions:    "nosniff",
    XFrameOptions:         "DENY",
    ContentSecurityPolicy: "default-src 'self'; script-src 'self' https://trusted.cdn.com",
    ReferrerPolicy:        "no-referrer",
    HSTS:                  "max-age=63072000; includeSubDomains; preload",
    CacheControl:          "no-cache, no-store, must-revalidate",
}
r.Use(middleware.SecureHeadersWithConfig(config))
```

### Authentication

Authenticates requests using JSON Web Tokens:

```go
// Create JWT service
jwtService, err := jwt.NewServiceWithKey([]byte("your-secret-key"))
if err != nil {
    log.Fatalf("Failed to create JWT service: %v", err)
}

// Use JWT middleware
r.Use(middleware.Auth(jwtService, "user"))
```

The middleware will:
1. Check for the Authorization header with the Bearer token
2. Validate the token
3. Add user information to the context
4. Call the next handler if authentication succeeds
5. Return a 401 error if authentication fails

#### Configuration

You can customize the authentication behavior:

```go
config := middleware.AuthConfig{
    JWTService:    jwtService,
    UserType:      "admin",
    TokenSource:   middleware.TokenSourceHeader, // or TokenSourceCookie, TokenSourceQuery
    TokenLookup:   "Authorization",
    TokenPrefix:   "Bearer",
    ContextKey:    "user",
    ErrorHandler:  customErrorHandler,
}
r.Use(middleware.AuthWithConfig(config))
```

### Cache

Caches responses to improve performance:

```go
// Cache responses for 5 minutes
r.Use(middleware.Cache(5 * time.Minute))

// Custom configuration
config := middleware.CacheConfig{
    Duration:   10 * time.Minute,
    SkipPaths:  []string{"/users", "/admin"},
    Methods:    []string{"GET"},
    KeyPrefix:  "api-cache:",
    KeyFunc:    customKeyFunction,
    Store:      customCacheStore, // Implements cache.Store interface
}
r.Use(middleware.CacheWithConfig(config))
```

## Creating Custom Middleware

You can create custom middleware by returning a `gra.HandlerFunc`:

```go
func RateLimiter(rps int, burst int) gra.HandlerFunc {
    // Create a rate limiter
    limiter := rate.NewLimiter(rate.Limit(rps), burst)
    
    // Return the middleware handler
    return func(c *gra.Context) {
        // Check if the request can proceed
        if !limiter.Allow() {
            c.Error(http.StatusTooManyRequests, "Too many requests")
            return
        }
        
        // Call the next handler
        c.Next()
    }
}

// Use the custom middleware
r.Use(RateLimiter(10, 30))
```

### Middleware Best Practices

1. **Call `c.Next()` to continue** the request chain (unless you want to short-circuit)
2. **Check `c.IsAborted()`** before executing code after `c.Next()`
3. **Use `c.Set()`** to share data between middleware and handlers
4. **Handle errors properly** and set appropriate status codes
5. **Keep middleware focused** on a single responsibility
6. **Order middleware correctly** (e.g., Recovery should be first)

```go
func MyMiddleware() gra.HandlerFunc {
    return func(c *gra.Context) {
        // Code executed before the request
        
        c.Next() // Call the next handler
        
        // Code executed after the request (if not aborted)
        if !c.IsAborted() {
            // Post-processing
        }
    }
}
```

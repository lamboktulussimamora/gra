# GRA Framework

[![Test and Coverage](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml/badge.svg)](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/lamboktulussimamora/gra/badge.svg?branch=main)](https://coveralls.io/github/lamboktulussimamora/gra?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/lamboktulussimamora/gra)](https://goreportcard.com/report/github.com/lamboktulussimamora/gra)
[![Go Reference](https://pkg.go.dev/badge/github.com/lamboktulussimamora/gra.svg)](https://pkg.go.dev/github.com/lamboktulussimamora/gra)
[![GitHub release](https://img.shields.io/github/release/lamboktulussimamora/gra.svg)](https://GitHub.com/lamboktulussimamora/gra/releases/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight HTTP framework for building web applications in Go, inspired by Gin.

## Features

- Context-based request handling
- HTTP routing with path parameters
- JWT authentication and authorization
- Secure HTTP headers middleware
- API versioning support
- Response caching
- Middleware support
- Request validation
- Standardized API responses
- Structured logging
- **Enhanced Entity Framework Core-like ORM**
- **Automatic database migrations**
- **Hybrid Migration System** with model-driven migration generation
- **LINQ-style querying with generics**
- **Change tracking and entity states**
- **Transaction management**
- **Database schema inspection and comparison**
- Clean architecture friendly

## Installation

```bash
go get github.com/lamboktulussimamora/gra
```

## Quick Start

```go
package main

import (
	"net/http"
	"github.com/lamboktulussimamora/gra"
)

func main() {
	// Create a new router
	r := gra.New()

	// Define a route
	r.GET("/hello", func(c *gra.Context) {
		c.Success(http.StatusOK, "Hello World", nil)
	})

	// Start the server
	gra.Run(":8080", r)
}
```

## Documentation

Full documentation is available at [https://lamboktulussimamora.github.io/gra/](https://lamboktulussimamora.github.io/gra/)

- [Getting Started Guide](https://lamboktulussimamora.github.io/gra/getting-started/)
- [API Reference](https://lamboktulussimamora.github.io/gra/api-reference/)
- [Examples](https://lamboktulussimamora.github.io/gra/examples/)

## Context

The `Context` provides a convenient way to handle HTTP requests and responses:

```go
// Get path parameters
id := c.GetParam("id")

// Get query parameters
name := c.GetQuery("name")

// Parse JSON request body
var user User
if err := c.BindJSON(&user); err != nil {
	c.Error(http.StatusBadRequest, "Invalid request")
	return
}

// Send JSON response
c.JSON(http.StatusOK, map[string]any{
	"message": "Success",
})

// Send standardized success response
c.Success(http.StatusOK, "User created", user)

// Send standardized error response
c.Error(http.StatusNotFound, "User not found")
```

## Router

The `Router` handles HTTP routing:

```go
r := core.New()

// Register routes
r.GET("/users", listUsers)
r.POST("/users", createUser)
r.GET("/users/:id", getUser)
r.PUT("/users/:id", updateUser)
r.DELETE("/users/:id", deleteUser)
```

### Route Groups

You can group routes with a common prefix:

```go
// Create an API group
api := r.Group("/api")

// Add routes to the group
api.GET("/users", listUsers)
api.POST("/users", createUser)

// Create nested groups
v1 := api.Group("/v1")
v1.GET("/products", listProductsV1)

v2 := api.Group("/v2") 
v2.GET("/products", listProductsV2)
```

The above code will create these routes:
- `/api/users` (GET, POST)
- `/api/v1/products` (GET)
- `/api/v2/products` (GET)

## Middleware

Middleware functions can be used to add common functionality:

```go
// Use global middleware
r.Use(
	middleware.Logger(),
	middleware.Recovery(),
	middleware.CORS("*"),
	middleware.SecureHeaders(),
)

// Apply middleware to a specific group
authRoutes := r.Group("/api")
authRoutes.Use(middleware.Auth(jwtService, "user"))
```

### Available Middleware

1. **Logger**: Logs HTTP requests and responses
2. **Recovery**: Recovers from panics in handlers
3. **CORS**: Configures Cross-Origin Resource Sharing
4. **Auth**: JWT authentication middleware
5. **SecureHeaders**: Adds security-related HTTP headers
6. **Cache**: HTTP response caching (see Cache section)

### JWT Authentication

The JWT middleware authenticates requests using JSON Web Tokens:

```go
// Create JWT service
jwtService, _ := jwt.NewServiceWithKey([]byte("your-secret-key"))

// Use JWT middleware
protectedRoutes.Use(middleware.Auth(jwtService, "user"))
```

### Secure Headers

The secure headers middleware adds security-related HTTP headers:

```go
// Use default secure headers
app.Use(middleware.SecureHeaders())

// Or use custom configuration
config := middleware.DefaultSecureHeadersConfig()
config.ContentSecurityPolicy = "default-src 'self'"
config.XFrameOptions = "DENY"
app.Use(middleware.SecureHeadersWithConfig(config))
```

Included headers:
- X-XSS-Protection
- X-Content-Type-Options
- X-Frame-Options
- Strict-Transport-Security (HSTS)
- Content-Security-Policy
- Referrer-Policy
- Cross-Origin-Resource-Policy

## Validation

Validate request data using struct tags:

```go
type User struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"enum=user,admin,guest"`
	Phone    string `json:"phone" validate:"regexp=^[0-9]{10}$"`
}

func createUser(c *core.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.Error(http.StatusBadRequest, "Invalid request")
		return
	}

	v := validator.New()
	errors := v.Validate(user)
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, map[string]any{
			"status": "error",
			"errors": errors,
		})
		return
	}

	// Process validated user...
}
```

### Validation Rules

The validator supports the following validation rules:

- `required`: Field cannot be empty/zero
- `email`: Field must be a valid email address
- `min=X`: String length or number value must be at least X
- `max=X`: String length or number value must be at most X
- `regexp=pattern`: String must match the specified regular expression pattern
- `enum=val1,val2,val3`: String must be one of the specified values
- `range=min,max`: Number must be within the specified inclusive range

### Custom Error Messages

You can specify custom error messages for validation rules:

```go
type Product struct {
    Name  string `json:"name" validate:"required|Name is mandatory"`
    Price int    `json:"price" validate:"range=1,10000|Price must be between $1 and $10,000"`
}
```

### Batch Validation

You can validate multiple objects at once:

```go
// Products to validate
products := []any{
    Product{Name: "Product 1", Price: 100},
    Product{Name: "", Price: -5},  // Invalid
    Product{Name: "Product 3", Price: 9999},
}

v := validator.New()
results := v.ValidateBatch(products)

// Check if any validation errors
if v.HasBatchErrors(results) {
    // Get only invalid results
    invalid := v.FilterInvalid(results)
    
    for _, result := range invalid {
        fmt.Printf("Item at index %d has %d errors\n", result.Index, len(result.Errors))
        for _, err := range result.Errors {
            fmt.Printf("  - Field: %s, Error: %s\n", err.Field, err.Message)
        }
    }
}
```

## Examples

See the `examples` directory for more usage examples.

## Testing

The framework has comprehensive test coverage:

- Overall test coverage: 90.2%
- Full test coverage for middleware and adapter packages
- High coverage for router (95.7%), context (90.0%), logger (87.9%), and validator (82.6%)

To run tests:

```bash
go test ./...
```

To check test coverage:

```bash
go test ./... -cover
```

## API Versioning

API versioning helps you maintain backward compatibility while evolving your API. The versioning package provides multiple strategies for API versioning:

```go
import (
    "github.com/lamboktulussimamora/gra/versioning"
)

// Create a new router
r := gra.New()

// Set up versioning with URL path strategy (default)
// This will handle URLs like /v1/users, /v2/users, etc.
v := versioning.New().
    WithSupportedVersions("1", "2").
    WithDefaultVersion("1")

// Apply versioning middleware
r.Use(v.Middleware())
```

### Versioning Strategies

You can choose from different versioning strategies:

#### Path Versioning

Uses the URL path to specify the version:

```go
// /v1/users, /v2/users, etc.
v := versioning.New().
    WithStrategy(&versioning.PathVersionStrategy{Prefix: "v"})
```

#### Query Parameter Versioning

Uses a query parameter to specify the version:

```go
// /users?version=1, /users?v=2, etc.
v := versioning.New().
    WithStrategy(&versioning.QueryVersionStrategy{ParamName: "version"})
```

#### Header Versioning

Uses a custom HTTP header to specify the version:

```go
// HTTP Header: Accept-Version: 1
v := versioning.New().
    WithStrategy(&versioning.HeaderVersionStrategy{HeaderName: "Accept-Version"})
```

#### Media Type Versioning

Uses the Accept header with vendor media type to specify the version:

```go
// HTTP Header: Accept: application/vnd.api.v1+json
v := versioning.New().
    WithStrategy(&versioning.MediaTypeVersionStrategy{MediaTypePrefix: "application/vnd."})
```

### Accessing Version Information

You can access the API version in your handlers:

```go
func getUser(c *gra.Context) {
    // Get version info
    versionInfo, exists := versioning.GetAPIVersion(c)
    if exists {
        fmt.Println("API Version:", versionInfo.Version)
    }
    
    // Handle request normally...
}
```

## Response Caching

The cache middleware improves performance by caching responses to GET requests:

```go
import (
    "github.com/lamboktulussimamora/gra/cache"
)

// Create a new router
r := gra.New()

// Add cache middleware with default settings (5-minute TTL)
r.Use(cache.New())
```

### Custom Cache Configuration

Configure caching behavior:

```go
// Create a custom cache configuration
config := cache.DefaultCacheConfig()
config.TTL = time.Minute * 10  // Set TTL to 10 minutes
config.Methods = []string{http.MethodGet, http.MethodHead}  // Cache GET and HEAD requests
config.MaxBodySize = 5 * 1024 * 1024  // Increase max cache size to 5MB

// Use custom configuration
r.Use(cache.WithConfig(config))
```

### Cache Stores

The default in-memory store works for single-instance applications. For distributed applications, you can implement the `CacheStore` interface:

```go
type MyCacheStore struct {
    // Your implementation details
}

// Implement the CacheStore interface methods
func (s *MyCacheStore) Get(key string) (*cache.CacheEntry, bool) {
    // Your implementation
}

func (s *MyCacheStore) Set(key string, entry *cache.CacheEntry, ttl time.Duration) {
    // Your implementation
}

func (s *MyCacheStore) Delete(key string) {
    // Your implementation
}

func (s *MyCacheStore) Clear() {
    // Your implementation
}

// Use your custom store
config := cache.DefaultCacheConfig()
config.Store = &MyCacheStore{}
r.Use(cache.WithConfig(config))
```

### Cache Control

Manually control cache behavior:

```go
// Clear entire cache
cache.ClearCache(myStore)

// Invalidate specific entry
cache.InvalidateCache(myStore, "GET:/api/users/123")
```

## JWT Authentication

The JWT (JSON Web Tokens) package provides authentication functionality:

```go
import "github.com/lamboktulussimamora/gra/jwt"
```

### Creating a JWT Service

```go
// Create with a signing key
jwtService, err := jwt.NewServiceWithKey([]byte("your-secret-signing-key"))

// Or create with custom configuration
config := jwt.DefaultConfig()
config.SigningKey = []byte("your-secret-key")
config.ExpirationTime = time.Hour * 48  // 48 hours
config.Issuer = "my-application"
jwtService, err := jwt.NewService(config)
```

### Generating Tokens

```go
// Create claims
claims := jwt.StandardClaims{
    Subject: "user-123",  // Required
    Custom: map[string]interface{}{
        "username": "johndoe",
        "role": "admin",
    },
}

// Generate token
token, err := jwtService.GenerateToken(claims)
```

### Validating Tokens

```go
// Validate a token
claims, err := jwtService.ValidateToken(tokenString)
if err != nil {
    // Handle error: jwt.ErrInvalidToken, jwt.ErrExpiredToken
    return
}

// Access claims
userID := claims["sub"].(string)
username := claims["username"].(string)
role := claims["role"].(string)
```

### Refreshing Tokens

```go
// Refresh a token (generate new token with same claims)
newToken, err := jwtService.RefreshToken(oldTokenString)
```

### Using with Middleware

```go
// Protect routes with JWT authentication
protectedRoutes := r.Group("/api")
protectedRoutes.Use(middleware.Auth(jwtService, "user"))

// Access claims in your handlers
func getUserProfile(c *context.Context) {
    // Get user claims from context
    userClaims := c.Value("user").(map[string]interface{})
    userID := userClaims["sub"].(string)
    
    // Handle request
    // ...
}
```

## Enhanced ORM System

GRA includes a comprehensive Entity Framework Core-inspired ORM system that provides advanced database operations with automatic migrations, change tracking, and LINQ-style querying.

### Key Features

- **Entity Framework Core-like architecture** with DbContext and DbSet patterns
- **Automatic database migrations** based on struct definitions
- **LINQ-style querying** with Go generics support
- **Change tracking** with entity states (Added, Modified, Deleted, Unchanged)
- **Transaction management** with SaveChanges pattern
- **Fluent query interface** (Where, OrderBy, Take, Skip, Include)
- **Advanced query execution** (First, FirstOrDefault, Single, Any, ToList, Count)
- **Read-only queries** with AsNoTracking support
- **Relationship management** with foreign key support

### Quick Start

```go
import (
    "github.com/lamboktulussimamora/gra/orm/dbcontext"
    "github.com/lamboktulussimamora/gra/orm/models"
    "github.com/lamboktulussimamora/gra/orm/migrations"
)

// Define your entities (already included)
type User struct {
    models.BaseEntity
    FirstName string `db:"first_name"`
    LastName  string `db:"last_name"`
    Email     string `db:"email"`
    IsActive  bool   `db:"is_active"`
}

// Initialize database context
db, err := sql.Open("sqlite3", "app.db")
if err != nil {
    log.Fatal(err)
}

ctx := dbcontext.NewEnhancedDbContext(db)

// Run automatic migrations
migrationRunner := migrations.NewMigrationRunner(db)
err = migrationRunner.AutoMigrate(
    &models.User{},
    &models.Product{},
    &models.Category{},
    // ... other entities
)
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

### Basic CRUD Operations

```go
// Create new entities
user := &models.User{
    FirstName: "John",
    LastName:  "Doe",
    Email:     "john.doe@example.com",
    IsActive:  true,
}

// Add to context (tracks as "Added")
err := ctx.Add(user)
if err != nil {
    log.Fatal(err)
}

// Save changes to database
err = ctx.SaveChanges()
if err != nil {
    log.Fatal(err)
}

// Read operations
userSet := dbcontext.EnhancedSet[models.User](ctx)

// Find by ID
foundUser, err := userSet.Where("id = ?", user.ID).FirstOrDefault()
if err != nil {
    log.Fatal(err)
}

// Update entity
foundUser.Email = "newemail@example.com"
err = ctx.Update(foundUser)
err = ctx.SaveChanges()

// Delete entity
err = ctx.Delete(foundUser)
err = ctx.SaveChanges()
```

### Advanced Querying

```go
userSet := dbcontext.EnhancedSet[models.User](ctx)

// LINQ-style querying with method chaining
activeUsers, err := userSet.
    Where("is_active = ?", true).
    OrderBy("first_name").
    Take(10).
    ToList()

// Complex queries with multiple conditions
specificUsers, err := userSet.
    Where("is_active = ? AND email LIKE ?", true, "%@company.com").
    OrderByDescending("created_at").
    Skip(20).
    Take(10).
    ToList()

// Check existence
hasActiveUsers, err := userSet.
    Where("is_active = ?", true).
    Any()

// Count records
totalUsers, err := userSet.Count()
activeUserCount, err := userSet.
    Where("is_active = ?", true).
    Count()

// Single record operations
firstUser, err := userSet.
    OrderBy("created_at").
    First() // Throws error if no results

firstUserOrNil, err := userSet.
    Where("email = ?", "specific@email.com").
    FirstOrDefault() // Returns nil if no results
```

### Change Tracking

```go
// Enable change tracking (default)
user, err := userSet.Where("id = ?", 1).First()

// Modify entity
user.Email = "updated@example.com"
user.IsActive = false

// Context automatically tracks changes
fmt.Printf("Entity State: %v\n", ctx.ChangeTracker.GetEntityState(user))
// Output: Modified

// Save all tracked changes
err = ctx.SaveChanges()

// Read-only queries (no change tracking)
readOnlyUsers, err := userSet.
    AsNoTracking().
    Where("is_active = ?", true).
    ToList()
// These entities won't be tracked for changes
```

### Transaction Management

```go
// Begin transaction
tx, err := ctx.Database.Begin()
if err != nil {
    log.Fatal(err)
}

// Create transaction context
txCtx := dbcontext.NewEnhancedDbContextWithTx(tx)

// Perform operations within transaction
user1 := &models.User{FirstName: "User", LastName: "One", Email: "user1@example.com"}
user2 := &models.User{FirstName: "User", LastName: "Two", Email: "user2@example.com"}

err = txCtx.Add(user1)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

err = txCtx.Add(user2)
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

// Save all changes in transaction
err = txCtx.SaveChanges()
if err != nil {
    tx.Rollback()
    log.Fatal(err)
}

// Commit transaction
err = tx.Commit()
if err != nil {
    log.Fatal(err)
}
```

### Migration System

GRA provides multiple migration approaches to suit different development workflows:

#### 1. Automatic Migration System

The automatic migration system creates and updates database tables based on your entity definitions:

```go
// Create migration runner
migrationRunner := migrations.NewMigrationRunner(db)

// Register entities for migration
entities := []interface{}{
    &models.User{},
    &models.Product{},
    &models.Category{},
    &models.Order{},
    &models.OrderItem{},
    &models.Review{},
    &models.Role{},
    &models.UserRole{},
}

// Run automatic migrations
err := migrationRunner.AutoMigrate(entities...)
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

#### 2. Hybrid Migration System (New in v1.0.7)

The hybrid migration system combines model-driven development with explicit migration file generation:

```go
import (
    "github.com/lamboktulussimamora/gra/orm/migrations"
)

// Create hybrid migrator
migrator := migrations.NewHybridMigrator(db, config)

// Register your models
registry := migrations.NewModelRegistry()
registry.RegisterModel(&models.User{})
registry.RegisterModel(&models.Product{})
registry.RegisterModel(&models.Category{})

// Check migration status (automatically initializes schema)
status, err := migrator.GetMigrationStatus()
if err != nil {
    log.Fatal(err)
}

// Generate migration from model changes
migration, err := migrator.CreateMigration("create_initial_schema", "Initial database schema")
if err != nil {
    log.Fatal(err)
}

// Apply migrations
err = migrator.ApplyMigrations()
if err != nil {
    log.Fatal(err)
}
```

#### Key Features of Hybrid Migration System:
- **Model Registration**: Register Go structs as database models
- **Change Detection**: Automatic detection of schema differences
- **SQL Generation**: Automated DDL generation (CREATE TABLE, ALTER TABLE, etc.)
- **Migration Files**: Generates proper up/down migration scripts
- **Database Inspection**: Multi-database schema comparison
- **Safety Checks**: Warns about destructive changes

#### 3. Entity Framework Core-like CLI

For teams familiar with Entity Framework Core, use the CLI tool:

```bash
# Install the CLI tool
go install github.com/lamboktulussimamora/gra/tools/ef-migrate

# Create new migration
ef-migrate add-migration CreateUsersTable

# Apply migrations
ef-migrate update-database

# Check migration status
ef-migrate status

# Rollback to specific migration
ef-migrate rollback InitialCreate
```

### Entity Relationships

Define relationships using struct tags and foreign keys:

```go
type Order struct {
    models.BaseEntity
    UserID     uint           `db:"user_id" fk:"users(id)"`
    User       *User          `db:"-"` // Navigation property
    OrderItems []OrderItem    `db:"-"` // Collection navigation
    Total      float64        `db:"total"`
    Status     string         `db:"status"`
}

type OrderItem struct {
    models.BaseEntity
    OrderID   uint    `db:"order_id" fk:"orders(id)"`
    ProductID uint    `db:"product_id" fk:"products(id)"`
    Order     *Order  `db:"-"`
    Product   *Product `db:"-"`
    Quantity  int     `db:"quantity"`
    Price     float64 `db:"price"`
}
```

### Backward Compatibility

The enhanced ORM maintains compatibility with existing code through a compatibility wrapper:

```go
// Legacy usage still works
ctx := dbcontext.NewDbContext(db)
userDbSet := ctx.Set(&models.User{})

users, err := userDbSet.Where("is_active = ?", true).ToList()
```

### Complete Examples

The framework includes comprehensive examples demonstrating different aspects:

#### 1. Enhanced ORM Demo
See the comprehensive example at `examples/comprehensive-orm-demo/` for a full demonstration including:

- Database setup and migrations
- CRUD operations with change tracking
- Advanced querying scenarios
- Transaction management
- Error handling patterns
- Performance optimization with AsNoTracking

```bash
# Run the comprehensive demo
cd examples/comprehensive-orm-demo
go run main.go
```

#### 2. Hybrid Migration Demo (New in v1.0.7)
Complete demonstration of the hybrid migration system at `examples/hybrid-migration-demo/`:

- Model registration and management
- Automatic change detection
- Migration file generation
- Schema comparison and validation
- Multi-database support

```bash
# Run the hybrid migration demo
cd examples/hybrid-migration-demo
go run demo.go
```

#### 3. Manual Migration System
Traditional migration approach at `examples/manual_migrations/`:

- Direct SQL migration files
- Shell script-based migration runner
- Version tracking and rollback support

```bash
# Run manual migration example
cd examples/manual_migrations
./db_migrate_v2.sh status
```
cd examples/comprehensive-orm-demo
go run main.go
```

For detailed implementation examples and advanced usage patterns, refer to the example documentation in the `examples/comprehensive-orm-demo/README.md` file.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -am 'Add my feature'`)
6. Push to the branch (`git push origin feature/my-feature`)
7. Create a new Pull Request

### Code Quality

Before submitting your code, please:

1. Run `make verify` to check code formatting and detect issues
2. Ensure all tests pass with `make test`
3. Run `make clean` to clean up any temporary files

## Development

### Prerequisites

- Go 1.24 or later
- Make

> **Note:** As of v1.0.6 (May 2025), GRA framework requires Go 1.24 or later. Previous versions required Go 1.21+. The framework takes advantage of the latest Go features and improvements for better performance and security.

### Go Version Compatibility

GRA framework follows these compatibility principles:

- **Minimum Go Version:** Each release specifies the minimum Go version required
- **Testing:** All components are continuously tested against the minimum supported Go version
- **New Feature Adoption:** We adopt new Go language features when they provide significant benefits
- **Migration Path:** Major version changes are documented in release notes with clear migration steps

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make race

# Generate coverage report
make coverage
```

### Project Cleanup

The project includes a cleanup system to maintain a clean codebase:

```bash
# Clean up generated files, backups, and binaries
make clean

# For a more thorough cleanup before release
./scripts/clean_project.sh
```

The cleanup removes:
- Coverage output files (*.out)
- Benchmark results
- Temporary files (*.bak, *.new, *.tmp)

## Benchmarks

GRA framework is designed with performance in mind. Here are some benchmark results from our testing:

```
BenchmarkRouterSimple/SimpleRoute-8         	  368399	      3249 ns/op
BenchmarkRouterSimple/ParameterizedRoute-8  	  293060	      4102 ns/op
BenchmarkRouterComplex/ManyRoutes_Simple-8  	  230602	      5204 ns/op
BenchmarkRouterComplex/ManyRoutes_WithParameter-8 	  183795	      6518 ns/op
BenchmarkRouterComplex/DeepNestedParameters-8    	  147219	      8142 ns/op
```

To run the benchmarks yourself:

```bash
cd router
go test -bench=. -benchmem
```

## Community and Support

- **GitHub Discussions**: Ask questions and share ideas
- **Issue Tracker**: Report bugs or request features
- **Pull Requests**: Contributions are welcome following our guidelines

Join our community to help shape the future of GRA framework!
- Compiled binaries in the examples directory
- Editor backup files

## License

MIT

# Basic GRA Framework Example

This example demonstrates the fundamental features of the GRA framework, including routing, middleware, validation, and JSON handling.

## 🚀 Quick Start

```bash
# Navigate to the basic example directory
cd examples/basic

# Run the example
go run main.go
```

The server will start on `http://localhost:8080` with the following endpoints:

## 📚 API Endpoints

### GET `/`
Returns API information and available endpoints.

**Response:**
```json
{
  "status": "success",
  "message": "Welcome to GRA Framework",
  "data": {
    "version": "1.0.3",
    "time": "2025-06-24T22:39:15.123456789+07:00",
    "endpoints": [
      {"method": "GET", "path": "/", "description": "API information"},
      {"method": "GET", "path": "/users/:id", "description": "Get user by ID"},
      {"method": "POST", "path": "/users", "description": "Create new user"}
    ]
  }
}
```

### GET `/users/:id`
Retrieves a user by ID (mock implementation).

**Example:** `GET /users/123`

**Response:**
```json
{
  "status": "success",
  "message": "User found",
  "data": {
    "id": "123",
    "name": "John Doe",
    "email": "john.doe@example.com"
  }
}
```

### POST `/users`
Creates a new user with validation.

**Request Body:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "securepassword123"
}
```

**Response (Success):**
```json
{
  "status": "success",
  "message": "User created successfully",
  "data": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "********"
  }
}
```

**Response (Validation Error):**
```json
{
  "status": "error",
  "error": "Validation failed",
  "errors": [
    {
      "field": "name",
      "message": "Field 'name' is required"
    }
  ]
}
```

## 🏗️ Architecture Overview

### Project Structure
```
examples/basic/
├── main.go       # Main application code
├── main_test.go  # Comprehensive test suite
├── go.mod        # Go module definition
└── README.md     # This file
```

### Key Components

#### 1. **Router Setup** (`setupRoutes()`)
- Creates a new GRA router
- Configures middleware (Logger, Recovery, CORS)
- Defines all API routes
- Returns configured router instance

#### 2. **User Model**
```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
```

#### 3. **Middleware Stack**
- **Logger**: Logs all HTTP requests
- **Recovery**: Handles panics gracefully
- **CORS**: Enables cross-origin requests

#### 4. **Validation**
- Uses GRA's built-in validator
- Supports struct tag validation
- Provides detailed error messages

## 🧪 Testing

The example includes comprehensive tests covering:

### Test Coverage: **67.7%**

- ✅ **Route Testing**: All endpoints tested
- ✅ **Validation Testing**: Valid/invalid user data
- ✅ **Error Handling**: JSON parsing errors, validation failures
- ✅ **Middleware Testing**: CORS headers, logging
- ✅ **JSON Serialization**: Marshal/unmarshal operations
- ✅ **Function Testing**: Individual function unit tests

### Running Tests

```bash
# Run all tests
go test -v

# Run tests with coverage
go test -cover

# Generate detailed coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Structure

```go
// Example test case
func TestCreateUserRoute(t *testing.T) {
    testCases := []struct {
        name           string
        requestBody    interface{}
        expectedStatus int
        expectSuccess  bool
        expectErrors   bool
    }{
        {
            name: "valid_user",
            requestBody: User{
                Name:     "John Doe",
                Email:    "john@example.com",
                Password: "password123",
            },
            expectedStatus: http.StatusCreated,
            expectSuccess:  true,
            expectErrors:   false,
        },
        // ... more test cases
    }
    // ... test implementation
}
```

## 📖 Learning Objectives

This example teaches:

1. **Basic GRA Setup**: How to create and configure a GRA application
2. **Routing**: Defining different HTTP methods and routes
3. **Middleware**: Adding cross-cutting concerns
4. **JSON Handling**: Binding JSON requests and returning JSON responses
5. **Validation**: Using struct tags for input validation
6. **Error Handling**: Returning appropriate error responses
7. **Testing**: Writing comprehensive tests for web applications

## 🔧 Customization

### Adding New Routes
```go
// In setupRoutes() function
r.GET("/health", func(c *gra.Context) {
    c.Success(http.StatusOK, "Server is healthy", map[string]any{
        "uptime": time.Since(startTime),
        "status": "ok",
    })
})
```

### Adding Middleware
```go
// Custom middleware example
func customMiddleware() gra.Middleware {
    return func(c *gra.Context) {
        // Pre-processing
        c.Header("X-Custom-Header", "MyValue")
        
        // Continue to next middleware/handler
        c.Next()
        
        // Post-processing (after handler execution)
    }
}

// Add to middleware stack
r.Use(customMiddleware())
```

### Extending Validation
```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=128"`
    Age      int    `json:"age" validate:"min=18,max=120"`
}
```

## 🚀 Next Steps

After mastering this basic example, explore:

1. **Database Integration**: Connect to PostgreSQL/MySQL
2. **Authentication**: JWT tokens and user sessions
3. **Advanced Routing**: Route groups and parameters
4. **File Uploads**: Handling multipart form data
5. **WebSocket Support**: Real-time communication
6. **Deployment**: Docker containers and cloud deployment

## 📋 Dependencies

- **GRA Framework**: Core web framework
- **GRA Router**: HTTP routing
- **GRA Middleware**: Common middleware functions
- **GRA Validator**: Input validation
- **GRA Context**: Request/response handling

## 🐛 Common Issues

### Port Already in Use
```bash
# Kill process using port 8080
lsof -ti:8080 | xargs kill -9
```

### Import Path Issues
Make sure you're in the correct directory and the module path is properly set in `go.mod`.

### Validation Not Working
Ensure struct tags are properly formatted:
- Use `validate:"required"` not `validation:"required"`
- Separate multiple rules with commas: `validate:"required,email"`

## 💡 Tips

1. **Use Table-Driven Tests**: Makes adding new test cases easy
2. **Test Edge Cases**: Empty inputs, invalid JSON, etc.
3. **Mock External Dependencies**: Keep tests fast and reliable
4. **Use Meaningful Variable Names**: Makes code self-documenting
5. **Handle Errors Gracefully**: Provide helpful error messages

This basic example provides a solid foundation for building more complex web applications with the GRA framework!

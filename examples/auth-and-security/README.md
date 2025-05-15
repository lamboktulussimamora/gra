# Authentication and Security Example

This example demonstrates how to implement JWT authentication and secure HTTP headers in a GRA Framework application.

## Features

1. **JWT Authentication**
   - Login endpoint that generates JWT tokens
   - Protected API routes using JWT middleware
   - Role-based access control (user & admin roles)

2. **Security Headers**
   - Implementation of secure HTTP headers
   - XSS protection
   - Content type options
   - Frame options
   - HSTS (HTTP Strict Transport Security)
   - And more...

## Running the Example

```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### Public Endpoints

- `GET /` - Home page (public access)
- `POST /login` - Login endpoint, returns JWT token

### Protected Endpoints

- `GET /api/profile` - Get user profile (requires authentication)
- `GET /api/admin/dashboard` - Admin dashboard (requires admin role)

## Usage Example

### 1. Login

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "password1"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "1",
    "username": "user1",
    "role": "user"
  }
}
```

### 2. Access Profile (Protected)

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Response:
```json
{
  "id": "1",
  "username": "user1",
  "role": "user"
}
```

### 3. Access Admin Dashboard (Admin Only)

```bash
# First login as admin
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Then use the returned token
curl -X GET http://localhost:8080/api/admin/dashboard \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Security Headers

This example demonstrates the use of the following security headers:

- `X-XSS-Protection`: Prevents reflected XSS attacks
- `X-Content-Type-Options`: Prevents MIME-type sniffing
- `X-Frame-Options`: Controls whether the page can be displayed in a frame
- `Strict-Transport-Security`: Forces HTTPS usage
- `Content-Security-Policy`: Restricts resource loading
- `Referrer-Policy`: Controls referrer information
- `Cross-Origin-Resource-Policy`: Controls resource sharing

You can verify the headers using:

```bash
curl -I http://localhost:8080
```

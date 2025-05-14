# API Versioning and Caching Example

This example demonstrates how to use GRA framework's API versioning and response caching features.

## Features

- API versioning with URL path strategy (e.g., /v1/products, /v2/products)
- Response caching for improved performance
- Different response schemas based on API version
- Secure headers middleware

## Running the Example

```bash
go run main.go
```

## Testing the API

### API Version 1

```bash
# Get all products (v1)
curl http://localhost:8080/api/v1/products

# Get a specific product (v1)
curl http://localhost:8080/api/v1/products/1
```

### API Version 2

```bash
# Get all products (v2)
curl http://localhost:8080/api/v2/products

# Get a specific product (v2)
curl http://localhost:8080/api/v2/products/1
```

### Default Version

The default version (v1) will be used if no version is specified:

```bash
# Uses default version (v1)
curl http://localhost:8080/api/products
```

### Testing Caching

The responses are cached for 30 seconds. Make the same request multiple times and observe the `X-Cache` header:

```bash
# First call should show "X-Cache: MISS"
curl -v http://localhost:8080/api/v1/products

# Second call (within 30 seconds) should show "X-Cache: HIT"
curl -v http://localhost:8080/api/v1/products
```

### Health Check

A simple health check endpoint is available (not versioned):

```bash
curl http://localhost:8080/api/health
```

# GRA Framework v1.0.4 Release Notes

**Release Date:** May 15, 2025

## Overview

GRA Framework v1.0.4 introduces significant enhancements to security and API management capabilities. This release adds JWT authentication, secure headers middleware, API versioning, and HTTP response caching.

## New Features

### JWT Authentication

- Comprehensive JWT implementation for authentication and authorization
- Token generation with customizable claims
- Token validation and verification
- Support for token refresh
- Middleware for protecting routes
- Role-based access control

### Secure Headers Middleware

- Default security headers for protection against common web vulnerabilities
- X-XSS-Protection header to mitigate cross-site scripting
- X-Content-Type-Options to prevent MIME type sniffing
- X-Frame-Options to control iframe embedding
- Strict-Transport-Security (HSTS) for enforcing HTTPS
- Content-Security-Policy (CSP) with flexible builder pattern
- Referrer-Policy control
- Cross-origin policy headers

### API Versioning

- Multiple versioning strategies:
  - URL path-based versioning
  - Query parameter versioning
  - Header-based versioning
  - Media type (Accept header) versioning
- Middleware for automatic version handling
- Customizable version extraction

### Response Caching

- HTTP response caching with ETag support
- Cache invalidation capabilities
- Configurable cache stores
- Cache middleware for easy integration

## Improvements

- Enhanced context management with better value handling
- Improved router performance
- More comprehensive validation options
- Better error handling and reporting

## Examples

- Added auth-and-security example demonstrating JWT and secure headers
- Added versioning-and-cache example

## Documentation

- Updated README with examples for all new features
- Added detailed security documentation
- Improved API documentation

## Requirements

- Go 1.24+

## Migration

No breaking changes in this release. All new features are additive and backward compatible.
# GRA Framework v1.0.5 Release Notes

**Release Date:** May 15, 2025

## Overview

GRA Framework v1.0.5 is a maintenance release that fixes issues in the example applications and improves documentation. This release follows v1.0.4 which introduced JWT authentication, secure headers middleware, API versioning, and HTTP response caching.

## Bug Fixes

### Example Applications
- Fixed routing issues in the auth-and-security example
- Resolved compatibility issues in examples with local module usage
- Fixed incorrect header routing in versioning-and-cache example

### Validator Package
- Fixed regex pattern validation for usernames and phone numbers
- Fixed commented-out test assertions in validator_test.go

## Improvements

### Validator Package
- Refactored validator.go to reduce cognitive complexity
- Improved regex pattern handling with proper anchoring
- Modularized validation functions for better maintainability
- Enhanced test coverage for regex pattern validation

### Documentation
- Enhanced example applications with better documentation
- Updated README files for all examples
- Clarified usage instructions for authentication and API versioning
- Updated ENHANCEMENT_SUMMARY.md with validator improvements

### Project Organization
- Updated .gitignore to exclude example binaries
- Improved test script organization
- Removed unused script files

## Requirements

- Go 1.16+

## Migration

No breaking changes in this release. All changes are backward compatible with v1.0.4.

To use this framework in your projects:
```bash
go get github.com/lamboktulussimamora/gra@v1.0.5
```

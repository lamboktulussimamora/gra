# Changelog

## v1.0.5 (2025-05-15)

### Bug Fixes and Improvements

#### Bug Fixes:
- Fixed routing issues in auth-and-security example
- Resolved compatibility issues in examples with local module usage
- Fixed incorrect header routing in versioning-and-cache example

#### Improvements:
- Enhanced example applications with better documentation
- Updated .gitignore to exclude example binaries
- Improved test script organization

## v1.0.4 (2025-05-15)

### Feature Enhancements

#### New Features:
- Added API versioning package with multiple strategies (path, query, header, media type)
- Implemented HTTP response caching middleware with ETag support
- Implemented JWT authentication package with token generation and validation
- Added secure headers middleware for improved security
- Added versioning and caching example application
- Added authentication and security example application

#### Documentation:
- Updated README.md with comprehensive documentation for new features
- Created ENHANCEMENT_SUMMARY.md with details of implementations
- Added example code for all new features

## v1.0.3 (2025-05-11)

### Bug Fixes and Test Coverage Improvements

#### Bug Fixes:
- Fixed a bug in validator's max validation for unsigned integers
- Improved error handling in JSON encoding/decoding

#### Test Coverage Improvements:
- Enhanced test coverage for validator package from 84.6% to 98.9%
- Added tests for various data types in validator
- Added tests for edge cases in all validation rules
- Improved overall test coverage from 92.6% to 98.0%

## v1.0.2 (2025-05-11)

### Framework Improvements and Testing Enhancements

#### Development Improvements:
- Added Makefile for standardized development workflows
- Created comprehensive TEST_PLAN.md for maintaining test quality
- Added benchmark tests for router performance analysis
- Fixed linting issues in test files
- Improved documentation for testing procedures

#### Test Coverage Improvements:
- Added tests for Fatal logger methods with proper mocking
- Enhanced router tests with complex path parameter combinations
- Added validator tests for arrays of nested structs
- Achieved 100% test coverage in context package with edge case testing
- Added tests for the JSONData function to ensure proper handling of raw JSON responses
- Fixed bug in validator's max validation for unsigned integers
- Expanded test coverage for validator with edge cases and multiple data types
- Improved overall test coverage from 90.2% to 98.0%

## v1.0.1 (2025-05-11)

### Test Coverage Improvements

Added comprehensive test coverage across all packages.

#### Improvements:
- Added tests for all packages: adapter, context, logger, middleware, router, validator, and core
- Achieved overall test coverage of 90.2%
- Fixed test functions to ensure proper error handling and edge cases
- Added GitHub Actions workflows for automated test coverage reporting
- Integrated with Coveralls and CodeCov for coverage visualization
- Set up GitHub Pages for publishing coverage reports

## v1.0.1 (2025-05-10)

- Renamed the package from `core` to `gra`
- Renamed the file from `core.go` to `gra.go`
- Updated all import paths in the examples
- Updated the README to reflect the new package name
- Updated examples to use the new package name
- Updated Quick Start Guide to reflect the new package name

## v1.0.0 (2025-05-10)

### Framework Renamed from go-core-framework to gra

This release marks the official renaming of the framework from "go-core-framework" to "gra".

#### Changes:
- Renamed the GitHub repository
- Updated the module name in go.mod
- Updated all import paths throughout the codebase
- Updated examples to use the new import paths
- Updated documentation to reflect the new name
- Added migration guide for existing users

#### Migration:
- See MIGRATION.md for instructions on how to update your code

#### Functionality:
- No functional changes or breaking changes were introduced with this rename
- All APIs remain the same, only the import paths have changed

## v1.0.0 (2025-05-10)

- Initial framework implementation
- Core context handling
- Router implementation
- Middleware support
- Validation utilities
- Basic examples
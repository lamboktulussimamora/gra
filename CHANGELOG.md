# Changelog

## v1.0.7 (2025-06-05)

### Hybrid Migration System Bug Fix

#### Bug Fixes:
- **Fixed HybridMigrator initialization issue**: Fixed runtime error where `migrator.EnsureSchema()` was called directly but `HybridMigrator` doesn't expose this method
- **Enhanced GetMigrationStatus() method**: Added automatic schema initialization for both EF migration schema and migration history table
- **Improved demo reliability**: Updated hybrid migration demo to use proper API patterns and automatic initialization

#### New Features:
- **Enhanced Hybrid Migration System**: Complete hybrid migration system with model-driven migration generation
- **Model Registry**: Added ModelRegistry for registering Go structs as database models
- **Change Detection**: Implemented automatic change detection between Go models and database schema
- **SQL Generation**: Added automated SQL generation for schema changes (CREATE TABLE, ALTER TABLE, etc.)
- **Migration File Generation**: Automated migration file creation with proper up/down scripts
- **Database Inspector**: Added database schema inspection capabilities for multiple database engines

#### Examples:
- **Hybrid Migration Demo**: Complete working demo showing model registration, change detection, and migration generation
- **Integration Tests**: Added comprehensive test suite for hybrid migration functionality

#### Documentation:
- Updated hybrid migration documentation with proper usage patterns
- Added troubleshooting guide for common migration issues
- Enhanced examples with proper error handling

## v1.0.6 (2025-05-17)

### Go Version Update

#### Changes:
- Updated minimum Go version requirement from 1.21 to 1.24
- Updated all go.mod files in main project and example applications
- Updated documentation to reflect new Go version requirements
- Ensured compatibility with Go 1.24 across all components

## v1.0.5 (2025-05-15)

### Bug Fixes and Improvements

#### Bug Fixes:
- Fixed routing issues in auth-and-security example
- Resolved compatibility issues in examples with local module usage
- Fixed incorrect header routing in versioning-and-cache example
- Fixed regex pattern validation in validator package for usernames and phone numbers

#### Improvements:
- Enhanced example applications with better documentation
- Updated .gitignore to exclude example binaries
- Improved test script organization
- Refactored validator.go to reduce cognitive complexity
- Improved regex pattern handling with proper anchoring
- Modularized validation functions for better maintainability

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
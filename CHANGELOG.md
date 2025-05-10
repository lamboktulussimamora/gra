# Changelog

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
- Improved overall test coverage to 93.5%

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
#!/bin/bash

# Script to commit all the framework improvements

echo "Staging changes..."

# Stage all modified files
git add gra.go
git add CHANGELOG.md
git add TEST_COVERAGE.md
git add context/context.go
git add context/context_test.go
git add context/mock_writer_test.go
git add validator/validator.go
git add validator/validator_test.go
git add scripts/fix_context_constants.sh
git add scripts/fix_context_tests.sh

# Commit with a descriptive message
git commit -m "Version 1.0.3: Test improvements and validator bug fix

- Updated version to 1.0.3
- Fixed validator bug for max validation of unsigned integers
- Improved test coverage from 90.2% to 98.0%
- Added JSONData function with comprehensive tests
- Added edge case testing for context and validator packages
- Created mock objects for error testing
- Enhanced validator testing with all data types
- Added tests for embedded structs and JSON tag handling
- Updated TEST_COVERAGE.md with new metrics"

echo "Changes committed successfully!"
echo "You can now push these changes to GitHub with: git push origin main"

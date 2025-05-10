#!/bin/bash

# Generate test coverage report
echo "Generating test coverage..."
go test ./... -coverprofile=coverage.out

# Convert to HTML
echo "Creating HTML report..."
go tool cover -html=coverage.out -o coverage.html
echo "HTML report created: coverage.html"

# Show coverage percentage
go tool cover -func=coverage.out

echo ""
echo "Coverage report generation complete"

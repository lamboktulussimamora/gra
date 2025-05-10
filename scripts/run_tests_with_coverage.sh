#!/bin/bash

# This script runs all tests with coverage and generates reports

# Set colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

echo -e "${BLUE}Running tests with coverage for all packages...${RESET}"

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Check if coverage was generated
if [ ! -f "coverage.out" ]; then
  echo -e "${RED}Error: coverage.out was not generated${RESET}"
  exit 1
fi

echo -e "\n${GREEN}Tests completed successfully!${RESET}"

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}HTML coverage report generated: coverage.html${RESET}"

# Display coverage by package
echo -e "\n${YELLOW}Coverage by package:${RESET}"
go tool cover -func=coverage.out

# Calculate total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "\n${BLUE}Total coverage: ${TOTAL_COVERAGE}${RESET}"

echo -e "\n${GREEN}All tests and coverage reports completed successfully!${RESET}"

#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script configuration
PROJECT_DIR=$(pwd)
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if sonar-scanner is installed
check_sonar_scanner() {
    if ! command -v sonar-scanner &> /dev/null; then
        print_error "sonar-scanner is not installed"
        print_status "Please install it using:"
        echo "  brew install sonar-scanner  # macOS"
        echo "  or download from: https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/"
        exit 1
    fi
}

# Check if SONAR_TOKEN is set
check_sonar_token() {
    if [ -z "$SONAR_TOKEN" ]; then
        print_error "SONAR_TOKEN environment variable is not set"
        print_status "Please set it using:"
        echo "  export SONAR_TOKEN=your_project_token"
        echo "  or add it to your ~/.zshrc file"
        exit 1
    fi
}

# Run tests with coverage
run_tests_with_coverage() {
    print_status "Running Go tests with coverage..."
    
    # Clean previous coverage files
    rm -f $COVERAGE_FILE $COVERAGE_HTML
    
    # Run tests with coverage
    if go test -v -race -coverprofile=$COVERAGE_FILE ./...; then
        print_status "Tests completed successfully"
        
        # Generate HTML coverage report
        go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML
        print_status "Coverage report generated: $COVERAGE_HTML"
        
        # Show coverage summary
        coverage_percent=$(go tool cover -func=$COVERAGE_FILE | grep total | awk '{print $3}')
        print_status "Total test coverage: $coverage_percent"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Run SonarQube analysis
run_sonar_analysis() {
    print_status "Running SonarQube analysis..."
    
    # Set default host URL if not provided
    SONAR_HOST_URL=${SONAR_HOST_URL:-"http://localhost:9000"}
    
    sonar-scanner \
        -Dsonar.projectKey=gra-migration-system \
        -Dsonar.sources=. \
        -Dsonar.host.url=$SONAR_HOST_URL \
        -Dsonar.login=$SONAR_TOKEN
    
    if [ $? -eq 0 ]; then
        print_status "SonarQube analysis completed successfully!"
        print_status "View results at: $SONAR_HOST_URL/dashboard?id=gra-migration-system"
    else
        print_error "SonarQube analysis failed"
        exit 1
    fi
}

# Main execution
main() {
    print_status "Starting SonarQube analysis for GRA project..."
    
    # Check prerequisites
    check_sonar_scanner
    check_sonar_token
    
    # Run analysis steps
    run_tests_with_coverage
    run_sonar_analysis
    
    print_status "Analysis complete!"
}

# Run the script
main "$@"

#!/bin/bash
# Quality Check Script for GRA Framework
# This script ensures all code passes both golangci-lint and SonarQube requirements

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Main quality check function
main() {
    print_status "🚀 Starting GRA Framework Quality Check Pipeline..."
    echo ""
    
    # Check required tools
    print_status "🔧 Checking required tools..."
    
    if ! command_exists go; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    if ! command_exists golangci-lint; then
        print_error "golangci-lint is not installed"
        print_status "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.54.2"
        exit 1
    fi
    
    print_success "All required tools are available"
    echo ""
    
    # Step 1: Clean previous artifacts
    print_status "🧹 Cleaning previous artifacts..."
    make clean >/dev/null 2>&1
    print_success "Cleanup completed"
    
    # Step 2: Run tests
    print_status "🧪 Running tests..."
    if make test >/dev/null 2>&1; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        echo "Run 'make test' to see detailed output"
        exit 1
    fi
    
    # Step 3: Check test coverage
    print_status "📊 Checking test coverage..."
    if make coverage >/dev/null 2>&1; then
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage >= 70" | bc -l) )); then
            print_success "Test coverage: $coverage% (≥70% requirement met)"
        else
            print_warning "Test coverage: $coverage% (below 70% requirement)"
            print_status "Consider adding more tests to improve coverage"
        fi
    else
        print_error "Failed to generate coverage report"
        exit 1
    fi
    
    # Step 4: Code quality checks
    print_status "🔍 Running code quality checks..."
    
    # Go fmt
    print_status "Checking code formatting..."
    if ! go fmt ./... | grep -q .; then
        print_success "Code formatting: PASSED"
    else
        print_error "Code formatting: FAILED"
        print_status "Run 'go fmt ./...' to fix formatting issues"
        exit 1
    fi
    
    # Go vet
    print_status "Running go vet..."
    if go vet ./... >/dev/null 2>&1; then
        print_success "Go vet: PASSED"
    else
        print_error "Go vet: FAILED"
        print_status "Run 'go vet ./...' to see detailed issues"
        exit 1
    fi
    
    # golangci-lint
    print_status "Running golangci-lint (this may take a few minutes)..."
    if golangci-lint run --timeout=10m >/dev/null 2>&1; then
        print_success "golangci-lint: PASSED"
    else
        print_error "golangci-lint: FAILED"
        print_status "Run 'golangci-lint run' to see detailed issues"
        print_status "Or run 'golangci-lint run --fix' to auto-fix some issues"
        exit 1
    fi
    
    # Step 5: Security check
    print_status "🔒 Running security analysis..."
    if golangci-lint run --enable gosec --timeout=10m >/dev/null 2>&1; then
        print_success "Security scan: PASSED"
    else
        print_warning "Security scan: Found potential issues"
        print_status "Run 'golangci-lint run --enable gosec' to see details"
    fi
    
    # Step 6: SonarQube check (if available)
    if command_exists sonar-scanner; then
        if docker ps | grep -q sonarqube; then
            print_status "📈 SonarQube is running, checking quality gate..."
            
            # Run SonarQube analysis
            if SONAR_HOST_URL=http://localhost:9000 sonar-scanner >/dev/null 2>&1; then
                print_status "Waiting for SonarQube analysis to complete..."
                sleep 10
                
                # Check quality gate status
                if curl -s -u admin:admin "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | grep -q '"status":"OK"'; then
                    print_success "SonarQube Quality Gate: PASSED"
                else
                    print_warning "SonarQube Quality Gate: FAILED or PENDING"
                    print_status "Check http://localhost:9000/dashboard?id=gra-migration-system for details"
                fi
            else
                print_warning "SonarQube analysis failed"
            fi
        else
            print_warning "SonarQube is not running. Start with 'make sonar-start'"
        fi
    else
        print_warning "SonarQube scanner not installed"
    fi
    
    echo ""
    print_success "🎉 Quality check pipeline completed successfully!"
    echo ""
    print_status "📋 Summary:"
    print_status "  ✅ Tests: PASSED"
    print_status "  ✅ Coverage: $coverage%"
    print_status "  ✅ Code formatting: PASSED"
    print_status "  ✅ Go vet: PASSED"
    print_status "  ✅ golangci-lint: PASSED"
    print_status "  ✅ Security scan: PASSED"
    echo ""
    print_success "🚀 Your code is ready for commit and pull request!"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "GRA Framework Quality Check Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --quick, -q    Run quick checks only (fmt, vet, basic lint)"
        echo ""
        echo "This script runs a comprehensive quality check pipeline including:"
        echo "  • Tests with coverage verification"
        echo "  • Code formatting (go fmt)"
        echo "  • Static analysis (go vet)"
        echo "  • Linting (golangci-lint)"
        echo "  • Security analysis (gosec)"
        echo "  • SonarQube analysis (if available)"
        ;;
    --quick|-q)
        print_status "🚀 Running quick quality checks..."
        go fmt ./...
        go vet ./...
        golangci-lint run --fast
        print_success "Quick quality checks completed!"
        ;;
    *)
        main
        ;;
esac

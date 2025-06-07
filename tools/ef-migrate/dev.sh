#!/bin/bash

# EF Migrate Development Helper Script
# This script provides convenient commands for developing and testing the EF migration tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
BINARY_NAME="ef-migrate"
BINARY_PATH="./bin/${BINARY_NAME}"
TEST_DB_PATH="./test.db"
EXAMPLES_DIR="./examples"
MIGRATIONS_DIR="./migrations"

# Function to print colored output
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to show help
show_help() {
    cat << EOF
🚀 EF Migrate Development Helper

Usage: $0 <command> [options]

Commands:
  build           Build the ef-migrate binary
  test            Run tests with coverage
  test-watch      Run tests in watch mode
  clean           Clean build artifacts and test files
  dev-setup       Setup development environment
  lint            Run linting and formatting
  
Database Commands:
  db-start        Start PostgreSQL with Docker
  db-stop         Stop PostgreSQL Docker container
  db-reset        Reset test database
  
Migration Commands:
  demo            Run a complete migration demo
  example-init    Initialize with example migrations
  example-test    Test with example migrations
  
Docker Commands:
  docker-build    Build Docker image
  docker-run      Run in Docker container
  docker-compose  Start full development environment
  
Utility Commands:
  install         Install binary to /usr/local/bin
  release         Create release build
  help            Show this help message

Examples:
  $0 build && $0 demo
  $0 test-watch
  $0 docker-compose && $0 example-test

Environment Variables:
  DATABASE_URL    - Database connection string
  VERBOSE         - Enable verbose output (true/false)
  
EOF
}

# Function to build the binary
build() {
    print_status "Building ${BINARY_NAME}..."
    mkdir -p bin
    
    VERSION=${VERSION:-"dev"}
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    go build \
        -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o "${BINARY_PATH}" \
        .
    
    print_success "Built ${BINARY_NAME} -> ${BINARY_PATH}"
}

# Function to run tests
run_tests() {
    print_status "Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out ./...
    
    if [ -f coverage.out ]; then
        go tool cover -html=coverage.out -o coverage.html
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        print_success "Test coverage: ${coverage}"
        print_status "Coverage report: coverage.html"
    fi
}

# Function to watch tests
test_watch() {
    print_status "Running tests in watch mode..."
    print_status "Press Ctrl+C to stop"
    
    # Install fswatch if not available
    if ! command -v fswatch &> /dev/null; then
        print_warning "fswatch not found. Install with: brew install fswatch"
        return 1
    fi
    
    fswatch -o . -e ".*" -i "\\.go$" | while read f; do
        clear
        echo -e "${BLUE}🔄 File changed, running tests...${NC}"
        go test -v ./... 2>&1 || true
        echo -e "${BLUE}⏳ Watching for changes...${NC}"
    done
}

# Function to clean up
clean() {
    print_status "Cleaning up..."
    rm -rf bin/
    rm -f coverage.out coverage.html
    rm -f "${TEST_DB_PATH}"
    rm -rf test_migrations/
    print_success "Cleaned up build artifacts and test files"
}

# Function to setup development environment
dev_setup() {
    print_status "Setting up development environment..."
    
    # Create directories
    mkdir -p bin migrations examples
    
    # Check dependencies
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        return 1
    fi
    
    # Install dependencies
    go mod download
    go mod verify
    
    # Build the tool
    build
    
    print_success "Development environment ready!"
    print_status "Try: $0 demo"
}

# Function to run linting
lint() {
    print_status "Running linting and formatting..."
    
    # Format code
    go fmt ./...
    
    # Run vet
    go vet ./...
    
    # Run golangci-lint if available
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --timeout=5m
        print_success "Linting completed"
    else
        print_warning "golangci-lint not found. Install with: brew install golangci/tap/golangci-lint"
    fi
}

# Function to start PostgreSQL
db_start() {
    print_status "Starting PostgreSQL with Docker..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        return 1
    fi
    
    docker run -d \
        --name gra-postgres-dev \
        -p 5432:5432 \
        -e POSTGRES_USER=postgres \
        -e POSTGRES_PASSWORD=postgres \
        -e POSTGRES_DB=gra_dev \
        postgres:15-alpine
    
    print_success "PostgreSQL started on port 5432"
    print_status "Connection: postgres://postgres:postgres@localhost:5432/gra_dev"
}

# Function to stop PostgreSQL
db_stop() {
    print_status "Stopping PostgreSQL Docker container..."
    docker stop gra-postgres-dev 2>/dev/null || true
    docker rm gra-postgres-dev 2>/dev/null || true
    print_success "PostgreSQL stopped"
}

# Function to reset database
db_reset() {
    print_status "Resetting test database..."
    rm -f "${TEST_DB_PATH}"
    print_success "Test database reset"
}

# Function to run a complete demo
demo() {
    print_status "Running EF Migration Tool Demo..."
    
    if [ ! -f "${BINARY_PATH}" ]; then
        build
    fi
    
    # Clean up
    rm -f "${TEST_DB_PATH}"
    
    print_status "1. Showing version..."
    "${BINARY_PATH}" version
    
    print_status "2. Checking initial status..."
    "${BINARY_PATH}" -connection "${TEST_DB_PATH}" status
    
    print_status "3. Creating a migration..."
    "${BINARY_PATH}" -connection "${TEST_DB_PATH}" add-migration "DemoMigration" "Demo migration for testing"
    
    print_status "4. Listing migrations..."
    "${BINARY_PATH}" -connection "${TEST_DB_PATH}" list
    
    print_status "5. Updating database..."
    "${BINARY_PATH}" -connection "${TEST_DB_PATH}" update-database
    
    print_status "6. Checking final status..."
    "${BINARY_PATH}" -connection "${TEST_DB_PATH}" status
    
    print_success "Demo completed successfully!"
}

# Function to initialize with examples
example_init() {
    print_status "Initializing with example migrations..."
    
    if [ ! -f "${BINARY_PATH}" ]; then
        build
    fi
    
    # Copy examples to migrations directory
    mkdir -p "${MIGRATIONS_DIR}"
    cp -r "${EXAMPLES_DIR}"/* "${MIGRATIONS_DIR}/" 2>/dev/null || true
    
    print_success "Example migrations copied to ${MIGRATIONS_DIR}/"
}

# Function to test with examples
example_test() {
    print_status "Testing with example migrations..."
    
    if [ ! -f "${BINARY_PATH}" ]; then
        build
    fi
    
    # Use SQLite for testing
    local test_db="./example_test.db"
    rm -f "${test_db}"
    
    print_status "Testing with database: ${test_db}"
    
    # Test each command
    "${BINARY_PATH}" -connection "${test_db}" status
    "${BINARY_PATH}" -connection "${test_db}" add-migration "ExampleTest" "Testing with examples"
    "${BINARY_PATH}" -connection "${test_db}" update-database
    "${BINARY_PATH}" -connection "${test_db}" list
    "${BINARY_PATH}" -connection "${test_db}" status
    
    print_success "Example testing completed!"
    print_status "Test database: ${test_db}"
}

# Function to build Docker image
docker_build() {
    print_status "Building Docker image..."
    
    VERSION=${VERSION:-"dev"}
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg GIT_COMMIT="${GIT_COMMIT}" \
        -t gra-ef-migrate:latest \
        -f Dockerfile \
        ../../
    
    print_success "Docker image built: gra-ef-migrate:latest"
}

# Function to run in Docker
docker_run() {
    print_status "Running in Docker container..."
    
    docker run --rm -it \
        -v "$(pwd)/migrations:/app/migrations" \
        gra-ef-migrate:latest \
        "$@"
}

# Function to start Docker Compose
docker_compose() {
    print_status "Starting development environment with Docker Compose..."
    
    if [ ! -f docker-compose.yml ]; then
        print_error "docker-compose.yml not found"
        return 1
    fi
    
    docker-compose up -d
    print_success "Development environment started"
    print_status "PostgreSQL: localhost:5432"
    print_status "pgAdmin: http://localhost:8080"
}

# Function to install binary
install() {
    if [ ! -f "${BINARY_PATH}" ]; then
        build
    fi
    
    print_status "Installing ${BINARY_NAME} to /usr/local/bin..."
    sudo cp "${BINARY_PATH}" /usr/local/bin/
    print_success "Installed ${BINARY_NAME} to /usr/local/bin/"
}

# Function to create release
release() {
    print_status "Creating release build..."
    
    VERSION=${1:-"v1.0.0"}
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    mkdir -p release
    
    # Build for multiple platforms
    for GOOS in linux darwin windows; do
        for GOARCH in amd64 arm64; do
            if [ "$GOOS" = "windows" ] && [ "$GOARCH" = "arm64" ]; then
                continue
            fi
            
            BINARY_NAME_RELEASE="${BINARY_NAME}-${GOOS}-${GOARCH}"
            if [ "$GOOS" = "windows" ]; then
                BINARY_NAME_RELEASE="${BINARY_NAME_RELEASE}.exe"
            fi
            
            print_status "Building for ${GOOS}/${GOARCH}..."
            
            env GOOS=$GOOS GOARCH=$GOARCH go build \
                -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
                -o "release/${BINARY_NAME_RELEASE}" \
                .
        done
    done
    
    print_success "Release builds created in release/"
    ls -la release/
}

# Main command router
case "${1:-help}" in
    build)
        build
        ;;
    test)
        run_tests
        ;;
    test-watch)
        test_watch
        ;;
    clean)
        clean
        ;;
    dev-setup)
        dev_setup
        ;;
    lint)
        lint
        ;;
    db-start)
        db_start
        ;;
    db-stop)
        db_stop
        ;;
    db-reset)
        db_reset
        ;;
    demo)
        demo
        ;;
    example-init)
        example_init
        ;;
    example-test)
        example_test
        ;;
    docker-build)
        docker_build
        ;;
    docker-run)
        shift
        docker_run "$@"
        ;;
    docker-compose)
        docker_compose
        ;;
    install)
        install
        ;;
    release)
        release "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac

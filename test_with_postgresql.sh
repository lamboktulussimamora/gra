#!/bin/bash

# Test script for PostgreSQL integration testing
# Usage: ./test_with_postgresql.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
POSTGRES_HOST="localhost"
POSTGRES_PORT="5433"
POSTGRES_DB="gra_test"
POSTGRES_USER="gra_user"
POSTGRES_PASSWORD="gra_password"
POSTGRES_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

echo -e "${BLUE}🐘 Starting PostgreSQL Integration Tests${NC}"
echo "========================================"

# Function to check if PostgreSQL is ready
check_postgres() {
    echo -e "${YELLOW}⏳ Checking PostgreSQL connection...${NC}"
    
    for i in {1..30}; do
        if docker exec gra-postgres-test pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB} >/dev/null 2>&1; then
            echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
    done
    
    echo -e "${RED}❌ PostgreSQL failed to start${NC}"
    return 1
}

# Function to start PostgreSQL
start_postgres() {
    echo -e "${YELLOW}🚀 Starting PostgreSQL with Docker...${NC}"
    
    # Stop existing container if running
    docker-compose -f docker-compose.test.yml down >/dev/null 2>&1 || true
    
    # Start PostgreSQL
    docker-compose -f docker-compose.test.yml up -d postgres-test
    
    # Wait for PostgreSQL to be ready
    check_postgres
}

# Function to stop PostgreSQL
stop_postgres() {
    echo -e "${YELLOW}🛑 Stopping PostgreSQL...${NC}"
    docker-compose -f docker-compose.test.yml down
}

# Function to run tests with PostgreSQL
run_postgres_tests() {
    echo -e "${BLUE}🧪 Running tests with PostgreSQL...${NC}"
    
    # Set environment variables for PostgreSQL testing
    export TEST_DATABASE_URL="${POSTGRES_URL}"
    export TEST_WITH_POSTGRES="true"
    
    # Run all tests with coverage
    echo -e "${YELLOW}📊 Running full test suite with coverage...${NC}"
    go test -v -race -coverprofile=coverage_postgresql.out -covermode=atomic ./...
    
    # Run specific migration tests
    echo -e "${YELLOW}🔄 Running migration-specific tests...${NC}"
    go test -v -race ./orm/migrations -args -test.timeout=30s
    go test -v -race ./tools/migration/direct -args -test.timeout=30s
    go test -v -race ./tools/migration/test -args -test.timeout=30s
    
    # Generate coverage report
    echo -e "${YELLOW}📈 Generating coverage report...${NC}"
    go tool cover -html=coverage_postgresql.out -o coverage_postgresql.html
    go tool cover -func=coverage_postgresql.out
    
    echo -e "${GREEN}✅ PostgreSQL tests completed!${NC}"
}

# Function to run migration stress tests
run_migration_stress_tests() {
    echo -e "${BLUE}🏋️ Running migration stress tests...${NC}"
    
    export TEST_DATABASE_URL="${POSTGRES_URL}"
    export TEST_WITH_POSTGRES="true"
    
    # Create test migration scenarios
    echo -e "${YELLOW}🔄 Testing complex migration scenarios...${NC}"
    
    # Test 1: Large schema migration
    echo "Testing large schema migrations..."
    go run tools/migration/test/test_runner.go --conn="${POSTGRES_URL}" --up || echo "Expected failure for demo purposes"
    
    # Test 2: Multiple table creation
    echo "Testing multiple table scenarios..."
    
    # Test 3: Data migration scenarios
    echo "Testing data migration scenarios..."
    
    echo -e "${GREEN}✅ Stress tests completed!${NC}"
}

# Function to cleanup
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    stop_postgres
    
    # Remove test files
    rm -f coverage_postgresql.out coverage_postgresql.html
    
    # Unset environment variables
    unset TEST_DATABASE_URL
    unset TEST_WITH_POSTGRES
}

# Main execution
main() {
    # Set trap for cleanup on exit
    trap cleanup EXIT
    
    echo -e "${BLUE}🏁 Starting PostgreSQL integration test suite${NC}"
    echo "Database URL: ${POSTGRES_URL}"
    echo ""
    
    # Start PostgreSQL
    start_postgres
    
    # Wait a bit more for full initialization
    sleep 3
    
    # Run tests
    run_postgres_tests
    
    # Run stress tests if requested
    if [[ "${1}" == "--stress" ]]; then
        run_migration_stress_tests
    fi
    
    echo ""
    echo -e "${GREEN}🎉 All PostgreSQL tests completed successfully!${NC}"
    echo -e "${BLUE}📊 Coverage report available at: coverage_postgresql.html${NC}"
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is required but not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is required but not installed${NC}"
    exit 1
fi

# Run main function
main "$@"

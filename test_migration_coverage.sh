#!/bin/bash

# Migration Test Coverage Script
# This script sets up PostgreSQL and runs migration tests with high coverage

set -e

echo "🚀 Starting Migration Test Coverage Script"

# Function to cleanup on exit
cleanup() {
    echo "🧹 Cleaning up..."
    docker stop test-postgres 2>/dev/null || true
    docker rm test-postgres 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is required but not installed. Please install Docker first."
    exit 1
fi

echo "📦 Setting up PostgreSQL container..."

# Remove existing container if it exists
docker stop test-postgres 2>/dev/null || true
docker rm test-postgres 2>/dev/null || true

# Start PostgreSQL container
echo "🐘 Starting PostgreSQL container..."
docker run --name test-postgres \
    -e POSTGRES_PASSWORD=testpass \
    -e POSTGRES_DB=testdb \
    -p 5433:5432 \
    -d postgres:15-alpine

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Verify PostgreSQL is running
if ! docker exec test-postgres pg_isready -U postgres; then
    echo "❌ PostgreSQL failed to start properly"
    exit 1
fi

echo "✅ PostgreSQL is ready!"

# Run migration tests with coverage
echo "🧪 Running migration tests with coverage..."
cd "$(dirname "$0")"

# Run the tests
go test -coverprofile=tools_migration_test_cover.out ./tools/migration/test

# Generate coverage report
echo "📊 Generating coverage report..."
go tool cover -func=tools_migration_test_cover.out

# Generate HTML coverage report
go tool cover -html=tools_migration_test_cover.out -o migration_test_coverage.html

echo "✅ Coverage test completed!"
echo "📄 HTML report generated: migration_test_coverage.html"

# Test the main function directly to verify it works
echo "🔍 Testing main function directly..."
cd tools/migration/test

# Use environment variable for database password
DB_PASSWORD="${DB_PASSWORD:-testpass}"
CONNECTION_STRING="postgres://postgres:${DB_PASSWORD}@localhost:5433/testdb?sslmode=disable"

go run test_runner.go --conn "${CONNECTION_STRING}" --up

echo "🎉 All tests completed successfully!"
echo "💡 Note: PostgreSQL container will be cleaned up automatically"

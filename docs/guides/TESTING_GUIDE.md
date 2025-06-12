# GRA Framework Testing Guide

Complete guide for running tests with both SQLite and PostgreSQL databases using Docker Compose.

## Prerequisites

### Required Software

1. **Go** (version 1.19 or later)
   ```bash
   go version
   ```

2. **Docker & Docker Compose**
   ```bash
   docker --version
   docker-compose --version
   ```

3. **Git** (for cloning the repository)
   ```bash
   git --version
   ```

### Database Drivers

The project already includes the necessary database drivers:
- SQLite: `github.com/mattn/go-sqlite3`
- PostgreSQL: `github.com/lib/pq`

## Testing Scenarios

### 1. SQLite Testing (Default)

SQLite testing is the default and fastest option, suitable for local development and CI/CD pipelines.

#### Quick SQLite Tests

```bash
# Run all tests with SQLite (default)
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific migration tests
go test -v ./orm/migrations

# Run with race detection
go test -race ./...
```

#### SQLite Migration Tests

```bash
# Run auto migration tests
go test -v ./orm/migrations -run TestAutoMigration

# Run migration-specific tests
go test -v ./tools/migration/test

# Test with coverage
go test -coverprofile=migration_coverage.out ./orm/migrations
go tool cover -func=migration_coverage.out
```

### 2. PostgreSQL Testing (Docker)

PostgreSQL testing provides comprehensive database integration testing using Docker containers.

#### Setting Up PostgreSQL with Docker Compose

The project includes a pre-configured Docker Compose setup for PostgreSQL testing:

```bash
# Start PostgreSQL container
docker-compose -f docker-compose.test.yml up -d postgres-test

# Verify PostgreSQL is running
docker-compose -f docker-compose.test.yml ps

# Check logs
docker-compose -f docker-compose.test.yml logs postgres-test
```

#### Configuration Details

The `docker-compose.test.yml` includes:
- **PostgreSQL 15 Alpine** (lightweight, fast startup)
- **Port 5433** (to avoid conflicts with local PostgreSQL)
- **Database**: `gra_test`
- **User**: `gra_user`
- **Password**: `gra_password`
- **Health checks** for reliable startup detection

#### Running PostgreSQL Tests

##### Option 1: Automated Script (Recommended)

```bash
# Run complete PostgreSQL test suite
./test_with_postgresql.sh

# Run with stress testing
./test_with_postgresql.sh --stress
```

This script automatically:
- Starts PostgreSQL container
- Waits for readiness
- Runs all tests with coverage
- Generates HTML coverage report
- Cleans up resources

##### Option 2: Manual Setup

```bash
# 1. Start PostgreSQL
docker-compose -f docker-compose.test.yml up -d postgres-test

# 2. Set environment variables
export TEST_DATABASE_URL="postgres://gra_user:gra_password@localhost:5433/gra_test?sslmode=disable"
export TEST_WITH_POSTGRES="true"

# 3. Run tests
go test -v ./orm/migrations

# 4. Run with coverage
go test -coverprofile=postgres_coverage.out ./orm/migrations
go tool cover -html=postgres_coverage.out -o postgres_coverage.html

# 5. Cleanup
docker-compose -f docker-compose.test.yml down
unset TEST_DATABASE_URL TEST_WITH_POSTGRES
```

#### PostgreSQL-Specific Tests

```bash
# Run PostgreSQL integration tests
export TEST_WITH_POSTGRES="true"
export TEST_DATABASE_URL="postgres://gra_user:gra_password@localhost:5433/gra_test?sslmode=disable"

# Test PostgreSQL-specific features
go test -v ./orm/migrations -run TestPostgreSQL

# Test database driver compatibility
go test -v ./orm/migrations -run TestDatabaseDriverCompatibility

# Test large dataset migration
go test -v ./orm/migrations -run TestLargeDatasetMigration
```

### 3. Multi-Database Testing

Test the same functionality across both databases:

```bash
# Run cross-database compatibility tests
go test -v ./orm/migrations -run TestMultiDatabaseCompatibility

# Test with both databases using the automated script
./test_with_postgresql.sh
```

## Test Categories

### Core Migration Tests

```bash
# Auto migration functionality
go test -v ./orm/migrations -run TestAutoMigration

# Model registration and snapshot creation
go test -v ./orm/migrations -run TestModelRegistry

# Schema comparison and change detection
go test -v ./orm/migrations -run TestSchemaComparison
```

### Database Driver Tests

```bash
# Test SQLite-specific features
go test -v ./orm/migrations -run TestSQLite

# Test PostgreSQL-specific features (requires PostgreSQL container)
export TEST_WITH_POSTGRES="true"
go test -v ./orm/migrations -run TestPostgreSQL
```

### Performance Tests

```bash
# Large dataset migration performance
go test -v ./orm/migrations -run TestLargeDataset

# Migration idempotency testing
go test -v ./orm/migrations -run TestIdempotency

# Stress testing with PostgreSQL
./test_with_postgresql.sh --stress
```

### EF Migration System Tests

```bash
# Test EF Core-like migration tools
go test -v ./tools/ef-migrate

# Test migration CLI
go test -v ./tools/migration/test

# Integration tests
go test -v ./tools/migration/direct
```

## Environment Variables

### PostgreSQL Testing Variables

```bash
# Required for PostgreSQL testing
export TEST_DATABASE_URL="postgres://gra_user:gra_password@localhost:5433/gra_test?sslmode=disable"
export TEST_WITH_POSTGRES="true"

# Optional PostgreSQL configuration
export POSTGRES_HOST="localhost"
export POSTGRES_PORT="5433"
export POSTGRES_DB="gra_test"
export POSTGRES_USER="gra_user"
export POSTGRES_PASSWORD="gra_password"
```

### Coverage and Debugging

```bash
# Enable detailed logging
export DEBUG="true"

# Coverage configuration
export COVERAGE_MIN="60"  # Minimum coverage percentage

# Test timeout (for slow tests)
export TEST_TIMEOUT="30s"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test-sqlite:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.19'
      - run: go test -race -coverprofile=coverage.out ./...
      
  test-postgresql:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: gra_password
          POSTGRES_USER: gra_user
          POSTGRES_DB: gra_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.19'
      - run: ./test_with_postgresql.sh
        env:
          TEST_DATABASE_URL: postgres://gra_user:gra_password@localhost:5432/gra_test?sslmode=disable
          TEST_WITH_POSTGRES: "true"
```

### Local Development

```bash
# Create a development script
cat > test_dev.sh << 'EOF'
#!/bin/bash
set -e

echo "🧪 Running Development Tests"

# Run SQLite tests first (fast)
echo "📱 Testing with SQLite..."
go test -short ./orm/migrations

# Run PostgreSQL tests if Docker is available
if command -v docker &> /dev/null; then
    echo "🐘 Testing with PostgreSQL..."
    ./test_with_postgresql.sh
else
    echo "⚠️  Docker not available, skipping PostgreSQL tests"
fi

echo "✅ All tests completed!"
EOF

chmod +x test_dev.sh
./test_dev.sh
```

## Troubleshooting

### Common Issues

#### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL container is running
docker-compose -f docker-compose.test.yml ps

# Check PostgreSQL logs
docker-compose -f docker-compose.test.yml logs postgres-test

# Test connection manually
docker exec -it gra-postgres-test psql -U gra_user -d gra_test -c "SELECT version();"
```

#### Port Conflicts

```bash
# Check what's using port 5433
lsof -i :5433

# Use different port in docker-compose.test.yml
# ports:
#   - "5434:5432"  # Change to available port
```

#### Permission Issues

```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Fix file permissions
chmod +x test_with_postgresql.sh
chmod +x test_dev.sh
```

#### Memory Issues

```bash
# Increase Docker memory (Docker Desktop)
# Settings > Resources > Memory > 4GB+

# Run tests with lower parallelism
go test -p 1 ./...
```

### Database-Specific Issues

#### SQLite Lock Issues

```bash
# Use separate databases for parallel tests
go test -parallel 1 ./orm/migrations

# Check for unclosed database connections in tests
```

#### PostgreSQL Schema Issues

```bash
# Clean up test schemas
docker exec -it gra-postgres-test psql -U gra_user -d gra_test -c "
DROP SCHEMA IF EXISTS test_schema_123 CASCADE;
"

# Reset test database
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d postgres-test
```

## Performance Optimization

### Test Performance Tips

```bash
# Run tests in parallel (default)
go test -parallel 4 ./...

# Run specific test packages
go test ./orm/migrations ./tools/migration/test

# Use build cache
go clean -testcache  # Clear cache if needed
go test -count=1 ./...  # Skip cache for specific run

# Profile test performance
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./orm/migrations
go tool pprof cpu.prof
```

### Database Performance

```bash
# PostgreSQL: Check query performance
docker exec -it gra-postgres-test psql -U gra_user -d gra_test -c "
EXPLAIN ANALYZE SELECT * FROM information_schema.tables;
"

# Monitor database connections
docker exec -it gra-postgres-test psql -U gra_user -d gra_test -c "
SELECT count(*) FROM pg_stat_activity;
"
```

## Test Coverage Reports

### Generate Coverage Reports

```bash
# Combined coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Per-package coverage
go test -coverprofile=migrations.out ./orm/migrations
go tool cover -func=migrations.out

# Coverage with PostgreSQL
./test_with_postgresql.sh
# Creates postgres_coverage.html automatically
```

### Coverage Targets

- **Overall Project**: 60%+
- **Migration System**: 50%+
- **Core ORM**: 70%+
- **Critical Paths**: 80%+

## Summary

### Quick Commands Reference

```bash
# SQLite testing (fast, default)
go test ./...

# PostgreSQL testing (comprehensive)
./test_with_postgresql.sh

# Development testing
./test_dev.sh

# Specific package testing
go test -v ./orm/migrations

# Coverage testing
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Clean up
docker-compose -f docker-compose.test.yml down
```

### Test Strategy

1. **Local Development**: Use SQLite for rapid feedback
2. **Feature Testing**: Use PostgreSQL for database-specific features
3. **CI/CD**: Run both SQLite and PostgreSQL tests
4. **Pre-commit**: Quick SQLite test suite
5. **Pre-release**: Full PostgreSQL integration testing

The GRA framework provides comprehensive testing capabilities across multiple database systems, ensuring robust and reliable database operations in production environments.

# Migration Example

This example demonstrates how to use the GRA framework's MigrationRunner for automatic database migrations.

## Overview

The MigrationRunner provides functionality to:
- Create database tables automatically from Go struct definitions
- Handle table existence checks
- Track migration history
- Support both PostgreSQL and SQLite databases

## Running the Example

### Basic Usage

```bash
go run main.go
```

This will attempt to connect to a PostgreSQL database and run migrations.

### Running Tests

#### Unit Tests Only

```bash
go test -short
```

#### Integration Tests with Docker PostgreSQL

To run the full integration tests with a real PostgreSQL database:

1. **Start the test database:**
   ```bash
   docker-compose -f ../../docker-compose.test.yml up -d
   ```

2. **Run all tests:**
   ```bash
   go test -v
   ```

3. **Run tests with coverage:**
   ```bash
   go test -cover
   ```

4. **Stop the test database:**
   ```bash
   docker-compose -f ../../docker-compose.test.yml down
   ```

#### Alternative: Use Development Database

You can also use the main development database:

```bash
docker-compose -f ../../docker-compose.yml up -d
go test -v
```

## Test Coverage

The test suite includes:

- **Unit Tests**: Test individual functions and error handling without requiring a database
- **Integration Tests**: Test the complete migration flow with a real PostgreSQL database
- **Error Handling**: Test various failure scenarios and edge cases

Current test coverage: **73.3%**

### Coverage Details

- `NewMigrationRunner`: 100% (full constructor testing)
- `NewMigrationRunnerWithDriver`: 83.3% (multiple driver support)
- `Close`: 100% (connection cleanup)
- `AutoMigrate`: 62.5% (main migration logic with error handling)
- `createMigrationsTable`: 100% (migrations table creation)
- `migrateEntity`: 80% (individual entity migration)
- `tableExists`: 100% (table existence checks)
- `createTable`: 100% (table creation)
- `getTableName`: 100% (table name generation)
- `ShowStatus`: 56.2% (migration status display)
- `runMigrations`: 72.7% (main entry point)

## Database Configuration

### Test Database (docker-compose.test.yml)

- Host: `localhost`
- Port: `5433` 
- Database: `gra_test`
- User: `gra_user`
- Password: `gra_password`

### Development Database (docker-compose.yml)

- Host: `localhost`
- Port: `5432`
- Database: `gra_dev`
- User: `postgres`
- Password: `postgres`

## Features Tested

- ✅ Database connection management
- ✅ PostgreSQL and SQLite driver support
- ✅ Automatic table creation from Go structs
- ✅ Migration table tracking
- ✅ Table existence verification
- ✅ Error handling and recovery
- ✅ Idempotent migrations (can run multiple times safely)
- ✅ Individual entity migration
- ✅ Migration status reporting

## Known Issues

- Some entities (Order, OrderItem) have schema generation issues with default values that require quotes
- This is handled gracefully in tests with appropriate error checking

## Example Entities

The migration example works with these entity types:

- `Role` - User roles and permissions
- `Category` - Product categories
- `User` - Application users
- `Product` - E-commerce products
- `Order` - Customer orders (has schema generation issues)
- `OrderItem` - Items within orders (has schema generation issues)
- `Review` - Product reviews
- `UserRole` - Many-to-many user-role relationships

## Integration with CI/CD

The tests are designed to work in CI/CD environments:

- Integration tests are skipped when PostgreSQL is not available
- Tests use the `-short` flag to run only unit tests in resource-constrained environments
- Docker Compose provides consistent test environments

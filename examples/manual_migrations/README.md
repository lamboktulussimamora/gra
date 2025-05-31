# Manual Migrations Example

This example demonstrates a simple, direct database migration system in Go. The system provides:

1. **Direct Migration Runner**: Self-contained migration tool with hardcoded migrations
2. **CLI Interface**: Bash script wrapper for easy migration management
3. **Transaction Safety**: Each migration runs in a transaction with proper rollback
4. **Status Tracking**: Migration tracking via `schema_migrations` table

## Quick Overview

This example has been streamlined to focus on the essential migration functionality without complex abstractions. The migrations are hardcoded in the migration runner, making it self-contained and easy to understand.

## Structure

- `db_migrate_v2.sh` - Main CLI script for running migrations
- `go.mod` / `go.sum` - Go module definition
- `../../tools/migration/` - Migration tools (moved to root-level tools directory)
  - `direct_runner.go` - Complete migration runner with hardcoded migrations
  - `test_runner.go` - Simple connection test utility

## Migration Tools

The actual migration tools have been moved to `../../tools/migration/` for better organization:
- **`direct_runner`** - Full-featured migration runner (up/down/status commands)
- **`test_runner`** - Simple database connection and basic setup utility

## How to Use

### Quick Start

The migration system is ready to use out of the box:

```bash
# Test database connection
./db_migrate_v2.sh test

# Check migration status
./db_migrate_v2.sh status

# Apply all pending migrations
./db_migrate_v2.sh up

# Show help
./db_migrate_v2.sh -h
```

### Available Commands

- **`test`** - Test database connection
- **`status`** - Show current migration status  
- **`up`** - Apply all pending migrations
- **`down`** - Rollback last migration (shows "not implemented" message)

### Database Configuration

The CLI script uses these default connection parameters:
```bash
DB_HOST="localhost"
DB_PORT="5432" 
DB_USER="postgres"
DB_PASSWORD="MyPassword_123"
DB_NAME="gra"
```

You can override these with command-line options:
```bash
./db_migrate_v2.sh status --host myhost --port 5433 --user myuser --password mypass --dbname mydb
```

### Current Migrations

The system includes these hardcoded migrations in `direct_runner.go`:

1. **Migration 1** - Create initial schema:
   - Creates `users` table with id, name, email, timestamps
   - Creates `products` table with id, name, price, description, user_id FK, timestamps

2. **Migration 2** - Add performance indexes:
   - Adds index on `users.email`
   - Adds index on `products.user_id`

3. **Migration 3** - Add categories:
   - Creates `categories` table with id, name, description, created_at

### Adding New Migrations

To add new migrations, edit `../../tools/migration/direct_runner.go` and add to the migrations slice:

```go
{
    Version:     4,
    Description: "Add your new migration description",
    SQL: `
        -- Your SQL statements here
        ALTER TABLE products ADD COLUMN category_id INTEGER REFERENCES categories(id);
    `,
},
```

## Direct Tool Usage

You can also use the migration tools directly from the root-level tools directory:

```bash
# Build the tools (if needed)
cd ../../tools/migration
go build -o direct_runner direct_runner.go
go build -o test_runner test_runner.go

# Use direct runner
./../../tools/migration/direct_runner --conn "postgres://user:pass@host:port/db" --status
./../../tools/migration/direct_runner --conn "postgres://user:pass@host:port/db" --up

# Use test runner for simple connection testing
./../../tools/migration/test_runner --conn "postgres://user:pass@host:port/db" --up
```

## Requirements

- Go 1.16+
- PostgreSQL database
- `github.com/lib/pq` driver (automatically handled by go.mod)

## Architecture Notes

This example demonstrates a **simplified, self-contained approach** to database migrations:

- **No complex abstractions** - Direct SQL in migration runner
- **No external files** - Migrations are hardcoded in Go source
- **Minimal dependencies** - Only database driver required
- **Transaction safety** - Each migration runs in a transaction
- **Simple tracking** - Uses standard `schema_migrations` table

For more complex migration needs with model-driven approaches, see other examples in the repository.
````

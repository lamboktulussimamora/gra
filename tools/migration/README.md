# Migration Tools

Database migration utilities for the GRA framework.

## Tools

### direct_runner.go
A comprehensive migration runner that supports:
- **Up migrations**: Apply pending migrations
- **Status check**: Show applied migrations  
- **Transaction safety**: Each migration runs in a transaction
- **Error handling**: Proper rollback on failures

**Features:**
- Hardcoded migrations (no external files needed)
- Migration tracking via `schema_migrations` table
- Verbose output support
- PostgreSQL support

**Usage:**
```bash
# Build
go build -o direct_runner direct_runner.go

# Apply migrations
./direct_runner --conn "postgres://user:pass@host/db" --up

# Check status
./direct_runner --conn "postgres://user:pass@host/db" --status

# Verbose output
./direct_runner --conn "postgres://user:pass@host/db" --up --verbose
```

### test_runner.go
A simple migration test utility for basic database setup:
- **Connection testing**: Verify database connectivity
- **Basic schema**: Creates users table and migrations tracking
- **Minimal footprint**: Simple, focused functionality

**Usage:**
```bash
# Build
go build -o test_runner test_runner.go

# Test connection and apply basic migration
./test_runner --conn "postgres://user:pass@host/db" --up
```

## Migration Schema

Both tools use a `schema_migrations` table to track applied migrations:
```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Examples

For complete usage examples with CLI scripts, see:
- `examples/manual_migrations/` - Full example with CLI wrapper
- `examples/manual_migrations/db_migrate_v2.sh` - Bash script wrapper

## Requirements

- Go 1.16+
- PostgreSQL database
- github.com/lib/pq driver (included in go.mod)

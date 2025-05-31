# GRA Tools

This directory contains utility tools for the GRA framework.

## Migration Tools

The `migration/` directory contains database migration utilities:

- **`direct/direct_runner.go`** - Complete migration runner with up/down/status commands
- **`test/test_runner.go`** - Simple migration test utility

### Usage

#### Building Tools
```bash
# Build direct runner
go build -o tools/migration/direct_runner tools/migration/direct/direct_runner.go

# Build test runner  
go build -o tools/migration/test_runner tools/migration/test/test_runner.go
```

#### Running Migrations
```bash
# Check migration status
./tools/migration/direct_runner --conn "postgres://user:pass@host/db" --status

# Apply pending migrations
./tools/migration/direct_runner --conn "postgres://user:pass@host/db" --up

# Test database connection
./tools/migration/test_runner --conn "postgres://user:pass@host/db" --up
```

### Examples

See the `examples/manual_migrations/` directory for complete usage examples and CLI scripts.

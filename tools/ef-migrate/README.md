# EF Migrate - Entity Framework Migration Tool for Go

A comprehensive Entity Framework-like migration CLI tool for the GRA framework that provides database migration management with commands similar to Entity Framework Core.

## Features

- 🚀 **Multiple Database Support**: PostgreSQL and SQLite
- 📁 **File-based Migrations**: Create and manage migration files
- 🔄 **Rollback Support**: Safely rollback migrations
- 📊 **Migration Status**: View current migration state
- 📜 **Script Generation**: Generate SQL scripts for migrations
- 🌐 **Unicode Support**: Handle international characters and emoji
- 🔒 **Security**: Password masking and secure connection handling
- ⚡ **Performance**: Optimized for large-scale migrations
- 🧪 **Comprehensive Testing**: 18+ test categories with edge cases

## Installation

```bash
# Build the tool
go build -o ef-migrate main.go

# Or run directly
go run main.go [command] [options]
```

## Commands

### Core Migration Commands

#### Add Migration
Create a new migration file:
```bash
./ef-migrate add-migration InitialCreate
./ef-migrate add-migration "Add Users Table" --verbose
```

#### Update Database
Apply pending migrations:
```bash
./ef-migrate update-database
./ef-migrate update-database TargetMigration
```

#### Migration Status
View current migration status:
```bash
./ef-migrate status
```

#### List Migrations
Display migration history:
```bash
./ef-migrate list
./ef-migrate get-migration
```

#### Generate Script
Create SQL script for migrations:
```bash
./ef-migrate script
./ef-migrate script FromMigration ToMigration
```

#### Rollback
Rollback to a previous migration:
```bash
./ef-migrate rollback PreviousMigration
```

#### Remove Migration
Remove the last migration:
```bash
./ef-migrate remove-migration
```

## Connection Options

### Connection String
```bash
# PostgreSQL
./ef-migrate status -connection "postgres://user:password@localhost/mydb?sslmode=disable"

# SQLite
./ef-migrate status -connection "./mydb.db"
```

### Individual Parameters (PostgreSQL)
```bash
./ef-migrate status -host localhost -port 5432 -user myuser -password mypass -database mydb -sslmode disable
```

### Environment Variable
```bash
export DATABASE_URL="postgres://user:password@localhost/mydb"
./ef-migrate status
```

## Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `-connection` | Database connection string | `""` |
| `-migrations-dir` | Directory for migration files | `./migrations` |
| `-verbose` | Enable verbose logging | `false` |
| `-host` | Database host (PostgreSQL) | `""` |
| `-port` | Database port (PostgreSQL) | `5432` |
| `-user` | Database user (PostgreSQL) | `""` |
| `-password` | Database password (PostgreSQL) | `""` |
| `-database` | Database name (PostgreSQL) | `""` |
| `-sslmode` | SSL mode (PostgreSQL) | `disable` |

## Migration File Structure

Migration files are stored in the migrations directory with the following structure:

```
migrations/
├── 20240101_120000_InitialCreate.sql
├── 20240102_140000_AddUsersTable.sql
└── 20240103_160000_AddIndexes.sql
```

Each migration file contains:
```sql
-- Migration: InitialCreate
-- Description: Initial database setup
-- TODO: Add your SQL here

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- Rollback for: InitialCreate
-- TODO: Add rollback SQL here

DROP TABLE users;
```

## Examples

### Basic Usage
```bash
# Create a new migration
./ef-migrate add-migration "Create Users Table"

# Edit the generated migration file
# migrations/20240101_120000_Create_Users_Table.sql

# Apply the migration
./ef-migrate update-database

# Check status
./ef-migrate status
```

### Advanced Usage
```bash
# Rollback to specific migration
./ef-migrate rollback 20240101_120000_InitialCreate

# Generate script between migrations
./ef-migrate script InitialCreate AddUsersTable

# Remove last migration
./ef-migrate remove-migration

# Verbose logging
./ef-migrate update-database --verbose
```

### Unicode Support
The tool supports Unicode characters in migration names:
```bash
./ef-migrate add-migration "添加用户表"  # Chinese
./ef-migrate add-migration "إضافة جدول المستخدمين"  # Arabic
./ef-migrate add-migration "Добавить таблицу пользователей"  # Russian
./ef-migrate add-migration "Migration with 🚀 emoji"  # Emoji
```

## Testing

Run the comprehensive test suite:
```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestAddMigration

# Run with coverage
go test -v -cover
```

The test suite covers:
- Basic migration operations
- Database schema management
- Command-line interface
- Error handling and validation
- Memory leak prevention
- Performance optimization
- Security validation
- Unicode and internationalization
- Edge cases and boundary conditions

## Architecture

The tool is built on top of the GRA framework's migration system:

```
┌─────────────────────────────────────┐
│           CLI Interface             │
├─────────────────────────────────────┤
│        Command Routing              │
├─────────────────────────────────────┤
│     EF Migration Manager            │
├─────────────────────────────────────┤
│    Database Connection Layer        │
├─────────────────────────────────────┤
│   PostgreSQL    │    SQLite         │
└─────────────────────────────────────┘
```

## Database Support

### PostgreSQL
- Full feature support
- Connection string or individual parameters
- SSL/TLS support
- Transaction management

### SQLite
- File and in-memory databases
- Automatic database creation
- Foreign key support
- WAL mode compatibility

## Error Handling

The tool provides comprehensive error handling:
- Connection validation
- Migration file validation
- SQL syntax checking
- Rollback safety checks
- Unicode encoding validation

## Security Features

- Password masking in logs
- SQL injection prevention
- Secure connection handling
- Migration file validation
- Access control checks

## Performance

- Optimized for large migration sets
- Memory efficient processing
- Concurrent-safe operations
- Connection pooling
- Batch processing support

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run the test suite
5. Submit a pull request

## License

This tool is part of the GRA framework and follows the same license terms.

## Support

For issues and questions:
1. Check the test suite for usage examples
2. Review the source code documentation
3. Create an issue in the GRA framework repository

# GRA Hybrid Migration System

A hybrid database migration system for the GRA framework that combines automatic schema detection with SQL file generation for review and version control, similar to Entity Framework Core's approach.

## Features

- **Automatic Schema Detection**: Analyzes Go struct definitions to detect schema changes
- **Hybrid Approach**: Combines fast development iteration with production safety
- **EF Core-Style API**: Familiar DbSet registration pattern for model management  
- **Change Detection**: Automatically detects table and column additions, modifications, and deletions
- **SQL Generation**: Creates reviewable migration scripts with up/down support
- **Multiple Database Support**: PostgreSQL, MySQL, and SQLite
- **Migration History**: Tracks applied migrations with checksums and rollback support
- **Safety Features**: Destructive change detection and multiple migration modes

## Quick Start

### 1. Define Your Models

```go
type User struct {
    ID        int64     `db:"id" migration:"primary_key,auto_increment"`
    Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
    Name      string    `db:"name" migration:"not_null,max_length:100"`
    IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
    CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

type Post struct {
    ID       int64  `db:"id" migration:"primary_key,auto_increment"`
    UserID   int64  `db:"user_id" migration:"not_null,foreign_key:users.id"`
    Title    string `db:"title" migration:"not_null,max_length:255"`
    Content  string `db:"content" migration:"type:TEXT"`
}
```

### 2. Register Models and Create Migrations

```go
// Connect to database
db, _ := sql.Open("postgres", "your-connection-string")

// Create migrator
migrator := migrations.NewHybridMigrator(db, migrations.PostgreSQL, "./migrations")

// Register models (EF Core-style)
migrator.DbSet(&User{})
migrator.DbSet(&Post{})

// Create migration for detected changes
migration, err := migrator.AddMigration("initial_schema", migrations.Interactive)
if err != nil {
    log.Fatal(err)
}

// Apply migrations
err = migrator.ApplyMigrations(migrations.Automatic)
```

### 3. Use CLI Tool

```bash
# Create migration
./migrate -db "postgres://user:pass@localhost/db" add "add_user_table"

# Apply migrations
./migrate -db "postgres://user:pass@localhost/db" apply

# Check status
./migrate -db "postgres://user:pass@localhost/db" status

# Revert last migration
./migrate -db "postgres://user:pass@localhost/db" revert
```

## Migration Tags

The system uses struct tags to define schema properties:

### Basic Tags
- `primary_key` - Marks field as primary key
- `auto_increment` - Auto-incrementing field (SERIAL/AUTO_INCREMENT)
- `not_null` - Field cannot be null
- `nullable` - Field can be null (default)
- `unique` - Unique constraint
- `index` - Create index on field

### Type and Size Tags
- `type:TEXT` - Specify database type
- `max_length:255` - Maximum string length
- `precision:10,scale:2` - Decimal precision and scale

### Default Values
- `default:true` - Boolean default
- `default:'value'` - String default (quoted)
- `default:CURRENT_TIMESTAMP` - SQL function default

### Relationships
- `foreign_key:table.column` - Foreign key constraint

### Example
```go
type User struct {
    ID       int64  `db:"id" migration:"primary_key,auto_increment"`
    Email    string `db:"email" migration:"unique,not_null,max_length:255,index"`
    Name     string `db:"name" migration:"not_null,max_length:100"`
    IsActive bool   `db:"is_active" migration:"not_null,default:true"`
    Balance  float64 `db:"balance" migration:"precision:10,scale:2,default:0.00"`
}
```

## Migration Modes

### Automatic Mode
- Safe for non-destructive changes only
- No user interaction required
- Fails on destructive operations

```go
migrator.ApplyMigrations(migrations.Automatic)
```

### Interactive Mode  
- Prompts for confirmation on destructive changes
- Default mode for development
- Provides detailed change information

```go
migrator.ApplyMigrations(migrations.Interactive)
```

### Generate Only Mode
- Creates migration files without applying them
- Perfect for production pipelines
- Allows manual review before deployment

```go
migrator.AddMigration("migration_name", migrations.GenerateOnly)
```

### Force Destructive Mode
- Applies all changes without confirmation
- Use with extreme caution
- Suitable for test environments

```go
migrator.ApplyMigrations(migrations.ForceDestructive)
```

## Change Detection

The system automatically detects:

### Table Changes
- **Create Table**: New model added
- **Drop Table**: Model removed (destructive)

### Column Changes  
- **Add Column**: New field in model
- **Drop Column**: Field removed from model (destructive)
- **Alter Column**: Field properties changed (potentially destructive)

### Index Changes
- **Create Index**: Index tag added to field
- **Drop Index**: Index tag removed from field

### Constraint Changes
- **Foreign Keys**: Detected from `foreign_key` tags
- **Unique Constraints**: Detected from `unique` tags
- **Check Constraints**: Future enhancement

## Generated Migration Files

Migration files follow a structured format:

```sql
-- Migration: add_user_profile
-- Created: 2024-01-15T10:30:00Z
-- Checksum: abc123def456
-- Mode: Interactive
-- Has Destructive: false
-- Requires Review: false

-- +migrate Up
CREATE TABLE "users" (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "idx_users_email" ON "users" (email);

-- +migrate Down  
DROP INDEX IF EXISTS "idx_users_email";
DROP TABLE IF EXISTS "users";
```

## Database Support

### PostgreSQL
- Full feature support
- SERIAL/BIGSERIAL for auto-increment
- Complete constraint support

### MySQL
- Full feature support  
- AUTO_INCREMENT for auto-increment
- InnoDB foreign key support

### SQLite
- Basic support
- Limited ALTER TABLE capabilities
- Foreign key support with PRAGMA

## Architecture

### Core Components

1. **ModelRegistry** - Manages model registration and snapshots
2. **DatabaseInspector** - Reads current database schema
3. **ChangeDetector** - Compares models with database state
4. **SQLGenerator** - Creates migration scripts
5. **HybridMigrator** - Orchestrates the entire process

### Data Flow

```
Models → Registry → Snapshots → Detector → Changes → Generator → SQL Files → Migrator → Database
```

## CLI Tool

The included CLI tool provides command-line access to all migration features:

```bash
# Build the CLI
go build -o migrate ./cmd/migrate

# Available commands
./migrate -help

# Example usage
./migrate -db "postgres://localhost/mydb" -migrations-dir "./db/migrations" add "create_users"
./migrate -db "postgres://localhost/mydb" apply --auto
./migrate -db "postgres://localhost/mydb" status
```

### CLI Options
- `-db`: Database connection string
- `-driver`: Database driver (postgres, mysql, sqlite)
- `-migrations-dir`: Migration files directory
- `-models-dir`: Models directory (for auto-discovery)

## Advanced Usage

### Custom Table Names
```go
// Explicit table name
migrator.DbSet(&User{}, "custom_users")

// Multiple models to same table (inheritance scenarios)
migrator.DbSet(&AdminUser{}, "users")
migrator.DbSet(&RegularUser{}, "users")
```

### Complex Relationships
```go
type Order struct {
    ID         int64 `db:"id" migration:"primary_key,auto_increment"`
    CustomerID int64 `db:"customer_id" migration:"not_null,foreign_key:customers.id,index"`
    ProductID  int64 `db:"product_id" migration:"not_null,foreign_key:products.id,index"`
    Quantity   int   `db:"quantity" migration:"not_null,default:1"`
}
```

### Embedded Structs
```go
type BaseModel struct {
    ID        int64     `db:"id" migration:"primary_key,auto_increment"`
    CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

type User struct {
    BaseModel // Embedded struct fields are included
    Email string `db:"email" migration:"unique,not_null,max_length:255"`
    Name  string `db:"name" migration:"not_null,max_length:100"`
}
```

## Production Deployment

### Recommended Workflow

1. **Development**: Use Interactive mode for rapid iteration
2. **Testing**: Generate migrations with GenerateOnly mode
3. **Review**: Manual review of generated SQL scripts  
4. **CI/CD**: Apply migrations with Automatic mode
5. **Production**: Use pre-generated migration files only

### Example CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Generate Migrations
  run: ./migrate -db "$DB_URL" generate "deploy_$(date +%Y%m%d)"
  
- name: Review Migration
  run: cat migrations/*.sql
  
- name: Apply Migration
  run: ./migrate -db "$PROD_DB_URL" apply --auto
```

## Best Practices

### Model Design
- Use consistent naming conventions
- Always include created_at/updated_at timestamps
- Consider soft deletes with deleted_at fields
- Use appropriate data types and constraints

### Migration Strategy
- Create small, focused migrations
- Test migrations on copy of production data
- Keep migrations reversible when possible
- Use descriptive migration names

### Production Safety
- Never auto-generate migrations in production
- Always review generated SQL before applying
- Test rollback procedures
- Monitor migration performance

## Error Handling

The system provides comprehensive error handling:

### Validation Errors
- Circular dependency detection
- Orphaned foreign key detection  
- Data loss potential warnings

### Migration Errors
- SQL execution failures with context
- Rollback on partial failure
- Detailed error messages

### Example Error Handling
```go
migration, err := migrator.AddMigration("risky_change", migrations.Interactive)
if err != nil {
    // Handle migration creation errors
    log.Printf("Migration creation failed: %v", err)
    return
}

if len(migration.Warnings) > 0 {
    // Review warnings before proceeding
    for _, warning := range migration.Warnings {
        log.Printf("WARNING: %s", warning)
    }
}

err = migrator.ApplyMigrations(migrations.Automatic)
if err != nil {
    // Handle application errors
    log.Printf("Migration application failed: %v", err)
    // Consider rollback or manual intervention
}
```

## Comparison with Other Systems

### vs. Entity Framework Core
- ✅ Similar DbSet registration pattern
- ✅ Automatic change detection
- ✅ Up/Down migration scripts
- ✅ Migration history tracking
- ➖ No LINQ support (not applicable to Go)

### vs. Rails ActiveRecord
- ✅ Migration file generation
- ✅ Rollback support
- ✅ Schema versioning
- ➕ Automatic change detection (Rails requires manual migration writing)

### vs. Golang-Migrate
- ➕ Automatic schema detection (golang-migrate is manual)
- ✅ Multiple database support
- ✅ Migration history
- ➖ Less mature ecosystem

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

### Development Setup
```bash
git clone https://github.com/your-org/gra
cd gra/orm/migrations
go mod tidy
go test ./...
```

## License

This project is part of the GRA framework and follows the same license terms.

## Support

- GitHub Issues: Report bugs and feature requests
- Documentation: See examples/ directory for more usage patterns
- Community: Join the GRA framework community discussions

---

## Roadmap

### v1.1 (Planned)
- MySQL and SQLite inspector implementations
- Advanced constraint support (CHECK constraints)
- Migration squashing for optimization
- Performance monitoring and metrics

### v1.2 (Future)
- Visual migration designer
- Database schema diff tool
- Integration with popular Go ORMs
- Cloud database provider integrations

### v2.0 (Future)
- Multi-database migration support
- Advanced rollback strategies
- Schema branching and merging
- Enterprise features (audit logs, approval workflows)

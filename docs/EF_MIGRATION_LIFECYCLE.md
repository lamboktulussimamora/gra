# Entity Framework Core Migration Lifecycle in GRA

The GRA Framework provides a complete Entity Framework Core-like migration system that allows you to manage database schema changes with the same familiar commands and lifecycle as EF Core.

## 🚀 Quick Start

### Installation

```bash
# Build the EF migration CLI tool
cd tools/ef-migrate
go build -o ef-migrate main.go

# Make it executable (Linux/Mac)
chmod +x ef-migrate

# Add to PATH (optional)
sudo mv ef-migrate /usr/local/bin/
```

### Basic Usage

```bash
# Set your database connection
export DATABASE_URL="postgres://user:password@localhost/mydb?sslmode=disable"

# Create your first migration
ef-migrate add-migration InitialCreate "Create initial database schema"

# Apply migrations to database
ef-migrate update-database

# Check migration status
ef-migrate status
```

## 📋 Migration Lifecycle Commands

### 1. Add-Migration (Create New Migration)

Creates a new migration file with UP and DOWN SQL scripts.

```bash
# Basic usage
ef-migrate add-migration <MigrationName>

# With description
ef-migrate add-migration CreateUsersTable "Initial user table with authentication"

# Example output:
🔧 Creating migration: CreateUsersTable
✅ Migration created: 1703123456_CreateUsersTable
📁 File: ./migrations/1703123456_CreateUsersTable.sql
📝 Edit the migration file and run 'update-database' to apply
```

**Generated Migration File:**
```sql
-- Migration: CreateUsersTable
-- Description: Initial user table with authentication
-- Created: 2023-12-21 10:30:45
-- Version: 1703123456

-- UP Migration
-- Migration: CreateUsersTable
-- Description: Initial user table with authentication
-- TODO: Add your SQL here

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- DOWN Migration (for rollback)
-- Rollback for: CreateUsersTable
-- TODO: Add rollback SQL here

DROP TABLE IF EXISTS users;
```

### 2. Update-Database (Apply Migrations)

Applies pending migrations to the database.

```bash
# Apply all pending migrations
ef-migrate update-database

# Apply migrations up to a specific migration
ef-migrate update-database CreateUsersTable

# Apply migrations up to a specific version
ef-migrate update-database 1703123456_CreateUsersTable
```

**Example Output:**
```
🚀 Updating database...
📋 Migration History:
⏳ Pending Migrations (2):
   ⏳ 1703123456_CreateUsersTable - Initial user table with authentication
   ⏳ 1703123500_AddUserProfiles - Add user profile information

Applying 2 migration(s)...
Applying migration: 1703123456_CreateUsersTable
✓ Applied migration: 1703123456_CreateUsersTable (125ms)
Applying migration: 1703123500_AddUserProfiles
✓ Applied migration: 1703123500_AddUserProfiles (89ms)
✓ All migrations applied successfully
✅ Database updated successfully!
```

### 3. Get-Migration (List Migrations)

Shows the complete migration history with status.

```bash
ef-migrate get-migration
# or
ef-migrate list
```

**Example Output:**
```
📋 Migration History:
====================

✅ Applied Migrations (2):
   ✅ 1703123456_CreateUsersTable (2023-12-21 10:35:22) - Initial user table
   ✅ 1703123500_AddUserProfiles (2023-12-21 10:36:15) - Add user profiles

⏳ Pending Migrations (1):
   ⏳ 1703123600_AddUserSettings - User preference settings

📊 Summary: 2 applied, 1 pending, 0 failed
```

### 4. Rollback (Update-Database with Target)

Rolls back migrations to a specific point.

```bash
# Rollback to a specific migration
ef-migrate rollback CreateUsersTable

# Rollback to migration by ID
ef-migrate rollback 1703123456_CreateUsersTable
```

**Example Output:**
```
⏪ Rolling back to migration: CreateUsersTable
Rolling back migration: 1703123600_AddUserSettings
✓ Rolled back migration: 1703123600_AddUserSettings (67ms)
Rolling back migration: 1703123500_AddUserProfiles  
✓ Rolled back migration: 1703123500_AddUserProfiles (45ms)
✅ Rollback completed successfully!
```

### 5. Status (Quick Overview)

Shows a quick summary of migration status.

```bash
ef-migrate status
```

**Example Output:**
```
📊 Migration Status:
===================
Database: myapp_db
Applied:  3 migrations
Pending:  1 migrations
Failed:   0 migrations
Latest:   1703123500_AddUserProfiles (2023-12-21 10:36:15)
Next:     1703123600_AddUserSettings
```

### 6. Script (Generate SQL)

Generates SQL scripts for migrations without applying them.

```bash
# Generate script for all pending migrations
ef-migrate script

# Generate script up to specific migration
ef-migrate script AddUserProfiles
```

**Example Output:**
```
📜 Generating migration script...
-- Generated Migration Script
-- Generated at: 2023-12-21 10:45:30
-- Migrations: 2
-- ==========================================

-- Migration 1: 1703123456_CreateUsersTable
-- Description: Initial user table
-- ------------------------------------------
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Migration 2: 1703123500_AddUserProfiles
-- Description: Add user profiles
-- ------------------------------------------
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    bio TEXT
);

-- End of migration script
```

## 🏗️ Migration System Architecture

### Database Schema

The EF migration system creates three tables:

1. **`__ef_migrations_history`** - EF Core compatible history table
2. **`__migration_history`** - Detailed migration tracking
3. **`__model_snapshot`** - Model versioning and snapshots

```sql
-- EF Core compatible table
CREATE TABLE __ef_migrations_history (
    migration_id VARCHAR(150) NOT NULL PRIMARY KEY,
    product_version VARCHAR(32) NOT NULL
);

-- Detailed tracking table
CREATE TABLE __migration_history (
    id SERIAL PRIMARY KEY,
    migration_id VARCHAR(150) NOT NULL,
    name VARCHAR(255) NOT NULL,
    version BIGINT NOT NULL,
    description TEXT,
    up_sql TEXT NOT NULL,
    down_sql TEXT NOT NULL,
    applied_at TIMESTAMP,
    rolled_back_at TIMESTAMP,
    execution_time_ms INTEGER,
    state VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    checksum VARCHAR(64)
);
```

### Migration States

```go
type MigrationState int

const (
    MigrationStatePending MigrationState = iota  // Not yet applied
    MigrationStateApplied                        // Successfully applied
    MigrationStateFailed                         // Failed to apply
)
```

### Migration Structure

```go
type Migration struct {
    ID          string             `json:"id"`           // Unique migration identifier
    Name        string             `json:"name"`         // Human-readable name
    Version     int64              `json:"version"`      // Unix timestamp version
    Description string             `json:"description"`  // Migration description
    UpSQL       string             `json:"up_sql"`       // Forward migration SQL
    DownSQL     string             `json:"down_sql"`     // Rollback migration SQL
    AppliedAt   time.Time          `json:"applied_at"`   // When applied
    State       MigrationState     `json:"state"`        // Current state
}
```

## 🔄 Complete Lifecycle Example

Here's a complete example showing the migration lifecycle:

### Step 1: Initialize Project

```bash
# Create new project directory
mkdir myapp && cd myapp

# Initialize migration system
ef-migrate status  # This initializes the schema tables
```

### Step 2: Create Initial Migration

```bash
ef-migrate add-migration InitialCreate "Create initial database schema"
```

Edit the generated migration file:

```sql
-- UP Migration
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- DOWN Migration
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Step 3: Apply Initial Migration

```bash
ef-migrate update-database
```

### Step 4: Add More Migrations

```bash
# Add user profiles
ef-migrate add-migration AddUserProfiles "Add user profile information"

# Add user settings
ef-migrate add-migration AddUserSettings "Add user preference settings"

# Add audit logging
ef-migrate add-migration AddAuditLog "Add audit trail for user actions"
```

### Step 5: Apply Specific Migration

```bash
# Apply only up to user profiles
ef-migrate update-database AddUserProfiles
```

### Step 6: Check Status

```bash
ef-migrate status
# Shows: 2 applied, 2 pending
```

### Step 7: Apply All Remaining

```bash
ef-migrate update-database
```

### Step 8: Rollback if Needed

```bash
# Rollback to before user settings
ef-migrate rollback AddUserProfiles
```

## 🛠️ Advanced Features

### Automatic Migration Generation

Generate migrations automatically from Go structs:

```go
type User struct {
    ID        int    `db:"id" migrations:"primary_key,auto_increment"`
    Email     string `db:"email" migrations:"unique,not_null,type:varchar(255)"`
    Name      string `db:"name" migrations:"not_null,type:varchar(100)"`
    CreatedAt string `db:"created_at" migrations:"default:CURRENT_TIMESTAMP"`
}

// Generate migration from entity
migration, err := manager.GenerateMigrationFromEntity("User", User{})
```

### Transaction Safety

All migrations run in database transactions:
- ✅ **Atomic**: Either fully applied or fully rolled back
- ✅ **Consistent**: Database remains in valid state
- ✅ **Isolated**: Concurrent operations don't interfere
- ✅ **Durable**: Changes are permanently stored

### Migration Validation

Built-in validation ensures:
- ✅ **Dependency Resolution**: Migrations applied in correct order
- ✅ **Checksum Verification**: Migration content hasn't changed
- ✅ **State Consistency**: Database state matches migration history
- ✅ **Error Recovery**: Failed migrations can be retried or rolled back

## 🔧 Configuration Options

### Environment Variables

```bash
export DATABASE_URL="postgres://user:pass@localhost/db?sslmode=disable"
export MIGRATION_TABLE="__ef_migrations_history"
export MIGRATION_HISTORY_TABLE="__migration_history"
export MIGRATION_SNAPSHOT_TABLE="__model_snapshot"
```

### CLI Options

```bash
ef-migrate -connection "..." -migrations-dir "./db/migrations" -verbose update-database
```

### Programmatic Configuration

```go
manager := migrations.NewEFMigrationManager(db, logger)
manager.SetMigrationTable("custom_ef_migrations")
manager.SetHistoryTable("custom_migration_history")
manager.SetSnapshotTable("custom_model_snapshots")
```

## 🆚 EF Core vs GRA Migrations

| Feature | EF Core | GRA Framework | Status |
|---------|---------|---------------|--------|
| Add-Migration | ✅ | ✅ | Equivalent |
| Update-Database | ✅ | ✅ | Equivalent |
| Remove-Migration | ✅ | ✅ | Equivalent |
| Script-Migration | ✅ | ✅ | Equivalent |
| Get-Migration | ✅ | ✅ | Enhanced with more details |
| Rollback Support | ✅ | ✅ | Full support |
| Transaction Safety | ✅ | ✅ | Full support |
| Auto-generation | ✅ | ✅ | From Go structs |
| Migration History | ✅ | ✅ | Enhanced tracking |
| Model Snapshots | ✅ | ✅ | Planned |
| Seed Data | ✅ | ⚠️ | In development |

## 🚨 Best Practices

### 1. Migration Naming
```bash
# ✅ Good
ef-migrate add-migration CreateUsersTable
ef-migrate add-migration AddEmailIndexToUsers
ef-migrate add-migration UpdateUserPasswordPolicy

# ❌ Bad
ef-migrate add-migration Migration1
ef-migrate add-migration Fix
ef-migrate add-migration Update
```

### 2. Migration Content
```sql
-- ✅ Good: Explicit and reversible
-- UP Migration
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL
);
CREATE INDEX idx_users_email ON users(email);

-- DOWN Migration  
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;

-- ❌ Bad: Not reversible
-- UP Migration
CREATE TABLE users (id SERIAL, email TEXT);
-- DOWN Migration
-- TODO: Add rollback
```

### 3. Production Deployment
```bash
# ✅ Always generate and review scripts first
ef-migrate script > migration-script-v1.2.sql
# Review the script
# Apply in production with monitoring

# ❌ Don't apply directly in production
ef-migrate update-database  # Dangerous in production
```

### 4. Team Collaboration
```bash
# ✅ Check migration status before creating new ones
ef-migrate status
ef-migrate get-migration

# ✅ Pull latest changes and apply migrations
git pull
ef-migrate update-database

# ✅ Create descriptive migrations
ef-migrate add-migration AddUserRoles "Add role-based access control for users"
```

## 🐛 Troubleshooting

### Common Issues

#### 1. Migration Order Conflicts
```bash
# Problem: Migrations applied out of order
# Solution: Check migration history and dependencies
ef-migrate get-migration
ef-migrate rollback <last-good-migration>
ef-migrate update-database
```

#### 2. Failed Migrations
```bash
# Problem: Migration failed halfway through
# Solution: Check error details and fix
ef-migrate status  # Shows failed migrations
# Fix the migration SQL
ef-migrate update-database  # Retry
```

#### 3. Connection Issues
```bash
# Problem: Cannot connect to database
# Solution: Verify connection string
ef-migrate -connection "postgres://user:pass@host/db" status
```

### Recovery Procedures

#### Reset Migration State
```bash
# WARNING: This will lose migration history
# Only use in development

# 1. Drop migration tables
psql -c "DROP TABLE IF EXISTS __ef_migrations_history, __migration_history, __model_snapshot;"

# 2. Reinitialize
ef-migrate status

# 3. Create new initial migration
ef-migrate add-migration InitialCreate
```

#### Repair Corrupted History
```sql
-- Manual repair of migration history
UPDATE __migration_history 
SET state = 'applied', applied_at = CURRENT_TIMESTAMP 
WHERE migration_id = 'problematic_migration_id';
```

## 📚 Additional Resources

- [GRA Framework Documentation](https://github.com/your-org/gra/docs)
- [Entity Framework Core Migrations](https://docs.microsoft.com/en-us/ef/core/managing-schemas/migrations/)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
- [Database Schema Versioning](https://martinfowler.com/articles/evodb.html)

## 🤝 Contributing

To contribute to the EF migration system:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

### Development Setup

```bash
git clone https://github.com/your-org/gra.git
cd gra/orm/migrations
go test ./...
```

### Testing

```bash
# Run migration tests
cd orm/migrations
go test -v

# Run CLI tests  
cd tools/ef-migrate
go test -v
```

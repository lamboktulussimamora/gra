# PostgreSQL Setup Guide for GRA Framework

## Quick Start with PostgreSQL

### 1. Install PostgreSQL Dependencies

```bash
go get github.com/lib/pq
```

### 2. Setup PostgreSQL Database

#### Option A: Local PostgreSQL Installation

```bash
# Install PostgreSQL (macOS with Homebrew)
brew install postgresql
brew services start postgresql

# Create database
createdb gra_demo

# Create user (optional)
psql -d gra_demo -c "CREATE USER gra_user WITH PASSWORD 'password';"
psql -d gra_demo -c "GRANT ALL PRIVILEGES ON DATABASE gra_demo TO gra_user;"
```

#### Option B: Docker PostgreSQL

```bash
# Run PostgreSQL in Docker
docker run --name gra-postgres \
  -e POSTGRES_DB=gra_demo \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -d postgres:15

# Verify connection
docker exec -it gra-postgres psql -U postgres -d gra_demo -c "SELECT version();"
```

### 3. Connection Strings

#### Basic Connection
```go
connectionString := "host=localhost port=5432 user=postgres dbname=gra_demo sslmode=disable password=postgres"
```

#### With Environment Variables
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gra_demo?sslmode=disable"
```

### 4. Code Examples

#### Basic Usage
```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/lamboktulussimamora/gra/orm/dbcontext"
    _ "github.com/lib/pq"
)

func main() {
    // PostgreSQL connection
    db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=gra_demo sslmode=disable password=postgres")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create context - driver automatically detected
    ctx := dbcontext.NewEnhancedDbContextWithDB(db)
    
    // Use normally - placeholders automatically converted
    user := &User{Name: "John", Email: "john@example.com"}
    ctx.Add(user)
    ctx.SaveChanges()
    
    // Queries work seamlessly
    users := dbcontext.NewEnhancedDbSet[User](ctx)
    found, _ := users.Where("name = ?", "John").FirstOrDefault()
}
```

#### With Auto-Migration
```go
import (
    "github.com/lamboktulussimamora/gra/orm/migrations"
    "github.com/lamboktulussimamora/gra/orm/models"
)

func main() {
    db, _ := sql.Open("postgres", connectionString)
    ctx := dbcontext.NewEnhancedDbContextWithDB(db)
    
    // Run migrations - automatically uses PostgreSQL schemas
    migrator := migrations.NewAutoMigrator(ctx, db)
    err := migrator.MigrateModels(&models.User{}, &models.Product{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Tables created with PostgreSQL-optimized schema
}
```

### 5. Environment Configuration

#### Production Environment Variables
```bash
# .env file
DATABASE_URL=postgres://username:password@hostname:5432/database_name?sslmode=require

# Or individual components
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gra_demo
DB_SSLMODE=disable
```

#### Load Environment in Go
```go
import "os"

func getDatabaseURL() string {
    if url := os.Getenv("DATABASE_URL"); url != "" {
        return url
    }
    
    host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "postgres")
    password := getEnv("DB_PASSWORD", "postgres")
    dbname := getEnv("DB_NAME", "gra_demo")
    sslmode := getEnv("DB_SSLMODE", "disable")
    
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 6. Testing

#### Run the Demo
```bash
cd examples/enhanced-orm-demo

# With PostgreSQL available
go run main.go

# Expected output:
# ✅ Connected to PostgreSQL database
# 📦 Running Migrations (PostgreSQL)...
# ✅ All operations complete
```

#### Test Connection
```go
func testConnection() {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        log.Fatal("Failed to open:", err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping:", err)
    }
    
    fmt.Println("✅ PostgreSQL connection successful")
}
```

### 7. Troubleshooting

#### Common Issues

**Connection Refused**
```
pq: connection refused
```
- Ensure PostgreSQL is running: `brew services status postgresql`
- Check port: `lsof -i :5432`

**Authentication Failed**
```
pq: password authentication failed
```
- Verify username/password
- Check `pg_hba.conf` authentication settings
- Try connecting with `psql` first

**Database Does Not Exist**
```
pq: database "gra_demo" does not exist
```
- Create database: `createdb gra_demo`
- Or connect to `postgres` database initially

**SSL Issues**
```
pq: SSL is not enabled on the server
```
- Add `sslmode=disable` to connection string for local development
- Use `sslmode=require` for production

### 8. Performance Optimization

#### Connection Pooling
```go
import "time"

func setupDatabase() *sql.DB {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        log.Fatal(err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db
}
```

#### Index Management
```sql
-- Add indexes for frequently queried columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 9. Migration from SQLite

#### Automatic Migration
The framework automatically detects the database type and adjusts:
- ✅ **Placeholders**: `?` → `$1, $2, $3`
- ✅ **Schema queries**: `PRAGMA table_info()` → `information_schema.columns`
- ✅ **Data types**: Automatically mapped by PostgreSQL driver

#### Data Migration
```go
// Export from SQLite
func exportFromSQLite() []User {
    sqliteDB, _ := sql.Open("sqlite3", "./old.db")
    ctx := dbcontext.NewEnhancedDbContextWithDB(sqliteDB)
    users := dbcontext.NewEnhancedDbSet[User](ctx)
    return users.ToList()
}

// Import to PostgreSQL
func importToPostgreSQL(users []User) {
    pgDB, _ := sql.Open("postgres", pgConnectionString)
    ctx := dbcontext.NewEnhancedDbContextWithDB(pgDB)
    
    for _, user := range users {
        ctx.Add(&user)
    }
    ctx.SaveChanges()
}
```

### 10. Production Deployment

#### Environment Setup
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/gra_prod
    depends_on:
      - postgres
      
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: gra_prod
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
volumes:
  postgres_data:
```

#### Health Checks
```go
func healthCheck(db *sql.DB) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return db.PingContext(ctx)
}
```

---

## Summary

The GRA Framework now provides seamless PostgreSQL support with:
- ✅ Automatic driver detection
- ✅ Zero configuration required
- ✅ Full backward compatibility
- ✅ Production-ready performance
- ✅ Comprehensive migration support

Simply change your connection string from SQLite to PostgreSQL, and everything works automatically!

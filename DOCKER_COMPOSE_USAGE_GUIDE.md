# Docker Compose Usage Guide

This guide explains how to use the Docker Compose files in the GRA project effectively.

## File Overview

### Core Files (Use these)
- **`docker-compose.yml`** - Main development environment
- **`docker-compose.override.yml`** - Additional development services
- **`docker-compose.test.yml`** - Testing environment
- **`docker-compose.sonar.yml`** - Code quality analysis

### ~~Legacy Files~~ (Removed)
- ~~`tools/ef-migrate/docker-compose.yml`~~ - Replaced by main compose files

## Usage Patterns

### 1. Basic Development
```bash
# Start main development stack (PostgreSQL only)
docker-compose up -d

# Start with additional dev tools (pgAdmin, Redis)
docker-compose --profile dev up -d

# Start everything including ef-migrate tool
docker-compose --profile full up -d
```

### 2. Migration Development
```bash
# Start with migration tools
docker-compose --profile migration up -d

# Or start everything
docker-compose --profile full up -d
```

### 3. Testing
```bash
# Start test database (different port, isolated data)
docker-compose -f docker-compose.test.yml up -d

# Run tests with test database
go test ./...

# Clean up test environment
docker-compose -f docker-compose.test.yml down -v
```

### 4. Code Quality Analysis
```bash
# Start SonarQube with dedicated database
docker-compose -f docker-compose.sonar.yml up -d

# Access SonarQube at http://localhost:9000
# Run sonar-scanner to analyze code

# Clean up
docker-compose -f docker-compose.sonar.yml down -v
```

## Service Details

### Development Services (`docker-compose.yml`)
| Service | Port | Purpose | Profile |
|---------|------|---------|---------|
| postgres | 5432 | Main development database | always |
| pgadmin | 8080 | Database management UI | dev, full |
| redis | 6379 | Caching (optional) | dev, full |

### Additional Services (`docker-compose.override.yml`)
| Service | Purpose | Profile |
|---------|---------|---------|
| ef-migrate | Migration tool | migration, full |

### Test Services (`docker-compose.test.yml`)
| Service | Port | Purpose |
|---------|------|---------|
| postgres-test | 5433 | Isolated test database |

### SonarQube Services (`docker-compose.sonar.yml`)
| Service | Port | Purpose |
|---------|------|---------|
| sonarqube | 9000 | Code quality analysis |
| db | - | SonarQube database |

## Profile-Based Service Selection

### Available Profiles
- **`dev`** - Development tools (pgAdmin, Redis)
- **`migration`** - Migration development (ef-migrate)
- **`full`** - Everything (dev + migration)

### Profile Usage Examples
```bash
# Minimal (PostgreSQL only)
docker-compose up -d

# Development (PostgreSQL + pgAdmin + Redis)
docker-compose --profile dev up -d

# Migration work (PostgreSQL + ef-migrate tool)
docker-compose --profile migration up -d

# Everything (PostgreSQL + pgAdmin + Redis + ef-migrate)
docker-compose --profile full up -d
```

## Database Connections

### Development Database
```bash
Host: localhost
Port: 5432
Database: gra_dev
Username: postgres
Password: postgres
```

### Test Database
```bash
Host: localhost
Port: 5433
Database: gra_test
Username: gra_user
Password: gra_password
```

### pgAdmin Access
```bash
URL: http://localhost:8080
Email: admin@gra.dev
Password: admin
```

## Best Practices

### 1. Environment Isolation
- **Never mix test and development databases**
- Use separate compose files for different purposes
- Always use `-f` flag for non-default compose files

### 2. Resource Management
```bash
# Start only what you need
docker-compose up -d postgres  # Just database

# Clean up when done
docker-compose down  # Stop containers
docker-compose down -v  # Stop and remove volumes
```

### 3. Port Management
- Development: Standard ports (5432, 8080, 6379)
- Testing: Alternative ports (5433)
- SonarQube: Dedicated ports (9000)

### 4. Data Persistence
- Development data persists in named volumes
- Test data can be ephemeral (`docker-compose down -v`)
- Use volumes for important data, tmpfs for temporary data

## Troubleshooting

### Port Conflicts
```bash
# Check what's using a port
lsof -i :5432

# Use different compose file with different ports
docker-compose -f docker-compose.test.yml up -d
```

### Database Connection Issues
```bash
# Check container status
docker-compose ps

# Check logs
docker-compose logs postgres

# Test database connectivity
docker-compose exec postgres psql -U postgres -d gra_dev
```

### Clean Slate
```bash
# Stop everything and remove all data
docker-compose down -v
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.sonar.yml down -v

# Remove all GRA-related Docker resources
docker system prune -f
```

## Summary

This multi-file approach provides:
- ✅ **Clear separation of concerns**
- ✅ **Fast, targeted startup**
- ✅ **No port conflicts**
- ✅ **Environment isolation**
- ✅ **Flexible development workflows**
- ✅ **CI/CD friendly**

Instead of one monolithic file, you have specialized tools for each purpose, making development more efficient and less error-prone.

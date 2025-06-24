# Docker Compose Configuration Guide

This project uses multiple Docker Compose files for different purposes. Here's how to use them:

## 🚀 Quick Start Commands

### Development (Main Database Only)
```bash
# Start main PostgreSQL database
docker-compose up -d

# Start with pgAdmin for database management  
docker-compose --profile dev up -d

# Start everything including Redis
docker-compose --profile full up -d
```

### Testing
```bash
# Start test database for automated testing
docker-compose -f docker-compose.test.yml up -d

# Run tests with PostgreSQL integration
go test -cover ./tools/migration/test
```

### Code Quality Analysis
```bash
# Start SonarQube for code analysis
docker-compose -f docker-compose.sonar.yml up -d

# Access SonarQube at http://localhost:9000
```

### EF Migration Tool
```bash
# Start with migration tool
docker-compose --profile migration up -d

# Or use the specific ef-migrate compose
docker-compose -f tools/ef-migrate/docker-compose.yml up -d
```

## 📁 File Structure

```
├── docker-compose.yml              # Main development environment
├── docker-compose.override.yml     # Additional development services
├── docker-compose.test.yml         # Testing environment
├── docker-compose.sonar.yml        # SonarQube code analysis
└── tools/ef-migrate/
    └── docker-compose.yml          # EF Migration tool environment
```

## 🎯 Purpose of Each Configuration

### 1. **Main Development** (`docker-compose.yml`)
- **Services**: PostgreSQL (port 5432), optional pgAdmin & Redis
- **Database**: `gra_dev` with `postgres:postgres`
- **Usage**: Primary development database
- **Profiles**: 
  - Default: PostgreSQL only
  - `dev`: + pgAdmin
  - `full`: + pgAdmin + Redis

### 2. **Testing** (`docker-compose.test.yml`)
- **Services**: PostgreSQL test database (port 5433)
- **Database**: `gra_test` with `gra_user:gra_password`
- **Usage**: Automated testing, CI/CD
- **Isolation**: Separate from development data

### 3. **Code Quality** (`docker-compose.sonar.yml`)
- **Services**: SonarQube + PostgreSQL (internal)
- **Usage**: Code quality analysis
- **Access**: http://localhost:9000
- **Isolation**: Own network and database

### 4. **EF Migration** (`tools/ef-migrate/docker-compose.yml`)
- **Services**: Migration tool + full dev stack
- **Usage**: Database migration management
- **Features**: Built-in migration tool

## 🔧 Service Details

### PostgreSQL Instances
| Environment | Port | Database | User | Password | Purpose |
|-------------|------|----------|------|----------|---------|
| Development | 5432 | gra_dev | postgres | postgres | Main development |
| Testing | 5433 | gra_test | gra_user | gra_password | Automated testing |
| SonarQube | internal | sonar | sonar | sonar | Code analysis |

### Additional Services
| Service | Port | Purpose | Profile |
|---------|------|---------|---------|
| pgAdmin | 8080 | Database management | dev, full |
| Redis | 6379 | Caching | full |
| SonarQube | 9000 | Code analysis | sonar compose |

## 💡 Best Practices

### For Development
```bash
# Daily development
docker-compose up -d

# With database management UI
docker-compose --profile dev up -d

# Full stack development
docker-compose --profile full up -d
```

### For Testing
```bash
# Before running tests
docker-compose -f docker-compose.test.yml up -d

# Run tests
go test -cover ./...

# Cleanup after testing
docker-compose -f docker-compose.test.yml down
```

### For Code Analysis
```bash
# Start SonarQube
docker-compose -f docker-compose.sonar.yml up -d

# Run analysis (after SonarQube is ready)
./scripts/sonar-scan.sh

# Stop when done
docker-compose -f docker-compose.sonar.yml down
```

## 🚦 Health Checks

All databases include health checks:
- **PostgreSQL**: `pg_isready` command
- **Redis**: `redis-cli ping`
- **Intervals**: 10s with 5s timeout, 5 retries

## 🔒 Security Notes

1. **Different Networks**: Each environment uses isolated networks
2. **Different Credentials**: Test vs dev vs analysis environments
3. **Port Separation**: No conflicts between environments
4. **Data Isolation**: Separate volumes for each purpose

## 🛠️ Troubleshooting

### Port Conflicts
If you get port conflicts, check what's running:
```bash
docker ps
lsof -i :5432
lsof -i :5433
```

### Database Connection Issues
```bash
# Check container health
docker-compose ps

# View logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U postgres
```

### Clean Start
```bash
# Stop all and remove volumes
docker-compose -f docker-compose.yml -f docker-compose.test.yml -f docker-compose.sonar.yml down -v

# Remove all project containers
docker container prune -f
```

## 📊 Resource Usage

| Configuration | RAM Usage | Startup Time | Services |
|---------------|-----------|--------------|----------|
| Main (minimal) | ~200MB | ~10s | PostgreSQL only |
| Dev profile | ~300MB | ~15s | + pgAdmin |
| Full profile | ~350MB | ~20s | + Redis |
| Testing | ~200MB | ~10s | Test PostgreSQL |
| SonarQube | ~2GB | ~60s | SonarQube + DB |

Choose the configuration that matches your current task to optimize resource usage.

# SonarQube Integration Guide for GRA

This guide explains how to set up and use SonarQube for code quality analysis in the GRA project.

## Prerequisites

1. **Docker and Docker Compose** - For running SonarQube locally
2. **SonarQube Scanner** - For code analysis
3. **Go 1.21+** - For running tests and building the project

## Installation

### 1. Install SonarQube Scanner

```bash
# macOS (using Homebrew)
brew install sonar-scanner

# Or download from official site
# https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/
```

### 2. Install Docker (if not already installed)

```bash
# macOS
brew install --cask docker

# Or download Docker Desktop from https://www.docker.com/products/docker-desktop
```

## Quick Start

### 1. Start SonarQube Server

```bash
make sonar-start
```

This will start SonarQube at `http://localhost:9000` using Docker Compose.

### 2. Initial Setup

1. Open `http://localhost:9000` in your browser
2. Login with default credentials: `admin` / `admin`
3. Change the admin password when prompted
4. Create a new project:
   - Project Key: `gra-migration-system`
   - Display Name: `GRA - Go Rapid API with Migration System`
5. Generate a project token and save it securely

### 3. Configure Environment

```bash
# Add to your ~/.zshrc or ~/.bash_profile
export SONAR_TOKEN="your_project_token_here"
export SONAR_HOST_URL="http://localhost:9000"

# Reload your shell
source ~/.zshrc
```

### 4. Run Analysis

```bash
make sonar-analyze
```

This will:
- Run all tests with coverage
- Generate coverage reports
- Execute SonarQube analysis
- Upload results to the SonarQube server

## Available Commands

| Command | Description |
|---------|-------------|
| `make sonar-start` | Start SonarQube server with Docker |
| `make sonar-stop` | Stop SonarQube server |
| `make sonar-analyze` | Run complete analysis (tests + SonarQube) |
| `make sonar-clean` | Clean SonarQube data and volumes |

## Manual Analysis

You can also run the analysis script directly:

```bash
./scripts/run-sonar.sh
```

## CI/CD Integration

### GitHub Actions

The project includes a GitHub Actions workflow (`.github/workflows/sonarqube.yml`) that:

- Runs on push to main branches and pull requests
- Executes tests with coverage
- Performs SonarQube analysis
- Uploads results to SonarQube server

### Required Secrets

Add these secrets to your GitHub repository:

- `SONAR_TOKEN` - Your SonarQube project token
- `SONAR_HOST_URL` - Your SonarQube server URL

## Configuration Files

### sonar-project.properties

Main configuration file with project settings:
- Project identification
- Source and test file patterns
- Coverage report paths
- Exclusions for non-source files

### docker-compose.sonar.yml

Docker Compose configuration for running SonarQube with PostgreSQL database.

## Quality Gates

SonarQube will analyze your code for:

- **Bugs** - Logic errors and potential runtime issues
- **Vulnerabilities** - Security issues
- **Code Smells** - Maintainability issues
- **Coverage** - Test coverage percentage
- **Duplication** - Code duplication detection

## Viewing Results

After analysis:

1. Open `http://localhost:9000`
2. Navigate to your project dashboard
3. Review issues and metrics
4. Use filters to focus on specific issue types
5. View coverage reports and hotspots

## Troubleshooting

### Common Issues

1. **sonar-scanner not found**
   ```bash
   brew install sonar-scanner
   ```

2. **SONAR_TOKEN not set**
   ```bash
   export SONAR_TOKEN="your_token"
   ```

3. **SonarQube not accessible**
   - Check if Docker is running
   - Wait for SonarQube to fully initialize (2-3 minutes)
   - Check `docker-compose logs` for errors

4. **Analysis fails**
   - Ensure tests pass: `make test`
   - Check network connectivity to SonarQube server
   - Verify project key matches SonarQube configuration

### Logs and Debugging

```bash
# Check SonarQube logs
docker-compose -f docker-compose.sonar.yml logs -f

# Check Docker containers
docker-compose -f docker-compose.sonar.yml ps

# Test connectivity
curl http://localhost:9000/api/system/status
```

## Best Practices

1. **Run analysis regularly** - Integrate into your development workflow
2. **Fix issues incrementally** - Address new issues as they arise
3. **Set quality gates** - Define minimum quality standards
4. **Monitor trends** - Track quality metrics over time
5. **Review security hotspots** - Pay special attention to security issues

## Project-Specific Exclusions

The configuration excludes:
- Test files from main analysis
- Documentation files (*.md)
- Generated files (migrations/*.sql)
- Dependencies (vendor/, node_modules/)
- Build artifacts (bin/, debug/)

## Integration with IDE

Many IDEs support SonarQube integration:
- **VS Code**: SonarLint extension
- **GoLand**: SonarLint plugin
- **Vim/Neovim**: Various SonarQube plugins

## Support

For issues with SonarQube integration:
1. Check this guide first
2. Review SonarQube logs
3. Consult [SonarQube documentation](https://docs.sonarqube.org/)
4. Open an issue in the project repository

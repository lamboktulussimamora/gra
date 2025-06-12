# SonarQube Commands Reference for GRA Project

This document provides a comprehensive reference for all SonarQube commands used in the GRA (Go Rapid API) migration system project.

## Table of Contents
- [SonarQube Server Management](#sonarqube-server-management)
- [Authentication & Configuration](#authentication--configuration)
- [Project Management](#project-management)
- [Code Analysis](#code-analysis)
- [Quality Gate Management](#quality-gate-management)
- [Reporting & Monitoring](#reporting--monitoring)
- [Troubleshooting](#troubleshooting)
- [Integration with CI/CD](#integration-with-cicd)

## SonarQube Server Management

### Starting SonarQube Server
```bash
# Start SonarQube using Docker Compose
make sonar-start

# Alternative: Start manually using Docker
docker run -d --name sonarqube \
  -p 9000:9000 \
  -v sonarqube_data:/opt/sonarqube/data \
  -v sonarqube_logs:/opt/sonarqube/logs \
  -v sonarqube_extensions:/opt/sonarqube/extensions \
  sonarqube:community

# Check if SonarQube is running
curl -s http://localhost:9000/api/system/status
```

### Stopping SonarQube Server
```bash
# Stop SonarQube using make
make sonar-stop

# Alternative: Stop manually
docker stop sonarqube
docker rm sonarqube
```

### Server Status and Health Check
```bash
# Check server status
curl -s http://localhost:9000/api/system/status | jq '.'

# Check server health
curl -s http://localhost:9000/api/system/health | jq '.'

# Get server version
curl -s http://localhost:9000/api/server/version
```

## Authentication & Configuration

### Initial Setup (Default Credentials)
- **Default Username**: `admin`
- **Default Password**: `admin`
- **Updated Password**: `MyPassword_123`

### Change Admin Password
```bash
# Change password via API
curl -X POST \
  -u admin:admin \
  "http://localhost:9000/api/users/change_password" \
  -d "login=admin&previousPassword=admin&password=MyPassword_123"
```

### Generate Authentication Token
```bash
# Generate project-specific token
curl -X POST \
  -u admin:MyPassword_123 \
  "http://localhost:9000/api/user_tokens/generate" \
  -d "name=gra-migration-system-token&type=PROJECT_ANALYSIS_TOKEN"

# Example response token: sqp_4f609734abcfd2163ca877c6d27ade784c890347
```

### Export Token for Environment
```bash
# Set token as environment variable
export SONAR_TOKEN="sqp_4f609734abcfd2163ca877c6d27ade784c890347"

# Verify token is set
echo $SONAR_TOKEN
```

## Project Management

### Create New Project
```bash
# Create project via API
curl -X POST \
  -u admin:MyPassword_123 \
  "http://localhost:9000/api/projects/create" \
  -d "project=gra-migration-system&name=GRA Migration System"
```

### List Projects
```bash
# List all projects
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/projects/search" | jq '.components[] | {key: .key, name: .name}'

# Get specific project info
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/projects/search?projects=gra-migration-system" | jq '.'
```

### Delete Project
```bash
# Delete project (use with caution)
curl -X POST \
  -u admin:MyPassword_123 \
  "http://localhost:9000/api/projects/delete" \
  -d "project=gra-migration-system"
```

## Code Analysis

### Run Complete Analysis with Coverage
```bash
# Full analysis pipeline
make test-coverage && make sonar-analyze

# Manual analysis with all parameters
sonar-scanner \
  -Dsonar.projectKey=gra-migration-system \
  -Dsonar.sources=. \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=$SONAR_TOKEN \
  -Dsonar.go.coverage.reportPaths=coverage.out \
  -Dsonar.exclusions="**/*_test.go,**/vendor/**,**/node_modules/**"
```

### Analysis Configuration (sonar-project.properties)
```properties
# Project identification
sonar.projectKey=gra-migration-system
sonar.projectName=GRA Migration System
sonar.projectVersion=1.0

# Source code settings
sonar.sources=.
sonar.language=go
sonar.sourceEncoding=UTF-8

# Coverage settings
sonar.go.coverage.reportPaths=coverage.out

# Exclusions
sonar.exclusions=**/*_test.go,**/vendor/**,**/node_modules/**,**/*.pb.go

# Test settings
sonar.tests=.
sonar.test.inclusions=**/*_test.go
sonar.test.exclusions=**/vendor/**
```

### Quick Analysis (Without Tests)
```bash
# Quick scan for immediate feedback
sonar-scanner \
  -Dsonar.projectKey=gra-migration-system \
  -Dsonar.sources=. \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.token=$SONAR_TOKEN
```

## Quality Gate Management

### Check Quality Gate Status
```bash
# Get current quality gate status
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | jq '.'

# Check specific conditions
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | \
  jq '.projectStatus.conditions[] | {status: .status, metric: .metricKey, threshold: .errorThreshold, actual: .actualValue}'
```

### Quality Gate Results Interpretation
```bash
# Parse quality gate results
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | \
  jq -r '.projectStatus | "Status: \(.status)", "Conditions:", (.conditions[] | "  \(.metricKey): \(.status) (threshold: \(.errorThreshold), actual: \(.actualValue))")'
```

### List Available Quality Gates
```bash
# Get all quality gates
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/list" | jq '.qualitygates[] | {id: .id, name: .name, isDefault: .isDefault}'
```

## Reporting & Monitoring

### Get Project Issues
```bash
# Get all open issues
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/issues/search?componentKeys=gra-migration-system&statuses=OPEN" | \
  jq '.issues[] | {key: .key, rule: .rule, severity: .severity, message: .message, line: .line, component: .component}'

# Get issues by severity
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/issues/search?componentKeys=gra-migration-system&severities=BLOCKER,CRITICAL" | \
  jq '.issues[] | {severity: .severity, message: .message, file: .component, line: .line}'

# Get issues by type
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/issues/search?componentKeys=gra-migration-system&types=BUG,VULNERABILITY,CODE_SMELL" | \
  jq '.issues[] | {type: .type, severity: .severity, message: .message}'
```

### Get Security Hotspots
```bash
# Get security hotspots requiring review
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/hotspots/search?projectKey=gra-migration-system&statuses=TO_REVIEW" | \
  jq '.hotspots[] | {key: .key, message: .message, file: .component, line: .line, status: .status}'

# Get security hotspot details
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/hotspots/show?hotspot=HOTSPOT_KEY" | jq '.'
```

### Get Project Metrics
```bash
# Get comprehensive project metrics
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/measures/component?component=gra-migration-system&metricKeys=ncloc,complexity,coverage,duplicated_lines_density,vulnerabilities,bugs,code_smells" | \
  jq '.component.measures[] | {metric: .metric, value: .value}'

# Get coverage metrics specifically
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/measures/component?component=gra-migration-system&metricKeys=coverage,line_coverage,branch_coverage,uncovered_lines" | \
  jq '.component.measures[] | {metric: .metric, value: .value}'
```

### Analysis History
```bash
# Get analysis history
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/project_analyses/search?project=gra-migration-system" | \
  jq '.analyses[] | {date: .date, projectVersion: .projectVersion, buildString: .buildString}'
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Connection Issues
```bash
# Test SonarQube connectivity
curl -v http://localhost:9000/api/system/ping

# Check if SonarQube is ready
timeout 300 bash -c 'until curl -s http://localhost:9000/api/system/status | grep -q "UP"; do echo "Waiting for SonarQube..."; sleep 5; done'
```

#### 2. Authentication Issues
```bash
# Verify token validity
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/authentication/validate" | jq '.'

# List user tokens
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/user_tokens/search" | jq '.'
```

#### 3. Analysis Issues
```bash
# Check scanner logs
sonar-scanner -X  # Enable debug mode

# Verify project exists
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/components/show?component=gra-migration-system"
```

#### 4. Clean and Reset
```bash
# Clean SonarQube data (destructive operation)
make sonar-clean

# Remove analysis cache
rm -rf .scannerwork/

# Reset project data
curl -X POST -u admin:MyPassword_123 \
  "http://localhost:9000/api/projects/delete" \
  -d "project=gra-migration-system"
```

## Integration with CI/CD

### GitHub Actions Integration
```yaml
# .github/workflows/sonar.yml
- name: SonarQube Analysis
  run: |
    sonar-scanner \
      -Dsonar.projectKey=gra-migration-system \
      -Dsonar.sources=. \
      -Dsonar.host.url=${{ secrets.SONAR_HOST_URL }} \
      -Dsonar.token=${{ secrets.SONAR_TOKEN }} \
      -Dsonar.go.coverage.reportPaths=coverage.out
```

### Quality Gate in CI
```bash
# Wait for quality gate result in CI
curl -s -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | \
  jq -e '.projectStatus.status == "OK"' || exit 1
```

### Make Targets for Automation
```bash
# Available make targets
make sonar-start      # Start SonarQube server
make sonar-stop       # Stop SonarQube server
make sonar-analyze    # Run SonarQube analysis
make sonar-clean      # Clean SonarQube data
make sonar-status     # Check SonarQube status
```

## Environment Variables

### Required Environment Variables
```bash
export SONAR_HOST_URL="http://localhost:9000"
export SONAR_TOKEN="sqp_4f609734abcfd2163ca877c6d27ade784c890347"
export SONAR_PROJECT_KEY="gra-migration-system"
```

### Optional Configuration
```bash
export SONAR_SCANNER_OPTS="-Xmx512m"
export SONAR_LOG_LEVEL="INFO"  # DEBUG, INFO, WARN, ERROR
```

## Quality Metrics Explained

### Coverage Metrics
- `coverage`: Overall test coverage percentage
- `line_coverage`: Line coverage percentage
- `branch_coverage`: Branch coverage percentage
- `new_coverage`: Coverage on new code

### Security Metrics
- `vulnerabilities`: Number of security vulnerabilities
- `security_hotspots`: Number of security hotspots
- `security_rating`: Security rating (A-E)

### Maintainability Metrics
- `code_smells`: Number of maintainability issues
- `technical_debt`: Technical debt in minutes
- `sqale_rating`: Maintainability rating (A-E)

### Reliability Metrics
- `bugs`: Number of bugs
- `reliability_rating`: Reliability rating (A-E)

## Advanced Usage

### Custom Quality Gate
```bash
# Create custom quality gate
curl -X POST -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/create" \
  -d "name=GRA Custom Gate"

# Add conditions to quality gate
curl -X POST -u admin:MyPassword_123 \
  "http://localhost:9000/api/qualitygates/create_condition" \
  -d "gateId=1&metric=coverage&op=LT&error=80"
```

### Exclude Files from Analysis
```bash
# In sonar-project.properties
sonar.exclusions=**/*_test.go,**/vendor/**,**/migrations/**
sonar.coverage.exclusions=**/*_test.go,**/mocks/**
sonar.cpd.exclusions=**/*_generated.go
```

This comprehensive guide covers all SonarQube commands and operations used in the GRA project. For the most up-to-date information, refer to the [SonarQube API documentation](https://docs.sonarqube.org/latest/extend/web-api/).

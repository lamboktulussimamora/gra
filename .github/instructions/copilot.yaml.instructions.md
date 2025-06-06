# GRA Framework - Project Standards and Guidelines for AI Assistants
# This file provides comprehensive instructions for GitHub Copilot and other AI assistants
# to maintain consistency and quality when generating code for the GRA project.

instructions:
  # ========================
  # PROJECT OVERVIEW
  # ========================
  - "GRA is a lightweight HTTP framework for Go, inspired by Gin, with Entity Framework Core-like ORM capabilities."
  - "The project follows Clean Architecture principles, emphasizing separation of concerns and testability."
  - "Core features: HTTP routing, JWT auth, middleware, API versioning, caching, ORM with migrations, LINQ-style querying."
  - "Target: High-performance web applications with database operations, following Go best practices."

  # ========================
  # CODING STANDARDS
  # ========================
  - "Use idiomatic Go code following official Go guidelines and effective Go practices."
  - "All exported functions and types MUST have GoDoc comments starting with the function/type name."
  - "Use context.Context in all exported functions for cancellation and timeouts."
  - "Prefer interfaces over concrete types for better testability and flexibility."
  - "Use dependency injection instead of global variables or singletons."
  - "Error handling: Use error wrapping with fmt.Errorf or errors package for context."
  - "Avoid magic numbers and strings; define named constants with descriptive names."
  - "Use struct tags for JSON, validation, and database mapping consistently."
  - "Follow Go naming conventions: PascalCase for exported, camelCase for unexported."

  # ========================
  # PROJECT STRUCTURE (Clean Architecture)
  # ========================
  - "Project follows Clean Architecture with domain-driven folder structure:"
  
  # Core Infrastructure
  - "adapter/ : Database adapters and external system interfaces. Implements repository patterns and database-specific logic. Example: adapter/postgres_adapter.go with connection pooling and query optimization."
  - "cache/ : HTTP response caching middleware and strategies. Supports memory, Redis, and custom cache backends. Example: cache/memory_cache.go with TTL and LRU eviction policies."
  - "context/ : HTTP request context management, extending Go's context.Context. Includes request parsing, response helpers, and middleware context. Example: context/context.go with BindJSON, Success, Error methods."
  - "router/ : HTTP routing engine with path parameters, route groups, and middleware support. Example: router/router.go with GET, POST, PUT, DELETE methods and route registration."
  
  # Authentication & Security
  - "jwt/ : JWT token generation, validation, and middleware. Supports multiple signing algorithms and token refresh. Example: jwt/service.go with GenerateToken, ValidateToken methods."
  - "middleware/ : HTTP middleware for authentication, logging, CORS, security headers, rate limiting. Example: middleware/auth.go, middleware/cors.go with configurable options."
  - "validator/ : Input validation using struct tags and custom validators. Example: validator/validator.go with email, phone, custom business rule validation."
  
  # Data Layer
  - "orm/ : Entity Framework Core-like ORM with change tracking, LINQ-style queries, and automatic migrations."
  - "orm/dbcontext/ : Database context for connection management and transaction handling."
  - "orm/models/ : Entity models with struct tags for database mapping and validation."
  - "orm/migrations/ : Migration engine for schema changes and data transformations."
  - "orm/schema/ : Database schema inspection and comparison tools."
  - "migrations/ : SQL migration files with timestamp-based naming (YYYYMMDDHHMMSS_Description.sql)."
  
  # Application Layer
  - "versioning/ : API versioning strategies (path, query, header, media-type). Example: versioning/path_strategy.go for /v1/users, /v2/users routing."
  - "logger/ : Structured logging with configurable levels and outputs. Use this package for ALL logging - do not use log package or external loggers."
  
  # Development & Tooling
  - "examples/ : Complete example applications demonstrating framework features."
  - "examples/basic/ : Simple CRUD API example"
  - "examples/comprehensive-orm-demo/ : Advanced ORM features showcase"
  - "examples/auth-and-security/ : Authentication and security patterns"
  - "tools/ : CLI tools and code generators. Example: tools/migration-generator for creating new migrations."
  - "scripts/ : Build, test, and deployment automation scripts. Example: scripts/coverage.sh for test coverage reports."
  - "debug/ : Development and debugging utilities, migration test harnesses."
  - "docs/ : Documentation website, API references, guides, and tutorials."

  # ========================
  # DATABASE STANDARDS
  # ========================
  - "Support multiple databases: PostgreSQL (primary), SQLite (testing), MySQL (optional)."
  - "Use connection pooling with configurable max connections, idle timeout, and connection lifetime."
  - "Database migrations must be reversible with both Up and Down operations."
  - "Migration naming: timestamp_PascalCaseDescription.sql (e.g., 20240606120000_CreateUsersTable.sql)."
  - "Use prepared statements and parameterized queries to prevent SQL injection."
  - "Database models should use struct tags: json field_name db field_name validate required"
  - "Implement repository pattern in adapter/ for database operations with interfaces."
  - "Use transactions for multi-step operations with proper rollback handling."
  - "Database connection strings should be configurable via environment variables."

  # ========================
  # SECURITY STANDARDS
  # ========================
  - "JWT tokens: Use RS256 or HS256 algorithms with configurable expiration times."
  - "Implement CORS middleware with configurable origins, methods, and headers."
  - "Security headers: X-XSS-Protection, X-Content-Type-Options, X-Frame-Options, HSTS, CSP."
  - "Input validation: Validate ALL user inputs using the validator package."
  - "Rate limiting: Implement configurable rate limiting for API endpoints."
  - "Password hashing: Use bcrypt with cost factor 12 or higher."
  - "Secure session management with HttpOnly, Secure, and SameSite cookie attributes."
  - "Environment-based configuration for secrets (never hardcode sensitive data)."
  - "SQL injection prevention through prepared statements and ORM usage."
  - "HTTPS enforcement in production environments."

  # ========================
  # TESTING STANDARDS
  # ========================
  - "Maintain minimum 80% test coverage across all packages."
  - "Use table-driven tests for comprehensive scenario coverage."
  - "Test files: *_test.go in the same package as the code being tested."
  - "Unit tests: Test individual functions and methods in isolation."
  - "Integration tests: Test database operations, HTTP endpoints, and middleware."
  - "Mock external dependencies using interfaces and dependency injection."
  - "Benchmark tests: Include performance benchmarks for critical paths (router, ORM)."
  - "Test naming: TestFunctionName_Scenario_ExpectedResult."
  - "Use testify/assert for assertions and testify/mock for mocking."
  - "Setup and teardown: Use setup/teardown functions for test data and cleanup."

  # ========================
  # CODE QUALITY REQUIREMENTS - MANDATORY
  # ========================
  
  # ALL CODE MUST PASS THESE QUALITY GATES BEFORE COMMIT
  - "ZERO tolerance policy: All code MUST pass both golangci-lint and SonarQube without ANY issues."
  - "Run 'make verify' before every commit - this includes golangci-lint, go vet, and go fmt checks."
  - "Use 'make sonar-analyze' to validate SonarQube compliance before pull requests."
  - "Code coverage MUST be maintained above 70% for new features and bug fixes."
  - "Cyclomatic complexity MUST NOT exceed 15 for any function (enforced by gocyclo)."
  
  # SonarQube Quality Requirements
  - "SonarQube Quality Gate MUST pass with A rating for:"
  - "  - Reliability: No bugs allowed in production code"
  - "  - Security: No security hotspots or vulnerabilities"
  - "  - Maintainability: Technical debt ratio < 5%"
  - "  - Coverage: Minimum 70% test coverage on new code"
  - "  - Duplications: No code duplications > 3% density"
  - "Fix ALL SonarQube issues before merge - use sonar.exclusions only for generated code."
  
  # golangci-lint Mandatory Rules
  - "ALL enabled linters MUST pass without warnings or errors:"
  - "  - errcheck: Check ALL error returns - no ignored errors allowed"
  - "  - gosec: Security issues MUST be resolved or properly justified"
  - "  - govet: All go vet warnings MUST be fixed"
  - "  - staticcheck: Static analysis issues MUST be resolved"
  - "  - gocyclo: Functions with complexity > 15 MUST be refactored"
  - "  - ineffassign: Remove all ineffective assignments"
  - "  - unused: Remove all unused code, variables, and imports"
  - "  - misspell: Fix ALL spelling mistakes in comments and strings"
  - "  - goconst: Extract repeated strings (≥3 occurrences) to constants"
  - "  - revive: Follow Go style guidelines strictly"
  
  # Pre-commit Quality Checklist
  - "Before every commit, run these commands and ensure ALL pass:"
  - "  1. make verify    # Runs gofmt, go vet, and golangci-lint"
  - "  2. make test      # All tests must pass"
  - "  3. make coverage  # Verify coverage meets minimum threshold"
  - "  4. make sonar-analyze  # SonarQube analysis (if SONAR_TOKEN available)"
  
  # Quality Standards for New Code
  - "New functions MUST have:"
  - "  - Comprehensive GoDoc comments with examples for exported functions"
  - "  - Input validation with proper error messages"
  - "  - Error handling with context using fmt.Errorf or errors.Wrap"
  - "  - Unit tests with minimum 80% coverage"
  - "  - Integration tests for database operations"
  - "  - Benchmark tests for performance-critical code"
  
  # Code Review Quality Gates
  - "Pull requests MUST include:"
  - "  - Evidence of successful 'make verify' execution"
  - "  - SonarQube quality gate status: PASSED"
  - "  - Test coverage report showing no decrease in overall coverage"
  - "  - Performance impact analysis for critical paths"
  - "  - Security impact assessment for authentication/authorization changes"

  # ========================
  # TOOLS AND USAGE
  # ========================
  - "golangci-lint: Run before commits with timeout=10m. Fix ALL linting errors."
  - "go mod tidy: Keep dependencies clean and up-to-date."
  - "gofmt/goimports: Format code automatically before commits."
  - "scripts/coverage.sh: Generate coverage reports and enforce minimum thresholds."
  - "Makefile: Use for common development tasks (build, test, lint, deploy)."
  - "Docker: Support containerized development and deployment."
  - "GitHub Actions: Automated testing, coverage reporting, and quality gates."
  - "SonarQube: Code quality analysis and security vulnerability scanning."

  # ========================
  # HOW TO RUN COMMANDS
  # ========================
  
  # Running Tests
  - "Run all tests: go test ./..."
  - "Run tests with coverage: go test -v -race -coverprofile=coverage.out ./..."
  - "Generate HTML coverage report: go tool cover -html=coverage.out -o coverage.html"
  - "Run specific package tests: go test ./router -v"
  - "Run specific test function: go test ./router -run TestRouter_GET"
  - "Run benchmarks: go test -bench=. ./router"
  - "Use coverage script: ./scripts/coverage.sh"
  
  # Code Quality and Linting (MANDATORY BEFORE COMMIT)
  - "Full quality check: make verify  # Runs gofmt, go vet, golangci-lint"
  - "Quick lint check: golangci-lint run --timeout=10m"
  - "Fix auto-fixable issues: golangci-lint run --fix --timeout=10m"
  - "Security scan: golangci-lint run --enable gosec --timeout=10m"
  - "Performance analysis: golangci-lint run --enable prealloc,bodyclose --timeout=10m"
  - "Format code: gofmt -w . && goimports -w ."
  - "Tidy dependencies: go mod tidy && go mod verify"
  - "Check for vulnerabilities: go list -json -deps ./... | nancy sleuth"
  
  # SonarQube Analysis (MANDATORY FOR PULL REQUESTS)
  - "Start SonarQube locally: make sonar-start"
  - "Run full SonarQube analysis: make sonar-analyze  # Requires SONAR_TOKEN"
  - "Quick local scan: sonar-scanner -Dsonar.host.url=http://localhost:9000"
  - "Check quality gate: curl -u admin:admin http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system"
  - "View detailed results: open http://localhost:9000/dashboard?id=gra-migration-system"
  - "Clean SonarQube data: make sonar-clean"
  - "Stop SonarQube: make sonar-stop"
  
  # Combined Quality Pipeline (RUN BEFORE EVERY COMMIT)
  - "Complete quality check pipeline:"
  - "  make clean && make test && make coverage && make verify && make sonar-analyze"
  - "Quick pre-commit check: make pre-commit"
  - "CI/CD quality gate: make test && make coverage && make verify # + SonarQube in CI"
  
  # Development Environment Setup
  - "Setup development environment: ./scripts/setup-dev.sh"
  - "Install pre-commit hooks: ./scripts/setup-dev.sh"
  - "Check tools only: ./scripts/setup-dev.sh --check"
  - "Run comprehensive quality check: ./scripts/quality-check.sh"
  - "Quick quality check: ./scripts/quality-check.sh --quick"
  
  # Database Migrations
  - "Run all migrations: go run tools/ef-migrate/main.go up"
  - "Create new migration: go run tools/ef-migrate/main.go create CreateTableName"
  - "Rollback last migration: go run tools/ef-migrate/main.go down"
  - "Check migration status: go run tools/ef-migrate/main.go status"
  - "Reset database: go run tools/ef-migrate/main.go reset"
  - "Test migration lifecycle: ./test_complete_migration_lifecycle.sh"
  
  # Running Examples
  - "Basic example: cd examples/basic && go run main.go"
  - "ORM demo: cd examples/comprehensive-orm-demo && go run main.go"
  - "Auth example: cd examples/auth-and-security && go run main.go"
  - "Migration example: cd examples/migration-example && go run main.go"
  - "Versioning example: cd examples/versioning-and-cache && go run main.go"
  
  # Build and Development
  - "Build project: go build ."
  - "Install dependencies: go mod download"
  - "Clean build cache: go clean -cache"
  - "Build for different platforms: GOOS=linux GOARCH=amd64 go build ."
  - "Run with race detection: go run -race main.go"
  
  # Documentation
  - "Generate docs: go doc -all ."
  - "Serve docs locally: godoc -http=:6060"
  - "Check links in docs: ./scripts/check_links.sh"
  - "Deploy docs: ./scripts/generate_version_docs.sh"
  
  # Database Operations
  - "Test PostgreSQL setup: ./test_postgresql_complete.sh"
  - "Test SQLite operations: go test ./adapter -tags=sqlite"
  - "Clean test databases: rm -f *.db test_*.db"
  
  # Release and Deployment
  - "Create release: ./release.sh"
  - "Verify installation: ./verify.sh"
  - "Clean project: ./scripts/clean_project.sh"

  # ========================
  # DEVELOPMENT WORKFLOW
  # ========================
  - "Feature branches: Create feature branches from develop, merge via pull requests."
  - "Commit messages: Use conventional commits format (feat:, fix:, docs:, test:, refactor:)."
  - "Pull requests: Include description, testing notes, and breaking changes."
  - "Code review: Require approval from maintainers before merging."
  - "CI/CD: All tests must pass, coverage thresholds met, no linting errors."

  # ========================
  # QUALITY-COMPLIANT CODE EXAMPLES
  # ========================
  
  # Example: HTTP Handler (SonarQube + golangci-lint compliant)
  - "Quality-compliant handler pattern:"
  - "// GetUser retrieves a user by ID with proper validation and error handling"
  - "func GetUser(c *gra.Context) {"
  - "    userID := c.Param(\"id\")"
  - "    if userID == \"\" {"
  - "        c.Error(http.StatusBadRequest, \"missing user ID\")"
  - "        return"
  - "    }"
  - "    "
  - "    user, err := userService.GetByID(c.Request().Context(), userID)"
  - "    if err != nil {"
  - "        if errors.Is(err, ErrUserNotFound) {"
  - "            c.Error(http.StatusNotFound, \"user not found\")"
  - "            return"
  - "        }"
  - "        c.Error(http.StatusInternalServerError, fmt.Errorf(\"failed to get user: %w\", err).Error())"
  - "        return"
  - "    }"
  - "    "
  - "    c.Success(http.StatusOK, user)"
  - "}"
  
  # Example: Error Handling (Required for errcheck linter)
  - "ALWAYS handle errors explicitly - never ignore:"
  - "// BAD: Ignored error (fails errcheck)"
  - "json.Unmarshal(data, &result)"
  - ""
  - "// GOOD: Proper error handling"
  - "if err := json.Unmarshal(data, &result); err != nil {"
  - "    return fmt.Errorf(\"failed to unmarshal data: %w\", err)"
  - "}"
  
  # Example: Security-compliant code (Required for gosec)
  - "Security best practices for gosec compliance:"
  - "// BAD: Hardcoded credentials (fails gosec)"
  - "const secretKey = \"hardcoded-secret\""
  - ""
  - "// GOOD: Environment-based configuration"
  - "secretKey := os.Getenv(\"JWT_SECRET_KEY\")"
  - "if secretKey == \"\" {"
  - "    return errors.New(\"JWT_SECRET_KEY environment variable required\")"
  - "}"
  
  # Example: Constants for repeated strings (goconst compliance)
  - "Extract repeated strings to constants:"
  - "// BAD: Repeated strings (fails goconst)"
  - "log.Info(\"user created\")"
  - "log.Info(\"user created\")"
  - "log.Info(\"user created\")"
  - ""
  - "// GOOD: Use constants"
  - "const msgUserCreated = \"user created\""
  - "log.Info(msgUserCreated)"
  
  # Example: Reduce cyclomatic complexity (gocyclo compliance)
  - "Keep functions simple (complexity < 15):"
  - "// BAD: High complexity function"
  - "func processUser(user *User) error {"
  - "    if user == nil { /* logic */ }"
  - "    if user.Email == \"\" { /* logic */ }"
  - "    if user.Age < 18 { /* logic */ }"
  - "    // ... many more conditions"
  - "}"
  - ""
  - "// GOOD: Extract validation to separate functions"
  - "func processUser(user *User) error {"
  - "    if err := validateUser(user); err != nil { return err }"
  - "    if err := processUserLogic(user); err != nil { return err }"
  - "    return nil"
  - "}"

  # ========================
  # EXAMPLES AND PATTERNS
  # ========================
  
  # Adding HTTP Endpoints
  - "Create handler functions with proper error handling and response formatting."
  - "func GetUsers(c *gra.Context) { /* validate input, call service, return response */ }"
  - "Use c.BindJSON() for request parsing, c.Success() for responses, c.Error() for errors."
  - "Register routes: r.GET(/users, GetUsers) with appropriate middleware."
  
  # Adding Middleware
  - "Implement middleware with proper context handling."
  - "func CustomMiddleware() gra.MiddlewareFunc { return func(c *gra.Context) { /* logic */ c.Next() } }"
  - "Test middleware behavior with mock contexts and edge cases."
  
  # Adding Database Models
  - "Define models in orm/models/ with proper struct tags."
  - "type User struct { ID int json id db id Name string json name db name validate required }"
  - "Create corresponding migrations in migrations/ directory."
  
  # Adding Cache Strategies
  - "Implement cache interfaces in cache/ with TTL and eviction policies."
  - "type CacheStrategy interface { Get(key string) ([]byte, error); Set(key string, value []byte, ttl time.Duration) error }"
  
  # Adding Authentication
  - "Use JWT middleware for protected routes: authRoutes.Use(middleware.Auth(jwtService, user))"
  - "Access user info via c.GetUser() in handlers."

  # ========================
  # DOCUMENTATION STANDARDS
  # ========================
  - "Update README.md for any new features or breaking changes."
  - "Add examples in examples/ directory for new major features."
  - "Update docs/ website for comprehensive feature documentation."
  - "Include inline code comments for complex algorithms or business logic."
  - "Document API endpoints with request/response examples."
  - "Maintain CHANGELOG.md with version history and migration guides."

  # ========================
  # QUALITY ENFORCEMENT RULES
  # ========================
  
  # MANDATORY: Every commit MUST pass these checks
  - "Pre-commit hook automatically runs: gofmt, go vet, golangci-lint, security scan, tests"
  - "If pre-commit hook fails, commit is BLOCKED until issues are fixed"
  - "Override pre-commit hook ONLY for emergency hotfixes: git commit --no-verify"
  
  # MANDATORY: Every pull request MUST pass
  - "Full quality pipeline: make quality (test + coverage + verify + security)"
  - "SonarQube Quality Gate status: PASSED (A rating required)"
  - "Test coverage: No decrease in overall coverage, ≥70% for new code"
  - "Zero tolerance: No bugs, vulnerabilities, or security hotspots"
  
  # Code review requirements
  - "Pull request description MUST include: make verify output, coverage report link"
  - "Breaking changes MUST be documented with migration guide"
  - "New features MUST include: examples, tests, documentation"
  - "Security-related changes MUST include: threat model, security review"
  
  # Continuous Integration (CI) requirements
  - "All CI checks MUST pass: tests, coverage, quality gates, security scans"
  - "SonarQube analysis MUST run on every PR with quality gate enforcement"
  - "Automated deployment ONLY after ALL quality gates pass"

# For detailed information, refer to:
# - README.md: Project overview and quick start
# - CONTRIBUTING.md: Development guidelines and processes  
# - docs/: Comprehensive documentation and guides
# - examples/: Working code examples and demonstrations

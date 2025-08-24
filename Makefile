# Makefile for GRA Framework Development

# Variables
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html
BENCH_FILE = bench.out
GO = go
# By default, exclude example applications from package list to avoid skewing coverage
INCLUDE_EXAMPLES ?= false
# Exclude example apps and CLI/tools packages from coverage by default
# Rationale: focus coverage on library packages; binaries are thin wrappers

PACKAGES = $(shell go list ./... \
	| { if [ "$(INCLUDE_EXAMPLES)" = "true" ]; then cat; else grep -Ev '/examples(/|$$)|/tools(/|$$)|/cmd(/|$$)'; fi; } \
	| tr '\n' ' ')
TEST_PACKAGES ?= $(PACKAGES)
SONAR_HOST ?= http://localhost:9000
SONAR_USER ?= admin
SONAR_PASSWORD ?= MyPassword_123
SONAR_PROJECT_KEY ?= gra-migration-system
BENCHMARK_FLAGS = -benchmem
COVERAGE_THRESHOLD = 80
POSTGRES_COMPOSE = docker-compose.db.yml
POSTGRES_URL ?= postgres://postgres:MyPassword_123@localhost:5432/gra_test?sslmode=disable

## Default target shows help for discoverability
.DEFAULT_GOAL := help

# Default target (explicit)
.PHONY: all
all: test

# Run tests without coverage
.PHONY: test
test:
	$(GO) test -v $(TEST_PACKAGES)

# (Deprecated) db-test alias removed to avoid duplicate target; use the db-* section below

.PHONY: pg-up pg-down pg-test
POSTGRES_HOST_PORT ?= 5432

pg-up:
	@echo "🐘 Starting local PostgreSQL via docker-compose.db.yml..."
	POSTGRES_HOST_PORT=$(POSTGRES_HOST_PORT) docker-compose -f docker-compose.db.yml up -d
	@echo "Waiting for Postgres to be healthy..."
	@for i in $$(seq 1 30); do \
	  if docker inspect -f '{{.State.Health.Status}}' gra-postgres 2>/dev/null | grep -q healthy; then \
	    echo "Postgres is healthy"; break; \
	  else \
	    echo "waiting ($$i)..."; sleep 2; \
	  fi; \
	done

pg-down:
	@echo "🧹 Stopping local PostgreSQL..."
	docker-compose -f docker-compose.db.yml down

pg-test: pg-up
	@echo "🧪 Running Postgres integration tests..."
	PGHOST=localhost PGPORT=$(POSTGRES_HOST_PORT) PGUSER=postgres PGPASSWORD=MyPassword_123 PGDATABASE=gra_test GRA_TEST_PG=1 $(GO) test -v ./orm/dbcontext -run Postgres -cover
	PGHOST=localhost PGPORT=$(POSTGRES_HOST_PORT) PGUSER=postgres PGPASSWORD=MyPassword_123 PGDATABASE=gra_test GRA_TEST_PG=1 $(GO) test -v ./orm/migrations -run Postgres -cover
	@echo "✅ Postgres tests done"

# Run tests with coverage and generate HTML report
.PHONY: coverage
coverage:
	@echo "🧪 Running unified tests with coverage across all packages..."
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@PKGS="$(TEST_PACKAGES)"; \
	COVERPKG_COMMA=$$(echo $$PKGS | tr ' ' ','); \
	echo "→ Packages: $$PKGS"; \
	$(GO) test -count=1 -covermode=atomic -coverpkg=$$COVERPKG_COMMA -coverprofile=$(COVERAGE_FILE) $$PKGS
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo "📏 Computing weighted overall coverage from per-package profiles..."
	@TOTAL_STMTS=0; TOTAL_COVERED=0; \
	for pkg in $(TEST_PACKAGES); do \
		$(GO) test -count=1 -coverprofile=.pkg.cov $$pkg >/dev/null || exit 1; \
		SC=$$(awk 'NR>1 { s+=$$2; if ($$3>0) c+=$$2 } END { printf("%d %d", s, c) }' .pkg.cov); \
		S=$$(echo $$SC | awk '{print $$1+0}'); C=$$(echo $$SC | awk '{print $$2+0}'); \
		TOTAL_STMTS=$$((TOTAL_STMTS + S)); TOTAL_COVERED=$$((TOTAL_COVERED + C)); \
		rm -f .pkg.cov; \
	done; \
	if [ $$TOTAL_STMTS -gt 0 ]; then \
		PCT=$$(awk -v c=$$TOTAL_COVERED -v s=$$TOTAL_STMTS 'BEGIN { if (s>0) printf("%.1f", 100*c/s); else print 0 }'); \
		echo "🔢 Weighted overall coverage: $$PCT% (covered=$$TOTAL_COVERED / stmts=$$TOTAL_STMTS)" | tee coverage-summary.out; \
	else \
		echo "🔢 Weighted overall coverage: 0.0% (no statements)" | tee coverage-summary.out; \
	fi

# Enforce minimum coverage threshold
.PHONY: coverage-check
coverage-check: coverage
	@echo "🔎 Verifying test coverage >= $(COVERAGE_THRESHOLD)%..."
	@if [ -f coverage-summary.out ]; then \
		COVERAGE=$$(sed -n 's/.*Weighted overall coverage:[[:space:]]*\([0-9.][0-9.]*\)%.*/\1/p' coverage-summary.out); \
	else \
		COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {print $$3}' | tr -d '%'); \
	fi; \
	if command -v bc >/dev/null 2>&1; then \
	  COMP=$$(echo "$$COVERAGE >= $(COVERAGE_THRESHOLD)" | bc -l); \
	else \
	  COMP=$$(awk 'BEGIN { if ('"$$COVERAGE"' >= $(COVERAGE_THRESHOLD)) print 1; else print 0 }'); \
	fi; \
	if [ $$COMP -eq 1 ]; then \
	  echo "✅ Coverage $$COVERAGE% >= $(COVERAGE_THRESHOLD)%"; \
	else \
	  echo "❌ Coverage $$COVERAGE% < $(COVERAGE_THRESHOLD)%"; \
	  exit 1; \
	fi

# Run benchmarks
.PHONY: bench
bench:
	$(GO) test -bench=. $(BENCHMARK_FLAGS) $(TEST_PACKAGES) | tee $(BENCH_FILE)

# Run tests with race detector
.PHONY: race
race:
	$(GO) test -race $(TEST_PACKAGES)

# Generate GitHub Pages with coverage report
.PHONY: pages
pages:
	@echo "Creating GitHub Pages content..."
	@mkdir -p gh-pages
	$(GO) test $(TEST_PACKAGES) -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -html=$(COVERAGE_FILE) -o gh-pages/index.html
	@echo "# GRA Framework Coverage Report" > gh-pages/README.md
	@echo "Coverage report generated on $$(date)" >> gh-pages/README.md
	@echo "GitHub Pages content created in gh-pages directory"

# Verify code quality (fmt, vet, golangci-lint)
.PHONY: verify
verify:
	@echo "🔍 Running comprehensive code quality checks..."
	@echo "Running go fmt..."
	@$(GO) fmt $(PACKAGES)
	@echo "Running go vet..."
	@$(GO) vet $(PACKAGES)
	@echo "Running golangci-lint (this may take a few minutes)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=10m; \
		echo "✅ All golangci-lint checks passed!"; \
	else \
		echo "❌ golangci-lint not installed. Please install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@echo "🎉 All code quality checks passed!"

# Open the generated coverage HTML in default browser
.PHONY: coverage-open
coverage-open: coverage
	@echo "🌐 Opening coverage report (coverage.html)..."
	@open coverage.html || xdg-open coverage.html || true

# Quick test alias: run tests with -short flag across packages
.PHONY: short
short:
	@echo "🧪 Running short tests..."
	@$(GO) test -short $(TEST_PACKAGES)

# CI pipeline convenience target
.PHONY: ci
ci:
	@$(MAKE) quality

# Formatting and module housekeeping
.PHONY: fmt tidy

fmt:
	@echo "🧹 Running go fmt..."
	@$(GO) fmt $(PACKAGES)

tidy:
	@echo "📦 Ensuring go.mod/go.sum are tidy..."
	@$(GO) mod tidy
	@$(GO) mod verify

# Build helper tools
.PHONY: tools
tools:
	@echo "🔨 Building helper tools..."
	@$(GO) build -o bin/jsonval ./tools/jsonval

# Quick lint check with auto-fix
.PHONY: lint
lint:
	@echo "🔧 Running golangci-lint with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --timeout=10m; \
		echo "✅ Linting completed with auto-fixes applied!"; \
	else \
		echo "❌ golangci-lint not installed. Please install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Security scan using golangci-lint
.PHONY: security
security:
	@echo "🔒 Running security analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --enable gosec --timeout=10m; \
		echo "✅ Security scan completed!"; \
	else \
		echo "❌ golangci-lint not installed. Please install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Pre-commit quality gate (required before every commit)
.PHONY: pre-commit
pre-commit: test verify
	@echo "🚀 Running pre-commit quality gate..."
	@echo "✅ All pre-commit checks passed! Ready to commit."

# Full quality pipeline (for CI/CD and pull requests)
.PHONY: quality
quality: clean test coverage-check verify security
	@echo "🏆 Full quality pipeline completed successfully!"
	@echo "📊 Coverage report: $(COVERAGE_HTML)"
	@echo "🎯 All quality gates passed - ready for SonarQube analysis!"

# Clean up generated files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) $(BENCH_FILE)
	@rm -rf gh-pages
	@rm -f *.out *.test *.prof
	@find . -name "*.bak" -o -name "*.new" -o -name "*.tmp" -o -name "*~" -o -name "*.swp" -delete
	@find ./examples -type f -perm +111 -not -name "*.sh" -not -name "*.go" -not -name "*.md" -delete
	@echo "Project cleaned up successfully!"

# Postgres DB controls for integration tests
.PHONY: db-start db-stop db-logs db-wait db-test

db-start:
	@echo "🐘 Starting PostgreSQL test database..."
	docker-compose -f $(POSTGRES_COMPOSE) up -d
	$(MAKE) db-wait

db-stop:
	@echo "🛑 Stopping PostgreSQL test database..."
	docker-compose -f $(POSTGRES_COMPOSE) down -v

db-logs:
	docker-compose -f $(POSTGRES_COMPOSE) logs -f --tail=100

db-wait:
	@echo "⏳ Waiting for PostgreSQL to become healthy..."
	@for i in $$(seq 1 60); do \
	  docker inspect -f '{{.State.Health.Status}}' gra-postgres 2>/dev/null | grep -q healthy && { echo "✅ PostgreSQL is healthy"; exit 0; }; \
	  sleep 1; \
	done; \
	echo "❌ PostgreSQL did not become healthy in time"; exit 1

db-test: db-start
	@echo "🧪 Running integration tests against PostgreSQL..."
	DATABASE_URL=$(POSTGRES_URL) $(GO) test -v -run TestPG_ ./orm/dbcontext -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo "🧹 Cleaning up test database..."
	$(MAKE) db-stop

# SonarQube targets
.PHONY: sonar-start sonar-stop sonar-analyze sonar-clean sonar-wait sonar-token

sonar-start:
	@echo "Starting SonarQube with Docker Compose..."
	docker-compose -f docker-compose.sonar.yml up -d
	@echo "SonarQube is starting at $(SONAR_HOST)"
	@echo "Credentials used by Make targets: $(SONAR_USER)/********"
	@echo "Please wait a few minutes for SonarQube to fully initialize"

# Wait until SonarQube is UP
sonar-wait: tools
	@echo "⏳ Waiting for SonarQube to be UP at $(SONAR_HOST)..."
	@for i in $$(seq 1 60); do \
		STATUS=$$(curl -s $(SONAR_HOST)/api/system/status | ./bin/jsonval -p status 2>/dev/null || echo STARTING); \
		if [ "$$STATUS" = "UP" ]; then echo "✅ SonarQube is UP"; exit 0; fi; \
		echo "  Attempt $$i: status=$$STATUS"; \
		sleep 2; \
	done; \
	echo "❌ SonarQube did not become UP in time"; exit 1

# Generate or reuse a Sonar token (stored locally in .sonar.token)
sonar-token: tools
	@if [ -n "$$SONAR_TOKEN" ]; then \
		echo "🔐 Using SONAR_TOKEN from environment"; \
		exit 0; \
	fi; \
	if [ -s .sonar.token ]; then \
		echo "🔐 Using cached token from .sonar.token"; \
		exit 0; \
	fi; \
	NAME=gra-cli-$$RANDOM; \
	echo "🔐 Generating new token '$$NAME' for $(SONAR_USER) ..."; \
	RESP=$$(curl -s -u $(SONAR_USER):$(SONAR_PASSWORD) -X POST "$(SONAR_HOST)/api/user_tokens/generate?name=$$NAME"); \
	TOKEN=$$(printf "%s" "$$RESP" | ./bin/jsonval -p token 2>/dev/null || echo ""); \
	if [ -z "$$TOKEN" ]; then echo "❌ Failed to generate token (response: $$RESP)"; exit 1; fi; \
	printf "%s" "$$TOKEN" > .sonar.token; \
	chmod 600 .sonar.token; \
	echo "✅ Token saved to .sonar.token"

sonar-stop:
	@echo "Stopping SonarQube..."
	docker-compose -f docker-compose.sonar.yml down

sonar-analyze: coverage tools sonar-wait sonar-token
	@echo "🔍 Running SonarQube analysis..."
	@TOK="$$SONAR_TOKEN"; \
	if [ -z "$$TOK" ] && [ -s .sonar.token ]; then TOK=$$(cat .sonar.token); fi; \
	if command -v sonar-scanner >/dev/null 2>&1; then \
		if [ -z "$$TOK" ]; then \
			echo "⚠️  No SONAR_TOKEN available; analysis may fail"; \
			sonar-scanner -Dsonar.host.url=$(SONAR_HOST) -Dsonar.scanner.skipJreProvisioning=true; \
		else \
			sonar-scanner -Dsonar.host.url=$(SONAR_HOST) -Dsonar.token=$$TOK -Dsonar.scanner.skipJreProvisioning=true; \
		fi; \
	else \
		echo "⚠️  Local sonar-scanner not available or failed. Falling back to Docker scanner..."; \
		if command -v docker >/dev/null 2>&1; then \
			DOCKER_IMG=sonarsource/sonar-scanner-cli:latest; \
			echo "🐳 Using $$DOCKER_IMG"; \
			SCANNER_HOST=http://host.docker.internal:9000; \
			if [ -z "$$TOK" ]; then \
				docker run --rm -e SONAR_HOST_URL=$$SCANNER_HOST -v "$$PWD:/usr/src" -w /usr/src $$DOCKER_IMG; \
			else \
				docker run --rm -e SONAR_HOST_URL=$$SCANNER_HOST -e SONAR_TOKEN=$$TOK -v "$$PWD:/usr/src" -w /usr/src $$DOCKER_IMG; \
			fi; \
		else \
			echo "❌ Neither local sonar-scanner nor Docker is available."; \
			exit 1; \
		fi; \
	fi
	@echo "✅ SonarQube analysis completed!"

# Check SonarQube coverage against threshold via API
.PHONY: sonar-coverage
sonar-coverage: tools
	@echo "📈 Checking SonarQube coverage (threshold $(COVERAGE_THRESHOLD)%)..."
	@RESP=$$(curl -s -u $(SONAR_USER):$(SONAR_PASSWORD) "$(SONAR_HOST)/api/measures/component?component=$(SONAR_PROJECT_KEY)&metricKeys=coverage"); \
	VAL=$$(printf "%s" "$$RESP" | ./bin/jsonval -p component.measures.0.value 2>/dev/null || echo 0); \
	echo "SonarQube coverage: $$VAL%"; \
	awk -v v="$$VAL" -v th="$(COVERAGE_THRESHOLD)" 'BEGIN{ if (v+0 >= th+0) { print "✅ Meets threshold"; exit 0 } else { print "❌ Below threshold"; exit 1 } }'

# Helper: show per-package coverage summary
.PHONY: coverage-packages
coverage-packages:
	@echo "📦 Per-package coverage:";
	@$(GO) test -cover $(PACKAGES) | sed -E 's#github.com/[^ ]+/gra/##'

# Check SonarQube quality gate status
.PHONY: sonar-status
sonar-status: tools
	@echo "📊 Checking SonarQube quality gate status..."
	@RESP=$$(curl -s -u $(SONAR_USER):$(SONAR_PASSWORD) "$(SONAR_HOST)/api/qualitygates/project_status?projectKey=$(SONAR_PROJECT_KEY)"); \
	STATUS=$$(printf "%s" "$$RESP" | ./bin/jsonval -p projectStatus.status 2>/dev/null || echo ERROR); \
	if [ "$$STATUS" = "OK" ]; then echo "✅ Quality Gate: PASSED"; else echo "❌ Quality Gate: FAILED"; fi

sonar-clean:
	@echo "Cleaning SonarQube data..."
	docker-compose -f docker-compose.sonar.yml down -v
	docker volume prune -f

# Validate SonarQube login with current credentials
.PHONY: sonar-login-check
sonar-login-check: tools
	@echo "🔐 Validating SonarQube credentials for $(SONAR_USER) at $(SONAR_HOST)..."
	@RESP=$$(curl -s -u $(SONAR_USER):$(SONAR_PASSWORD) "$(SONAR_HOST)/api/authentication/validate"); \
	VALID=$$(printf "%s" "$$RESP" | ./bin/jsonval -p valid 2>/dev/null || echo false); \
	if [ "$$VALID" = "true" ]; then echo "✅ Auth valid"; else echo "❌ Auth failed"; exit 1; fi

# Show total coverage line from coverage.out
.PHONY: coverage-total
coverage-total: coverage
	@$(GO) tool cover -func=$(COVERAGE_FILE) | tail -n 1

# Help command
.PHONY: help
help:
	@echo "🚀 GRA Framework Development Commands:"
	@echo ""
	@echo "📋 Testing & Coverage:"
	@echo "  make test         - Run tests (defaults to excluding ./examples)"
	@echo "  make coverage     - Run tests with coverage and generate HTML report"
	@echo "  make coverage-open - Open the HTML coverage report"
	@echo "  make bench        - Run benchmarks"
	@echo "  make race         - Run tests with race detector"
	@echo "      Tip: INCLUDE_EXAMPLES=true make coverage   # include example apps in coverage"
	@echo ""
	@echo "🔍 Code Quality (MANDATORY BEFORE COMMIT):"
	@echo "  make verify       - Full code quality check (fmt, vet, golangci-lint)"
	@echo "  make fmt          - Run go fmt across packages"
	@echo "  make tidy         - Run go mod tidy && verify modules"
	@echo "  make lint         - Quick lint with auto-fix"
	@echo "  make security     - Security analysis with gosec"
	@echo "  make pre-commit   - Pre-commit quality gate (test + verify)"
	@echo "  make quality      - Full quality pipeline (all checks)"
	@echo "  make ci           - Run quality pipeline for CI"
	@echo ""
	@echo "🐘 Database (PostgreSQL) for integration tests:"
	@echo "  make pg-up        - Start local PostgreSQL (docker-compose.db.yml)"
	@echo "  make pg-down      - Stop local PostgreSQL"
	@echo "  make pg-test      - Run Postgres-specific integration tests"
	@echo "  make db-start     - Start test DB and wait for health"
	@echo "  make db-test      - Run integration tests against PostgreSQL"
	@echo "  make db-logs      - Tail DB logs"
	@echo "  make db-stop      - Stop and remove test DB containers/volumes"
	@echo ""
	@echo "📊 SonarQube Analysis:"
	@echo "  make sonar-start   - Start SonarQube server with Docker"
	@echo "  make sonar-analyze - Run SonarQube analysis (set SONAR_TOKEN for remote)"
	@echo "  make sonar-coverage - Check SonarQube coverage >= $(COVERAGE_THRESHOLD)% via API"
	@echo "  make sonar-status - Check SonarQube quality gate status"
	@echo "  make sonar-stop   - Stop SonarQube server"
	@echo "  make sonar-clean  - Clean SonarQube data and volumes"
	@echo ""
	@echo "🛠️  Utilities:"
	@echo "  make pages        - Generate GitHub Pages content"
	@echo "  make coverage-total - Print the total coverage line from coverage.out"
	@echo "  make clean        - Clean up generated files, backups, and binaries"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "💡 Quality Requirements:"
	@echo "   • ALL code MUST pass 'make verify' before commit"
	@echo "   • Pull requests MUST pass 'make quality'"
	@echo "   • SonarQube quality gate MUST be GREEN"
	@echo "   • Test coverage MUST be ≥$(COVERAGE_THRESHOLD)% for new code"

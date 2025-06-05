# Makefile for GRA Framework Development

# Variables
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html
BENCH_FILE = bench.out
GO = go
PACKAGES = ./...
BENCHMARK_FLAGS = -benchmem

# Default target
.PHONY: all
all: test

# Run tests without coverage
.PHONY: test
test:
	$(GO) test -v $(PACKAGES)

# Run tests with coverage and generate HTML report
.PHONY: coverage
coverage:
	$(GO) test -v -coverprofile=$(COVERAGE_FILE) $(PACKAGES)
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@$(GO) tool cover -func=$(COVERAGE_FILE)

# Run benchmarks
.PHONY: bench
bench:
	$(GO) test -bench=. $(BENCHMARK_FLAGS) $(PACKAGES) | tee $(BENCH_FILE)

# Run tests with race detector
.PHONY: race
race:
	$(GO) test -race $(PACKAGES)

# Generate GitHub Pages with coverage report
.PHONY: pages
pages:
	@echo "Creating GitHub Pages content..."
	@mkdir -p gh-pages
	$(GO) test $(PACKAGES) -coverprofile=$(COVERAGE_FILE)
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
quality: clean test coverage verify security
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

# SonarQube targets
.PHONY: sonar-start sonar-stop sonar-analyze sonar-clean

sonar-start:
	@echo "Starting SonarQube with Docker Compose..."
	docker-compose -f docker-compose.sonar.yml up -d
	@echo "SonarQube is starting at http://localhost:9000"
	@echo "Default credentials: admin/admin"
	@echo "Please wait a few minutes for SonarQube to fully initialize"

sonar-stop:
	@echo "Stopping SonarQube..."
	docker-compose -f docker-compose.sonar.yml down

sonar-analyze: coverage
	@echo "🔍 Running SonarQube analysis..."
	@if [ -z "$$SONAR_TOKEN" ]; then \
		echo "⚠️  SONAR_TOKEN not set. Running local analysis..."; \
		if command -v sonar-scanner >/dev/null 2>&1; then \
			sonar-scanner -Dsonar.host.url=http://localhost:9000; \
		else \
			echo "❌ sonar-scanner not installed. Please install SonarQube Scanner."; \
			exit 1; \
		fi \
	else \
		echo "🚀 Running SonarQube analysis with token..."; \
		if command -v sonar-scanner >/dev/null 2>&1; then \
			sonar-scanner -Dsonar.token=$$SONAR_TOKEN; \
		else \
			echo "❌ sonar-scanner not installed. Please install SonarQube Scanner."; \
			exit 1; \
		fi \
	fi
	@echo "✅ SonarQube analysis completed!"

# Check SonarQube quality gate status
.PHONY: sonar-status
sonar-status:
	@echo "📊 Checking SonarQube quality gate status..."
	@curl -s -u admin:admin "http://localhost:9000/api/qualitygates/project_status?projectKey=gra-migration-system" | \
		python3 -c "import sys, json; data = json.load(sys.stdin); print('✅ Quality Gate: PASSED' if data['projectStatus']['status'] == 'OK' else '❌ Quality Gate: FAILED')"

sonar-clean:
	@echo "Cleaning SonarQube data..."
	docker-compose -f docker-compose.sonar.yml down -v
	docker volume prune -f

# Help command
.PHONY: help
help:
	@echo "🚀 GRA Framework Development Commands:"
	@echo ""
	@echo "📋 Testing & Coverage:"
	@echo "  make test         - Run tests"
	@echo "  make coverage     - Run tests with coverage and generate HTML report"
	@echo "  make bench        - Run benchmarks"
	@echo "  make race         - Run tests with race detector"
	@echo ""
	@echo "🔍 Code Quality (MANDATORY BEFORE COMMIT):"
	@echo "  make verify       - Full code quality check (fmt, vet, golangci-lint)"
	@echo "  make lint         - Quick lint with auto-fix"
	@echo "  make security     - Security analysis with gosec"
	@echo "  make pre-commit   - Pre-commit quality gate (test + verify)"
	@echo "  make quality      - Full quality pipeline (all checks)"
	@echo ""
	@echo "📊 SonarQube Analysis:"
	@echo "  make sonar-start  - Start SonarQube server with Docker"
	@echo "  make sonar-analyze- Run SonarQube analysis (set SONAR_TOKEN for remote)"
	@echo "  make sonar-status - Check SonarQube quality gate status"
	@echo "  make sonar-stop   - Stop SonarQube server"
	@echo "  make sonar-clean  - Clean SonarQube data and volumes"
	@echo ""
	@echo "🛠️  Utilities:"
	@echo "  make pages        - Generate GitHub Pages content"
	@echo "  make clean        - Clean up generated files, backups, and binaries"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "💡 Quality Requirements:"
	@echo "   • ALL code MUST pass 'make verify' before commit"
	@echo "   • Pull requests MUST pass 'make quality'"
	@echo "   • SonarQube quality gate MUST be GREEN"
	@echo "   • Test coverage MUST be ≥70% for new code"

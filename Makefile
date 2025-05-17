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

# Verify code quality (fmt, vet, lint)
.PHONY: verify
verify:
	@echo "Running go fmt..."
	@$(GO) fmt $(PACKAGES)
	@echo "Running go vet..."
	@$(GO) vet $(PACKAGES)
	@if command -v golint >/dev/null 2>&1; then \
		echo "Running golint..."; \
		golint $(PACKAGES); \
	else \
		echo "golint not installed. Skipping."; \
	fi

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

# Help command
.PHONY: help
help:
	@echo "GRA Framework Development Commands:"
	@echo "  make test      - Run tests"
	@echo "  make coverage  - Run tests with coverage and generate HTML report"
	@echo "  make bench     - Run benchmarks"
	@echo "  make race      - Run tests with race detector"
	@echo "  make pages     - Generate GitHub Pages content"
	@echo "  make verify    - Verify code quality (fmt, vet, lint)"
	@echo "  make clean     - Clean up generated files, backups, and binaries"
	@echo "  make help      - Show this help message"

# EF Migrate - Makefile for build automation

# Variables
BINARY_NAME=ef-migrate
BINARY_PATH=./bin/$(BINARY_NAME)
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION?=v1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean test build

# Build the binary
.PHONY: build
build: $(BINARY_PATH)

$(BINARY_PATH): $(GO_FILES)
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "✅ Built $(BINARY_NAME) -> $(BINARY_PATH)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✅ Built for multiple platforms"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v -race -cover ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# Run benchmark tests
.PHONY: benchmark
benchmark:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Install the binary
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed $(BINARY_NAME) to /usr/local/bin/"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run
	@echo "✅ Code linted"

# Run security scan
.PHONY: security
security:
	@echo "🛡️  Running security scan..."
	@gosec ./...
	@echo "✅ Security scan completed"

# Create release
.PHONY: release
release: clean test build-all
	@echo "🚀 Creating release $(VERSION)..."
	@mkdir -p release
	@cp bin/* release/
	@tar -czf release/$(BINARY_NAME)-$(VERSION).tar.gz -C bin .
	@echo "✅ Release created: release/$(BINARY_NAME)-$(VERSION).tar.gz"

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "⚙️  Setting up development environment..."
	@go mod tidy
	@go mod download
	@echo "✅ Development environment ready"

# Quick development test
.PHONY: dev-test
dev-test: build
	@echo "🧪 Running quick development test..."
	@$(BINARY_PATH) help
	@echo "✅ Development test passed"

# Documentation
.PHONY: docs
docs:
	@echo "📚 Generating documentation..."
	@go doc -all . > docs.txt
	@echo "✅ Documentation generated: docs.txt"

# Help
.PHONY: help
help:
	@echo "EF Migrate - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  install       - Install binary to /usr/local/bin"
	@echo "  release       - Create release package"
	@echo ""
	@echo "Test Commands:"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  benchmark     - Run benchmark tests"
	@echo ""
	@echo "Quality Commands:"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Run security scan"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev-setup     - Setup development environment"
	@echo "  dev-test      - Quick development test"
	@echo "  docs          - Generate documentation"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Example Usage:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make install"
	@echo "  make release VERSION=v1.1.0"

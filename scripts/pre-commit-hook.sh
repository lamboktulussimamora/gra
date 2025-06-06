#!/bin/bash
# Pre-commit hook for GRA Framework
# This hook ensures all code passes quality checks before allowing commits

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[PRE-COMMIT]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_status "🚨 Running pre-commit quality checks..."

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f ".golangci.yml" ]; then
    print_error "This doesn't appear to be the GRA project root directory"
    exit 1
fi

# Get list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    print_status "No Go files staged for commit, skipping quality checks"
    exit 0
fi

print_status "Found staged Go files: $(echo $STAGED_GO_FILES | wc -w | tr -d ' ') files"

# Run go fmt on staged files
print_status "🔧 Checking code formatting..."
UNFORMATTED_FILES=$(echo "$STAGED_GO_FILES" | xargs gofmt -l)
if [ -n "$UNFORMATTED_FILES" ]; then
    print_error "The following files need formatting:"
    echo "$UNFORMATTED_FILES"
    print_status "Run 'gofmt -w $UNFORMATTED_FILES' to fix formatting"
    exit 1
fi
print_success "Code formatting: PASSED"

# Run go vet
print_status "🔍 Running go vet..."
if ! go vet ./... 2>/dev/null; then
    print_error "go vet found issues"
    print_status "Run 'go vet ./...' to see details"
    exit 1
fi
print_success "Go vet: PASSED"

# Run golangci-lint on staged files
print_status "🧹 Running golangci-lint..."
if command -v golangci-lint >/dev/null 2>&1; then
    # Create a temporary config for faster pre-commit checks
    cat > .golangci-precommit.yml << EOF
run:
  timeout: 5m
  issues-exit-code: 1

linters:
  disable-all: true
  enable:
    - errcheck
    - gosec
    - govet
    - staticcheck
    - ineffassign
    - unused
    - misspell

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
EOF

    if ! golangci-lint run --config .golangci-precommit.yml --new-from-patch=HEAD^ 2>/dev/null; then
        rm -f .golangci-precommit.yml
        print_error "golangci-lint found issues in staged files"
        print_status "Run 'golangci-lint run --new-from-patch=HEAD^' to see details"
        print_status "Or run 'golangci-lint run --fix' to auto-fix some issues"
        exit 1
    fi
    rm -f .golangci-precommit.yml
    print_success "golangci-lint: PASSED"
else
    print_error "golangci-lint not found"
    print_status "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
    exit 1
fi

# Run tests on packages that have staged files
print_status "🧪 Running tests for affected packages..."
AFFECTED_PACKAGES=$(echo "$STAGED_GO_FILES" | xargs -I {} dirname {} | sort -u | grep -v '^examples' | grep -v '^debug' | tr '\n' ' ')

if [ -n "$AFFECTED_PACKAGES" ]; then
    for pkg in $AFFECTED_PACKAGES; do
        if [ -f "$pkg"/*_test.go ] 2>/dev/null; then
            if ! go test "./$pkg" -timeout=30s >/dev/null 2>&1; then
                print_error "Tests failed for package: $pkg"
                print_status "Run 'go test ./$pkg -v' to see details"
                exit 1
            fi
        fi
    done
    print_success "Tests: PASSED"
else
    print_status "No testable packages affected, skipping tests"
fi

# Security check on staged files
print_status "🔒 Running security analysis..."
if golangci-lint run --config <(echo "
run:
  timeout: 2m
linters:
  disable-all: true
  enable:
    - gosec
") --new-from-patch=HEAD^ >/dev/null 2>&1; then
    print_success "Security check: PASSED"
else
    print_error "Security issues found in staged files"
    print_status "Run 'golangci-lint run --enable gosec' to see details"
    exit 1
fi

print_success "🎉 All pre-commit checks passed! Proceeding with commit..."

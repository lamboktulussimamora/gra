#!/bin/bash
# Setup script for GRA Framework development environment
# This script installs pre-commit hooks and validates the development setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[SETUP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

main() {
    print_status "🚀 Setting up GRA Framework development environment..."
    echo ""
    
    # Check if we're in a git repository
    if [ ! -d ".git" ]; then
        print_error "This is not a git repository"
        print_status "Initialize git with: git init"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f ".golangci.yml" ]; then
        print_error "This doesn't appear to be the GRA project root directory"
        exit 1
    fi
    
    # Check required tools
    print_status "🔧 Checking required development tools..."
    
    MISSING_TOOLS=()
    
    if ! command_exists go; then
        MISSING_TOOLS+=("go")
    else
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go $GO_VERSION installed"
    fi
    
    if ! command_exists golangci-lint; then
        MISSING_TOOLS+=("golangci-lint")
    else
        GOLANGCI_VERSION=$(golangci-lint --version | awk '{print $4}')
        print_success "golangci-lint $GOLANGCI_VERSION installed"
    fi
    
    if ! command_exists docker; then
        print_warning "Docker not found (optional for SonarQube)"
    else
        print_success "Docker installed"
    fi
    
    if ! command_exists sonar-scanner; then
        print_warning "SonarQube scanner not found (optional)"
    else
        print_success "SonarQube scanner installed"
    fi
    
    # Install missing tools
    if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
        print_error "Missing required tools: ${MISSING_TOOLS[*]}"
        echo ""
        
        for tool in "${MISSING_TOOLS[@]}"; do
            case $tool in
                go)
                    print_status "Install Go from: https://golang.org/dl/"
                    ;;
                golangci-lint)
                    print_status "Install golangci-lint:"
                    print_status "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
                    ;;
            esac
        done
        
        print_error "Please install missing tools and run this script again"
        exit 1
    fi
    
    # Install pre-commit hook
    print_status "📋 Installing pre-commit hook..."
    
    HOOK_SOURCE="./scripts/pre-commit-hook.sh"
    HOOK_DEST=".git/hooks/pre-commit"
    
    if [ -f "$HOOK_DEST" ]; then
        print_warning "Pre-commit hook already exists"
        read -p "Replace existing pre-commit hook? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Keeping existing pre-commit hook"
        else
            cp "$HOOK_SOURCE" "$HOOK_DEST"
            chmod +x "$HOOK_DEST"
            print_success "Pre-commit hook updated"
        fi
    else
        cp "$HOOK_SOURCE" "$HOOK_DEST"
        chmod +x "$HOOK_DEST"
        print_success "Pre-commit hook installed"
    fi
    
    # Validate configuration files
    print_status "📝 Validating configuration files..."
    
    if [ -f ".golangci.yml" ]; then
        if golangci-lint config path > /dev/null 2>&1; then
            print_success "golangci-lint configuration is valid"
        else
            print_error "golangci-lint configuration is invalid"
            exit 1
        fi
    else
        print_error ".golangci.yml not found"
        exit 1
    fi
    
    if [ -f "sonar-project.properties" ]; then
        print_success "SonarQube configuration found"
    else
        print_warning "sonar-project.properties not found"
    fi
    
    # Test the development workflow
    print_status "🧪 Testing development workflow..."
    
    # Clean first
    if make clean >/dev/null 2>&1; then
        print_success "Clean: PASSED"
    else
        print_error "Clean: FAILED"
        exit 1
    fi
    
    # Run quick tests
    if make test >/dev/null 2>&1; then
        print_success "Tests: PASSED"
    else
        print_error "Tests: FAILED"
        print_status "Fix test failures before proceeding"
        exit 1
    fi
    
    # Run quality checks
    if make verify >/dev/null 2>&1; then
        print_success "Code quality checks: PASSED"
    else
        print_warning "Code quality checks: FAILED"
        print_status "Some quality issues found - run 'make verify' to see details"
    fi
    
    # Create quality check alias (optional)
    print_status "⚡ Setting up convenient aliases..."
    
    ALIAS_FILE="$HOME/.gra_aliases"
    cat > "$ALIAS_FILE" << 'EOF'
# GRA Framework Development Aliases
alias gra-quality="cd $(git rev-parse --show-toplevel) && ./scripts/quality-check.sh"
alias gra-quick="cd $(git rev-parse --show-toplevel) && make pre-commit"
alias gra-full="cd $(git rev-parse --show-toplevel) && make quality"
alias gra-sonar="cd $(git rev-parse --show-toplevel) && make sonar-start && sleep 30 && make sonar-analyze"
EOF
    
    print_status "Quality check aliases created in $ALIAS_FILE"
    print_status "Add to your shell profile (.zshrc, .bashrc): source $ALIAS_FILE"
    
    echo ""
    print_success "🎉 GRA Framework development environment setup completed!"
    echo ""
    print_status "📋 Next steps:"
    print_status "  1. Source aliases: source $ALIAS_FILE"
    print_status "  2. Test pre-commit: git add . && git commit -m 'test' (will be blocked by hook)"
    print_status "  3. Run quality check: ./scripts/quality-check.sh"
    print_status "  4. For SonarQube: make sonar-start && make sonar-analyze"
    echo ""
    print_status "🔍 Available commands:"
    print_status "  make help          - Show all available commands"
    print_status "  make pre-commit    - Quick pre-commit checks"
    print_status "  make quality       - Full quality pipeline"
    print_status "  make verify        - Code quality verification"
    echo ""
    print_success "Happy coding! 🚀"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "GRA Framework Development Environment Setup"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --check, -c    Only check tools, don't install hooks"
        echo ""
        echo "This script will:"
        echo "  • Check for required development tools"
        echo "  • Install pre-commit quality hooks"
        echo "  • Validate configuration files"
        echo "  • Test the development workflow"
        echo "  • Set up convenient aliases"
        ;;
    --check|-c)
        print_status "🔧 Checking development tools only..."
        # Just run the tool checks part
        if command_exists go; then
            print_success "Go installed"
        else
            print_error "Go not installed"
        fi
        if command_exists golangci-lint; then
            print_success "golangci-lint installed"
        else
            print_error "golangci-lint not installed"
        fi
        ;;
    *)
        main
        ;;
esac

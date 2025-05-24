#!/bin/bash
# GRA Framework - Deploy and Test Search Enhancements

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}GRA Framework - Search Enhancement Deployment${NC}"
echo "=================================================="

# Function to check if command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check for required commands
if ! command_exists node; then
  echo -e "${RED}Error: Node.js is required but not installed.${NC}"
  exit 1
fi

# Check for npm packages
echo -e "${YELLOW}Checking for required Node.js packages...${NC}"
for package in lunr cheerio; do
  if ! npm list --depth=0 | grep -q "$package"; then
    echo -e "Installing $package..."
    npm install "$package" --no-save
  fi
done

# Set up directories
DOCS_DIR="./docs"
SCRIPTS_DIR="./scripts"
TEST_DIR="./test-results"
SEARCH_INDEX_SCRIPT="$SCRIPTS_DIR/generate_search_index.js"

# Create test directory if it doesn't exist
mkdir -p "$TEST_DIR"

# Generate the main search index
echo -e "\n${YELLOW}Generating main search index...${NC}"
node "$SEARCH_INDEX_SCRIPT" --verbose

# Check if we need to generate versioned search indexes
if [[ -d "$DOCS_DIR/versions" ]]; then
  echo -e "\n${YELLOW}Generating versioned search indexes...${NC}"
  
  # Find all version directories
  VERSION_DIRS=$(find "$DOCS_DIR/versions" -mindepth 1 -maxdepth 1 -type d)
  
  for version_dir in $VERSION_DIRS; do
    version=$(basename "$version_dir")
    echo -e "Processing version: $version"
    
    # Create assets directory if it doesn't exist
    mkdir -p "$version_dir/assets"
    
    # Generate search index for this version
    DOCS_DIR="$version_dir" node "$SEARCH_INDEX_SCRIPT" --output-file "$version_dir/assets/search-index.json"
  done
fi

# Run search tests
echo -e "\n${YELLOW}Running search tests...${NC}"
if [[ -f "$SCRIPTS_DIR/advanced_search_test.js" ]]; then
  node "$SCRIPTS_DIR/advanced_search_test.js"
  TEST_STATUS=$?
  
  if [[ $TEST_STATUS -ne 0 ]]; then
    echo -e "${RED}Search tests failed!${NC}"
    echo "Check test results for more information."
  else
    echo -e "${GREEN}Search tests passed!${NC}"
  fi
else
  echo -e "${RED}Advanced search test script not found.${NC}"
  echo "Please ensure $SCRIPTS_DIR/advanced_search_test.js exists."
fi

echo -e "\n${GREEN}Search enhancement deployment complete!${NC}"
echo "=================================================="
echo -e "To validate the search functionality:"
echo -e "1. Open the documentation in a browser"
echo -e "2. Try searching for various terms"
echo -e "3. Check that highlighting and keyword display work correctly"
echo -e "4. Verify version awareness if using versioned documentation"

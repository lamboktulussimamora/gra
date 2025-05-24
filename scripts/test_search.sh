#!/bin/bash
# Test script for verifying search functionality

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Search Functionality Test Script${NC}"
echo "==============================="

# Check if lunr is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is required but not installed.${NC}"
    exit 1
fi

# Check for npm packages
echo "Checking for required Node.js packages..."
if ! npm list -g lunr | grep -q lunr; then
    echo -e "${YELLOW}Warning: lunr not found in global packages. Attempting to install...${NC}"
    npm install --location=global lunr
fi

if ! npm list -g cheerio | grep -q cheerio; then
    echo -e "${YELLOW}Warning: cheerio not found in global packages. Attempting to install...${NC}"
    npm install --location=global cheerio
fi

# Create test directory
TEST_DIR=$(mktemp -d)
echo "Created temporary test directory: $TEST_DIR"

# Create test markdown files
echo "Creating test markdown files..."
mkdir -p "$TEST_DIR/docs"

cat > "$TEST_DIR/docs/test1.md" <<EOF
# Test Document 1

This is a test document containing the word middleware.

## Example

\`\`\`go
app.Use(middleware.Logger())
\`\`\`
EOF

cat > "$TEST_DIR/docs/test2.md" <<EOF
# Test Document 2

This document talks about routing and HTTP methods.

## Example

\`\`\`go
app.GET("/api/users", handlers.ListUsers)
app.POST("/api/users", handlers.CreateUser)
\`\`\`
EOF

# Copy the search index generator
echo "Copying search index generator script..."
mkdir -p "$TEST_DIR/scripts" "$TEST_DIR/docs/assets"
cp ./scripts/generate_search_index.js "$TEST_DIR/scripts/"

# Create a modified version for testing
cat > "$TEST_DIR/scripts/search_test.js" <<EOF
const fs = require('fs');
const path = require('path');
const lunr = require('lunr');

// Load the search index
const indexData = JSON.parse(fs.readFileSync(path.join(__dirname, '../docs/assets/search-index.json'), 'utf8'));
const index = lunr.Index.load(indexData.index);
const documents = indexData.documents;

// Test searches
function testSearch(query, expectedDocIds) {
    console.log(\`Testing search for "\${query}"...\`);
    
    const results = index.search(query);
    
    // Check if we found any results
    if (results.length === 0) {
        console.log(\`  ❌ No results found for "\${query}"\`);
        return false;
    }
    
    // Check if we found the expected documents
    let allFound = true;
    for (const expectedId of expectedDocIds) {
        const found = results.some(result => result.ref === expectedId);
        if (!found) {
            console.log(\`  ❌ Expected document "\${expectedId}" not found in results\`);
            allFound = false;
        }
    }
    
    if (allFound) {
        console.log(\`  ✅ Found all expected documents\`);
    }
    
    // Show the results
    console.log(\`  Results: \${results.length}\`);
    results.forEach(result => {
        const doc = documents[result.ref];
        if (doc) {
            console.log(\`    - "\${doc.title}" (score: \${Math.round(result.score * 100) / 100})\`);
        }
    });
    
    return allFound;
}

// Generate the search index
console.log('Generating search index...');
require('./generate_search_index.js');

// Give time for the index to be generated
setTimeout(() => {
    // Run tests
    console.log('\nRunning search tests...\n');
    
    const tests = [
        { query: 'middleware', expectedDocs: ['test1'] },
        { query: 'routing', expectedDocs: ['test2'] },
        { query: 'GET', expectedDocs: ['test2'] },
        { query: 'POST', expectedDocs: ['test2'] }
    ];
    
    let passed = 0;
    for (const test of tests) {
        if (testSearch(test.query, test.expectedDocs)) {
            passed++;
        }
        console.log('');
    }
    
    console.log(\`\${passed} of \${tests.length} tests passed\`);
    
    if (passed === tests.length) {
        console.log('\n✅ All search tests passed!');
        process.exit(0);
    } else {
        console.log('\n❌ Some search tests failed.');
        process.exit(1);
    }
}, 1000);
EOF

# Run the test
echo "Generating search index for test documents..."
cd "$TEST_DIR"
node scripts/generate_search_index.js

echo -e "\n${YELLOW}Running search functionality test...${NC}"
node scripts/search_test.js

# Cleanup
echo -e "\nCleaning up test directory..."
rm -rf "$TEST_DIR"
echo -e "${GREEN}Test completed!${NC}"

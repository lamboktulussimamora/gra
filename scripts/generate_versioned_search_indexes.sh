#!/bin/bash
# Script to generate search indexes for versioned documentation
# This script should be run after generate_version_docs.sh

# Check for Node.js
if ! command -v node &> /dev/null; then
    echo "Node.js is required but not installed. Please install Node.js first."
    exit 1
fi

# Base directories
DOCS_DIR="$(dirname "$0")/../docs"
VERSIONS_DIR="$DOCS_DIR/versions"

# Check if versions directory exists
if [ ! -d "$VERSIONS_DIR" ]; then
    echo "Versions directory not found. Please run generate_version_docs.sh first."
    exit 1
fi

# Generate search index for main documentation
echo "Generating search index for main documentation..."
node "$(dirname "$0")/generate_search_index.js"

# Get all version directories
versions=$(ls "$VERSIONS_DIR")

# Generate search index for each version
for version in $versions; do
    if [ -d "$VERSIONS_DIR/$version" ]; then
        echo "Generating search index for version $version..."
        
        # Create assets directory if it doesn't exist
        mkdir -p "$VERSIONS_DIR/$version/assets"
        
        # Set the current working directory to the version directory
        cd "$VERSIONS_DIR/$version"
        
        # Use the main script with the correct path
        DOCS_DIR="$VERSIONS_DIR/$version" \
        node "$(dirname "$0")/../../scripts/generate_search_index.js" \
            --output-file="$VERSIONS_DIR/$version/assets/search-index.json" \
            --exclude "SEARCH_IMPLEMENTATION" \
            --verbose
        
        # Return to the original directory
        cd - > /dev/null
    fi
done

echo "All search indexes generated successfully!"

#!/bin/bash
# Script to generate version-specific documentation for GRA Framework
# Usage: ./generate_version_docs.sh [version] [tag]

# Check parameters
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <version> <tag>"
    echo "Example: $0 1.0.0 v1.0.0"
    exit 1
fi

VERSION=$1
TAG=$2

# Directory structure
DOCS_DIR="docs"
VERSION_DIR="${DOCS_DIR}/versions/${VERSION}"

# Create version directory
mkdir -p "${VERSION_DIR}"

echo "Creating documentation for ${TAG} (${VERSION})"

# Copy current docs to version directory
cp -r ${DOCS_DIR}/core-concepts ${VERSION_DIR}/
cp -r ${DOCS_DIR}/api-reference ${VERSION_DIR}/
cp -r ${DOCS_DIR}/middleware ${VERSION_DIR}/
cp -r ${DOCS_DIR}/examples ${VERSION_DIR}/
cp -r ${DOCS_DIR}/getting-started ${VERSION_DIR}/
cp ${DOCS_DIR}/README.md ${VERSION_DIR}/

# Update version information
sed -i '' "s/Current version: .*/Current version: ${VERSION} (${TAG})/" ${VERSION_DIR}/README.md

# Update main README with version information
if ! grep -q "${VERSION}" ${DOCS_DIR}/README.md; then
    awk -v version="${VERSION}" '
    /## Documentation Sections/ {
        print "\n## Available Versions\n";
        print "- [Latest (" version ")](./)\n- [v" version "](versions/" version "/)\n";
        print $0;
        next;
    }
    { print }' ${DOCS_DIR}/README.md > ${DOCS_DIR}/README.md.new
    
    mv ${DOCS_DIR}/README.md.new ${DOCS_DIR}/README.md
fi

echo "Version-specific documentation created at ${VERSION_DIR}"
echo "Don't forget to commit and push the changes, then deploy to GitHub Pages"

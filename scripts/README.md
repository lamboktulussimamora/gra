# Documentation Scripts

This directory contains scripts for generating and testing documentation. To run these scripts locally, you'll need to install the required dependencies.

## Dependencies

The documentation scripts require Node.js modules. Install them with:

```bash
# Global installation
npm install --location=global lunr@2.3.9 cheerio@1.0.0-rc.12

# Or locally in this directory
npm init -y
npm install --save-dev lunr@2.3.9 cheerio@1.0.0-rc.12
```

## Available Scripts

### Search Index Generator

Generate the main search index for the latest documentation:

```bash
# Basic usage
node scripts/generate_search_index.js

# Advanced usage with options
node scripts/generate_search_index.js --output-file=custom/path/search-index.json --exclude "pattern" --verbose
```

Options:
- `--output-file`: Specify custom output path for the search index
- `--exclude`: Exclude files matching the specified pattern (can be used multiple times)
- `--verbose` or `-v`: Show detailed output during index generation

This creates a search index file at `docs/assets/search-index.json` used by the documentation site's search functionality.

### Versioned Search Index Generator

Generate search indexes for all versioned documentation:

```bash
./scripts/generate_versioned_search_indexes.sh
```

This script runs after `generate_version_docs.sh` and creates search indexes for each documentation version.

### Link Checker

Check for broken links in the documentation:

```bash
./scripts/check_links.sh
```

### Search Index Analyzer

Analyze search index files to get statistics and verify content:

```bash
node scripts/analyze_search_index.js
```

This script scans for search index files (both main and versioned) and provides detailed statistics about each index, including:
- Document count
- Snippet and heading coverage
- File size
- Content breakdown by section

### Search Functionality Tests

#### Automated Test Suite

Test the search functionality with sample content:

```bash
./scripts/test_search.sh
```

This creates test documents, generates a search index, and verifies that search queries return expected results.

#### Interactive Search Testing

Interactively test search queries against the documentation index:

```bash
# Basic usage with default index
node scripts/search_test_interactive.js

# Using a specific index (like a versioned index)
node scripts/search_test_interactive.js --index ./docs/versions/1.2.0/assets/search-index.json

# Run a single query and exit
node scripts/search_test_interactive.js --query "middleware"
```

This provides a command-line interface for testing search queries and examining the returned results, including snippets and highlighting of matched terms.

#### Advanced Search Testing

Run a series of predefined test queries to validate search functionality:

```bash
node scripts/advanced_search_test.js
```

This runs a comprehensive search quality test that evaluates how well various types of queries perform against your search index. It tests common terms, HTTP methods, code symbols, and section names, reporting the success rate and highlighting any issues.

## GitHub Actions

The GitHub Actions workflows will install these dependencies automatically when running in CI/CD, so you don't need to commit the `node_modules` directory or `package.json` files.

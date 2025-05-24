# Search Functionality Implementation

The GRA framework documentation includes a robust search functionality implemented using Lunr.js. This document explains how the search feature works and provides guidance for testing and troubleshooting.

## Components

The search functionality comprises the following key components:

1. **Search Index Generator** (`/scripts/generate_search_index.js`): A Node.js script that parses all markdown files in the documentation and builds a search index with the following features:
   - Full-text indexing of documentation content
   - Content snippet extraction for search result previews
   - Heading extraction for section context
   - Support for excluding specific files or patterns
   - Metadata about the index for version awareness

2. **Versioned Search Index Generator** (`/scripts/generate_versioned_search_indexes.sh`): A shell script that generates search indices for all versioned documentation.

3. **Search Index Analyzer** (`/scripts/analyze_search_index.js`): A utility to analyze search indices and provide statistics about content coverage and completeness.

4. **Client-Side Search Logic** (`/docs/assets/js/docs.js`): JavaScript code that loads the search index and handles user searches.

5. **Search UI** (embedded in `docs/README.md`): The search input field and results container in the documentation UI.

6. **Search Index Files**:
   - Main documentation: `/docs/assets/search-index.json`
   - Versioned documentation: `/docs/versions/{version}/assets/search-index.json`

## How It Works

1. When documentation is built or updated, the `generate_search_index.js` script scans all markdown files in the `/docs` directory.

2. For each file, it:
   - Extracts the title
   - Cleans the content (removes code blocks, HTML tags, etc.)
   - Extracts meaningful content snippets for result previews
   - Captures headings to provide section context
   - Creates a document record with ID, title, content, snippet, headings, and URL

3. A Lunr.js search index is created from these documents, which enables fast full-text search.

4. When a user types in the search box, the client-side JavaScript:
   - Determines which search index to load based on the current version
   - Performs a search against this index
   - Displays matching results with titles, snippets, and section information
   - Allows users to click on results to navigate to the relevant page

## Testing the Search

### Automated Testing

The framework provides dedicated scripts for testing search functionality:

1. **Basic Search Test**:
   ```bash
   ./scripts/test_search.sh
   ```
   This script creates test documents with known content, generates a search index, and performs test searches to verify that expected results are returned.

2. **Interactive Search Test**:
   ```bash
   node scripts/search_test_interactive.js
   ```
   This utility provides an interactive command-line interface for testing search queries against the actual documentation index. Options include:
   - `--index <path>`: Test with a specific index file
   - `--query <text>`: Run a single query and exit
   - `--help`: Show usage information

   Example usage with versioned documentation:
   ```bash
   node scripts/search_test_interactive.js --index ./docs/versions/1.2.0/assets/search-index.json
   ```

### Manual Testing

To manually test the search functionality:

1. Generate the search indices:
   ```bash
   # Install dependencies if needed
   npm install --location=global lunr cheerio
   
   # Generate the main index with verbose output
   node scripts/generate_search_index.js --verbose
   
   # Generate versioned indices (if versions exist)
   ./scripts/generate_versioned_search_indexes.sh
   ```

2. Analyze the search indices:
   ```bash
   node scripts/analyze_search_index.js
   ```

3. Serve the documentation site locally:
   ```bash
   cd docs
   python -m http.server 8080
   # Or any other local server
   ```

4. Open http://localhost:8080 in your browser

5. Try searching for different terms relevant to the GRA framework:
   - Try "middleware" to find middleware-related content
   - Try "router" to find routing-related documentation
   - Try "JWT" to find authentication documentation
   - Try specific method names like "Use" or "GET"
   - Try partially complete words to test matching

## Common Issues and Troubleshooting

### Missing Search Results

If search results aren't appearing as expected:

1. Check that `search-index.json` exists in `/docs/assets/`
2. Verify the content is properly indexed by examining the JSON file
3. Check browser console for JavaScript errors
4. Ensure the search input and results container have the correct IDs:
   - Search input: `docSearch`
   - Results container: `searchResults`

### Performance Issues

If search is slow:

1. Check the size of the search index file - large indices can impact performance
2. Consider limiting the amount of content indexed per document
3. Add debouncing to the search input to reduce search frequency during typing

## Integration with Documentation Workflow

The search index is automatically generated during the documentation deployment process:

1. The GitHub Actions workflow installs necessary Node.js dependencies
2. It runs the search index generator script
3. The generated index is deployed with the documentation

When making changes to the documentation structure or content, ensure the search index generator is updated accordingly.

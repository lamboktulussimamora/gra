#!/usr/bin/env node
/**
 * Interactive Search Test Utility
 * 
 * This utility allows you to interactively test the search functionality 
 * of the GRA documentation by performing searches against the generated index.
 */

const fs = require('fs');
const path = require('path');
const readline = require('readline');

// Check if lunr is installed
try {
  require.resolve('lunr');
} catch (e) {
  console.error('Error: lunr module is not installed. Please run:');
  console.error('npm install --location=global lunr');
  process.exit(1);
}

const lunr = require('lunr');

// Configuration
const docsDir = path.resolve('./docs');
const defaultIndexPath = path.join(docsDir, 'assets', 'search-index.json');

// Command line arguments
const args = process.argv.slice(2);
let indexPath = defaultIndexPath;
let isInteractive = true;
let singleQuery = null;

// Process arguments
for (let i = 0; i < args.length; i++) {
  if (args[i] === '--index' && args[i + 1]) {
    indexPath = args[i + 1];
    i++; // Skip the next arg
  } else if (args[i] === '--query' && args[i + 1]) {
    singleQuery = args[i + 1];
    isInteractive = false;
    i++; // Skip the next arg
  } else if (args[i] === '--help' || args[i] === '-h') {
    printHelp();
    process.exit(0);
  }
}

// Check if index file exists
if (!fs.existsSync(indexPath)) {
  console.error(`Error: Search index not found at ${indexPath}`);
  console.error('Run the search index generator first:');
  console.error('node scripts/generate_search_index.js');
  process.exit(1);
}

// Print help information
function printHelp() {
  console.log(`
Interactive Search Test Utility for GRA Documentation

Usage: 
  node search_test_interactive.js [options]

Options:
  --index <path>   Path to the search index JSON file
                  Default: ${defaultIndexPath}
  --query <text>   Run a single query and exit (non-interactive mode)
  --help, -h       Show this help information

Examples:
  # Run in interactive mode using default index
  node search_test_interactive.js
  
  # Run in interactive mode with a custom index
  node search_test_interactive.js --index ./docs/versions/1.2.0/assets/search-index.json
  
  # Run a single query and exit
  node search_test_interactive.js --query "middleware"`);
}

// Load the search index
let searchIndex;
let documents;
let metadata;

try {
  console.log(`Loading search index from ${indexPath}...`);
  const indexData = JSON.parse(fs.readFileSync(indexPath, 'utf8'));
  searchIndex = lunr.Index.load(indexData.index);
  documents = indexData.documents;
  metadata = indexData.metadata || {};
  
  console.log('Search index loaded successfully!');
  if (metadata.documentCount) {
    console.log(`Index contains ${metadata.documentCount} documents`);
  }
  if (metadata.isVersioned) {
    console.log(`Version: ${metadata.version || 'unknown'} (versioned documentation)`);
  }
  console.log('-----------------------------------');
} catch (error) {
  console.error(`Error loading search index: ${error.message}`);
  process.exit(1);
}

// If single query mode
if (singleQuery) {
  performSearch(singleQuery);
  process.exit(0);
}

// Interactive mode
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

console.log('Welcome to the GRA Documentation Search Test Utility');
console.log('Type your search queries or "exit" to quit');
console.log('-----------------------------------');

// Interactive loop
function promptForQuery() {
  rl.question('Search query> ', (query) => {
    if (query.toLowerCase() === 'exit' || query.toLowerCase() === 'quit') {
      console.log('Exiting search utility.');
      rl.close();
      return;
    }
    
    if (query.trim() === '') {
      promptForQuery();
      return;
    }

    performSearch(query.trim());
    promptForQuery();
  });
}

// Perform a search
function performSearch(query) {
  console.log(`\nSearching for: "${query}"`);
  
  try {
    const results = searchIndex.search(query);
    
    if (results.length === 0) {
      console.log('No results found.\n');
      return;
    }
    
    console.log(`Found ${results.length} results:\n`);
    
    results.forEach((result, index) => {
      const doc = documents[result.ref];
      if (!doc) {
        console.log(`[${index + 1}] Unknown document (${result.ref})`);
        return;
      }
      
      console.log(`[${index + 1}] ${doc.title} (score: ${result.score.toFixed(2)})`);
      console.log(`    URL: ${doc.url}`);
      
      if (doc.snippet) {
        // Highlight query terms in snippet
        const highlightedSnippet = highlightTerms(doc.snippet, query.split(' '));
        console.log(`    ${highlightedSnippet}`);
      }
      
      if (doc.headings && doc.headings.length > 0) {
        console.log(`    In sections: ${doc.headings.join(' > ')}`);
      }
      
      console.log('');
    });
    
    console.log('-----------------------------------');
  } catch (error) {
    console.error(`Error performing search: ${error.message}`);
  }
}

// Highlight terms in text
function highlightTerms(text, terms) {
  let result = text;
  terms.forEach(term => {
    if (term.length < 3) return; // Skip short terms
    
    const regex = new RegExp(`(${term})`, 'gi');
    result = result.replace(regex, '\x1b[33m$1\x1b[0m'); // Yellow highlight
  });
  
  return result;
}

// Start the interactive prompt
promptForQuery();

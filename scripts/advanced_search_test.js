#!/usr/bin/env node
/**
 * Advanced Search Test Script for GRA Framework Documentation
 * 
 * This utility tests various search query types against the generated index
 * and provides detailed analysis of search results. It helps to validate the 
 * effectiveness of the search indexing approach.
 */

const fs = require('fs');
const path = require('path');

// Check for required dependencies
try {
  require.resolve('lunr');
} catch (e) {
  console.error('Error: lunr module is not installed. Please run:');
  console.error('npm install --location=global lunr');
  process.exit(1);
}

const lunr = require('lunr');
const testQueries = [
  { 
    name: "Basic framework terms", 
    queries: ["middleware", "router", "context", "handler"]
  },
  {
    name: "HTTP methods",
    queries: ["GET", "POST", "PUT", "DELETE"]
  },
  {
    name: "Code symbols", 
    queries: ["app.Use", "router.GET", "NewApp", "ServeHTTP"]
  },
  {
    name: "Complex combinations",
    queries: ["auth middleware", "json response", "error handling"]
  },
  {
    name: "Documentation sections",
    queries: ["getting started", "core concepts", "middleware", "examples"]
  }
];

// Configuration
const docsDir = path.resolve('./docs');
const indexPath = path.join(docsDir, 'assets', 'search-index.json');

/**
 * Run search tests against the index
 */
async function runSearchTests() {
  console.log('=== GRA Framework Documentation Search Test ===\n');
  
  // Check if index exists
  if (!fs.existsSync(indexPath)) {
    console.error(`Error: Search index not found at ${indexPath}`);
    console.error('Please run the search index generator first:');
    console.error('node scripts/generate_search_index.js');
    process.exit(1);
  }
  
  // Load index
  let searchIndex;
  let documents;
  let metadata;
  
  try {
    console.log(`Loading search index from ${indexPath}...`);
    const data = JSON.parse(fs.readFileSync(indexPath, 'utf8'));
    searchIndex = lunr.Index.load(data.index);
    documents = data.documents;
    metadata = data.metadata || {};
    
    console.log('Search index loaded successfully!');
    console.log(`Index contains ${Object.keys(documents).length} documents`);
    console.log(`Generated: ${metadata.generatedAt || 'Unknown'}`);
    console.log('================================\n');
  } catch (error) {
    console.error(`Error loading search index: ${error.message}`);
    process.exit(1);
  }
  
  // Run test queries
  let testsPassed = 0;
  let testsFailed = 0;
  let totalQueries = 0;
  
  for (const testCase of testQueries) {
    console.log(`\nTest Case: ${testCase.name}`);
    console.log('---------------------------------');
    
    for (const query of testCase.queries) {
      totalQueries++;
      try {
        const results = searchIndex.search(query);
        const pass = results.length > 0;
        
        if (pass) {
          testsPassed++;
          console.log(`✓ Query "${query}" returned ${results.length} results`);
          
          // Display top 3 results
          results.slice(0, 3).forEach((result, idx) => {
            const doc = documents[result.ref];
            if (doc) {
              console.log(`   ${idx+1}. [${result.score.toFixed(2)}] ${doc.title}`);
            }
          });
        } else {
          testsFailed++;
          console.log(`✗ Query "${query}" returned 0 results`);
        }
      } catch (error) {
        testsFailed++;
        console.error(`✗ Error with query "${query}": ${error.message}`);
      }
    }
  }
  
  // Summary
  console.log('\n=== Test Summary ===');
  console.log(`Total queries: ${totalQueries}`);
  console.log(`Successful queries: ${testsPassed}`);
  console.log(`Failed queries: ${testsFailed}`);
  console.log(`Success rate: ${((testsPassed / totalQueries) * 100).toFixed(1)}%`);
  
  if (testsFailed > 0) {
    console.log('\nSome queries failed! Consider improving your search index.');
  } else {
    console.log('\nAll queries successful!');
  }
}

runSearchTests().catch(console.error);

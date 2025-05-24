#!/usr/bin/env node
// Script to analyze the search index and display statistics

const fs = require('fs');
const path = require('path');

// Find all search index files
function findSearchIndexFiles(rootDir) {
  const indexFiles = [];
  
  // Check main index
  const mainIndex = path.join(rootDir, 'docs', 'assets', 'search-index.json');
  if (fs.existsSync(mainIndex)) {
    indexFiles.push(mainIndex);
  }
  
  // Check versioned indices
  const versionsDir = path.join(rootDir, 'docs', 'versions');
  if (fs.existsSync(versionsDir)) {
    const versions = fs.readdirSync(versionsDir);
    
    versions.forEach(version => {
      const versionIndexPath = path.join(versionsDir, version, 'assets', 'search-index.json');
      if (fs.existsSync(versionIndexPath)) {
        indexFiles.push(versionIndexPath);
      }
    });
  }
  
  return indexFiles;
}

// Analyze a search index file
function analyzeIndex(filePath) {
  try {
    const data = fs.readFileSync(filePath, 'utf8');
    const index = JSON.parse(data);
    
    const metadata = index.metadata || {};
    const documents = index.documents || {};
    const documentCount = Object.keys(documents).length;
    
    // Calculate stats
    const stats = {
      documentCount,
      averageSnippetLength: 0,
      hasSnippets: false,
      hasHeadings: false,
      topSections: {},
      fileSize: fs.statSync(filePath).size,
    };
    
    // Check if index has snippets and headings
    let snippetCount = 0;
    let snippetTotalLength = 0;
    let headingCount = 0;
    
    Object.values(documents).forEach(doc => {
      // Count sections
      const urlParts = doc.url.split('/').filter(Boolean);
      const section = urlParts.length > 0 ? urlParts[0] : 'Main';
      
      if (!stats.topSections[section]) {
        stats.topSections[section] = 0;
      }
      stats.topSections[section]++;
      
      // Check snippets
      if (doc.snippet && doc.snippet.length > 0) {
        snippetCount++;
        snippetTotalLength += doc.snippet.length;
      }
      
      // Check headings
      if (doc.headings && doc.headings.length > 0) {
        headingCount++;
      }
    });
    
    stats.hasSnippets = snippetCount > 0;
    stats.hasHeadings = headingCount > 0;
    stats.snippetCoverage = documentCount > 0 ? (snippetCount / documentCount) * 100 : 0;
    stats.headingCoverage = documentCount > 0 ? (headingCount / documentCount) * 100 : 0;
    
    if (snippetCount > 0) {
      stats.averageSnippetLength = snippetTotalLength / snippetCount;
    }
    
    // Sort sections by count
    const sortedSections = Object.entries(stats.topSections)
      .sort((a, b) => b[1] - a[1])
      .reduce((obj, [key, value]) => {
        obj[key] = value;
        return obj;
      }, {});
      
    stats.topSections = sortedSections;
    
    return {
      path: filePath,
      version: metadata.version || 'unknown',
      isVersioned: metadata.isVersioned || false,
      generatedAt: metadata.generatedAt || 'unknown',
      stats
    };
    
  } catch (error) {
    return {
      path: filePath,
      error: error.message,
      stats: {}
    };
  }
}

// Format file size
function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' bytes';
  else if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
  else return (bytes / 1048576).toFixed(1) + ' MB';
}

// Main function
function main() {
  const rootDir = path.resolve('.');
  console.log(`Analyzing search indices from: ${rootDir}\n`);
  
  const indexFiles = findSearchIndexFiles(rootDir);
  
  if (indexFiles.length === 0) {
    console.log('No search index files found.');
    return;
  }
  
  console.log(`Found ${indexFiles.length} search index file(s):`);
  
  indexFiles.forEach(filePath => {
    console.log(`\n=====================================`);
    console.log(`Analyzing: ${path.relative(rootDir, filePath)}`);
    console.log(`=====================================`);
    
    const analysis = analyzeIndex(filePath);
    
    if (analysis.error) {
      console.log(`Error: ${analysis.error}`);
      return;
    }
    
    const stats = analysis.stats;
    
    console.log(`Version: ${analysis.version}`);
    console.log(`Generated at: ${analysis.generatedAt}`);
    console.log(`Versioned: ${analysis.isVersioned ? 'Yes' : 'No'}`);
    console.log(`File size: ${formatSize(stats.fileSize)}`);
    console.log(`Document count: ${stats.documentCount}`);
    console.log(`Enhanced features:`);
    console.log(`  - Snippets: ${stats.hasSnippets ? 'Yes' : 'No'} (${stats.snippetCoverage.toFixed(1)}% coverage)`);
    if (stats.hasSnippets) {
      console.log(`  - Average snippet length: ${stats.averageSnippetLength.toFixed(1)} characters`);
    }
    console.log(`  - Headings: ${stats.hasHeadings ? 'Yes' : 'No'} (${stats.headingCoverage.toFixed(1)}% coverage)`);
    
    console.log(`\nContent breakdown by section:`);
    Object.entries(stats.topSections).slice(0, 5).forEach(([section, count]) => {
      const percentage = (count / stats.documentCount * 100).toFixed(1);
      console.log(`  - ${section}: ${count} documents (${percentage}%)`);
    });
  });
}

main();

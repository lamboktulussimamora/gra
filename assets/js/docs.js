// GRA Documentation Search and Version Selector

// Wait for DOM to be loaded
document.addEventListener('DOMContentLoaded', function() {
  // Initialize search
  setupSearch();
  
  // Initialize version selector
  setupVersionSelector();
});

// Set up search functionality
async function setupSearch() {
  const searchInput = document.getElementById('docSearch');
  const searchResults = document.getElementById('searchResults');
  
  if (!searchInput || !searchResults) return;
  
  let searchIndex;
  let searchDocs;
  let indexMetadata;
  
  try {
    // Determine the correct path for the search index based on URL
    let indexPath = '/assets/search-index.json';
    
    // If we're in a versioned documentation page, load the appropriate index
    const currentVersion = getCurrentVersion();
    if (currentVersion) {
      indexPath = `/versions/${currentVersion}/assets/search-index.json`;
    }
    
    // Load search index
    const response = await fetch(indexPath);
    if (!response.ok) throw new Error(`Failed to load search index from ${indexPath}`);
    
    const data = await response.json();
    searchIndex = lunr.Index.load(data.index);
    searchDocs = data.documents;
    indexMetadata = data.metadata;
    
    console.log('Search index loaded:', indexMetadata);
  } catch (error) {
    console.error('Error loading search index:', error);
    return;
  }
  
  // Search input event handler
  searchInput.addEventListener('input', function() {
    const query = this.value.trim();
    
    if (!query || query.length < 2) {
      searchResults.style.display = 'none';
      return;
    }
    
    try {
      // Search for results
      const results = searchIndex.search(query);
      displaySearchResults(results, searchDocs, searchResults, query, indexMetadata);
    } catch (error) {
      console.error('Search error:', error);
      searchResults.innerHTML = '<div class="search-result">Error performing search</div>';
      searchResults.style.display = 'block';
    }
  });
  
  // Close search results when clicking outside
  document.addEventListener('click', function(event) {
    if (!searchInput.contains(event.target) && !searchResults.contains(event.target)) {
      searchResults.style.display = 'none';
    }
  });
}

// Create a search result element
function createSearchResultElement(doc, result, queryTerms, indexMetadata) {
  // Get relative path parts for highlighting section
  const urlParts = doc.url.split('/').filter(Boolean);
  const section = urlParts.length > 0 ? urlParts[0] : 'Main';
  
  // Create headings breadcrumb if available
  let breadcrumb = '';
  if (doc.headings && doc.headings.length > 0) {
    breadcrumb = `<div class="result-breadcrumb">${doc.headings[0]}</div>`;
  }
  
  // Get snippet
  let snippetHtml = '';
  if (doc.snippet && doc.snippet.length > 0) {
    // Get most relevant snippet based on search terms, if function exists
    let snippet = doc.snippet;
    if (typeof getRelevantSnippet === 'function' && doc.content) {
      snippet = getRelevantSnippet(doc.content, queryTerms);
    }
    
    // Highlight search terms if available
    const highlightedSnippet = queryTerms.length > 0 ? 
      highlightSearchTerms(snippet, queryTerms) : snippet;
    snippetHtml = `<div class="result-snippet">${highlightedSnippet}</div>`;
  }
  
  // Get keywords
  let keywordsHtml = '';
  if (doc.keywords && doc.keywords.length > 0) {
    // Get relevant keywords based on search terms
    const relevantKeywords = typeof getRelevantKeywords === 'function' ? 
      getRelevantKeywords(doc.keywords, queryTerms) :
      doc.keywords.filter(kw => kw.length > 1).slice(0, 5);
    
    if (relevantKeywords.length > 0) {
      keywordsHtml = `<div class="result-keywords">
        ${relevantKeywords.map(kw => `<span class="keyword-tag">${kw}</span>`).join('')}
      </div>`;
    }
  }
  
  // Get version display if available
  const versionDisplay = typeof getVersionDisplay === 'function' && indexMetadata ? 
    getVersionDisplay(indexMetadata) : '';
    
  // Highlight title if needed
  const highlightedTitle = queryTerms.length > 0 ? 
    highlightSearchTerms(doc.title, queryTerms) : doc.title;
  
  const resultElement = document.createElement('div');
  resultElement.className = 'search-result';
  resultElement.innerHTML = `
    <h3>${highlightedTitle}</h3>
    ${breadcrumb}
    ${snippetHtml}
    ${keywordsHtml}
    <div class="result-meta">
      <span class="result-section">${section}</span>
      <span class="result-score">Relevance: ${Math.round(result.score * 100) / 100}</span>
      ${versionDisplay}
    </div>
  `;
  
  resultElement.addEventListener('click', function() {
    window.location.href = doc.url;
  });
  
  return resultElement;
}

// Add results header to container
function addResultsHeader(resultsContainer, count) {
  const resultsHeader = document.createElement('div');
  resultsHeader.className = 'search-results-header';
  resultsHeader.innerHTML = `<span>${count} result${count !== 1 ? 's' : ''} found</span>`;
  resultsContainer.appendChild(resultsHeader);
}

// Add "see more" element for additional results
function addSeeMoreElement(resultsContainer, remainingCount) {
  const seeMoreElement = document.createElement('div');
  seeMoreElement.className = 'search-result search-more';
  seeMoreElement.innerHTML = `<span>See ${remainingCount} more results...</span>`;
  resultsContainer.appendChild(seeMoreElement);
}

// Process query string into terms for highlighting
function processQueryTerms(query) {
  return query.toLowerCase().split(/\s+/).filter(term => term.length > 2);
}

// Display search results
function displaySearchResults(results, documents, resultsContainer, query='', indexMetadata=null) {
  // Handle no results case
  if (results.length === 0) {
    resultsContainer.innerHTML = '<div class="search-result">No results found</div>';
    resultsContainer.style.display = 'block';
    return;
  }
  
  // Process query terms for highlighting
  const queryTerms = processQueryTerms(query);
  resultsContainer.innerHTML = '';
  
  // Add results count header
  addResultsHeader(resultsContainer, results.length);
  
  // Display top 10 results
  results.slice(0, 10).forEach(result => {
    const doc = documents[result.ref];
    if (doc) {
      const resultElement = createSearchResultElement(doc, result, queryTerms, indexMetadata);
      resultsContainer.appendChild(resultElement);
    }
  });
  
  // Add "see more" if there are more results
  if (results.length > 10) {
    addSeeMoreElement(resultsContainer, results.length - 10);
  }
  
  resultsContainer.style.display = 'block';
}

// Set up version selector
function setupVersionSelector() {
  const versionSelect = document.getElementById('versionSelect');
  
  if (!versionSelect) return;
  
  // Switch version handler
  versionSelect.addEventListener('change', function() {
    const version = this.value;
    if (!version) {
      window.location.href = '/'; // Latest version
      return;
    }
    
    window.location.href = `/versions/${version}/`;
  });
  
  // Get current version from URL
  const currentVersion = getCurrentVersion();
  if (currentVersion) {
    // Add current version to selector if it's not already there
    let found = false;
    
    // Using Array.from to convert HTMLCollection to array for modern iteration
    Array.from(versionSelect.options).forEach(option => {
      if (option.value === currentVersion) {
        option.selected = true;
        found = true;
      }
    });
    
    if (!found) {
      const option = new Option(`v${currentVersion}`, currentVersion, true, true);
      versionSelect.appendChild(option);
    }
  }
}

// Get current version from URL
function getCurrentVersion() {
  const path = window.location.pathname;
  const regex = /\/versions\/([^/]+)/;
  const match = regex.exec(path);
  return match ? match[1] : null;
}

// Switch to a different documentation version
function switchVersion(version) {
  if (!version) {
    window.location.href = '/'; // Latest version
    return;
  }
  
  window.location.href = `/versions/${version}/`;
}

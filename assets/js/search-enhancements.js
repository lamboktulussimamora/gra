// GRA Documentation Search Enhancements

/**
 * Highlights search terms in a text string
 * @param {string} text - The text to highlight
 * @param {string[]} terms - Array of search terms to highlight
 * @returns {string} HTML with highlighted terms
 */
function highlightSearchTerms(text, terms) {
  if (!text || !terms || terms.length === 0) {
    return text;
  }
  
  // Create a safe copy to manipulate
  let result = text;
  
  // Escape special regex characters
  const escapeRegExp = (string) => string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  
  // Create a regex that matches any of the terms with word boundaries
  const termPattern = terms
    .filter(term => term.length > 2) // Only highlight terms of reasonable length
    .map(escapeRegExp)
    .join('|');
  
  if (!termPattern) {
    return text;
  }
  
  try {
    const regex = new RegExp(`(${termPattern})`, 'gi');
    result = result.replace(regex, '<span class="search-highlight">$1</span>');
  } catch (e) {
    console.error('Error highlighting search terms:', e);
  }
  
  return result;
}

/**
 * Formats version information for display
 * @param {Object} metadata - Search index metadata
 * @returns {string} HTML for version display or empty string
 */
function getVersionDisplay(metadata) {
  if (!metadata) {
    return '';
  }
  
  let versionClass = 'version-latest';
  let versionLabel = 'Latest';
  
  if (metadata.isVersioned) {
    versionLabel = metadata.version || 'Unknown';
    versionClass = 'version-previous';
    
    // Check if it's marked as development version
    if (versionLabel.includes('dev') || versionLabel.includes('alpha') || versionLabel.includes('beta')) {
      versionClass = 'version-development';
    }
  }
  
  return `<span class="result-version ${versionClass}">${versionLabel}</span>`;
}

/**
 * Returns relevant keywords for display based on search terms
 * @param {string[]} keywords - Array of all document keywords
 * @param {string[]} searchTerms - Array of search terms
 * @returns {string[]} Array of most relevant keywords
 */
function getRelevantKeywords(keywords, searchTerms) {
  if (!keywords || keywords.length === 0) {
    return [];
  }
  
  // First, include any keywords that match search terms
  const matchingKeywords = [];
  const otherKeywords = [];
  
  keywords.forEach(keyword => {
    let isMatching = false;
    
    // Check if this keyword matches any search term
    if (searchTerms && searchTerms.length > 0) {
      for (const term of searchTerms) {
        if (term.length > 2 && 
            (keyword.toLowerCase().includes(term.toLowerCase()) || 
             term.toLowerCase().includes(keyword.toLowerCase()))) {
          isMatching = true;
          break;
        }
      }
    }
    
    if (isMatching) {
      matchingKeywords.push(keyword);
    } else {
      otherKeywords.push(keyword);
    }
  });
  
  // Return matching keywords first, then others up to a maximum of 5 total
  return [...matchingKeywords, ...otherKeywords].slice(0, 5);
}

/**
 * Extracts the most relevant snippet for search results
 * @param {string} content - Full content of the document
 * @param {string[]} searchTerms - Array of search terms
 * @param {number} maxLength - Maximum length of snippet
 * @returns {string} Best text snippet for search result
 */
function getRelevantSnippet(content, searchTerms, maxLength = 160) {
  if (!content || content.length === 0) {
    return '';
  }

  // If no search terms, return the first part of content
  if (!searchTerms || searchTerms.length === 0) {
    return content.substring(0, maxLength) + (content.length > maxLength ? '...' : '');
  }

  // Split content into paragraphs
  const paragraphs = content.split(/\n\n+/);
  
  // Score each paragraph based on search term occurrence
  const scoredParagraphs = paragraphs.map(paragraph => {
    let score = 0;
    searchTerms.forEach(term => {
      // Count occurrences of search term
      const regex = new RegExp(term, 'gi');
      const matches = paragraph.match(regex);
      if (matches) {
        score += matches.length;
      }
    });
    return { text: paragraph, score };
  });
  
  // Sort by score (highest first)
  scoredParagraphs.sort((a, b) => b.score - a.score);
  
  // If no paragraphs matched, return first paragraph
  if (scoredParagraphs[0].score === 0) {
    return paragraphs[0].substring(0, maxLength) + (paragraphs[0].length > maxLength ? '...' : '');
  }
  
  // Return best matching paragraph
  const bestParagraph = scoredParagraphs[0].text.trim();
  return bestParagraph.substring(0, maxLength) + (bestParagraph.length > maxLength ? '...' : '');
}

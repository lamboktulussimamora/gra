// Generate search index for GRA documentation
const fs = require('fs');
const path = require('path');
const lunr = require('lunr');
const cheerio = require('cheerio');

// Parse command line arguments
const args = process.argv.slice(2);
let outputFilePath = null;
let excludePatterns = [];
let verbose = false;

for (let i = 0; i < args.length; i++) {
  if (args[i] === '--output-file' && args[i + 1]) {
    outputFilePath = args[i + 1];
    i++; // Skip the next argument
  } else if (args[i] === '--exclude' && args[i + 1]) {
    excludePatterns.push(new RegExp(args[i + 1]));
    i++; // Skip the next argument
  } else if (args[i] === '--verbose' || args[i] === '-v') {
    verbose = true;
  }
}

// Configuration
const docsDir = process.env.DOCS_DIR || path.resolve('./docs');
const outputFile = outputFilePath || path.join(docsDir, 'assets', 'search-index.json');
const supportedExtensions = ['.md'];
const ignoredDirs = ['node_modules', '.git', 'assets/js', 'assets/css'];

// Storage for all documents
const documents = [];

// Process markdown content
function processMarkdown(content, filePath) {
  // Convert relative links to absolute
  const relativePath = path.relative(docsDir, filePath);
  const urlPath = relativePath.replace(/README\.md$/i, '').replace(/\.md$/i, '');
  
  // Extract title from content
  const titleMatch = content.match(/^#\s+(.+)$/m);
  const title = titleMatch ? titleMatch[1] : path.basename(filePath, '.md');

  // Remove code blocks and HTML to create clean text for indexing
  const cleanContent = content
    .replace(/```[\s\S]*?```/g, '') // Remove code blocks
    .replace(/`([^`]+)`/g, '$1')    // Remove inline code marks
    .replace(/<!--[\s\S]*?-->/g, '') // Remove HTML comments
    .replace(/<[^>]+>/g, '')       // Remove HTML tags
    .replace(/\n+/g, ' ')          // Replace multiple newlines with space
    .trim();                       // Trim whitespace
    
  // Extract meaningful sections for content snippets
  const sections = [];
  const headings = [];
  
  // Extract headings for section context
  const headingMatches = content.matchAll(/^(#{2,}) (.+)$/gm);
  for (const match of headingMatches) {
    headings.push({
      level: match[1].length,
      text: match[2],
      position: match.index
    });
  }
  
  // Extract good paragraph snippets (at least 50 chars)
  const paragraphs = cleanContent.split(/\n\n+/);
  paragraphs.forEach(paragraph => {
    if (paragraph.trim().length >= 50) {
      sections.push(paragraph.trim());
    }
  });
  
  // If no good paragraphs found, use the whole cleaned content
  if (sections.length === 0) {
    sections.push(cleanContent);
  }
  
  // Extract keywords from the content (code methods, important terms)
  const keywords = extractKeywords(content, filePath);
  
  // Create document
  return {
    id: urlPath || '/',
    title: title,
    content: cleanContent,
    snippet: sections[0].substring(0, 160) + (sections[0].length > 160 ? '...' : ''),
    headings: headings.slice(0, 5).map(h => h.text),  // Include up to 5 headings for context
    keywords: keywords,
    url: urlPath ? `/${urlPath}/` : '/',
  };
}

// Extract important keywords from content
function extractKeywords(content, filePath) {
  const keywords = new Set();
  
  // Extract Go method names (like app.GET, router.Use)
  const methodMatches = content.match(/\b[a-zA-Z]+\.[A-Z][a-zA-Z]+\b/g) || [];
  
  // Extract Go function/method definitions
  const functionMatches = content.match(/\bfunc\s+\(?[a-zA-Z]*\)?\s*([A-Za-z]+\w*)/g) || [];
  if (functionMatches.length > 0) {
    functionMatches.forEach(match => {
      // Extract function name from the match
      const funcNameMatch = match.match(/\bfunc\s+\(?[a-zA-Z]*\)?\s*([A-Za-z]+\w*)/);
      if (funcNameMatch && funcNameMatch[1]) {
        keywords.add(funcNameMatch[1]);
      }
    });
  }
  methodMatches.forEach(match => keywords.add(match));
  
  // Extract code parameters in backticks
  const codeMatches = content.match(/`([^`]+)`/g) || [];
  codeMatches.forEach(matchStr => {
    // Remove the backticks
    const processedMatch = matchStr.replace(/`/g, '').trim();
    // Only add if it looks like a code term (not a general term in backticks)
    if (processedMatch.length > 2 && (
        /^[A-Z]/.test(processedMatch) || // Starts with capital letter
        processedMatch.includes('.') ||   // Contains a dot
        processedMatch.includes('(') ||   // Contains parenthesis
        /^[a-z]+[A-Z]/.test(processedMatch) // camelCase
      )) {
      keywords.add(processedMatch);
    }
  });
  
  // Extract HTTP methods if they're mentioned alone
  const httpMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'];
  httpMethods.forEach(method => {
    if (content.includes(method)) {
      keywords.add(method);
    }
  });
  
  // Extract common Go package names and frameworks
  const goPackages = ['gra', 'http', 'net/http', 'context', 'json', 'middleware', 'router', 'handler'];
  goPackages.forEach(pkg => {
    // Check for common Go import patterns or usage patterns
    if (content.toLowerCase().includes(pkg.toLowerCase())) {
      keywords.add(pkg);
    }
  });
  
  // Extract struct and interface names
  const structMatches = content.match(/type\s+([A-Z][a-zA-Z0-9]*)\s+(struct|interface)/g) || [];
  structMatches.forEach(match => {
    const nameMatch = match.match(/type\s+([A-Z][a-zA-Z0-9]*)/);
    if (nameMatch && nameMatch[1]) {
      keywords.add(nameMatch[1]);
    }
  });
  
  // Add section context based on the file path
  const pathParts = filePath.split('/');
  const sectionName = pathParts[pathParts.length - 2]; // Parent directory name
  if (sectionName && sectionName !== 'docs') {
    keywords.add(sectionName);
  }
  
  return Array.from(keywords);
}

// Walk through directories recursively
function walkDir(dir) {
  const files = fs.readdirSync(dir);

  files.forEach(file => {
    const filePath = path.join(dir, file);
    const stat = fs.statSync(filePath);

    // Skip ignored directories
    if (stat.isDirectory()) {
      if (!ignoredDirs.includes(file)) {
        walkDir(filePath);
      }
      return;
    }

    // Process only supported file extensions
    const ext = path.extname(file);
    if (!supportedExtensions.includes(ext.toLowerCase())) {
      return;
    }

    // Check if file should be excluded
    const relativePath = path.relative(docsDir, filePath);
    const shouldExclude = excludePatterns.some(pattern => pattern.test(relativePath));
    
    if (shouldExclude) {
      if (verbose) {
        console.log(`Excluding file: ${relativePath}`);
      }
      return;
    }

    // Read file content
    const content = fs.readFileSync(filePath, 'utf8');
    const doc = processMarkdown(content, filePath);
    documents.push(doc);
    
    if (verbose) {
      console.log(`Indexed document: ${doc.id} (${doc.title})`);
    }
  });
}

// Main function
function main() {
  console.log('Generating search index...');
  
  if (verbose) {
    console.log(`Source directory: ${docsDir}`);
    console.log(`Output file: ${outputFile}`);
    console.log(`Exclude patterns: ${excludePatterns.map(p => p.toString()).join(', ') || 'none'}`);
    console.log('----------------------------');
  }
  
  // Create assets directory if it doesn't exist
  const outputDir = path.dirname(outputFile);
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }
  
  // Process all documentation files
  walkDir(docsDir);
  console.log(`Found ${documents.length} documents`);
  
  // Create index
  console.log('Building search index...');
  const idx = lunr(function() {
    this.ref('id');
    this.field('title', { boost: 10 });
    this.field('content');
    this.field('keywords', { boost: 5 }); // Add keywords field with boosted relevance
    
    // Add metadata for more granular filtering
    this.metadataWhitelist = ['position'];
    
    documents.forEach(function(doc) {
      this.add(doc);
    }, this);
  });
  
  // Write index and documents to file
  const output = {
    index: idx,
    documents: documents.reduce((acc, doc) => {
      acc[doc.id] = {
        title: doc.title,
        url: doc.url,
        snippet: doc.snippet || '',
        headings: doc.headings || [],
        keywords: doc.keywords || []
      };
      return acc;
    }, {}),
    metadata: {
      generatedAt: new Date().toISOString(),
      documentCount: documents.length,
      version: '1.2.0', // GRA framework version
      source: path.basename(docsDir),
      // Include path info for versioned docs
      isVersioned: docsDir.includes('/versions/'),
      versionPath: docsDir.includes('/versions/') ? 
        path.basename(path.dirname(docsDir)) + '/' + path.basename(docsDir) : 
        null
    }
  };
  
  fs.writeFileSync(outputFile, JSON.stringify(output));
  console.log(`Search index written to ${outputFile}`);
  
  if (verbose) {
    console.log('----------------------------');
    console.log('Search index metadata:');
    console.log(JSON.stringify(output.metadata, null, 2));
  }
}

main();

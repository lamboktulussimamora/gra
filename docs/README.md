# GRA Framework Documentation

[![Test and Coverage](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml/badge.svg)](https://github.com/lamboktulussimamora/gra/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/lamboktulussimamora/gra/badge.svg?branch=main)](https://coveralls.io/github/lamboktulussimamora/gra?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/lamboktulussimamora/gra)](https://goreportcard.com/report/github.com/lamboktulussimamora/gra)
[![GitHub release](https://img.shields.io/github/release/lamboktulussimamora/gra.svg)](https://GitHub.com/lamboktulussimamora/gra/releases/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/lamboktulussimamora/gra.svg)](https://pkg.go.dev/github.com/lamboktulussimamora/gra)
[![GitHub stars](https://img.shields.io/github/stars/lamboktulussimamora/gra.svg)](https://GitHub.com/lamboktulussimamora/gra/stargazers/)

Welcome to the GRA Framework documentation. GRA is a lightweight HTTP framework for building web applications in Go, inspired by Gin but with a focus on simplicity, performance, and developer experience.

Current version: 1.2.0

<div class="header-bar">
  <div class="search-container">
    <input type="text" id="docSearch" placeholder="Search documentation..." />
    <div id="searchResults" style="display: none;"></div>
  </div>

  <div class="version-selector">
    <label for="versionSelect">Version: </label>
    <select id="versionSelect" onchange="switchVersion(this.value)">
      <option value="" selected>Latest (1.2.0)</option>
      <!-- Additional versions will be added automatically -->
    </select>
  </div>
</div>

<div class="doc-sections">
  <div class="doc-section">
    <h2>📚 Getting Started</h2>
    <p>Set up your first GRA application and learn the basics.</p>
    <ul>
      <li><a href="getting-started/#installation">Installation</a></li>
      <li><a href="getting-started/#quick-start">Quick Start Guide</a></li>
      <li><a href="getting-started/#project-structure">Project Structure</a></li>
    </ul>
    <a href="getting-started/" class="section-link">Explore Getting Started →</a>
  </div>
  
  <div class="doc-section">
    <h2>🎓 Tutorial</h2>
    <p>Step-by-step guide to build a complete REST API.</p>
    <ul>
      <li><a href="tutorial/#what-well-build">Project Setup</a></li>
      <li><a href="tutorial/#step-3-create-an-in-memory-store">Data Storage</a></li>
      <li><a href="tutorial/#step-4-create-api-handlers">API Handlers</a></li>
    </ul>
    <a href="tutorial/" class="section-link">Start Tutorial →</a>
  </div>
  
  <div class="doc-section">
    <h2>🧠 Core Concepts</h2>
    <p>Understand the fundamental concepts of GRA.</p>
    <ul>
      <li><a href="core-concepts/#router">Router</a></li>
      <li><a href="core-concepts/#context">Context</a></li>
      <li><a href="core-concepts/#middleware">Middleware</a></li>
      <li><a href="core-concepts/#request-handling">Request Handling</a></li>
    </ul>
    <a href="core-concepts/" class="section-link">Explore Core Concepts →</a>
  </div>
  
  <div class="doc-section">
    <h2>📋 API Reference</h2>
    <p>Detailed documentation for all GRA components.</p>
    <ul>
      <li><a href="api-reference/#core-package">Core Package</a></li>
      <li><a href="api-reference/#router-package">Router Package</a></li>
      <li><a href="api-reference/#context-package">Context Package</a></li>
      <li><a href="api-reference/#middleware-package">Middleware Package</a></li>
    </ul>
    <a href="api-reference/" class="section-link">Explore API Reference →</a>
  </div>
  
  <div class="doc-section">
    <h2>🔌 Middleware</h2>
    <p>Learn about built-in middleware and creating custom ones.</p>
    <ul>
      <li><a href="middleware/#using-middleware">Using Middleware</a></li>
      <li><a href="middleware/#built-in-middleware">Built-in Middleware</a></li>
      <li><a href="middleware/#creating-custom-middleware">Creating Custom Middleware</a></li>
    </ul>
    <a href="middleware/" class="section-link">Explore Middleware →</a>
  </div>
  
  <div class="doc-section">
    <h2>💡 Examples</h2>
    <p>Real-world examples and use cases.</p>
    <ul>
      <li><a href="examples/#basic-http-server">Basic HTTP Server</a></li>
      <li><a href="examples/#rest-api-with-crud-operations">REST API with CRUD</a></li>
      <li><a href="examples/#authentication-and-security">Authentication</a></li>
    </ul>
    <a href="examples/" class="section-link">Explore Examples →</a>
  </div>
</div>

## Features

- Context-based request handling
- HTTP routing with path parameters
- JWT authentication and authorization
- Secure HTTP headers middleware
- API versioning support
- Response caching
- Middleware support
- Request validation
- Standardized API responses
- Structured logging
- Clean architecture friendly

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/lamboktulussimamora/gra"
)

func main() {
    // Create a new router
    r := gra.New()

    // Define a route
    r.GET("/hello", func(c *gra.Context) {
        c.Success(http.StatusOK, "Hello World", nil)
    })

    // Start the server
    gra.Run(":8080", r)
}
```

## Installation

```bash
go get github.com/lamboktulussimamora/gra
```

## Benchmarks

GRA framework is designed with performance in mind. Here are some benchmark results:

```
BenchmarkRouterSimple/SimpleRoute-8               368399      3249 ns/op
BenchmarkRouterSimple/ParameterizedRoute-8        293060      4102 ns/op
BenchmarkRouterComplex/ManyRoutes_Simple-8        230602      5204 ns/op
BenchmarkRouterComplex/ManyRoutes_WithParameter-8 183795      6518 ns/op
BenchmarkRouterComplex/DeepNestedParameters-8     147219      8142 ns/op
```

## Community and Support

<div class="community-section">
  <div class="community-card">
    <h3>💬 GitHub Discussions</h3>
    <p>Ask questions, share ideas, and connect with other GRA users.</p>
    <a href="https://github.com/lamboktulussimamora/gra/discussions" target="_blank">Join the conversation</a>
  </div>
  
  <div class="community-card">
    <h3>🐛 Issue Tracker</h3>
    <p>Report bugs or request features for the framework.</p>
    <a href="https://github.com/lamboktulussimamora/gra/issues" target="_blank">Submit an issue</a>
  </div>
  
  <div class="community-card">
    <h3>👩‍💻 Contributing</h3>
    <p>Learn how to contribute to the GRA framework.</p>
    <a href="https://github.com/lamboktulussimamora/gra/blob/main/CONTRIBUTING.md" target="_blank">Read contributing guide</a>
  </div>
</div>

<div class="footer">
  <p>GRA Framework is released under the <a href="https://opensource.org/licenses/MIT" target="_blank">MIT License</a></p>
  <p>Documentation last updated: May 24, 2025</p>
  <p><a href="https://github.com/lamboktulussimamora/gra" target="_blank">GitHub Repository</a> | <a href="https://pkg.go.dev/github.com/lamboktulussimamora/gra" target="_blank">Go Reference</a></p>
</div>

<script src="/assets/js/lunr.min.js"></script>
<script src="/assets/js/docs.js"></script>

<link rel="stylesheet" href="/assets/css/style.css" />
    searchResults.style.display = 'block';
    
    // Example search results (to be replaced with actual implementation)
    const pages = [
      { title: 'Getting Started', url: 'getting-started/' },
      { title: 'Core Concepts - Router', url: 'core-concepts/#router' },
      { title: 'Core Concepts - Context', url: 'core-concepts/#context' },
      { title: 'Middleware Overview', url: 'middleware/' },
      { title: 'API Reference', url: 'api-reference/' },
      { title: 'Examples - Basic HTTP Server', url: 'examples/#basic-http-server' },
      // More pages would be indexed in reality
    ];
    
    const matches = pages.filter(page => 
      page.title.toLowerCase().includes(query)
    );
    
    if (matches.length === 0) {
      searchResults.innerHTML = '<div class="no-results">No results found</div>';
      return;
    }
    
    const resultsList = document.createElement('ul');
    matches.forEach(match => {
      const item = document.createElement('li');
      const link = document.createElement('a');
      link.href = match.url;
      link.textContent = match.title;
      item.appendChild(link);
      resultsList.appendChild(item);
    });
    
    searchResults.appendChild(resultsList);
  });
  
  // Close search results when clicking outside
  document.addEventListener('click', function(event) {
    if (!searchInput.contains(event.target) && !searchResults.contains(event.target)) {
      searchResults.style.display = 'none';
    }
  });
});

// Version switcher
function switchVersion(version) {
  if (version === '') {
    window.location.href = '/gra/';
  } else {
    window.location.href = '/gra/versions/' + version + '/';
  }
}

// Add styles
const style = document.createElement('style');
style.textContent = `
  .search-container {
    position: relative;
    margin: 20px 0;
  }
  
  #docSearch {
    width: 100%;
    padding: 10px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 16px;
  }
  
  #searchResults {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: white;
    border: 1px solid #ddd;
    border-radius: 0 0 4px 4px;
    box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    max-height: 300px;
    overflow-y: auto;
    z-index: 100;
  }
  
  #searchResults ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  #searchResults li {
    padding: 0;
    margin: 0;
  }
  
  #searchResults a {
    display: block;
    padding: 10px;
    text-decoration: none;
    color: #333;
    border-bottom: 1px solid #eee;
  }
  
  #searchResults a:hover {
    background-color: #f5f5f5;
  }
  
  .no-results {
    padding: 10px;
    color: #666;
    font-style: italic;
  }
  
  .version-selector {
    margin: 20px 0;
  }
  
  #versionSelect {
    padding: 5px;
    border-radius: 4px;
    border: 1px solid #ddd;
  }
`;
document.head.appendChild(style);
</script>

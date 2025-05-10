#!/bin/bash
# This script replaces all instances of "Content-Type" and "application/json" with constants

# Replace in context_test.go
sed -i '' 's/w.Header().Get("Content-Type")/w.Header().Get(headerContentType)/g' /Users/lamboktulussimamora/Projects/gra/context/context_test.go
sed -i '' 's/c.Writer.Header().Set("Content-Type", "application\/json")/c.Writer.Header().Set(headerContentType, contentTypeJSON)/g' /Users/lamboktulussimamora/Projects/gra/context/context_test.go
sed -i '' 's/contentType != "application\/json"/contentType != contentTypeJSON/g' /Users/lamboktulussimamora/Projects/gra/context/context_test.go

echo "Completed replacing string literals with constants"

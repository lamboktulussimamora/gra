#!/bin/bash

echo "GRA Framework Verification Script"
echo "================================="

# Check module name
echo -n "Checking go.mod... "
if grep -q "github.com/lamboktulussimamora/gra" /Users/lamboktulussimamora/Projects/gra/go.mod; then
  echo "✓ Module name updated"
else
  echo "✗ Module name not updated"
fi

# Check core.go
echo -n "Checking core.go... "
if grep -q "const Version = \"1.0.1\"" /Users/lamboktulussimamora/Projects/gra/gra.go; then
  echo "✓ Version updated"
else
  echo "✗ Version not updated"
fi

# Check for old references
echo -n "Checking for old references... "
if grep -r "go-core-framework" --include="*.go" /Users/lamboktulussimamora/Projects/gra > /dev/null; then
  echo "✗ Found old references"
else
  echo "✓ No old references found"
fi

# Check example imports
echo -n "Checking example imports... "
if grep -q "github.com/lamboktulussimamora/gra" /Users/lamboktulussimamora/Projects/gra/examples/basic/main.go; then
  echo "✓ Example imports updated"
else
  echo "✗ Example imports not updated"
fi

echo "================================="
echo "Verification complete"

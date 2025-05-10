# Migration Guide: go-core-framework → gra

This guide helps you migrate from `go-core-framework` to the renamed `gra` framework.

## Step 1: Update Dependencies

Update your `go.mod` file to use the new module:

```bash
go get github.com/lamboktulussimamora/gra
go mod tidy
```

## Step 2: Update Import Paths

Replace all import paths in your codebase:

```bash
# Find all occurrences
grep -r "github.com/lamboktulussimamora/go-core-framework" --include="*.go" .

# Replace them (Unix/Linux/macOS)
find . -type f -name "*.go" -exec sed -i '' 's|github.com/lamboktulussimamora/go-core-framework|github.com/lamboktulussimamora/gra|g' {} \;
```

## Step 3: Check for Breaking Changes

The framework was renamed without functional changes, so your code should continue to work as before.

## Step 4: Update Documentation

If you have documentation that references the old framework name, update it to use "gra" instead.

## Common Issues

If you encounter any issues after migration, please check:

1. Import paths in rarely edited files
2. Documentation references
3. Build scripts or CI/CD pipelines that may reference the old module

## Example Migration

Before:
```go
import (
    "github.com/lamboktulussimamora/go-core-framework/core"
    "github.com/lamboktulussimamora/go-core-framework/middleware"
)
```

After:
```go
import (
    "github.com/lamboktulussimamora/gra/core"
    "github.com/lamboktulussimamora/gra/middleware"
)
```

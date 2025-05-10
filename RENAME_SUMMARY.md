# Framework Rename Summary

The framework has been renamed from "go-core-framework" to "gra". The following changes were made:

## Directory Structure
- Renamed `/Users/lamboktulussimamora/Projects/go-core-framework` to `/Users/lamboktulussimamora/Projects/gra`

## Module Name
- Updated the module name in go.mod from `github.com/lamboktulussimamora/go-core-framework` to `github.com/lamboktulussimamora/gra`

## Import Paths
- Updated all import paths in the codebase:
  - core.go
  - router/router.go
  - middleware/middleware.go
  - adapter/adapter.go
  - examples/basic/main.go

## Documentation
- Updated README.md to reflect the new framework name
- Updated example code to use the new import paths

## Framework Name
- Changed all references from "Go Core Framework" to "GRA Framework"

## Next Steps
1. Push the renamed repository to GitHub:
   ```
   git remote set-url origin https://github.com/lamboktulussimamora/gra.git
   git add .
   git commit -m "Rename framework from go-core-framework to gra"
   git push -u origin main
   ```

2. Update any projects using the old framework name to use the new import paths

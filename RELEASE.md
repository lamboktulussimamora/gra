# GRA Framework Release Plan

This document outlines the steps to publish the GRA framework as a public Go module.

## 1. Pre-Release Checklist

- [ ] Review code for quality and consistency
- [ ] Ensure all files have proper documentation
- [ ] Verify that all tests pass
- [ ] Update version number in `core.go`
- [ ] Update CHANGELOG.md with release notes
- [ ] Prepare documentation

## 2. Publishing to GitHub

### 2.1 Create GitHub Repository

1. Go to [GitHub](https://github.com/new)
2. Create a new repository named "gra"
3. Keep it public
4. Do not initialize with README, .gitignore, or license (we'll push our existing code)

### 2.2 Push Code to GitHub

```bash
# Initialize Git repository if not already done
git init

# Add all files
git add .

# Commit changes
git commit -m "Initial release of GRA framework"

# Set the remote repository
git remote add origin git@github.com:lamboktulussimamora/gra.git

# Push to GitHub
git push -u origin main
```

## 3. Creating a Release

### 3.1 Create a Git Tag

```bash
# Create a tag
git tag v1.0.0

# Push the tag
git push origin v1.0.0
```

### 3.2 Create GitHub Release

1. Go to the repository on GitHub
2. Click "Releases" on the right sidebar
3. Click "Create a new release"
4. Choose the v1.0.0 tag
5. Add a title: "GRA Framework v1.0.0"
6. Add release notes (use the content from CHANGELOG.md)
7. Click "Publish release"

## 4. Verifying the Module

### 4.1 Test with a New Project

```bash
# Create a test project
mkdir -p ~/testgra
cd ~/testgra
go mod init testgra

# Add dependency
go get github.com/lamboktulussimamora/gra@v1.0.0

# Create main.go
cat > main.go << 'EOT'
package main

import (
    "fmt"
    "net/http"

    "github.com/lamboktulussimamora/gra/core"
)

func main() {
    r := core.New()
    r.GET("/", func(c *core.Context) {
        c.Success(http.StatusOK, "GRA Framework Test", nil)
    })
    fmt.Println("Server started at http://localhost:8080")
    core.Run(":8080", r)
}
EOT

# Run the test project
go run main.go
```

### 4.2 Verify with curl

```bash
curl http://localhost:8080/
```

Should return:
```json
{"status":"success","message":"GRA Framework Test"}
```

## 5. Post-Release Activities

- [ ] Update existing projects to use the new framework
- [ ] Announce the release to users/community
- [ ] Gather feedback and plan future improvements
- [ ] Set up CI/CD for the framework repository

## 6. Future Release Planning

For future releases, follow the semantic versioning pattern:

- **Patch version (v1.0.x)** - Bug fixes that don't affect the API
- **Minor version (v1.x.0)** - New features in a backward-compatible manner
- **Major version (vx.0.0)** - Breaking changes to the API

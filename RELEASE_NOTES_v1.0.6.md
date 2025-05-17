# GRA Framework v1.0.6 Release Notes

*Release Date: May 17, 2025*

## Overview

GRA Framework v1.0.6 updates the minimum required Go version from 1.21 to 1.24. This update ensures the framework takes advantage of the latest Go language features and improvements, while maintaining backward compatibility for existing applications.

## Changes

- Updated minimum Go version requirement from 1.21 to 1.24
- Updated all go.mod files in main project and example applications
- Updated documentation to reflect new Go version requirements
- Ensured compatibility with Go 1.24 across all components

## Requirements

- Go 1.24+

## Migration

For users updating from a previous version, follow these steps:

1. Update your local Go installation to version 1.24 or later
2. Update your project's go.mod file to use `go 1.24`
3. Run tests to ensure compatibility
4. If you're using any features that changed in Go 1.24, refer to the Go 1.24 release notes for migration steps

## Verification

All framework components and example applications have been tested with Go 1.24 to ensure full compatibility. No API changes were made in this release, so existing applications should work without modifications (apart from updating the Go version).

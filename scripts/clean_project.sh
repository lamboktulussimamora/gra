#!/bin/zsh

# clean_project.sh - Script to clean up the project before release
# This script removes temporary files, compiled binaries, and other artifacts

set -e

echo "Cleaning project for release..."

# First use the make clean target
if [ -f Makefile ]; then
  echo "Running make clean..."
  make clean
else
  echo "No Makefile found, skipping make clean"
fi

# Remove any potential leftover files
echo "Removing profiling and temporary files..."
find . -name "*.out" -delete
find . -name "*.test" -delete
find . -name "*.prof" -delete
find . -name "*.callgraph.out" -delete

# Remove backup files
echo "Removing backup files..."
find . -name "*.bak" -delete
find . -name "*.new" -delete
find . -name "*.tmp" -delete
find . -name "*~" -delete
find . -name "*.swp" -delete

# Remove compiled binaries in examples
echo "Removing compiled binaries in examples..."
find ./examples -type f -perm +111 -not -name "*.sh" -not -name "*.go" -not -name "*.md" -exec rm -f {} \;

# Git clean (optional, uncomment if needed)
# echo "Running git clean to remove untracked files..."
# git clean -xdf

echo "Project cleaned successfully!"

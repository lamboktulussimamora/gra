#!/bin/bash

# GRA Framework Release Script
# This script helps with publishing the GRA framework to GitHub

VERSION="v1.0.0"
REPO_URL="github.com/lamboktulussimamora/gra"
REPO_SSH="git@github.com:lamboktulussimamora/gra.git"
REPO_HTTPS="https://github.com/lamboktulussimamora/gra.git"

echo "GRA Framework Release Script"
echo "==========================="

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed. Please install git and try again."
    exit 1
fi

# Check if we're in a git repo
if [ ! -d .git ]; then
    echo "Initializing git repository..."
    git init
    
    # Check if initialization was successful
    if [ ! -d .git ]; then
        echo "Error: Failed to initialize git repository."
        exit 1
    fi
fi

# Check current remote
REMOTE_EXISTS=$(git remote -v | grep -c origin)
if [ "$REMOTE_EXISTS" -eq 0 ]; then
    echo "No remote 'origin' found. Adding remote..."
    
    # Ask for SSH or HTTPS
    read -p "Use SSH for GitHub remote? (y/N): " use_ssh
    if [[ "$use_ssh" =~ ^[Yy]$ ]]; then
        git remote add origin "$REPO_SSH"
    else
        git remote add origin "$REPO_HTTPS"
    fi
else
    echo "Remote 'origin' already exists."
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    echo "You have uncommitted changes. Please commit or stash them before proceeding."
    
    read -p "Add and commit all changes? (y/N): " add_all
    if [[ "$add_all" =~ ^[Yy]$ ]]; then
        git add .
        git commit -m "Initial commit of GRA framework v1.0.0"
    else
        echo "Please handle your uncommitted changes manually."
        exit 1
    fi
fi

# Ensure we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Currently on branch '$CURRENT_BRANCH'. Switching to 'main'..."
    
    # Check if main exists
    MAIN_EXISTS=$(git branch --list main)
    if [ -z "$MAIN_EXISTS" ]; then
        git branch -m "$CURRENT_BRANCH" main
    else
        git checkout main
    fi
fi

# Set up tag
TAG_EXISTS=$(git tag -l "$VERSION")
if [ -n "$TAG_EXISTS" ]; then
    echo "Tag $VERSION already exists."
    read -p "Delete existing tag? (y/N): " delete_tag
    if [[ "$delete_tag" =~ ^[Yy]$ ]]; then
        git tag -d "$VERSION"
        echo "Deleted existing tag $VERSION."
    else
        echo "Using existing tag."
    fi
fi

# Create tag if needed
if [ -z "$(git tag -l "$VERSION")" ]; then
    echo "Creating tag $VERSION..."
    git tag "$VERSION"
fi

# Push to GitHub
echo "Ready to push to GitHub."
read -p "Push now? (y/N): " push_now
if [[ "$push_now" =~ ^[Yy]$ ]]; then
    echo "Pushing to GitHub..."
    git push -u origin main
    git push origin --tags
    
    echo "Success! The GRA framework has been pushed to GitHub."
    echo "Tag: $VERSION"
    echo ""
    echo "Next Steps:"
    echo "1. Go to https://github.com/lamboktulussimamora/gra/releases"
    echo "2. Create a new release based on tag $VERSION"
    echo "3. Add release notes and publish"
else
    echo "Not pushing. You can push manually with:"
    echo "  git push -u origin main"
    echo "  git push origin --tags"
fi

echo ""
echo "To use this framework in other projects:"
echo "  go get $REPO_URL@$VERSION"

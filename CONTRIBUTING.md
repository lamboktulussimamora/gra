# Contributing to GRA Framework

Thank you for your interest in contributing to the GRA framework! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct:

- Be respectful and inclusive
- Focus on constructive feedback
- Be patient with new contributors
- Disagree respectfully

## How to Contribute

### Reporting Bugs

1. **Check existing issues** to see if the bug has already been reported
2. **Use the bug report template** when opening a new issue
3. **Include detailed steps to reproduce** the bug
4. **Include your environment details** (Go version, OS, etc.)
5. **Include screenshots or logs** if applicable

### Suggesting Features

1. **Check existing issues** to see if the feature has already been suggested
2. **Use the feature request template** when opening a new issue
3. **Explain the problem** your feature would solve
4. **Describe the solution** you'd like to see implemented
5. **Consider alternatives** and why they might not work as well

### Pull Requests

1. **Fork the repository**
2. **Create a branch** for your feature or bugfix
3. **Make your changes** following our code style
4. **Add or update tests** for your changes
5. **Run tests** to ensure they pass
6. **Update documentation** if necessary
7. **Submit a pull request** with a clear description of the changes

## Development Setup

### Prerequisites

- Go 1.18 or later
- Git

### Local Development

1. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/gra.git
   cd gra
   ```

2. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/lamboktulussimamora/gra.git
   ```

3. Create your feature branch:
   ```bash
   git checkout -b feature/amazing-feature
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Run tests:
   ```bash
   go test ./...
   ```

## Code Style

- Follow standard Go code style guidelines
- Document all public functions and types
- Write clear commit messages
- Keep changes focused on a single concern

## Testing

- Add tests for all new features and bug fixes
- Maintain or improve test coverage
- Run both unit and integration tests

## Documentation

- Update any relevant documentation
- Include code examples for new features
- Ensure documentation is clear and concise

## Review Process

1. Maintainers will review your PR
2. Feedback may be provided for changes
3. Once approved, your PR will be merged
4. You'll be credited as a contributor

## Getting Help

If you need assistance with the contribution process, feel free to:

- Open an issue with questions
- Reach out to maintainers
- Check existing documentation and examples

Thank you for contributing to GRA Framework!

# Migration Examples

This directory contains examples demonstrating the GRA framework's migration system capabilities.

## Files

- **`usage_example.go`** - A comprehensive demonstration of the GRA hybrid migration system showing:
  - Model definition with migration tags
  - Database connection setup
  - Migration creation and execution
  - Status checking and error handling
  - Complete migration lifecycle

- **`usage_example_test.go`** - Test cases for the migration usage example

## Running the Example

To run the migration usage example:

```bash
cd examples/migrations
go run usage_example.go
```

**Note**: Make sure you have a PostgreSQL database running and accessible with the connection string used in the example, or modify the connection string in `initializeDatabase()` function to match your setup.

## What You'll Learn

- How to define models with migration constraints using struct tags
- How to set up the GRA hybrid migrator
- How to check migration status
- How to create and apply migrations
- How to handle both automatic and interactive migration modes
- Best practices for error handling in migration workflows

## Model Examples

The example includes three related models:

- **User**: Basic user information with email uniqueness
- **Post**: Blog posts with foreign key relationship to users
- **Comment**: Comments with relationships to both posts and users

These models demonstrate various migration features including:
- Primary keys and auto-increment
- Foreign key relationships
- Unique constraints
- Default values
- Text and timestamp fields

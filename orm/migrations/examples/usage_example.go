// Package main demonstrates usage examples for the GRA migration system.
// This file provides example models and migration scenarios for documentation and testing.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq" // Import for PostgreSQL driver (required for database/sql)
)

// User represents an example user model for migration demonstration.
type User struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

// Post represents an example blog post model for migration demonstration.
type Post struct {
	ID          int64     `db:"id" migration:"primary_key,auto_increment"`
	UserID      int64     `db:"user_id" migration:"not_null,foreign_key:users.id"`
	Title       string    `db:"title" migration:"not_null,max_length:255"`
	Content     string    `db:"content" migration:"type:TEXT"`
	IsPublished bool      `db:"is_published" migration:"not_null,default:false"`
	CreatedAt   time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

// Comment represents an example comment model for migration demonstration.
type Comment struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	PostID    int64     `db:"post_id" migration:"not_null,foreign_key:posts.id"`
	UserID    int64     `db:"user_id" migration:"not_null,foreign_key:users.id"`
	Content   string    `db:"content" migration:"not_null,type:TEXT"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

func main() {
	// Database connection (adjust for your environment)
	db, err := sql.Open("postgres", "postgres://user:password@localhost/testdb?sslmode=disable")
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database: %v", closeErr)
		}
	}()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return
	}

	// Create hybrid migrator
	migrator := migrations.NewHybridMigrator(
		db,
		migrations.PostgreSQL,
		"./migrations", // migrations directory
	)

	fmt.Println("=== Hybrid Migration System Example ===")

	// Example 1: Register models (EF Core-style DbSet)
	fmt.Println("1. Registering models...")
	migrator.DbSet(&User{})    // Will use "users" table (pluralized)
	migrator.DbSet(&Post{})    // Will use "posts" table (pluralized)
	migrator.DbSet(&Comment{}) // Will use "comments" table (pluralized)
	fmt.Println("   ✓ Models registered")

	// Example 2: Check current migration status
	fmt.Println("2. Checking migration status...")
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
		return
	}

	fmt.Printf("   Applied migrations: %d\n", len(status.AppliedMigrations))
	fmt.Printf("   Pending migrations: %d\n", len(status.PendingMigrations))
	fmt.Printf("   Has pending changes: %t\n", status.HasPendingChanges)

	if status.HasPendingChanges {
		fmt.Printf("   Changes summary: %s\n", status.Summary)
	}
	fmt.Println()

	// Example 3: Create a new migration (if there are changes)
	if status.HasPendingChanges {
		fmt.Println("3. Creating migration for detected changes...")

		migrationFile, err := migrator.AddMigration(
			"initial_schema",
			migrations.ModeInteractive, // Will prompt for destructive changes
		)
		if err != nil {
			log.Printf("Failed to create migration: %v", err)
			return
		}

		fmt.Printf("   ✓ Migration created: %s\n", migrationFile.Filename)
		fmt.Printf("   Has destructive changes: %t\n", migrationFile.HasDestructiveChanges())
		fmt.Printf("   Changes count: %d\n", len(migrationFile.Changes))

		if warnings := migrationFile.GetWarnings(); len(warnings) > 0 {
			fmt.Println("   Warnings:")
			for _, warning := range warnings {
				fmt.Printf("     - %s\n", warning)
			}
		}
		fmt.Println()

		// Example 4: Apply the migration
		fmt.Println("4. Applying migrations...")
		err = migrator.ApplyMigrations(migrations.ModeAutomatic)
		if err != nil {
			// If automatic mode fails due to destructive changes, try interactive
			fmt.Printf("   Automatic mode failed: %v\n", err)
			fmt.Println("   Trying interactive mode...")

			err = migrator.ApplyMigrations(migrations.ModeInteractive)
			if err != nil {
				log.Printf("Failed to apply migrations: %v", err)
				return
			}
		}
		fmt.Println("   ✓ Migrations applied successfully")
	} else {
		fmt.Println("3. No changes detected, skipping migration creation")
	}

	// Example 5: Show final status
	fmt.Println("5. Final migration status...")
	finalStatus, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get final status: %v", err)
		return
	}

	fmt.Printf("   Applied migrations: %d\n", len(finalStatus.AppliedMigrations))
	fmt.Printf("   Pending migrations: %d\n", len(finalStatus.PendingMigrations))
	fmt.Printf("   Database is up to date: %t\n", !finalStatus.HasPendingChanges)

	fmt.Println("\n=== Example Complete ===")
}

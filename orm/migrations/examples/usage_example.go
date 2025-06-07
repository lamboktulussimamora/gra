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

// initializeDatabase establishes and tests the database connection.
func initializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://user:password@localhost/testdb?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w, and failed to close connection: %w", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// setupMigrator creates and configures the hybrid migrator with models.
func setupMigrator(db *sql.DB) *migrations.HybridMigrator {
	migrator := migrations.NewHybridMigrator(
		db,
		migrations.PostgreSQL,
		"./migrations", // migrations directory
	)

	fmt.Println("1. Registering models...")
	migrator.DbSet(&User{})    // Will use "users" table (pluralized)
	migrator.DbSet(&Post{})    // Will use "posts" table (pluralized)
	migrator.DbSet(&Comment{}) // Will use "comments" table (pluralized)
	fmt.Println("   ✓ Models registered")

	return migrator
}

// displayMigrationStatus shows the current migration status information.
func displayMigrationStatus(status *migrations.MigrationStatus) {
	if status == nil {
		fmt.Println("   [ERROR] Migration status is nil")
		return
	}
	fmt.Printf("   Applied migrations: %d\n", len(status.AppliedMigrations))
	fmt.Printf("   Pending migrations: %d\n", len(status.PendingMigrations))
	fmt.Printf("   Has pending changes: %t\n", status.HasPendingChanges)

	if status.HasPendingChanges {
		fmt.Printf("   Changes summary: %s\n", status.Summary)
	}
	fmt.Println()
}

// displayMigrationFileInfo shows information about a created migration file.
func displayMigrationFileInfo(migrationFile *migrations.MigrationFile) {
	if migrationFile == nil {
		fmt.Println("   [ERROR] Migration file is nil")
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
}

// createAndApplyMigration creates a new migration and applies it if needed.
func createAndApplyMigration(migrator *migrations.HybridMigrator, status *migrations.MigrationStatus) error {
	if migrator == nil {
		return fmt.Errorf("migrator is nil")
	}
	if status == nil {
		return fmt.Errorf("migration status is nil")
	}
	if !status.HasPendingChanges {
		fmt.Println("3. No changes detected, skipping migration creation")
		return nil
	}

	fmt.Println("3. Creating migration for detected changes...")
	migrationFile, err := migrator.AddMigration(
		"initial_schema",
		migrations.ModeInteractive, // Will prompt for destructive changes
	)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	displayMigrationFileInfo(migrationFile)

	// Apply the migration
	fmt.Println("4. Applying migrations...")
	err = migrator.ApplyMigrations(migrations.ModeAutomatic)
	if err != nil {
		// If automatic mode fails due to destructive changes, try interactive
		fmt.Printf("   Automatic mode failed: %v\n", err)
		fmt.Println("   Trying interactive mode...")

		err = migrator.ApplyMigrations(migrations.ModeInteractive)
		if err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}
	fmt.Println("   ✓ Migrations applied successfully")

	return nil
}

// showFinalStatus displays the final migration status after all operations.
func showFinalStatus(migrator *migrations.HybridMigrator) error {
	if migrator == nil {
		return fmt.Errorf("migrator is nil")
	}
	fmt.Println("5. Final migration status...")
	finalStatus, err := migrator.GetMigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get final status: %w", err)
	}

	fmt.Printf("   Applied migrations: %d\n", len(finalStatus.AppliedMigrations))
	fmt.Printf("   Pending migrations: %d\n", len(finalStatus.PendingMigrations))
	fmt.Printf("   Database is up to date: %t\n", !finalStatus.HasPendingChanges)

	return nil
}

func main() {
	// Initialize database connection
	db, err := initializeDatabase()
	if err != nil {
		log.Printf("Database initialization failed: %v", err)
		return
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database: %v", closeErr)
		}
	}()

	fmt.Println("=== Hybrid Migration System Example ===")

	// Setup migrator and register models
	migrator := setupMigrator(db)

	// Check current migration status
	fmt.Println("2. Checking migration status...")
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
		return
	}

	displayMigrationStatus(status)

	// Create and apply migrations if needed
	if err := createAndApplyMigration(migrator, status); err != nil {
		log.Printf("Migration process failed: %v", err)
		return
	}

	// Show final status
	if err := showFinalStatus(migrator); err != nil {
		log.Printf("Failed to show final status: %v", err)
		return
	}

	fmt.Println("\n=== Example Complete ===")
}

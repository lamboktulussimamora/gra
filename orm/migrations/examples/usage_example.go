package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Example models - these would typically be in your models package
type User struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

type Post struct {
	ID          int64     `db:"id" migration:"primary_key,auto_increment"`
	UserID      int64     `db:"user_id" migration:"not_null,foreign_key:users.id"`
	Title       string    `db:"title" migration:"not_null,max_length:255"`
	Content     string    `db:"content" migration:"type:TEXT"`
	IsPublished bool      `db:"is_published" migration:"not_null,default:false"`
	CreatedAt   time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

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

// Example of how to use the system in different scenarios
func exampleScenarios() {
	// This function shows various usage patterns
	// (not called in main, just for documentation)

	var migrator *migrations.HybridMigrator
	// ... setup migrator ...

	// Scenario 1: Development workflow
	developmentWorkflow(migrator)

	// Scenario 2: Production deployment
	productionDeployment(migrator)

	// Scenario 3: Rollback scenario
	rollbackScenario(migrator)
}

func developmentWorkflow(migrator *migrations.HybridMigrator) {
	fmt.Println("=== Development Workflow ===")

	// 1. Developer adds a new field to User model
	// (This would be done by modifying the struct)

	// 2. Check what changes would be generated
	status, _ := migrator.GetMigrationStatus()
	if status.HasPendingChanges {
		fmt.Printf("Changes detected: %s\n", status.Summary)

		// 3. Generate migration in interactive mode (allows review)
		migration, err := migrator.AddMigration("add_user_profile_fields", migrations.ModeInteractive)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Migration created: %s\n", migration.Name)

		// 4. Apply migration automatically (since it's non-destructive)
		err = migrator.ApplyMigrations(migrations.ModeAutomatic)
		if err != nil {
			fmt.Printf("Error applying migration: %v\n", err)
		}
	}
}

func productionDeployment(migrator *migrations.HybridMigrator) {
	fmt.Println("=== Production Deployment ===")

	// 1. In production, only apply pre-generated migrations
	// Never generate new migrations in production

	// 2. Apply pending migrations with careful mode
	err := migrator.ApplyMigrations(migrations.ModeInteractive)
	if err != nil {
		fmt.Printf("Production migration failed: %v\n", err)
		// In production, you might want to halt deployment here
	}

	// 3. Verify database state
	status, _ := migrator.GetMigrationStatus()
	if status.HasPendingChanges {
		fmt.Println("WARNING: Production database has unexpected pending changes!")
	}
}

func rollbackScenario(migrator *migrations.HybridMigrator) {
	fmt.Println("=== Rollback Scenario ===")

	// 1. Check current status
	status, _ := migrator.GetMigrationStatus()
	fmt.Printf("Applied migrations: %d\n", len(status.AppliedMigrations))

	// 2. Rollback last migration if needed
	if len(status.AppliedMigrations) > 0 {
		err := migrator.RevertMigration()
		if err != nil {
			fmt.Printf("Rollback failed: %v\n", err)
		} else {
			fmt.Println("Successfully rolled back last migration")
		}
	}
}

// Example of advanced model with complex relationships
type AdvancedUser struct {
	// Base fields
	ID       int64  `db:"id" migration:"primary_key,auto_increment"`
	Email    string `db:"email" migration:"unique,not_null,max_length:255,index"`
	Username string `db:"username" migration:"unique,not_null,max_length:50,index"`

	// Profile information
	FirstName string `db:"first_name" migration:"max_length:100"`
	LastName  string `db:"last_name" migration:"max_length:100"`
	Bio       string `db:"bio" migration:"type:TEXT"`
	AvatarURL string `db:"avatar_url" migration:"max_length:500"`

	// Status and permissions
	IsActive   bool   `db:"is_active" migration:"not_null,default:true,index"`
	IsVerified bool   `db:"is_verified" migration:"not_null,default:false"`
	Role       string `db:"role" migration:"not_null,default:'user',max_length:50,index"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP,index"`
	UpdatedAt time.Time  `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	LastLogin *time.Time `db:"last_login" migration:"nullable"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" migration:"nullable,index"`
}

// Example showing how the system handles different change types
func changeTypeExamples() {
	/*
		Change Type Examples:

		1. CreateTable - Adding AdvancedUser model
		   - Generates CREATE TABLE with all columns, indexes, constraints

		2. AddColumn - Adding new field to existing model
		   - Generates ALTER TABLE ADD COLUMN

		3. DropColumn - Removing field from model
		   - Generates ALTER TABLE DROP COLUMN (destructive)

		4. AlterColumn - Changing field properties
		   - Type change: migration:"type:TEXT" -> migration:"max_length:500"
		   - Nullable change: not_null -> nullable
		   - Default change: default:true -> default:false

		5. CreateIndex - Adding index tag to field
		   - migration:"max_length:255" -> migration:"max_length:255,index"

		6. DropIndex - Removing index tag from field
		   - migration:"max_length:255,index" -> migration:"max_length:255"

		7. CreateConstraint - Adding foreign key or unique constraint
		   - migration:"foreign_key:users.id"
		   - migration:"unique"

		All changes are detected automatically by comparing current model state
		with database schema, similar to Entity Framework Core.
	*/
}

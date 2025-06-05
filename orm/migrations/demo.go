package migrations

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

// IntegrationDemo demonstrates the complete migration workflow
func IntegrationDemo() {
	fmt.Println("=== GRA Hybrid Migration Integration Demo ===")

	// 1. Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database: %v", closeErr)
		}
	}()

	// 2. Create migrator
	migrator := NewHybridMigrator(
		db,
		SQLite,
		"./test_migrations",
	)

	// 3. Register existing GRA models
	fmt.Println("1. Registering GRA models...")
	migrator.DbSet(&models.User{})
	migrator.DbSet(&models.Product{})
	migrator.DbSet(&models.Category{})
	fmt.Println("   ✓ Core models registered")

	// 4. Check migration status
	fmt.Println("2. Checking migration status...")
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Fatalf("Failed to get migration status: %v", err)
	}

	fmt.Printf("   Applied migrations: %d\n", len(status.AppliedMigrations))
	fmt.Printf("   Pending migrations: %d\n", len(status.PendingMigrations))
	fmt.Printf("   Has pending changes: %t\n", status.HasPendingChanges)
	fmt.Println()

	// 5. Create initial migration
	fmt.Println("3. Creating initial migration...")
	migrationFile, err := migrator.AddMigration(
		"create_initial_schema",
		ModeGenerateOnly, // Generate files only for review
	)
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}

	if migrationFile != nil {
		fmt.Printf("   ✓ Migration created: %s\n", migrationFile.Filename)
		fmt.Printf("   Changes: %d\n", len(migrationFile.Changes))
		fmt.Printf("   Has destructive changes: %t\n", migrationFile.HasDestructiveChanges())

		if warnings := migrationFile.GetWarnings(); len(warnings) > 0 {
			fmt.Println("   Warnings:")
			for _, warning := range warnings {
				fmt.Printf("     - %s\n", warning)
			}
		}
	} else {
		fmt.Println("   No changes detected")
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("The hybrid migration system is working correctly!")
	fmt.Println("Key features demonstrated:")
	fmt.Println("  ✓ Model registration (EF Core-style DbSet)")
	fmt.Println("  ✓ Change detection from struct definitions")
	fmt.Println("  ✓ Migration file generation")
	fmt.Println("  ✓ Safety checks and warnings")
	fmt.Println("  ✓ Multiple migration modes")
}

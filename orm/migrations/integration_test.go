package migrations

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

// IntegrationTest demonstrates the complete migration workflow
func IntegrationTest() {
	fmt.Println("=== GRA Hybrid Migration Integration Test ===")

	// 1. Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return
	}

	// 2. Create migrator
	migrator := NewHybridMigrator(
		db,
		SQLite,
		"./test_migrations",
	)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database: %v", closeErr)
		}
	}()

	// 3. Register existing GRA models
	fmt.Println("1. Registering GRA models...")
	migrator.DbSet(&models.User{})
	migrator.DbSet(&models.Product{})
	migrator.DbSet(&models.Category{})
	migrator.DbSet(&models.Order{})
	migrator.DbSet(&models.OrderItem{})
	migrator.DbSet(&models.Role{})
	migrator.DbSet(&models.UserRole{})
	migrator.DbSet(&models.Review{})
	fmt.Println("   ✓ All models registered")
	fmt.Println()

	// 4. Check migration status
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
		fmt.Printf("   Changes detected: %s\n", status.Summary)
	}
	fmt.Println()

	// 5. Create initial migration
	fmt.Println("3. Creating initial migration...")
	migrationFile, err := migrator.AddMigration(
		"create_initial_schema",
		ModeGenerateOnly, // Generate files only for review
	)
	if err != nil {
		log.Printf("Failed to create migration: %v", err)
		return
	}

	if migrationFile != nil {
		fmt.Printf("   ✓ Migration created: %s\n", migrationFile.Filename)
		fmt.Printf("   Changes: %d\n", len(migrationFile.Changes))
		fmt.Printf("   Has destructive changes: %t\n", migrationFile.HasDestructiveChanges())

		// Show change summary
		fmt.Println("   Change summary:")
		for _, change := range migrationFile.Changes {
			fmt.Printf("     - %s: %s.%s\n", change.Type, change.TableName, change.ColumnName)
		}

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

	// 6. Apply migrations
	fmt.Println("4. Applying migrations...")
	err = migrator.ApplyMigrations(ModeAutomatic)
	if err != nil {
		log.Printf("Failed to apply migrations: %v", err)
	} else {
		fmt.Println("   ✓ Migrations applied successfully")
	}
	fmt.Println()

	// 7. Verify database schema
	fmt.Println("5. Verifying database schema...")
	tables := []string{"users", "products", "categories", "orders", "order_items", "roles", "user_roles", "reviews"}

	// Get current schema to check table existence
	schema, err := migrator.inspector.GetCurrentSchema()
	if err != nil {
		fmt.Printf("   Error getting current schema: %v\n", err)
	} else {
		for _, table := range tables {
			if _, exists := schema[table]; exists {
				fmt.Printf("   ✓ Table '%s' exists\n", table)
			} else {
				fmt.Printf("   ✗ Table '%s' missing\n", table)
			}
		}
	}
	fmt.Println()

	// 8. Test adding a new field (simulate model change)
	fmt.Println("6. Testing model evolution...")

	// Create a modified user model to simulate development
	type ModifiedUser struct {
		models.User
		PhoneNumber *string `db:"phone_number" migration:"max_length:20"`
		IsVerified  bool    `db:"is_verified" migration:"not_null,default:false"`
	}

	// Register the modified model
	migrator.DbSet(&ModifiedUser{})

	// Check for new changes
	newStatus, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get updated status: %v", err)
		return
	}

	if newStatus.HasPendingChanges {
		fmt.Printf("   ✓ Changes detected for modified model\n")
		fmt.Printf("   Pending changes: %s\n", newStatus.Summary)

		// Create evolution migration
		evolutionMigration, err := migrator.AddMigration(
			"add_user_fields",
			ModeGenerateOnly,
		)
		if err != nil {
			log.Printf("Failed to create evolution migration: %v", err)
		} else {
			fmt.Printf("   ✓ Evolution migration created: %s\n", evolutionMigration.Filename)
		}
	} else {
		fmt.Println("   No additional changes detected")
	}
	fmt.Println()

	// 9. Final status
	fmt.Println("7. Final status...")
	finalStatus, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get final status: %v", err)
		return
	}

	fmt.Printf("   Total applied migrations: %d\n", len(finalStatus.AppliedMigrations))
	fmt.Printf("   Database is current: %t\n", !finalStatus.HasPendingChanges)

	fmt.Println("=== Integration Test Complete ===")
}

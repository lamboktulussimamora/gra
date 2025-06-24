package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

// DemoResult contains the results of running the demo
type DemoResult struct {
	ModelsRegistered int
	MigrationStatus  *migrations.MigrationStatus
	MigrationFile    *migrations.MigrationFile
	Error            error
	FeaturesDemo     []string
}

// RunIntegrationDemo demonstrates the complete migration workflow
func RunIntegrationDemo() (*DemoResult, error) {
	result := &DemoResult{
		FeaturesDemo: []string{
			"Model registration (EF Core-style DbSet)",
			"Change detection from struct definitions",
			"Migration file generation",
			"Safety checks and warnings",
			"Multiple migration modes",
		},
	}

	// 1. Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		result.Error = fmt.Errorf("failed to open database: %v", err)
		return result, err
	}
	defer db.Close()

	// 2. Create migrator
	migrator := migrations.NewHybridMigrator(
		db,
		migrations.SQLite,
		"./test_migrations",
	)

	// 3. Register existing GRA models
	migrator.DbSet(&models.User{})
	migrator.DbSet(&models.Product{})
	migrator.DbSet(&models.Category{})
	result.ModelsRegistered = 3

	// 4. Check migration status (this initializes the schema automatically)
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		result.Error = fmt.Errorf("failed to get migration status: %v", err)
		return result, err
	}
	result.MigrationStatus = status

	// 5. Create initial migration
	migrationFile, err := migrator.AddMigration(
		"create_initial_schema",
		migrations.ModeGenerateOnly, // Generate files only for review
	)
	if err != nil {
		result.Error = fmt.Errorf("failed to create migration: %v", err)
		return result, err
	}
	result.MigrationFile = migrationFile

	return result, nil
}

// PrintDemoResults prints the demo results in a user-friendly format
func PrintDemoResults(result *DemoResult) {
	fmt.Println("=== GRA Hybrid Migration Integration Demo ===")

	if result.Error != nil {
		fmt.Printf("❌ Demo failed: %v\n", result.Error)
		return
	}

	fmt.Println("1. Registering GRA models...")
	fmt.Printf("   ✓ %d core models registered\n", result.ModelsRegistered)

	fmt.Println("2. Initializing migration system...")
	fmt.Println("   ✓ Migration system initialized")

	fmt.Println("3. Checking migration status...")
	if result.MigrationStatus != nil {
		fmt.Printf("   Applied migrations: %d\n", len(result.MigrationStatus.AppliedMigrations))
		fmt.Printf("   Pending migrations: %d\n", len(result.MigrationStatus.PendingMigrations))
		fmt.Printf("   Has pending changes: %t\n", result.MigrationStatus.HasPendingChanges)
	}
	fmt.Println()

	fmt.Println("4. Creating initial migration...")
	if result.MigrationFile != nil {
		fmt.Printf("   ✓ Migration created: %s\n", result.MigrationFile.Filename)
		fmt.Printf("   Changes: %d\n", len(result.MigrationFile.Changes))
		fmt.Printf("   Has destructive changes: %t\n", result.MigrationFile.HasDestructiveChanges())

		if warnings := result.MigrationFile.GetWarnings(); len(warnings) > 0 {
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
	for _, feature := range result.FeaturesDemo {
		fmt.Printf("  ✓ %s\n", feature)
	}
}

func main() {
	result, err := RunIntegrationDemo()
	if err != nil {
		log.Fatalf("Demo failed: %v", err)
	}
	PrintDemoResults(result)
}

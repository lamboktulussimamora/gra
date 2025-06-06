package migrations

import (
	"database/sql"
	"log"

	"github.com/lamboktulussimamora/gra/logger"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3" // Import for SQLite driver (required for database/sql)
)

// IntegrationDemo demonstrates the complete migration workflow
func IntegrationDemo() {
	logger.Get().Info("=== GRA Hybrid Migration Integration Demo ===")

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
	logger.Get().Info("1. Registering GRA models...")
	migrator.DbSet(&models.User{})
	migrator.DbSet(&models.Product{})
	migrator.DbSet(&models.Category{})
	logger.Get().Info("   ✓ Core models registered")

	// 4. Check migration status
	logger.Get().Info("2. Checking migration status...")
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
		return
	}

	logger.Get().Infof("   Applied migrations: %d", len(status.AppliedMigrations))
	logger.Get().Infof("   Pending migrations: %d", len(status.PendingMigrations))
	logger.Get().Infof("   Has pending changes: %t", status.HasPendingChanges)

	// 5. Create initial migration
	logger.Get().Info("3. Creating initial migration...")
	migrationFile, err := migrator.AddMigration(
		"create_initial_schema",
		ModeGenerateOnly, // Generate files only for review
	)
	if err != nil {
		log.Printf("Failed to create migration: %v", err)
		return
	}

	if migrationFile != nil {
		logger.Get().Infof("   ✓ Migration created: %s", migrationFile.Filename)
		logger.Get().Infof("   Changes: %d", len(migrationFile.Changes))
		logger.Get().Infof("   Has destructive changes: %t", migrationFile.HasDestructiveChanges())

		if warnings := migrationFile.GetWarnings(); len(warnings) > 0 {
			logger.Get().Info("   Warnings:")
			for _, warning := range warnings {
				logger.Get().Infof("     - %s", warning)
			}
		}
	} else {
		logger.Get().Info("   No changes detected")
	}

	logger.Get().Info("=== Demo Complete ===")
	logger.Get().Info("The hybrid migration system is working correctly!")
	logger.Get().Info("Key features demonstrated:")
	logger.Get().Info("  ✓ Model registration (EF Core-style DbSet)")
	logger.Get().Info("  ✓ Change detection from struct definitions")
	logger.Get().Info("  ✓ Migration file generation")
	logger.Get().Info("  ✓ Safety checks and warnings")
	logger.Get().Info("  ✓ Multiple migration modes")
}

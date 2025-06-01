// EF Core-like Migration CLI Tool for GRA Framework
// Provides commands similar to Entity Framework Core migration commands
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Constants for error messages and formatting
const (
	ErrorFailedToGetHistory = "❌ Failed to get migration history:"
	FormatMigrationLine     = "   %s\n"
	TimeFormat              = "2006-01-02 15:04:05"
)

type CLIConfig struct {
	ConnectionString string
	MigrationsDir    string
	Verbose          bool
}

func main() {
	config := CLIConfig{}

	// Define CLI flags
	flag.StringVar(&config.ConnectionString, "connection", "", "Database connection string")
	flag.StringVar(&config.MigrationsDir, "migrations-dir", "./migrations", "Directory to store migration files")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Get command
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Handle help command before database setup
	if command == "help" || command == "-h" || command == "--help" {
		printUsage()
		return
	}

	// Setup database connection
	if config.ConnectionString == "" {
		config.ConnectionString = os.Getenv("DATABASE_URL")
		if config.ConnectionString == "" {
			log.Fatal("❌ Database connection string required. Use -connection flag or DATABASE_URL env var")
		}
	}

	// Detect database driver
	var driverName string
	if strings.HasPrefix(config.ConnectionString, "postgres://") || strings.Contains(config.ConnectionString, "user=") {
		driverName = "postgres"
	} else if strings.HasSuffix(config.ConnectionString, ".db") || strings.Contains(config.ConnectionString, "sqlite") {
		driverName = "sqlite3"
	} else {
		// Default to postgres for backward compatibility
		driverName = "postgres"
	}

	db, err := sql.Open(driverName, config.ConnectionString)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Create migration manager
	migrationConfig := migrations.DefaultEFMigrationConfig()
	if config.Verbose {
		migrationConfig.Logger = log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags)
	} else {
		migrationConfig.Logger = log.New(os.Stderr, "", 0)
	}

	manager := migrations.NewEFMigrationManager(db, migrationConfig)

	// Initialize schema if needed
	if err := manager.EnsureSchema(); err != nil {
		log.Fatal("❌ Failed to initialize migration schema:", err)
	}

	// Execute command
	switch command {
	case "add-migration", "add":
		addMigration(manager, args[1:], config)
	case "update-database", "update":
		updateDatabase(manager, args[1:], config)
	case "get-migration", "list":
		getMigrations(manager, config)
	case "rollback":
		rollbackMigration(manager, args[1:], config)
	case "status":
		showStatus(manager, config)
	case "script":
		generateScript(manager, args[1:], config)
	case "remove-migration", "remove":
		removeMigration(manager, args[1:], config)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// addMigration implements Add-Migration command
func addMigration(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	if len(args) == 0 {
		log.Fatal("❌ Migration name required. Usage: add-migration <name>")
	}

	name := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

	fmt.Printf("🔧 Creating migration: %s\n", name)

	// For now, create empty migration that user can fill
	upSQL := fmt.Sprintf("-- Migration: %s\n-- Description: %s\n-- TODO: Add your SQL here\n\n", name, description)
	downSQL := fmt.Sprintf("-- Rollback for: %s\n-- TODO: Add rollback SQL here\n\n", name)

	migration := manager.AddMigration(name, description, upSQL, downSQL)

	// Save migration to file
	if err := saveMigrationToFile(migration, config.MigrationsDir); err != nil {
		log.Fatal("❌ Failed to save migration file:", err)
	}

	fmt.Printf("✅ Migration created: %s\n", migration.ID)
	fmt.Printf("📁 File: %s/%s.sql\n", config.MigrationsDir, migration.ID)
	fmt.Println("📝 Edit the migration file and run 'update-database' to apply")
}

// updateDatabase implements Update-Database command
func updateDatabase(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	fmt.Println("🚀 Updating database...")

	var targetMigration []string
	if len(args) > 0 {
		targetMigration = []string{args[0]}
		fmt.Printf("🎯 Target migration: %s\n", args[0])
	}

	if err := manager.UpdateDatabase(targetMigration...); err != nil {
		log.Fatal("❌ Failed to update database:", err)
	}

	fmt.Println("✅ Database updated successfully!")
}

// getMigrations implements Get-Migration command
func getMigrations(manager *migrations.EFMigrationManager, config CLIConfig) {
	fmt.Println("📋 Migration History:")
	fmt.Println("====================")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("❌ Failed to get migration history:", err)
	}

	if len(history.Applied) == 0 && len(history.Pending) == 0 && len(history.Failed) == 0 {
		fmt.Println("📭 No migrations found")
		return
	}

	// Applied migrations
	if len(history.Applied) > 0 {
		fmt.Printf("\n✅ Applied Migrations (%d):\n", len(history.Applied))
		for _, m := range history.Applied {
			fmt.Printf(FormatMigrationLine, formatMigrationInfo(m, "applied"))
		}
	}

	// Pending migrations
	if len(history.Pending) > 0 {
		fmt.Printf("\n⏳ Pending Migrations (%d):\n", len(history.Pending))
		for _, m := range history.Pending {
			fmt.Printf(FormatMigrationLine, formatMigrationInfo(m, "pending"))
		}
	}

	// Failed migrations
	if len(history.Failed) > 0 {
		fmt.Printf("\n❌ Failed Migrations (%d):\n", len(history.Failed))
		for _, m := range history.Failed {
			fmt.Printf(FormatMigrationLine, formatMigrationInfo(m, "failed"))
		}
	}

	fmt.Printf("\n📊 Summary: %d applied, %d pending, %d failed\n",
		len(history.Applied), len(history.Pending), len(history.Failed))
}

// rollbackMigration implements rollback functionality
func rollbackMigration(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	if len(args) == 0 {
		log.Fatal("❌ Target migration required. Usage: rollback <migration-name-or-id>")
	}

	target := args[0]
	fmt.Printf("⏪ Rolling back to migration: %s\n", target)

	if err := manager.RollbackMigration(target); err != nil {
		log.Fatal("❌ Failed to rollback migration:", err)
	}

	fmt.Println("✅ Rollback completed successfully!")
}

// showStatus shows current migration status
func showStatus(manager *migrations.EFMigrationManager, config CLIConfig) {
	fmt.Println("📊 Migration Status:")
	fmt.Println("===================")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("❌ Failed to get migration status:", err)
	}

	fmt.Printf("Database: %s\n", extractDBName(config.ConnectionString))
	fmt.Printf("Applied:  %d migrations\n", len(history.Applied))
	fmt.Printf("Pending:  %d migrations\n", len(history.Pending))
	fmt.Printf("Failed:   %d migrations\n", len(history.Failed))

	if len(history.Applied) > 0 {
		latest := history.Applied[len(history.Applied)-1]
		fmt.Printf("Latest:   %s (%s)\n", latest.ID, latest.AppliedAt.Format("2006-01-02 15:04:05"))
	}

	if len(history.Pending) > 0 {
		fmt.Printf("Next:     %s\n", history.Pending[0].ID)
	}
}

// generateScript generates SQL script for migrations
func generateScript(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	fmt.Println("📜 Generating migration script...")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("❌ Failed to get migration history:", err)
	}

	if len(history.Pending) == 0 {
		fmt.Println("📭 No pending migrations to script")
		return
	}

	var migrations []migrations.Migration
	if len(args) > 0 {
		// Script to specific migration
		target := args[0]
		for _, m := range history.Pending {
			migrations = append(migrations, m)
			if m.ID == target || m.Name == target {
				break
			}
		}
	} else {
		// Script all pending migrations
		migrations = history.Pending
	}

	// Generate script
	fmt.Println("-- Generated Migration Script")
	fmt.Printf("-- Generated at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("-- Migrations: %d\n", len(migrations))
	fmt.Println("-- ==========================================")

	for i, migration := range migrations {
		fmt.Printf("\n-- Migration %d: %s\n", i+1, migration.ID)
		fmt.Printf("-- Description: %s\n", migration.Description)
		fmt.Println("-- ------------------------------------------")
		fmt.Println(migration.UpSQL)
	}

	fmt.Println("\n-- End of migration script")
}

// removeMigration removes the last migration
func removeMigration(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	fmt.Println("🗑️  Removing last migration...")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Fatal("❌ Failed to get migration history:", err)
	}

	if len(history.Pending) == 0 {
		log.Fatal("❌ No pending migrations to remove")
	}

	// Remove the last pending migration
	lastMigration := history.Pending[len(history.Pending)-1]

	fmt.Printf("🗑️  Removing migration: %s\n", lastMigration.ID)

	// TODO: Implement removal logic in EFMigrationManager
	fmt.Println("⚠️  Note: Migration removal from database not yet implemented")
	fmt.Printf("📁 Please manually delete: %s/%s.sql\n", config.MigrationsDir, lastMigration.ID)
}

// Helper functions

func formatMigrationInfo(m migrations.Migration, status string) string {
	var statusIcon string
	switch status {
	case "applied":
		statusIcon = "✅"
	case "pending":
		statusIcon = "⏳"
	case "failed":
		statusIcon = "❌"
	default:
		statusIcon = "❓"
	}

	result := fmt.Sprintf("%s %s", statusIcon, m.ID)
	if !m.AppliedAt.IsZero() {
		result += fmt.Sprintf(" (%s)", m.AppliedAt.Format("2006-01-02 15:04:05"))
	}
	if m.Description != "" {
		result += fmt.Sprintf(" - %s", m.Description)
	}
	return result
}

func extractDBName(connectionString string) string {
	parts := strings.Split(connectionString, "/")
	if len(parts) > 0 {
		dbPart := parts[len(parts)-1]
		if idx := strings.Index(dbPart, "?"); idx > -1 {
			return dbPart[:idx]
		}
		return dbPart
	}
	return "unknown"
}

func saveMigrationToFile(migration *migrations.Migration, dir string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create migration file
	filename := fmt.Sprintf("%s/%s.sql", dir, migration.ID)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write migration content
	content := fmt.Sprintf(`-- Migration: %s
-- Description: %s
-- Created: %s
-- Version: %d

-- UP Migration
%s

-- DOWN Migration (for rollback)
-- %s
`,
		migration.Name,
		migration.Description,
		time.Now().Format("2006-01-02 15:04:05"),
		migration.Version,
		migration.UpSQL,
		migration.DownSQL,
	)

	_, err = file.WriteString(content)
	return err
}

func printUsage() {
	fmt.Println(`🚀 GRA Entity Framework Core-like Migration Tool`)
	fmt.Println(`===============================================`)
	fmt.Println()
	fmt.Println(`USAGE:`)
	fmt.Println(`  ef-migrate [options] <command> [arguments]`)
	fmt.Println()
	fmt.Println(`OPTIONS:`)
	fmt.Println(`  -connection <string>    Database connection string`)
	fmt.Println(`  -migrations-dir <path>  Directory for migration files (default: ./migrations)`)
	fmt.Println(`  -verbose               Enable verbose logging`)
	fmt.Println()
	fmt.Println(`COMMANDS:`)
	fmt.Println()
	fmt.Println(`📝 Migration Management:`)
	fmt.Println(`  add-migration <name> [description]  Create a new migration`)
	fmt.Println(`  update-database [target]            Apply pending migrations`)
	fmt.Println(`  rollback <target>                   Rollback to specific migration`)
	fmt.Println(`  remove-migration                    Remove the last migration`)
	fmt.Println()
	fmt.Println(`📋 Information:`)
	fmt.Println(`  get-migration                       List all migrations`)
	fmt.Println(`  status                              Show migration status`)
	fmt.Println(`  script [target]                     Generate SQL script`)
	fmt.Println()
	fmt.Println(`EXAMPLES:`)
	fmt.Println(`  # Create a new migration`)
	fmt.Println(`  ef-migrate add-migration CreateUsersTable "Initial user table"`)
	fmt.Println()
	fmt.Println(`  # Apply all pending migrations`)
	fmt.Println(`  ef-migrate update-database`)
	fmt.Println()
	fmt.Println(`  # Apply migrations up to a specific one`)
	fmt.Println(`  ef-migrate update-database CreateUsersTable`)
	fmt.Println()
	fmt.Println(`  # Rollback to a specific migration`)
	fmt.Println(`  ef-migrate rollback InitialMigration`)
	fmt.Println()
	fmt.Println(`  # View migration status`)
	fmt.Println(`  ef-migrate status`)
	fmt.Println()
	fmt.Println(`  # List all migrations`)
	fmt.Println(`  ef-migrate get-migration`)
	fmt.Println()
	fmt.Println(`ENVIRONMENT:`)
	fmt.Println(`  DATABASE_URL    Default database connection string`)
	fmt.Println()
	fmt.Println(`📚 More info: https://github.com/your-org/gra/docs/migrations`)
}

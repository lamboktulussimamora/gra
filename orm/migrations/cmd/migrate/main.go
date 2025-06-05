package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Configuration for the migration CLI
type Config struct {
	DatabaseURL   string
	Driver        string
	MigrationsDir string
	ModelsDir     string
}

func main() {
	var config Config
	var command string

	// Define command line flags
	flag.StringVar(&config.DatabaseURL, "db", "", "Database connection URL")
	flag.StringVar(&config.Driver, "driver", "postgres", "Database driver (postgres, mysql, sqlite)")
	flag.StringVar(&config.MigrationsDir, "migrations-dir", "./migrations", "Directory for migration files")
	flag.StringVar(&config.ModelsDir, "models-dir", "./models", "Directory containing model files")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  add <name>      Create a new migration with the given name\n")
		fmt.Fprintf(os.Stderr, "  apply           Apply all pending migrations\n")
		fmt.Fprintf(os.Stderr, "  revert          Revert the last applied migration\n")
		fmt.Fprintf(os.Stderr, "  status          Show migration status\n")
		fmt.Fprintf(os.Stderr, "  generate <name> Generate migration script only (no database changes)\n")
		fmt.Fprintf(os.Stderr, "  force <name>    Create migration with force destructive mode\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Get command
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}
	command = flag.Arg(0)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Connect to database
	db, err := connectDatabase(&config)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Create migrator
	driver := getDriver(config.Driver)
	migrator := migrations.NewHybridMigrator(db, driver, config.MigrationsDir)

	// Register models (this would typically be done automatically by scanning the models directory)
	if err := registerModels(migrator, config.ModelsDir); err != nil {
		log.Fatalf("Model registration error: %v", err)
	}

	// Execute command
	switch command {
	case "add":
		err = cmdAddMigration(migrator, flag.Args()[1:])
	case "apply":
		err = cmdApplyMigrations(migrator, flag.Args()[1:])
	case "revert":
		err = cmdRevertMigration(migrator)
	case "status":
		err = cmdMigrationStatus(migrator)
	case "generate":
		err = cmdGenerateMigration(migrator, flag.Args()[1:])
	case "force":
		err = cmdForceMigration(migrator, flag.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Command error: %v", err)
	}
}

// validateConfig validates the CLI configuration
func validateConfig(config *Config) error {
	if config.DatabaseURL == "" {
		return fmt.Errorf("database URL is required (use -db flag)")
	}

	if config.Driver == "" {
		config.Driver = "postgres"
	}

	if config.MigrationsDir == "" {
		config.MigrationsDir = "./migrations"
	}

	if config.ModelsDir == "" {
		config.ModelsDir = "./models"
	}

	return nil
}

// connectDatabase establishes a database connection
func connectDatabase(config *Config) (*sql.DB, error) {
	db, err := sql.Open(config.Driver, config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getDriver converts string driver name to migrations.DatabaseDriver
func getDriver(driverName string) migrations.DatabaseDriver {
	switch driverName {
	case "postgres", "postgresql":
		return migrations.PostgreSQL
	case "mysql":
		return migrations.MySQL
	case "sqlite", "sqlite3":
		return migrations.SQLite
	default:
		log.Fatalf("Unsupported driver: %s", driverName)
		return ""
	}
}

// registerModels registers models with the migrator
// In a real implementation, this would scan the models directory and register all found models
func registerModels(migrator *migrations.HybridMigrator, modelsDir string) error {
	// This is a placeholder implementation
	// In practice, you would:
	// 1. Scan the models directory for Go files
	// 2. Parse the Go files to find struct definitions with migration tags
	// 3. Register each model with the migrator

	fmt.Printf("Note: Model registration from %s not implemented in this example\n", modelsDir)
	fmt.Printf("In practice, you would call migrator.DbSet() for each model\n")

	// Example model registration (you would replace this with actual model scanning):
	// migrator.DbSet(&User{}, "users")
	// migrator.DbSet(&Post{}, "posts")

	return nil
}

// cmdAddMigration creates a new migration
func cmdAddMigration(migrator *migrations.HybridMigrator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("migration name is required")
	}

	name := args[0]
	mode := migrations.ModeInteractive

	fmt.Printf("Creating migration: %s\n", name)

	migrationFile, err := migrator.AddMigration(name, mode)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	fmt.Printf("Migration created: %s\n", migrationFile.Filename)

	if migrationFile.HasDestructiveChanges() {
		fmt.Printf("⚠️  WARNING: This migration contains destructive changes\n")
	}

	if len(migrationFile.GetWarnings()) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warning := range migrationFile.GetWarnings() {
			fmt.Printf("  - %s\n", warning)
		}
	}

	return nil
}

// cmdApplyMigrations applies pending migrations
func cmdApplyMigrations(migrator *migrations.HybridMigrator, args []string) error {
	mode := migrations.ModeInteractive

	// Check for force flag
	for _, arg := range args {
		if arg == "--force" {
			mode = migrations.ModeForceDestructive
			break
		}
		if arg == "--auto" {
			mode = migrations.ModeAutomatic
			break
		}
	}

	fmt.Printf("Applying migrations in %s mode...\n", mode)

	err := migrator.ApplyMigrations(mode)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Printf("All migrations applied successfully\n")
	return nil
}

// cmdRevertMigration reverts the last migration
func cmdRevertMigration(migrator *migrations.HybridMigrator) error {
	fmt.Printf("Reverting last migration...\n")

	err := migrator.RevertMigration()
	if err != nil {
		return fmt.Errorf("failed to revert migration: %w", err)
	}

	fmt.Printf("Migration reverted successfully\n")
	return nil
}

// cmdMigrationStatus shows migration status
func cmdMigrationStatus(migrator *migrations.HybridMigrator) error {
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	fmt.Printf("Migration Status\n")
	fmt.Printf("================\n\n")

	// Applied migrations
	fmt.Printf("Applied Migrations (%d):\n", len(status.AppliedMigrations))
	if len(status.AppliedMigrations) == 0 {
		fmt.Printf("  None\n")
	} else {
		for _, migration := range status.AppliedMigrations {
			fmt.Printf("  ✓ %s (%s)\n", migration.Name, migration.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Printf("\n")

	// Pending migrations
	fmt.Printf("Pending Migrations (%d):\n", len(status.PendingMigrations))
	if len(status.PendingMigrations) == 0 {
		fmt.Printf("  None\n")
	} else {
		for _, migration := range status.PendingMigrations {
			icon := "○"
			if migration.HasDestructiveChanges() {
				icon = "⚠️"
			}
			fmt.Printf("  %s %s (%s)\n", icon, migration.Name, migration.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Printf("\n")

	// Current changes
	if status.HasPendingChanges {
		fmt.Printf("Pending Changes:\n")
		fmt.Printf("  %s\n", status.Summary)
		if status.HasDestructiveChanges {
			fmt.Printf("  ⚠️  Contains destructive changes\n")
		}
	} else {
		fmt.Printf("No pending changes detected\n")
	}

	return nil
}

// cmdGenerateMigration generates a migration script without applying it
func cmdGenerateMigration(migrator *migrations.HybridMigrator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("migration name is required")
	}

	name := args[0]
	mode := migrations.ModeGenerateOnly

	fmt.Printf("Generating migration script: %s\n", name)

	migrationFile, err := migrator.AddMigration(name, mode)
	if err != nil {
		return fmt.Errorf("failed to generate migration: %w", err)
	}

	fmt.Printf("Migration script generated: %s\n", migrationFile.Filename)
	fmt.Printf("Review the script before applying with 'apply' command\n")

	return nil
}

// cmdForceMigration creates a migration with force destructive mode
func cmdForceMigration(migrator *migrations.HybridMigrator, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("migration name is required")
	}

	name := args[0]
	mode := migrations.ModeForceDestructive

	fmt.Printf("Creating migration with force destructive mode: %s\n", name)
	fmt.Printf("⚠️  WARNING: This allows destructive changes without confirmation\n")

	migrationFile, err := migrator.AddMigration(name, mode)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	fmt.Printf("Migration created: %s\n", migrationFile.Filename)

	return nil
}

// Example models (these would typically be in separate files)
// These are just examples to show the expected structure

/*
// User model example
type User struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

// Post model example
type Post struct {
	ID       int64  `db:"id" migration:"primary_key,auto_increment"`
	UserID   int64  `db:"user_id" migration:"not_null,foreign_key:users.id"`
	Title    string `db:"title" migration:"not_null,max_length:255"`
	Content  string `db:"content" migration:"type:TEXT"`
	IsPublic bool   `db:"is_public" migration:"not_null,default:false"`
}

// To register these models, you would call:
// migrator.DbSet(&User{})
// migrator.DbSet(&Post{})
*/

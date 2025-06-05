// EF Core-like Migration CLI Tool for GRA Framework
// Provides commands similar to Entity Framework Core migration commands
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Constants for error messages and formatting
const (
	ErrorFailedToGetHistoryFmt = "❌ Failed to get migration history: %v"
	FormatMigrationLine        = "   %s\n"
	TimeFormat                 = "2006-01-02 15:04:05"
)

// CLIConfig is the configuration for the CLI migration tool.
type CLIConfig struct {
	ConnectionString string
	MigrationsDir    string
	Verbose          bool
	// Individual connection parameters for PostgreSQL
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func main() {
	config := CLIConfig{}

	// Define CLI flags
	flag.StringVar(&config.ConnectionString, "connection", "", "Database connection string")
	flag.StringVar(&config.MigrationsDir, "migrations-dir", "./migrations", "Directory to store migration files")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")

	// PostgreSQL specific flags
	flag.StringVar(&config.Host, "host", "", "Database host (PostgreSQL only)")
	flag.StringVar(&config.Port, "port", "5432", "Database port (PostgreSQL only)")
	flag.StringVar(&config.User, "user", "", "Database user (PostgreSQL only)")
	flag.StringVar(&config.Password, "password", "", "Database password (PostgreSQL only)")
	flag.StringVar(&config.Database, "database", "", "Database name (PostgreSQL only)")
	flag.StringVar(&config.SSLMode, "sslmode", "disable", "SSL mode (PostgreSQL only)")

	flag.Parse()

	// Get command
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
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
			// Try to build PostgreSQL connection string from individual parameters
			if config.Host != "" && config.User != "" && config.Database != "" {
				config.ConnectionString = buildPostgreSQLConnectionString(config)
				fmt.Printf("🔗 Built connection string from parameters for database: %s\n", config.Database)
			} else {
				log.Printf("❌ Database connection required. Use -connection flag, DATABASE_URL env var, or provide -host, -user, -database flags")
				return
			}
		}
	}

	// Detect database driver
	var driverName string
	switch {
	case strings.HasPrefix(config.ConnectionString, "postgres://"), strings.Contains(config.ConnectionString, "user="):
		driverName = "postgres"
	case strings.HasSuffix(config.ConnectionString, ".db"), strings.Contains(config.ConnectionString, "sqlite"):
		driverName = "sqlite3"
	default:
		driverName = "postgres" // Default to postgres for backward compatibility
	}

	db, err := sql.Open(driverName, config.ConnectionString)
	if err != nil {
		log.Printf("❌ Failed to connect to database: %v", err)
		return
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("Warning: failed to close db: %v", cerr)
		}
	}()

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
		log.Printf("❌ Failed to initialize migration schema: %v", err)
		return
	}

	// Load migrations from filesystem before executing commands
	if err := loadMigrationsFromFilesystem(manager, config.MigrationsDir); err != nil {
		log.Printf("❌ Failed to load migrations from filesystem: %v", err)
		return
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
		return
	}
}

// addMigration implements Add-Migration command
func addMigration(manager *migrations.EFMigrationManager, args []string, config CLIConfig) {
	if len(args) == 0 {
		log.Printf("❌ Migration name required. Usage: add-migration <name>")
		return
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
		log.Printf("❌ Failed to save migration file: %v", err)
		return
	}

	fmt.Printf("✅ Migration created: %s\n", migration.ID)
	fmt.Printf("📁 File: %s/%s.sql\n", config.MigrationsDir, migration.ID)
	fmt.Println("📝 Edit the migration file and run 'update-database' to apply")
}

// updateDatabase implements Update-Database command
func updateDatabase(manager *migrations.EFMigrationManager, args []string, _ CLIConfig) {
	fmt.Println("🚀 Updating database...")

	var targetMigration []string
	if len(args) > 0 {
		targetMigration = []string{args[0]}
		fmt.Printf("🎯 Target migration: %s\n", args[0])
	}

	if err := manager.UpdateDatabase(targetMigration...); err != nil {
		log.Printf("❌ Failed to update database: %v", err)
		return
	}

	fmt.Println("✅ Database updated successfully!")
}

// getMigrations implements Get-Migration command
func getMigrations(manager *migrations.EFMigrationManager, _ CLIConfig) {
	fmt.Println("📋 Migration History:")
	fmt.Println("====================")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Printf(ErrorFailedToGetHistoryFmt, err)
		return
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
func rollbackMigration(manager *migrations.EFMigrationManager, args []string, _ CLIConfig) {
	if len(args) == 0 {
		log.Printf("❌ Target migration required. Usage: rollback <migration-name-or-id>")
		return
	}

	target := args[0]
	fmt.Printf("⏪ Rolling back to migration: %s\n", target)

	if err := manager.RollbackMigration(target); err != nil {
		log.Printf("❌ Failed to rollback migration: %v", err)
		return
	}

	fmt.Println("✅ Rollback completed successfully!")
}

// showStatus shows current migration status
func showStatus(manager *migrations.EFMigrationManager, config CLIConfig) {
	fmt.Println("📊 Migration Status:")
	fmt.Println("===================")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Printf("❌ Failed to get migration status: %v", err)
		return
	}

	sanitizedConnectionString := sanitizeConnectionString(config.ConnectionString)
	fmt.Printf("Database: %s\n", extractDBName(sanitizedConnectionString))
	fmt.Printf("Applied:  %d migrations\n", len(history.Applied))
	fmt.Printf("Pending:  %d migrations\n", len(history.Pending))
	fmt.Printf("Failed:   %d migrations\n", len(history.Failed))

	if len(history.Applied) > 0 {
		latest := history.Applied[len(history.Applied)-1]
		fmt.Printf("Latest:   %s (%s)\n", latest.ID, latest.AppliedAt.Format(TimeFormat))
	}

	if len(history.Pending) > 0 {
		fmt.Printf("Next:     %s\n", history.Pending[0].ID)
	}
}

// generateScript generates SQL script for migrations
func generateScript(manager *migrations.EFMigrationManager, args []string, _ CLIConfig) {
	fmt.Println("📜 Generating migration script...")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Printf(ErrorFailedToGetHistoryFmt, err)
		return
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
	fmt.Printf("-- Generated at: %s\n", time.Now().Format(TimeFormat))
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
func removeMigration(manager *migrations.EFMigrationManager, _ []string, config CLIConfig) {
	fmt.Println("🗑️  Removing last migration...")

	history, err := manager.GetMigrationHistory()
	if err != nil {
		log.Printf(ErrorFailedToGetHistoryFmt, err)
		return
	}

	if len(history.Pending) == 0 {
		log.Printf("❌ No pending migrations to remove")
		return
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
		result += fmt.Sprintf(" (%s)", m.AppliedAt.Format(TimeFormat))
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
	// #nosec G301 -- Directory must be user-accessible for migration files
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// Create migration file
	filename := fmt.Sprintf("%s/%s.sql", dir, migration.ID)
	// #nosec G304 -- File creation is controlled by migration logic, not user input
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Warning: failed to close file: %v", cerr)
		}
	}()

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
		time.Now().Format(TimeFormat),
		migration.Version,
		migration.UpSQL,
		migration.DownSQL,
	)

	_, err = file.WriteString(content)
	return err
}

// buildPostgreSQLConnectionString builds a PostgreSQL connection string from individual parameters
func buildPostgreSQLConnectionString(config CLIConfig) string {
	host := config.Host
	if host == "" {
		host = "localhost"
	}

	port := config.Port
	if port == "" {
		port = "5432"
	}

	sslmode := config.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, host, port, config.Database, sslmode)
}

func sanitizeConnectionString(connectionString string) string {
	re := regexp.MustCompile(`(postgres://.*:)(.*)(@.*)`)
	return re.ReplaceAllString(connectionString, "${1}*****${3}")
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
	fmt.Println(`PostgreSQL Connection Options:`)
	fmt.Println(`  -host <string>         Database host (default: localhost)`)
	fmt.Println(`  -port <string>         Database port (default: 5432)`)
	fmt.Println(`  -user <string>         Database user`)
	fmt.Println(`  -password <string>     Database password`)
	fmt.Println(`  -database <string>     Database name`)
	fmt.Println(`  -sslmode <string>      SSL mode (default: disable)`)
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
	fmt.Println()
	fmt.Println(`Connection Examples:`)
	fmt.Println(`  # Using individual PostgreSQL parameters`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra status`)
	fmt.Println()
	fmt.Println(`  # Using connection string`)
	fmt.Println(`  ef-migrate -connection "postgres://user:pass@localhost:5432/gra?sslmode=disable" status`)
	fmt.Println()
	fmt.Println(`Migration Examples:`)
	fmt.Println(`  # Create a new migration`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra add-migration CreateUsersTable "Initial user table"`)
	fmt.Println()
	fmt.Println(`  # Apply all pending migrations`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra update-database`)
	fmt.Println()
	fmt.Println(`  # Apply migrations up to a specific one`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra update-database CreateUsersTable`)
	fmt.Println()
	fmt.Println(`  # Rollback to a specific migration`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra rollback InitialMigration`)
	fmt.Println()
	fmt.Println(`  # View migration status`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra status`)
	fmt.Println()
	fmt.Println(`  # List all migrations`)
	fmt.Println(`  ef-migrate -host localhost -user postgres -password MyPass123 -database gra get-migration`)
	fmt.Println()
	fmt.Println(`ENVIRONMENT:`)
	fmt.Println(`  DATABASE_URL    Default database connection string`)
	fmt.Println()
	fmt.Println(`📚 More info: https://github.com/your-org/gra/docs/migrations`)
}

// loadMigrationsFromFilesystem loads migration files from the filesystem
func loadMigrationsFromFilesystem(manager *migrations.EFMigrationManager, migrationsDir string) error {
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil // No migrations directory, no error
	}

	// Get all .sql files in the migrations directory
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Regular expression to parse migration filename: VERSION_NAME.sql
	migrationRegex := regexp.MustCompile(`^(\d+)_(.+)\.sql$`)

	for _, file := range files {
		filename := filepath.Base(file)
		matches := migrationRegex.FindStringSubmatch(filename)

		if len(matches) != 3 {
			continue // Skip files that don't match the pattern
		}

		versionStr := matches[1]
		name := matches[2]

		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			continue // Skip files with invalid version
		}

		// Read migration file content
		// #nosec G304 -- File path is determined by migration manager logic, not user input
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Parse migration content to extract UP and DOWN SQL
		upSQL, downSQL := parseMigrationContent(string(content))

		// Create migration ID
		migrationID := fmt.Sprintf("%d_%s", version, name)

		// Add migration to manager
		migration := migrations.Migration{
			ID:          migrationID,
			Name:        strings.ReplaceAll(name, "_", " "),
			Version:     version,
			Description: fmt.Sprintf("Migration loaded from %s", filename),
			UpSQL:       upSQL,
			DownSQL:     downSQL,
			State:       migrations.MigrationStatePending,
		}

		// Add to manager's pending migrations if not already applied
		manager.AddLoadedMigration(migration)
	}

	return nil
}

// parseMigrationContent parses migration file content to extract UP and DOWN SQL
func parseMigrationContent(content string) (upSQL, downSQL string) {
	lines := strings.Split(content, "\n")
	var upLines, downLines []string
	var inDownSection bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines for section detection
		if strings.HasPrefix(trimmed, "--") {
			if strings.Contains(strings.ToLower(trimmed), "down migration") ||
				strings.Contains(strings.ToLower(trimmed), "rollback") {
				inDownSection = true
				continue
			}
			if strings.Contains(strings.ToLower(trimmed), "up migration") {
				inDownSection = false
				continue
			}
		}

		// Add lines to appropriate section
		if inDownSection {
			downLines = append(downLines, line)
		} else {
			// Skip header comments for UP section
			if !strings.HasPrefix(trimmed, "--") || strings.Contains(trimmed, "Migration:") || strings.Contains(trimmed, "Description:") || strings.Contains(trimmed, "Created:") || strings.Contains(trimmed, "Version:") {
				if !strings.HasPrefix(trimmed, "--") {
					upLines = append(upLines, line)
				}
			} else {
				upLines = append(upLines, line)
			}
		}
	}

	upSQL = strings.TrimSpace(strings.Join(upLines, "\n"))
	downSQL = strings.TrimSpace(strings.Join(downLines, "\n"))

	// Remove comment prefixes from DOWN SQL
	if downSQL != "" {
		downLines = strings.Split(downSQL, "\n")
		var cleanDownLines []string
		for _, line := range downLines {
			if strings.HasPrefix(strings.TrimSpace(line), "-- ") {
				cleanDownLines = append(cleanDownLines, strings.TrimPrefix(strings.TrimSpace(line), "-- "))
			} else {
				cleanDownLines = append(cleanDownLines, line)
			}
		}
		downSQL = strings.TrimSpace(strings.Join(cleanDownLines, "\n"))
	}

	return upSQL, downSQL
}

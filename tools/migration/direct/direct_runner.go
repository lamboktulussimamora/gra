// Package main provides a CLI tool for running direct database migrations.
// It supports applying and tracking schema migrations for PostgreSQL databases.
//
// Usage:
//
//	direct_runner --conn 'postgres://user:pass@host/db' --up
//	direct_runner --conn 'postgres://user:pass@host/db' --status
//
// Flags:
//
//	--up      Apply pending migrations
//	--status  Show migration status
//	--down    Roll back the last applied migration (not implemented)
//
// Example:
//
//	direct_runner --conn 'postgres://localhost:5432/mydb?sslmode=disable' --up
//
// See README.md for more details.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

const (
	tableUsers            = "users"
	tableProducts         = "products"
	tableCategories       = "categories"
	tableSchemaMigrations = "schema_migrations"
)

const errNilDB = "db is nil"

var (
	upFlag     = flag.Bool("up", false, "Apply pending migrations")
	downFlag   = flag.Bool("down", false, "Roll back the last applied migration")
	connFlag   = flag.String("conn", "", "Database connection string")
	verbose    = flag.Bool("verbose", false, "Show verbose output")
	statusFlag = flag.Bool("status", false, "Show migration status")
)

const warnCloseDB = "Warning: failed to close db: %v"

func closeDBWithWarn(db *sql.DB) {
	if db == nil {
		return
	}
	if cerr := db.Close(); cerr != nil {
		log.Printf(warnCloseDB, cerr)
	}
}

func exitWithDBClose(db *sql.DB, msg string, args ...interface{}) {
	closeDBWithWarn(db)
	log.Fatalf(msg, args...)
}

func main() {
	flag.Parse()

	if *connFlag == "" {
		fmt.Println("Error: Database connection string is required")
		fmt.Println("Usage: direct_runner --conn 'postgres://user:pass@host/db' --up")
		fmt.Println("       direct_runner --conn 'postgres://user:pass@host/db' --status")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", *connFlag)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		exitWithDBClose(db, "Database connection failed: %v", err)
	}

	if *verbose {
		fmt.Println("✓ Connected to database successfully")
	}

	if err := ensureMigrationTable(db); err != nil {
		exitWithDBClose(db, "Failed to ensure migration table: %v", err)
	}

	if *statusFlag {
		if err := showStatus(db); err != nil {
			exitWithDBClose(db, "Status failed: %v", err)
		}
		closeDBWithWarn(db)
		return
	}

	if *upFlag {
		if err := migrateUp(db); err != nil {
			exitWithDBClose(db, "Migration up failed: %v", err)
		}
		closeDBWithWarn(db)
		return
	}

	if *downFlag {
		closeDBWithWarn(db)
		fmt.Println("Migration down not implemented yet")
		return
	}

	flag.Usage()
	closeDBWithWarn(db)
	os.Exit(1)
}

// ensureMigrationTable creates the schema_migrations table if it does not exist.
func ensureMigrationTable(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("%s", errNilDB)
	}
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + tableSchemaMigrations + ` (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a map of applied migration versions.
func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	if db == nil {
		return nil, fmt.Errorf("%s", errNilDB)
	}
	applied := make(map[int]bool)

	rows, err := db.Query("SELECT version FROM " + tableSchemaMigrations + " ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("Warning: failed to close rows: %v", cerr)
		}
	}()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// showStatus prints the current migration status to stdout.
func showStatus(db *sql.DB) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Printf("Applied migrations: %d\n", len(applied))

	if len(applied) > 0 {
		fmt.Println("Applied versions:")
		for version := range applied {
			fmt.Printf("  - Version %d\n", version)
		}
	} else {
		fmt.Println("No migrations applied yet")
	}

	return nil
}

// migrateUp applies all pending migrations in order.
func migrateUp(db *sql.DB) error {
	if *verbose {
		fmt.Println("Starting migration up...")
	}

	migrations := getMigrationsList()

	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			if *verbose {
				fmt.Printf("Migration %d already applied, skipping\n", migration.Version)
			}
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return err
		}
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

// getMigrationsList returns the list of migrations to apply.
func getMigrationsList() []struct {
	Version     int
	Description string
	SQL         string
} {
	return []struct {
		Version     int
		Description string
		SQL         string
	}{
		{
			Version:     1,
			Description: "Create initial schema with users and products tables",
			SQL: `
				CREATE TABLE IF NOT EXISTS ` + tableUsers + ` (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					email VARCHAR(255) UNIQUE NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				CREATE TABLE IF NOT EXISTS ` + tableProducts + ` (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					price DECIMAL(10,2) NOT NULL,
					description TEXT,
					user_id INTEGER REFERENCES users(id),
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
		{
			Version:     2,
			Description: "Add indexes for better performance",
			SQL: `
				CREATE INDEX IF NOT EXISTS idx_users_email ON ` + tableUsers + `(email);
				CREATE INDEX IF NOT EXISTS idx_products_user_id ON ` + tableProducts + `(user_id);
			`,
		},
		{
			Version:     3,
			Description: "Add categories table",
			SQL: `
				CREATE TABLE IF NOT EXISTS ` + tableCategories + ` (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NOT NULL UNIQUE,
					description TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
	}
}

// applyMigration applies a single migration in a transaction.
func applyMigration(db *sql.DB, migration struct {
	Version     int
	Description string
	SQL         string
}) error {
	if db == nil {
		return fmt.Errorf("%s", errNilDB)
	}
	if *verbose {
		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
	}

	if _, err := tx.Exec(migration.SQL); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rerr)
		}
		return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
	}

	_, err = tx.Exec("INSERT INTO "+tableSchemaMigrations+" (version) VALUES ($1)", migration.Version)
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rerr)
		}
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	fmt.Printf("✓ Applied migration %d: %s\n", migration.Version, migration.Description)
	return nil
}

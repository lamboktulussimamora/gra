package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	upFlag     = flag.Bool("up", false, "Apply pending migrations")
	downFlag   = flag.Bool("down", false, "Roll back the last applied migration")
	connFlag   = flag.String("conn", "", "Database connection string")
	verbose    = flag.Bool("verbose", false, "Show verbose output")
	statusFlag = flag.Bool("status", false, "Show migration status")
)

const migrationTableName = "schema_migrations"

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
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	if *verbose {
		fmt.Println("✓ Connected to database successfully")
	}

	if err := ensureMigrationTable(db); err != nil {
		log.Fatalf("Failed to ensure migration table: %v", err)
	}

	if *statusFlag {
		if err := showStatus(db); err != nil {
			log.Fatalf("Status failed: %v", err)
		}
	} else if *upFlag {
		if err := migrateUp(db); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	} else if *downFlag {
		fmt.Println("Migration down not implemented yet")
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %v", err)
	}
	return nil
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %v", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

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

func migrateUp(db *sql.DB) error {
	if *verbose {
		fmt.Println("Starting migration up...")
	}

	migrations := []struct {
		Version     int
		Description string
		SQL         string
	}{
		{
			Version:     1,
			Description: "Create initial schema with users and products tables",
			SQL: `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					email VARCHAR(255) UNIQUE NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				CREATE TABLE IF NOT EXISTS products (
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
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_products_user_id ON products(user_id);
			`,
		},
		{
			Version:     3,
			Description: "Add categories table",
			SQL: `
				CREATE TABLE IF NOT EXISTS categories (
					id SERIAL PRIMARY KEY,
					name VARCHAR(255) NOT NULL UNIQUE,
					description TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
	}

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

		if *verbose {
			fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %v", migration.Version, err)
		}

		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %v", migration.Version, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %v", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %v", migration.Version, err)
		}

		fmt.Printf("✓ Applied migration %d: %s\n", migration.Version, migration.Description)
	}

	fmt.Println("All migrations applied successfully")
	return nil
}

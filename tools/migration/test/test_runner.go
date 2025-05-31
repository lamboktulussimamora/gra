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
	up = flag.Bool("up", false, "Apply migrations")
	conn = flag.String("conn", "", "Connection string")
)

func main() {
	flag.Parse()
	if *conn == "" {
		fmt.Println("Usage: test_runner --conn 'postgres://...' --up")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", *conn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Connection failed:", err)
	}

	fmt.Println("✓ Database connection successful!")

	if *up {
		// Create migrations table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			log.Fatal("Failed to create migrations table:", err)
		}

		// Create users table
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		if err != nil {
			log.Fatal("Failed to create users table:", err)
		}

		fmt.Println("✓ Users table created successfully!")

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT DO NOTHING")
		if err != nil {
			log.Fatal("Failed to record migration:", err)
		}

		fmt.Println("✓ Migration completed!")
	}
}

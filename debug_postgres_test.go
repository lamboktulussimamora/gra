package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	_ "github.com/lib/pq"
)

type TestUserDebug struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at" sql:"-"`
	UpdatedAt time.Time `db:"updated_at" sql:"-"`
}

func (TestUserDebug) TableName() string {
	return "test_users_pg"
}

func main() {
	// Connect to test database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost", "5433", "gra_user", "gra_password", "gra_test")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create table
	createTable := `
		CREATE TABLE IF NOT EXISTS test_users_pg (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE,
			age INTEGER,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createTable); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Test the insert
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	user := TestUserDebug{
		Name:     "Debug User",
		Email:    "debug@example.com",
		Age:      30,
		IsActive: true,
	}

	fmt.Printf("Before Add: ID=%d, Name=%s\n", user.ID, user.Name)

	ctx.Add(&user)
	affected, err := ctx.SaveChanges()

	fmt.Printf("After SaveChanges: affected=%d, err=%v\n", affected, err)
	fmt.Printf("After SaveChanges: ID=%d, Name=%s\n", user.ID, user.Name)
}

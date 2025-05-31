package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/migrations"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("🚀 GRA Framework - Enhanced ORM Demo (Multi-Database)")
	fmt.Println("====================================================")

	// Check for PostgreSQL first, fallback to SQLite
	var db *sql.DB
	var err error
	var dbType string

	// Try PostgreSQL connection
	postgresConnectionString := "host=localhost port=5432 user=postgres dbname=gra_demo sslmode=disable password=postgres"
	if pgUrl := os.Getenv("DATABASE_URL"); pgUrl != "" {
		postgresConnectionString = pgUrl
	}

	db, err = sql.Open("postgres", postgresConnectionString)
	if err == nil && db.Ping() == nil {
		dbType = "PostgreSQL"
		fmt.Println("✅ Connected to PostgreSQL database")
	} else {
		fmt.Println("⚠️  PostgreSQL not available, falling back to SQLite...")
		// Fallback to SQLite
		db, err = sql.Open("sqlite3", "./enhanced_demo.db")
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %v", err)
		}
		dbType = "SQLite"
		fmt.Println("✅ Connected to SQLite database")
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Run migrations
	fmt.Printf("\n📦 Running Migrations (%s)...\n", dbType)
	migrator := migrations.NewAutoMigrator(ctx, db)
	if err := migrator.MigrateModels(&models.User{}, &models.Category{}, &models.Product{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed")

	// Test basic operations
	fmt.Println("\n🎯 Testing Basic Operations...")

	// Create a user
	user := &models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		IsActive:  true,
	}

	ctx.Add(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to save user: %v", err)
	}
	fmt.Printf("✅ Created user: %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)

	// Test queries using enhanced set
	fmt.Println("\n🔍 Testing LINQ-style Queries...")
	users := dbcontext.NewEnhancedDbSet[models.User](ctx)

	// Count users
	count, err := users.Count()
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	fmt.Printf("✅ Total users: %d\n", count)

	// Find user by email
	foundUser, err := users.Where("email = ?", "john@example.com").FirstOrDefault()
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}
	if foundUser != nil {
		fmt.Printf("✅ Found user: %s %s\n", foundUser.FirstName, foundUser.LastName)
	}

	// Test change tracking
	fmt.Println("\n📊 Testing Change Tracking...")
	foundUser.Email = "john.updated@example.com"
	ctx.Update(foundUser)

	state := ctx.ChangeTracker.GetEntityState(foundUser)
	fmt.Printf("✅ Entity state: %s\n", state.String())

	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to save changes: %v", err)
	}
	fmt.Println("✅ Changes saved")

	// Test timestamp management
	fmt.Println("\n⏰ Testing Timestamp Management...")

	// Check if timestamps were set during creation
	if !user.CreatedAt.IsZero() {
		fmt.Printf("✅ CreatedAt timestamp set: %v\n", user.CreatedAt)
	} else {
		fmt.Println("⚠️  CreatedAt timestamp not set")
	}

	if !user.UpdatedAt.IsZero() {
		fmt.Printf("✅ UpdatedAt timestamp set: %v\n", user.UpdatedAt)
	} else {
		fmt.Println("⚠️  UpdatedAt timestamp not set")
	}

	// Test update timestamp changes
	originalUpdatedAt := foundUser.UpdatedAt
	foundUser.FirstName = "John Updated"

	ctx.Update(foundUser)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to update user for timestamp test: %v", err)
	}

	if foundUser.UpdatedAt.After(originalUpdatedAt) {
		fmt.Println("✅ UpdatedAt timestamp automatically updated on modification")
	} else {
		fmt.Println("⚠️  UpdatedAt timestamp was not updated")
	}

	// Test BaseEntity field inclusion
	fmt.Println("\n🏗️  Testing BaseEntity Field Inclusion...")

	// Verify we can query using BaseEntity fields
	recentUsers := dbcontext.NewEnhancedDbSet[models.User](ctx)
	foundRecentUsers, err := recentUsers.Where("created_at > ?", "2024-01-01").ToList()
	if err != nil {
		log.Fatalf("Failed to query by created_at: %v", err)
	}
	fmt.Printf("✅ Successfully queried by BaseEntity field 'created_at': found %d users\n", len(foundRecentUsers))

	fmt.Println("\n🎉 All tests completed successfully!")
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/migrations"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("🐛 Debug Delete Operation")
	fmt.Println("=========================")

	// Setup database
	dbPath := "./debug_delete.db"
	os.Remove(dbPath) // Fresh start

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Run migrations
	fmt.Println("\n📦 Running Migrations...")
	migrator := migrations.NewAutoMigrator(ctx, db)
	if err := migrator.MigrateModels(&models.User{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed")

	// Create a single user
	fmt.Println("\n🎯 Creating User...")
	user := &models.User{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		IsActive:  true,
	}

	ctx.Add(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("✅ Created user with ID: %d\n", user.ID)

	// Check database contents before delete
	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	beforeUsers, err := userSet.ToList()
	if err != nil {
		log.Fatalf("Failed to query before delete: %v", err)
	}
	fmt.Printf("📊 Users before delete: %d\n", len(beforeUsers))

	// Debug the ID value
	fmt.Printf("🔍 User ID before delete: %d\n", user.ID)
	fmt.Printf("🔍 Entity state before delete: %s\n", ctx.ChangeTracker.GetEntityState(user).String())

	// Debug what's actually in the database
	var dbID int
	var dbEmail string
	err = db.QueryRow("SELECT id, email FROM users WHERE id = ?", user.ID).Scan(&dbID, &dbEmail)
	if err != nil {
		fmt.Printf("🔍 No record found with ID %d: %v\n", user.ID, err)
	} else {
		fmt.Printf("🔍 Found in DB: ID=%d, Email=%s\n", dbID, dbEmail)
	}

	// Delete the user
	fmt.Println("\n🗑️  Deleting User...")
	ctx.Delete(user)
	fmt.Printf("🔍 Entity state after Delete(): %s\n", ctx.ChangeTracker.GetEntityState(user).String())

	// Save changes
	rowsAffected, err := ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Printf("✅ SaveChanges completed, rows affected: %d\n", rowsAffected)

	// Check database contents after delete
	afterUsers, err := userSet.ToList()
	if err != nil {
		log.Fatalf("Failed to query after delete: %v", err)
	}
	fmt.Printf("📊 Users after delete: %d\n", len(afterUsers))

	// Raw SQL query to verify
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to run raw count query: %v", err)
	}
	fmt.Printf("📊 Raw SQL count: %d\n", count)

	if count == 0 {
		fmt.Println("🎉 Delete operation successful!")
	} else {
		fmt.Println("❌ Delete operation failed - records still exist")
	}
}

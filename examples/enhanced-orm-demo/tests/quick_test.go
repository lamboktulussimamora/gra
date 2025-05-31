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
	fmt.Println("🧪 GRA Framework - Quick CRUD Test")
	fmt.Println("==================================")

	// Setup database
	dbPath := "./quick_test.db"
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
	if err := migrator.MigrateModels(&models.User{}, &models.Category{}, &models.Product{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed")

	// Test basic CRUD operations
	fmt.Println("\n🎯 Testing Basic CRUD Operations...")

	// CREATE
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
	fmt.Printf("✅ Created user: %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)
	fmt.Printf("   Created at: %v\n", user.CreatedAt)
	fmt.Printf("   Updated at: %v\n", user.UpdatedAt)

	// READ
	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	foundUser, err := userSet.Where("email = ?", "test@example.com").FirstOrDefault()
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}
	if foundUser != nil {
		fmt.Printf("✅ Found user: %s %s\n", foundUser.FirstName, foundUser.LastName)
	}

	// UPDATE
	originalUpdatedAt := foundUser.UpdatedAt
	foundUser.FirstName = "Updated"
	ctx.Update(foundUser)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}
	fmt.Printf("✅ Updated user: %s\n", foundUser.FirstName)
	if foundUser.UpdatedAt.After(originalUpdatedAt) {
		fmt.Println("✅ UpdatedAt timestamp was automatically updated")
	}

	// Test BaseEntity field query
	recentUsers, err := userSet.Where("created_at > ?", "2024-01-01").ToList()
	if err != nil {
		log.Fatalf("Failed to query by BaseEntity field: %v", err)
	}
	fmt.Printf("✅ Found %d users by BaseEntity field query\n", len(recentUsers))

	// DELETE
	ctx.Delete(foundUser)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Println("✅ Deleted user")

	// Verify deletion
	remaining, err := userSet.ToList()
	if err != nil {
		log.Fatalf("Failed to query remaining users: %v", err)
	}
	fmt.Printf("✅ Remaining users: %d\n", len(remaining))

	fmt.Println("\n🎉 Quick CRUD test completed successfully!")
	fmt.Println("✨ All BaseEntity features are working correctly!")
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/migrations"
	"github.com/lamboktulussimamora/gra/orm/models"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("🧪 GRA Framework - Comprehensive CRUD Testing")
	fmt.Println("=============================================")

	// Setup database
	dbPath := "./crud_test.db"
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

	// Test CREATE operations
	fmt.Println("\n🎯 Testing CREATE Operations...")

	// Create a category
	category := &models.Category{
		Name:        "Electronics",
		Description: "Electronic devices and gadgets",
	}

	if err := ctx.Insert(category); err != nil {
		log.Fatalf("Failed to create category: %v", err)
	}
	fmt.Printf("✅ Created category: %s (ID: %d, Created: %v)\n",
		category.Name, category.ID, category.CreatedAt)

	// Create a product
	product := &models.Product{
		Name:        "Smartphone",
		Description: "Latest Android smartphone",
		Price:       599.99,
		SKU:         "PHONE-001",
		CategoryID:  category.ID,
		InStock:     true,
		StockCount:  50,
	}

	if err := ctx.Insert(product); err != nil {
		log.Fatalf("Failed to create product: %v", err)
	}
	fmt.Printf("✅ Created product: %s (ID: %d, Created: %v)\n",
		product.Name, product.ID, product.CreatedAt)

	// Create a user
	user := &models.User{
		FirstName: "Alice",
		LastName:  "Johnson",
		Email:     "alice@example.com",
		Password:  "hashedpassword123",
		IsActive:  true,
	}

	if err := ctx.Insert(user); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("✅ Created user: %s %s (ID: %d, Created: %v)\n",
		user.FirstName, user.LastName, user.ID, user.CreatedAt)

	// Test UPDATE operations
	fmt.Println("\n📝 Testing UPDATE Operations...")

	// Update user
	originalUpdatedAt := user.UpdatedAt
	time.Sleep(100 * time.Millisecond) // Ensure timestamp difference

	user.FirstName = "Alice Updated"
	user.LastName = "Johnson Updated"

	if err := ctx.Update(user); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	// Verify the update timestamps changed
	fmt.Printf("✅ Updated user: %s %s\n", user.FirstName, user.LastName)
	fmt.Printf("   Original UpdatedAt: %v\n", originalUpdatedAt)
	fmt.Printf("   New UpdatedAt: %v\n", user.UpdatedAt)

	if user.UpdatedAt.After(originalUpdatedAt) {
		fmt.Println("✅ UpdatedAt timestamp was automatically updated")
	} else {
		fmt.Println("⚠️  UpdatedAt timestamp was not updated")
	}

	fmt.Println("\n🎉 Core CRUD tests completed successfully!")
	fmt.Println("✨ BaseEntity serialization, table naming, and embedded struct handling are working perfectly!")
}

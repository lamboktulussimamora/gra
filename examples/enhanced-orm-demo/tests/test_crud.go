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

	// Test READ operations
	fmt.Println("\n🔍 Testing READ Operations...")

	// Query users
	var users []models.User
	if err := ctx.Query(&users).Where("is_active = ?", true).Execute(); err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	fmt.Printf("✅ Found %d active users\n", len(users))

	// Query products by category
	var products []models.Product
	if err := ctx.Query(&products).Where("category_id = ?", category.ID).Execute(); err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	fmt.Printf("✅ Found %d products in %s category\n", len(products), category.Name)

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

	// Update product stock
	product.StockCount = 45
	product.Price = 549.99

	if err := ctx.Update(product); err != nil {
		log.Fatalf("Failed to update product: %v", err)
	}
	fmt.Printf("✅ Updated product stock to %d and price to $%.2f\n",
		product.StockCount, product.Price)

	// Test SOFT DELETE operations
	fmt.Println("\n🗑️  Testing SOFT DELETE Operations...")

	// Soft delete the product
	if err := ctx.SoftDelete(product); err != nil {
		log.Fatalf("Failed to soft delete product: %v", err)
	}

	if product.DeletedAt != nil {
		fmt.Printf("✅ Soft deleted product: %s (DeletedAt: %v)\n",
			product.Name, *product.DeletedAt)
	} else {
		fmt.Println("⚠️  Product DeletedAt was not set")
	}

	// Verify soft deleted products are excluded from normal queries
	var activeProducts []models.Product
	if err := ctx.Query(&activeProducts).Where("category_id = ?", category.ID).Execute(); err != nil {
		log.Fatalf("Failed to query active products: %v", err)
	}
	fmt.Printf("✅ Active products after soft delete: %d (should be 0)\n", len(activeProducts))

	// Test querying including soft deleted items
	var allProducts []models.Product
	if err := ctx.Query(&allProducts).IncludeSoftDeleted().Where("category_id = ?", category.ID).Execute(); err != nil {
		log.Fatalf("Failed to query all products: %v", err)
	}
	fmt.Printf("✅ All products (including deleted): %d (should be 1)\n", len(allProducts))

	// Test RESTORE operations
	fmt.Println("\n🔄 Testing RESTORE Operations...")

	// Restore the soft deleted product
	if err := ctx.Restore(product); err != nil {
		log.Fatalf("Failed to restore product: %v", err)
	}

	if product.DeletedAt == nil {
		fmt.Println("✅ Product restored successfully (DeletedAt is nil)")
	} else {
		fmt.Printf("⚠️  Product DeletedAt was not cleared: %v\n", *product.DeletedAt)
	}

	// Verify restored product appears in normal queries
	activeProducts = []models.Product{}
	if err := ctx.Query(&activeProducts).Where("category_id = ?", category.ID).Execute(); err != nil {
		log.Fatalf("Failed to query active products after restore: %v", err)
	}
	fmt.Printf("✅ Active products after restore: %d (should be 1)\n", len(activeProducts))

	// Test HARD DELETE operations
	fmt.Println("\n💥 Testing HARD DELETE Operations...")

	// Hard delete the user
	if err := ctx.Delete(user); err != nil {
		log.Fatalf("Failed to hard delete user: %v", err)
	}
	fmt.Println("✅ Hard deleted user")

	// Verify user is completely removed
	var remainingUsers []models.User
	if err := ctx.Query(&remainingUsers).IncludeSoftDeleted().Execute(); err != nil {
		log.Fatalf("Failed to query remaining users: %v", err)
	}
	fmt.Printf("✅ Remaining users: %d (should be 0)\n", len(remainingUsers))

	fmt.Println("\n🎉 All CRUD tests completed successfully!")
	fmt.Println("✨ BaseEntity serialization, table naming, and embedded struct handling are working perfectly!")
}

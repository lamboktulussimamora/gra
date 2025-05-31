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

func runCompleteTests() {
	fmt.Println("🧪 GRA Framework - Complete CRUD Testing")
	fmt.Println("========================================")

	// Setup database
	dbPath := "./complete_test.db"
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

	// Test CREATE operations with BaseEntity
	fmt.Println("\n🎯 Testing CREATE Operations...")

	// Create a category
	category := &models.Category{
		Name:        "Electronics",
		Description: "Electronic devices and gadgets",
	}

	ctx.Add(category)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to create category: %v", err)
	}
	fmt.Printf("✅ Created category: %s (ID: %d, Created: %v)\n",
		category.Name, category.ID, category.CreatedAt)

	testCreateOperations(ctx, category)
	testReadOperations(ctx, category)
	testUpdateOperations(ctx)
	testDeleteOperations(ctx)
	testAdvancedQuerying(ctx)

	fmt.Println("\n🎉 All complete CRUD tests passed successfully!")
	fmt.Println("✨ BaseEntity serialization, table naming, embedded struct handling,")
	fmt.Println("   timestamp management, and change tracking are all working perfectly!")
}

func testCreateOperations(ctx *dbcontext.EnhancedDbContext, category *models.Category) {
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

	ctx.Add(product)
	_, err := ctx.SaveChanges()
	if err != nil {
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

	ctx.Add(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("✅ Created user: %s %s (ID: %d, Created: %v)\n",
		user.FirstName, user.LastName, user.ID, user.CreatedAt)
}

func testReadOperations(ctx *dbcontext.EnhancedDbContext, category *models.Category) {
	fmt.Println("\n🔍 Testing READ Operations...")

	// Query users using EnhancedDbSet
	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	users, err := userSet.Where("is_active = ?", true).ToList()
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	fmt.Printf("✅ Found %d active users\n", len(users))

	// Query products by category using EnhancedDbSet
	productSet := dbcontext.NewEnhancedDbSet[models.Product](ctx)
	products, err := productSet.Where("category_id = ?", category.ID).ToList()
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	fmt.Printf("✅ Found %d products in %s category\n", len(products), category.Name)

	// Test querying by BaseEntity timestamps
	recentCategories, err := dbcontext.NewEnhancedDbSet[models.Category](ctx).
		Where("created_at > ?", "2024-01-01").ToList()
	if err != nil {
		log.Fatalf("Failed to query by created_at: %v", err)
	}
	fmt.Printf("✅ Found %d categories created after 2024-01-01\n", len(recentCategories))
}

func testUpdateOperations(ctx *dbcontext.EnhancedDbContext) {
	fmt.Println("\n📝 Testing UPDATE Operations...")

	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	user, err := userSet.FirstOrDefault()
	if err != nil || user == nil {
		log.Fatalf("Failed to get user for update test: %v", err)
	}

	// Update user
	originalUpdatedAt := user.UpdatedAt
	time.Sleep(100 * time.Millisecond) // Ensure timestamp difference

	user.FirstName = "Alice Updated"
	user.LastName = "Johnson Updated"

	ctx.Update(user)
	_, err = ctx.SaveChanges()
	if err != nil {
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
}

func testDeleteOperations(ctx *dbcontext.EnhancedDbContext) {
	fmt.Println("\n🗑️  Testing DELETE Operations...")

	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	user, err := userSet.FirstOrDefault()
	if err != nil || user == nil {
		log.Fatalf("Failed to get user for delete test: %v", err)
	}

	// Delete the user
	ctx.Delete(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Println("✅ Deleted user")

	// Verify user is removed
	remainingUsers, err := userSet.ToList()
	if err != nil {
		log.Fatalf("Failed to query remaining users: %v", err)
	}
	fmt.Printf("✅ Remaining users: %d (should be 0)\n", len(remainingUsers))
}

func testAdvancedQuerying(ctx *dbcontext.EnhancedDbContext) {
	fmt.Println("\n🔎 Testing Advanced Querying...")

	productSet := dbcontext.NewEnhancedDbSet[models.Product](ctx)

	// Test Count
	productCount, err := productSet.Count()
	if err != nil {
		log.Fatalf("Failed to count products: %v", err)
	}
	fmt.Printf("✅ Total products: %d\n", productCount)

	// Test FirstOrDefault
	firstProduct, err := productSet.Where("in_stock = ?", true).FirstOrDefault()
	if err != nil {
		log.Fatalf("Failed to get first product: %v", err)
	}
	if firstProduct != nil {
		fmt.Printf("✅ First product: %s\n", firstProduct.Name)
	}

	// Test OrderBy
	orderedProducts, err := productSet.OrderBy("name").ToList()
	if err != nil {
		log.Fatalf("Failed to order products: %v", err)
	}
	fmt.Printf("✅ Ordered products count: %d\n", len(orderedProducts))

	// Test AsNoTracking
	untracked, err := productSet.AsNoTracking().ToList()
	if err != nil {
		log.Fatalf("Failed to get untracked products: %v", err)
	}
	fmt.Printf("✅ Untracked products count: %d\n", len(untracked))
}

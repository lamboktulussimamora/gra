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

const (
	johnEmail    = "john.doe@example.com"
	updatedEmail = "john.doe.updated@example.com"
)

func main() {
	fmt.Println("🚀 GRA Framework - Enhanced Entity Framework Core-style ORM")
	fmt.Println("=========================================================")

	// Setup database
	dbPath := getEnvDefault("DB_PATH", "./enhanced_demo.db")

	// Remove existing database for fresh demo
	os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create enhanced context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Run auto migrations
	fmt.Println("\n📦 Step 1: Auto-Migrating Database Schema")
	migrator := migrations.NewAutoMigrator(ctx, db)
	if err := migrator.MigrateModels(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Demonstrate EF Core-style operations
	fmt.Println("\n🎯 Step 2: Entity Framework Core-style Operations")

	if err := demonstrateEntitySetOperations(ctx); err != nil {
		log.Fatalf("Entity set operations failed: %v", err)
	}

	if err := demonstrateAdvancedQueries(ctx); err != nil {
		log.Fatalf("Advanced queries failed: %v", err)
	}

	if err := demonstrateChangeTracking(ctx); err != nil {
		log.Fatalf("Change tracking failed: %v", err)
	}

	if err := demonstrateTransactions(ctx); err != nil {
		log.Fatalf("Transactions failed: %v", err)
	}

	fmt.Println("\n🎉 All demonstrations completed successfully!")
}

func demonstrateEntitySetOperations(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   📝 Entity Set Operations (LINQ-style)")

	// Create entity sets
	users := dbcontext.NewEnhancedDbSet[models.User](ctx)
	products := dbcontext.NewEnhancedDbSet[models.Product](ctx)

	// Create some categories
	electronics := &models.Category{
		Name:        "Electronics",
		Description: "Electronic devices and gadgets",
	}

	books := &models.Category{
		Name:        "Books",
		Description: "Books and literature",
	}

	// Add categories
	ctx.Add(electronics)
	ctx.Add(books)
	_, err := ctx.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("      ✅ Created categories: %s (ID: %d), %s (ID: %d)\n",
		electronics.Name, electronics.ID, books.Name, books.ID)

	// Create users
	john := &models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     johnEmail,
		Password:  "password123",
		IsActive:  true,
	}

	jane := &models.User{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Password:  "password456",
		IsActive:  true,
	}

	// Add users using entity sets
	ctx.Add(john)
	ctx.Add(jane)
	_, err = ctx.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("      ✅ Created users: %s %s (ID: %d), %s %s (ID: %d)\n",
		john.FirstName, john.LastName, john.ID,
		jane.FirstName, jane.LastName, jane.ID)

	// Create products
	laptop := &models.Product{
		Name:        "Gaming Laptop",
		Description: "High-performance gaming laptop",
		Price:       1299.99,
		SKU:         "LAPTOP-001",
		CategoryID:  electronics.ID,
		InStock:     true,
		StockCount:  10,
	}

	book := &models.Product{
		Name:        "Go Programming Book",
		Description: "Learn Go programming language",
		Price:       49.99,
		SKU:         "BOOK-001",
		CategoryID:  books.ID,
		InStock:     true,
		StockCount:  50,
	}

	ctx.Add(laptop)
	ctx.Add(book)
	_, err = ctx.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("      ✅ Created products: %s ($%.2f), %s ($%.2f)\n",
		laptop.Name, laptop.Price, book.Name, book.Price)

	// Demonstrate LINQ-style queries
	fmt.Println("\n   🔍 LINQ-style Queries")

	// Find user by email
	foundUser, err := users.Where("email = ?", johnEmail).FirstOrDefault()
	if err != nil {
		return err
	}
	if foundUser != nil {
		fmt.Printf("      ✅ Found user by email: %s %s\n", foundUser.FirstName, foundUser.LastName)
	}

	// Get all active users
	activeUsers, err := users.Where("is_active = ?", true).ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Found %d active users\n", len(activeUsers))

	// Get products by category
	electronicsProducts, err := products.Where("category_id = ?", electronics.ID).ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Found %d electronics products\n", len(electronicsProducts))

	// Get products within price range
	affordableProducts, err := products.Where("price <= ? AND in_stock = ?", 100.0, true).
		OrderBy("price").
		ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Found %d affordable products (≤ $100)\n", len(affordableProducts))

	return nil
}

func demonstrateAdvancedQueries(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   🔍 Advanced Query Operations")

	products := dbcontext.NewEnhancedDbSet[models.Product](ctx)

	// Count products
	totalProducts, err := products.Count()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Total products: %d\n", totalProducts)

	// Get first product
	firstProduct, err := products.OrderBy("created_at").FirstOrDefault()
	if err != nil {
		return err
	}
	if firstProduct != nil {
		fmt.Printf("      ✅ First product: %s\n", firstProduct.Name)
	}

	// Get products with pagination
	pageSize := 1
	page1Products, err := products.OrderBy("name").Take(pageSize).ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Page 1 products (limit %d): %d items\n", pageSize, len(page1Products))

	page2Products, err := products.OrderBy("name").Skip(pageSize).Take(pageSize).ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Page 2 products (skip %d, limit %d): %d items\n", pageSize, pageSize, len(page2Products))

	// Search products by name pattern
	searchResults, err := products.Where("name LIKE ?", "%Go%").ToList()
	if err != nil {
		return err
	}
	fmt.Printf("      ✅ Products matching 'Go': %d items\n", len(searchResults))

	return nil
}

func demonstrateChangeTracking(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   📊 Change Tracking Operations")

	users := dbcontext.NewEnhancedDbSet[models.User](ctx)

	// Get a user to modify
	user, err := users.Where("email = ?", johnEmail).FirstOrDefault()
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	fmt.Printf("      📝 Original user: %s %s (%s)\n", user.FirstName, user.LastName, user.Email)

	// Modify the user
	user.Email = updatedEmail
	user.LastName = "Updated"

	// Update in context (tracks changes)
	ctx.Update(user)

	fmt.Printf("      📝 Modified user: %s %s (%s)\n", user.FirstName, user.LastName, user.Email)

	// Check change tracker state
	state := ctx.ChangeTracker.GetEntityState(user)
	fmt.Printf("      📊 Entity state: %s\n", state.String())

	// Save changes
	_, err = ctx.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("      ✅ Changes saved to database\n")

	// Verify changes were saved
	updatedUser, err := users.Where("id = ?", user.ID).FirstOrDefault()
	if err != nil {
		return err
	}
	if updatedUser != nil {
		fmt.Printf("      ✅ Verified updated user: %s %s (%s)\n",
			updatedUser.FirstName, updatedUser.LastName, updatedUser.Email)
	}

	return nil
}

func demonstrateTransactions(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   💳 Transaction Operations")

	// Start a transaction
	tx, err := ctx.Database.Begin()
	if err != nil {
		return err
	}

	// Create a new context with the transaction
	txCtx := dbcontext.NewEnhancedDbContextWithTx(tx)

	// Create multiple entities in transaction
	order := &models.Order{
		UserID:      1,
		OrderNumber: "ORD-001",
		TotalAmount: 1349.98,
		Status:      "pending",
	}

	txCtx.Add(order)
	_, err = txCtx.SaveChanges()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Create order items
	item1 := &models.OrderItem{
		OrderID:   order.ID,
		ProductID: 1,
		Quantity:  1,
		UnitPrice: 1299.99,
		Total:     1299.99,
	}

	item2 := &models.OrderItem{
		OrderID:   order.ID,
		ProductID: 2,
		Quantity:  1,
		UnitPrice: 49.99,
		Total:     49.99,
	}

	txCtx.Add(item1)
	txCtx.Add(item2)
	_, err = txCtx.SaveChanges()
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("      ✅ Created order with ID %d and 2 items in transaction\n", order.ID)

	// Verify the order exists
	orders := dbcontext.NewEnhancedDbSet[models.Order](ctx)
	savedOrder, err := orders.Where("id = ?", order.ID).FirstOrDefault()
	if err != nil {
		return err
	}
	if savedOrder != nil {
		fmt.Printf("      ✅ Verified order: ID %d, Total: $%.2f, Status: %s\n",
			savedOrder.ID, savedOrder.TotalAmount, savedOrder.Status)
	}

	// Verify order items
	orderItems := dbcontext.NewEnhancedDbSet[models.OrderItem](ctx)
	items, err := orderItems.Where("order_id = ?", order.ID).ToList()
	if err != nil {
		return err
	}

	fmt.Printf("      ✅ Verified %d order items\n", len(items))

	return nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

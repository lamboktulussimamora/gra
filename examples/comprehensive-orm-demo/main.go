// Package main demonstrates a comprehensive ORM usage example for the GRA framework.
// This example covers migrations, enhanced ORM features, and best practices.
// Run this file to see a full demonstration of the framework's capabilities.
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

const isActiveWhere = "is_active = ?"

func main() {
	// Database connection string (SQLite for demo)
	connectionString := getConnectionString()

	fmt.Println("🚀 GRA Framework - Enhanced ORM Demonstration")
	fmt.Println("============================================")

	// Step 1: Run Migrations
	fmt.Println("\n📦 Step 1: Running Database Migrations")
	if err := runMigrations(connectionString); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed successfully")

	// Step 2: Demonstrate Enhanced ORM Features
	fmt.Println("\n🎯 Step 2: Demonstrating Enhanced ORM Features")
	if err := demonstrateORM(connectionString); err != nil {
		log.Fatalf("ORM demonstration failed: %v", err)
	}
	fmt.Println("✅ ORM demonstration completed successfully")

	fmt.Println("\n🎉 All demonstrations completed successfully!")
}

func getConnectionString() string {
	// Use SQLite for demo (easier setup)
	dbPath := getEnvDefault("DB_PATH", "./demo.db")
	return dbPath
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrations(connectionString string) error {
	// Open database connection
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database connection: %v", closeErr)
		}
	}()

	// Create enhanced database context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Create migration runner
	migrationRunner := migrations.NewAutoMigrator(ctx, db)

	// Define entities to migrate
	entities := []interface{}{
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
		&models.Role{},
		&models.UserRole{},
	}

	// Run automatic migrations
	return migrationRunner.MigrateModels(entities...)
}

func demonstrateORM(connectionString string) error {
	// Open database connection
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database connection: %v", closeErr)
		}
	}()

	// Create enhanced database context
	ctx := dbcontext.NewEnhancedDbContextWithDB(db)

	// Demonstrate basic CRUD operations
	if err := demonstrateBasicCRUD(ctx); err != nil {
		return fmt.Errorf("basic CRUD demonstration failed: %w", err)
	}

	// Demonstrate advanced querying
	if err := demonstrateAdvancedQuerying(ctx); err != nil {
		return fmt.Errorf("advanced querying demonstration failed: %w", err)
	}

	// Demonstrate transactions
	if err := demonstrateTransactions(ctx); err != nil {
		return fmt.Errorf("transaction demonstration failed: %w", err)
	}

	// Demonstrate change tracking
	if err := demonstrateChangeTracking(ctx); err != nil {
		return fmt.Errorf("change tracking demonstration failed: %w", err)
	}

	return nil
}

func demonstrateBasicCRUD(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   📝 Basic CRUD Operations")

	// Create new user
	user := &models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  true,
	}

	// Add user to context (tracks as "Added")
	ctx.Add(user)

	// Save changes to database
	_, err := ctx.SaveChanges()
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("      ✅ Created user: %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)

	// Read user back
	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	foundUser, err := userSet.Where("id = ?", user.ID).FirstOrDefault()
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if foundUser != nil {
		fmt.Printf("      ✅ Found user: %s %s (Email: %s)\n", foundUser.FirstName, foundUser.LastName, foundUser.Email)

		// Update user
		foundUser.Email = "john.doe.updated@example.com"
		ctx.Update(foundUser)

		_, err = ctx.SaveChanges()
		if err != nil {
			return fmt.Errorf("failed to save updated user: %w", err)
		}

		fmt.Printf("      ✅ Updated user email to: %s\n", foundUser.Email)

		// Delete user
		ctx.Delete(foundUser)

		_, err = ctx.SaveChanges()
		if err != nil {
			return fmt.Errorf("failed to save deleted user: %w", err)
		}

		fmt.Println("      ✅ Deleted user successfully")
	}

	return nil
}

func demonstrateAdvancedQuerying(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   🔍 Advanced Querying")

	// Create sample users for querying
	users := []*models.User{
		{FirstName: "Alice", LastName: "Johnson", Email: "alice@example.com", IsActive: true},
		{FirstName: "Bob", LastName: "Smith", Email: "bob@example.com", IsActive: false},
		{FirstName: "Charlie", LastName: "Brown", Email: "charlie@example.com", IsActive: true},
		{FirstName: "Diana", LastName: "Wilson", Email: "diana@example.com", IsActive: true},
	}

	// Add all users
	for _, user := range users {
		ctx.Add(user)
	}

	_, err := ctx.SaveChanges()
	if err != nil {
		return fmt.Errorf("failed to save users: %w", err)
	}

	fmt.Printf("      ✅ Created %d sample users\n", len(users))

	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)

	// Query active users
	activeUsers, err := userSet.Where(isActiveWhere, true).ToList()
	if err != nil {
		return fmt.Errorf("failed to query active users: %w", err)
	}

	fmt.Printf("      ✅ Found %d active users\n", len(activeUsers))

	// Query with ordering and limiting
	orderedUsers, err := userSet.
		Where(isActiveWhere, true).
		OrderBy("first_name").
		Take(2).
		ToList()
	if err != nil {
		return fmt.Errorf("failed to query ordered users: %w", err)
	}

	fmt.Printf("      ✅ Found %d ordered users (limited to 2)\n", len(orderedUsers))

	// Count operations
	totalCount, err := userSet.Count()
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	activeCount, err := userSet.Where(isActiveWhere, true).Count()
	if err != nil {
		return fmt.Errorf("failed to count active users: %w", err)
	}

	fmt.Printf("      ✅ Total users: %d, Active users: %d\n", totalCount, activeCount)

	// Check existence
	hasUsers, err := userSet.Any()
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	fmt.Printf("      ✅ Has users: %t\n", hasUsers)

	return nil
}

func demonstrateTransactions(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   💳 Transaction Management")

	// Begin transaction
	tx, err := ctx.Database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create transaction context
	txCtx := dbcontext.NewEnhancedDbContextWithTx(tx)

	// Create users within transaction
	user1 := &models.User{FirstName: "Trans", LastName: "User1", Email: "trans1@example.com", IsActive: true}
	user2 := &models.User{FirstName: "Trans", LastName: "User2", Email: "trans2@example.com", IsActive: true}

	txCtx.Add(user1)
	txCtx.Add(user2)

	// Save changes within transaction
	_, err = txCtx.SaveChanges()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to save changes in transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("      ✅ Transaction completed successfully")
	fmt.Printf("      ✅ Created users: %s and %s\n", user1.FirstName, user2.FirstName)

	return nil
}

func demonstrateChangeTracking(ctx *dbcontext.EnhancedDbContext) error {
	fmt.Println("\n   📊 Change Tracking")

	// Create a user
	user := &models.User{
		FirstName: "Track",
		LastName:  "Test",
		Email:     "track@example.com",
		IsActive:  true,
	}

	ctx.Add(user)

	// Check entity state
	state := ctx.ChangeTracker.GetEntityState(user)
	fmt.Printf("      ✅ Entity state after Add: %v\n", state)

	_, err := ctx.SaveChanges()
	if err != nil {
		return fmt.Errorf("failed to save tracked user: %w", err)
	}

	// Check state after save
	state = ctx.ChangeTracker.GetEntityState(user)
	fmt.Printf("      ✅ Entity state after SaveChanges: %v\n", state)

	// Modify entity
	user.Email = "track.modified@example.com"
	ctx.Update(user)

	// Check state after modification
	state = ctx.ChangeTracker.GetEntityState(user)
	fmt.Printf("      ✅ Entity state after Update: %v\n", state)

	// Demo read-only queries (no tracking)
	userSet := dbcontext.NewEnhancedDbSet[models.User](ctx)
	readOnlyUsers, err := userSet.AsNoTracking().Where(isActiveWhere, true).ToList()
	if err != nil {
		return fmt.Errorf("failed to execute no-tracking query: %w", err)
	}

	fmt.Printf("      ✅ Read-only query returned %d users (not tracked)\n", len(readOnlyUsers))

	return nil
}

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lamboktulussimamora/gra/orm/dbcontext"
	"github.com/lamboktulussimamora/gra/orm/models"
)

func main() {
	// Database connection string
	dbURI := "host=localhost port=5432 user=postgres password=MyPassword_123 dbname=gra_test sslmode=disable"

	// Initialize DbContext
	dbCtx, err := dbcontext.NewDbContext(dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbCtx.Close()

	fmt.Println("🚀 Testing GRA ORM System (GORM-free)")
	fmt.Println("=====================================")

	// Test 1: Create tables manually (since auto-migration has issues)
	fmt.Println("\n1. Creating database tables...")
	if err := createTables(dbCtx); err != nil {
		log.Printf("Warning: Failed to create tables: %v", err)
	}

	// Test 2: Basic CRUD operations
	fmt.Println("\n2. Testing basic CRUD operations...")
	testBasicCRUD(dbCtx)

	// Test 3: Transaction support
	fmt.Println("\n3. Testing transaction support...")
	testTransactions(dbCtx)

	// Test 4: Relationships
	fmt.Println("\n4. Testing relationships...")
	testRelationships(dbCtx)

	fmt.Println("\n✅ ORM system testing completed!")
}

func createTables(dbCtx *dbcontext.DbContext) error {
	// Create tables in dependency order
	tables := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(50) NOT NULL UNIQUE,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			parent_id BIGINT REFERENCES categories(id)
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			first_name VARCHAR(50) NOT NULL,
			last_name VARCHAR(50) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			last_login TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			sku VARCHAR(100) NOT NULL UNIQUE,
			category_id BIGINT REFERENCES categories(id),
			in_stock BOOLEAN DEFAULT true,
			stock_count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id BIGINT NOT NULL REFERENCES users(id),
			order_number VARCHAR(50) NOT NULL UNIQUE,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			total_amount DECIMAL(10,2) NOT NULL,
			shipped_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			order_id BIGINT NOT NULL REFERENCES orders(id),
			product_id BIGINT NOT NULL REFERENCES products(id),
			quantity INTEGER NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			total DECIMAL(10,2) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id BIGINT NOT NULL REFERENCES users(id),
			product_id BIGINT NOT NULL REFERENCES products(id),
			rating INTEGER NOT NULL,
			title VARCHAR(200) NOT NULL,
			comment TEXT NOT NULL,
			is_verified BOOLEAN DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			user_id BIGINT NOT NULL REFERENCES users(id),
			role_id BIGINT NOT NULL REFERENCES roles(id),
			UNIQUE(user_id, role_id)
		)`,
	}

	for _, sql := range tables {
		if err := dbCtx.ExecuteSQL(sql); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func testBasicCRUD(dbCtx *dbcontext.DbContext) {
	// Get DbSets
	userDbSet := dbcontext.GetDbSet[*models.User](dbCtx)
	categoryDbSet := dbcontext.GetDbSet[*models.Category](dbCtx)

	// Create a category
	category := &models.Category{
		Name:        "Electronics",
		Description: "Electronic products and gadgets",
	}

	if err := categoryDbSet.Add(category); err != nil {
		log.Printf("Error creating category: %v", err)
		return
	}
	fmt.Printf("✓ Created category: %s (ID: %d)\n", category.Name, category.ID)

	// Create a user
	user := &models.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "hashedpassword123",
		IsActive:  true,
	}

	if err := userDbSet.Add(user); err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}
	fmt.Printf("✓ Created user: %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)

	// Read users
	users, err := userDbSet.ToList()
	if err != nil {
		log.Printf("Error reading users: %v", err)
		return
	}
	fmt.Printf("✓ Found %d users in database\n", len(users))

	// Update user
	if len(users) > 0 {
		firstUser := users[0]
		firstUser.LastLogin = &[]time.Time{time.Now()}[0]
		if err := userDbSet.Update(firstUser); err != nil {
			log.Printf("Error updating user: %v", err)
		} else {
			fmt.Printf("✓ Updated user login time\n")
		}
	}

	// Count
	count, err := userDbSet.Count()
	if err != nil {
		log.Printf("Error counting users: %v", err)
	} else {
		fmt.Printf("✓ Total users: %d\n", count)
	}
}

func testTransactions(dbCtx *dbcontext.DbContext) {
	// Begin transaction
	if err := dbCtx.BeginTransaction(); err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}

	userDbSet := dbcontext.GetDbSet[*models.User](dbCtx)

	// Create user in transaction
	user := &models.User{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Password:  "hashedpassword456",
		IsActive:  true,
	}

	if err := userDbSet.Add(user); err != nil {
		log.Printf("Error creating user in transaction: %v", err)
		dbCtx.RollbackTransaction()
		return
	}

	// Commit transaction
	if err := dbCtx.CommitTransaction(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return
	}

	fmt.Printf("✓ Transaction completed successfully - Created user: %s\n", user.Email)
}

func testRelationships(dbCtx *dbcontext.DbContext) {
	productDbSet := dbcontext.GetDbSet[*models.Product](dbCtx)
	categoryDbSet := dbcontext.GetDbSet[*models.Category](dbCtx)

	// Find a category
	categories, err := categoryDbSet.ToList()
	if err != nil || len(categories) == 0 {
		fmt.Printf("No categories found for relationship test\n")
		return
	}

	// Create a product linked to the category
	product := &models.Product{
		Name:        "Smartphone",
		Description: "Latest smartphone with advanced features",
		Price:       599.99,
		SKU:         "PHONE-001",
		CategoryID:  categories[0].ID,
		InStock:     true,
		StockCount:  50,
	}

	if err := productDbSet.Add(product); err != nil {
		log.Printf("Error creating product: %v", err)
		return
	}

	fmt.Printf("✓ Created product: %s linked to category: %s\n", product.Name, categories[0].Name)

	// Test finding by specific criteria
	expensiveProducts := productDbSet.Where("price > $1", 500.0)
	products, err := expensiveProducts.ToList()
	if err != nil {
		log.Printf("Error querying expensive products: %v", err)
	} else {
		fmt.Printf("✓ Found %d expensive products (>$500)\n", len(products))
	}
}

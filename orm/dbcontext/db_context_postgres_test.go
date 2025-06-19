package dbcontext

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Constants for PostgreSQL testing
const (
	postgresHost     = "localhost"
	postgresPort     = "5433"
	postgresUser     = "gra_user"
	postgresPassword = "gra_password"
	postgresDBName   = "gra_test"

	pgTestTable = `
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

	pgTestTableUsers = `
		CREATE TABLE IF NOT EXISTS users_pg (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE,
			full_name VARCHAR(255),
			age INTEGER CHECK (age >= 0),
			salary DECIMAL(10,2),
			is_verified BOOLEAN DEFAULT false,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
)

// TestUserPG represents a test user for PostgreSQL
type TestUserPG struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at" sql:"-"`
	UpdatedAt time.Time `db:"updated_at" sql:"-"`
}

// TableName returns the table name for TestUserPG
func (TestUserPG) TableName() string {
	return "test_users_pg"
}

// UserPG represents a more complex user entity for PostgreSQL
type UserPG struct {
	ID         int64     `db:"id"`
	Username   string    `db:"username"`
	Email      string    `db:"email"`
	FullName   string    `db:"full_name"`
	Age        int       `db:"age"`
	Salary     float64   `db:"salary"`
	IsVerified bool      `db:"is_verified"`
	Metadata   string    `db:"metadata"`
	CreatedAt  time.Time `db:"created_at" sql:"-"`
	UpdatedAt  time.Time `db:"updated_at" sql:"-"`
}

// TableName returns the table name for UserPG
func (UserPG) TableName() string {
	return "users_pg"
}

// setupPostgresDB sets up a PostgreSQL database connection for testing
func setupPostgresDB(t *testing.T) *sql.DB {
	// Skip PostgreSQL tests if not available
	if os.Getenv("SKIP_POSTGRES_TESTS") == "true" {
		t.Skip("Skipping PostgreSQL tests (SKIP_POSTGRES_TESTS=true)")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Failed to connect to PostgreSQL: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("PostgreSQL not available: %v", err)
	}

	// Create test tables
	if _, err := db.Exec(pgTestTable); err != nil {
		db.Close()
		t.Fatalf("Failed to create test table: %v", err)
	}

	if _, err := db.Exec(pgTestTableUsers); err != nil {
		db.Close()
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

// cleanupPostgresDB cleans up test data
func cleanupPostgresDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM test_users_pg")
	if err != nil {
		t.Logf("Failed to clean up test_users_pg: %v", err)
	}

	_, err = db.Exec("DELETE FROM users_pg")
	if err != nil {
		t.Logf("Failed to clean up users_pg: %v", err)
	}
}

// TestPostgreSQLIntegration tests the dbcontext with actual PostgreSQL database
func TestPostgreSQLIntegration(t *testing.T) {
	db := setupPostgresDB(t)
	defer func() {
		cleanupPostgresDB(t, db)
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	// Clean up before starting tests
	cleanupPostgresDB(t, db)

	ctx := NewEnhancedDbContextWithDB(db)

	t.Run("PostgreSQL driver detection", func(t *testing.T) {
		if ctx.driver != "postgres" {
			t.Errorf("Expected driver 'postgres', got '%s'", ctx.driver)
		}
	})

	t.Run("INSERT with PostgreSQL", func(t *testing.T) {
		cleanupPostgresDB(t, db)              // Clean before each subtest
		ctx := NewEnhancedDbContextWithDB(db) // Fresh context for each test
		user := TestUserPG{
			Name:     "John PostgreSQL",
			Email:    "john.pg@example.com",
			Age:      28,
			IsActive: true,
		}

		ctx.Add(&user)
		affected, err := ctx.SaveChanges()

		if err != nil {
			t.Fatalf("Failed to save changes: %v", err)
		}

		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		if user.ID == 0 {
			t.Error("Expected ID to be set after insert")
		}

		// Verify the user was actually inserted
		dbSet := NewEnhancedDbSet[TestUserPG](ctx)
		users, err := dbSet.Where("email = ?", "john.pg@example.com").ToList()
		if err != nil {
			t.Fatalf("Failed to fetch user: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "John PostgreSQL" {
			t.Errorf("Expected name 'John PostgreSQL', got '%s'", users[0].Name)
		}
	})

	t.Run("UPDATE with PostgreSQL", func(t *testing.T) {
		cleanupPostgresDB(t, db)              // Clean before each subtest
		ctx := NewEnhancedDbContextWithDB(db) // Fresh context for each test
		// First insert a user
		user := TestUserPG{
			Name:     "Alice PostgreSQL",
			Email:    "alice.pg@example.com",
			Age:      25,
			IsActive: true,
		}

		ctx.Add(&user)
		_, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		// Update the user
		user.Name = "Alice Updated PG"
		user.Age = 26
		ctx.Update(&user)

		affected, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		// Verify the update
		dbSet := NewEnhancedDbSet[TestUserPG](ctx)
		users, err := dbSet.Where("id = ?", user.ID).ToList()
		if err != nil {
			t.Fatalf("Failed to fetch updated user: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		if users[0].Name != "Alice Updated PG" {
			t.Errorf("Expected name 'Alice Updated PG', got '%s'", users[0].Name)
		}

		if users[0].Age != 26 {
			t.Errorf("Expected age 26, got %d", users[0].Age)
		}
	})

	t.Run("DELETE with PostgreSQL", func(t *testing.T) {
		cleanupPostgresDB(t, db)              // Clean before each subtest
		ctx := NewEnhancedDbContextWithDB(db) // Fresh context for each test
		// First insert a user
		user := TestUserPG{
			Name:     "Bob PostgreSQL",
			Email:    "bob.pg@example.com",
			Age:      30,
			IsActive: true,
		}

		ctx.Add(&user)
		_, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		userID := user.ID

		// Delete the user
		ctx.Delete(&user)
		affected, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		// Verify the user was deleted
		dbSet := NewEnhancedDbSet[TestUserPG](ctx)
		users, err := dbSet.Where("id = ?", userID).ToList()
		if err != nil {
			t.Fatalf("Failed to check deleted user: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("Expected 0 users after deletion, got %d", len(users))
		}
	})

	t.Run("complex query operations with PostgreSQL", func(t *testing.T) {
		cleanupPostgresDB(t, db)              // Clean before each subtest
		ctx := NewEnhancedDbContextWithDB(db) // Fresh context for each test
		// Insert multiple test users
		users := []TestUserPG{
			{Name: "User1", Email: "user1@pg.com", Age: 20, IsActive: true},
			{Name: "User2", Email: "user2@pg.com", Age: 25, IsActive: false},
			{Name: "User3", Email: "user3@pg.com", Age: 30, IsActive: true},
			{Name: "User4", Email: "user4@pg.com", Age: 35, IsActive: false},
		}

		for i := range users {
			ctx.Add(&users[i])
		}

		_, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to insert test users: %v", err)
		}

		dbSet := NewEnhancedDbSet[TestUserPG](ctx)

		// Test complex WHERE clauses
		activeUsers, err := dbSet.Where("is_active = ?", true).Where("age >= ?", 25).ToList()
		if err != nil {
			t.Fatalf("Failed to fetch active users: %v", err)
		}

		if len(activeUsers) != 1 { // Only User3 should match (active and age >= 25)
			t.Errorf("Expected 1 active user with age >= 25, got %d", len(activeUsers))
		}

		// Test ordering
		orderedUsers, err := dbSet.OrderBy("age DESC").Take(2).ToList()
		if err != nil {
			t.Fatalf("Failed to fetch ordered users: %v", err)
		}

		if len(orderedUsers) != 2 {
			t.Errorf("Expected 2 users, got %d", len(orderedUsers))
		}

		if orderedUsers[0].Age < orderedUsers[1].Age {
			t.Error("Users should be ordered by age DESC")
		}

		// Test Count
		count, err := dbSet.Where("age >= ?", 25).Count()
		if err != nil {
			t.Fatalf("Failed to count users: %v", err)
		}

		if count != 3 { // User2, User3, User4
			t.Errorf("Expected count 3, got %d", count)
		}

		// Test Any
		hasActiveUsers, err := dbSet.Where("is_active = ?", true).Any()
		if err != nil {
			t.Fatalf("Failed to check if any active users: %v", err)
		}

		if !hasActiveUsers {
			t.Error("Expected to have active users")
		}
	})
}

// TestPostgreSQLTransactions tests transaction handling with PostgreSQL
func TestPostgreSQLTransactions(t *testing.T) {
	db := setupPostgresDB(t)
	defer func() {
		cleanupPostgresDB(t, db)
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)

	t.Run("successful transaction", func(t *testing.T) {
		tx, err := ctx.Database.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		txCtx := NewEnhancedDbContextWithTx(tx)
		txCtx.driver = "postgres" // Set driver explicitly

		user := TestUserPG{
			Name:     "Transaction User",
			Email:    "tx.user@example.com",
			Age:      25,
			IsActive: true,
		}

		txCtx.Add(&user)
		_, err = txCtx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to save in transaction: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify the user was actually committed
		dbSet := NewEnhancedDbSet[TestUserPG](ctx)
		users, err := dbSet.Where("email = ?", "tx.user@example.com").ToList()
		if err != nil {
			t.Fatalf("Failed to fetch user after commit: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user after commit, got %d", len(users))
		}
	})

	t.Run("rolled back transaction", func(t *testing.T) {
		tx, err := ctx.Database.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		txCtx := NewEnhancedDbContextWithTx(tx)
		txCtx.driver = "postgres" // Set driver explicitly

		user := TestUserPG{
			Name:     "Rollback User",
			Email:    "rollback.user@example.com",
			Age:      30,
			IsActive: true,
		}

		txCtx.Add(&user)
		_, err = txCtx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to save in transaction: %v", err)
		}

		// Rollback instead of commit
		if err := tx.Rollback(); err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// Verify the user was NOT committed
		dbSet := NewEnhancedDbSet[TestUserPG](ctx)
		users, err := dbSet.Where("email = ?", "rollback.user@example.com").ToList()
		if err != nil {
			t.Fatalf("Failed to check for rolled back user: %v", err)
		}

		if len(users) != 0 {
			t.Errorf("Expected 0 users after rollback, got %d", len(users))
		}
	})
}

// TestPostgreSQLErrorHandling tests error scenarios with PostgreSQL
func TestPostgreSQLErrorHandling(t *testing.T) {
	db := setupPostgresDB(t)
	defer func() {
		cleanupPostgresDB(t, db)
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)

	t.Run("duplicate key error", func(t *testing.T) {
		user1 := TestUserPG{
			Name:     "Duplicate Test",
			Email:    "duplicate@example.com",
			Age:      25,
			IsActive: true,
		}

		user2 := TestUserPG{
			Name:     "Duplicate Test 2",
			Email:    "duplicate@example.com", // Same email - should cause constraint violation
			Age:      30,
			IsActive: true,
		}

		ctx.Add(&user1)
		_, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to insert first user: %v", err)
		}

		ctx.Add(&user2)
		_, err = ctx.SaveChanges()
		if err == nil {
			t.Error("Expected error due to duplicate email constraint")
		}
	})

	t.Run("invalid data type error", func(t *testing.T) {
		// This test might be tricky with the current struct-based approach
		// But we can test by trying to insert invalid data through direct SQL
		_, err := db.Exec("INSERT INTO test_users_pg (name, email, age) VALUES ($1, $2, $3)",
			"Test User", "test@example.com", "invalid_age")
		if err == nil {
			t.Error("Expected error due to invalid age data type")
		}
	})
}

// TestPostgreSQLPlaceholderConversion tests placeholder conversion for PostgreSQL
func TestPostgreSQLPlaceholderConversion(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single placeholder",
			input:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "multiple placeholders",
			input:    "SELECT * FROM users WHERE id = ? AND name = ? AND age > ?",
			expected: "SELECT * FROM users WHERE id = $1 AND name = $2 AND age > $3",
		},
		{
			name:     "no placeholders",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "placeholder in string literal",
			input:    "SELECT * FROM users WHERE note = 'Contains ? character' AND id = ?",
			expected: "SELECT * FROM users WHERE note = 'Contains $1 character' AND id = $2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertQueryPlaceholders(tc.input, "postgres")
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}

	// Test with non-PostgreSQL driver (should not convert)
	t.Run("sqlite3 driver", func(t *testing.T) {
		input := "SELECT * FROM users WHERE id = ? AND name = ?"
		result := convertQueryPlaceholders(input, "sqlite3")
		if result != input {
			t.Errorf("Expected no conversion for sqlite3, got '%s'", result)
		}
	})
}

// TestComplexPostgreSQLDataTypes tests complex data types and operations
func TestComplexPostgreSQLDataTypes(t *testing.T) {
	db := setupPostgresDB(t)
	defer func() {
		cleanupPostgresDB(t, db)
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	t.Run("decimal and json data types", func(t *testing.T) {
		cleanupPostgresDB(t, db)              // Clean before test
		ctx := NewEnhancedDbContextWithDB(db) // Fresh context
		user := UserPG{
			Username:   "complexuser",
			Email:      "complex@example.com",
			FullName:   "Complex User",
			Age:        28,
			Salary:     75000.50,
			IsVerified: true,
			Metadata:   `{"department": "engineering", "level": "senior"}`,
		}

		ctx.Add(&user)
		affected, err := ctx.SaveChanges()
		if err != nil {
			t.Fatalf("Failed to save complex user: %v", err)
		}

		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		if user.ID == 0 {
			t.Error("Expected ID to be set after insert")
		}

		// Fetch and verify
		dbSet := NewEnhancedDbSet[UserPG](ctx)
		users, err := dbSet.Where("username = ?", "complexuser").ToList()
		if err != nil {
			t.Fatalf("Failed to fetch complex user: %v", err)
		}

		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}

		fetchedUser := users[0]
		if fetchedUser.Salary != 75000.50 {
			t.Errorf("Expected salary 75000.50, got %f", fetchedUser.Salary)
		}

		if fetchedUser.Metadata != `{"department": "engineering", "level": "senior"}` &&
			fetchedUser.Metadata != `{"level": "senior", "department": "engineering"}` {
			t.Errorf("Expected metadata to match JSON content, got '%s'", fetchedUser.Metadata)
		}
	})
}

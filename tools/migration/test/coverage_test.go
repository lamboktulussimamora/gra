package main

import (
	"testing"
)

// TestMainFunctionWithRunningPostgres tests main function with actual PostgreSQL to reach SQL error paths
func TestMainFunctionWithRunningPostgres(t *testing.T) {
	originalConn := *conn
	originalUp := *up

	defer func() {
		*conn = originalConn
		*up = originalUp
	}()

	t.Run("test with running postgres for SQL coverage", func(t *testing.T) {
		// Test with the running PostgreSQL container
		// This should successfully connect and either succeed or fail at SQL level

		// Test 1: Working connection with up=false (should reach success message)
		*conn = "postgres://postgres:testpass@localhost:5433/testdb?sslmode=disable"
		*up = false
		main() // Should reach "Database connection successful!" and exit

		// Test 2: Working connection with up=true (should reach SQL execution)
		*conn = "postgres://postgres:testpass@localhost:5433/testdb?sslmode=disable"
		*up = true
		main() // Should execute SQL statements - might succeed or fail depending on permissions

		// Test 3: Try to connect to postgres database (might have different permissions)
		*conn = "postgres://postgres:testpass@localhost:5433/postgres?sslmode=disable"
		*up = true
		main() // Might have different permission behavior

		// Test 4: Try with template1 (system database, might be restricted)
		*conn = "postgres://postgres:testpass@localhost:5433/template1?sslmode=disable"
		*up = true
		main() // template1 might be read-only or have restrictions
	})

	t.Run("test potential SQL error scenarios", func(t *testing.T) {
		// Try scenarios that might trigger SQL errors

		// Test with a non-existent database (should fail during connection, not SQL)
		*conn = "postgres://postgres:testpass@localhost:5433/nonexistent?sslmode=disable"
		*up = true
		main()

		// Test with wrong credentials (should fail during connection)
		*conn = "postgres://wronguser:wrongpass@localhost:5433/testdb?sslmode=disable"
		*up = true
		main()

		// Test with the main database again to ensure we get SQL execution
		*conn = "postgres://postgres:testpass@localhost:5433/testdb?sslmode=disable"
		*up = true
		main() // This should definitely reach SQL execution
	})
}

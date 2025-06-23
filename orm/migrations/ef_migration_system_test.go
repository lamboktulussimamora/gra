package migrations

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMigrationState tests the MigrationState enum and its string representation
func TestMigrationState(t *testing.T) {
	tests := []struct {
		state    MigrationState
		expected string
	}{
		{MigrationStatePending, "Pending"},
		{MigrationStateApplied, "Applied"},
		{MigrationStateFailed, "Failed"},
		{MigrationState(999), "Unknown"}, // Test unknown state
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("Expected MigrationState(%d).String() = %s, got %s",
				int(test.state), test.expected, result)
		}
	}
}

// TestDetectDatabaseDriver tests the database driver detection functionality
func TestDetectDatabaseDriver(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{db: db}

	// Test SQLite detection
	driver := manager.detectDatabaseDriver()
	if !strings.Contains(driver, "sqlite") {
		t.Errorf("Expected driver to contain 'sqlite', got: %s", driver)
	}
}

// TestConvertQueryPlaceholders tests query placeholder conversion
func TestConvertQueryPlaceholders(t *testing.T) {
	manager := &EFMigrationManager{}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "no placeholders",
			query:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "single placeholder",
			query:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = ?",
		},
		{
			name:     "multiple placeholders",
			query:    "INSERT INTO users (name, email) VALUES (?, ?)",
			expected: "INSERT INTO users (name, email) VALUES (?, ?)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := manager.ConvertQueryPlaceholders(test.query)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

// TestConvertQueryPlaceholdersPostgreSQL tests PostgreSQL-specific placeholder conversion
func TestConvertQueryPlaceholdersPostgreSQL(t *testing.T) {
	manager := &EFMigrationManager{driver: "postgres"}

	tests := []struct {
		query    string
		expected string
	}{
		{
			query:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			query:    "INSERT INTO users (name, email) VALUES (?, ?)",
			expected: "INSERT INTO users (name, email) VALUES ($1, $2)",
		},
		{
			query:    "UPDATE users SET name = ?, email = ? WHERE id = ?",
			expected: "UPDATE users SET name = $1, email = $2 WHERE id = $3",
		},
	}

	for _, test := range tests {
		result := manager.convertQueryPlaceholders(test.query)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

// TestGetAutoIncrementSQL tests auto-increment SQL generation
func TestGetAutoIncrementSQL(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{"sqlite3", "AUTOINCREMENT"},
		{"postgres", ""},
		{"mysql", "AUTO_INCREMENT"},
		{"unknown", ""},
	}

	manager := &EFMigrationManager{}
	for _, test := range tests {
		manager.driver = test.driver
		result := manager.getAutoIncrementSQL()
		if result != test.expected {
			t.Errorf("For driver %s, expected %s, got %s",
				test.driver, test.expected, result)
		}
	}
}

// TestEnsureSchemaTables tests schema table creation
func TestEnsureSchemaTables(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{
		db:     db,
		driver: "sqlite3",
	}

	err = manager.ensureSchemaTables()
	if err != nil {
		t.Errorf("Failed to ensure schema tables: %v", err)
	}

	// Verify tables were created
	tables := []string{"__EFMigrationsHistory", "__EFMigrationSnapshots"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err = db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist", table)
		}
	}
}

// TestEnsureSchemaIndexes tests schema index creation
func TestEnsureSchemaIndexes(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{
		db:     db,
		driver: "sqlite3",
	}

	// First create the schema tables
	err = manager.ensureSchemaTables()
	if err != nil {
		t.Fatalf("Failed to ensure schema tables: %v", err)
	}

	// Then create indexes
	err = manager.ensureSchemaIndexes()
	if err != nil {
		t.Errorf("Failed to ensure schema indexes: %v", err)
	}
}

// TestDebugSQLiteSchema tests SQLite schema debugging
func TestDebugSQLiteSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{
		db:     db,
		driver: "sqlite3",
	}

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Debug schema (this should not error)
	manager.debugSQLiteSchema()
}

// TestEnsureSchema tests the complete schema initialization
func TestEnsureSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	config := DefaultEFMigrationConfig()
	manager, err := NewEFMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	err = manager.EnsureSchema()
	if err != nil {
		t.Errorf("Failed to ensure schema: %v", err)
	}

	// Verify schema was created properly
	tables := []string{"__EFMigrationsHistory", "__EFMigrationSnapshots"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err = db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist", table)
		}
	}
}

// TestNewEFMigrationManager tests the migration manager creation
func TestNewEFMigrationManager(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Test with default config
	config := DefaultEFMigrationConfig()
	manager, err := NewEFMigrationManager(db, config)
	if err != nil {
		t.Errorf("Failed to create migration manager: %v", err)
	}
	if manager == nil {
		t.Error("Expected non-nil migration manager")
	}

	// Test with nil database
	_, err = NewEFMigrationManager(nil, config)
	if err == nil {
		t.Error("Expected error when creating manager with nil database")
	}
}

// TestDefaultEFMigrationConfig tests the default configuration
func TestDefaultEFMigrationConfig(t *testing.T) {
	config := DefaultEFMigrationConfig()

	if config == nil {
		t.Error("Expected non-nil config")
	}

	if config.MigrationsTable == "" {
		t.Error("Expected non-empty migrations table name")
	}

	if config.SnapshotsTable == "" {
		t.Error("Expected non-empty snapshots table name")
	}
}

// TestMigrationManagerWithPostgreSQLCompatibility tests PostgreSQL-specific functionality
func TestMigrationManagerWithPostgreSQLCompatibility(t *testing.T) {
	// Since we can't easily test with actual PostgreSQL in unit tests,
	// we'll test the logic paths that would be taken for PostgreSQL

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{
		db:     db,
		driver: "postgres", // Set as postgres to test postgres-specific code paths
	}

	// Test PostgreSQL placeholder conversion
	query := "INSERT INTO users (name) VALUES (?)"
	converted := manager.convertQueryPlaceholders(query)
	expected := "INSERT INTO users (name) VALUES ($1)"
	if converted != expected {
		t.Errorf("Expected %s, got %s", expected, converted)
	}

	// Test PostgreSQL auto-increment SQL
	autoIncrement := manager.getAutoIncrementSQL()
	if autoIncrement != "" {
		t.Errorf("Expected empty string for PostgreSQL auto-increment, got %s", autoIncrement)
	}
}

// TestMigrationManagerDatabaseDriverDetection tests driver detection edge cases
func TestMigrationManagerDatabaseDriverDetection(t *testing.T) {
	// Test with actual SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := &EFMigrationManager{db: db}

	// Test driver detection
	driver := manager.detectDatabaseDriver()
	if driver == "" {
		t.Error("Expected non-empty driver detection result")
	}

	// The exact driver string may vary, but it should contain meaningful information
	if len(driver) < 3 {
		t.Errorf("Expected driver string to be at least 3 characters, got: %s", driver)
	}
}

// TestMigrationManagerErrorHandling tests error handling in various scenarios
func TestMigrationManagerErrorHandling(t *testing.T) {
	// Test with closed database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	db.Close() // Close immediately to test error handling

	manager := &EFMigrationManager{
		db:     db,
		driver: "sqlite3",
	}

	// This should fail because the database is closed
	err = manager.ensureSchemaTables()
	if err == nil {
		t.Error("Expected error when ensuring schema tables on closed database")
	}
}

// TestMigrationStateConstants tests all migration state constants
func TestMigrationStateConstants(t *testing.T) {
	// Test that constants have expected values
	if MigrationStatePending != 0 {
		t.Errorf("Expected MigrationStatePending to be 0, got %d", MigrationStatePending)
	}

	if MigrationStateApplied != 1 {
		t.Errorf("Expected MigrationStateApplied to be 1, got %d", MigrationStateApplied)
	}

	if MigrationStateFailed != 2 {
		t.Errorf("Expected MigrationStateFailed to be 2, got %d", MigrationStateFailed)
	}
}

// TestQueryPlaceholderEdgeCases tests edge cases in query placeholder conversion
func TestQueryPlaceholderEdgeCases(t *testing.T) {
	manager := &EFMigrationManager{driver: "postgres"}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "empty query",
			query:    "",
			expected: "",
		},
		{
			name:     "query without placeholders",
			query:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "query with ? in string literal",
			query:    "SELECT * FROM users WHERE name = 'test?'",
			expected: "SELECT * FROM users WHERE name = 'test$1'", // This tests the current behavior
		},
		{
			name:     "complex query with multiple placeholders",
			query:    "UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ? AND active = ?",
			expected: "UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE id = $4 AND active = $5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := manager.convertQueryPlaceholders(test.query)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

// TestAutoIncrementSQLForAllDrivers tests auto-increment SQL for all supported drivers
func TestAutoIncrementSQLForAllDrivers(t *testing.T) {
	manager := &EFMigrationManager{}

	drivers := []struct {
		name     string
		expected string
	}{
		{"sqlite3", "AUTOINCREMENT"},
		{"sqlite", "AUTOINCREMENT"},
		{"postgres", ""},
		{"postgresql", ""},
		{"mysql", "AUTO_INCREMENT"},
		{"mariadb", "AUTO_INCREMENT"},
		{"unknown", ""},
		{"", ""},
	}

	for _, driver := range drivers {
		t.Run(driver.name, func(t *testing.T) {
			manager.driver = driver.name
			result := manager.getAutoIncrementSQL()
			if result != driver.expected {
				t.Errorf("For driver %s, expected %s, got %s",
					driver.name, driver.expected, result)
			}
		})
	}
}

// TestEFMigrationManagerIntegration tests integration scenarios
func TestEFMigrationManagerIntegration(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Test complete initialization flow
	config := DefaultEFMigrationConfig()
	manager, err := NewEFMigrationManager(db, config)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	// Test schema initialization
	err = manager.EnsureSchema()
	if err != nil {
		t.Errorf("Failed to ensure schema: %v", err)
	}

	// Verify the manager is properly configured
	if manager.db == nil {
		t.Error("Expected non-nil database connection")
	}

	if manager.driver == "" {
		t.Error("Expected non-empty driver")
	}

	// Test that schema exists
	var historyCount, snapshotCount int

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='__EFMigrationsHistory'").Scan(&historyCount)
	if err != nil {
		t.Errorf("Failed to check for migrations history table: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='__EFMigrationSnapshots'").Scan(&snapshotCount)
	if err != nil {
		t.Errorf("Failed to check for snapshots table: %v", err)
	}

	if historyCount != 1 {
		t.Errorf("Expected migrations history table to exist")
	}

	if snapshotCount != 1 {
		t.Errorf("Expected snapshots table to exist")
	}
}

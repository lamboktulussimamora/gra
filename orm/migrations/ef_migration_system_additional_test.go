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

	manager := NewEFMigrationManager(db, nil)

	// Test SQLite detection
	driver := manager.detectDatabaseDriver()
	if !strings.Contains(string(driver), "sqlite") {
		t.Errorf("Expected driver to contain 'sqlite', got: %s", driver)
	}
}

// TestConvertQueryPlaceholders tests query placeholder conversion
func TestConvertQueryPlaceholders(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)
	manager.driver = PostgreSQL // Set to PostgreSQL for testing

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
		driver   DatabaseDriver
		expected string
	}{
		{SQLite, "AUTOINCREMENT"},
		{PostgreSQL, ""},
		{MySQL, "AUTO_INCREMENT"},
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)
	for _, test := range tests {
		manager.driver = test.driver
		result := manager.getAutoIncrementSQL()
		if result != test.expected {
			t.Errorf("For driver %s, expected %s, got %s",
				test.driver, test.expected, result)
		}
	}
}

// TestEnsureSchema tests the complete schema initialization
func TestEnsureSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	config := DefaultEFMigrationConfig()
	manager := NewEFMigrationManager(db, config)

	err = manager.EnsureSchema()
	if err != nil {
		t.Errorf("Failed to ensure schema: %v", err)
	}

	// Verify schema was created properly
	tables := []string{config.MigrationTable, config.HistoryTable, config.SnapshotTable}
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
	manager := NewEFMigrationManager(db, config)
	if manager == nil {
		t.Error("Expected non-nil migration manager")
	}

	// Test with nil config
	manager2 := NewEFMigrationManager(db, nil)
	if manager2 == nil {
		t.Error("Expected non-nil migration manager even with nil config")
	}
}

// TestDefaultEFMigrationConfig tests the default configuration
func TestDefaultEFMigrationConfig(t *testing.T) {
	config := DefaultEFMigrationConfig()

	if config == nil {
		t.Error("Expected non-nil config")
	}

	if config.MigrationTable == "" {
		t.Error("Expected non-empty migration table name")
	}

	if config.HistoryTable == "" {
		t.Error("Expected non-empty history table name")
	}

	if config.SnapshotTable == "" {
		t.Error("Expected non-empty snapshot table name")
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

	manager := NewEFMigrationManager(db, nil)
	manager.driver = PostgreSQL // Set as postgres to test postgres-specific code paths

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

	manager := NewEFMigrationManager(db, nil)

	// Test driver detection
	driver := manager.detectDatabaseDriver()
	if driver == "" {
		t.Error("Expected non-empty driver detection result")
	}

	// The exact driver string may vary, but it should contain meaningful information
	if len(string(driver)) < 3 {
		t.Errorf("Expected driver string to be at least 3 characters, got: %s", driver)
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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)
	manager.driver = PostgreSQL

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
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

	drivers := []struct {
		name     DatabaseDriver
		expected string
	}{
		{SQLite, "AUTOINCREMENT"},
		{PostgreSQL, ""},
		{MySQL, "AUTO_INCREMENT"},
	}

	for _, driver := range drivers {
		t.Run(string(driver.name), func(t *testing.T) {
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
	manager := NewEFMigrationManager(db, config)

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
	var historyCount, snapshotCount, migrationCount int

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", config.HistoryTable).Scan(&historyCount)
	if err != nil {
		t.Errorf("Failed to check for history table: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", config.SnapshotTable).Scan(&snapshotCount)
	if err != nil {
		t.Errorf("Failed to check for snapshots table: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", config.MigrationTable).Scan(&migrationCount)
	if err != nil {
		t.Errorf("Failed to check for migration table: %v", err)
	}

	if historyCount != 1 {
		t.Errorf("Expected history table to exist")
	}

	if snapshotCount != 1 {
		t.Errorf("Expected snapshots table to exist")
	}

	if migrationCount != 1 {
		t.Errorf("Expected migration table to exist")
	}
}

// TestEFMigrationManagerDebugFunctionality tests debug-related functions
func TestEFMigrationManagerDebugFunctionality(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

	// Create schema first
	err = manager.EnsureSchema()
	if err != nil {
		t.Fatalf("Failed to ensure schema: %v", err)
	}

	// Test debug functionality (should not error)
	manager.debugSQLiteSchema()
}

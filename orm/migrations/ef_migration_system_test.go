package migrations

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Setup test database for EF migration testing
func setupEFMigrationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

func TestNewEFMigrationManager(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	// Test with nil config (should use defaults)
	manager := NewEFMigrationManager(db, nil)
	if manager == nil {
		t.Fatal("NewEFMigrationManager should not return nil")
	}

	if manager.migrationTable != "__ef_migrations_history" {
		t.Errorf("Expected migration table '__ef_migrations_history', got '%s'", manager.migrationTable)
	}

	if manager.historyTable != "__ef_migration_history" {
		t.Errorf("Expected history table '__ef_migration_history', got '%s'", manager.historyTable)
	}

	if manager.snapshotTable != "__model_snapshot" {
		t.Errorf("Expected snapshot table '__model_snapshot', got '%s'", manager.snapshotTable)
	}

	// Test with custom config
	customConfig := &EFMigrationConfig{
		AutoMigrate:    true,
		MigrationTable: "custom_migrations",
		HistoryTable:   "custom_history",
		SnapshotTable:  "custom_snapshot",
		Logger:         log.New(os.Stdout, "TEST: ", log.LstdFlags),
	}

	manager2 := NewEFMigrationManager(db, customConfig)
	if manager2.autoMigrate != true {
		t.Error("Expected autoMigrate to be true")
	}

	if manager2.migrationTable != "custom_migrations" {
		t.Errorf("Expected migration table 'custom_migrations', got '%s'", manager2.migrationTable)
	}
}

func TestDefaultEFMigrationConfig(t *testing.T) {
	config := DefaultEFMigrationConfig()

	if config == nil {
		t.Fatal("DefaultEFMigrationConfig should not return nil")
	}

	if config.AutoMigrate != false {
		t.Error("Expected AutoMigrate to be false by default")
	}

	if config.MigrationTable != "__ef_migrations_history" {
		t.Errorf("Expected default migration table '__ef_migrations_history', got '%s'", config.MigrationTable)
	}

	if config.HistoryTable != "__ef_migration_history" {
		t.Errorf("Expected default history table '__ef_migration_history', got '%s'", config.HistoryTable)
	}

	if config.SnapshotTable != "__model_snapshot" {
		t.Errorf("Expected default snapshot table '__model_snapshot', got '%s'", config.SnapshotTable)
	}

	if config.Logger == nil {
		t.Error("Expected logger to be set in default config")
	}
}

func TestMigrationStateString(t *testing.T) {
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
			t.Errorf("Expected state %d to return '%s', got '%s'", test.state, test.expected, result)
		}
	}
}

func TestDetectDatabaseDriver(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

	// Should detect SQLite
	if manager.driver != SQLite {
		t.Errorf("Expected SQLite driver, got %v", manager.driver)
	}
}

func TestConvertQueryPlaceholders(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

	// Test SQLite placeholders (should remain unchanged)
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	result := manager.ConvertQueryPlaceholders(query)
	if result != query {
		t.Errorf("SQLite query should remain unchanged, got '%s'", result)
	}

	// Test PostgreSQL placeholders conversion
	manager.driver = PostgreSQL
	expected := "SELECT * FROM users WHERE id = $1 AND name = $2"
	result = manager.ConvertQueryPlaceholders(query)
	if result != expected {
		t.Errorf("Expected PostgreSQL query '%s', got '%s'", expected, result)
	}

	// Test query with no placeholders
	noPlaceholderQuery := "SELECT * FROM users"
	result = manager.ConvertQueryPlaceholders(noPlaceholderQuery)
	if result != noPlaceholderQuery {
		t.Errorf("Query with no placeholders should remain unchanged, got '%s'", result)
	}

	// Test MySQL (should remain unchanged like SQLite)
	manager.driver = MySQL
	result = manager.ConvertQueryPlaceholders(query)
	if result != query {
		t.Errorf("MySQL query should remain unchanged, got '%s'", result)
	}
}

func TestMigrationStruct(t *testing.T) {
	migration := Migration{
		ID:          "20231201120000_InitialCreate",
		Name:        "InitialCreate",
		Version:     20231201120000,
		Description: "Initial database creation",
		UpSQL:       "CREATE TABLE users (id INTEGER PRIMARY KEY);",
		DownSQL:     "DROP TABLE users;",
		AppliedAt:   time.Now(),
		State:       MigrationStateApplied,
	}

	// Test that migration struct fields are properly set
	if migration.ID != "20231201120000_InitialCreate" {
		t.Errorf("Expected ID '20231201120000_InitialCreate', got '%s'", migration.ID)
	}

	if migration.Name != "InitialCreate" {
		t.Errorf("Expected Name 'InitialCreate', got '%s'", migration.Name)
	}

	if migration.Version != 20231201120000 {
		t.Errorf("Expected Version 20231201120000, got %d", migration.Version)
	}

	if migration.State != MigrationStateApplied {
		t.Errorf("Expected State MigrationStateApplied, got %v", migration.State)
	}
}

func TestMigrationHistory(t *testing.T) {
	now := time.Now()

	history := MigrationHistory{
		Applied: []Migration{
			{
				ID:        "001_Initial",
				Name:      "Initial",
				State:     MigrationStateApplied,
				AppliedAt: now,
			},
		},
		Pending: []Migration{
			{
				ID:    "002_AddUsers",
				Name:  "AddUsers",
				State: MigrationStatePending,
			},
		},
		Failed: []Migration{
			{
				ID:    "003_AddProducts",
				Name:  "AddProducts",
				State: MigrationStateFailed,
			},
		},
	}

	// Test that history contains expected migrations
	if len(history.Applied) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(history.Applied))
	}

	if len(history.Pending) != 1 {
		t.Errorf("Expected 1 pending migration, got %d", len(history.Pending))
	}

	if len(history.Failed) != 1 {
		t.Errorf("Expected 1 failed migration, got %d", len(history.Failed))
	}

	if history.Applied[0].State != MigrationStateApplied {
		t.Errorf("Expected applied migration state to be Applied, got %v", history.Applied[0].State)
	}

	if history.Applied[0].AppliedAt.IsZero() {
		t.Error("Expected applied migration to have AppliedAt time set")
	}
}

func TestEFMigrationManagerFields(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)

	// Test that manager fields are properly initialized
	if manager.db == nil {
		t.Error("Expected db to be set")
	}

	if manager.logger == nil {
		t.Error("Expected logger to be set")
	}

	if manager.loadedMigrations == nil {
		t.Error("Expected loadedMigrations to be initialized")
	}

	if manager.pendingMigrations == nil {
		t.Error("Expected pendingMigrations to be initialized")
	}

	if len(manager.loadedMigrations) != 0 {
		t.Errorf("Expected loadedMigrations to be empty initially, got %d", len(manager.loadedMigrations))
	}

	if len(manager.pendingMigrations) != 0 {
		t.Errorf("Expected pendingMigrations to be empty initially, got %d", len(manager.pendingMigrations))
	}
}

func TestComplexQueryPlaceholderConversion(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	manager := NewEFMigrationManager(db, nil)
	manager.driver = PostgreSQL

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple INSERT",
			input:    "INSERT INTO users (name, email) VALUES (?, ?)",
			expected: "INSERT INTO users (name, email) VALUES ($1, $2)",
		},
		{
			name:     "Complex query with multiple clauses",
			input:    "SELECT u.*, p.name FROM users u JOIN profiles p ON u.id = p.user_id WHERE u.active = ? AND p.updated_at > ? ORDER BY u.created_at",
			expected: "SELECT u.*, p.name FROM users u JOIN profiles p ON u.id = p.user_id WHERE u.active = $1 AND p.updated_at > $2 ORDER BY u.created_at",
		},
		{
			name:     "UPDATE with WHERE",
			input:    "UPDATE users SET name = ?, email = ? WHERE id = ?",
			expected: "UPDATE users SET name = $1, email = $2 WHERE id = $3",
		},
		{
			name:     "DELETE with multiple conditions",
			input:    "DELETE FROM users WHERE active = ? OR created_at < ?",
			expected: "DELETE FROM users WHERE active = $1 OR created_at < $2",
		},
		{
			name:     "Query with no placeholders",
			input:    "SELECT COUNT(*) FROM users",
			expected: "SELECT COUNT(*) FROM users",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := manager.ConvertQueryPlaceholders(test.input)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func TestMigrationConfigValidation(t *testing.T) {
	db := setupEFMigrationTestDB(t)
	defer db.Close()

	// Test with empty strings in config
	config := &EFMigrationConfig{
		AutoMigrate:    false,
		MigrationTable: "",  // Empty string
		HistoryTable:   "",  // Empty string
		SnapshotTable:  "",  // Empty string
		Logger:         nil, // Nil logger
	}

	manager := NewEFMigrationManager(db, config)

	// Should handle empty strings gracefully
	if manager.migrationTable != "" {
		t.Errorf("Expected empty migration table, got '%s'", manager.migrationTable)
	}

	if manager.historyTable != "" {
		t.Errorf("Expected empty history table, got '%s'", manager.historyTable)
	}

	if manager.snapshotTable != "" {
		t.Errorf("Expected empty snapshot table, got '%s'", manager.snapshotTable)
	}

	// Should handle nil logger gracefully
	if manager.logger != nil {
		t.Error("Expected logger to be nil when passed nil")
	}
}

func TestMigrationManagerWithClosedDB(t *testing.T) {
	db := setupEFMigrationTestDB(t)

	manager := NewEFMigrationManager(db, nil)

	// Close the database
	db.Close()

	// Test driver detection with closed database
	driver := manager.detectDatabaseDriver()

	// Should default to SQLite when detection fails
	if driver != SQLite {
		t.Errorf("Expected SQLite driver when detection fails, got %v", driver)
	}
}

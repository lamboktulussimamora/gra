package dbcontext

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestDetectDatabaseDriver tests the detectDatabaseDriver function
func TestDetectDatabaseDriverExtended(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	// Test SQLite detection
	driver := detectDatabaseDriver(db)
	if driver != "sqlite3" {
		t.Errorf("Expected driver to be 'sqlite3', got '%s'", driver)
	}
}

// TestConvertQueryPlaceholdersExtended tests the convertQueryPlaceholders function
func TestConvertQueryPlaceholdersExtended(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		driver   string
		expected string
	}{
		{
			name:     "SQLite placeholders unchanged",
			query:    "SELECT * FROM users WHERE id = ? AND name = ?",
			driver:   "sqlite3",
			expected: "SELECT * FROM users WHERE id = ? AND name = ?",
		},
		{
			name:     "MySQL placeholders unchanged",
			query:    "SELECT * FROM users WHERE id = ? AND name = ?",
			driver:   "mysql",
			expected: "SELECT * FROM users WHERE id = ? AND name = ?",
		},
		{
			name:     "PostgreSQL placeholders converted",
			query:    "SELECT * FROM users WHERE id = ? AND name = ?",
			driver:   "postgres",
			expected: "SELECT * FROM users WHERE id = $1 AND name = $2",
		},
		{
			name:     "PostgreSQL single placeholder",
			query:    "SELECT * FROM users WHERE id = ?",
			driver:   "postgres",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "PostgreSQL no placeholders",
			query:    "SELECT * FROM users",
			driver:   "postgres",
			expected: "SELECT * FROM users",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertQueryPlaceholders(test.query, test.driver)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

// TestNewDatabaseExtended tests the NewDatabase function
func TestNewDatabaseExtended(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	// Test NewDatabase
	database := NewDatabase(db)
	if database == nil {
		t.Fatal("Expected database to be created, got nil")
	}

	if database.db != db {
		t.Error("Expected database.db to be set correctly")
	}
}

// TestDatabaseBeginExtended tests the Begin transaction method
func TestDatabaseBeginExtended(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	database := NewDatabase(db)

	// Test beginning a transaction
	tx, err := database.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	if tx == nil {
		t.Fatal("Expected transaction to be created, got nil")
	}

	// Clean up transaction
	tx.Rollback()
}

// TestNewChangeTrackerExtended tests the NewChangeTracker function
func TestNewChangeTrackerExtended(t *testing.T) {
	tracker := NewChangeTracker()

	if tracker == nil {
		t.Fatal("Expected change tracker to be created, got nil")
	}

	if tracker.entities == nil {
		t.Error("Expected entities map to be initialized")
	}
}

// TestChangeTrackerEntityStatesExtended tests the entity state management in ChangeTracker
func TestChangeTrackerEntityStatesExtended(t *testing.T) {
	tracker := NewChangeTracker()

	testEntity := struct {
		ID   int
		Name string
	}{
		ID:   1,
		Name: "Test",
	}

	// Test getting state of untracked entity
	state := tracker.GetEntityState(&testEntity)
	if state != EntityStateUnchanged {
		t.Errorf("Expected state to be Unchanged for untracked entity, got %v", state)
	}

	// Test setting entity state
	tracker.SetEntityState(&testEntity, EntityStateAdded)
	state = tracker.GetEntityState(&testEntity)
	if state != EntityStateAdded {
		t.Errorf("Expected state to be Added, got %v", state)
	}

	// Test tracking entity
	tracker.TrackEntity(&testEntity, EntityStateUnchanged)
	state = tracker.GetEntityState(&testEntity)
	if state != EntityStateUnchanged {
		t.Errorf("Expected state to be Unchanged after tracking, got %v", state)
	}
}

// TestChangeTrackerStringExtended tests the String method of ChangeTracker
func TestChangeTrackerStringExtended(t *testing.T) {
	tracker := NewChangeTracker()

	// Test empty tracker
	result := tracker.String()
	if !strings.Contains(result, "No tracked entities") {
		t.Error("Expected string to indicate no tracked entities")
	}

	testEntity1 := struct {
		ID   int
		Name string
	}{ID: 1, Name: "Test1"}

	testEntity2 := struct {
		ID   int
		Name string
	}{ID: 2, Name: "Test2"}

	// Track some entities with different states
	tracker.SetEntityState(&testEntity1, EntityStateAdded)
	tracker.SetEntityState(&testEntity2, EntityStateModified)

	result = tracker.String()
	if result == "" {
		t.Error("Expected non-empty string representation")
	}

	// Should contain information about tracked entities
	if !strings.Contains(result, "2 tracked entities") {
		t.Error("Expected string to contain '2 tracked entities'")
	}
}

// TestEntityStateStringExtended tests the String method of EntityState
func TestEntityStateStringExtended(t *testing.T) {
	tests := []struct {
		state    EntityState
		expected string
	}{
		{EntityStateUnchanged, "Unchanged"},
		{EntityStateAdded, "Added"},
		{EntityStateModified, "Modified"},
		{EntityStateDeleted, "Deleted"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.state.String()
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

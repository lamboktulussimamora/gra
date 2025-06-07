package main

import (
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/migrations"
	_ "github.com/mattn/go-sqlite3"
)

func TestUserStruct(t *testing.T) {
	user := User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected Name to be 'Test User', got %s", user.Name)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestPostStruct(t *testing.T) {
	post := Post{
		ID:          1,
		UserID:      1,
		Title:       "Test Post",
		Content:     "This is a test post content",
		IsPublished: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if post.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", post.ID)
	}
	if post.UserID != 1 {
		t.Errorf("Expected UserID to be 1, got %d", post.UserID)
	}
	if post.Title != "Test Post" {
		t.Errorf("Expected Title to be 'Test Post', got %s", post.Title)
	}
	if post.IsPublished {
		t.Error("Expected IsPublished to be false")
	}
}

func TestCommentStruct(t *testing.T) {
	comment := Comment{
		ID:        1,
		PostID:    1,
		UserID:    1,
		Content:   "This is a test comment",
		CreatedAt: time.Now(),
	}

	if comment.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", comment.ID)
	}
	if comment.PostID != 1 {
		t.Errorf("Expected PostID to be 1, got %d", comment.PostID)
	}
	if comment.UserID != 1 {
		t.Errorf("Expected UserID to be 1, got %d", comment.UserID)
	}
	if comment.Content != "This is a test comment" {
		t.Errorf("Expected Content to be 'This is a test comment', got %s", comment.Content)
	}
}

func TestUserStructTags(t *testing.T) {
	userType := reflect.TypeOf(User{})

	// Test ID field tags
	idField, found := userType.FieldByName("ID")
	if !found {
		t.Fatal("ID field not found")
	}

	if idField.Tag.Get("db") != "id" {
		t.Errorf("Expected db tag to be 'id', got %s", idField.Tag.Get("db"))
	}
	if idField.Tag.Get("migration") != "primary_key,auto_increment" {
		t.Errorf("Expected migration tag to be 'primary_key,auto_increment', got %s", idField.Tag.Get("migration"))
	}

	// Test Email field tags
	emailField, found := userType.FieldByName("Email")
	if !found {
		t.Fatal("Email field not found")
	}

	if emailField.Tag.Get("db") != "email" {
		t.Errorf("Expected db tag to be 'email', got %s", emailField.Tag.Get("db"))
	}
	if emailField.Tag.Get("migration") != "unique,not_null,max_length:255" {
		t.Errorf("Expected migration tag to be 'unique,not_null,max_length:255', got %s", emailField.Tag.Get("migration"))
	}
}

func TestPostStructTags(t *testing.T) {
	postType := reflect.TypeOf(Post{})

	// Test UserID field tags for foreign key
	userIDField, found := postType.FieldByName("UserID")
	if !found {
		t.Fatal("UserID field not found")
	}

	migrationTag := userIDField.Tag.Get("migration")
	if !strings.Contains(migrationTag, "foreign_key:users.id") {
		t.Errorf("Expected migration tag to contain 'foreign_key:users.id', got %s", migrationTag)
	}

	// Test Content field for TEXT type
	contentField, found := postType.FieldByName("Content")
	if !found {
		t.Fatal("Content field not found")
	}

	migrationTag = contentField.Tag.Get("migration")
	if !strings.Contains(migrationTag, "type:TEXT") {
		t.Errorf("Expected migration tag to contain 'type:TEXT', got %s", migrationTag)
	}
}

func TestInitializeDatabase_WithSQLite(t *testing.T) {
	// Test with SQLite (in-memory) since we don't have PostgreSQL in tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close test database: %v", closeErr)
		}
	}()

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestSetupMigrator(t *testing.T) {
	// Create SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close test database: %v", closeErr)
		}
	}()

	// Create temporary directory for migrations
	tempDir, err := os.MkdirTemp("", "test_migrations")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp directory: %v", err)
		}
	}()

	// Test migrator setup
	migrator := migrations.NewHybridMigrator(
		db,
		migrations.SQLite, // Use SQLite for testing
		tempDir,
	)

	if migrator == nil {
		t.Fatal("Migrator should not be nil")
	}

	// Test registering models (this should not panic)
	migrator.DbSet(&User{})
	migrator.DbSet(&Post{})
	migrator.DbSet(&Comment{})
}

func TestDisplayMigrationStatus_WithEmptyStatus(_ *testing.T) {
	status := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{},
		PendingMigrations: []*migrations.MigrationFile{},
		HasPendingChanges: false,
		Summary:           "",
	}

	// Test that this doesn't panic
	displayMigrationStatus(status)
}

func TestDisplayMigrationStatus_WithData(_ *testing.T) {
	// Create mock migration files
	appliedMigration1 := &migrations.MigrationFile{
		Name:      "001_initial",
		Timestamp: time.Now().Add(-2 * time.Hour),
	}
	appliedMigration2 := &migrations.MigrationFile{
		Name:      "002_add_users",
		Timestamp: time.Now().Add(-1 * time.Hour),
	}
	pendingMigration := &migrations.MigrationFile{
		Name:      "003_add_posts",
		Timestamp: time.Now(),
	}

	status := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{appliedMigration1, appliedMigration2},
		PendingMigrations: []*migrations.MigrationFile{pendingMigration},
		HasPendingChanges: true,
		Summary:           "Add posts table",
	}

	// Test that this doesn't panic
	displayMigrationStatus(status)
}

func TestDisplayMigrationFileInfo_WithBasicInfo(_ *testing.T) {
	// Create a mock migration file with minimal required fields
	migrationFile := &migrations.MigrationFile{
		Filename: "test_migration.sql",
		Changes:  []migrations.MigrationChange{}, // Empty changes
	}

	// Test that this doesn't panic
	displayMigrationFileInfo(migrationFile)
}

// Mock implementation to test the main workflow without actual database operations
func TestMainWorkflow_Logic(t *testing.T) {
	// Test the logical components that main() uses

	// Test struct creation
	user := User{}
	post := Post{}
	comment := Comment{}

	// Verify structs are properly initialized with zero values
	if user.ID != 0 {
		t.Errorf("Expected User.ID to be 0, got %d", user.ID)
	}
	if post.UserID != 0 {
		t.Errorf("Expected Post.UserID to be 0, got %d", post.UserID)
	}
	if comment.PostID != 0 {
		t.Errorf("Expected Comment.PostID to be 0, got %d", comment.PostID)
	}

	// Test type reflection (used in schema generation)
	userType := reflect.TypeOf(User{})
	if userType.Kind() != reflect.Struct {
		t.Errorf("Expected User to be a struct, got %v", userType.Kind())
	}

	// Check that all expected fields exist
	expectedFields := []string{"ID", "Email", "Name", "IsActive", "CreatedAt", "UpdatedAt"}
	for _, fieldName := range expectedFields {
		_, found := userType.FieldByName(fieldName)
		if !found {
			t.Errorf("Expected field %s not found in User struct", fieldName)
		}
	}
}

func TestStructFieldTypes(t *testing.T) {
	// Test that all struct fields have the correct types
	userType := reflect.TypeOf(User{})

	testCases := []struct {
		fieldName    string
		expectedType reflect.Type
	}{
		{"ID", reflect.TypeOf(int64(0))},
		{"Email", reflect.TypeOf("")},
		{"Name", reflect.TypeOf("")},
		{"IsActive", reflect.TypeOf(true)},
		{"CreatedAt", reflect.TypeOf(time.Time{})},
		{"UpdatedAt", reflect.TypeOf(time.Time{})},
	}

	for _, tc := range testCases {
		field, found := userType.FieldByName(tc.fieldName)
		if !found {
			t.Errorf("Field %s not found", tc.fieldName)
			continue
		}

		if field.Type != tc.expectedType {
			t.Errorf("Field %s has type %v, expected %v", tc.fieldName, field.Type, tc.expectedType)
		}
	}
}

// Test that foreign key relationships are properly defined
func TestForeignKeyRelationships(t *testing.T) {
	// Test Post -> User relationship
	postType := reflect.TypeOf(Post{})
	userIDField, found := postType.FieldByName("UserID")
	if !found {
		t.Fatal("UserID field not found in Post struct")
	}

	migrationTag := userIDField.Tag.Get("migration")
	if !strings.Contains(migrationTag, "foreign_key:users.id") {
		t.Errorf("Post.UserID should have foreign key to users.id, got migration tag: %s", migrationTag)
	}

	// Test Comment -> Post relationship
	commentType := reflect.TypeOf(Comment{})
	postIDField, found := commentType.FieldByName("PostID")
	if !found {
		t.Fatal("PostID field not found in Comment struct")
	}

	migrationTag = postIDField.Tag.Get("migration")
	if !strings.Contains(migrationTag, "foreign_key:posts.id") {
		t.Errorf("Comment.PostID should have foreign key to posts.id, got migration tag: %s", migrationTag)
	}

	// Test Comment -> User relationship
	commentUserIDField, found := commentType.FieldByName("UserID")
	if !found {
		t.Fatal("UserID field not found in Comment struct")
	}

	migrationTag = commentUserIDField.Tag.Get("migration")
	if !strings.Contains(migrationTag, "foreign_key:users.id") {
		t.Errorf("Comment.UserID should have foreign key to users.id, got migration tag: %s", migrationTag)
	}
}

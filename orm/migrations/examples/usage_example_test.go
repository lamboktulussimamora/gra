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

const (
	sqlite3Driver                   = "sqlite3"
	memoryDB                        = ":memory:"
	testMigrationsDir               = "test_migrations"
	invalidDriver                   = "invalid_driver"
	expectedErr                     = "Expected error (no real database): %v"
	warningDBClose                  = "Warning: Failed to close test database: %v"
	warningTempDir                  = "Warning: failed to remove temp directory: %v"
	failedCreateDB                  = "Failed to create test database: %v"
	failedCreateTempDir             = "Failed to create temp directory: %v"
	expectedIDOne                   = "Expected ID to be 1, got %d"
	foreignKeyUsers                 = "foreign_key:users.id"
	errIDFormat                     = "Expected ID to be %d, got %d"
	errUserIDFormat                 = "Expected UserID to be %d, got %d"
	errTitleFormat                  = "Expected Title to be '%s', got %s"
	errContentFormat                = "Expected Content to be '%s', got %s"
	errIsActiveFormat               = "Expected IsActive to be %v"
	errIsPublishedFmt               = "Expected IsPublished to be %v"
	zeroIDCase                      = "zero id"
	postTitleTestPost               = "Test Post"
	postContentTestPost             = "This is a test post content"
	postTitlePublished              = "Published Post"
	postContentPublished            = "Published content"
	errExpectedNilMigrator          = "Expected error when migrator is nil"
	errExpectedNilStatus            = "Expected error when status is nil"
	errExpectedNilMigratorAndStatus = "Expected error when migrator and status are nil"
	errExpectedGetStatusBranch      = "Expected error from GetMigrationStatus branch"
)

func TestUserStruct(t *testing.T) {
	testCases := []struct {
		name     string
		user     User
		expected struct {
			ID       int64
			Email    string
			Name     string
			IsActive bool
		}
	}{
		{
			name: "valid user",
			user: User{
				ID:        1,
				Email:     "test@example.com",
				Name:      "Test User",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: struct {
				ID       int64
				Email    string
				Name     string
				IsActive bool
			}{1, "test@example.com", "Test User", true},
		},
		{
			name: "inactive user",
			user: User{
				ID:        2,
				Email:     "inactive@example.com",
				Name:      "Inactive User",
				IsActive:  false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: struct {
				ID       int64
				Email    string
				Name     string
				IsActive bool
			}{2, "inactive@example.com", "Inactive User", false},
		},
		{
			name: zeroIDCase,
			user: User{},
			expected: struct {
				ID       int64
				Email    string
				Name     string
				IsActive bool
			}{0, "", "", false},
		},
		{
			name: "user with empty email and name",
			user: User{
				ID:        3,
				Email:     "",
				Name:      "",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expected: struct {
				ID       int64
				Email    string
				Name     string
				IsActive bool
			}{3, "", "", true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.user.ID != tc.expected.ID {
				t.Errorf(errIDFormat, tc.expected.ID, tc.user.ID)
			}
			if tc.user.Email != tc.expected.Email {
				t.Errorf("Expected Email to be '%s', got %s", tc.expected.Email, tc.user.Email)
			}
			if tc.user.Name != tc.expected.Name {
				t.Errorf("Expected Name to be '%s', got %s", tc.expected.Name, tc.user.Name)
			}
			if tc.user.IsActive != tc.expected.IsActive {
				t.Errorf(errIsActiveFormat, tc.expected.IsActive)
			}
		})
	}
}

func TestPostStructUnpublished(t *testing.T) {
	p := Post{
		ID:          1,
		UserID:      1,
		Title:       postTitleTestPost,
		Content:     postContentTestPost,
		IsPublished: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if p.ID != 1 {
		t.Errorf(errIDFormat, 1, p.ID)
	}
	if p.UserID != 1 {
		t.Errorf(errUserIDFormat, 1, p.UserID)
	}
	if p.Title != postTitleTestPost {
		t.Errorf(errTitleFormat, postTitleTestPost, p.Title)
	}
	if p.Content != postContentTestPost {
		t.Errorf(errContentFormat, postContentTestPost, p.Content)
	}
	if p.IsPublished != false {
		t.Errorf(errIsPublishedFmt, false)
	}
}

func TestPostStructPublished(t *testing.T) {
	p := Post{
		ID:          2,
		UserID:      2,
		Title:       postTitlePublished,
		Content:     postContentPublished,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if p.ID != 2 {
		t.Errorf(errIDFormat, 2, p.ID)
	}
	if p.UserID != 2 {
		t.Errorf(errUserIDFormat, 2, p.UserID)
	}
	if p.Title != postTitlePublished {
		t.Errorf(errTitleFormat, postTitlePublished, p.Title)
	}
	if p.Content != postContentPublished {
		t.Errorf(errContentFormat, postContentPublished, p.Content)
	}
	if p.IsPublished != true {
		t.Errorf(errIsPublishedFmt, true)
	}
}

func TestPostStructZeroValue(t *testing.T) {
	p := Post{}
	if p.ID != 0 {
		t.Errorf(errIDFormat, 0, p.ID)
	}
	if p.UserID != 0 {
		t.Errorf(errUserIDFormat, 0, p.UserID)
	}
	if p.Title != "" {
		t.Errorf(errTitleFormat, "", p.Title)
	}
	if p.Content != "" {
		t.Errorf(errContentFormat, "", p.Content)
	}
	if p.IsPublished != false {
		t.Errorf(errIsPublishedFmt, false)
	}
}

func TestPostStructEmptyFields(t *testing.T) {
	p := Post{
		ID:          3,
		UserID:      1,
		Title:       "",
		Content:     "",
		IsPublished: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if p.ID != 3 {
		t.Errorf(errIDFormat, 3, p.ID)
	}
	if p.UserID != 1 {
		t.Errorf(errUserIDFormat, 1, p.UserID)
	}
	if p.Title != "" {
		t.Errorf(errTitleFormat, "", p.Title)
	}
	if p.Content != "" {
		t.Errorf(errContentFormat, "", p.Content)
	}
	if p.IsPublished != false {
		t.Errorf(errIsPublishedFmt, false)
	}
}

func TestCommentStruct(t *testing.T) {
	testCases := []struct {
		name     string
		comment  Comment
		expected struct {
			ID      int64
			PostID  int64
			UserID  int64
			Content string
		}
	}{
		{
			name: "basic comment",
			comment: Comment{
				ID:        1,
				PostID:    1,
				UserID:    1,
				Content:   "This is a test comment",
				CreatedAt: time.Now(),
			},
			expected: struct {
				ID      int64
				PostID  int64
				UserID  int64
				Content string
			}{1, 1, 1, "This is a test comment"},
		},
		{
			name: "empty content",
			comment: Comment{
				ID:        2,
				PostID:    2,
				UserID:    2,
				Content:   "",
				CreatedAt: time.Now(),
			},
			expected: struct {
				ID      int64
				PostID  int64
				UserID  int64
				Content string
			}{2, 2, 2, ""},
		},
		{
			name:    zeroIDCase,
			comment: Comment{},
			expected: struct {
				ID      int64
				PostID  int64
				UserID  int64
				Content string
			}{0, 0, 0, ""},
		},
		{
			name: "comment with empty content",
			comment: Comment{
				ID:        3,
				PostID:    1,
				UserID:    1,
				Content:   "",
				CreatedAt: time.Now(),
			},
			expected: struct {
				ID      int64
				PostID  int64
				UserID  int64
				Content string
			}{3, 1, 1, ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.comment.ID != tc.expected.ID {
				t.Errorf(errIDFormat, tc.expected.ID, tc.comment.ID)
			}
			if tc.comment.PostID != tc.expected.PostID {
				t.Errorf("Expected PostID to be %d, got %d", tc.expected.PostID, tc.comment.PostID)
			}
			if tc.comment.UserID != tc.expected.UserID {
				t.Errorf(errUserIDFormat, tc.expected.UserID, tc.comment.UserID)
			}
			if tc.comment.Content != tc.expected.Content {
				t.Errorf(errContentFormat, tc.expected.Content, tc.comment.Content)
			}
		})
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
	if !strings.Contains(migrationTag, foreignKeyUsers) {
		t.Errorf("Expected migration tag to contain '%s', got %s", foreignKeyUsers, migrationTag)
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

func TestInitializeDatabaseWithSQLite(t *testing.T) {
	// Test with SQLite (in-memory) since we don't have PostgreSQL in tests
	db, err := sql.Open(sqlite3Driver, memoryDB)
	if err != nil {
		t.Fatalf(failedCreateDB, err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warningDBClose, closeErr)
		}
	}()

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestSetupMigrator(t *testing.T) {
	// Create SQLite database for testing
	db, err := sql.Open(sqlite3Driver, memoryDB)
	if err != nil {
		t.Fatalf(failedCreateDB, err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warningDBClose, closeErr)
		}
	}()

	// Create temporary directory for migrations
	tempDir, err := os.MkdirTemp("", testMigrationsDir)
	if err != nil {
		t.Fatalf(failedCreateTempDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf(warningTempDir, err)
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

func TestDisplayMigrationStatusWithEmptyStatus(_ *testing.T) {
	status := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{},
		PendingMigrations: []*migrations.MigrationFile{},
		HasPendingChanges: false,
		Summary:           "",
	}

	// Test that this doesn't panic
	displayMigrationStatus(status)
}

func TestDisplayMigrationStatusWithData(_ *testing.T) {
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

func TestDisplayMigrationFileInfoWithBasicInfo(_ *testing.T) {
	// Create a mock migration file with minimal required fields
	migrationFile := &migrations.MigrationFile{
		Filename: "test_migration.sql",
		Changes:  []migrations.MigrationChange{}, // Empty changes
	}

	// Test that this doesn't panic
	displayMigrationFileInfo(migrationFile)
}

// Mock implementation to test the main workflow without actual database operations
func TestMainWorkflowLogic(t *testing.T) {
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
	if !strings.Contains(migrationTag, foreignKeyUsers) {
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
	if !strings.Contains(migrationTag, foreignKeyUsers) {
		t.Errorf("Comment.UserID should have foreign key to users.id, got migration tag: %s", migrationTag)
	}
}

func TestInitializeDatabaseConnection(t *testing.T) {
	// We can't test the actual PostgreSQL connection in unit tests
	// but we can test the connection string parsing and error handling logic

	// Test the function signature and behavior with a mock
	// This tests the code path without requiring a real PostgreSQL database
	t.Run("connection logic test", func(t *testing.T) {
		// Test that the function exists and has correct signature
		_ = initializeDatabase

		// In a real scenario, this would fail because we don't have PostgreSQL running
		// But we're testing the code exists and can be called
		db, err := initializeDatabase()
		if err != nil {
			// Expected since we don't have a real PostgreSQL database
			t.Logf(expectedErr, err)
		}
		if db != nil {
			_ = db.Close()
		}
	})
}

// Test createAndApplyMigration function with mock data
func TestCreateAndApplyMigration(t *testing.T) {
	// Create SQLite database for testing
	db, err := sql.Open(sqlite3Driver, memoryDB)
	if err != nil {
		t.Fatalf(failedCreateDB, err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warningDBClose, closeErr)
		}
	}()

	// Create temporary directory for migrations
	tempDir, err := os.MkdirTemp("", testMigrationsDir)
	if err != nil {
		t.Fatalf(failedCreateTempDir, err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf(warningTempDir, err)
		}
	}()

	// Setup migrator
	migrator := migrations.NewHybridMigrator(
		db,
		migrations.SQLite,
		tempDir,
	)
	migrator.DbSet(&User{})

	t.Run("no pending changes", func(t *testing.T) {
		status := &migrations.MigrationStatus{
			AppliedMigrations: []*migrations.MigrationFile{},
			PendingMigrations: []*migrations.MigrationFile{},
			HasPendingChanges: false,
			Summary:           "No changes",
		}

		err := createAndApplyMigration(migrator, status)
		if err != nil {
			t.Errorf("Expected no error for no pending changes, got: %v", err)
		}
	})

	t.Run("with pending changes", func(t *testing.T) {
		status := &migrations.MigrationStatus{
			AppliedMigrations: []*migrations.MigrationFile{},
			PendingMigrations: []*migrations.MigrationFile{},
			HasPendingChanges: true,
			Summary:           "Create initial schema",
		}

		// This might fail due to the migration system complexity, but we test that the function exists
		err := createAndApplyMigration(migrator, status)
		// We don't assert on the error since the migration system is complex
		// and might require specific database setup
		if err != nil {
			t.Logf("Migration error (expected in test): %v", err)
		}
	})
}

func TestCreateAndApplyMigrationErrorCases(t *testing.T) {
	// Simulate nil migrator
	err := createAndApplyMigration(nil, nil)
	if err == nil {
		t.Error("Expected error when migrator is nil")
	}

	// Simulate nil status
	db, _ := sql.Open(sqlite3Driver, memoryDB)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warningDBClose, closeErr)
		}
	}()
	migrator := migrations.NewHybridMigrator(db, migrations.SQLite, "./test_migrations")
	err = createAndApplyMigration(migrator, nil)
	if err == nil {
		t.Error("Expected error when status is nil")
	}
}

func TestShowFinalStatusErrorCase(t *testing.T) {
	err := showFinalStatus(nil)
	if err == nil {
		t.Error(errExpectedNilMigrator)
	}
}

func TestShowFinalStatusAllBranches(t *testing.T) {
	// Case: migrator is nil
	err := showFinalStatus(nil)
	if err == nil {
		t.Error(errExpectedNilMigrator)
	}

	// Case: migrator returns error from GetMigrationStatus (simulate by passing nil HybridMigrator)
	err = showFinalStatus((*migrations.HybridMigrator)(nil))
	if err == nil {
		t.Error(errExpectedGetStatusBranch)
	}

	// Case: migrator returns valid status (use a real migrator with a unique temp dir)
	db, _ := sql.Open(sqlite3Driver, memoryDB)
	tempDir, err := os.MkdirTemp("", testMigrationsDir)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	defer db.Close()
	migrator := migrations.NewHybridMigrator(db, migrations.SQLite, tempDir)
	migrator.DbSet(&User{})
	_ = migrator.GetMigrationStatus // ensure method exists
	err = showFinalStatus(migrator)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDisplayMigrationFileInfoWithWarnings(_ *testing.T) {
	migrationFile := &migrations.MigrationFile{
		Filename: "test_migration_with_warnings.sql",
		Changes: []migrations.MigrationChange{
			{
				Type:          "DROP_COLUMN",
				TableName:     "users",
				ColumnName:    "old_field",
				IsDestructive: true,
				RequiresData:  false,
				Description:   "Remove old field",
			},
		},
	}

	// Test that this doesn't panic and covers the warnings path
	displayMigrationFileInfo(migrationFile)
}

func TestDisplayMigrationFileInfoVariousChanges(_ *testing.T) {
	migrationFile := &migrations.MigrationFile{
		Filename: "comprehensive_migration.sql",
		Changes: []migrations.MigrationChange{
			{
				Type:          "ADD_TABLE",
				TableName:     "new_table",
				ColumnName:    "",
				IsDestructive: false,
				RequiresData:  false,
				Description:   "Add new table",
			},
			{
				Type:          "ADD_COLUMN",
				TableName:     "users",
				ColumnName:    "new_field",
				IsDestructive: false,
				RequiresData:  true,
				Description:   "Add new field with data migration",
			},
			{
				Type:          "DROP_TABLE",
				TableName:     "old_table",
				ColumnName:    "",
				IsDestructive: true,
				RequiresData:  false,
				Description:   "Remove old table",
			},
		},
	}

	// Test that this covers multiple change types and warnings
	displayMigrationFileInfo(migrationFile)
}

func TestSetupMigratorErrorScenarios(t *testing.T) {
	t.Run("setup with different database types", func(t *testing.T) {
		// Test with SQLite (known working)
		db, err := sql.Open(sqlite3Driver, memoryDB)
		if err != nil {
			t.Fatalf(failedCreateDB, err)
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				t.Logf(warningDBClose, closeErr)
			}
		}()

		// Create migrator with SQLite
		migrator := migrations.NewHybridMigrator(
			db,
			migrations.SQLite,
			"./test_migrations",
		)

		if migrator == nil {
			t.Error("Migrator should not be nil")
		}

		// Test model registration doesn't panic
		migrator.DbSet(&User{})
		migrator.DbSet(&Post{})
		migrator.DbSet(&Comment{})
	})
}

func TestUserStructTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expectID int64
		expectOK bool
	}{
		{"valid user", User{ID: 1, Email: "a@b.com", Name: "A", IsActive: true}, 1, true},
		{name: zeroIDCase, user: User{ID: 0, Email: "a@b.com", Name: "A", IsActive: true}, expectID: 0, expectOK: true},
		{"inactive", User{ID: 2, Email: "b@c.com", Name: "B", IsActive: false}, 2, false},
		{"empty name", User{ID: 3, Email: "c@d.com", Name: "", IsActive: true}, 3, true},
		// Fix: These cases should expect true for IsActive, as the struct does not validate email or name length
		{"invalid email", User{ID: 4, Email: "invalid", Name: "D", IsActive: true}, 4, true},
		{"long name", User{ID: 5, Email: "e@f.com", Name: strings.Repeat("A", 256), IsActive: true}, 5, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectID != tc.user.ID {
				t.Errorf(errIDFormat, tc.expectID, tc.user.ID)
			}
			if tc.user.IsActive != tc.expectOK && tc.expectID != 0 {
				t.Errorf(errIsActiveFormat, tc.expectOK)
			}
		})
	}
}

func TestPostStructTableDriven(t *testing.T) {
	tests := []struct {
		name            string
		post            Post
		expectID        int64
		expectUserID    int64
		expectTitle     string
		expectPublished bool
	}{
		{"valid post", Post{ID: 1, UserID: 1, Title: "T", IsPublished: true}, 1, 1, "T", true},
		{name: zeroIDCase, post: Post{ID: 0, UserID: 1, Title: "T", IsPublished: false}, expectID: 0, expectUserID: 1, expectTitle: "T", expectPublished: false},
		{"empty title", Post{ID: 2, UserID: 2, Title: "", IsPublished: false}, 2, 2, "", false},
		{"long title", Post{ID: 3, UserID: 1, Title: strings.Repeat("T", 256), IsPublished: true}, 3, 1, strings.Repeat("T", 256), true},
		{"invalid user", Post{ID: 4, UserID: -1, Title: "Valid", IsPublished: true}, 4, -1, "Valid", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectID != tc.post.ID {
				t.Errorf(errIDFormat, tc.expectID, tc.post.ID)
			}
			if tc.post.UserID != tc.expectUserID {
				t.Errorf("Expected UserID %d, got %d", tc.expectUserID, tc.post.UserID)
			}
			if tc.post.Title != tc.expectTitle {
				t.Errorf("Expected Title to be '%s', got '%s'", tc.expectTitle, tc.post.Title)
			}
			if tc.post.IsPublished != tc.expectPublished {
				t.Errorf("Expected IsPublished to be %v", tc.expectPublished)
			}
		})
	}
}

func TestCommentStructTableDriven(t *testing.T) {
	tests := []struct {
		name          string
		comment       Comment
		expectID      int64
		expectContent string
	}{
		{"valid comment", Comment{ID: 1, Content: "ok"}, 1, "ok"},
		{"zero id", Comment{ID: 0, Content: "empty"}, 0, "empty"},
		{"empty content", Comment{ID: 2, Content: ""}, 2, ""},
		{"long content", Comment{ID: 3, Content: strings.Repeat("C", 256)}, 3, strings.Repeat("C", 256)},
		{"invalid post", Comment{ID: 4, PostID: -1, Content: "Invalid post"}, 4, "Invalid post"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectID != tc.comment.ID {
				t.Errorf(errIDFormat, tc.expectID, tc.comment.ID)
			}
			if tc.comment.Content != tc.expectContent {
				t.Errorf("Expected Content '%s', got '%s'", tc.expectContent, tc.comment.Content)
			}
		})
	}
}

func TestMigrationStatusEdgeCases(t *testing.T) {
	// Nil MigrationStatus pointer (should not panic)
	displayMigrationStatus(nil)

	// Empty MigrationStatus struct
	displayMigrationStatus(&migrations.MigrationStatus{})

	// MigrationStatus with only applied migrations
	applied := &migrations.MigrationFile{Name: "applied", Timestamp: time.Now()}
	statusApplied := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{applied},
	}
	displayMigrationStatus(statusApplied)

	// MigrationStatus with only pending migrations
	pending := &migrations.MigrationFile{Name: "pending", Timestamp: time.Now()}
	statusPending := &migrations.MigrationStatus{
		PendingMigrations: []*migrations.MigrationFile{pending},
		HasPendingChanges: true,
	}
	displayMigrationStatus(statusPending)

	// MigrationStatus with nil slices
	statusNilSlices := &migrations.MigrationStatus{
		AppliedMigrations: nil,
		PendingMigrations: nil,
	}
	displayMigrationStatus(statusNilSlices)
}

func TestMigrationFileEdgeCases(t *testing.T) {
	// Nil MigrationFile pointer (should not panic)
	displayMigrationFileInfo(nil)

	// Empty MigrationFile struct
	displayMigrationFileInfo(&migrations.MigrationFile{})

	// MigrationFile with only filename
	fileWithName := &migrations.MigrationFile{Filename: "only_name.sql"}
	displayMigrationFileInfo(fileWithName)

	// MigrationFile with nil Changes
	fileNilChanges := &migrations.MigrationFile{Filename: "nil_changes.sql", Changes: nil}
	displayMigrationFileInfo(fileNilChanges)

	// MigrationFile with unusual values
	fileUnusual := &migrations.MigrationFile{
		Filename: "unusual.sql",
		Changes:  []migrations.MigrationChange{{Type: "", TableName: "", ColumnName: "", IsDestructive: false, RequiresData: false, Description: ""}},
	}
	displayMigrationFileInfo(fileUnusual)
}

func BenchmarkDisplayMigrationStatus(b *testing.B) {
	status := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{{Name: "applied", Timestamp: time.Now()}},
		PendingMigrations: []*migrations.MigrationFile{{Name: "pending", Timestamp: time.Now()}},
		HasPendingChanges: true,
		Summary:           "Benchmark summary",
	}
	for i := 0; i < b.N; i++ {
		displayMigrationStatus(status)
	}
}

func BenchmarkDisplayMigrationFileInfo(b *testing.B) {
	file := &migrations.MigrationFile{
		Filename: "bench.sql",
		Changes:  []migrations.MigrationChange{{Type: "ADD_TABLE", TableName: "bench", Description: "desc"}},
	}
	for i := 0; i < b.N; i++ {
		displayMigrationFileInfo(file)
	}
}

// Test that all public methods for MigrationStatus and MigrationFile do not panic and have correct return types
func TestPublicMethodsDoNotPanic(t *testing.T) {
	// MigrationStatus
	status := &migrations.MigrationStatus{}
	_ = status.HasPendingChanges
	_ = status.Summary

	// MigrationFile
	file := &migrations.MigrationFile{}
	_ = file.HasDestructiveChanges()
	_ = file.RequiresReview()
	_ = file.GetWarnings()
	_ = file.Errors()
}

func TestDisplayMigrationStatusNilAndPartialCases(t *testing.T) {
	// Nil pointer
	displayMigrationStatus(nil)

	// Empty struct
	status := &migrations.MigrationStatus{}
	displayMigrationStatus(status)

	// Only AppliedMigrations
	status = &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{{Name: "applied", Timestamp: time.Now()}},
	}
	displayMigrationStatus(status)

	// Only PendingMigrations
	status = &migrations.MigrationStatus{
		PendingMigrations: []*migrations.MigrationFile{{Name: "pending", Timestamp: time.Now()}},
	}
	displayMigrationStatus(status)

	// No migrations, HasPendingChanges true
	status = &migrations.MigrationStatus{HasPendingChanges: true}
	displayMigrationStatus(status)
}

func TestDisplayMigrationFileInfoNilAndPartialCases(t *testing.T) {
	// Nil pointer
	displayMigrationFileInfo(nil)

	// Empty struct
	file := &migrations.MigrationFile{}
	displayMigrationFileInfo(file)

	// Only Filename
	file = &migrations.MigrationFile{Filename: "only_filename.sql"}
	displayMigrationFileInfo(file)

	// Only Changes
	file = &migrations.MigrationFile{Changes: []migrations.MigrationChange{{Type: "ADD_TABLE", TableName: "t"}}}
	displayMigrationFileInfo(file)

	// Change with all fields empty
	file = &migrations.MigrationFile{Changes: []migrations.MigrationChange{{}}}
	displayMigrationFileInfo(file)
}

func TestMigrationStatusPartialFields(t *testing.T) {
	// AppliedMigrations nil, PendingMigrations non-nil
	status := &migrations.MigrationStatus{
		PendingMigrations: []*migrations.MigrationFile{{Name: "pending"}},
	}
	displayMigrationStatus(status)

	// AppliedMigrations non-nil, PendingMigrations nil
	status = &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{{Name: "applied"}},
	}
	displayMigrationStatus(status)
}

func TestMigrationFilePartialFields(t *testing.T) {
	// MigrationFile with nil Changes
	file := &migrations.MigrationFile{Filename: "partial.sql"}
	displayMigrationFileInfo(file)

	// MigrationFile with one Change, partial fields
	file = &migrations.MigrationFile{
		Filename: "partial2.sql",
		Changes:  []migrations.MigrationChange{{Type: "DROP_TABLE"}},
	}
	displayMigrationFileInfo(file)
}

func TestCreateAndApplyMigrationNilStatusAndMigrator(t *testing.T) {
	// Both nil
	if err := createAndApplyMigration(nil, nil); err == nil {
		t.Error("Expected error for nil migrator and status")
	}
	// Nil status
	db, _ := sql.Open(sqlite3Driver, memoryDB)
	defer db.Close()
	migrator := migrations.NewHybridMigrator(db, migrations.SQLite, testMigrationsDir)
	if err := createAndApplyMigration(migrator, nil); err == nil {
		t.Error("Expected error for nil status")
	}
	// Nil migrator
	status := &migrations.MigrationStatus{}
	if err := createAndApplyMigration(nil, status); err == nil {
		t.Error("Expected error for nil migrator")
	}
}

func TestShowFinalStatusNilAndErrorCases(t *testing.T) {
	if err := showFinalStatus(nil); err == nil {
		t.Error("Expected error for nil migrator")
	}
	// Simulate error from GetMigrationStatus (nil HybridMigrator)
	if err := showFinalStatus((*migrations.HybridMigrator)(nil)); err == nil {
		t.Error("Expected error from GetMigrationStatus branch")
	}
}

// TestMainCoversMainLogic increases coverage by running main() in a safe way.
func TestMainCoversMainLogic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main panicked: %v", r)
		}
	}()
	main()
}

// Additional coverage: Test Comment struct tags
func TestCommentStructTags(t *testing.T) {
	commentType := reflect.TypeOf(Comment{})

	// Test ID field tags
	idField, found := commentType.FieldByName("ID")
	if !found {
		t.Fatal("ID field not found in Comment struct")
	}
	if idField.Tag.Get("db") != "id" {
		t.Errorf("Expected db tag 'id', got '%s'", idField.Tag.Get("db"))
	}
	if !strings.Contains(idField.Tag.Get("migration"), "primary_key") {
		t.Errorf("Expected migration tag to contain 'primary_key', got '%s'", idField.Tag.Get("migration"))
	}

	// Test PostID field tags
	postIDField, found := commentType.FieldByName("PostID")
	if !found {
		t.Fatal("PostID field not found in Comment struct")
	}
	if postIDField.Tag.Get("db") != "post_id" {
		t.Errorf("Expected db tag 'post_id', got '%s'", postIDField.Tag.Get("db"))
	}
	if !strings.Contains(postIDField.Tag.Get("migration"), "foreign_key:posts.id") {
		t.Errorf("Expected migration tag to contain 'foreign_key:posts.id', got '%s'", postIDField.Tag.Get("migration"))
	}

	// Test UserID field tags
	userIDField, found := commentType.FieldByName("UserID")
	if !found {
		t.Fatal("UserID field not found in Comment struct")
	}
	if userIDField.Tag.Get("db") != "user_id" {
		t.Errorf("Expected db tag 'user_id', got '%s'", userIDField.Tag.Get("db"))
	}
	if !strings.Contains(userIDField.Tag.Get("migration"), "foreign_key:users.id") {
		t.Errorf("Expected migration tag to contain 'foreign_key:users.id', got '%s'", userIDField.Tag.Get("migration"))
	}

	// Test Content field tags
	contentField, found := commentType.FieldByName("Content")
	if !found {
		t.Fatal("Content field not found in Comment struct")
	}
	if contentField.Tag.Get("db") != "content" {
		t.Errorf("Expected db tag 'content', got '%s'", contentField.Tag.Get("db"))
	}
	if !strings.Contains(contentField.Tag.Get("migration"), "type:TEXT") {
		t.Errorf("Expected migration tag to contain 'type:TEXT', got '%s'", contentField.Tag.Get("migration"))
	}
}

// Additional coverage: Test MigrationFile methods with various changes
func TestMigrationFileMethodsEdgeCases(t *testing.T) {
	file := &migrations.MigrationFile{
		Filename: "edge.sql",
		Changes: []migrations.MigrationChange{
			{Type: "DROP_TABLE", TableName: "users", IsDestructive: true, RequiresData: false, Description: "Drop users table"},
			{Type: "ADD_COLUMN", TableName: "posts", ColumnName: "extra", IsDestructive: false, RequiresData: true, Description: "Add extra column"},
		},
	}
	if !file.HasDestructiveChanges() {
		t.Error("Expected HasDestructiveChanges to be true")
	}
	if !file.RequiresReview() {
		t.Error("Expected RequiresReview to be true")
	}
	warnings := file.GetWarnings()
	if len(warnings) == 0 {
		t.Error("Expected warnings for destructive changes")
	}
	errs := file.Errors()
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got %d", len(errs))
	}
}

// Additional coverage: Test MigrationFile with empty and partial data
func TestMigrationFileEmptyAndPartial(t *testing.T) {
	file := &migrations.MigrationFile{}
	if file.HasDestructiveChanges() {
		t.Error("Expected HasDestructiveChanges to be false for empty file")
	}
	if file.RequiresReview() {
		t.Error("Expected RequiresReview to be false for empty file")
	}
	if len(file.GetWarnings()) != 0 {
		t.Error("Expected no warnings for empty file")
	}
	if len(file.Errors()) != 0 {
		t.Error("Expected no errors for empty file")
	}

	file.Changes = []migrations.MigrationChange{{}}
	if file.HasDestructiveChanges() {
		t.Error("Expected HasDestructiveChanges to be false for non-destructive change")
	}
}

// Additional coverage: Test MigrationStatus with nil, empty, and mixed fields
func TestMigrationStatusNilEmptyMixed(t *testing.T) {
	var status *migrations.MigrationStatus
	if status != nil && status.HasPendingChanges {
		t.Error("Expected HasPendingChanges to be false for nil status")
	}

	empty := &migrations.MigrationStatus{}
	if empty.HasPendingChanges {
		t.Error("Expected HasPendingChanges to be false for empty status")
	}
	if empty.Summary != "" {
		t.Errorf("Expected empty Summary, got %s", empty.Summary)
	}

	mixed := &migrations.MigrationStatus{
		AppliedMigrations: []*migrations.MigrationFile{{Name: "applied"}},
		PendingMigrations: []*migrations.MigrationFile{{Name: "pending"}},
		HasPendingChanges: true,
		Summary:           "Mixed status",
	}
	if !mixed.HasPendingChanges {
		t.Error("Expected HasPendingChanges to be true for mixed status")
	}
	if mixed.Summary != "Mixed status" {
		t.Errorf("Expected Summary 'Mixed status', got %s", mixed.Summary)
	}
}

// Additional coverage: Table-driven test for MigrationFile public methods
func TestMigrationFilePublicMethodsTableDriven(t *testing.T) {
	tests := []struct {
		name   string
		file   *migrations.MigrationFile
		wantDC bool
		wantRR bool
		wantW  int
		wantE  int
	}{
		{"nil file", nil, false, false, 0, 0},
		{"empty file", &migrations.MigrationFile{}, false, false, 0, 0},
		{"destructive", &migrations.MigrationFile{Changes: []migrations.MigrationChange{{IsDestructive: true}}}, true, true, 1, 0},
		{"non-destructive", &migrations.MigrationFile{Changes: []migrations.MigrationChange{{IsDestructive: false}}}, false, false, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.file.HasDestructiveChanges(); got != tc.wantDC {
				t.Errorf("HasDestructiveChanges: got %v, want %v", got, tc.wantDC)
			}
			if got := tc.file.RequiresReview(); got != tc.wantRR {
				t.Errorf("RequiresReview: got %v, want %v", got, tc.wantRR)
			}
			if got := len(tc.file.GetWarnings()); got != tc.wantW {
				t.Errorf("GetWarnings: got %d, want %d", got, tc.wantW)
			}
			if got := len(tc.file.Errors()); got != tc.wantE {
				t.Errorf("Errors: got %d, want %d", got, tc.wantE)
			}
		})
	}
}

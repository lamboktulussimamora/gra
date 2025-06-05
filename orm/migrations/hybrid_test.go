package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
)

// Test models
type TestUser struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
}

type TestPost struct {
	ID       int64  `db:"id" migration:"primary_key,auto_increment"`
	UserID   int64  `db:"user_id" migration:"not_null,foreign_key:users.id"`
	Title    string `db:"title" migration:"not_null,max_length:255"`
	Content  string `db:"content" migration:"type:TEXT"`
	IsPublic bool   `db:"is_public" migration:"not_null,default:false"`
}

// TestUserWithBio is a modified version of TestUser with an additional Bio field
// Used for testing column addition detection
type TestUserWithBio struct {
	ID        int64     `db:"id" migration:"primary_key,auto_increment"`
	Email     string    `db:"email" migration:"unique,not_null,max_length:255"`
	Name      string    `db:"name" migration:"not_null,max_length:100"`
	IsActive  bool      `db:"is_active" migration:"not_null,default:true"`
	CreatedAt time.Time `db:"created_at" migration:"not_null,default:CURRENT_TIMESTAMP"`
	Bio       string    `db:"bio" migration:"type:TEXT"` // New field
}

// TableName returns the same table name as TestUser to enable column change detection
func (TestUserWithBio) TableName() string {
	return testUsersTable
}

// UserWithoutEmail is used for testing destructive changes (column removal)
type UserWithoutEmail struct {
	ID   int64  `db:"id" migration:"primary_key,auto_increment"`
	Name string `db:"name" migration:"not_null,max_length:100"`
}

// TableName returns the same table name as TestUser to enable destructive change detection
func (UserWithoutEmail) TableName() string {
	return testUsersTable
}

const (
	testUsersTable           = "testusers"
	warnFailedToCloseDB      = "Warning: Failed to close database: %v"
	errFailedToDetectChanges = "Failed to detect changes: %v"
	testUsersTableQuery      = "SELECT name FROM sqlite_master WHERE type='table' AND name='testusers'"
)

// Test helpers
func setupTestDB(t *testing.T) (*sql.DB, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db, tmpDir
}

func setupTestMigrator(t *testing.T) (*HybridMigrator, *sql.DB, string) {
	db, tmpDir := setupTestDB(t)
	migrationsDir := filepath.Join(tmpDir, "migrations")

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("Failed to create migrations directory: %v", err)
	}

	migrator := NewHybridMigrator(db, SQLite, migrationsDir)
	return migrator, db, tmpDir
}

// Test ModelRegistry
func TestModelRegistry(t *testing.T) {
	registry := NewModelRegistry(SQLite)

	// Test model registration
	registry.RegisterModel(&TestUser{})
	registry.RegisterModel(&TestPost{})

	models := registry.GetModels()

	// Check user model
	userSnapshot, exists := models["testusers"]
	if !exists {
		t.Fatal("TestUser snapshot not found")
	}

	if userSnapshot.TableName != "testusers" {
		t.Errorf("Expected table name 'testusers', got '%s'", userSnapshot.TableName)
	}

	// Check columns
	if len(userSnapshot.Columns) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(userSnapshot.Columns))
	}

	// Check ID column
	idColumn, exists := userSnapshot.Columns["id"]
	if !exists {
		t.Fatal("ID column not found")
	}

	// Debug print
	t.Logf("ID column: %+v", idColumn)

	if !idColumn.IsPrimaryKey {
		t.Error("ID column should be primary key")
	}

	if !idColumn.IsIdentity {
		t.Error("ID column should be identity")
	}

	// Check email column
	emailColumn, exists := userSnapshot.Columns["email"]
	if !exists {
		t.Fatal("Email column not found")
	}

	if emailColumn.IsNullable {
		t.Error("Email column should not be nullable")
	}

	if emailColumn.MaxLength == nil || *emailColumn.MaxLength != 255 {
		t.Error("Email column should have max length of 255")
	}

	// Check post model
	postSnapshot, exists := models["testposts"]
	if !exists {
		t.Fatal("TestPost snapshot not found")
	}

	if postSnapshot.TableName != "testposts" {
		t.Errorf("Expected table name 'testposts', got '%s'", postSnapshot.TableName)
	}
}

// Test Change Detection
func TestChangeDetection(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Register models
	migrator.DbSet(&TestUser{})
	migrator.DbSet(&TestPost{})

	// Detect changes (should detect new tables)
	plan, err := migrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf(errFailedToDetectChanges, err)
	}

	if len(plan.Changes) == 0 {
		t.Fatal("Expected changes to be detected")
	}

	// Check that table creation changes are detected
	createTableCount := 0
	for _, change := range plan.Changes {
		if change.Type == CreateTable {
			createTableCount++
		}
	}

	if createTableCount != 2 {
		t.Errorf("Expected 2 CreateTable changes, got %d", createTableCount)
	}

	// Validate migration plan
	err = migrator.changeDetector.ValidateMigrationPlan(plan)
	if err != nil {
		t.Errorf("Migration plan validation failed: %v", err)
	}
}

// Test SQL Generation
func TestSQLGeneration(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	migrator.DbSet(&TestUser{})

	plan, err := migrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf(errFailedToDetectChanges, err)
	}

	migrationSQL, err := migrator.sqlGenerator.GenerateMigrationSQL(plan)
	if err != nil {
		t.Fatalf("Failed to generate SQL: %v", err)
	}

	// Check that SQL is generated
	if migrationSQL.UpScript == "" {
		t.Error("Up script should not be empty")
	}

	if migrationSQL.DownScript == "" {
		t.Error("Down script should not be empty")
	}

	// Check that CREATE TABLE is in up script
	if !contains(migrationSQL.UpScript, "CREATE TABLE") {
		t.Error("Up script should contain CREATE TABLE")
	}

	// Check that DROP TABLE is in down script
	if !contains(migrationSQL.DownScript, "DROP TABLE") {
		t.Error("Down script should contain DROP TABLE")
	}
}

// Test Migration Creation and Application
func registerModelAndLog(t *testing.T, migrator *HybridMigrator) {
	migrator.DbSet(&TestUser{})
	models := migrator.registry.GetModels()
	t.Logf("Registered models:")
	for name, snapshot := range models {
		t.Logf("  Model: %s, Table: %s", name, snapshot.TableName)
	}
}

func createAndValidateMigration(t *testing.T, migrator *HybridMigrator) *MigrationFile {
	migrationFile, err := migrator.AddMigration("create_users", Interactive)
	if err != nil {
		t.Fatalf("Failed to create migration: %v", err)
	}
	if migrationFile.Name != "create_users" {
		t.Errorf("Expected migration name 'create_users', got '%s'", migrationFile.Name)
	}
	if _, err := os.Stat(migrationFile.FilePath); os.IsNotExist(err) {
		t.Error("Migration file was not created")
	}
	return migrationFile
}

func logMigrationFileContent(t *testing.T, migrationFile *MigrationFile) {
	if content, err := os.ReadFile(migrationFile.FilePath); err == nil {
		t.Logf("Migration file content:\n%s", string(content))
	} else {
		t.Logf("Failed to read migration file: %v", err)
	}
}

func applyMigrationAndCheckTable(t *testing.T, migrator *HybridMigrator, db *sql.DB) {
	err := migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}
	var tableName string
	err = db.QueryRow(testUsersTableQuery).Scan(&tableName)
	if err != nil {
		t.Errorf("Table 'testusers' was not created: %v", err)
	}
}

func validateMigrationStatus(t *testing.T, migrator *HybridMigrator) {
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}
	if len(status.AppliedMigrations) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(status.AppliedMigrations))
	}
	if status.HasPendingChanges {
		t.Error("Should not have pending changes after applying migration")
	}
}

func TestMigrationCreationAndApplication(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	registerModelAndLog(t, migrator)
	migrationFile := createAndValidateMigration(t, migrator)
	logMigrationFileContent(t, migrationFile)
	applyMigrationAndCheckTable(t, migrator, db)
	validateMigrationStatus(t, migrator)
}

// Test Migration Rollback
func TestMigrationRollback(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Register model and create migration
	migrator.DbSet(&TestUser{})

	_, err := migrator.AddMigration("create_users", Interactive)
	if err != nil {
		t.Fatalf("Failed to create migration: %v", err)
	}

	// Apply migration
	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow(testUsersTableQuery).Scan(&tableName)
	if err != nil {
		t.Fatalf("Table should exist before rollback: %v", err)
	}

	// Rollback migration
	err = migrator.RevertMigration()
	if err != nil {
		t.Fatalf("Failed to revert migration: %v", err)
	}

	// Verify table no longer exists
	err = db.QueryRow(testUsersTableQuery).Scan(&tableName)
	if err == nil {
		t.Error("Table should not exist after rollback")
	}

	// Check migration status
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if len(status.AppliedMigrations) != 0 {
		t.Errorf("Expected 0 applied migrations after rollback, got %d", len(status.AppliedMigrations))
	}
}

// Test Column Changes
func TestColumnChanges(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Initial model
	migrator.DbSet(&TestUser{})

	// Create and apply initial migration
	_, err := migrator.AddMigration("initial", Interactive)
	if err != nil {
		t.Fatalf("Failed to create initial migration: %v", err)
	}

	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply initial migration: %v", err)
	}

	// Create new registry with modified model
	newMigrator := NewHybridMigrator(db, SQLite, migrator.migrationsDir)
	newMigrator.DbSet(&TestUserWithBio{})

	// Detect changes
	plan, err := newMigrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf(errFailedToDetectChanges, err)
	}

	// Debug: Print what was detected
	t.Logf("Total changes detected: %d", len(plan.Changes))
	for i, change := range plan.Changes {
		t.Logf("Change %d: Type=%s, Table=%s, Column=%s", i, change.Type, change.TableName, change.ColumnName)
	}

	// Should detect one AddColumn change
	addColumnCount := 0
	for _, change := range plan.Changes {
		if change.Type == AddColumn && change.ColumnName == "bio" {
			addColumnCount++
		}
	}

	if addColumnCount != 1 {
		t.Errorf("Expected 1 AddColumn change for 'bio', got %d", addColumnCount)
	}
}

// Test Multiple Migration Modes
func TestMigrationModes(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	migrator.DbSet(&TestUser{})

	// Test GenerateOnly mode
	migration, err := migrator.AddMigration("test_generate", GenerateOnly)
	if err != nil {
		t.Fatalf("Failed to create migration in GenerateOnly mode: %v", err)
	}

	if migration.Mode != GenerateOnly {
		t.Errorf("Expected GenerateOnly mode, got %s", migration.Mode)
	}

	// Check that migration file exists but table doesn't
	if _, err := os.Stat(migration.FilePath); os.IsNotExist(err) {
		t.Error("Migration file should exist in GenerateOnly mode")
	}

	var tableName string
	err = db.QueryRow(testUsersTableQuery).Scan(&tableName)
	if err == nil {
		t.Error("Table should not exist in GenerateOnly mode")
	}

	// Apply the generated migration
	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply generated migration: %v", err)
	}

	// Now table should exist
	err = db.QueryRow(testUsersTableQuery).Scan(&tableName)
	if err != nil {
		t.Errorf("Table should exist after applying migration: %v", err)
	}
}

// Test Database Inspector
func TestDatabaseInspector(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Create a table manually
	_, err := db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create index
	_, err = db.Exec("CREATE INDEX idx_test_email ON test_table(email)")
	if err != nil {
		t.Fatalf("Failed to create test index: %v", err)
	}

	// Inspect database
	schema, err := migrator.inspector.GetCurrentSchema()
	if err != nil {
		t.Fatalf("Failed to inspect database: %v", err)
	}

	// Check that table is detected
	table, exists := schema["test_table"]
	if !exists {
		t.Fatal("test_table should be detected by inspector")
	}

	// Check columns
	if len(table.Columns) < 4 {
		t.Errorf("Expected at least 4 columns, got %d", len(table.Columns))
	}

	// Check primary key
	if len(table.PrimaryKeys) != 1 || table.PrimaryKeys[0] != "id" {
		t.Errorf("Expected primary key 'id', got %v", table.PrimaryKeys)
	}

	// Check indexes (note: primary key index might be included)
	if len(table.Indexes) < 1 {
		t.Errorf("Expected at least 1 index, got %d", len(table.Indexes))
	}
}

// Test Error Handling
func TestErrorHandling(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Test adding migration without models
	_, err := migrator.AddMigration("empty_migration", Interactive)
	if err == nil {
		t.Error("Should fail when no changes are detected")
	}

	// Test invalid migration mode combination
	migrator.DbSet(&TestUser{})

	// First create the full user table
	_, err = migrator.AddMigration("create_full_user", Interactive)
	if err != nil {
		t.Fatalf("Failed to create initial migration: %v", err)
	}

	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply initial migration: %v", err)
	}

	// Now register reduced model (this would drop email column - destructive)
	newMigrator := NewHybridMigrator(db, SQLite, migrator.migrationsDir)
	newMigrator.DbSet(&UserWithoutEmail{})

	plan, err := newMigrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf(errFailedToDetectChanges, err)
	}

	// Should have destructive changes
	if !plan.HasDestructive {
		t.Error("Plan should have destructive changes")
	}

	// Automatic mode should fail
	err = newMigrator.ApplyMigrations(Automatic)
	if err == nil {
		t.Error("Automatic mode should fail with destructive changes")
	}
}

// Test helpers
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration test
func TestFullWorkflow(t *testing.T) {
	migrator, db, _ := setupTestMigrator(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(warnFailedToCloseDB, closeErr)
		}
	}()

	// Step 1: Create initial schema
	migrator.DbSet(&TestUser{})

	migration1, err := migrator.AddMigration("create_users", Interactive)
	if err != nil {
		t.Fatalf("Failed to create first migration: %v", err)
	}

	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply first migration: %v", err)
	}

	// Step 2: Add posts table
	migrator.DbSet(&TestPost{})

	migration2, err := migrator.AddMigration("create_posts", Interactive)
	if err != nil {
		t.Fatalf("Failed to create second migration: %v", err)
	}

	err = migrator.ApplyMigrations(Automatic)
	if err != nil {
		t.Fatalf("Failed to apply second migration: %v", err)
	}

	// Step 3: Check final status
	status, err := migrator.GetMigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get final status: %v", err)
	}

	if len(status.AppliedMigrations) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(status.AppliedMigrations))
	}

	if status.HasPendingChanges {
		t.Error("Should not have pending changes")
	}

	// Step 4: Rollback both migrations
	err = migrator.RevertMigration()
	if err != nil {
		t.Fatalf("Failed to revert posts migration: %v", err)
	}

	err = migrator.RevertMigration()
	if err != nil {
		t.Fatalf("Failed to revert users migration: %v", err)
	}

	// Step 5: Verify clean state
	finalStatus, err := migrator.GetMigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get final status after rollback: %v", err)
	}

	if len(finalStatus.AppliedMigrations) != 0 {
		t.Errorf("Expected 0 applied migrations after full rollback, got %d", len(finalStatus.AppliedMigrations))
	}

	// Verify migration files still exist
	if _, err := os.Stat(migration1.FilePath); os.IsNotExist(err) {
		t.Error("Migration files should still exist after rollback")
	}

	if _, err := os.Stat(migration2.FilePath); os.IsNotExist(err) {
		t.Error("Migration files should still exist after rollback")
	}
}

package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
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

	// #nosec G301 -- Directory must be user-accessible for migration files
	if err := os.MkdirAll(migrationsDir, 0750); err != nil {
		t.Fatalf("Failed to create migrations dir: %v", err)
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

	migrationQL, err := migrator.sqlGenerator.GenerateMigrationSQL(plan)
	if err != nil {
		t.Fatalf("failed to generate migration SQL: %v", err)
	}

	// Check that SQL is generated
	if migrationQL.UpScript == "" {
		t.Error("Up script should not be empty")
	}
	if migrationQL.DownScript == "" {
		t.Error("Down script should not be empty")
	}
	if !contains(migrationQL.UpScript, "CREATE TABLE") {
		t.Error("Up script should contain CREATE TABLE")
	}
	if !contains(migrationQL.DownScript, "DROP TABLE") {
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

// Test data-losing change detection functions
func TestChangeDetectorDataLosingFunctions(t *testing.T) {
	detector := setupChangeDetector(t)

	t.Run("isDataLosingAlterColumn", func(t *testing.T) {
		testIsDataLosingAlterColumn(t, detector)
	})

	t.Run("extractColumnInfoFromChange", func(t *testing.T) {
		testExtractColumnInfoFromChange(t, detector)
	})

	t.Run("hasDataLosingColumnChanges", func(t *testing.T) {
		testHasDataLosingColumnChanges(t, detector)
	})

	t.Run("isNullabilityChangeDataLosing", func(t *testing.T) {
		testIsNullabilityChangeDataLosing(t, detector)
	})

	t.Run("isLengthReductionDataLosing", func(t *testing.T) {
		testIsLengthReductionDataLosing(t, detector)
	})

	t.Run("isIncompatibleTypeChange", func(t *testing.T) {
		testIsIncompatibleTypeChange(t, detector)
	})

	t.Run("getIncompatibleTypeMap", func(t *testing.T) {
		testGetIncompatibleTypeMap(t, detector)
	})

	t.Run("checkTypeIncompatibility", func(t *testing.T) {
		testCheckTypeIncompatibility(t, detector)
	})
}

func testIsDataLosingAlterColumn(t *testing.T, detector *ChangeDetector) {
	// Test non-alter-column change
	change := MigrationChange{Type: CreateTable}
	if detector.isDataLosingAlterColumn(change) {
		t.Error("Expected false for non-alter-column change")
	}

	// Test alter column without proper values
	change = MigrationChange{Type: AlterColumn}
	if detector.isDataLosingAlterColumn(change) {
		t.Error("Expected false for alter column without proper values")
	}

	// Test alter column with proper values - data losing
	maxLength100 := 100
	maxLength50 := 50
	oldColumn := &DatabaseColumnInfo{
		IsNullable: true,
		MaxLength:  &maxLength100,
		DataType:   "VARCHAR",
	}
	newColumn := &ColumnInfo{
		IsNullable: false,
		MaxLength:  &maxLength50,
		DataType:   "VARCHAR",
	}
	change = MigrationChange{
		Type:     AlterColumn,
		OldValue: oldColumn,
		NewValue: newColumn,
	}
	if !detector.isDataLosingAlterColumn(change) {
		t.Error("Expected true for data-losing alter column")
	}
}

func testExtractColumnInfoFromChange(t *testing.T, detector *ChangeDetector) {
	// Test with invalid types
	change := MigrationChange{
		OldValue: "invalid",
		NewValue: "invalid",
	}
	_, _, ok := detector.extractColumnInfoFromChange(change)
	if ok {
		t.Error("Expected false for invalid types")
	}

	// Test with valid types
	oldColumn := &DatabaseColumnInfo{DataType: "VARCHAR"}
	newColumn := &ColumnInfo{DataType: "TEXT"}
	change = MigrationChange{
		OldValue: oldColumn,
		NewValue: newColumn,
	}
	old, new, ok := detector.extractColumnInfoFromChange(change)
	if !ok {
		t.Error("Expected true for valid types")
	}
	if old != oldColumn {
		t.Error("Expected old column to match")
	}
	if new != newColumn {
		t.Error("Expected new column to match")
	}
}

func testHasDataLosingColumnChanges(t *testing.T, detector *ChangeDetector) {
	// Test non-data-losing change
	oldColumn := &DatabaseColumnInfo{
		IsNullable: false,
		DataType:   "VARCHAR",
	}
	newColumn := &ColumnInfo{
		IsNullable: false,
		DataType:   "VARCHAR",
	}
	if detector.hasDataLosingColumnChanges(oldColumn, newColumn) {
		t.Error("Expected false for non-data-losing change")
	}

	// Test data-losing nullability change
	oldColumn.IsNullable = true
	newColumn.IsNullable = false
	if !detector.hasDataLosingColumnChanges(oldColumn, newColumn) {
		t.Error("Expected true for nullability data-losing change")
	}
}

func testIsNullabilityChangeDataLosing(t *testing.T, detector *ChangeDetector) {
	oldColumn := &DatabaseColumnInfo{IsNullable: true}
	newColumn := &ColumnInfo{IsNullable: false}
	if !detector.isNullabilityChangeDataLosing(oldColumn, newColumn) {
		t.Error("Expected true for nullable to non-nullable change")
	}

	oldColumn.IsNullable = false
	newColumn.IsNullable = true
	if detector.isNullabilityChangeDataLosing(oldColumn, newColumn) {
		t.Error("Expected false for non-nullable to nullable change")
	}
}

func testIsLengthReductionDataLosing(t *testing.T, detector *ChangeDetector) {
	maxLength100 := 100
	maxLength50 := 50
	maxLength200 := 200

	// Test length reduction
	oldColumn := &DatabaseColumnInfo{MaxLength: &maxLength100}
	newColumn := &ColumnInfo{MaxLength: &maxLength50}
	if !detector.isLengthReductionDataLosing(oldColumn, newColumn) {
		t.Error("Expected true for length reduction")
	}

	// Test length increase
	newColumn.MaxLength = &maxLength200
	if detector.isLengthReductionDataLosing(oldColumn, newColumn) {
		t.Error("Expected false for length increase")
	}

	// Test with nil lengths
	oldColumn.MaxLength = nil
	if detector.isLengthReductionDataLosing(oldColumn, newColumn) {
		t.Error("Expected false for nil old length")
	}
}

func testIsIncompatibleTypeChange(t *testing.T, detector *ChangeDetector) {
	// Test compatible change
	if detector.isIncompatibleTypeChange("VARCHAR(100)", "VARCHAR(200)") {
		t.Error("Expected false for compatible VARCHAR change")
	}

	// Test incompatible change
	if !detector.isIncompatibleTypeChange("TEXT", "INTEGER") {
		t.Error("Expected true for TEXT to INTEGER change")
	}

	if !detector.isIncompatibleTypeChange("BOOLEAN", "VARCHAR") {
		t.Error("Expected true for BOOLEAN to VARCHAR change")
	}

	// Test case insensitive
	if !detector.isIncompatibleTypeChange("text", "integer") {
		t.Error("Expected true for case-insensitive incompatible change")
	}
}

func testGetIncompatibleTypeMap(t *testing.T, detector *ChangeDetector) {
	typeMap := detector.getIncompatibleTypeMap()
	if len(typeMap) == 0 {
		t.Error("Expected non-empty type map")
	}

	// Test specific mappings
	textIncompatible, exists := typeMap["TEXT"]
	if !exists {
		t.Error("Expected TEXT to have incompatible types")
	}

	containsInteger := false
	for _, incompatible := range textIncompatible {
		if incompatible == "INTEGER" {
			containsInteger = true
			break
		}
	}
	if !containsInteger {
		t.Error("Expected INTEGER to be incompatible with TEXT")
	}
}

func testCheckTypeIncompatibility(t *testing.T, detector *ChangeDetector) {
	incompatibleMap := map[string][]string{
		"TEXT": {"INTEGER", "BOOLEAN"},
	}

	// Test incompatible
	if !detector.checkTypeIncompatibility("TEXT", "INTEGER", incompatibleMap) {
		t.Error("Expected true for incompatible types")
	}

	// Test compatible
	if detector.checkTypeIncompatibility("TEXT", "VARCHAR", incompatibleMap) {
		t.Error("Expected false for compatible types")
	}

	// Test non-existent source type
	if detector.checkTypeIncompatibility("UNKNOWN", "INTEGER", incompatibleMap) {
		t.Error("Expected false for unknown source type")
	}
}

// Test circular dependency detection
func TestChangeDetectorCircularDependencies(t *testing.T) {
	detector := setupChangeDetector(t)

	t.Run("hasCycleDFS", func(t *testing.T) {
		// Create a simple dependency graph without cycles
		deps := map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {},
		}

		visited := make(map[string]bool)
		recStack := make(map[string]bool)

		if detector.hasCycleDFS("A", deps, visited, recStack) {
			t.Error("Expected false for acyclic graph")
		}

		// Create a graph with cycles
		deps = map[string][]string{
			"A": {"B"},
			"B": {"C"},
			"C": {"A"}, // Creates cycle
		}

		visited = make(map[string]bool)
		recStack = make(map[string]bool)

		if !detector.hasCycleDFS("A", deps, visited, recStack) {
			t.Error("Expected true for cyclic graph")
		}
	})
}

// Test foreign key reference detection
func TestChangeDetectorForeignKeyReferences(t *testing.T) {
	detector := setupChangeDetector(t)

	t.Run("isForeignKeyReferencingDroppedTable", func(t *testing.T) {
		droppedTables := map[string]bool{
			"users":      true,
			"categories": true,
		}

		// Test constraint referencing dropped table
		constraint := &ConstraintInfo{
			Type:            "FOREIGN KEY",
			ReferencedTable: "users",
		}

		if !detector.isForeignKeyReferencingDroppedTable(constraint, droppedTables) {
			t.Error("Expected true for foreign key referencing dropped table")
		}

		// Test constraint not referencing dropped table
		constraint.ReferencedTable = "products"
		if detector.isForeignKeyReferencingDroppedTable(constraint, droppedTables) {
			t.Error("Expected false for foreign key not referencing dropped table")
		}

		// Test non-foreign key constraint
		constraint.Type = "UNIQUE"
		constraint.ReferencedTable = "users"
		if detector.isForeignKeyReferencingDroppedTable(constraint, droppedTables) {
			t.Error("Expected false for non-foreign key constraint")
		}
	})
}

// ...existing test code...

// Import required types and constants for hybrid migration tests
// These are re-exported for test visibility if not already imported
// (If these are already imported, this is a no-op)

// HybridMigrator, MigrationFile, MigrationStatus, NewHybridMigrator, NewModelRegistry, MigrationMode, etc.
// are defined in the migrations package and available for use in this test file.
// The following import ensures all symbols are available for type checking and test execution.

// Helper functions for test support

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// setupChangeDetector creates a change detector for testing
func setupChangeDetector(t *testing.T) *ChangeDetector {
	// Create a temporary directory for test
	tempDir := t.TempDir()

	// Create a test database
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("Failed to close test database: %v", closeErr)
		}
	}()

	// Create model registry and change detector
	registry := NewModelRegistry(SQLite)
	inspector := NewDatabaseInspector(db, SQLite)
	detector := NewChangeDetector(registry, inspector)

	return detector
}

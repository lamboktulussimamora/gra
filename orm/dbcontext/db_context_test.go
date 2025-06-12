package dbcontext

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test constants to avoid goconst violations
const (
	testDBCloseError           = "Failed to close database: %v"
	testChangeTrackerNil       = "ChangeTracker should not be nil"
	testUserName               = "Test User"
	testInsertQuery            = "INSERT INTO testusers (name, email, is_active) VALUES (?, ?, ?)"
	testInsertDataError        = "Failed to insert test data: %v"
	testNameCondition          = "name = ?"
	testJohnDoe                = "John Doe"
	testJohnEmail              = "john@example.com"
	testJohnUpdated            = "John Updated"
	testAliceName              = "Alice"
	testAliceEmail             = "alice@example.com"
	testBobName                = "Bob"
	testBobEmail               = "bob@example.com"
	testNonExistent            = "NonExistent"
	testCountQuery             = "SELECT COUNT(*) FROM testusers WHERE name = ?"
	testCountDeletedQuery      = "SELECT COUNT(*) FROM testusers WHERE id = ?"
	testSQLite3Driver          = "sqlite3"
	testMemoryDB               = ":memory:"
	testIsActiveCondition      = "is_active = ?"
	testIDCondition            = "id = ?"
	testEmailCondition         = "email = ?"
	testPostgresDriver         = "postgres"
	testMySQLDriver            = "mysql"
	testSQLite3URL             = "sqlite3://test.db"
	testPostgresURL            = "postgres://user:pass@localhost/db"
	testUnknownURL             = "unknown://test"
	testSelectQuery            = "SELECT * FROM users WHERE id = ? AND name = ?"
	testSelectSingleQuery      = "SELECT * FROM users WHERE id = ?"
	testPostgresQuery          = "SELECT * FROM users WHERE id = $1 AND name = $2"
	testExpectedGotFormat      = "Expected %s, got %s"
	testNameLikePattern        = "A%"
	testAliceLikeName          = "name"
	testExpectedErrorFormat    = "Expected error '%s', got '%s'"
	testUpdateNotImplemented   = "update not yet implemented"
	testDeleteNotImplemented   = "delete not yet implemented"
	testFindByIDNotImplemented = "findByID not yet implemented"
	testExpectedTableName      = "test_users"
	testExpectedBaseTableName  = "base_entitys"
	testCreateTableQuery       = `
		CREATE TABLE testusers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	testCreateEFTableQuery = `
		CREATE TABLE test_user_entitys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
)

// Test entity for testing
type TestUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

// TestUserEntity that implements EntityInterface for EF testing
type TestUserEntity struct {
	BaseEntity
	Name     string `db:"name"`
	Email    string `db:"email"`
	IsActive bool   `db:"is_active"`
}

// TableName returns the table name for TestUser
func (TestUser) TableName() string {
	return "testusers"
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open(testSQLite3Driver, testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test table
	_, err = db.Exec(testCreateTableQuery)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestNewEnhancedDbContext(t *testing.T) {
	ctx, err := NewEnhancedDbContext(testMemoryDB)
	if err != nil {
		t.Fatalf("Failed to create enhanced db context: %v", err)
	}
	defer func() {
		if closeErr := ctx.db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	if ctx.db == nil {
		t.Error("Database connection should not be nil")
	}
	if ctx.ChangeTracker == nil {
		t.Error(testChangeTrackerNil)
	}
	if ctx.Database == nil {
		t.Error("Database should not be nil")
	}
}

func TestNewEnhancedDbContextWithDB(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)
	if ctx.db != db {
		t.Error("Database should be the same as passed")
	}
	if ctx.ChangeTracker == nil {
		t.Error(testChangeTrackerNil)
	}
}

func TestNewEnhancedDbContextWithTx(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			t.Logf("Failed to rollback transaction: %v", rollbackErr)
		}
	}()

	ctx := NewEnhancedDbContextWithTx(tx)
	if ctx.tx != tx {
		t.Error("Transaction should be the same as passed")
	}
	if ctx.ChangeTracker == nil {
		t.Error(testChangeTrackerNil)
	}
}

func TestChangeTracker(t *testing.T) {
	tracker := NewChangeTracker()
	if tracker == nil {
		t.Fatal(testChangeTrackerNil)
	}

	user := &TestUser{Name: testUserName, Email: "test@example.com"}

	// Test tracking entity
	tracker.TrackEntity(user, EntityStateAdded)

	// Test getting state
	state := tracker.GetEntityState(user)
	if state != EntityStateAdded {
		t.Errorf("Expected EntityStateAdded, got %v", state)
	}

	// Test changing state
	tracker.TrackEntity(user, EntityStateModified)
	state = tracker.GetEntityState(user)
	if state != EntityStateModified {
		t.Errorf("Expected EntityStateModified, got %v", state)
	}
}

func TestDatabaseBegin(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	database := NewDatabase(db)
	tx, err := database.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			t.Logf("Failed to rollback transaction: %v", rollbackErr)
		}
	}()

	if tx == nil {
		t.Error("Transaction should not be nil")
	}
}

func TestAddUpdateDeleteSaveChanges(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)

	// Test Add and SaveChanges
	user := &TestUser{
		Name:     testJohnDoe,
		Email:    testJohnEmail,
		IsActive: true,
	}

	ctx.Add(user)
	_, err := ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to save changes: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be set after insert")
	}

	// Test Update
	user.Name = testJohnUpdated
	ctx.Update(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	var count int
	err = db.QueryRow(testCountQuery, testJohnUpdated).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query updated user: %v", err)
	}
	if count != 1 {
		t.Error("Updated user should exist in database")
	}

	// Test Delete
	ctx.Delete(user)
	_, err = ctx.SaveChanges()
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify delete
	err = db.QueryRow(testCountDeletedQuery, user.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query deleted user: %v", err)
	}
	if count != 0 {
		t.Error("Deleted user should not exist in database")
	}
}

func TestSetEntityState(t *testing.T) {
	tracker := NewChangeTracker()
	user := &TestUser{Name: testUserName}

	tracker.SetEntityState(user, EntityStateModified)
	state := tracker.GetEntityState(user)
	if state != EntityStateModified {
		t.Errorf("Expected EntityStateModified, got %v", state)
	}
}

func TestEnhancedDbSetWhere(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test Where
	results, err := set.Where(testIsActiveCondition, true).ToList()
	if err != nil {
		t.Fatalf("Failed to execute Where query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != testAliceName {
		t.Errorf(testExpectedGotFormat, testAliceName, results[0].Name)
	}
}

func TestEnhancedDbSetWhereLike(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test WhereLike
	results, err := set.WhereLike(testAliceLikeName, testNameLikePattern).ToList()
	if err != nil {
		t.Fatalf("Failed to execute WhereLike query: %v", err)
	}
	if len(results) != 1 || results[0].Name != testAliceName {
		t.Error("WhereLike should find Alice")
	}
}

func TestEnhancedDbSetWhereInAndOr(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test WhereIn
	results, err := set.WhereIn(testAliceLikeName, []interface{}{testAliceName, "Charlie"}).ToList()
	if err != nil {
		t.Fatalf("Failed to execute WhereIn query: %v", err)
	}
	if len(results) != 1 || results[0].Name != testAliceName {
		t.Error("WhereIn should find Alice")
	}

	// Test WhereOr
	results, err = set.WhereOr(testNameCondition+" OR "+testNameCondition, testAliceName, testBobName).ToList()
	if err != nil {
		t.Fatalf("Failed to execute WhereOr query: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("WhereOr should find 2 results, got %d", len(results))
	}
}

func TestEnhancedDbSetOrderBy(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test OrderBy
	results, err := set.OrderBy(testAliceLikeName).ToList()
	if err != nil {
		t.Fatalf("Failed to execute OrderBy query: %v", err)
	}
	if len(results) != 2 || results[0].Name != testAliceName {
		t.Error("OrderBy should return Alice first")
	}

	// Test OrderByDescending
	results, err = set.OrderByDescending(testAliceLikeName).ToList()
	if err != nil {
		t.Fatalf("Failed to execute OrderByDescending query: %v", err)
	}
	if len(results) != 2 || results[0].Name != testBobName {
		t.Error("OrderByDescending should return Bob first")
	}
}

func TestEnhancedDbSetTakeSkip(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test Take
	results, err := set.Take(1).ToList()
	if err != nil {
		t.Fatalf("Failed to execute Take query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Take(1) should return 1 result, got %d", len(results))
	}

	// Test Skip
	results, err = set.Skip(1).ToList()
	if err != nil {
		t.Fatalf("Failed to execute Skip query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Skip(1) should return 1 result, got %d", len(results))
	}
}

func TestEnhancedDbSetCountAny(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test Count
	count, err := set.Count()
	if err != nil {
		t.Fatalf("Failed to execute Count: %v", err)
	}
	if count != 2 {
		t.Errorf("Count should return 2, got %d", count)
	}

	// Test Any
	exists, err := set.Any()
	if err != nil {
		t.Fatalf("Failed to execute Any: %v", err)
	}
	if !exists {
		t.Error("Any should return true")
	}

	// Test Any with condition
	exists, err = set.Where(testNameCondition, testNonExistent).Any()
	if err != nil {
		t.Fatalf("Failed to execute Any with condition: %v", err)
	}
	if exists {
		t.Error("Any with non-existent condition should return false")
	}
}

func TestEnhancedDbSetFirstOrDefault(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test FirstOrDefault
	user, err := set.Where(testNameCondition, testAliceName).FirstOrDefault()
	if err != nil {
		t.Fatalf("Failed to execute FirstOrDefault: %v", err)
	}
	if user == nil || user.Name != testAliceName {
		t.Error("FirstOrDefault should return Alice")
	}

	// Test FirstOrDefault with no results
	user, err = set.Where(testNameCondition, testNonExistent).FirstOrDefault()
	if err != nil {
		t.Fatalf("Failed to execute FirstOrDefault with no results: %v", err)
	}
	if user != nil {
		t.Error("FirstOrDefault should return nil for non-existent record")
	}
}

func TestEnhancedDbSetFirstSingle(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test First
	user, err := set.Where(testNameCondition, testAliceName).First()
	if err != nil {
		t.Fatalf("Failed to execute First: %v", err)
	}
	if user.Name != testAliceName {
		t.Error("First should return Alice")
	}

	// Test Single
	user, err = set.Where(testNameCondition, testAliceName).Single()
	if err != nil {
		t.Fatalf("Failed to execute Single: %v", err)
	}
	if user.Name != testAliceName {
		t.Error("Single should return Alice")
	}
}

func TestEnhancedDbSetFind(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test Find
	var userID int64
	err = db.QueryRow("SELECT id FROM testusers WHERE "+testNameCondition, testAliceName).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	user, err := set.Find(userID)
	if err != nil {
		t.Fatalf("Failed to execute Find: %v", err)
	}
	if user == nil || user.Name != testAliceName {
		t.Error("Find should return Alice")
	}
}

func TestAsNoTracking(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test AsNoTracking
	noTrackingSet := set.AsNoTracking()
	if !noTrackingSet.noTracking {
		t.Error("AsNoTracking should set noTracking to true")
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test detectDatabaseDriver with actual database connections
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	driver := detectDatabaseDriver(db)
	if driver != testSQLite3Driver {
		t.Errorf(testExpectedGotFormat, testSQLite3Driver, driver)
	}

	// Test convertQueryPlaceholders
	query := convertQueryPlaceholders(testSelectQuery, testPostgresDriver)
	if query != testPostgresQuery {
		t.Errorf(testExpectedGotFormat, testPostgresQuery, query)
	}

	query = convertQueryPlaceholders(testSelectSingleQuery, testSQLite3Driver)
	if query != testSelectSingleQuery {
		t.Errorf(testExpectedGotFormat, testSelectSingleQuery, query)
	}
}

func TestStringMethod(t *testing.T) {
	// Test EntityState String method
	states := []EntityState{
		EntityStateUnchanged,
		EntityStateAdded,
		EntityStateModified,
		EntityStateDeleted,
	}

	expected := []string{
		"Unchanged",
		"Added",
		"Modified",
		"Deleted",
	}

	for i, state := range states {
		if state.String() != expected[i] {
			t.Errorf(testExpectedGotFormat, expected[i], state.String())
		}
	}

	// Test unknown state
	unknownState := EntityState(999)
	if unknownState.String() != "Unknown" {
		t.Errorf(testExpectedGotFormat, "Unknown", unknownState.String())
	}
}

// Test EF Context functionality
func TestEFContext(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Test NewEFContext
	ctx := NewEFContext(db)
	if ctx == nil {
		t.Fatal("EFContext should not be nil")
	}
	if ctx != nil && ctx.db != db {
		t.Error("EFContext should have the correct database")
	}

	// Test BaseEntity methods
	entity := &BaseEntity{ID: 1}

	// Test GetID
	id := entity.GetID()
	if id != uint(1) {
		t.Errorf("GetID should return 1, got %v", id)
	}

	// Test SetID
	entity.SetID(uint(42))
	if entity.ID != 42 {
		t.Errorf("SetID should set ID to 42, got %v", entity.ID)
	}

	// Test SetID with invalid type (should not crash)
	entity.SetID("invalid")
	if entity.ID != 42 {
		t.Error("SetID with invalid type should not change ID")
	}

	// Test ExtractFieldsForDebug
	columns, values := ctx.ExtractFieldsForDebug(entity)
	if len(columns) == 0 || len(values) == 0 {
		t.Error("ExtractFieldsForDebug should return columns and values")
	}

	// Test SaveChanges (should not error)
	err := ctx.SaveChanges()
	if err != nil {
		t.Errorf("SaveChanges should not error, got %v", err)
	}
}

func TestEFContextErrors(t *testing.T) {
	// Test with nil database
	ctx := NewEFContext(nil)

	entity := &BaseEntity{ID: 1}

	// Test Add with nil database
	err := ctx.Add(entity)
	if err == nil {
		t.Error("Add should return error with nil database")
	}

	// Test Update with nil database
	err = ctx.Update(entity)
	if err == nil {
		t.Error("Update should return error with nil database")
	}

	// Test Remove with nil database
	err = ctx.Remove(entity)
	if err == nil {
		t.Error("Remove should return error with nil database")
	}

	// Test Find with nil database
	err = ctx.Find(entity, 1)
	if err == nil {
		t.Error("Find should return error with nil database")
	}
}

// Test utility functions for better coverage
func TestUtilityFunctionsCoverage(t *testing.T) {
	// Test toSnakeCase function
	testCases := []struct {
		input    string
		expected string
	}{
		{"UserName", "user_name"},
		{"userName", "user_name"},
		{"user_name", "user_name"},
		{"", ""},
		{"A", "a"},
		{"CamelCase", "camel_case"},
	}

	for _, tc := range testCases {
		result := toSnakeCase(tc.input)
		if result != tc.expected {
			t.Errorf("toSnakeCase(%s): expected %s, got %s", tc.input, tc.expected, result)
		}
	}

	// Test toCamelCase function (already has some coverage but let's add more cases)
	camelCases := []struct {
		input    string
		expected string
	}{
		{"user_name", "UserName"},
		{"user_id", "UserId"},
		{"created_at", "CreatedAt"},
		{"", ""},
		{"id", "Id"},
		{"user", "User"},
		{"user_name_field", "UserNameField"},
	}

	for _, tc := range camelCases {
		result := toCamelCase(tc.input)
		if result != tc.expected {
			t.Errorf("toCamelCase(%s): expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}

func TestDatabaseDriverDetection(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Test detectDatabaseDriver with SQLite
	driver := detectDatabaseDriver(db)
	if driver != testSQLite3Driver {
		t.Errorf("Expected sqlite3 driver, got %s", driver)
	}
}

func TestFieldSettersIndirect(t *testing.T) {
	// We can't directly test the field setter functions because they don't return errors
	// Instead, we test them through setFieldValue which calls them indirectly

	// Test string field
	var strEntity struct {
		Value string
	}
	field := reflect.ValueOf(&strEntity).Elem().Field(0)
	err := setFieldValue(field, "test string")
	if err != nil {
		t.Errorf("setFieldValue for string should not error: %v", err)
	}
	if strEntity.Value != "test string" {
		t.Errorf("Expected 'test string', got %s", strEntity.Value)
	}

	// Test int field
	var intEntity struct {
		Value int64
	}
	field = reflect.ValueOf(&intEntity).Elem().Field(0)
	err = setFieldValue(field, int64(42))
	if err != nil {
		t.Errorf("setFieldValue for int should not error: %v", err)
	}
	if intEntity.Value != 42 {
		t.Errorf("Expected 42, got %d", intEntity.Value)
	}

	// Test uint field
	var uintEntity struct {
		Value uint64
	}
	field = reflect.ValueOf(&uintEntity).Elem().Field(0)
	err = setFieldValue(field, int64(42))
	if err != nil {
		t.Errorf("setFieldValue for uint should not error: %v", err)
	}
	if uintEntity.Value != 42 {
		t.Errorf("Expected 42, got %d", uintEntity.Value)
	}

	// Test float field
	var floatEntity struct {
		Value float64
	}
	field = reflect.ValueOf(&floatEntity).Elem().Field(0)
	err = setFieldValue(field, float64(3.14))
	if err != nil {
		t.Errorf("setFieldValue for float should not error: %v", err)
	}
	if floatEntity.Value != 3.14 {
		t.Errorf("Expected 3.14, got %f", floatEntity.Value)
	}

	// Test bool field
	var boolEntity struct {
		Value bool
	}
	field = reflect.ValueOf(&boolEntity).Elem().Field(0)
	err = setFieldValue(field, true)
	if err != nil {
		t.Errorf("setFieldValue for bool should not error: %v", err)
	}
	if !boolEntity.Value {
		t.Error("Expected true")
	}

	// Test time field
	var timeEntity struct {
		Value time.Time
	}
	field = reflect.ValueOf(&timeEntity).Elem().Field(0)
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	err = setFieldValue(field, now)
	if err != nil {
		t.Errorf("setFieldValue for time should not error: %v", err)
	}
	if !timeEntity.Value.Equal(now) {
		t.Errorf("Time values should be equal: got %v, expected %v", timeEntity.Value, now)
	}
}

func TestEmbeddedStructHandling(t *testing.T) {
	// Test embedded struct handling with a custom struct
	type EmbeddedStruct struct {
		Field1 string `db:"field1"`
		Field2 int    `db:"field2"`
	}

	type TestEntity struct {
		ID       int64          `db:"id"`
		Name     string         `db:"name"`
		Embedded EmbeddedStruct `db:"-"` // Embedded struct
	}

	entity := &TestEntity{
		ID:   1,
		Name: "test",
		Embedded: EmbeddedStruct{
			Field1: "embedded_value",
			Field2: 42,
		},
	}

	// Test getFieldData with embedded structs (this should cover handleEmbeddedStruct)
	columns, values, placeholders := getFieldData(entity, false, testSQLite3Driver)

	if len(columns) == 0 {
		t.Error("getFieldData should return columns for embedded struct")
	}
	if len(values) == 0 {
		t.Error("getFieldData should return values for embedded struct")
	}
	if len(placeholders) == 0 {
		t.Error("getFieldData should return placeholders for embedded struct")
	}
}

func TestGetTableName(t *testing.T) {
	// Test getTableName with TableName method
	user := &TestUser{}
	tableName := getTableName(user)
	if tableName != "testusers" {
		t.Errorf("Expected 'testusers', got %s", tableName)
	}

	// Test getTableName without TableName method
	type SimpleEntity struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	simple := &SimpleEntity{}
	tableName = getTableName(simple)
	if tableName != "simple_entity" {
		t.Errorf("Expected 'simple_entity', got %s", tableName)
	}
}

func TestPlaceholderHandling(t *testing.T) {
	// Test getPlaceholder for different drivers
	sqlitePlaceholder := getPlaceholder(testSQLite3Driver, 0)
	if sqlitePlaceholder != "?" {
		t.Errorf("Expected '?', got %s", sqlitePlaceholder)
	}

	postgresPlaceholder := getPlaceholder(testPostgresDriver, 0)
	if postgresPlaceholder != "$1" {
		t.Errorf("Expected '$1', got %s", postgresPlaceholder)
	}

	postgresPlaceholder2 := getPlaceholder(testPostgresDriver, 2)
	if postgresPlaceholder2 != "$3" {
		t.Errorf("Expected '$3', got %s", postgresPlaceholder2)
	}
}

func TestColumnNameExtraction(t *testing.T) {
	// Test getColumnNameFromFieldData
	type TestStruct struct {
		Field1 string `db:"custom_name"`
		Field2 string // No db tag, should use snake_case conversion
	}

	// Test with db tag
	field1, _ := reflect.TypeOf(TestStruct{}).FieldByName("Field1")
	columnName := getColumnNameFromFieldData(field1)
	if columnName != "custom_name" {
		t.Errorf("Expected 'custom_name', got %s", columnName)
	}

	// Test without db tag (should convert to snake_case)
	field2, _ := reflect.TypeOf(TestStruct{}).FieldByName("Field2")
	columnName = getColumnNameFromFieldData(field2)
	if columnName != "field2" {
		t.Errorf("Expected 'field2', got %s", columnName)
	}
}

func TestShouldSkipField(t *testing.T) {
	type TestStruct struct {
		ID           int64  `db:"id"`
		Name         string `db:"name"`
		IgnoredField string `db:"-"`
		SQLIgnored   string `sql:"-"`
		PublicField  string
	}

	structType := reflect.TypeOf(TestStruct{})

	// Test ID field exclusion
	idField, _ := structType.FieldByName("ID")
	if !shouldSkipField(idField, true) {
		t.Error("ID field should be skipped when excludeID is true")
	}
	if shouldSkipField(idField, false) {
		t.Error("ID field should not be skipped when excludeID is false")
	}

	// Test db tag exclusion
	ignoredField, _ := structType.FieldByName("IgnoredField")
	if !shouldSkipField(ignoredField, false) {
		t.Error("Field with db:\"-\" should be skipped")
	}

	// Test sql tag exclusion
	sqlIgnoredField, _ := structType.FieldByName("SQLIgnored")
	if !shouldSkipField(sqlIgnoredField, false) {
		t.Error("Field with sql:\"-\" should be skipped")
	}

	// Test public field inclusion
	publicField, _ := structType.FieldByName("PublicField")
	if shouldSkipField(publicField, false) {
		t.Error("Public field should not be skipped")
	}
}

func TestEntityStateEnumCoverage(t *testing.T) {
	// Test all possible entity states to ensure complete coverage
	tracker := NewChangeTracker()
	entity := &TestUser{Name: "Test"}

	// Test all entity states
	states := []EntityState{
		EntityStateUnchanged,
		EntityStateAdded,
		EntityStateModified,
		EntityStateDeleted,
	}

	for _, state := range states {
		tracker.SetEntityState(entity, state)
		retrievedState := tracker.GetEntityState(entity)
		if retrievedState != state {
			t.Errorf("Expected state %v, got %v", state, retrievedState)
		}
	}

	// Test getting state for untracked entity
	newEntity := &TestUser{Name: "New"}
	state := tracker.GetEntityState(newEntity)
	if state != EntityStateUnchanged {
		t.Errorf("Untracked entity should have Unchanged state, got %v", state)
	}
}

func TestAdvancedQueryScenarios(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}
	_, err = db.Exec(testInsertQuery, testBobName, testBobEmail, false)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test Single with multiple results (should error)
	_, err = set.Single()
	if err == nil {
		t.Error("Single should error when multiple results exist")
	}

	// Test First with no results (should error)
	_, err = set.Where("name = ?", testNonExistent).First()
	if err == nil {
		t.Error("First should error when no results exist")
	}
}

func TestScanEntityAdvanced(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	// Insert test data with various field types
	_, err := db.Exec(testInsertQuery, testAliceName, testAliceEmail, true)
	if err != nil {
		t.Fatalf(testInsertDataError, err)
	}

	// Query to test scanEntity function
	rows, err := db.Query("SELECT id, name, email, is_active, created_at FROM testusers WHERE name = ?", testAliceName)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			t.Logf("Failed to close rows: %v", closeErr)
		}
	}()

	if rows.Next() {
		entity := &TestUser{}
		err = scanEntity(rows, entity)
		if err != nil {
			t.Errorf("scanEntity should not error: %v", err)
		}
		if entity.Name != testAliceName {
			t.Errorf("Expected name %s, got %s", testAliceName, entity.Name)
		}
		if entity.Email != testAliceEmail {
			t.Errorf("Expected email %s, got %s", testAliceEmail, entity.Email)
		}
		if !entity.IsActive {
			t.Error("Expected IsActive to be true")
		}
	} else {
		t.Error("Expected at least one row")
	}
}

// Additional test for improved String method coverage on ChangeTracker
func TestChangeTrackerString(t *testing.T) {
	tracker := NewChangeTracker()
	entity1 := &TestUser{Name: "User1"}
	entity2 := &TestUser{Name: "User2"}

	// Test empty tracker
	str := tracker.String()
	if str == "" {
		t.Error("ChangeTracker String should return some representation")
	}

	// Add entities with different states
	tracker.SetEntityState(entity1, EntityStateAdded)
	tracker.SetEntityState(entity2, EntityStateModified)

	str = tracker.String()
	if str == "" {
		t.Error("ChangeTracker String should return representation with entities")
	}
}

func TestPlaceholderAdjustment(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[TestUser](ctx)

	// Test complex where conditions to improve adjustPlaceholdersForCondition coverage
	newSet := *set
	newSet.ctx.driver = testPostgresDriver

	// Test condition with multiple placeholders
	condition := "name = ? AND email = ? AND is_active = ?"
	adjustedCondition := newSet.adjustPlaceholdersForCondition(condition)
	expectedCondition := "name = $1 AND email = $2 AND is_active = $3"
	if adjustedCondition != expectedCondition {
		t.Errorf("Expected %s, got %s", expectedCondition, adjustedCondition)
	}

	// Test condition without placeholders
	condition = "is_active = true"
	adjustedCondition = newSet.adjustPlaceholdersForCondition(condition)
	if adjustedCondition != condition {
		t.Errorf("Condition without placeholders should remain unchanged")
	}
}

// Test EF Context insert method with 0% coverage
func TestEFContextInsertMethod(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf(testDBCloseError, closeErr)
		}
	}()

	ctx := NewEFContext(db)

	// Create table for testing
	_, err := db.Exec("DROP TABLE IF EXISTS test_user_entitys")
	if err != nil {
		t.Logf("Warning: could not drop table: %v", err)
	}
	_, err = db.Exec(testCreateEFTableQuery)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test insert method (0% coverage)
	entity := &TestUserEntity{
		Name:     testAliceName,
		Email:    testAliceEmail,
		IsActive: true,
	}
	err = ctx.insert(entity)
	if err != nil {
		t.Errorf("insert should work with valid entity and database: %v", err)
	}
}

// Test EF Context unimplemented methods
func TestEFContextUnimplementedMethods(t *testing.T) {
	ctx := NewEFContext(nil) // Don't need DB for error tests
	entity := &BaseEntity{ID: 1}

	// Test update method (0% coverage) - should return not implemented error
	err := ctx.update(entity)
	if err == nil {
		t.Error("update should return 'not yet implemented' error")
	}
	if err.Error() != testUpdateNotImplemented {
		t.Errorf(testExpectedErrorFormat, testUpdateNotImplemented, err.Error())
	}

	// Test delete method (0% coverage) - should return not implemented error
	err = ctx.delete(entity)
	if err == nil {
		t.Error("delete should return 'not yet implemented' error")
	}
	if err.Error() != testDeleteNotImplemented {
		t.Errorf(testExpectedErrorFormat, testDeleteNotImplemented, err.Error())
	}

	// Test findByID method (0% coverage) - should return not implemented error
	err = ctx.findByID(entity, 1)
	if err == nil {
		t.Error("findByID should return 'not yet implemented' error")
	}
	if err.Error() != testFindByIDNotImplemented {
		t.Errorf(testExpectedErrorFormat, testFindByIDNotImplemented, err.Error())
	}
}

// Test EF Context helper methods with 0% coverage
func TestEFContextHelperMethods(t *testing.T) {
	ctx := NewEFContext(nil) // Don't need DB for these helper methods

	// Test getTableNameFromType (0% coverage)
	userType := reflect.TypeOf(TestUser{})
	tableName := ctx.getTableNameFromType(userType)
	if tableName != testExpectedTableName {
		t.Errorf("Expected table name '%s', got '%s'", testExpectedTableName, tableName)
	}

	// Test with a different type
	baseType := reflect.TypeOf(BaseEntity{})
	baseTableName := ctx.getTableNameFromType(baseType)
	if baseTableName != testExpectedBaseTableName {
		t.Errorf("Expected table name '%s', got '%s'", testExpectedBaseTableName, baseTableName)
	}

	// Test toSnakeCaseEF (0% coverage)
	testCases := []struct {
		input    string
		expected string
	}{
		{"UserName", "user_name"},
		{"TestUser", "test_user"},
		{"BaseEntity", "base_entity"},
		{"ID", "i_d"},
		{"HTMLParser", "h_t_m_l_parser"},
		{"", ""},
		{"A", "a"},
	}

	for _, tc := range testCases {
		result := ctx.toSnakeCaseEF(tc.input)
		if result != tc.expected {
			t.Errorf("toSnakeCaseEF(%s): expected %s, got %s", tc.input, tc.expected, result)
		}
	}
}

// Test setTimestamps method (0% coverage)
func TestEFContextSetTimestamps(t *testing.T) {
	ctx := NewEFContext(nil)

	// Create entity with timestamp fields
	entity := &BaseEntity{}
	v := reflect.ValueOf(entity).Elem()

	// Test setTimestamps for insert (should set both CreatedAt and UpdatedAt)
	beforeTime := time.Now()
	ctx.setTimestamps(v, true)
	afterTime := time.Now()

	if entity.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set for insert")
	}
	if entity.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set for insert")
	}
	if entity.CreatedAt.Before(beforeTime) || entity.CreatedAt.After(afterTime) {
		t.Error("CreatedAt should be within expected time range")
	}
	if entity.UpdatedAt.Before(beforeTime) || entity.UpdatedAt.After(afterTime) {
		t.Error("UpdatedAt should be within expected time range")
	}

	// Test setTimestamps for update (should only set UpdatedAt)
	originalCreatedAt := entity.CreatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	beforeUpdateTime := time.Now()
	ctx.setTimestamps(v, false)
	afterUpdateTime := time.Now()

	if !entity.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should not change for update")
	}
	if entity.UpdatedAt.Before(beforeUpdateTime) || entity.UpdatedAt.After(afterUpdateTime) {
		t.Error("UpdatedAt should be updated within expected time range")
	}
}

// Test extractFieldsForInsert method (0% coverage)
func TestEFContextExtractFieldsForInsert(t *testing.T) {
	ctx := NewEFContext(nil)

	// Create test entity
	entity := &TestUserEntity{
		Name:     testAliceName,
		Email:    testAliceEmail,
		IsActive: true,
	}
	v := reflect.ValueOf(entity).Elem()

	// Test extractFieldsForInsert
	columns, values, placeholders := ctx.extractFieldsForInsert(v)

	if len(columns) == 0 {
		t.Error("extractFieldsForInsert should return columns")
	}
	if len(values) == 0 {
		t.Error("extractFieldsForInsert should return values")
	}
	if len(placeholders) == 0 {
		t.Error("extractFieldsForInsert should return placeholders")
	}
	if len(columns) != len(values) || len(values) != len(placeholders) {
		t.Error("columns, values, and placeholders should have same length")
	}

	// Check that ID is excluded from insert
	for _, col := range columns {
		if strings.ToLower(col) == "id" {
			t.Error("ID column should be excluded from insert")
		}
	}

	// Check for expected columns (name, email, is_active should be present)
	hasName := false
	hasEmail := false
	hasIsActive := false
	for _, col := range columns {
		switch col {
		case "name":
			hasName = true
		case "email":
			hasEmail = true
		case "is_active":
			hasIsActive = true
		}
	}
	if !hasName {
		t.Error("Expected 'name' column in insert")
	}
	if !hasEmail {
		t.Error("Expected 'email' column in insert")
	}
	if !hasIsActive {
		t.Error("Expected 'is_active' column in insert")
	}
}

package dbcontext

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test constants to avoid goconst violations
const (
	testDBCloseError      = "Failed to close database: %v"
	testChangeTrackerNil  = "ChangeTracker should not be nil"
	testUserName          = "Test User"
	testInsertQuery       = "INSERT INTO testusers (name, email, is_active) VALUES (?, ?, ?)"
	testInsertDataError   = "Failed to insert test data: %v"
	testNameCondition     = "name = ?"
	testJohnDoe           = "John Doe"
	testJohnEmail         = "john@example.com"
	testJohnUpdated       = "John Updated"
	testAliceName         = "Alice"
	testAliceEmail        = "alice@example.com"
	testBobName           = "Bob"
	testBobEmail          = "bob@example.com"
	testNonExistent       = "NonExistent"
	testCountQuery        = "SELECT COUNT(*) FROM testusers WHERE name = ?"
	testCountDeletedQuery = "SELECT COUNT(*) FROM testusers WHERE id = ?"
	testSQLite3Driver     = "sqlite3"
	testMemoryDB          = ":memory:"
	testIsActiveCondition = "is_active = ?"
	testIDCondition       = "id = ?"
	testEmailCondition    = "email = ?"
	testPostgresDriver    = "postgres"
	testMySQLDriver       = "mysql"
	testSQLite3URL        = "sqlite3://test.db"
	testPostgresURL       = "postgres://user:pass@localhost/db"
	testUnknownURL        = "unknown://test"
	testSelectQuery       = "SELECT * FROM users WHERE id = ? AND name = ?"
	testSelectSingleQuery = "SELECT * FROM users WHERE id = ?"
	testPostgresQuery     = "SELECT * FROM users WHERE id = $1 AND name = $2"
	testExpectedGotFormat = "Expected %s, got %s"
	testNameLikePattern   = "A%"
	testAliceLikeName     = "name"
)

// Test entity for testing
type TestUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
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
	_, err = db.Exec(`
		CREATE TABLE testusers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
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
	tracker := NewChangeTracker()
	user := &TestUser{Name: testUserName}

	tracker.TrackEntity(user, EntityStateAdded)

	// Test String method
	str := tracker.String()
	if str == "" {
		t.Error("String method should return non-empty string")
	}
}

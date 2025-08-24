package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
)

// openPGForMigrations mirrors dbcontext test helper
func openPGForMigrations(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("GRA_TEST_PG") == "0" {
		t.Skip("GRA_TEST_PG=0; skipping Postgres tests")
	}
	host := getenv("PGHOST", "localhost")
	port := getenv("PGPORT", "5432")
	user := getenv("PGUSER", "postgres")
	pass := getenv("PGPASSWORD", "MyPassword_123")
	dbname := getenv("PGDATABASE", "gra_test")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("postgres driver open failed: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	// Proactively drop unrelated tables that may be created by other package tests
	// to keep migration detection stable and deterministic.
	_, _ = db.Exec(`DROP TABLE IF EXISTS as_no_track`)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// simple dir helper mirroring hybrid_test style
func tempMigrationsDir(t *testing.T) string {
	dir := t.TempDir()
	return filepath.Join(dir, "migrations")
}

// local getenv helper
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// minimal model for PG
type PGUser struct {
	ID    int64  `db:"id" migration:"primary_key,auto_increment"`
	Email string `db:"email" migration:"unique,not_null,max_length:255"`
	Name  string `db:"name" migration:"not_null,max_length:100"`
}

const tblPGUsers = "pg_users"

func (PGUser) TableName() string { return tblPGUsers }

// Test that DatabaseInspector on Postgres returns expected schema after applying SQLGenerator output
func TestPostgres_Migrations_CreateAndInspect(t *testing.T) {
	db := openPGForMigrations(t)
	// clean
	_, _ = db.Exec(`DROP TABLE IF EXISTS pg_users`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS test_entities`)

	migrationsDir := tempMigrationsDir(t)
	migrator := NewHybridMigrator(db, PostgreSQL, migrationsDir)

	// Register model and detect changes
	migrator.DbSet(&PGUser{})
	plan, err := migrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges error: %v", err)
	}
	if len(plan.Changes) == 0 {
		t.Fatalf("expected changes for new table")
	}

	// Some concurrent package tests may leave unrelated tables around (e.g., as_no_track).
	// Filter changes to only include the CreateTable for our target to keep this test deterministic.
	filtered := make([]MigrationChange, 0, len(plan.Changes))
	for _, ch := range plan.Changes {
		if ch.Type == CreateTable && ch.TableName == "pg_users" {
			filtered = append(filtered, ch)
		}
	}
	plan = &MigrationPlan{Changes: filtered, PlanChecksum: plan.PlanChecksum}

	// Generate SQL and apply only our filtered plan
	ql, err := migrator.sqlGenerator.GenerateMigrationSQL(plan)
	if err != nil {
		t.Fatalf("GenerateMigrationSQL error: %v", err)
	}
	if _, err := db.Exec(ql.UpScript); err != nil {
		t.Fatalf("apply up script failed: %v\nSQL:\n%s", err, ql.UpScript)
	}

	// Inspect and verify table exists with columns
	schema, err := migrator.inspector.GetCurrentSchema()
	if err != nil {
		t.Fatalf("GetCurrentSchema error: %v", err)
	}
	tbl, ok := schema[tblPGUsers]
	if !ok {
		t.Fatalf("pg_users table not found in inspector")
	}
	if _, ok := tbl.Columns["id"]; !ok {
		t.Errorf("id column missing")
	}
	if _, ok := tbl.Columns["email"]; !ok {
		t.Errorf("email column missing")
	}
	if _, ok := tbl.Columns["name"]; !ok {
		t.Errorf("name column missing")
	}
}

// Evolve model: add column and ensure change detection + SQL generation handles ALTER on PG
type PGUserWithBio struct {
	PGUser
	Bio string `db:"bio" migration:"type:TEXT"`
}

func (PGUserWithBio) TableName() string { return "pg_users" }

func TestPostgres_Migrations_AlterColumn_AddAndDrop(t *testing.T) {
	db := openPGForMigrations(t)
	// Ensure base table exists
	_, _ = db.Exec(`DROP TABLE IF EXISTS pg_users`)
	_, err := db.Exec(`CREATE TABLE pg_users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		name VARCHAR(100) NOT NULL
	)`)
	if err != nil {
		t.Fatalf("bootstrap base table: %v", err)
	}

	migrator := NewHybridMigrator(db, PostgreSQL, tempMigrationsDir(t))
	migrator.DbSet(&PGUserWithBio{})
	plan, err := migrator.changeDetector.DetectChanges()
	if err != nil {
		t.Fatalf("DetectChanges: %v", err)
	}
	// Expect AddColumn for bio
	add := 0
	for _, ch := range plan.Changes {
		if ch.Type == AddColumn && ch.ColumnName == "bio" {
			add++
		}
	}
	if add != 1 {
		t.Fatalf("expected 1 AddColumn bio, got %d changes: %+v", add, plan.Changes)
	}
	// Normalize a minimal plan with only the AddColumn change and NewValue set for SQL generator
	filtered := make([]MigrationChange, 0, 1)
	for _, ch := range plan.Changes {
		if ch.Type == AddColumn && ch.ColumnName == "bio" {
			ch.NewValue = ch.NewColumn
			filtered = append(filtered, ch)
		}
	}
	normPlan := &MigrationPlan{Changes: filtered, PlanChecksum: "test"}
	ql, err := migrator.sqlGenerator.GenerateMigrationSQL(normPlan)
	if err != nil {
		t.Fatalf("GenerateMigrationSQL: %v", err)
	}
	if _, err := db.Exec(ql.UpScript); err != nil {
		t.Fatalf("apply up alter failed: %v\nSQL:\n%s", err, ql.UpScript)
	}
	// Verify
	row := db.QueryRow(`SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='pg_users' AND column_name='bio'`)
	var col string
	if err := row.Scan(&col); err != nil {
		t.Fatalf("bio column not found after alter: %v", err)
	}
}

// Verify inspector discovers PK, unique index, and foreign key constraints.
func TestPostgres_Inspector_PK_Index_Constraints(t *testing.T) {
	db := openPGForMigrations(t)
	// Clean any leftovers
	_, _ = db.Exec(`DROP TABLE IF EXISTS orders`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS customers`)

	// Create schema: customers with PK, unique email; orders with FK to customers(id)
	_, err := db.Exec(`CREATE TABLE customers (
		id SERIAL PRIMARY KEY,
		email VARCHAR(200) UNIQUE NOT NULL,
		name TEXT
	)`)
	if err != nil {
		t.Skipf("setup customers failed: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE orders (
		id SERIAL PRIMARY KEY,
		customer_id INTEGER NOT NULL,
		amount NUMERIC(10,2) NOT NULL,
		CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES customers(id)
	)`)
	if err != nil {
		t.Fatalf("setup orders failed: %v", err)
	}
	// extra index for coverage of index parsing
	_, _ = db.Exec(`CREATE INDEX ix_orders_customer_id ON orders(customer_id)`)

	inspector := NewDatabaseInspector(db, PostgreSQL)
	schema, err := inspector.GetCurrentSchema()
	if err != nil {
		t.Fatalf("GetCurrentSchema: %v", err)
	}

	// customers checks
	cust, ok := schema["customers"]
	if !ok {
		t.Fatalf("customers not found")
	}
	// Primary key contains id
	foundPK := false
	for _, c := range cust.PrimaryKeys {
		if c == "id" {
			foundPK = true
			break
		}
	}
	if !foundPK {
		t.Fatalf("customers id not found in primary keys: %+v", cust.PrimaryKeys)
	}
	// Unique constraint should exist on email
	hasUnique := false
	for _, con := range cust.Constraints {
		if con.Type == "UNIQUE" {
			// Columns may be sorted; check contains email
			for _, col := range con.Columns {
				if col == "email" {
					hasUnique = true
					break
				}
			}
		}
	}
	if !hasUnique {
		t.Errorf("expected UNIQUE constraint on customers.email, got: %+v", cust.Constraints)
	}

	// orders checks
	ord, ok := schema["orders"]
	if !ok {
		t.Fatalf("orders not found")
	}
	if _, ok := ord.Indexes["ix_orders_customer_id"]; !ok {
		t.Errorf("expected index ix_orders_customer_id present, got: %+v", ord.Indexes)
	}
	hasFK := false
	for _, con := range ord.Constraints {
		if con.Type == "FOREIGN KEY" && con.ReferencedTable == "customers" {
			hasFK = true
			break
		}
	}
	if !hasFK {
		t.Errorf("expected FK to customers present, got: %+v", ord.Constraints)
	}
}

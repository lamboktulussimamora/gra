package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// Reuse the Postgres opener from migrations_postgres_test.go if available; provide a local fallback.
func openPGForAutoMigrator(t *testing.T) *sql.DB {
	t.Helper()
	// Honor explicit skip flag
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
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// Models for auto-migration tests
type AMCustomer struct {
	ID    int64  `db:"id" migration:"primary_key,auto_increment"`
	Email string `db:"email" migration:"unique,not_null,max_length:255"`
	Name  string `db:"name"`
}

func (AMCustomer) TableName() string { return "am_customers" }

// Add a new column to trigger updateTableSchema path
type AMCustomerV2 struct {
	AMCustomer
	Bio string `db:"bio" migration:"type:TEXT"`
}

func (AMCustomerV2) TableName() string { return "am_customers" }

// TestAutoMigrator_CreateAndUpdate_Postgres validates create + update flows using a real Postgres DB.
func TestAutoMigrator_CreateAndUpdate_Postgres(t *testing.T) {
	db := openPGForAutoMigrator(t)

	// Clean prior runs
	_, _ = db.Exec(`DROP TABLE IF EXISTS am_customers`)
	_, _ = db.Exec(`DROP TABLE IF EXISTS __migrations`)

	am := NewAutoMigrator(nil, db)

	// 1) Create initial schema
	if err := am.MigrateModels(AMCustomer{}); err != nil {
		t.Fatalf("initial MigrateModels failed: %v", err)
	}

	// __migrations record should exist
	var checksum1 string
	err := db.QueryRow(`SELECT checksum FROM __migrations WHERE migration_name = $1`, "create_table_am_customers").Scan(&checksum1)
	if err != nil || checksum1 == "" {
		t.Fatalf("expected migration checksum recorded, err=%v checksum=%q", err, checksum1)
	}

	// Table and columns should exist
	var cnt int
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='am_customers' AND column_name='email'`).Scan(&cnt); err != nil || cnt == 0 {
		t.Fatalf("expected am_customers.email column present, err=%v cnt=%d", err, cnt)
	}

	// 2) Re-run with same model -> should be up-to-date (no error)
	if err := am.MigrateModels(AMCustomer{}); err != nil {
		t.Fatalf("re-run MigrateModels failed: %v", err)
	}

	// 3) Update model (add new column) -> triggers updateTableSchema
	if err := am.MigrateModels(AMCustomerV2{}); err != nil {
		t.Fatalf("update MigrateModels failed: %v", err)
	}

	// New column should exist
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='am_customers' AND column_name='bio'`).Scan(&cnt); err != nil || cnt == 0 {
		t.Fatalf("expected am_customers.bio column present after update, err=%v cnt=%d", err, cnt)
	}

	// Checksum should change after schema update
	var checksum2 string
	if err := db.QueryRow(`SELECT checksum FROM __migrations WHERE migration_name = $1`, "create_table_am_customers").Scan(&checksum2); err != nil {
		t.Fatalf("read updated checksum failed: %v", err)
	}
	if checksum1 == checksum2 {
		t.Fatalf("expected checksum to change after schema update")
	}
}

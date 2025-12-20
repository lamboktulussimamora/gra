package dbcontext

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type PgAudit struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PgUser struct {
	PgAudit
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
}

func ensurePg(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("GRA_TEST_PG") == "0" {
		t.Skip("GRA_TEST_PG=0; skipping Postgres integration test")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:MyPassword_123@localhost:5432/gra_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("postgres driver open failed: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	// ensure a clean, expected schema (drop to avoid prior conflicting definitions)
	_, _ = db.Exec(`DROP TABLE IF EXISTS pg_users`)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS pg_users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(100) NOT NULL,
        email VARCHAR(200) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    )`)
	return db
}

func (PgUser) TableName() string { return "pg_users" }

func TestPG_InsertUpdateDelete_AndQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	db := ensurePg(t)
	defer func() { _ = db.Close() }()

	ctx := NewEnhancedDbContextWithDB(db)
	if ctx.driver != driverPostgres {
		t.Fatalf("expected driver postgres, got %s", ctx.driver)
	}

	// Clean table
	_, _ = db.Exec("DELETE FROM pg_users")

	// Insert
	u := &PgUser{Username: "alice", Email: "alice@example.com"}
	ctx.Add(u)
	n, err := ctx.SaveChanges()
	if err != nil || n != 1 {
		t.Fatalf("insert SaveChanges err=%v n=%d", err, n)
	}
	if u.ID == 0 {
		t.Fatalf("expected auto id set, got %d", u.ID)
	}

	// Update
	u.Email = "alice+1@example.com"
	ctx.Update(u)
	n, err = ctx.SaveChanges()
	if err != nil || n != 1 {
		t.Fatalf("update SaveChanges err=%v n=%d", err, n)
	}

	// Queries via DbSet
	set := NewEnhancedDbSet[PgUser](ctx)
	// Count
	cnt, err := set.Count()
	if err != nil || cnt != 1 {
		t.Fatalf("count err=%v cnt=%d", err, cnt)
	}
	// Any
	hasAny, err := set.Any()
	if err != nil || !hasAny {
		t.Fatalf("any err=%v any=%v", err, hasAny)
	}
	// Find by id
	got, err := set.Find(u.ID)
	if err != nil || got == nil || got.ID != u.ID {
		t.Fatalf("find err=%v got=%+v", err, got)
	}
	// First
	first, err := set.First()
	if err != nil || first == nil || first.ID == 0 {
		t.Fatalf("first err=%v first=%+v", err, first)
	}
	// ToList with WhereLike
	list, err := set.WhereLike("username", "ali%").ToList()
	if err != nil || len(list) != 1 {
		t.Fatalf("tolist err=%v len=%d", err, len(list))
	}

	// Delete
	ctx.Delete(u)
	n, err = ctx.SaveChanges()
	if err != nil || n != 1 {
		t.Fatalf("delete SaveChanges err=%v n=%d", err, n)
	}
}

package dbcontext

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Test entity mapped to a real SQLite table
type TestEntity struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Score     float64   `db:"score"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (TestEntity) TableName() string { return "test_entities" }

// SQLite tests removed per project policy; Postgres integration tests only.

// Postgres integration test using local/docker Postgres.
// Set GRA_TEST_PG=1 to enable; connection read from env or defaults to docker-compose.db.yml.
func TestEnhancedDbContext_CRUD_WithPostgres(t *testing.T) {
	if os.Getenv("GRA_TEST_PG") == "0" {
		t.Skip("GRA_TEST_PG=0; skipping Postgres integration test")
	}

	host := getenv("PGHOST", "localhost")
	port := getenv("PGPORT", "55432")
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

	// Ensure table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS test_entities (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            age INTEGER NOT NULL,
            score DOUBLE PRECISION NOT NULL,
            active BOOLEAN NOT NULL,
            created_at TIMESTAMP,
            updated_at TIMESTAMP
        );
    `)
	if err != nil {
		t.Fatalf("failed to ensure schema: %v", err)
	}
	// Clean table
	_, _ = db.Exec("DELETE FROM test_entities")

	ctx := NewEnhancedDbContextWithDB(db)

	e := &TestEntity{Name: "PG-Alice", Age: 30, Score: 91.5, Active: true}
	ctx.Add(e)
	affected, err := ctx.SaveChanges()
	if err != nil {
		t.Fatalf("pg insert SaveChanges error: %v", err)
	}
	if affected != 1 || e.ID == 0 {
		t.Fatalf("expected insert affected=1 and id set; got affected=%d id=%d", affected, e.ID)
	}

	got, err := NewEnhancedDbSet[TestEntity](ctx).Find(e.ID)
	if err != nil || got == nil || got.Name == "" {
		t.Fatalf("pg find failed: err=%v got=%+v", err, got)
	}

	e.Score = 92.0
	ctx.Update(e)
	if _, err := ctx.SaveChanges(); err != nil {
		t.Fatalf("pg update SaveChanges error: %v", err)
	}

	ctx.Delete(e)
	if _, err := ctx.SaveChanges(); err != nil {
		t.Fatalf("pg delete SaveChanges error: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

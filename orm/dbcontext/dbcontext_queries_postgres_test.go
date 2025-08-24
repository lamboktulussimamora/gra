package dbcontext

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"testing"

	_ "github.com/lib/pq"
)

func openPGForTest(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("GRA_TEST_PG") == "0" {
		t.Skip("GRA_TEST_PG=0; skipping Postgres tests")
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
	return db
}

func ensureSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
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
		t.Fatalf("ensure schema: %v", err)
	}
	_, _ = db.Exec("DELETE FROM test_entities")
}

func seedEntities(t *testing.T, ctx *EnhancedDbContext) []*TestEntity {
	t.Helper()
	seeds := []*TestEntity{
		{Name: "Alpha", Age: 20, Score: 80.5, Active: true},
		{Name: "Beta", Age: 35, Score: 70.0, Active: false},
		{Name: "Gamma", Age: 28, Score: 90.1, Active: true},
		{Name: "Delta", Age: 40, Score: 60.2, Active: false},
		{Name: "Alpine", Age: 22, Score: 88.8, Active: true},
	}
	for _, e := range seeds {
		ctx.Add(e)
	}
	if affected, err := ctx.SaveChanges(); err != nil || affected != len(seeds) {
		t.Fatalf("seed SaveChanges err=%v affected=%d", err, affected)
	}
	return seeds
}

func TestEnhancedDbSet_AdvancedQueries_WithPostgres(t *testing.T) {
	db := openPGForTest(t)
	ensureSchema(t, db)
	ctx := NewEnhancedDbContextWithDB(db)

	seeds := seedEntities(t, ctx)

	set := NewEnhancedDbSet[TestEntity](ctx)

	// WhereLike
	likeList, err := set.WhereLike("name", "Al%").ToList()
	if err != nil {
		t.Fatalf("WhereLike ToList err: %v", err)
	}
	if len(likeList) != 2 { // Alpha, Alpine
		t.Fatalf("expected 2 rows for Al%%, got %d", len(likeList))
	}

	// Chained Where + WhereOr
	// score > 75 OR age < 25 should match Alpha(80.5), Gamma(90.1), Alpine(88.8), Alpha/Alpine age<25 as well
	chainCount, err := set.Where("score > ?", 75.0).WhereOr("age < ?", 25).Count()
	if err != nil {
		t.Fatalf("chained Count err: %v", err)
	}
	if chainCount < 3 || chainCount > 5 {
		t.Fatalf("unexpected chain count: %d", chainCount)
	}

	// OrderByDescending + Skip + Take should be deterministic over scores
	ordered, err := set.OrderByDescending("score").ToList()
	if err != nil {
		t.Fatalf("order desc ToList err: %v", err)
	}
	scores := make([]float64, len(ordered))
	for i, e := range ordered {
		scores[i] = e.Score
	}
	sorted := append([]float64(nil), scores...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })
	for i := range scores {
		if scores[i] != sorted[i] {
			t.Fatalf("scores not sorted desc: %v", scores)
		}
	}
	page, err := set.OrderByDescending("score").Skip(1).Take(2).ToList()
	if err != nil || len(page) != 2 {
		t.Fatalf("paging failed: err=%v len=%d", err, len(page))
	}

	// WhereIn subset by IDs
	subsetIDs := []interface{}{seeds[0].ID, seeds[2].ID, seeds[4].ID}
	subset, err := set.WhereIn("id", subsetIDs).ToList()
	if err != nil || len(subset) != 3 {
		t.Fatalf("WhereIn subset failed: err=%v len=%d", err, len(subset))
	}

	// Any true for active ones
	anyActive, err := set.Where("active = ?", true).Any()
	if err != nil || !anyActive {
		t.Fatalf("Any active failed: err=%v any=%v", err, anyActive)
	}
}

func TestEnhancedDbSet_FirstAndSingle_WithPostgres(t *testing.T) {
	db := openPGForTest(t)
	ensureSchema(t, db)
	ctx := NewEnhancedDbContextWithDB(db)

	seedEntities(t, ctx)

	set := NewEnhancedDbSet[TestEntity](ctx)

	// Single succeeds on unique ID
	one, err := set.Where("name = ?", "Beta").Single()
	if err != nil || one == nil || one.Name != "Beta" {
		t.Fatalf("Single unique failed: err=%v val=%+v", err, one)
	}

	// Single fails on multiple
	if _, err := set.WhereLike("name", "Al%").Single(); err == nil {
		t.Fatalf("expected Single to fail on multiple results")
	}

	// First fails on no results
	if _, err := set.Where("name = ?", "Nonexistent").First(); err == nil {
		t.Fatalf("expected First to fail on no results")
	}

	// FirstOrDefault returns nil on no results
	if v, err := set.Where("name = ?", "Nonexistent").FirstOrDefault(); err != nil || v != nil {
		t.Fatalf("FirstOrDefault expected (nil,nil), got (%v,%+v)", err, v)
	}
}

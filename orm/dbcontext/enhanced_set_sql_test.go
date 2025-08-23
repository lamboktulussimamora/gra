package dbcontext

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync"
	"testing"
)

// esDriver is a minimal SQL driver that returns canned rows for select and count queries
type esDriver struct{}
type esConn struct{}
type esStmt struct {
	cols []string
	rows [][]driver.Value
}
type esRows struct {
	cols []string
	rows [][]driver.Value
	idx  int
}

func (d esDriver) Open(name string) (driver.Conn, error) { return esConn{}, nil }
func (c esConn) Prepare(query string) (driver.Stmt, error) {
	q := strings.ToUpper(query)
	// Default columns map to esUser db tags
	cols := []string{"id", "name", "email", "role", "deleted_at", "updated_at"}
	rows := [][]driver.Value{
		{int64(1), "Alice", "alice@example.com", "admin", nil, nil},
		{int64(2), "Bob", "bob@example.com", "user", nil, nil},
	}
	if strings.Contains(q, "COUNT(*)") {
		cols = []string{"count"}
		rows = [][]driver.Value{{int64(2)}}
	} else if strings.HasPrefix(q, "SELECT ") {
		// Try to extract selected columns to match returned column names
		// between SELECT and FROM
		if i := strings.Index(q, " FROM "); i > len("SELECT ") {
			selected := strings.TrimSpace(query[len("SELECT "):i]) // use original case for tags
			// normalize and split by comma
			parts := strings.Split(selected, ",")
			trimmed := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "*" && p != "" {
					trimmed = append(trimmed, p)
				}
			}
			if len(trimmed) > 0 {
				// Lowercase to match db tags used by scanRows mapping
				cols = make([]string, len(trimmed))
				for i, p := range trimmed {
					cols[i] = strings.ToLower(p)
				}
				// Project values to selected columns by index
				// Build an index map from name to position in default row
				index := map[string]int{"id": 0, "name": 1, "email": 2, "role": 3, "deleted_at": 4, "updated_at": 5}
				projected := make([][]driver.Value, 0, len(rows))
				for _, r := range rows {
					pr := make([]driver.Value, len(cols))
					for j, c := range cols {
						pr[j] = r[index[c]]
					}
					projected = append(projected, pr)
				}
				rows = projected
			}
		}
	}
	return esStmt{cols: cols, rows: rows}, nil
}
func (c esConn) Close() error                                    { return nil }
func (c esConn) Begin() (driver.Tx, error)                       { return nil, nil }
func (s esStmt) Close() error                                    { return nil }
func (s esStmt) NumInput() int                                   { return -1 }
func (s esStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s esStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &esRows{cols: s.cols, rows: s.rows}, nil
}
func (r *esRows) Columns() []string { return r.cols }
func (r *esRows) Close() error      { return nil }
func (r *esRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

// Test entity mirroring es_users table with db tags
type esUser2 struct {
	ID        int64   `db:"id"`
	Name      string  `db:"name"`
	Email     string  `db:"email"`
	Role      string  `db:"role"`
	DeletedAt *string `db:"deleted_at"`
	UpdatedAt *string `db:"updated_at"`
}

func (esUser2) TableName() string { return "es_users" }

var esDriverRegisterOnce sync.Once

func registerESDriver() {
	esDriverRegisterOnce.Do(func() {
		sql.Register("esdriver", esDriver{})
	})
}

func setupESDB(t *testing.T) *sql.DB {
	t.Helper()
	registerESDriver()
	db, err := sql.Open("esdriver", "")
	if err != nil {
		t.Fatalf("failed to open esdriver: %v", err)
	}
	// Ping via a simple prepare/query to ensure driver is wired
	stmt, err := db.PrepareContext(context.Background(), "SELECT * FROM es_users")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	_ = stmt.Close()
	return db
}

func TestEnhancedSet_ToList_DBScan(t *testing.T) {
	db := setupESDB(t)
	defer db.Close()

	ctx := &EnhancedDbContext{Database: NewDatabase(db)}
	set := NewEnhancedSet[esUser2](ctx)

	// No WHERE; select all
	results, err := set.ToList()
	if err != nil {
		t.Fatalf("ToList error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "Alice" || results[1].Name != "Bob" {
		t.Fatalf("unexpected names: %#v", results)
	}
}

func TestEnhancedSet_First_Single(t *testing.T) {
	db := setupESDB(t)
	defer db.Close()

	ctx := &EnhancedDbContext{Database: NewDatabase(db)}
	set := NewEnhancedSet[esUser2](ctx)

	// First should return first row
	first, err := set.First()
	if err != nil {
		t.Fatalf("First error: %v", err)
	}
	if first.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", first.Name)
	}

	// Single should error because driver returns 2 rows
	if _, err := set.Single(); err == nil {
		t.Fatalf("expected error from Single with multiple rows")
	}
}

func TestEnhancedSet_Count_Any(t *testing.T) {
	db := setupESDB(t)
	defer db.Close()

	ctx := &EnhancedDbContext{Database: NewDatabase(db)}
	set := NewEnhancedSet[esUser2](ctx)

	cnt, err := set.Count()
	if err != nil {
		t.Fatalf("Count error: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected count 2, got %d", cnt)
	}

	any, err := set.Any()
	if err != nil {
		t.Fatalf("Any error: %v", err)
	}
	if !any {
		t.Fatalf("expected Any to be true")
	}
}

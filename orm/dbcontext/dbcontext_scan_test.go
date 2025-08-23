package dbcontext

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"testing"
)

// minimal in-memory driver to produce a single row for scanEntity tests
type singleRowDriver struct{}
type singleConn struct{}
type singleStmt struct {
	cols []string
	vals []driver.Value
}
type singleRows struct {
	cols []string
	vals []driver.Value
	done bool
}

func (d singleRowDriver) Open(name string) (driver.Conn, error) { return singleConn{}, nil }
func (c singleConn) Prepare(query string) (driver.Stmt, error) {
	return singleStmt{cols: []string{"id", "name", "age"}, vals: []driver.Value{int64(7), "Alice", int64(30)}}, nil
}
func (c singleConn) Close() error                                    { return nil }
func (c singleConn) Begin() (driver.Tx, error)                       { return nil, nil }
func (s singleStmt) Close() error                                    { return nil }
func (s singleStmt) NumInput() int                                   { return -1 }
func (s singleStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s singleStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &singleRows{cols: s.cols, vals: s.vals}, nil
}
func (r *singleRows) Columns() []string { return r.cols }
func (r *singleRows) Close() error      { return nil }
func (r *singleRows) Next(dest []driver.Value) error {
	if r.done {
		return driver.ErrBadConn
	}
	copy(dest, r.vals)
	r.done = true
	return nil
}

func TestScanEntity_SetsFields(t *testing.T) {
	sql.Register("single", singleRowDriver{})
	db, err := sql.Open("single", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Prepare and query to get Rows
	stmt, err := db.PrepareContext(context.Background(), "select 1")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	// Move to the first row before scanning
	if !rows.Next() {
		t.Fatal("expected one row")
	}

	type E struct {
		ID   int64
		Name string
		Age  int
	}
	var e E
	if err := scanEntity(rows, &e); err != nil {
		t.Fatalf("scanEntity error: %v", err)
	}
	if e.ID != 7 || e.Name != "Alice" || e.Age != 30 {
		t.Fatalf("unexpected entity values: %+v", e)
	}
}

func TestSetFieldValue_StringAndInt(t *testing.T) {
	type S struct {
		A string
		B int
	}
	var s S
	rv := reflect.ValueOf(&s).Elem()
	if err := setFieldValue(rv.FieldByName("A"), []byte("hi")); err != nil {
		t.Fatal(err)
	}
	if s.A != "hi" {
		t.Fatalf("expected hi, got %q", s.A)
	}
	if err := setFieldValue(rv.FieldByName("B"), "42"); err != nil {
		t.Fatal(err)
	}
	if s.B != 42 {
		t.Fatalf("expected 42, got %d", s.B)
	}
}

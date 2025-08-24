package dbcontext

import "testing"

func TestConvertQueryPlaceholders_Postgres(t *testing.T) {
	in := "SELECT * FROM users WHERE id = ? AND name = ? AND age > ?"
	out := convertQueryPlaceholders(in, "postgres")
	exp := "SELECT * FROM users WHERE id = $1 AND name = $2 AND age > $3"
	if out != exp {
		t.Fatalf("expected %q, got %q", exp, out)
	}
}

func TestConvertQueryPlaceholders_OtherDrivers(t *testing.T) {
	in := "INSERT INTO t(a,b) VALUES(?, ?)"
	if out := convertQueryPlaceholders(in, "sqlite3"); out != in {
		t.Fatalf("sqlite3 should remain unchanged")
	}
	if out := convertQueryPlaceholders(in, "mysql"); out != in {
		t.Fatalf("mysql should remain unchanged")
	}
}

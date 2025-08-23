package schema

import "testing"

func TestDetectDatabaseDriverFromConnectionString_Short(t *testing.T) {
	if got := DetectDatabaseDriverFromConnectionString("postgres"); got != PostgreSQL {
		t.Fatalf("expected PostgreSQL, got %v", got)
	}
	if got := DetectDatabaseDriverFromConnectionString("sqlite"); got != SQLite {
		t.Fatalf("expected SQLite, got %v", got)
	}
	if got := DetectDatabaseDriverFromConnectionString("mysql"); got != MySQL {
		t.Fatalf("expected MySQL, got %v", got)
	}
	if got := DetectDatabaseDriverFromConnectionString("unknown"); got != PostgreSQL {
		t.Fatalf("expected default PostgreSQL, got %v", got)
	}
}

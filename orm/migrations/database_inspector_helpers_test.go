package migrations

import "testing"

func TestParseIntValue(t *testing.T) {
	di := &DatabaseInspector{}
	if di.parseIntValue("") != 0 {
		t.Fatalf("empty string should return 0")
	}
	if di.parseIntValue("123") != 123 {
		t.Fatalf("expected 123")
	}
	if di.parseIntValue("12a3") != 0 {
		t.Fatalf("invalid chars should return 0")
	}
}

func TestParseSQLiteDataType_LengthAndPrecision(t *testing.T) {
	di := &DatabaseInspector{}
	col := &DatabaseColumnInfo{Name: "name"}
	di.parseSQLiteDataType(col, "VARCHAR(255)")
	if col.MaxLength == nil || *col.MaxLength != 255 {
		t.Fatalf("expected MaxLength=255, got %+v", col.MaxLength)
	}

	col2 := &DatabaseColumnInfo{Name: "price"}
	di.parseSQLiteDataType(col2, "DECIMAL(10,2)")
	if col2.Precision == nil || *col2.Precision != 10 || col2.Scale == nil || *col2.Scale != 2 {
		t.Fatalf("expected Precision=10 Scale=2, got p=%v s=%v", col2.Precision, col2.Scale)
	}

	col3 := &DatabaseColumnInfo{Name: "weird"}
	di.parseSQLiteDataType(col3, "NUMERIC(abc)")
	if col3.Precision != nil || col3.Scale != nil {
		t.Fatalf("invalid numeric params should be ignored, got p=%v s=%v", col3.Precision, col3.Scale)
	}
}

func TestIsSystemTable(t *testing.T) {
	di := &DatabaseInspector{}
	sys := []string{"__migration_history", "sqlite_master", "schema_migrations"}
	for _, name := range sys {
		if !di.isSystemTable(name) {
			t.Fatalf("expected %s to be system table", name)
		}
	}
	if di.isSystemTable("users") {
		t.Fatalf("users should not be system table")
	}
}

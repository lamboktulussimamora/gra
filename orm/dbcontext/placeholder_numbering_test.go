package dbcontext

import "testing"

// dummy entity type for set construction without hitting DB
type pnDummy struct{}

func (pnDummy) TableName() string { return "pn_dummy" }

func TestAdjustPlaceholders_Postgres_WhereChains(t *testing.T) {
	ctx := &EnhancedDbContext{driver: driverPostgres}
	set := &EnhancedDbSet[pnDummy]{
		ctx:       ctx,
		tableName: "pn_dummy",
	}

	// Chain multiple conditions to verify placeholder numbering continuity
	s1 := set.Where("a = ? AND b > ?", 1, 2)
	if got := s1.whereClause; got != "a = $1 AND b > $2" {
		t.Fatalf("unexpected where after first Where: %q", got)
	}

	s2 := s1.WhereOr("c IN (?, ?, ?)", 3, 4, 5)
	if got := s2.whereClause; got != "a = $1 AND b > $2 OR (c IN ($3, $4, $5))" {
		t.Fatalf("unexpected where after WhereOr: %q", got)
	}

	// WhereIn appends and should continue numbering
	s3 := s2.WhereIn("d", []interface{}{6, 7})
	if got := s3.whereClause; got != "a = $1 AND b > $2 OR (c IN ($3, $4, $5)) AND d IN ($6, $7)" {
		t.Fatalf("unexpected where after WhereIn: %q", got)
	}
}

// Test getFieldData placeholder numbering and embedded struct traversal for Postgres
func TestGetFieldData_Postgres_WithEmbedded(t *testing.T) {
	type Embedded struct {
		Street string
		Zip    int
	}
	type Entity struct {
		ID   int64
		Name string
		Embedded
		Age int
	}

	e := &Entity{ID: 10, Name: "x", Embedded: Embedded{Street: "s", Zip: 12345}, Age: 30}

	cols, vals, ph := getFieldData(e, true, driverPostgres) // exclude ID
	if len(cols) != len(vals) || len(vals) != len(ph) {
		t.Fatalf("length mismatch cols=%d vals=%d ph=%d", len(cols), len(vals), len(ph))
	}
	// Expect Name, Street, Zip, Age (order by struct fields with embedded expanded)
	if len(ph) != 4 {
		t.Fatalf("expected 4 placeholders, got %d", len(ph))
	}
	for i, p := range ph {
		want := "$" + itoa(i+1)
		if p != want {
			t.Fatalf("placeholder %d: want %s, got %s", i+1, want, p)
		}
	}

	// Verify ID is excluded when excludeID=true
	for _, c := range cols {
		if c == "id" {
			t.Fatalf("did not expect id column when excludeID=true")
		}
	}

	// Now include ID and ensure column count increased by one
	cols2, _, _ := getFieldData(e, false, driverPostgres)
	if len(cols2) != len(cols)+1 {
		t.Fatalf("expected one more column when excludeID=false, got %d vs %d", len(cols2), len(cols))
	}
}

// simple integer to string without importing strconv to keep test small
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

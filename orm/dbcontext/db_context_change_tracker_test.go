package dbcontext

import (
	"testing"
)

// Simple entity with TableName override to exercise getTableName path
type ctEntity struct {
	ID   int64
	Name string
}

func (ctEntity) TableName() string { return "ct_entities" }

func TestChangeTracker_BasicLifecycle(t *testing.T) {
	ct := NewChangeTracker()
	e := &ctEntity{Name: "n"}

	// default state is Unchanged
	if got := ct.GetEntityState(e); got != EntityStateUnchanged {
		t.Fatalf("default state should be Unchanged, got %v", got)
	}

	// Track and then change states
	ct.TrackEntity(e, EntityStateAdded)
	if got := ct.GetEntityState(e); got != EntityStateAdded {
		t.Fatalf("expected Added state, got %v", got)
	}
	ct.SetEntityState(e, EntityStateModified)
	if got := ct.GetEntityState(e); got != EntityStateModified {
		t.Fatalf("expected Modified state, got %v", got)
	}
	// ensure String() has a stable value for one of them
	if EntityStateDeleted.String() == "" {
		t.Fatalf("String() should not be empty")
	}
}

func TestEnhancedDbSet_PostgresPlaceholderNumbering(t *testing.T) {
	// Create a context with postgres driver without touching a real DB
	ctx := &EnhancedDbContext{driver: driverPostgres}

	type item struct {
		ID   int64
		Name string
	}
	set := NewEnhancedDbSet[item](ctx)

	// Build a complex where with multiple calls to ensure numbering sequences correctly
	ids := []interface{}{int64(1), int64(2)}
	q := set.Where("name = ?", "alice").WhereIn("id", ids).WhereOr("id > ?", 10).OrderBy("id").buildQuery()

	// Expect: $1 from first Where, $2,$3 from WhereIn, $4 from WhereOr
	want := "SELECT * FROM item WHERE name = $1 AND id IN ($2, $3) OR (id > $4) ORDER BY id"
	if q != want {
		t.Fatalf("unexpected postgres placeholder numbering:\nwant: %s\n got: %s", want, q)
	}
}

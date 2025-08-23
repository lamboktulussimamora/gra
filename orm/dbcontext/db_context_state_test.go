package dbcontext

import (
	"testing"
)

// entity used for AsNoTracking integration
type asNoTrackEntity struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (asNoTrackEntity) TableName() string { return "as_no_track" }

func TestEntityState_String(t *testing.T) {
	cases := map[EntityState]string{
		EntityStateUnchanged: "Unchanged",
		EntityStateAdded:     "Added",
		EntityStateModified:  "Modified",
		EntityStateDeleted:   "Deleted",
		EntityState(99):      "Unknown",
	}
	for s, exp := range cases {
		if got := s.String(); got != exp {
			t.Fatalf("%v.String()=%q, want %q", s, got, exp)
		}
	}
}

func TestChangeTracker_Basics(t *testing.T) {
	ct := NewChangeTracker()
	type e struct{ ID int }
	x := &e{ID: 1}

	if st := ct.GetEntityState(x); st != EntityStateUnchanged {
		t.Fatalf("new entity should be Unchanged, got %v", st)
	}
	ct.SetEntityState(x, EntityStateAdded)
	if st := ct.GetEntityState(x); st != EntityStateAdded {
		t.Fatalf("expected Added, got %v", st)
	}
	ct.TrackEntity(x, EntityStateModified)
	if st := ct.GetEntityState(x); st != EntityStateModified {
		t.Fatalf("expected Modified, got %v", st)
	}
}

func TestDatabase_Begin_Postgres(t *testing.T) {
	db := openPGForTest(t)
	d := NewDatabase(db)
	tx, err := d.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	_ = tx.Rollback()
}

func TestAsNoTracking_Effect(t *testing.T) {
	db := openPGForTest(t)
	// ensure table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS as_no_track (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL
    )`)
	if err != nil {
		t.Fatalf("ensure table: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DROP TABLE IF EXISTS as_no_track") })
	_, _ = db.Exec("DELETE FROM as_no_track")
	_, _ = db.Exec("INSERT INTO as_no_track(name) VALUES ('n1'), ('n2')")

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[asNoTrackEntity](ctx)

	// Without AsNoTracking should track results
	_, err = set.ToList()
	if err != nil {
		t.Fatalf("ToList: %v", err)
	}
	if len(ctx.ChangeTracker.entities) == 0 {
		t.Fatalf("expected tracked entities > 0 without AsNoTracking")
	}

	// With AsNoTracking should not add new tracked entities
	before := len(ctx.ChangeTracker.entities)
	_, err = set.AsNoTracking().ToList()
	if err != nil {
		t.Fatalf("ToList AsNoTracking: %v", err)
	}
	after := len(ctx.ChangeTracker.entities)
	if after != before {
		t.Fatalf("expected no new tracked entities with AsNoTracking, before=%d after=%d", before, after)
	}
}

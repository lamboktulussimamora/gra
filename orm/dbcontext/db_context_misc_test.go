package dbcontext

import (
	"testing"
)

func TestEnhancedDbSet_AdjustAndFind_NoTrackingFlag(t *testing.T) {
	// Minimal in-memory context with Postgres driver to test placeholder conversion
	ctx := &EnhancedDbContext{driver: driverPostgres, ChangeTracker: NewChangeTracker()}
	set := &EnhancedDbSet[struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}]{ctx: ctx, tableName: "t"}
	// Ensure adjustPlaceholdersForCondition increments from existing arg count
	set.whereArgs = []interface{}{1}
	cond := set.adjustPlaceholdersForCondition("name = ? AND id = ?")
	if cond != "name = $2 AND id = $3" {
		t.Fatalf("unexpected converted condition: %s", cond)
	}
	// AsNoTracking should set flag on cloned set
	set2 := set.AsNoTracking()
	if !set2.noTracking {
		t.Fatalf("AsNoTracking should set noTracking=true")
	}
}

func TestNewEnhancedDbContextWithTx_Defaults(t *testing.T) {
	ctx := NewEnhancedDbContextWithTx(nil)
	if ctx.driver != "sqlite3" { // default noted in constructor comment
		t.Fatalf("expected default driver sqlite3 for tx ctor, got %s", ctx.driver)
	}
	// ChangeTracker default state
	var e struct{}
	if st := ctx.ChangeTracker.GetEntityState(&e); st != EntityStateUnchanged {
		t.Fatalf("default entity state should be Unchanged")
	}
}

package migrations

import (
	"strings"
	"testing"
)

// helper to create pointer ints
func pint(v int) *int { return &v }

func TestChangeDetector_calculatePlanChecksum_StableAcrossOrder(t *testing.T) {
	cd := &ChangeDetector{}
	// Same logical changes but different order
	a := []MigrationChange{
		{Type: AddColumn, TableName: "users", ColumnName: "name"},
		{Type: CreateTable, TableName: "users"},
		{Type: DropIndex, TableName: "posts", IndexName: "ix_posts_user_id"},
	}
	b := []MigrationChange{
		{Type: DropIndex, TableName: "posts", IndexName: "ix_posts_user_id"},
		{Type: CreateTable, TableName: "users"},
		{Type: AddColumn, TableName: "users", ColumnName: "name"},
	}

	ca := cd.calculatePlanChecksum(a)
	cb := cd.calculatePlanChecksum(b)

	if ca != cb {
		t.Fatalf("expected stable checksum, got %s vs %s", ca, cb)
	}
}

func TestChangeDetector_getChangeTypePriority_Order(t *testing.T) {
	cd := &ChangeDetector{}
	if cd.getChangeTypePriority(CreateTable) >= cd.getChangeTypePriority(DropTable) {
		t.Fatalf("expected CreateTable to have higher priority than DropTable")
	}
	if cd.getChangeTypePriority(AddColumn) >= cd.getChangeTypePriority(AlterColumn) {
		t.Fatalf("expected AddColumn to have higher priority than AlterColumn")
	}
}

func TestChangeDetector_hasDestructiveAndRequiresReview(t *testing.T) {
	cd := &ChangeDetector{}
	nonDestructive := []MigrationChange{{Type: AddColumn}}
	destructive := []MigrationChange{{Type: DropColumn}}

	if cd.hasDestructiveChanges(nonDestructive) {
		t.Fatalf("unexpected destructive detection")
	}
	if !cd.hasDestructiveChanges(destructive) {
		t.Fatalf("expected destructive detection")
	}
	if !cd.requiresManualReview(destructive) {
		t.Fatalf("expected requires manual review for DropColumn")
	}
}

func TestChangeDetector_isDataLosingAlterColumn(t *testing.T) {
	cd := &ChangeDetector{}
	// Case 1: nullable -> not null
	oldCol := &DatabaseColumnInfo{IsNullable: true, DataType: "VARCHAR", MaxLength: pint(100)}
	newCol := &ColumnInfo{IsNullable: false, DataType: "VARCHAR", MaxLength: pint(100)}
	ch := MigrationChange{Type: AlterColumn, OldValue: oldCol, NewValue: newCol}
	if !cd.isDataLosingAlterColumn(ch) {
		t.Fatalf("expected data losing change for nullable->not null")
	}

	// Case 2: shrink max length
	oldCol = &DatabaseColumnInfo{IsNullable: false, DataType: "VARCHAR", MaxLength: pint(100)}
	newCol = &ColumnInfo{IsNullable: false, DataType: "VARCHAR", MaxLength: pint(50)}
	ch = MigrationChange{Type: AlterColumn, OldValue: oldCol, NewValue: newCol}
	if !cd.isDataLosingAlterColumn(ch) {
		t.Fatalf("expected data losing change for length shrink")
	}

	// Case 3: incompatible type change
	oldCol = &DatabaseColumnInfo{IsNullable: false, DataType: "TEXT"}
	newCol = &ColumnInfo{IsNullable: false, DataType: "INTEGER"}
	ch = MigrationChange{Type: AlterColumn, OldValue: oldCol, NewValue: newCol}
	if !cd.isDataLosingAlterColumn(ch) {
		t.Fatalf("expected data losing change for incompatible type change")
	}
}

func TestChangeDetector_isIncompatibleTypeChange(t *testing.T) {
	cd := &ChangeDetector{}
	cases := []struct {
		oldT, newT   string
		incompatible bool
	}{
		{"TEXT", "INTEGER", true},
		{"VARCHAR(255)", "BOOLEAN", true},
		{"INTEGER", "VARCHAR(10)", false},
		{"BOOLEAN", "INTEGER", true},
		{"DATE", "TIMESTAMP", false},
	}
	for _, c := range cases {
		got := cd.isIncompatibleTypeChange(c.oldT, c.newT)
		if got != c.incompatible {
			t.Fatalf("mismatch for %s -> %s: got %v", c.oldT, c.newT, got)
		}
	}
}

func TestChangeDetector_sortChangesByDependency(t *testing.T) {
	cd := &ChangeDetector{}
	changes := []MigrationChange{
		{Type: DropColumn, TableName: "users", ColumnName: "x"},
		{Type: CreateTable, TableName: "users"},
		{Type: AddColumn, TableName: "users", ColumnName: "y"},
	}
	cd.sortChangesByDependency(changes)
	// Expect CreateTable first, then AddColumn, then DropColumn
	if changes[0].Type != CreateTable || changes[1].Type != AddColumn || changes[2].Type != DropColumn {
		t.Fatalf("unexpected order after sort: %+v", changes)
	}
}

func TestChangeDetector_changeToString(t *testing.T) {
	cd := &ChangeDetector{}
	s := cd.changeToString(MigrationChange{Type: AddIndex, TableName: "t", ModelName: "m", ColumnName: "c", IndexName: "i"})
	if strings.Count(s, "|") != 4 {
		t.Fatalf("unexpected string format: %s", s)
	}
}

func TestValidateMigrationPlan_Warnings_NoErrors(t *testing.T) {
	cd := &ChangeDetector{}
	// Drop a table referenced by a FK -> orphaned FK warning
	snap := &ModelSnapshot{TableName: "posts", Constraints: map[string]*ConstraintInfo{
		"fk_posts_user": {Name: "fk_posts_user", Type: foreignKeyConstraintType, ReferencedTable: "users"},
	}}
	changes := []MigrationChange{
		{Type: DropTable, TableName: "users"},
		{Type: AddColumn, TableName: "posts", NewValue: &ColumnInfo{Constraints: map[string]*ConstraintInfo{
			"fk_posts_user": {Name: "fk_posts_user", Type: foreignKeyConstraintType, ReferencedTable: "users"},
		}}},
		{Type: CreateTable, TableName: "posts", NewValue: snap},
		{Type: DropColumn, TableName: "posts", ColumnName: "old"},
	}
	plan := &MigrationPlan{Changes: changes}
	err := cd.ValidateMigrationPlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Warnings) == 0 {
		t.Fatalf("expected warnings, got none")
	}
	// Ensure at least one warning mentions foreign key or data loss
	hasRelevant := false
	for _, w := range plan.Warnings {
		if strings.Contains(w, "Foreign key") || strings.Contains(w, "Potential data loss") {
			hasRelevant = true
			break
		}
	}
	if !hasRelevant {
		t.Fatalf("expected relevant warnings, got: %v", plan.Warnings)
	}
}

func TestValidateMigrationPlan_CircularDependency(t *testing.T) {
	cd := &ChangeDetector{}
	// users references posts and posts references users -> circular
	users := &ModelSnapshot{TableName: "users", Constraints: map[string]*ConstraintInfo{
		"fk_users_posts": {Name: "fk_users_posts", Type: foreignKeyConstraintType, ReferencedTable: "posts"},
	}}
	posts := &ModelSnapshot{TableName: "posts", Constraints: map[string]*ConstraintInfo{
		"fk_posts_users": {Name: "fk_posts_users", Type: foreignKeyConstraintType, ReferencedTable: "users"},
	}}
	plan := &MigrationPlan{Changes: []MigrationChange{
		{Type: CreateTable, TableName: "users", NewValue: users},
		{Type: CreateTable, TableName: "posts", NewValue: posts},
	}}
	err := cd.ValidateMigrationPlan(plan)
	if err == nil {
		t.Fatalf("expected circular dependency error, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetChangeSummary(t *testing.T) {
	cd := &ChangeDetector{}
	plan := &MigrationPlan{Changes: []MigrationChange{
		{Type: CreateTable}, {Type: CreateTable}, {Type: DropTable}, {Type: AddColumn}, {Type: DropColumn}, {Type: AlterColumn}, {Type: CreateIndex}, {Type: DropIndex},
	}}
	plan.HasDestructive = cd.hasDestructiveChanges(plan.Changes)
	plan.RequiresReview = cd.requiresManualReview(plan.Changes)
	summary := cd.GetChangeSummary(plan)
	for _, want := range []string{"2 table(s) to create", "1 table(s) to drop", "1 column(s) to add", "1 column(s) to drop", "1 column(s) to alter", "1 index(es) to create", "1 index(es) to drop"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
	if !strings.Contains(summary, "destructive") || !strings.Contains(summary, "requires manual review") {
		t.Fatalf("expected flags in summary, got: %s", summary)
	}
}

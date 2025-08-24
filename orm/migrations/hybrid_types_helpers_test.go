package migrations

import "testing"

func TestMigrationFile_FlagsAndWarnings(t *testing.T) {
	mf := &MigrationFile{}
	if mf.HasDestructiveChanges() {
		t.Fatalf("expected no destructive by default")
	}
	if mf.RequiresReview() {
		t.Fatalf("expected no review by default")
	}
	if len(mf.Warnings()) != 0 || len(mf.Errors()) != 0 {
		t.Fatalf("expected no warnings/errors by default")
	}

	destructive := true
	mf.ParsedHasDestructive = &destructive
	if !mf.HasDestructiveChanges() {
		t.Fatalf("parsed destructive flag should be honored when no Changes present")
	}

	// With Changes override ParsedHasDestructive
	mf.Changes = []MigrationChange{{IsDestructive: false}, {IsDestructive: true}}
	mf.ParsedHasDestructive = nil
	if !mf.HasDestructiveChanges() || !mf.RequiresReview() {
		t.Fatalf("destructive Changes should be reported")
	}
	ws := mf.Warnings()
	if len(ws) == 0 {
		t.Fatalf("expected warnings for destructive changes")
	}
}

func TestMigrationMode_ParseAndString(t *testing.T) {
	cases := []struct {
		in  string
		out MigrationMode
	}{
		{ModeAutomatic.String(), ModeAutomatic},
		{"Interactive", ModeInteractive},
		{ModeGenerateOnly.String(), ModeGenerateOnly},
		{"ForceDestructive", ModeForceDestructive},
		{"Unknown", ModeAutomatic},
	}
	for _, c := range cases {
		if got := ParseMigrationMode(c.in); got != c.out {
			t.Fatalf("ParseMigrationMode %s => %v", c.in, got)
		}
	}
	const (
		strAutomatic        = "Automatic"
		strForceDestructive = "ForceDestructive"
	)
	if ModeAutomatic.String() != strAutomatic || ModeForceDestructive.String() != strForceDestructive {
		t.Fatalf("unexpected String() output")
	}
}

func TestIsDataTypeCompatible(t *testing.T) {
	di := &DatabaseInspector{}
	pairs := [][2]string{
		{"VARCHAR", "CHARACTER VARYING"},
		{"TEXT", "VARCHAR"},
		{"INTEGER", "SERIAL"},
		{"BIGINT", "BIGSERIAL"},
		{"BOOLEAN", "BOOL"},
		{"TIMESTAMP", "TIMESTAMPTZ"},
		{"DECIMAL", "NUMERIC"},
	}
	for _, p := range pairs {
		if !di.isDataTypeCompatible(p[0], p[1]) {
			t.Fatalf("expected %s compatible with %s", p[0], p[1])
		}
	}
	if di.isDataTypeCompatible("INTEGER", "TEXT") {
		t.Fatalf("unexpected compatibility between INTEGER and TEXT")
	}
}

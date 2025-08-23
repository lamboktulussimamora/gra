package dbcontext

import (
	"reflect"
	"testing"
	"time"
)

// Test entity types for EFContext helper coverage
type efTestEntity struct {
	BaseEntity
	Name   string `db:"name" json:"name"`
	Age    int    `db:"age" json:"age"`
	Secret string `db:"-" json:"-"`
	// No tags -> falls back to snake_case
	DisplayName string
}

func TestToSnakeCaseEF(t *testing.T) {
	ctx := &EFContext{}
	cases := map[string]string{
		"Simple":           "simple",
		"UserID":           "user_i_d", // current simplistic behavior splits each uppercase
		"HTTPServerConfig": "h_t_t_p_server_config",
	}
	for in, want := range cases {
		if got := ctx.toSnakeCaseEF(in); got != want {
			t.Fatalf("toSnakeCaseEF(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestGetTableNameFromType(t *testing.T) {
	ctx := &EFContext{}
	var e efTestEntity
	got := ctx.getTableNameFromType(reflect.TypeOf(e))
	if want := "ef_test_entitys"; got != want { // pluralization is simple +"s"
		t.Fatalf("getTableNameFromType=%q, want %q", got, want)
	}
}

func TestGetColumnNameFromFieldPriority(t *testing.T) {
	ctx := &EFContext{}
	typ := reflect.TypeOf(efTestEntity{})

	// db tag wins
	fName, _ := typ.FieldByName("Name")
	if got := ctx.getColumnNameFromField(fName); got != "name" {
		t.Fatalf("db tag expected 'name', got %q", got)
	}

	// falls back to snake_case when no tags
	fDisp, _ := typ.FieldByName("DisplayName")
	if got := ctx.getColumnNameFromField(fDisp); got != "display_name" {
		t.Fatalf("snake_case expected 'display_name', got %q", got)
	}
}

func TestShouldSkipField_WithIgnoreTags(t *testing.T) {
	ctx := &EFContext{}
	typ := reflect.TypeOf(efTestEntity{})
	fSecret, _ := typ.FieldByName("Secret")
	if got := ctx.shouldSkipField(fSecret); got != true {
		t.Fatalf("expected Secret to be skipped due to db:\"-\" tag")
	}
}

func TestExtractFieldsForDebug_AndTimestamps(t *testing.T) {
	ctx := &EFContext{}
	e := &efTestEntity{Name: "alice", Age: 30}

	cols, vals := ctx.ExtractFieldsForDebug(e)
	// ID should be excluded on insert path; created_at/updated_at included (from BaseEntity)
	// DisplayName should appear as snake_case
	// Order depends on reflection traversal; ensure presence instead of exact order.
	wantCols := map[string]bool{"name": true, "age": true, "created_at": true, "updated_at": true, "display_name": true}
	if len(cols) == 0 || len(vals) == 0 {
		t.Fatalf("expected non-empty columns and values")
	}
	for _, c := range cols {
		delete(wantCols, c)
	}
	if len(wantCols) != 0 {
		t.Fatalf("missing expected columns: %v", wantCols)
	}

	// setTimestamps should set CreatedAt on insert and always set UpdatedAt
	// Work on the underlying value (non-pointer) per implementation
	rv := reflect.ValueOf(e).Elem()
	// Initially zero
	if !rv.FieldByName("CreatedAt").Interface().(time.Time).IsZero() {
		t.Fatalf("CreatedAt should start zero in this test")
	}
	ctx.setTimestamps(rv, true)
	if rv.FieldByName("CreatedAt").Interface().(time.Time).IsZero() {
		t.Fatalf("CreatedAt should be set on insert")
	}
	if rv.FieldByName("UpdatedAt").Interface().(time.Time).IsZero() {
		t.Fatalf("UpdatedAt should be set")
	}

	// Subsequent non-insert should not zero-out CreatedAt
	created := rv.FieldByName("CreatedAt").Interface().(time.Time)
	ctx.setTimestamps(rv, false)
	if got := rv.FieldByName("CreatedAt").Interface().(time.Time); !got.Equal(created) {
		t.Fatalf("CreatedAt should remain unchanged on non-insert")
	}
}

// Additional tests

// Local types for testing EFContext helpers
type efUser struct {
	BaseEntity
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	// Ignored fields via tags
	SkipDB   string `db:"-"`
	SkipJSON string `json:"-"`
}

func (efUser) TableName() string { return "ef_users" }

func TestEF_toSnakeCaseEF(t *testing.T) {
	cases := map[string]string{
		"User":        "user",
		"UserProfile": "user_profile",
		"HTTPServer":  "h_t_t_p_server", // simplistic algorithm behavior
		"simple":      "simple",
		"X":           "x",
	}
	var ctx EFContext
	for in, exp := range cases {
		if got := ctx.toSnakeCaseEF(in); got != exp {
			t.Fatalf("toSnakeCaseEF(%q) = %q, want %q", in, got, exp)
		}
	}
}

func TestEF_getTableAndColumnAndSkipRules(t *testing.T) {
	var ctx EFContext

	// Table name pluralization from type
	type UserProfile struct{ BaseEntity }
	tn := ctx.getTableNameFromType(reflect.TypeOf(UserProfile{}))
	if tn != "user_profiles" {
		t.Fatalf("table name expected user_profiles, got %s", tn)
	}

	// Column name from db tag, then json tag, then snake of field
	fieldDB, _ := reflect.TypeOf(efUser{}).FieldByName("Name")
	if col := ctx.getColumnNameFromField(fieldDB); col != "name" {
		t.Fatalf("db tag should win, got %s", col)
	}
	fieldJSON, _ := reflect.TypeOf(struct {
		BaseEntity
		DisplayName string `json:"display_name"`
	}{}).FieldByName("DisplayName")
	if col := ctx.getColumnNameFromField(fieldJSON); col != "display_name" {
		t.Fatalf("json tag should be used if db tag absent, got %s", col)
	}
	fieldDefault, _ := reflect.TypeOf(struct{ FooBar int }{}).FieldByName("FooBar")
	if col := ctx.getColumnNameFromField(fieldDefault); col != "foo_bar" {
		t.Fatalf("default snake case expected foo_bar, got %s", col)
	}

	// Skip rules: db:"-" or json:"-"
	fieldSkipDB, _ := reflect.TypeOf(efUser{}).FieldByName("SkipDB")
	if !ctx.shouldSkipField(fieldSkipDB) {
		t.Fatalf("field with db:\"-\" should be skipped")
	}
	fieldSkipJSON, _ := reflect.TypeOf(efUser{}).FieldByName("SkipJSON")
	if !ctx.shouldSkipField(fieldSkipJSON) {
		t.Fatalf("field with json:\"-\" should be skipped")
	}
}

func TestEF_ExtractFieldsForDebug_SkipsIDAndEmbeddedAndIgnored(t *testing.T) {
	var ctx EFContext
	u := &efUser{BaseEntity: BaseEntity{ID: 123}, Name: "Alice", Email: "a@example.com", SkipDB: "x", SkipJSON: "y"}
	cols, vals := ctx.ExtractFieldsForDebug(u)

	// ID should be excluded, SkipDB and SkipJSON excluded, created_at/updated_at included via embedded BaseEntity.
	// Order depends on struct layout; verify presence and matching values by column lookup.
	if len(cols) != len(vals) {
		t.Fatalf("cols and vals length mismatch: %d vs %d", len(cols), len(vals))
	}
	// Expect exactly these columns (unordered)
	expSet := map[string]bool{"name": true, "email": true, "created_at": true, "updated_at": true}
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(cols), cols)
	}
	for _, c := range cols {
		if !expSet[c] {
			t.Fatalf("unexpected column %q in %v", c, cols)
		}
	}
	// Validate that the values for name/email match via index lookup
	var nameIdx, emailIdx = -1, -1
	for i, c := range cols {
		if c == "name" {
			nameIdx = i
		}
		if c == "email" {
			emailIdx = i
		}
	}
	if nameIdx < 0 || emailIdx < 0 {
		t.Fatalf("missing name or email columns: %v", cols)
	}
	if vals[nameIdx] != "Alice" || vals[emailIdx] != "a@example.com" {
		t.Fatalf("unexpected name/email values: %v", vals)
	}
}

func TestEF_setTimestamps_CreateAndUpdate(t *testing.T) {
	// Use EFContext helpers directly to mutate embedded fields
	ent := &efUser{}
	v := reflect.ValueOf(ent).Elem()
	var ctx EFContext

	// create
	ctx.setTimestamps(v, true)
	if ent.CreatedAt.IsZero() || ent.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not set on create")
	}
	created := ent.CreatedAt
	time.Sleep(time.Millisecond)
	// update
	ctx.setTimestamps(v, false)
	if !ent.CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt should remain unchanged on update")
	}
	if !ent.UpdatedAt.After(created) {
		t.Fatalf("UpdatedAt should be refreshed on update")
	}
}

func TestEFContext_DBGuards(t *testing.T) {
	ctx := &EFContext{db: nil}
	u := &efUser{Name: "Bob", Email: "b@example.com"}

	if err := ctx.Add(u); err == nil || err.Error() != "database connection is nil" {
		t.Fatalf("Add should error on nil db, got %v", err)
	}
	if err := ctx.Update(u); err == nil || err.Error() != "database connection is nil" {
		t.Fatalf("Update should error on nil db, got %v", err)
	}
	if err := ctx.Remove(u); err == nil || err.Error() != "database connection is nil" {
		t.Fatalf("Remove should error on nil db, got %v", err)
	}
	if err := ctx.Find(u, 1); err == nil || err.Error() != "database connection is nil" {
		t.Fatalf("Find should error on nil db, got %v", err)
	}
	// SaveChanges is a no-op
	if err := ctx.SaveChanges(); err != nil {
		t.Fatalf("SaveChanges should be nil, got %v", err)
	}
}

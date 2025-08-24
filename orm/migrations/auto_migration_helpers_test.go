package migrations

import (
	"reflect"
	"strings"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/schema"
)

// Model with TableName method
type tableNamed struct {
	ID int `db:"id"`
}

func (tableNamed) TableName() string { return "custom_tbl" }

// Embedded structs for field processing tests
type embeddedInner struct {
	Name string `db:"name"`
}

type embeddedOuter struct {
	ID int `db:"id"`
	embeddedInner
	Ignored string `db:"-"`
}

func TestAutoMigrator_toSnakeCase(t *testing.T) {
	am := &AutoMigrator{}
	if got := am.toSnakeCase("UserProfileID"); got != "user_profile_i_d" {
		t.Fatalf("toSnakeCase got %q", got)
	}
}

func TestAutoMigrator_getTableName(t *testing.T) {
	am := &AutoMigrator{}
	// With TableName()
	if got := am.getTableName(tableNamed{}); got != "custom_tbl" {
		t.Fatalf("expected custom_tbl, got %q", got)
	}
	// Without TableName()
	type UserProfile struct{ ID int }
	if got := am.getTableName(UserProfile{}); got != "user_profile" {
		t.Fatalf("expected snake-case type name, got %q", got)
	}
}

func TestAutoMigrator_processStructFields_Embedded(t *testing.T) {
	am := &AutoMigrator{}
	var seen []string
	am.processStructFields(reflect.TypeOf(embeddedOuter{}), func(_ reflect.StructField, dbTag string) {
		seen = append(seen, dbTag)
	})
	got := strings.Join(seen, ",")
	// Expect id and name (embedded), but not ignored or unexported
	if !strings.Contains(got, "id") || !strings.Contains(got, "name") {
		t.Fatalf("expected to see id and name, got %q", got)
	}
	if strings.Contains(got, "ignored") {
		t.Fatalf("did not expect ignored tag to be processed: %q", got)
	}
}

func TestAutoMigrator_isEmbeddedStruct(t *testing.T) {
	am := &AutoMigrator{}
	typ := reflect.TypeOf(embeddedOuter{})
	// Find the embedded field
	var embeddedField reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Anonymous {
			embeddedField = f
			break
		}
	}
	if !am.isEmbeddedStruct(embeddedField) {
		t.Fatalf("expected embedded field to be detected")
	}
	// Non-embedded
	non := typ.Field(0)
	if am.isEmbeddedStruct(non) {
		t.Fatalf("did not expect non-embedded field to be detected as embedded")
	}
}

func TestAutoMigrator_generateColumnAndTableSQL(t *testing.T) {
	am := &AutoMigrator{db: nil}
	// Column generation
	ft, _ := reflect.TypeOf(embeddedInner{}).FieldByName("Name")
	col := am.generateColumnDefinition(ft, ft.Tag.Get("db"))
	if col == "" || !strings.Contains(col, "name ") { // driver-specific type varies; just check presence
		t.Fatalf("expected non-empty column definition containing name, got %q", col)
	}
	// Table SQL
	sql := am.generateCreateTableSQL("my_table", reflect.TypeOf(embeddedOuter{}))
	if !strings.HasPrefix(strings.TrimSpace(sql), "CREATE TABLE IF NOT EXISTS my_table (") {
		t.Fatalf("unexpected create table sql: %s", sql)
	}
}

func TestAutoMigrator_generateTableSchema_andChecksum(t *testing.T) {
	am := &AutoMigrator{}
	schema1 := am.generateTableSchema(reflect.TypeOf(embeddedOuter{}))
	if schema1 == "" {
		t.Fatalf("expected non-empty schema")
	}
	schema2 := am.generateTableSchema(reflect.TypeOf(embeddedInner{}))
	if schema2 == "" || schema2 == schema1 {
		t.Fatalf("expected different schema for inner vs outer")
	}
	sum1 := am.calculateChecksum(schema1)
	sum1b := am.calculateChecksum(schema1)
	sum2 := am.calculateChecksum(schema2)
	if sum1 != sum1b || sum1 == sum2 {
		t.Fatalf("checksum equality/inequality unexpected: %s %s %s", sum1, sum1b, sum2)
	}
}

func TestAutoMigrator_getTableColumnsQuery_builds(t *testing.T) {
	am := &AutoMigrator{}
	// Postgres
	q, args, err := am.getTableColumnsQuery(schema.PostgreSQL, "t")
	if err != nil || !strings.Contains(q, "information_schema.columns") || len(args) != 1 {
		t.Fatalf("postgres query unexpected: err=%v q=%q args=%v", err, q, args)
	}
	// SQLite
	q, args, err = am.getTableColumnsQuery(schema.SQLite, "t")
	if err != nil || !strings.Contains(q, "PRAGMA table_info(t)") || len(args) != 0 {
		t.Fatalf("sqlite query unexpected: err=%v q=%q args=%v", err, q, args)
	}
	// MySQL
	q, args, err = am.getTableColumnsQuery(schema.MySQL, "t")
	if err != nil || !strings.Contains(q, "information_schema.columns") || len(args) != 1 {
		t.Fatalf("mysql query unexpected: err=%v q=%q args=%v", err, q, args)
	}
}

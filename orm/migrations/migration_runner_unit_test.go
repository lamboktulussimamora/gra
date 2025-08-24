package migrations

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Helper struct to obtain reflect.StructField metadata easily
type runnerFieldSamples struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	NameSized string    `db:"name_sized" maxlength:"50"`
	Price     float64   `db:"price"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	Note      *string   `db:"note"`
	Ignored   struct{}  `db:"ignored"`
}

func fieldByName(t reflect.Type, name string) reflect.StructField {
	f, ok := t.FieldByName(name)
	if !ok {
		panic("field not found: " + name)
	}
	return f
}

func Test_sqlTypeForField_Various(t *testing.T) {
	typ := reflect.TypeOf(runnerFieldSamples{})

	// id -> SERIAL PRIMARY KEY
	f := fieldByName(typ, "ID")
	if sqlt, nullable := sqlTypeForField(f.Type, "id", f); sqlt != "SERIAL PRIMARY KEY" || nullable {
		t.Fatalf("id mapping unexpected: %q nullable=%v", sqlt, nullable)
	}

	// string with maxlength
	f = fieldByName(typ, "NameSized")
	if sqlt, _ := sqlTypeForField(f.Type, "name_sized", f); sqlt != "VARCHAR(50)" {
		t.Fatalf("maxlength mapping unexpected: %q", sqlt)
	}

	// plain string -> TEXT
	f = fieldByName(typ, "Name")
	if sqlt, _ := sqlTypeForField(f.Type, "name", f); sqlt != "TEXT" {
		t.Fatalf("string mapping unexpected: %q", sqlt)
	}

	// float -> DECIMAL(10,2)
	f = fieldByName(typ, "Price")
	if sqlt, _ := sqlTypeForField(f.Type, "price", f); sqlt != "DECIMAL(10,2)" {
		t.Fatalf("float mapping unexpected: %q", sqlt)
	}

	// bool -> BOOLEAN
	f = fieldByName(typ, "Active")
	if sqlt, _ := sqlTypeForField(f.Type, "active", f); sqlt != "BOOLEAN" {
		t.Fatalf("bool mapping unexpected: %q", sqlt)
	}

	// time -> TIMESTAMP
	f = fieldByName(typ, "CreatedAt")
	if sqlt, _ := sqlTypeForField(f.Type, "created_at", f); sqlt != "TIMESTAMP" {
		t.Fatalf("time mapping unexpected: %q", sqlt)
	}

	// pointer -> nullable true
	f = fieldByName(typ, "Note")
	if _, nullable := sqlTypeForField(f.Type, "note", f); !nullable {
		t.Fatalf("expected pointer field to be nullable")
	}

	// unsupported struct -> empty type
	f = fieldByName(typ, "Ignored")
	if sqlt, _ := sqlTypeForField(f.Type, "ignored", f); sqlt != "" {
		t.Fatalf("expected empty for unsupported struct, got %q", sqlt)
	}
}

func Test_addNotNullConstraint(t *testing.T) {
	if got := addNotNullConstraint("TEXT", "name", false); got != "TEXT NOT NULL" {
		t.Fatalf("expected NOT NULL added, got %q", got)
	}
	if got := addNotNullConstraint("TEXT", "id", false); got != "TEXT" {
		t.Fatalf("id should not add NOT NULL, got %q", got)
	}
	if got := addNotNullConstraint("TEXT", "name", true); got != "TEXT" {
		t.Fatalf("nullable should not add NOT NULL, got %q", got)
	}
}

func Test_addDefaultTimestamp(t *testing.T) {
	if got := addDefaultTimestamp("TIMESTAMP", "time.Time", "created_at"); got != "TIMESTAMP DEFAULT CURRENT_TIMESTAMP" {
		t.Fatalf("created_at default mismatch: %q", got)
	}
	if got := addDefaultTimestamp("TIMESTAMP", "time.Time", "updated_at"); got != "TIMESTAMP DEFAULT CURRENT_TIMESTAMP" {
		t.Fatalf("updated_at default mismatch: %q", got)
	}
	if got := addDefaultTimestamp("TIMESTAMP", "time.Time", "deleted_at"); got != "TIMESTAMP" {
		t.Fatalf("unexpected default added: %q", got)
	}
}

func Test_generateColumnDefinition_and_getTableName(t *testing.T) {
	typ := reflect.TypeOf(runnerFieldSamples{})
	// ID
	id := fieldByName(typ, "ID")
	if got := (&MigrationRunner{}).generateColumnDefinition(id, "id"); !strings.Contains(got, "id SERIAL PRIMARY KEY") {
		t.Fatalf("id column def mismatch: %q", got)
	}
	// Name (NOT NULL)
	name := fieldByName(typ, "Name")
	if got := (&MigrationRunner{}).generateColumnDefinition(name, "name"); got != "name TEXT NOT NULL" {
		t.Fatalf("name column def mismatch: %q", got)
	}
	// CreatedAt default with NOT NULL
	created := fieldByName(typ, "CreatedAt")
	if got := (&MigrationRunner{}).generateColumnDefinition(created, "created_at"); got != "created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP" {
		t.Fatalf("created_at column def mismatch: %q", got)
	}

	// getTableName pluralization + snake
	if got := (&MigrationRunner{}).getTableName("UserProfile"); got != "user_profiles" {
		t.Fatalf("table name mismatch: %q", got)
	}
}

func Test_generateCreateTableSQL(t *testing.T) {
	type Entity struct {
		ID        int       `db:"id"`
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at"`
		Ignored   string
	}
	sql := (&MigrationRunner{}).generateCreateTableSQL("entities", reflect.TypeOf(Entity{}))
	// Ensure wrapper and key columns exist
	if !strings.HasPrefix(strings.TrimSpace(sql), "CREATE TABLE IF NOT EXISTS entities (") || !strings.HasSuffix(strings.TrimSpace(sql), ")") {
		t.Fatalf("unexpected wrapper: %s", sql)
	}
	if !strings.Contains(sql, "id SERIAL PRIMARY KEY") || !strings.Contains(sql, "name TEXT NOT NULL") || !strings.Contains(sql, "created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("unexpected content: %s", sql)
	}
	if strings.Contains(sql, "Ignored") {
		t.Fatalf("should not include fields without db tag")
	}
}

package schema

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_ParseFieldToColumnForDriver_Postgres(t *testing.T) {
	type E struct {
		ID        int       `db:"id" sql:"primary_key;auto_increment"`
		Name      string    `db:"name" sql:"not_null"`
		Price     float64   `db:"price"`
		CreatedAt time.Time `db:"created_at" migration:"default:CURRENT_TIMESTAMP"`
		Note      *string   `db:"note"`
	}
	typ := reflect.TypeOf(E{})
	var defs []string
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Tag.Get("db") == "" {
			continue
		}
		def := ParseFieldToColumnForDriver(f, PostgreSQL)
		if def == "" {
			t.Fatalf("field %s not parsed", f.Name)
		}
		defs = append(defs, def)
	}
	if len(defs) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(defs))
	}
	all := strings.Join(defs, " ")
	if !strings.Contains(all, "id SERIAL PRIMARY KEY") {
		t.Fatalf("id definition mismatch: %s", all)
	}
	if !strings.Contains(all, "name VARCHAR(255) NOT NULL") {
		t.Fatalf("name definition mismatch: %s", all)
	}
	if !strings.Contains(all, "price DOUBLE PRECISION") {
		t.Fatalf("price type mismatch: %s", all)
	}
	if !strings.Contains(all, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("created_at default missing: %s", all)
	}
	if !strings.Contains(all, "note VARCHAR(255)") { // nullable by default for pointer
		t.Fatalf("note type mismatch: %s", all)
	}
}

func Test_GenerateCreateTableSQL_Postgres(t *testing.T) {
	type E struct {
		ID        int       `db:"id" sql:"primary_key;auto_increment"`
		Name      string    `db:"name" sql:"not_null"`
		CreatedAt time.Time `db:"created_at" migration:"default:CURRENT_TIMESTAMP"`
	}
	sql := GenerateCreateTableSQL(E{}, "entities")
	s := strings.TrimSpace(sql)
	if !strings.HasPrefix(s, "CREATE TABLE IF NOT EXISTS entities (") || !strings.HasSuffix(s, ");") {
		t.Fatalf("unexpected wrapper: %s", sql)
	}
	if !strings.Contains(sql, "id SERIAL PRIMARY KEY") {
		t.Fatalf("missing id pk: %s", sql)
	}
	if !strings.Contains(sql, "name VARCHAR(255) NOT NULL") {
		t.Fatalf("missing required name: %s", sql)
	}
	if !strings.Contains(sql, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("missing created_at default: %s", sql)
	}
}

func Test_ParseField_DefaultAndNullability(t *testing.T) {
	type E struct {
		// nullable pointer without not_null tag
		Note *string `db:"note"`
		// explicit not null and default value
		Count int `db:"count" sql:"not_null;default:0"`
	}
	typ := reflect.TypeOf(E{})
	var defs []string
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		d := ParseFieldToColumnForDriver(f, PostgreSQL)
		if d != "" {
			defs = append(defs, d)
		}
	}
	all := strings.Join(defs, " ")
	if !strings.Contains(all, "note VARCHAR(255)") { // no NOT NULL
		t.Fatalf("note should be nullable by default: %s", all)
	}
	if !strings.Contains(all, "count INTEGER NOT NULL DEFAULT 0") {
		t.Fatalf("count constraint/default mismatch: %s", all)
	}
}

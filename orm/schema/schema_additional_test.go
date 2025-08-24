package schema

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test additional SQL helper generators and type mappings across drivers
func Test_SQLHelpersAndTypeMappings(t *testing.T) {
	// Index SQL
	idx := GenerateIndexSQL("users", "idx_users_name", []string{"name"}, false)
	if !strings.Contains(idx, "CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);") {
		t.Fatalf("unexpected index sql: %s", idx)
	}

	uniqIdx := GenerateIndexSQL("users", "uidx_users_email", []string{"email"}, true)
	if !strings.Contains(uniqIdx, "CREATE UNIQUE INDEX IF NOT EXISTS uidx_users_email ON users (email);") {
		t.Fatalf("unexpected unique index sql: %s", uniqIdx)
	}

	// Foreign key SQL
	fk := GenerateForeignKeySQL("orders", "user_id", "users", "id")
	if !strings.Contains(fk, "ALTER TABLE orders ADD CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id);") {
		t.Fatalf("unexpected fk sql: %s", fk)
	}

	// Drop table SQL
	drop := GenerateDropTableSQL("temp_table")
	if drop != "DROP TABLE IF EXISTS temp_table CASCADE;" {
		t.Fatalf("unexpected drop sql: %s", drop)
	}

	// Type mapping through column generation per driver
	type Sample struct {
		S   string    `db:"s"`
		I   int       `db:"i"`
		I64 int64     `db:"i64"`
		F32 float32   `db:"f32"`
		F64 float64   `db:"f64"`
		B   bool      `db:"b"`
		T   time.Time `db:"t"`
		P   *string   `db:"p"`
	}

	typ := reflect.TypeOf(Sample{})
	// helper to collect per-driver column defs
	collect := func(driver DatabaseDriver) string {
		var cols []string
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			def := ParseFieldToColumnForDriver(f, driver)
			if def != "" {
				cols = append(cols, def)
			}
		}
		return strings.Join(cols, ", ")
	}

	// PostgreSQL expectations
	pg := collect(PostgreSQL)
	for _, exp := range []string{
		"s VARCHAR(255)",
		"i INTEGER",
		"i64 BIGINT",
		"f32 REAL",
		"f64 DOUBLE PRECISION",
		"b BOOLEAN",
		"t TIMESTAMP",
		"p VARCHAR(255)", // pointer still maps to base type
	} {
		if !strings.Contains(pg, exp) {
			t.Fatalf("postgres mapping missing %q in %s", exp, pg)
		}
	}

	// SQLite expectations
	sq := collect(SQLite)
	for _, exp := range []string{
		"s TEXT",
		"i INTEGER",
		"i64 INTEGER",
		"f32 REAL",
		"f64 REAL",
		"b INTEGER",  // boolean as INTEGER
		"t DATETIME", // time as DATETIME
		"p TEXT",
	} {
		if !strings.Contains(sq, exp) {
			t.Fatalf("sqlite mapping missing %q in %s", exp, sq)
		}
	}

	// MySQL expectations
	my := collect(MySQL)
	for _, exp := range []string{
		"s VARCHAR(255)",
		"i INT",
		"i64 BIGINT",
		"f32 FLOAT",
		"f64 DOUBLE",
		"b BOOLEAN",
		"t DATETIME",
		"p VARCHAR(255)",
	} {
		if !strings.Contains(my, exp) {
			t.Fatalf("mysql mapping missing %q in %s", exp, my)
		}
	}
}

// Validate that fields which are navigation-like are skipped (pointer to struct without db tag)
func Test_NavigationPropertySkipped(t *testing.T) {
	type Related struct {
		Name string
	}
	type Owner struct {
		ID   int      `db:"id"`
		Ref  *Related // no db tag => should be skipped
		Note string   `db:"note"`
	}
	sql := GenerateCreateTableSQLForDriver(Owner{}, "owners", PostgreSQL)
	if strings.Contains(sql, "Ref") || strings.Contains(sql, "related") {
		t.Fatalf("navigation property not skipped: %s", sql)
	}
	if !strings.Contains(sql, "id INTEGER") || !strings.Contains(sql, "note VARCHAR(255)") {
		t.Fatalf("expected columns missing: %s", sql)
	}
}

// Cover DetectDatabaseDriverFromConnectionString variant and extractSQLValue helper
func Test_ConnectionStringAndExtractSQLValue(t *testing.T) {
	if DetectDatabaseDriverFromConnectionString("postgresql") != PostgreSQL {
		t.Fatal("expected postgresql alias to resolve to PostgreSQL")
	}

	tag := "primary_key;default:CURRENT_TIMESTAMP;unique"
	if v := extractSQLValue(tag, "default"); v != "CURRENT_TIMESTAMP" {
		t.Fatalf("expected CURRENT_TIMESTAMP, got %q", v)
	}
	if v := extractSQLValue(tag, "missing"); v != "" {
		t.Fatalf("expected empty for missing key, got %q", v)
	}
}

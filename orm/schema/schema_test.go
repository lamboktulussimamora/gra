package schema

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test data structures to exercise parsing logic
type embeddedAudit struct {
	CreatedAt time.Time `db:"created_at" sql:"default:CURRENT_TIMESTAMP"`
}

type profile struct { // no db tags -> should be skipped when used as pointer
	Note string
}

type userEntity struct {
	ID            int64     `db:"id" sql:"primary_key;auto_increment"`
	Name          string    `db:"name" sql:"not_null;unique"`
	Email         *string   `db:"email"`
	Age           int       `db:"age" sql:"default:18"`
	Bio           *string   `db:"bio" migration:"type:TEXT;default:'hello'"`
	embeddedAudit           // embedded fields should be inlined
	Ignored       time.Time // no db tag, ensure skipped
	Posts         []string  // navigation property (slice) skipped
	Profile       *profile  // pointer to struct without db tag -> skipped
	Created       time.Time `db:"created"`
}

func TestGenerateCreateTableSQLForDriver_Postgres(t *testing.T) {
	sql := GenerateCreateTableSQLForDriver(userEntity{}, "users", PostgreSQL)

	checks := []string{
		// auto increment + primary key mapping to SERIAL types
		"id BIGSERIAL PRIMARY KEY",
		// constraints and flags
		"name VARCHAR(255) NOT NULL UNIQUE",
		// pointer type -> nullable column
		"email VARCHAR(255)",
		// integer default value
		"age INTEGER DEFAULT 18",
		// override type + default in migration tag (no quotes are re-added by implementation)
		"bio TEXT DEFAULT hello",
		// embedded field + CURRENT_TIMESTAMP default
		"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		// normal time mapping
		"created TIMESTAMP",
	}

	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("expected SQL to contain %q, got:\n%s", want, sql)
		}
	}

	if !strings.HasPrefix(sql, "CREATE TABLE IF NOT EXISTS users (") || !strings.HasSuffix(sql, ");") {
		t.Fatalf("unexpected SQL wrapper: %s", sql)
	}
}

func TestGenerateCreateTableSQLForDriver_SQLite(t *testing.T) {
	type s struct {
		ID int `db:"id" sql:"primary_key;auto_increment"`
	}
	sql := GenerateCreateTableSQLForDriver(s{}, "t", SQLite)
	if !strings.Contains(sql, "id INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Fatalf("sqlite autoincrement mapping not applied, sql=%s", sql)
	}
}

func TestGenerateCreateTableSQLForDriver_MySQL(t *testing.T) {
	type s struct {
		ID int64 `db:"id" sql:"primary_key;auto_increment"`
	}
	sql := GenerateCreateTableSQLForDriver(s{}, "t", MySQL)
	if !strings.Contains(sql, "id BIGINT PRIMARY KEY AUTO_INCREMENT") {
		t.Fatalf("mysql autoincrement mapping not applied, sql=%s", sql)
	}
}

func TestParseFieldToColumnForDriver_Various(t *testing.T) {
	var e userEntity
	typ := reflect.TypeOf(e)

	field := typ.Field(0) // ID
	col := ParseFieldToColumnForDriver(field, PostgreSQL)
	if !strings.Contains(col, "BIGSERIAL") || !strings.Contains(col, "PRIMARY KEY") {
		t.Fatalf("unexpected ID column: %s", col)
	}

	// migration tag overrides type and default
	bioField, _ := typ.FieldByName("Bio")
	col = ParseFieldToColumnForDriver(bioField, PostgreSQL)
	if !strings.Contains(col, "bio TEXT") || !strings.Contains(col, "DEFAULT hello") {
		t.Fatalf("unexpected Bio column: %s", col)
	}

	// CURRENT_TIMESTAMP default passthrough
	createdAtField, _ := reflect.TypeOf(embeddedAudit{}).FieldByName("CreatedAt")
	col = ParseFieldToColumnForDriver(createdAtField, PostgreSQL)
	if !strings.Contains(col, "TIMESTAMP") || !strings.Contains(col, "DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("unexpected CreatedAt column: %s", col)
	}
}

func TestExtractSQLValue(t *testing.T) {
	if got := extractSQLValue("default:CURRENT_TIMESTAMP;not_null", "default"); got != "CURRENT_TIMESTAMP" {
		t.Fatalf("want CURRENT_TIMESTAMP, got %s", got)
	}
	if got := extractSQLValue("type:TEXT;default:'hello'", "default"); got != "hello" {
		t.Fatalf("want hello, got %s", got)
	}
	if got := extractSQLValue("", "default"); got != "" {
		t.Fatalf("want empty, got %s", got)
	}
}

func TestHelpers_SQLGenerators(t *testing.T) {
	if got := GenerateDropTableSQL("users"); got != "DROP TABLE IF EXISTS users CASCADE;" {
		t.Fatalf("drop table mismatch: %s", got)
	}
	if got := GenerateIndexSQL("users", "idx_users_name", []string{"name"}, false); got != "CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);" {
		t.Fatalf("index sql mismatch: %s", got)
	}
	if got := GenerateIndexSQL("users", "uidx_users_email", []string{"email"}, true); got != "CREATE UNIQUE INDEX IF NOT EXISTS uidx_users_email ON users (email);" {
		t.Fatalf("unique index sql mismatch: %s", got)
	}
	if got := GenerateForeignKeySQL("orders", "user_id", "users", "id"); got != "ALTER TABLE orders ADD CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id);" {
		t.Fatalf("fk sql mismatch: %s", got)
	}
}

func TestDetectDatabaseDriverFromConnectionString(t *testing.T) {
	cases := map[string]DatabaseDriver{
		"postgres":   PostgreSQL,
		"PostgreSQL": PostgreSQL,
		"sqlite3":    SQLite,
		"sqlite":     SQLite,
		"mysql":      MySQL,
		"unknown":    PostgreSQL, // default fallback
	}
	for in, want := range cases {
		if got := DetectDatabaseDriverFromConnectionString(in); got != want {
			t.Fatalf("input=%s want=%s got=%s", in, want, got)
		}
	}
}

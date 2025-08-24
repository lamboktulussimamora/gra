package migrations

import (
	"reflect"
	"testing"
	"time"
)

// Test types for snapshotting
type regAudit struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type regUser struct {
	regAudit
	ID      int64   `db:"id" sql:"primary_key;auto_increment"`
	Name    string  `db:"name" sql:"size:100"`
	Bio     string  `db:"bio" migration:"type:TEXT"`
	Age     int     `db:"age"`
	Balance float64 `db:"balance" sql:"precision:10;scale:2"`
	Active  bool    `db:"active"`
	RoleID  int64   `db:"role_id" sql:"foreign_key:roles(id)"`
	Email   *string `db:"email" default:"''"`
	SkipMe  string  `db:"-"`
}

func (regUser) TableName() string { return "reg_users" }

func TestModelRegistry_CreateSnapshot_Postgres(t *testing.T) {
	mr := NewModelRegistry(PostgreSQL)
	mr.RegisterModel(regUser{})

	models := mr.GetModels()
	snap, ok := models["reg_users"]
	if !ok {
		t.Fatalf("expected snapshot for reg_users")
	}

	// Columns present
	cols := snap.Columns
	if _, ok := cols["id"]; !ok {
		t.Fatalf("missing id column")
	}
	if _, ok := cols["name"]; !ok {
		t.Fatalf("missing name column")
	}
	if _, ok := cols["bio"]; !ok {
		t.Fatalf("missing bio column")
	}
	if _, ok := cols["balance"]; !ok {
		t.Fatalf("missing balance column")
	}
	if _, ok := cols["role_id"]; !ok {
		t.Fatalf("missing role_id column")
	}
	if _, ok := cols["email"]; !ok {
		t.Fatalf("missing email column")
	}
	if _, ok := cols["skip_me"]; ok {
		t.Fatalf("skip_me should be omitted via db:\"-\"")
	}

	// Types and flags
	if !cols["id"].IsPrimaryKey || !cols["id"].IsIdentity {
		t.Fatalf("id should be primary key identity")
	}
	if cols["name"].SQLType != "VARCHAR(100)" {
		t.Fatalf("name SQLType expected VARCHAR(100), got %s", cols["name"].SQLType)
	}
	if cols["bio"].SQLType != sqlTypeText {
		t.Fatalf("bio should map to TEXT, got %s", cols["bio"].SQLType)
	}
	if cols["balance"].SQLType != "DECIMAL(10,2)" {
		t.Fatalf("balance should be DECIMAL(10,2), got %s", cols["balance"].SQLType)
	}
	if cols["active"].SQLType != "BOOLEAN" {
		t.Fatalf("active should be BOOLEAN on postgres, got %s", cols["active"].SQLType)
	}
	if !cols["role_id"].IsForeignKey || cols["role_id"].References == nil || cols["role_id"].References.Table != "roles" {
		t.Fatalf("role_id should be FK to roles(id)")
	}
	if cols["email"].Default == nil || *cols["email"].Default != "''" {
		t.Fatalf("email default extraction failed")
	}

	// Indexes and constraints inferred from tags
	if len(snap.Indexes) != 0 {
		t.Fatalf("no explicit indexes expected in this model")
	}
	// A FK constraint should be present in constraints map
	foundFK := false
	for _, c := range snap.Constraints {
		if c.Type == "FOREIGN_KEY" {
			foundFK = true
			break
		}
	}
	if !foundFK {
		t.Fatalf("expected a foreign key constraint entry")
	}

	// Checksum should be deterministic and non-empty
	if snap.Checksum == "" || len(snap.Checksum) < 16 {
		t.Fatalf("unexpected checksum: %q", snap.Checksum)
	}
}

func TestModelRegistry_TypeMappingAcrossDrivers(t *testing.T) {
	// Prepare a reflect.Field set using a struct with representative types
	type sample struct {
		A bool
		B int `sql:"primary_key"`
		C int64
		D float32
		E float64 `sql:"precision:8;scale:3"`
		F string  `sql:"size:50"`
		G time.Time
	}
	st := reflect.TypeOf(sample{})

	cases := []struct {
		drv      DatabaseDriver
		expBool  string
		expIntPK string
		expBig   string
		expReal  string
		expDbl   string
		expV     string
		expTime  string
	}{
		{SQLite, sqlTypeInteger, sqlTypeInteger, sqlTypeInteger, sqlTypeReal, "REAL", "VARCHAR(50)", "TIMESTAMP"},
		{MySQL, "TINYINT(1)", "INT", sqlTypeBigInt, "FLOAT", "DOUBLE", "VARCHAR(50)", "TIMESTAMP"},
		// For PostgreSQL, when precision/scale tags are present on float64, we map to DECIMAL(precision,scale)
		{PostgreSQL, "BOOLEAN", sqlTypeSerial, sqlTypeBigInt, sqlTypeReal, "DECIMAL(8,3)", "VARCHAR(50)", "TIMESTAMP"},
	}

	for _, c := range cases {
		mr := NewModelRegistry(c.drv)
		if got := mr.getSQLType(st.Field(0), st.Field(0).Type); got != c.expBool {
			t.Fatalf("%s bool: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(1), st.Field(1).Type); got != c.expIntPK {
			t.Fatalf("%s int pk: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(2), st.Field(2).Type); got != c.expBig {
			t.Fatalf("%s bigint: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(3), st.Field(3).Type); got != c.expReal {
			t.Fatalf("%s real: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(4), st.Field(4).Type); got != c.expDbl {
			t.Fatalf("%s double: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(5), st.Field(5).Type); got != c.expV {
			t.Fatalf("%s varchar: got %s", c.drv, got)
		}
		if got := mr.getSQLType(st.Field(6), st.Field(6).Type); got != c.expTime {
			t.Fatalf("%s time: got %s", c.drv, got)
		}
	}
}

func TestModelRegistry_ExplicitSQLTypeAndSizeParsing(t *testing.T) {
	type sample struct {
		T string `migration:"type:TEXT"`
		U string `sql:"type:TEXT;size:123"`
	}
	st := reflect.TypeOf(sample{})
	if tp, ok := getExplicitSQLType(st.Field(0)); !ok || tp != sqlTypeText {
		t.Fatalf("migration tag explicit type failed: %v %s", ok, tp)
	}
	if tp, ok := getExplicitSQLType(st.Field(1)); !ok || tp != sqlTypeText {
		t.Fatalf("sql tag explicit type failed: %v %s", ok, tp)
	}

	mr := NewModelRegistry(PostgreSQL)
	if sz := mr.getSizeFromSQLTag(st.Field(1)); sz != 123 {
		t.Fatalf("size parse failed, got %d", sz)
	}
}

func TestModelRegistry_PrimaryUniqueForeignKeyDetection(t *testing.T) {
	type m struct {
		ID int64  `sql:"primary_key;auto_increment"`
		U  string `sql:"unique"`
		R  int64  `sql:"foreign_key:roles(id)"`
		X  string `sql:"-"`
	}
	st := reflect.TypeOf(m{})
	mr := NewModelRegistry(PostgreSQL)
	if !mr.isPrimaryKey(st.Field(0)) || !mr.isAutoIncrement(st.Field(0)) {
		t.Fatalf("primary/auto detection failed")
	}
	if !mr.isUnique(st.Field(1)) {
		t.Fatalf("unique detection failed")
	}
	if !mr.isForeignKey(st.Field(2)) {
		t.Fatalf("fk detection failed")
	}
	if mr.isForeignKey(st.Field(3)) { // sql:"-"
		t.Fatalf("field with sql:- should not be FK")
	}
}

func TestModelRegistry_ForeignKeyParsing(t *testing.T) {
	info := parseForeignKey("roles(id)")
	if info == nil || info.Table != "roles" || info.Column != "id" {
		t.Fatalf("parseForeignKey failed: %+v", info)
	}
}

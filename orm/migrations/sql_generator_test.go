package migrations

import (
	"strings"
	"testing"
)

// helper to ptr int
func intPtr(v int) *int { return &v }

func TestSQLGenerator_CreateTable_PostgreSQL(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)

	snapshot := &ModelSnapshot{
		TableName: "products",
		Columns: map[string]*ColumnInfo{
			"id": {
				Name:         "id",
				DataType:     "INT64",
				IsPrimaryKey: true,
				IsIdentity:   true,
				IsNullable:   false,
			},
			"name": {
				Name:       "name",
				DataType:   "STRING",
				MaxLength:  intPtr(100),
				IsNullable: false,
				DefaultValue: func() *string {
					v := "''"
					return &v
				}(),
			},
			"description": {
				Name:       "description",
				DataType:   "TEXT",
				IsNullable: true,
			},
			"price": {
				Name:       "price",
				DataType:   "DECIMAL",
				Precision:  intPtr(10),
				Scale:      intPtr(2),
				IsNullable: false,
				DefaultValue: func() *string {
					v := "0"
					return &v
				}(),
			},
			"user_id": {
				Name:         "user_id",
				DataType:     "INT64",
				IsNullable:   false,
				IsForeignKey: true,
			},
			"created_at": {
				Name:       "created_at",
				DataType:   "TIME",
				IsNullable: false,
			},
		},
		Indexes: map[string]IndexInfo{
			"ix_products_name": {Name: "ix_products_name", Columns: []string{"name"}, IsUnique: true},
			"ix_products_user": {Name: "ix_products_user", Columns: []string{"user_id"}},
		},
		Constraints: map[string]*ConstraintInfo{
			"fk_products_user": {
				Name:              "fk_products_user",
				Type:              foreignKeyConstraintType,
				Columns:           []string{"user_id"},
				ReferencedTable:   "users",
				ReferencedColumns: []string{"id"},
			},
		},
	}

	change := MigrationChange{Type: CreateTable, TableName: snapshot.TableName, NewValue: snapshot}
	sql, err := sg.generateCreateTableSQL(change)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// id will be SERIAL (applyIdentityClause inspects DataType, not resolved type)
	if !strings.Contains(sql, `id SERIAL`) {
		t.Errorf("expected SERIAL id, got: %s", sql)
	}
	if !strings.Contains(sql, "PRIMARY KEY (id)") {
		t.Errorf("expected PRIMARY KEY constraint in: %s", sql)
	}
	// name should be VARCHAR(100) NOT NULL DEFAULT ''
	if !strings.Contains(sql, `name VARCHAR(100) NOT NULL DEFAULT ''`) {
		t.Errorf("expected name definition with NOT NULL and DEFAULT, got: %s", sql)
	}
	// price DECIMAL(10,2) NOT NULL DEFAULT 0
	if !strings.Contains(sql, `price DECIMAL(10,2) NOT NULL DEFAULT 0`) {
		t.Errorf("expected price definition, got: %s", sql)
	}
	// indexes
	if !strings.Contains(sql, `CREATE UNIQUE INDEX "ix_products_name" ON "products" ("name");`) {
		t.Errorf("expected unique index creation, got: %s", sql)
	}
	if !strings.Contains(sql, `CREATE INDEX "ix_products_user" ON "products" ("user_id");`) {
		t.Errorf("expected non-unique index creation, got: %s", sql)
	}
	// foreign key
	if !strings.Contains(sql, `ALTER TABLE "products" ADD CONSTRAINT "fk_products_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id");`) {
		t.Errorf("expected foreign key statement, got: %s", sql)
	}
}

func TestSQLGenerator_BasicDDL(t *testing.T) {
	// Test drop table, add/drop column
	sg := NewSQLGenerator(PostgreSQL)
	drop, err := sg.generateDropTableSQL(MigrationChange{Type: DropTable, TableName: "t"})
	if err != nil || drop != `DROP TABLE IF EXISTS "t";` {
		t.Fatalf("unexpected drop table: %v %s", err, drop)
	}

	addCol, err := sg.generateAddColumnSQL(MigrationChange{Type: AddColumn, TableName: "t", ColumnName: "c", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(10), IsNullable: false}})
	if err != nil || !strings.Contains(addCol, `ALTER TABLE "t" ADD COLUMN "c" VARCHAR(10) NOT NULL;`) {
		t.Fatalf("unexpected add column: %v %s", err, addCol)
	}

	dropCol, err := sg.generateDropColumnSQL(MigrationChange{Type: DropColumn, TableName: "t", ColumnName: "c"})
	if err != nil || dropCol != `ALTER TABLE "t" DROP COLUMN IF EXISTS "c";` {
		t.Fatalf("unexpected drop column: %v %s", err, dropCol)
	}
}

func TestSQLGenerator_AlterColumn_Postgres(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	def := "'x'"
	sql, err := sg.generateAlterColumnSQL(MigrationChange{Type: AlterColumn, TableName: "t", ColumnName: "c", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(50), IsNullable: false, DefaultValue: &def}})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(sql, `ALTER TABLE "t" ALTER COLUMN "c" TYPE VARCHAR(50);`) {
		t.Errorf("missing TYPE alter: %s", sql)
	}
	if !strings.Contains(sql, `ALTER TABLE "t" ALTER COLUMN "c" SET NOT NULL;`) {
		t.Errorf("missing NOT NULL alter: %s", sql)
	}
	if !strings.Contains(sql, `ALTER TABLE "t" ALTER COLUMN "c" SET DEFAULT 'x';`) {
		t.Errorf("missing DEFAULT alter: %s", sql)
	}

	// Now test dropping default and allowing nulls
	sql, err = sg.generateAlterColumnSQL(MigrationChange{Type: AlterColumn, TableName: "t", ColumnName: "c", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(20), IsNullable: true}})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(sql, `ALTER TABLE "t" ALTER COLUMN "c" DROP NOT NULL;`) {
		t.Errorf("missing DROP NOT NULL alter: %s", sql)
	}
	if !strings.Contains(sql, `ALTER TABLE "t" ALTER COLUMN "c" DROP DEFAULT;`) {
		t.Errorf("missing DROP DEFAULT alter: %s", sql)
	}
}

func TestSQLGenerator_AlterColumn_MySQL(t *testing.T) {
	sg := NewSQLGenerator(MySQL)
	sql, err := sg.generateAlterColumnSQL(MigrationChange{Type: AlterColumn, TableName: "t", ColumnName: "c", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(10), IsNullable: false}})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(sql, "ALTER TABLE `t` MODIFY COLUMN `c` VARCHAR(10) NOT NULL;") {
		t.Errorf("unexpected mysql alter: %s", sql)
	}
}

func TestSQLGenerator_IndexDDL_AllDrivers(t *testing.T) {
	// Postgres create/drop index
	pg := NewSQLGenerator(PostgreSQL)
	ci := &IndexInfo{Name: "ix_t_c", Columns: []string{"c"}}
	pgCreate, err := pg.generateCreateIndexSQL(MigrationChange{Type: CreateIndex, TableName: "t", IndexName: "ix_t_c", NewValue: ci})
	if err != nil || pgCreate != `CREATE INDEX "ix_t_c" ON "t" ("c");` {
		t.Fatalf("unexpected pg create index: %v %s", err, pgCreate)
	}
	pgDrop, err := pg.generateDropIndexSQL(MigrationChange{Type: DropIndex, TableName: "t", IndexName: "ix_t_c"})
	if err != nil || pgDrop != `DROP INDEX IF EXISTS "ix_t_c";` {
		t.Fatalf("unexpected pg drop index: %v %s", err, pgDrop)
	}

	// MySQL
	my := NewSQLGenerator(MySQL)
	myCreate, err := my.generateCreateIndexSQL(MigrationChange{Type: CreateIndex, TableName: "t", IndexName: "ix_t_c", NewValue: ci})
	if err != nil || myCreate != "CREATE INDEX `ix_t_c` ON `t` (`c`);" {
		t.Fatalf("unexpected mysql create index: %v %s", err, myCreate)
	}
	myDrop, err := my.generateDropIndexSQL(MigrationChange{Type: DropIndex, TableName: "t", IndexName: "ix_t_c"})
	if err != nil || myDrop != "DROP INDEX `ix_t_c` ON `t`;" {
		t.Fatalf("unexpected mysql drop index: %v %s", err, myDrop)
	}

	// SQLite
	sq := NewSQLGenerator(SQLite)
	sqCreate, err := sq.generateCreateIndexSQL(MigrationChange{Type: CreateIndex, TableName: "t", IndexName: "ix_t_c", NewValue: ci})
	if err != nil || sqCreate != `CREATE INDEX "ix_t_c" ON "t" ("c");` {
		t.Fatalf("unexpected sqlite create index: %v %s", err, sqCreate)
	}
	sqDrop, err := sq.generateDropIndexSQL(MigrationChange{Type: DropIndex, TableName: "t", IndexName: "ix_t_c"})
	if err != nil || sqDrop != `DROP INDEX IF EXISTS "ix_t_c";` {
		t.Fatalf("unexpected sqlite drop index: %v %s", err, sqDrop)
	}
}

func TestSQLGenerator_GenerateMigrationSQL_UpDown(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	// Simple plan: create table with one column, then add column, then create index
	snap := &ModelSnapshot{
		TableName: "a",
		Columns: map[string]*ColumnInfo{
			"id": {Name: "id", DataType: "INT64", IsPrimaryKey: true, IsIdentity: true, IsNullable: false},
		},
		Indexes:     map[string]IndexInfo{},
		Constraints: map[string]*ConstraintInfo{},
	}
	changes := []MigrationChange{
		{Type: CreateTable, TableName: "a", NewValue: snap},
		{Type: AddColumn, TableName: "a", ColumnName: "name", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(20), IsNullable: false}},
		{Type: CreateIndex, TableName: "a", IndexName: "ix_a_name", NewValue: &IndexInfo{Name: "ix_a_name", Columns: []string{"name"}}},
	}
	plan := &MigrationPlan{Changes: changes}
	out, err := sg.GenerateMigrationSQL(plan)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !strings.Contains(out.UpScript, `CREATE TABLE "a"`) {
		t.Errorf("up script missing create table: %s", out.UpScript)
	}
	if !strings.Contains(out.UpScript, `ALTER TABLE "a" ADD COLUMN "name" VARCHAR(20) NOT NULL;`) {
		t.Errorf("up script missing add column: %s", out.UpScript)
	}
	if !strings.Contains(out.DownScript, `DROP INDEX IF EXISTS "ix_a_name";`) {
		t.Errorf("down script missing drop index: %s", out.DownScript)
	}
	if !strings.Contains(out.DownScript, `ALTER TABLE "a" DROP COLUMN IF EXISTS "name";`) {
		t.Errorf("down script missing drop column: %s", out.DownScript)
	}
	if !strings.Contains(out.DownScript, `DROP TABLE IF EXISTS "a";`) {
		t.Errorf("down script missing drop table: %s", out.DownScript)
	}
}

func TestSQLGenerator_QuoteIdentifier(t *testing.T) {
	if NewSQLGenerator(PostgreSQL).quoteIdentifier("x") != `"x"` {
		t.Errorf("postgres quoting failed")
	}
	if NewSQLGenerator(MySQL).quoteIdentifier("x") != "`x`" {
		t.Errorf("mysql quoting failed")
	}
	if NewSQLGenerator(SQLite).quoteIdentifier("x") != `"x"` {
		t.Errorf("sqlite quoting failed")
	}
}

func TestSQLGenerator_CreateTable_SQLite_IdentityPK(t *testing.T) {
	sg := NewSQLGenerator(SQLite)
	snap := &ModelSnapshot{
		TableName: "test",
		Columns: map[string]*ColumnInfo{
			"id":   {Name: "id", DataType: "INT64", IsPrimaryKey: true, IsIdentity: true, IsNullable: false},
			"name": {Name: "name", DataType: "STRING", MaxLength: intPtr(10), IsNullable: false},
		},
		Indexes:     map[string]IndexInfo{},
		Constraints: map[string]*ConstraintInfo{},
	}
	sql, err := sg.generateCreateTableSQL(MigrationChange{Type: CreateTable, TableName: "test", NewValue: snap})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT") {
		t.Errorf("expected sqlite identity inline primary key: %s", sql)
	}
	if strings.Contains(sql, "PRIMARY KEY (id)") {
		t.Errorf("should not add separate PK constraint for sqlite identity: %s", sql)
	}
}

func TestSQLGenerator_IdentityBigInt_Postgres(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	col := &ColumnInfo{DataType: "BIGINT", IsIdentity: true, IsNullable: false}
	def := sg.generateColumnDefinition(col)
	if !strings.HasPrefix(def, "BIGSERIAL") {
		t.Errorf("expected BIGSERIAL for bigint identity, got: %s", def)
	}
}

func TestSQLGenerator_TypeResolutionAndLength(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	// supportsLength
	if !sg.supportsLength("VARCHAR") {
		t.Errorf("expected VARCHAR to support length")
	}
	if sg.supportsLength("TEXT") {
		t.Errorf("TEXT should not support length")
	}
	// precision/scale vs length
	c1 := &ColumnInfo{DataType: "DECIMAL", Precision: intPtr(8), Scale: intPtr(3)}
	if got := sg.resolveColumnDataType(c1); got != "DECIMAL(8,3)" {
		t.Errorf("expected DECIMAL(8,3), got %s", got)
	}
	c2 := &ColumnInfo{DataType: "STRING", MaxLength: intPtr(50)}
	if got := sg.resolveColumnDataType(c2); got != "VARCHAR(50)" {
		t.Errorf("expected VARCHAR(50), got %s", got)
	}
}

func TestSQLGenerator_CreateIndex_UniqueMultiColumn(t *testing.T) {
	// MySQL variant
	sg := NewSQLGenerator(MySQL)
	idx := &IndexInfo{Name: "ix_t_ab", Columns: []string{"a", "b"}, IsUnique: true}
	stmt := sg.generateCreateIndexStatement("t", "ix_t_ab", idx)
	if stmt != "CREATE UNIQUE INDEX `ix_t_ab` ON `t` (`a`, `b`);" {
		t.Errorf("unexpected unique index stmt: %s", stmt)
	}
}

func TestSQLGenerator_GenerateMigrationSQL_EmptyPlan(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	out, err := sg.GenerateMigrationSQL(&MigrationPlan{Changes: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out.UpScript, "-- No changes detected") {
		t.Errorf("unexpected up script: %s", out.UpScript)
	}
	if !strings.HasPrefix(out.DownScript, "-- No changes to revert") {
		t.Errorf("unexpected down script: %s", out.DownScript)
	}
}

func TestSQLGenerator_AddForeignKey_Composite(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	c := &ConstraintInfo{Type: foreignKeyConstraintType, Columns: []string{"a", "b"}, ReferencedTable: "r", ReferencedColumns: []string{"x", "y"}}
	stmt := sg.generateAddForeignKeySQL("t", "fk_t_r", c)
	expected := `ALTER TABLE "t" ADD CONSTRAINT "fk_t_r" FOREIGN KEY ("a", "b") REFERENCES "r" ("x", "y");`
	if stmt != expected {
		t.Errorf("unexpected fk stmt: %s", stmt)
	}
}

func TestSQLGenerator_AlterColumn_SQLiteError(t *testing.T) {
	sg := NewSQLGenerator(SQLite)
	_, err := sg.generateAlterColumnSQL(MigrationChange{Type: AlterColumn, TableName: "t", ColumnName: "c", NewValue: &ColumnInfo{DataType: "STRING", MaxLength: intPtr(10)}})
	if err == nil || !strings.Contains(err.Error(), "SQLite does not support ALTER COLUMN directly") {
		t.Errorf("expected sqlite alter column error, got: %v", err)
	}
}

func TestSQLGenerator_NullabilityAndDefaultClause(t *testing.T) {
	sg := NewSQLGenerator(PostgreSQL)
	def := "now()"
	parts := sg.nullabilityAndDefaultClause(&ColumnInfo{IsNullable: false, DefaultValue: &def})
	joined := strings.Join(parts, " ")
	if !strings.Contains(joined, "NOT NULL") || !strings.Contains(joined, "DEFAULT now()") {
		t.Errorf("unexpected clauses: %v", parts)
	}
}

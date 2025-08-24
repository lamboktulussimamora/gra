package dbcontext

import (
	"reflect"
	"testing"
)

type efEntity struct {
	BaseEntity
	Name string `db:"name" json:"name"`
	Skip string `db:"-" json:"-"`
}

func TestEFContext_Utils(t *testing.T) {
	ctx := NewEFContext(nil)

	if ctx.toSnakeCaseEF("UserAccount") != "user_account" {
		t.Fatalf("toSnakeCaseEF failed")
	}

	tn := ctx.getTableNameFromType(reflect.TypeOf(efEntity{}))
	if tn != "ef_entitys" { // current implementation: snake_case + 's'
		t.Fatalf("unexpected table name: %s", tn)
	}

	field, _ := reflect.TypeOf(efEntity{}).FieldByName("Name")
	if col := ctx.getColumnNameFromField(field); col != tColName {
		t.Fatalf("unexpected column name: %s", col)
	}

	skipField, _ := reflect.TypeOf(efEntity{}).FieldByName("Skip")
	if !ctx.shouldSkipField(skipField) {
		t.Fatalf("expected Skip to be skipped by tags")
	}
}

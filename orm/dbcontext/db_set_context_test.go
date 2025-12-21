package dbcontext

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type ctxUser struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (ctxUser) TableName() string { return "ctx_users" }

func TestEnhancedDbSet_ToListContext_Canceled(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE ctx_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO ctx_users(name) VALUES ('alice')`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[ctxUser](ctx)

	opCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = set.ToListContext(opCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestEnhancedDbSet_CountContext_Canceled(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE ctx_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	ctx := NewEnhancedDbContextWithDB(db)
	set := NewEnhancedDbSet[ctxUser](ctx)

	opCtx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = set.CountContext(opCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

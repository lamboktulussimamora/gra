package main

import "testing"

func TestEnsureMigrationTable_NilDB(t *testing.T) {
	if err := ensureMigrationTable(nil); err == nil {
		t.Fatalf("expected error on nil db")
	}
}

func TestGetAppliedMigrations_NilDB(t *testing.T) {
	if _, err := getAppliedMigrations(nil); err == nil {
		t.Fatalf("expected error on nil db")
	}
}

package main

import (
	"fmt"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/migrations"
)

func TestSimpleHybridMigration(t *testing.T) {
	fmt.Println("=== Simple Hybrid Migration Test ===")

	// Test 1: Create model registry
	fmt.Println("1. Testing ModelRegistry...")
	registry := migrations.NewModelRegistry(migrations.SQLite)

	// Simple test model
	type TestUser struct {
		ID    int64  `db:"id" migration:"primary_key,auto_increment"`
		Email string `db:"email" migration:"unique,not_null,max_length:255"`
		Name  string `db:"name" migration:"not_null,max_length:100"`
	}

	// Register model
	registry.RegisterModel(&TestUser{})
	models := registry.GetModels()

	if len(models) == 0 {
		t.Fatal("Expected at least 1 model to be registered")
	}

	fmt.Printf("   ✓ Registered %d models\n", len(models))
	for tableName, snapshot := range models {
		fmt.Printf("   ✓ Table: %s with %d columns\n", tableName, len(snapshot.Columns))
		
		if len(snapshot.Columns) == 0 {
			t.Errorf("Table %s should have columns", tableName)
		}
		
		for colName, col := range snapshot.Columns {
			fmt.Printf("     - %s: %s (%s)\n", colName, col.Type, col.SQLType)
		}
	}

	// Test 2: Check migration types
	fmt.Println("\n2. Testing MigrationFile...")
	migrationFile := &migrations.MigrationFile{
		Name:        "test_migration",
		Description: "Test migration for demo",
		Changes: []migrations.MigrationChange{
			{
				Type:          migrations.AddColumn,
				TableName:     "users",
				ColumnName:    "phone",
				IsDestructive: false,
			},
		},
	}

	if migrationFile.Name == "" {
		t.Error("Migration name should not be empty")
	}

	if len(migrationFile.Changes) == 0 {
		t.Error("Migration should have at least one change")
	}

	fmt.Printf("   ✓ Migration: %s\n", migrationFile.Name)
	fmt.Printf("   ✓ Has destructive changes: %t\n", migrationFile.HasDestructiveChanges())
	fmt.Printf("   ✓ Warnings: %v\n", migrationFile.GetWarnings())

	// Verify the migration is not destructive for this test case
	if migrationFile.HasDestructiveChanges() {
		t.Error("Test migration should not have destructive changes")
	}

	fmt.Println("\n=== Basic Test Complete ===")
	fmt.Println("Core migration types and registry are working!")
}

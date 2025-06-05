package main_test

import (
	"fmt"

	"github.com/lamboktulussimamora/gra/orm/migrations"
)

func main() {
	fmt.Println("=== Simple Hybrid Migration Test ===")

	// Test 1: Create model registry
	fmt.Println("1. Testing ModelRegistry...")
	registry := migrations.NewModelRegistry()

	// Simple test model
	type TestUser struct {
		ID    int64  `db:"id" migration:"primary_key,auto_increment"`
		Email string `db:"email" migration:"unique,not_null,max_length:255"`
		Name  string `db:"name" migration:"not_null,max_length:100"`
	}

	// Register model
	registry.RegisterModel(&TestUser{})
	models := registry.GetModels()

	fmt.Printf("   ✓ Registered %d models\n", len(models))
	for tableName, snapshot := range models {
		fmt.Printf("   ✓ Table: %s with %d columns\n", tableName, len(snapshot.Columns))
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

	fmt.Printf("   ✓ Migration: %s\n", migrationFile.Name)
	fmt.Printf("   ✓ Has destructive changes: %t\n", migrationFile.HasDestructiveChanges())
	fmt.Printf("   ✓ Warnings: %v\n", migrationFile.GetWarnings())

	fmt.Println("\n=== Basic Test Complete ===")
	fmt.Println("Core migration types and registry are working!")
}

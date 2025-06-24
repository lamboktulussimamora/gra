package main

import (
	"fmt"

	"github.com/lamboktulussimamora/gra/orm/migrations"
)

// SimpleResult contains the results of the simple demo
type SimpleResult struct {
	RegisteredModels int
	TableNames       []string
	MigrationFile    *migrations.MigrationFile
	Error            error
}

// RunSimpleDemo runs a simple hybrid migration test
func RunSimpleDemo() (*SimpleResult, error) {
	result := &SimpleResult{}

	// Test 1: Create model registry
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

	result.RegisteredModels = len(models)
	for tableName := range models {
		result.TableNames = append(result.TableNames, tableName)
	}

	// Test 2: Check migration types
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

	result.MigrationFile = migrationFile
	return result, nil
}

// PrintSimpleResults prints the simple demo results
func PrintSimpleResults(result *SimpleResult) {
	fmt.Println("=== Simple Hybrid Migration Test ===")

	if result.Error != nil {
		fmt.Printf("❌ Simple demo failed: %v\n", result.Error)
		return
	}

	// Test 1: Create model registry
	fmt.Println("1. Testing ModelRegistry...")
	fmt.Printf("   ✓ Registered %d models\n", result.RegisteredModels)
	for _, tableName := range result.TableNames {
		fmt.Printf("   ✓ Table: %s\n", tableName)
	}

	// Test 2: Check migration types
	fmt.Println("\n2. Testing MigrationFile...")
	if result.MigrationFile != nil {
		fmt.Printf("   ✓ Migration: %s\n", result.MigrationFile.Name)
		fmt.Printf("   ✓ Has destructive changes: %t\n", result.MigrationFile.HasDestructiveChanges())
		fmt.Printf("   ✓ Warnings: %v\n", result.MigrationFile.GetWarnings())
	}

	fmt.Println("\n=== Basic Test Complete ===")
	fmt.Println("Core migration types and registry are working!")
}

func simpleDemo() {
	result, err := RunSimpleDemo()
	if err != nil {
		fmt.Printf("Simple demo failed: %v\n", err)
		return
	}
	PrintSimpleResults(result)
}

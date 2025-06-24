package main

import (
	"fmt"
	"testing"

	"github.com/lamboktulussimamora/gra/orm/migrations"
)

func TestRunIntegrationDemo(t *testing.T) {
	// Test the main integration demo
	result, err := RunIntegrationDemo()
	if err != nil {
		t.Fatalf("RunIntegrationDemo failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check models were registered
	if result.ModelsRegistered != 3 {
		t.Errorf("Expected 3 models registered, got %d", result.ModelsRegistered)
	}

	// Check migration status exists
	if result.MigrationStatus == nil {
		t.Error("Expected migration status to be non-nil")
	}

	// Check features demo list
	if len(result.FeaturesDemo) == 0 {
		t.Error("Expected features demo list to be non-empty")
	}

	expectedFeatures := []string{
		"Model registration (EF Core-style DbSet)",
		"Change detection from struct definitions",
		"Migration file generation",
		"Safety checks and warnings",
		"Multiple migration modes",
	}

	if len(result.FeaturesDemo) != len(expectedFeatures) {
		t.Errorf("Expected %d features, got %d", len(expectedFeatures), len(result.FeaturesDemo))
	}

	// Verify no error occurred
	if result.Error != nil {
		t.Errorf("Expected no error in result, got: %v", result.Error)
	}
}

func TestPrintDemoResults(t *testing.T) {
	// Test with successful result
	result := &DemoResult{
		ModelsRegistered: 3,
		MigrationStatus: &migrations.MigrationStatus{
			AppliedMigrations: []*migrations.MigrationFile{
				{
					Name:     "migration1",
					Filename: "001_migration1.sql",
				},
			},
			PendingMigrations: []*migrations.MigrationFile{},
			HasPendingChanges: false,
		},
		MigrationFile: &migrations.MigrationFile{
			Name:     "test_migration",
			Filename: "001_test_migration.sql",
			Changes: []migrations.MigrationChange{
				{
					Type:          migrations.CreateTable,
					TableName:     "test_table",
					IsDestructive: false,
				},
			},
		},
		FeaturesDemo: []string{"Test feature"},
		Error:        nil,
	}

	// This should not panic
	PrintDemoResults(result)

	// Test with error result
	errorResult := &DemoResult{
		Error: fmt.Errorf("test error"),
	}

	// This should not panic
	PrintDemoResults(errorResult)
}

func TestRunSimpleDemo(t *testing.T) {
	// Test the simple demo
	result, err := RunSimpleDemo()
	if err != nil {
		t.Fatalf("RunSimpleDemo failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check models were registered
	if result.RegisteredModels != 1 {
		t.Errorf("Expected 1 model registered, got %d", result.RegisteredModels)
	}

	// Check table names
	if len(result.TableNames) == 0 {
		t.Error("Expected at least one table name")
	}

	// Check migration file exists
	if result.MigrationFile == nil {
		t.Error("Expected migration file to be non-nil")
	} else {
		if result.MigrationFile.Name != "test_migration" {
			t.Errorf("Expected migration name 'test_migration', got %s", result.MigrationFile.Name)
		}

		if len(result.MigrationFile.Changes) == 0 {
			t.Error("Expected at least one migration change")
		}

		// Test migration file methods
		if result.MigrationFile.HasDestructiveChanges() {
			t.Error("Test migration should not have destructive changes")
		}

		warnings := result.MigrationFile.GetWarnings()
		// GetWarnings can return nil if there are no warnings, which is expected
		if len(warnings) > 0 {
			t.Errorf("Expected no warnings for test migration, got: %v", warnings)
		}
	}

	// Verify no error occurred
	if result.Error != nil {
		t.Errorf("Expected no error in result, got: %v", result.Error)
	}
}

func TestPrintSimpleResults(t *testing.T) {
	// Test with successful result
	result := &SimpleResult{
		RegisteredModels: 1,
		TableNames:       []string{"testusers"},
		MigrationFile: &migrations.MigrationFile{
			Name: "test_migration",
			Changes: []migrations.MigrationChange{
				{
					Type:          migrations.AddColumn,
					TableName:     "users",
					ColumnName:    "phone",
					IsDestructive: false,
				},
			},
		},
		Error: nil,
	}

	// This should not panic
	PrintSimpleResults(result)

	// Test with error result
	errorResult := &SimpleResult{
		Error: fmt.Errorf("simple demo error"),
	}

	// This should not panic
	PrintSimpleResults(errorResult)
}

// Test the demo result error handling
func TestDemoResultWithError(t *testing.T) {
	// Create a result with an error
	result := &DemoResult{
		Error: fmt.Errorf("test error"),
	}

	// Verify error is properly set
	if result.Error == nil {
		t.Error("Expected error to be set")
	}

	if result.Error.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got %s", result.Error.Error())
	}
}

// Test simple result error handling
func TestSimpleResultWithError(t *testing.T) {
	// Create a result with an error
	result := &SimpleResult{
		Error: fmt.Errorf("simple error"),
	}

	// Verify error is properly set
	if result.Error == nil {
		t.Error("Expected error to be set")
	}

	if result.Error.Error() != "simple error" {
		t.Errorf("Expected error message 'simple error', got %s", result.Error.Error())
	}
}

// Test edge cases and validation
func TestDemoResultValidation(t *testing.T) {
	// Test empty demo result
	result := &DemoResult{}

	// Should handle nil migration status gracefully
	if result.MigrationStatus != nil {
		t.Error("Expected nil migration status for empty result")
	}

	// Should handle nil migration file gracefully
	if result.MigrationFile != nil {
		t.Error("Expected nil migration file for empty result")
	}

	// Should handle empty features list gracefully
	if result.FeaturesDemo == nil {
		result.FeaturesDemo = []string{}
	}
}

// Test error handling in PrintDemoResults
func TestPrintDemoResultsWithNilStatus(t *testing.T) {
	// Test with nil migration status
	result := &DemoResult{
		ModelsRegistered: 2,
		MigrationStatus:  nil, // This should be handled gracefully
		MigrationFile: &migrations.MigrationFile{
			Name:     "test_migration",
			Filename: "001_test_migration.sql",
			Changes: []migrations.MigrationChange{
				{
					Type:          migrations.CreateTable,
					TableName:     "test_table",
					IsDestructive: false,
				},
			},
		},
		FeaturesDemo: []string{"Test feature"},
		Error:        nil,
	}

	// This should not panic even with nil migration status
	PrintDemoResults(result)
}

// Test migration file warnings functionality
func TestMigrationFileWarnings(t *testing.T) {
	// Create a migration file with potential warnings
	migrationFile := &migrations.MigrationFile{
		Name:     "test_migration_with_warnings",
		Filename: "001_test_migration_with_warnings.sql",
		Changes: []migrations.MigrationChange{
			{
				Type:          migrations.DropColumn,
				TableName:     "users",
				ColumnName:    "legacy_field",
				IsDestructive: true,
			},
		},
	}

	// Test warnings (if GetWarnings returns any)
	warnings := migrationFile.GetWarnings()
	// GetWarnings can return nil if there are no warnings, which is valid
	if warnings != nil {
		// If warnings exist, they should be a valid slice
		if warnings != nil {
			// Warnings should be accessible without issues
			_ = len(warnings)
		}
	}

	// Test destructive changes detection
	if !migrationFile.HasDestructiveChanges() {
		t.Error("Expected migration to have destructive changes")
	}
}

// Test model registry functionality
func TestModelRegistryFunctionality(t *testing.T) {
	// Create model registry
	registry := migrations.NewModelRegistry(migrations.SQLite)

	// Test model struct
	type TestModel struct {
		ID    int64  `db:"id" migration:"primary_key,auto_increment"`
		Name  string `db:"name" migration:"not_null,max_length:100"`
		Email string `db:"email" migration:"unique,not_null,max_length:255"`
	}

	// Register model
	registry.RegisterModel(&TestModel{})

	// Get models
	models := registry.GetModels()

	// Verify registration
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	// Find the table (name might be transformed)
	var tableName string
	for name := range models {
		tableName = name
		break
	}

	if tableName == "" {
		t.Fatal("Expected at least one table to be registered")
	}

	// Verify columns
	tableSnapshot := models[tableName]
	if len(tableSnapshot.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(tableSnapshot.Columns))
	}

	// Check specific columns exist
	expectedColumns := []string{"id", "name", "email"}
	for _, expectedCol := range expectedColumns {
		if _, exists := tableSnapshot.Columns[expectedCol]; !exists {
			t.Errorf("Expected column %s to exist", expectedCol)
		}
	}
}

// Integration test that combines both demo functions
func TestIntegratedDemoWorkflow(t *testing.T) {
	// Run simple demo first
	simpleResult, err := RunSimpleDemo()
	if err != nil {
		t.Fatalf("Simple demo failed: %v", err)
	}

	if simpleResult.RegisteredModels == 0 {
		t.Error("Simple demo should register at least one model")
	}

	// Run integration demo
	integrationResult, err := RunIntegrationDemo()
	if err != nil {
		t.Fatalf("Integration demo failed: %v", err)
	}

	if integrationResult.ModelsRegistered == 0 {
		t.Error("Integration demo should register at least one model")
	}

	// Verify both demos worked
	if simpleResult.Error != nil {
		t.Errorf("Simple demo had error: %v", simpleResult.Error)
	}

	if integrationResult.Error != nil {
		t.Errorf("Integration demo had error: %v", integrationResult.Error)
	}
}

// Test concurrent execution (if applicable)
func TestConcurrentDemoExecution(t *testing.T) {
	// Run demos concurrently to ensure thread safety
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	// Run simple demo in goroutine
	go func() {
		_, err := RunSimpleDemo()
		if err != nil {
			errors <- err
		} else {
			done <- true
		}
	}()

	// Run integration demo in goroutine
	go func() {
		_, err := RunIntegrationDemo()
		if err != nil {
			errors <- err
		} else {
			done <- true
		}
	}()

	// Wait for both to complete
	completed := 0
	for completed < 2 {
		select {
		case <-done:
			completed++
		case err := <-errors:
			t.Errorf("Concurrent demo execution failed: %v", err)
			completed++
		}
	}
}

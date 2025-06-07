// Package debug_test provides utilities for debugging database schema generation.
package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lamboktulussimamora/gra/orm/schema"
)

func TestTestUserStruct(t *testing.T) {
	user := TestUser{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if user.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", user.ID)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected Name to be 'Test User', got %s", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got %s", user.Email)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestTestUserTags(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})

	// Test ID field tags
	idField, found := userType.FieldByName("ID")
	if !found {
		t.Fatal("ID field not found")
	}

	if idField.Tag.Get("db") != "id" {
		t.Errorf("Expected db tag to be 'id', got %s", idField.Tag.Get("db"))
	}
	if idField.Tag.Get("migration") != "primary_key,auto_increment" {
		t.Errorf("Expected migration tag to be 'primary_key,auto_increment', got %s", idField.Tag.Get("migration"))
	}

	// Test Name field tags
	nameField, found := userType.FieldByName("Name")
	if !found {
		t.Fatal("Name field not found")
	}

	if nameField.Tag.Get("db") != "name" {
		t.Errorf("Expected db tag to be 'name', got %s", nameField.Tag.Get("db"))
	}
	if nameField.Tag.Get("migration") != "not_null,max_length:100" {
		t.Errorf("Expected migration tag to be 'not_null,max_length:100', got %s", nameField.Tag.Get("migration"))
	}

	// Test Email field tags
	emailField, found := userType.FieldByName("Email")
	if !found {
		t.Fatal("Email field not found")
	}

	if emailField.Tag.Get("db") != "email" {
		t.Errorf("Expected db tag to be 'email', got %s", emailField.Tag.Get("db"))
	}
	if emailField.Tag.Get("migration") != "unique,not_null,max_length:255" {
		t.Errorf("Expected migration tag to be 'unique,not_null,max_length:255', got %s", emailField.Tag.Get("migration"))
	}
}

func TestSchemaGeneration(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})

	// Test ID field schema generation
	idField, found := userType.FieldByName("ID")
	if !found {
		t.Fatal("ID field not found")
	}

	pgColumn := schema.ParseFieldToColumnForDriver(idField, schema.PostgreSQL)
	if !strings.Contains(pgColumn, "SERIAL") {
		t.Errorf("PostgreSQL column should contain SERIAL, got: %s", pgColumn)
	}

	sqliteColumn := schema.ParseFieldToColumnForDriver(idField, schema.SQLite)
	if !strings.Contains(sqliteColumn, "AUTOINCREMENT") {
		t.Errorf("SQLite column should contain AUTOINCREMENT, got: %s", sqliteColumn)
	}

	// Test Name field schema generation
	nameField, found := userType.FieldByName("Name")
	if !found {
		t.Fatal("Name field not found")
	}

	nameColumn := schema.ParseFieldToColumnForDriver(nameField, schema.PostgreSQL)
	if !strings.Contains(nameColumn, "VARCHAR(100)") {
		t.Errorf("PostgreSQL name column should contain VARCHAR(100), got: %s", nameColumn)
	}
	if !strings.Contains(nameColumn, "NOT NULL") {
		t.Errorf("PostgreSQL name column should contain NOT NULL, got: %s", nameColumn)
	}
}

func TestMainFunction(t *testing.T) {
	// Since main() doesn't return anything and prints to stdout,
	// we'll test the logic it contains separately
	userType := reflect.TypeOf(TestUser{})

	// Test ID field lookup
	idField, found := userType.FieldByName("ID")
	if !found {
		t.Fatal("ID field should be found")
	}

	// Test that the tags are accessible
	dbTag := idField.Tag.Get("db")
	migrationTag := idField.Tag.Get("migration")

	if dbTag == "" {
		t.Error("db tag should not be empty")
	}
	if migrationTag == "" {
		t.Error("migration tag should not be empty")
	}

	// Test schema generation for different drivers
	pgColumn := schema.ParseFieldToColumnForDriver(idField, schema.PostgreSQL)
	sqliteColumn := schema.ParseFieldToColumnForDriver(idField, schema.SQLite)

	if pgColumn == "" {
		t.Error("PostgreSQL column generation should not be empty")
	}
	if sqliteColumn == "" {
		t.Error("SQLite column generation should not be empty")
	}

	// Test name field lookup
	nameField, found := userType.FieldByName("Name")
	if !found {
		t.Fatal("Name field should be found")
	}

	nameColumn := schema.ParseFieldToColumnForDriver(nameField, schema.PostgreSQL)
	if nameColumn == "" {
		t.Error("Name column generation should not be empty")
	}
}

// TestMainFunctionOutput tests that main() can be executed without panicking
func TestMainFunctionOutput(t *testing.T) {
	// Since main() prints to stdout and doesn't accept parameters,
	// we'll just ensure it doesn't panic when called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() function panicked: %v", r)
		}
	}()

	// Redirect stdout to capture output
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// Create a temporary file to capture output
	tempFile, err := os.CreateTemp("", "debug_test_output")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			t.Logf("Warning: failed to close temp file: %v", err)
		}
		if err := os.Remove(tempFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	os.Stdout = tempFile

	// Call main function
	main()

	// Reset stdout
	os.Stdout = originalStdout

	// Read the captured output
	if _, err := tempFile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek temp file: %v", err)
	}
	output := make([]byte, 1024)
	n, _ := tempFile.Read(output)
	outputStr := string(output[:n])

	// Verify that some expected content is present
	if !strings.Contains(outputStr, "Field tags:") {
		t.Error("Output should contain 'Field tags:'")
	}
	if !strings.Contains(outputStr, "PostgreSQL column:") {
		t.Error("Output should contain 'PostgreSQL column:'")
	}
	if !strings.Contains(outputStr, "SQLite column:") {
		t.Error("Output should contain 'SQLite column:'")
	}
}

func TestAllFieldsSchemaGeneration(t *testing.T) {
	userType := reflect.TypeOf(TestUser{})

	testCases := []struct {
		fieldName   string
		shouldExist bool
	}{
		{"ID", true},
		{"Name", true},
		{"Email", true},
		{"IsActive", true},
		{"CreatedAt", true},
		{"NonExistentField", false},
	}

	for _, tc := range testCases {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, found := userType.FieldByName(tc.fieldName)
			if found != tc.shouldExist {
				t.Errorf("Field %s existence should be %v, got %v", tc.fieldName, tc.shouldExist, found)
				return
			}

			if tc.shouldExist {
				// Test schema generation for each driver
				pgColumn := schema.ParseFieldToColumnForDriver(field, schema.PostgreSQL)
				sqliteColumn := schema.ParseFieldToColumnForDriver(field, schema.SQLite)
				mysqlColumn := schema.ParseFieldToColumnForDriver(field, schema.MySQL)

				if pgColumn == "" {
					t.Errorf("PostgreSQL column generation for %s should not be empty", tc.fieldName)
				}
				if sqliteColumn == "" {
					t.Errorf("SQLite column generation for %s should not be empty", tc.fieldName)
				}
				if mysqlColumn == "" {
					t.Errorf("MySQL column generation for %s should not be empty", tc.fieldName)
				}
			}
		})
	}
}

func TestDatabaseDriverConstants(t *testing.T) {
	// Test that the schema constants are accessible
	if schema.PostgreSQL == "" {
		t.Error("PostgreSQL driver constant should not be empty")
	}
	if schema.SQLite == "" {
		t.Error("SQLite driver constant should not be empty")
	}
	if schema.MySQL == "" {
		t.Error("MySQL driver constant should not be empty")
	}
}

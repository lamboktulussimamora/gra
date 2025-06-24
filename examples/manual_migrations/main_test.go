package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestManualMigrationsStructure tests the manual migrations example structure
func TestManualMigrationsStructure(t *testing.T) {
	t.Run("RequiredFilesExist", func(t *testing.T) {
		// Check if required files exist
		requiredFiles := []string{
			"README.md",
			"db_migrate_v2.sh",
			"go.mod",
			"go.sum",
		}

		for _, file := range requiredFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Required file %s does not exist", file)
			}
		}
	})

	t.Run("ShellScriptIsExecutable", func(t *testing.T) {
		scriptPath := "db_migrate_v2.sh"
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Fatalf("Failed to get file info for %s: %v", scriptPath, err)
		}

		// Check if the file has execute permissions
		mode := info.Mode()
		if mode&0111 == 0 {
			t.Errorf("Shell script %s is not executable", scriptPath)
		}
	})

	t.Run("ReadmeHasContent", func(t *testing.T) {
		content, err := os.ReadFile("README.md")
		if err != nil {
			t.Fatalf("Failed to read README.md: %v", err)
		}

		if len(content) == 0 {
			t.Errorf("README.md is empty")
		}

		// Check for key sections in the README
		readmeContent := string(content)
		expectedSections := []string{
			"Manual Migrations Example",
			"Quick Start",
			"Available Commands",
		}

		for _, section := range expectedSections {
			if !containsIgnoreCase(readmeContent, section) {
				t.Errorf("README.md does not contain expected section: %s", section)
			}
		}
	})

	t.Run("GoModHasCorrectModule", func(t *testing.T) {
		content, err := os.ReadFile("go.mod")
		if err != nil {
			t.Fatalf("Failed to read go.mod: %v", err)
		}

		goModContent := string(content)
		if !containsIgnoreCase(goModContent, "module") {
			t.Errorf("go.mod does not contain module declaration")
		}
	})
}

// TestShellScriptStructure tests the shell script structure and content
func TestShellScriptStructure(t *testing.T) {
	t.Run("ShellScriptHasCorrectShebang", func(t *testing.T) {
		content, err := os.ReadFile("db_migrate_v2.sh")
		if err != nil {
			t.Fatalf("Failed to read db_migrate_v2.sh: %v", err)
		}

		scriptContent := string(content)
		if len(scriptContent) == 0 {
			t.Fatalf("Shell script is empty")
		}

		// Check for proper shebang
		if !stringStartsWith(scriptContent, "#!/bin/bash") {
			t.Errorf("Shell script does not start with proper shebang")
		}
	})

	t.Run("ShellScriptHasRequiredCommands", func(t *testing.T) {
		content, err := os.ReadFile("db_migrate_v2.sh")
		if err != nil {
			t.Fatalf("Failed to read db_migrate_v2.sh: %v", err)
		}

		scriptContent := string(content)
		expectedCommands := []string{
			"up",
			"down",
			"status",
			"test",
		}

		for _, cmd := range expectedCommands {
			if !containsIgnoreCase(scriptContent, cmd) {
				t.Errorf("Shell script does not contain expected command: %s", cmd)
			}
		}
	})

	t.Run("ShellScriptHasUsageFunction", func(t *testing.T) {
		content, err := os.ReadFile("db_migrate_v2.sh")
		if err != nil {
			t.Fatalf("Failed to read db_migrate_v2.sh: %v", err)
		}

		scriptContent := string(content)
		if !containsIgnoreCase(scriptContent, "usage()") {
			t.Errorf("Shell script does not contain usage function")
		}
	})
}

// TestMigrationToolsReference tests references to migration tools
func TestMigrationToolsReference(t *testing.T) {
	t.Run("ReadmeReferencesTools", func(t *testing.T) {
		content, err := os.ReadFile("README.md")
		if err != nil {
			t.Fatalf("Failed to read README.md: %v", err)
		}

		readmeContent := string(content)
		expectedReferences := []string{
			"direct_runner",
			"test_runner",
			"tools/migration",
		}

		for _, ref := range expectedReferences {
			if !containsIgnoreCase(readmeContent, ref) {
				t.Errorf("README.md does not reference expected tool: %s", ref)
			}
		}
	})

	t.Run("ToolsDirectoryStructure", func(t *testing.T) {
		// Check if the tools directory exists relative to the project root
		toolsPath := filepath.Join("..", "..", "tools", "migration")
		if _, err := os.Stat(toolsPath); os.IsNotExist(err) {
			t.Logf("Tools directory %s does not exist - this might be expected", toolsPath)
			// This is not a failure since the tools might be organized differently
		} else {
			t.Logf("Tools directory %s exists", toolsPath)
		}
	})
}

// TestDatabaseConfiguration tests database configuration handling
func TestDatabaseConfiguration(t *testing.T) {
	t.Run("ShellScriptHasDBVariables", func(t *testing.T) {
		content, err := os.ReadFile("db_migrate_v2.sh")
		if err != nil {
			t.Fatalf("Failed to read db_migrate_v2.sh: %v", err)
		}

		scriptContent := string(content)
		expectedVariables := []string{
			"DB_HOST",
			"DB_PORT",
			"DB_USER",
			"DB_PASSWORD",
			"DB_NAME",
		}

		for _, variable := range expectedVariables {
			if !containsIgnoreCase(scriptContent, variable) {
				t.Errorf("Shell script does not contain expected DB variable: %s", variable)
			}
		}
	})

	t.Run("ReadmeDocumentsConfiguration", func(t *testing.T) {
		content, err := os.ReadFile("README.md")
		if err != nil {
			t.Fatalf("Failed to read README.md: %v", err)
		}

		readmeContent := string(content)
		if !containsIgnoreCase(readmeContent, "Database Configuration") {
			t.Errorf("README.md does not document database configuration")
		}
	})
}

// Helper functions
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && indexIgnoreCase(s, substr) >= 0
}

func indexIgnoreCase(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return i
		}
	}
	return -1
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func stringStartsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

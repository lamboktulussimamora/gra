package main

import (
	"strings"
	"testing"
)

func TestBuildPostgreSQLConnectionString(t *testing.T) {
	cs := buildPostgreSQLConnectionString(CLIConfig{
		Host:     "db.local",
		Port:     "55432",
		User:     "postgres",
		Password: "MyPassword_123",
		Database: "gra_test",
		SSLMode:  "disable",
	})
	want := "postgres://postgres:MyPassword_123@db.local:55432/gra_test?sslmode=disable"
	if cs != want {
		t.Fatalf("want %q, got %q", want, cs)
	}
}

func TestSanitizeConnectionString(t *testing.T) {
	in := "postgres://user:secret@localhost:5432/db?sslmode=disable"
	out := sanitizeConnectionString(in)
	if strings.Contains(out, "secret") || !strings.Contains(out, "*****") {
		t.Fatalf("password should be masked, got %q", out)
	}
}

func TestExtractDBName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"postgres://u:p@h:5432/mydb?sslmode=disable", "mydb"},
		{"postgres://u:p@h:5432/simple", "simple"},
		{"", ""},
	}
	for _, c := range cases {
		if got := extractDBName(c.in); got != c.want {
			t.Fatalf("extractDBName(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestParseMigrationContent(t *testing.T) {
	content := `-- Migration: test
-- Description: demo
-- UP Migration
CREATE TABLE users(id SERIAL PRIMARY KEY);

-- DOWN Migration (for rollback)
-- DROP TABLE users;`
	up, down := parseMigrationContent(content)
	if !strings.Contains(up, "CREATE TABLE users") {
		t.Fatalf("up not parsed correctly: %q", up)
	}
	if !strings.Contains(down, "DROP TABLE users;") {
		t.Fatalf("down not parsed correctly: %q", down)
	}
}

package main

import "testing"

func TestGetDriver(t *testing.T) {
	if d := getDriver("postgres"); d != "postgres" {
		t.Fatalf("expected postgres, got %q", d)
	}
	if d := getDriver("postgresql"); d != "postgres" {
		t.Fatalf("expected postgres, got %q", d)
	}
	if d := getDriver("sqlite"); d != "sqlite3" {
		t.Fatalf("expected sqlite, got %q", d)
	}
	if d := getDriver("sqlite3"); d != "sqlite3" {
		t.Fatalf("expected sqlite, got %q", d)
	}
	if d := getDriver("mysql"); d != "mysql" {
		t.Fatalf("expected mysql, got %q", d)
	}
}

func TestValidateConfig(t *testing.T) {
	// missing DB
	if err := validateConfig(&Config{}); err == nil {
		t.Fatalf("expected error when db url empty")
	}
	// defaults
	cfg := &Config{DatabaseURL: "postgres://x"}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Driver == "" || cfg.MigrationsDir == "" || cfg.ModelsDir == "" {
		t.Fatalf("expected defaults to be set, got: %#v", cfg)
	}
}

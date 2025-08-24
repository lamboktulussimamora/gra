package main

import "testing"

const (
	drvPostgres = "postgres"
	drvSQLite3  = "sqlite3"
	drvSQLite   = "sqlite"
	drvMySQL    = "mysql"
)

func TestGetDriver(t *testing.T) {
	if d := getDriver(drvPostgres); d != drvPostgres {
		t.Fatalf("expected postgres, got %q", d)
	}
	if d := getDriver("postgresql"); d != drvPostgres {
		t.Fatalf("expected postgres, got %q", d)
	}
	if d := getDriver(drvSQLite); d != drvSQLite3 {
		t.Fatalf("expected sqlite, got %q", d)
	}
	if d := getDriver(drvSQLite3); d != drvSQLite3 {
		t.Fatalf("expected sqlite, got %q", d)
	}
	if d := getDriver(drvMySQL); d != drvMySQL {
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

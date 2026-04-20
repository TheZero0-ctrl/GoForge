package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigUsesDSNFlagFirst(t *testing.T) {
	cfg, err := ParseConfig(testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/flag_db?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.DatabaseName != "flag_db" {
		t.Fatalf("expected database flag_db, got %q", cfg.DatabaseName)
	}
}

func TestParseConfigFallsBackToDatabaseToml(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.MkdirAll(filepath.Join("config"), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	content := []byte("[development]\ndsn = \"postgres://localhost:5432/file_db?sslmode=disable\"\n")
	if err := os.WriteFile(filepath.Join("config", "database.toml"), content, 0o644); err != nil {
		t.Fatalf("write database.toml: %v", err)
	}

	cfg, err := ParseConfig(testParams{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.DatabaseName != "file_db" {
		t.Fatalf("expected database file_db, got %q", cfg.DatabaseName)
	}

}

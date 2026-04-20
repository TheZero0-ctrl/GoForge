//go:build e2e

package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goforge/test/testutil/e2e"
)

func TestDBMigrateCommandE2EAppliesMigrationWhenDSNProvided(t *testing.T) {
	dsn := os.Getenv("GOFORGE_E2E_DATABASE_URL")
	if dsn == "" {
		t.Skip("set GOFORGE_E2E_DATABASE_URL to run db:migrate apply integration test")
	}

	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	if err := os.MkdirAll(filepath.Join(workspace, "cmd", "api"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/api: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workspace, "cmd", "api", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write cmd/api/main.go: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(workspace, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}

	version := time.Now().UTC().Format("20060102150405")
	tableName := fmt.Sprintf("goforge_e2e_%d", time.Now().UTC().UnixNano())

	up := []byte(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id bigserial PRIMARY KEY);\n", tableName))
	down := []byte(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", tableName))

	upPath := filepath.Join(workspace, "migrations", version+"_create_"+tableName+".up.sql")
	downPath := filepath.Join(workspace, "migrations", version+"_create_"+tableName+".down.sql")

	if err := os.WriteFile(upPath, up, 0o644); err != nil {
		t.Fatalf("write up migration: %v", err)
	}

	if err := os.WriteFile(downPath, down, 0o644); err != nil {
		t.Fatalf("write down migration: %v", err)
	}

	result := e2e.Run(t, binary, workspace, "db:migrate", "--dsn", dsn)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "DONE db:migrate finished")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

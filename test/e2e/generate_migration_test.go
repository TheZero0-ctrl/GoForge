//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goforge/test/testutil/e2e"
)

func TestGenerateMigrationAliasDryRunPrintsMigrationPairPlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "g", "migration", "create_users", "name:string")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Generate migration files")
	e2e.AssertContains(t, result.Stdout, "PLAN MKDIR migrations")
	e2e.AssertContains(t, result.Stdout, ".up.sql")
	e2e.AssertContains(t, result.Stdout, ".down.sql")

	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

func TestGenerateMigrationCreatePatternWritesScaffoldedSQL(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "g", "migration", "create_users", "name:string")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	upPath, downPath := findMigrationPair(t, filepath.Join(workspace, "migrations"), "create_users")

	upData, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	downData, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	e2e.AssertContains(t, string(upData), "CREATE TABLE IF NOT EXISTS \"users\"")
	e2e.AssertContains(t, string(upData), "\"name\" text")
	e2e.AssertContains(t, string(downData), "DROP TABLE IF EXISTS \"users\"")
}

func TestGenerateMigrationCustomNameWritesEmptyFiles(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "g", "migration", "backfill_users")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	upPath, downPath := findMigrationPair(t, filepath.Join(workspace, "migrations"), "backfill_users")

	upData, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	downData, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	if len(upData) != 0 || len(downData) != 0 {
		t.Fatalf("expected empty migration files for custom name, got up=%q down=%q", string(upData), string(downData))
	}
}

func TestGenerateMigrationRemovePatternWithoutTypeKeepsDownEmpty(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "g", "migration", "remove_email_from_users")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	upPath, downPath := findMigrationPair(t, filepath.Join(workspace, "migrations"), "remove_email_from_users")

	upData, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	downData, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	e2e.AssertContains(t, string(upData), "ALTER TABLE \"users\" DROP COLUMN IF EXISTS \"email\";")
	if len(downData) != 0 {
		t.Fatalf("expected empty down migration when remove pattern has no type, got %q", string(downData))
	}
}

func TestGenerateMigrationAddPatternQuotesReservedIdentifier(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "g", "migration", "add_desc_to_products", "desc:string")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	upPath, _ := findMigrationPair(t, filepath.Join(workspace, "migrations"), "add_desc_to_products")

	upData, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	e2e.AssertContains(t, string(upData), "ALTER TABLE \"products\" ADD COLUMN IF NOT EXISTS \"desc\" text;")
}

func findMigrationPair(t *testing.T, migrationsDir, name string) (string, string) {
	t.Helper()

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	upPath := ""
	downPath := ""

	for _, entry := range entries {
		filename := entry.Name()
		if strings.HasSuffix(filename, "_"+name+".up.sql") {
			upPath = filepath.Join(migrationsDir, filename)
		}
		if strings.HasSuffix(filename, "_"+name+".down.sql") {
			downPath = filepath.Join(migrationsDir, filename)
		}
	}

	if upPath == "" || downPath == "" {
		t.Fatalf("expected migration pair for %q, got up=%q down=%q", name, upPath, downPath)
	}

	return upPath, downPath
}

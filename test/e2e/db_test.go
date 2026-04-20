//go:build e2e

package e2e_test

import (
	"testing"

	"goforge/test/testutil/e2e"
)

func TestDBCreateCommandE2EDryRunPrintsCreatePlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "db:create", "--dsn", "postgres://localhost:5432/demo_app?sslmode=disable")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Create database")
	e2e.AssertContains(t, result.Stdout, "CREATE DATABASE \"demo_app\";")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

func TestDBDropCommandE2EDryRunPrintsDropPlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "db:drop", "--dsn", "postgres://localhost:5432/demo_app?sslmode=disable")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Drop database")
	e2e.AssertContains(t, result.Stdout, "DROP DATABASE IF EXISTS \"demo_app\";")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

func TestDBMigrateCommandE2EDryRunPrintsMigratePlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "db:migrate", "--dsn", "postgres://localhost:5432/demo_app?sslmode=disable")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Apply database migrations")
	e2e.AssertContains(t, result.Stdout, "PLAN MIGRATE UP file://")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

func TestDBRollbackCommandE2EDryRunPrintsRollbackPlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "db:rollback", "2", "--dsn", "postgres://localhost:5432/demo_app?sslmode=disable")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Rollback database migrations")
	e2e.AssertContains(t, result.Stdout, "PLAN MIGRATE DOWN 2 file://")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

func TestDBMigrateForceCommandE2EDryRunPrintsForcePlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "--dry-run", "db:migrate:force", "20260420034348", "--dsn", "postgres://localhost:5432/demo_app?sslmode=disable")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Force migration version")
	e2e.AssertContains(t, result.Stdout, "PLAN MIGRATE FORCE 20260420034348 file://")
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
}

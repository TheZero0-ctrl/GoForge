//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"goforge/test/testutil/e2e"
)

func TestNewCommandE2EScaffoldsApp(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	result := e2e.Run(t, binary, workspace, "new", "demo-api", "--skip-git", "--skip-tidy")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	appRoot := filepath.Join(workspace, "demo-api")
	e2e.AssertFileExists(t, filepath.Join(appRoot, "go.mod"))
	e2e.AssertFileExists(t, filepath.Join(appRoot, "cmd", "api", "main.go"))
	e2e.AssertFileExists(t, filepath.Join(appRoot, "README.md"))
	e2e.AssertFileExists(t, filepath.Join(appRoot, ".gitignore"))
	e2e.AssertFileExists(t, filepath.Join(appRoot, "migrations", ".keep"))

	goMod, err := os.ReadFile(filepath.Join(appRoot, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	e2e.AssertContains(t, string(goMod), "module demo-api")
}

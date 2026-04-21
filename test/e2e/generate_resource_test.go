//go:build e2e

package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"goforge/test/testutil/e2e"
)

func TestGenerateResourceCreatesFilesAndBuilds(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	newResult := e2e.Run(t, binary, workspace, "new", "demo-api", "--skip-git")
	if newResult.ExitCode != 0 {
		t.Fatalf("expected new exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", newResult.ExitCode, newResult.Stdout, newResult.Stderr)
	}

	appRoot := filepath.Join(workspace, "demo-api")

	genResult := e2e.Run(t, binary, appRoot, "g", "resource", "movie", "title:string", "year:int", "genres:string[]")
	if genResult.ExitCode != 0 {
		t.Fatalf("expected generate resource exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", genResult.ExitCode, genResult.Stdout, genResult.Stderr)
	}

	e2e.AssertFileExists(t, filepath.Join(appRoot, "internal", "data", "movies.go"))
	e2e.AssertFileExists(t, filepath.Join(appRoot, "cmd", "api", "movies.go"))
	_, _ = findResourceMigrationPair(t, filepath.Join(appRoot, "migrations"), "create_movies")

	routesData, err := os.ReadFile(filepath.Join(appRoot, "cmd", "api", "routes.go"))
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	e2e.AssertContains(t, string(routesData), "router.HandlerFunc(http.MethodGet, \"/v1/movies\", app.listMoviesHandler)")
	e2e.AssertContains(t, string(routesData), "router.HandlerFunc(http.MethodPost, \"/v1/movies\", app.createMovieHandler)")

	modelsData, err := os.ReadFile(filepath.Join(appRoot, "internal", "data", "models.go"))
	if err != nil {
		t.Fatalf("read models.go: %v", err)
	}
	e2e.AssertContains(t, string(modelsData), "Movies MovieModel")

	resourceData, err := os.ReadFile(filepath.Join(appRoot, "internal", "data", "movies.go"))
	if err != nil {
		t.Fatalf("read movies.go: %v", err)
	}
	if !strings.Contains(string(resourceData), "CreatedAt") {
		t.Fatalf("expected generated resource struct to include CreatedAt, got:\n%s", string(resourceData))
	}
	if !strings.Contains(string(resourceData), "Version") {
		t.Fatalf("expected generated resource struct to include Version, got:\n%s", string(resourceData))
	}
	e2e.AssertContains(t, string(resourceData), "func (m MovieModel) GetAll() ([]*Movie, error)")

	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = appRoot
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected generated app to compile: %v\n%s", err, string(buildOut))
	}
}

func TestGenerateResourceDryRunPrintsCreateAndUpdatePlan(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	newResult := e2e.Run(t, binary, workspace, "new", "demo-api", "--skip-git")
	if newResult.ExitCode != 0 {
		t.Fatalf("expected new exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", newResult.ExitCode, newResult.Stdout, newResult.Stderr)
	}

	appRoot := filepath.Join(workspace, "demo-api")
	result := e2e.Run(t, binary, appRoot, "--dry-run", "g", "resource", "movie", "title:string")
	if result.ExitCode != 0 {
		t.Fatalf("expected dry-run exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", result.ExitCode, result.Stdout, result.Stderr)
	}

	e2e.AssertContains(t, result.Stdout, "INFO DRY-RUN Generate resource files and migration")
	e2e.AssertContains(t, result.Stdout, "PLAN WRITE internal/data/movies.go")
	e2e.AssertContains(t, result.Stdout, "PLAN WRITE cmd/api/movies.go")
	e2e.AssertContains(t, result.Stdout, "PLAN UPDATE cmd/api/routes.go")
	e2e.AssertContains(t, result.Stdout, "PLAN UPDATE internal/data/models.go")
	e2e.AssertContains(t, result.Stdout, "_create_movies.up.sql")
	e2e.AssertContains(t, result.Stdout, "_create_movies.down.sql")
}

func TestGenerateResourceConflictsWhenCreateMigrationAlreadyExists(t *testing.T) {
	repoRoot := e2e.RepoRoot(t)
	binary := e2e.BuildBinary(t, repoRoot)
	workspace := t.TempDir()

	newResult := e2e.Run(t, binary, workspace, "new", "demo-api", "--skip-git")
	if newResult.ExitCode != 0 {
		t.Fatalf("expected new exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", newResult.ExitCode, newResult.Stdout, newResult.Stderr)
	}

	appRoot := filepath.Join(workspace, "demo-api")
	first := e2e.Run(t, binary, appRoot, "g", "resource", "movie", "title:string")
	if first.ExitCode != 0 {
		t.Fatalf("expected first run success, got %d\nstdout:\n%s\nstderr:\n%s", first.ExitCode, first.Stdout, first.Stderr)
	}

	second := e2e.Run(t, binary, appRoot, "g", "resource", "movie", "title:string")
	if second.ExitCode != 3 {
		t.Fatalf("expected conflict exit code 3, got %d\nstdout:\n%s\nstderr:\n%s", second.ExitCode, second.Stdout, second.Stderr)
	}

	e2e.AssertContains(t, second.Stderr, "conflict: create migration for \"movies\" already exists")
}

func findResourceMigrationPair(t *testing.T, migrationsDir, name string) (string, string) {
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

package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func RepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}

	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root containing go.mod")
		}
		dir = parent
	}
}

func BuildBinary(t *testing.T, repoRoot string) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "goforge-e2e-bin")
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/goforge")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build goforge binary: %v\n%s", err, string(out))
	}

	return binPath
}

func Run(t *testing.T, binaryPath, cwd string, args ...string) Result {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = cwd

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run command failed: %v", err)
		}
	}

	return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code}
}

func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (%v)", path, err)
	}
}

func AssertContains(t *testing.T, got, want string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

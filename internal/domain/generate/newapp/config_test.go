package newapp

import "testing"

type testParams struct {
	values map[string]string
}

func (p testParams) Param(key string) string {
	return p.values[key]
}

func (p testParams) BoolParam(key string) bool {
	return p.values[key] == "true"
}

func TestParseConfigRequiresSingleArg(t *testing.T) {
	t.Parallel()

	_, err := ParseConfig(nil, testParams{})
	if err == nil {
		t.Fatal("expected error for missing app name")
	}

	_, err = ParseConfig([]string{"one", "two"}, testParams{})
	if err == nil {
		t.Fatal("expected error for extra args")
	}
}

func TestParseConfigRejectsEmptyAppName(t *testing.T) {
	t.Parallel()

	_, err := ParseConfig([]string{"   \t  "}, testParams{})
	if err == nil {
		t.Fatal("expected error for empty app name")
	}

	if got := err.Error(); got != "app name cannot be empty" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestParseConfigDefaultsModuleToAppName(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfig([]string{"demo-api"}, testParams{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.ModulePath != "demo-api" {
		t.Fatalf("expected module to default to app name, got %q", cfg.ModulePath)
	}
}

func TestParseConfigUsesExplicitModule(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfig([]string{"demo-api"}, testParams{values: map[string]string{"module": "github.com/acme/demo-api"}})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.ModulePath != "github.com/acme/demo-api" {
		t.Fatalf("expected explicit module path, got %q", cfg.ModulePath)
	}
}

func TestParseConfigReadsSkipFlags(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfig([]string{"demo-api"}, testParams{values: map[string]string{
		"skip-git":  "true",
		"skip-tidy": "true",
	}})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if !cfg.SkipGit {
		t.Fatal("expected SkipGit true")
	}

	if !cfg.SkipTidy {
		t.Fatal("expected SkipTidy true")
	}
}

func TestParseConfigSanitizesAppName(t *testing.T) {
	t.Parallel()

	cfg, err := ParseConfig([]string{"My App!@#"}, testParams{})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if cfg.NormalizeAppName != "My_App_" {
		t.Fatalf("expected sanitized app name, got %q", cfg.NormalizeAppName)
	}
}

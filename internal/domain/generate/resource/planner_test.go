package resource

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goforge/internal/domain/plan"
)

type testParams struct {
	values map[string]string
}

func (p testParams) Param(key string) string {
	return p.values[key]
}

func (p testParams) BoolParam(key string) bool {
	return p.values[key] == "true"
}

func TestValidateRequiresField(t *testing.T) {
	err := Validate([]string{"movie"}, testParams{})
	if err == nil {
		t.Fatal("expected missing field error")
	}
}

func TestPatchRoutesFileAddsResourceRoutes(t *testing.T) {
	src := []byte(`package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	return app.recoverPanic(router)
}
`)

	names, _, err := parseInput([]string{"movie", "title:string"})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}

	updated, changed, err := patchRoutesFile(src, names)
	if err != nil {
		t.Fatalf("patch routes: %v", err)
	}
	if !changed {
		t.Fatal("expected routes to change")
	}

	text := string(updated)
	if !strings.Contains(text, "router.HandlerFunc(http.MethodGet, \"/v1/movies\", app.listMoviesHandler)") {
		t.Fatalf("expected list route, got:\n%s", text)
	}
	if !strings.Contains(text, "router.HandlerFunc(http.MethodPost, \"/v1/movies\", app.createMovieHandler)") {
		t.Fatalf("expected create route, got:\n%s", text)
	}
}

func TestPatchRoutesFileConflictsOnExistingDifferentHandler(t *testing.T) {
	src := []byte(`package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.healthcheckHandler)
	return app.recoverPanic(router)
}
`)

	names, _, err := parseInput([]string{"movie", "title:string"})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}

	_, _, err = patchRoutesFile(src, names)
	if err == nil || !strings.Contains(err.Error(), "route conflict") {
		t.Fatalf("expected route conflict error, got %v", err)
	}
}

func TestPatchModelsFileAddsModelFieldAndConstructor(t *testing.T) {
	src := []byte(`package data

import "database/sql"

type Models struct {
}

func NewModels(db *sql.DB) Models {
	return Models{}
}
`)

	names, _, err := parseInput([]string{"movie", "title:string"})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}

	updated, changed, err := patchModelsFile(src, names)
	if err != nil {
		t.Fatalf("patch models: %v", err)
	}
	if !changed {
		t.Fatal("expected models to change")
	}

	text := string(updated)
	if !strings.Contains(text, "Movies MovieModel") {
		t.Fatalf("expected models field, got:\n%s", text)
	}
	if !strings.Contains(text, "Movies: MovieModel{DB: db}") {
		t.Fatalf("expected constructor wiring, got:\n%s", text)
	}
}

func TestPatchModelsFileConflictsOnExistingMismatchedType(t *testing.T) {
	src := []byte(`package data

import "database/sql"

type Models struct {
	Movies int
}

func NewModels(db *sql.DB) Models {
	return Models{}
}
`)

	names, _, err := parseInput([]string{"movie", "title:string"})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}

	_, _, err = patchModelsFile(src, names)
	if err == nil || !strings.Contains(err.Error(), "models conflict") {
		t.Fatalf("expected models conflict error, got %v", err)
	}
}

func TestPatchModelsFileConflictsOnExistingMismatchedConstructorValue(t *testing.T) {
	src := []byte(`package data

import "database/sql"

type Models struct {
	Movies MovieModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{},
	}
}
`)

	names, _, err := parseInput([]string{"movie", "title:string"})
	if err != nil {
		t.Fatalf("parse input: %v", err)
	}

	_, _, err = patchModelsFile(src, names)
	if err == nil || !strings.Contains(err.Error(), "models conflict") {
		t.Fatalf("expected models conflict error, got %v", err)
	}
}

func TestPlanGeneratesResourceFilesAndUpdates(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, workspace, "go.mod", "module example.com/demo\n\ngo 1.25.0\n")
	writeFile(t, workspace, "cmd/api/main.go", "package main\n")
	writeFile(t, workspace, "internal/validator/validator.go", "package validator\n")
	writeFile(t, workspace, "cmd/api/routes.go", `package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	return app.recoverPanic(router)
}
`)
	writeFile(t, workspace, "internal/data/models.go", `package data

import "database/sql"

type Models struct {
}

func NewModels(db *sql.DB) Models {
	return Models{}
}
`)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	planned, err := Plan(context.Background(), []string{"movie", "title:string", "genres:string[]", "year:int"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if planned.CommandID != "generate:resource" {
		t.Fatalf("unexpected command id %q", planned.CommandID)
	}

	hasDataFile := false
	hasHandlersFile := false
	hasRoutesUpdate := false
	hasModelsUpdate := false
	hasFmt := false
	hasMigrationDir := false
	hasMigrationUp := false
	hasMigrationDown := false

	for _, op := range planned.Ops {
		switch {
		case op.Type == plan.OpWriteFile && op.Path == filepath.Join("internal", "data", "movies.go"):
			hasDataFile = true
		case op.Type == plan.OpWriteFile && op.Path == filepath.Join("cmd", "api", "movies.go"):
			hasHandlersFile = true
		case op.Type == plan.OpUpdateFile && op.Path == filepath.Join("cmd", "api", "routes.go"):
			hasRoutesUpdate = true
		case op.Type == plan.OpUpdateFile && op.Path == filepath.Join("internal", "data", "models.go"):
			hasModelsUpdate = true
		case op.Type == plan.OpRun && len(op.Cmd) >= 2 && op.Cmd[0] == "gofmt" && op.Cmd[1] == "-w":
			hasFmt = true
		case op.Type == plan.OpMkdir && op.Path == "migrations":
			hasMigrationDir = true
		case op.Type == plan.OpWriteFile && strings.HasSuffix(op.Path, "_create_movies.up.sql"):
			hasMigrationUp = true
		case op.Type == plan.OpWriteFile && strings.HasSuffix(op.Path, "_create_movies.down.sql"):
			hasMigrationDown = true
		}
	}

	if !hasDataFile || !hasHandlersFile || !hasRoutesUpdate || !hasModelsUpdate || !hasFmt || !hasMigrationDir || !hasMigrationUp || !hasMigrationDown {
		t.Fatalf("missing expected operations: %+v", planned.Ops)
	}
}

func TestPlanReturnsConflictOpWhenCreateMigrationExists(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, workspace, "go.mod", "module example.com/demo\n\ngo 1.25.0\n")
	writeFile(t, workspace, "cmd/api/main.go", "package main\n")
	writeFile(t, workspace, "internal/validator/validator.go", "package validator\n")
	writeFile(t, workspace, "cmd/api/routes.go", "package main\n")
	writeFile(t, workspace, "internal/data/models.go", "package data\n")
	writeFile(t, workspace, "migrations/20260421000000_create_movies.up.sql", "")

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	planned, err := Plan(context.Background(), []string{"movie", "title:string"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if len(planned.Ops) != 1 {
		t.Fatalf("expected a single conflict op, got %d", len(planned.Ops))
	}

	op := planned.Ops[0]
	if op.Type != plan.OpEnsureNotExists {
		t.Fatalf("expected ensure_not_exists op, got %q", op.Type)
	}

	if !strings.Contains(op.Message, "conflict") {
		t.Fatalf("expected conflict message, got %q", op.Message)
	}
}

func writeFile(t *testing.T, root, relPath, data string) {
	t.Helper()
	abs := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(data), 0o644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
}

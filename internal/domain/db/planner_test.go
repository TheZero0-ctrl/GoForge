package db

import (
	"context"
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

func TestPlanCreateIncludesExpectedOperations(t *testing.T) {
	t.Parallel()

	planned, err := PlanCreate(context.Background(), nil, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}

	if planned.CommandID != "db:create" {
		t.Fatalf("expected command id db:create, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpEnsureExists || planned.Ops[0].Path != "config/database.toml" {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	if planned.Ops[1].Type != plan.OpRun {
		t.Fatalf("expected second op to be run, got %q", planned.Ops[1].Type)
	}

	want := []string{"psql", "postgres://localhost:5432/postgres?sslmode=disable", "-v", "ON_ERROR_STOP=1", "-c", "CREATE DATABASE \"demo_app\";"}
	if len(planned.Ops[1].Cmd) != len(want) {
		t.Fatalf("unexpected create command length: %v", planned.Ops[1].Cmd)
	}

	for i := range want {
		if planned.Ops[1].Cmd[i] != want[i] {
			t.Fatalf("unexpected create command: %v", planned.Ops[1].Cmd)
		}
	}
}

func TestPlanDropIncludesExpectedOperations(t *testing.T) {
	t.Parallel()

	planned, err := PlanDrop(context.Background(), nil, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("plan drop: %v", err)
	}

	if planned.CommandID != "db:drop" {
		t.Fatalf("expected command id db:drop, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpEnsureExists || planned.Ops[0].Path != "config/database.toml" {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	if planned.Ops[1].Type != plan.OpRun {
		t.Fatalf("expected second op to be run, got %q", planned.Ops[1].Type)
	}

	want := []string{"psql", "postgres://localhost:5432/postgres?sslmode=disable", "-v", "ON_ERROR_STOP=1", "-c", "DROP DATABASE IF EXISTS \"demo_app\";"}
	if len(planned.Ops[1].Cmd) != len(want) {
		t.Fatalf("unexpected drop command length: %v", planned.Ops[1].Cmd)
	}

	for i := range want {
		if planned.Ops[1].Cmd[i] != want[i] {
			t.Fatalf("unexpected drop command: %v", planned.Ops[1].Cmd)
		}
	}
}

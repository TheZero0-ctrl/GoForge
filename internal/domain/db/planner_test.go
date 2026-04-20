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

	if len(planned.Ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpRun {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	want := []string{"psql", "postgres://localhost:5432/postgres?sslmode=disable", "-v", "ON_ERROR_STOP=1", "-c", "CREATE DATABASE \"demo_app\";"}
	if len(planned.Ops[0].Cmd) != len(want) {
		t.Fatalf("unexpected create command length: %v", planned.Ops[0].Cmd)
	}

	for i := range want {
		if planned.Ops[0].Cmd[i] != want[i] {
			t.Fatalf("unexpected create command: %v", planned.Ops[0].Cmd)
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

	if len(planned.Ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpRun {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	want := []string{"psql", "postgres://localhost:5432/postgres?sslmode=disable", "-v", "ON_ERROR_STOP=1", "-c", "DROP DATABASE IF EXISTS \"demo_app\";"}
	if len(planned.Ops[0].Cmd) != len(want) {
		t.Fatalf("unexpected drop command length: %v", planned.Ops[0].Cmd)
	}

	for i := range want {
		if planned.Ops[0].Cmd[i] != want[i] {
			t.Fatalf("unexpected drop command: %v", planned.Ops[0].Cmd)
		}
	}
}

func TestPlanMigrateIncludesExpectedOperations(t *testing.T) {
	t.Parallel()

	planned, err := PlanMigrate(context.Background(), nil, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("plan migrate: %v", err)
	}

	if planned.CommandID != "db:migrate" {
		t.Fatalf("expected command id db:migrate, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpEnsureExists || planned.Ops[0].Path != "cmd/api/main.go" {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	if planned.Ops[1].Type != plan.OpEnsureExists || planned.Ops[1].Path != "migrations" {
		t.Fatalf("unexpected second op: %+v", planned.Ops[1])
	}

	if planned.Ops[2].Type != plan.OpMigrateUp {
		t.Fatalf("expected third op migrate_up, got %q", planned.Ops[2].Type)
	}

	if planned.Ops[2].Params[plan.MigrateParamDatabaseURL] != "postgres://localhost:5432/demo_app?sslmode=disable" {
		t.Fatalf("unexpected migration database URL: %q", planned.Ops[2].Params[plan.MigrateParamDatabaseURL])
	}

	if planned.Ops[2].Params[plan.MigrateParamSourceURL] == "" {
		t.Fatal("expected migration source URL to be set")
	}
}

func TestPlanRollbackIncludesExpectedOperations(t *testing.T) {
	t.Parallel()

	planned, err := PlanRollback(context.Background(), nil, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("plan rollback: %v", err)
	}

	if planned.CommandID != "db:rollback" {
		t.Fatalf("expected command id db:rollback, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[2].Type != plan.OpMigrateDown {
		t.Fatalf("expected third op migrate_down, got %q", planned.Ops[2].Type)
	}

	if planned.Ops[2].Params[plan.MigrateParamSteps] != "1" {
		t.Fatalf("expected default rollback steps to be 1, got %q", planned.Ops[2].Params[plan.MigrateParamSteps])
	}
}

func TestPlanMigrateForceIncludesExpectedOperations(t *testing.T) {
	t.Parallel()

	planned, err := PlanMigrateForce(context.Background(), []string{"20260420034348"}, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err != nil {
		t.Fatalf("plan migrate force: %v", err)
	}

	if planned.CommandID != "db:migrate:force" {
		t.Fatalf("expected command id db:migrate:force, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[2].Type != plan.OpMigrateForce {
		t.Fatalf("expected third op migrate_force, got %q", planned.Ops[2].Type)
	}

	if planned.Ops[2].Params[plan.MigrateParamVersion] != "20260420034348" {
		t.Fatalf("unexpected force version: %q", planned.Ops[2].Params[plan.MigrateParamVersion])
	}
}

func TestValidateRollbackRejectsNonPositiveSteps(t *testing.T) {
	t.Parallel()

	err := ValidateRollback([]string{"0"}, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err == nil {
		t.Fatal("expected validation error for non-positive steps")
	}
}

func TestValidateMigrateForceRejectsMissingVersion(t *testing.T) {
	t.Parallel()

	err := ValidateMigrateForce(nil, testParams{values: map[string]string{
		"dsn": "postgres://localhost:5432/demo_app?sslmode=disable",
	}})
	if err == nil {
		t.Fatal("expected validation error when version is missing")
	}
}

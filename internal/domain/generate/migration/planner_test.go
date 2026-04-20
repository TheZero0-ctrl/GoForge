package migration

import (
	"context"
	"path/filepath"
	"regexp"
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

func TestValidateRequiresName(t *testing.T) {
	t.Parallel()

	err := Validate(nil, testParams{})
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestValidateRequiresSnakeCaseName(t *testing.T) {
	t.Parallel()

	err := Validate([]string{"CreateUsers"}, testParams{})
	if err == nil {
		t.Fatal("expected snake_case validation error")
	}
}

func TestValidateRejectsInvalidFieldForMatchedPattern(t *testing.T) {
	t.Parallel()

	err := Validate([]string{"add_email_to_users", "email:unknown"}, testParams{})
	if err == nil {
		t.Fatal("expected unsupported field type error")
	}
}

func TestPlanCreatesMigrationPairWithSharedTimestamp(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"create_users"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if planned.CommandID != "generate:migration" {
		t.Fatalf("expected command id generate:migration, got %q", planned.CommandID)
	}

	if len(planned.Ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpMkdir || planned.Ops[0].Path != "migrations" {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	up := planned.Ops[1]
	down := planned.Ops[2]

	if up.Type != plan.OpWriteFile {
		t.Fatalf("expected second op write_file, got %q", up.Type)
	}

	if down.Type != plan.OpWriteFile {
		t.Fatalf("expected third op write_file, got %q", down.Type)
	}

	if filepath.Ext(strings.TrimSuffix(up.Path, ".sql")) != ".up" {
		t.Fatalf("expected .up.sql file, got %q", up.Path)
	}

	if filepath.Ext(strings.TrimSuffix(down.Path, ".sql")) != ".down" {
		t.Fatalf("expected .down.sql file, got %q", down.Path)
	}

	if !strings.Contains(string(up.Data), "CREATE TABLE IF NOT EXISTS \"users\"") {
		t.Fatalf("expected create table scaffold in up migration, got %q", string(up.Data))
	}

	if !strings.Contains(string(down.Data), "DROP TABLE IF EXISTS \"users\"") {
		t.Fatalf("expected drop table scaffold in down migration, got %q", string(down.Data))
	}

	upBase := strings.TrimSuffix(filepath.Base(up.Path), ".up.sql")
	downBase := strings.TrimSuffix(filepath.Base(down.Path), ".down.sql")
	if upBase != downBase {
		t.Fatalf("expected shared timestamp/name base, got up=%q down=%q", upBase, downBase)
	}

	match := regexp.MustCompile(`^[0-9]{14}_create_users$`).MatchString
	if !match(upBase) {
		t.Fatalf("expected base to match timestamped migration pattern, got %q", upBase)
	}
}

func TestPlanFallsBackToEmptyTemplatesForCustomNames(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"backfill_widgets"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	up := planned.Ops[1]
	down := planned.Ops[2]

	if len(up.Data) != 0 || len(down.Data) != 0 {
		t.Fatalf("expected empty file contents for custom migration, got up=%q down=%q", string(up.Data), string(down.Data))
	}
}

func TestPlanRemovePatternWithoutTypeKeepsDownEmpty(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"remove_email_from_users"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	up := planned.Ops[1]
	down := planned.Ops[2]

	if !strings.Contains(string(up.Data), "ALTER TABLE \"users\" DROP COLUMN IF EXISTS \"email\";") {
		t.Fatalf("expected drop column scaffold in up migration, got %q", string(up.Data))
	}

	if len(down.Data) != 0 {
		t.Fatalf("expected empty down migration for remove pattern without type, got %q", string(down.Data))
	}
}

func TestPlanAddPatternQuotesReservedColumnName(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"add_desc_to_products", "desc:string"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	up := planned.Ops[1]
	if !strings.Contains(string(up.Data), "ALTER TABLE \"products\" ADD COLUMN IF NOT EXISTS \"desc\" text;") {
		t.Fatalf("expected quoted add column scaffold, got %q", string(up.Data))
	}
}

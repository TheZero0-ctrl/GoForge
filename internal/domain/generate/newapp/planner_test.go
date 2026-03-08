package newapp

import (
	"context"
	"testing"

	"goforge/internal/domain/plan"
)

func TestValidateMatchesConfigValidation(t *testing.T) {
	t.Parallel()

	err := Validate([]string{"\n\t"}, testParams{})
	if err == nil {
		t.Fatal("expected validation error for empty app name")
	}
}

func TestPlanIncludesRequiredBaseOperations(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"demo-api"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if planned.CommandID != "new" {
		t.Fatalf("expected command id new, got %q", planned.CommandID)
	}

	if len(planned.Ops) < 3 {
		t.Fatalf("expected at least 3 ops, got %d", len(planned.Ops))
	}

	if planned.Ops[0].Type != plan.OpEnsureEmptyDir || planned.Ops[0].Path != "demo-api" {
		t.Fatalf("unexpected first op: %+v", planned.Ops[0])
	}

	if planned.Ops[1].Type != plan.OpMkdir || planned.Ops[1].Path != "demo-api" {
		t.Fatalf("unexpected second op: %+v", planned.Ops[1])
	}

	if planned.Ops[2].Type != plan.OpRun {
		t.Fatalf("expected third op to be run, got %q", planned.Ops[2].Type)
	}

	if planned.Ops[2].Path != "demo-api" {
		t.Fatalf("expected go mod init cwd to be app dir, got %q", planned.Ops[2].Path)
	}

	if len(planned.Ops[2].Cmd) != 4 || planned.Ops[2].Cmd[0] != "go" || planned.Ops[2].Cmd[1] != "mod" || planned.Ops[2].Cmd[2] != "init" || planned.Ops[2].Cmd[3] != "demo-api" {
		t.Fatalf("unexpected go mod init command: %v", planned.Ops[2].Cmd)
	}
}

func TestPlanIncludesManifestWriteOperations(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"demo-api"}, testParams{})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	writeCount := 0
	hasReadme := false
	hasMain := false

	for _, op := range planned.Ops {
		if op.Type != plan.OpWriteFile {
			continue
		}

		writeCount++
		if op.Path == "demo-api/README.md" {
			hasReadme = true
		}
		if op.Path == "demo-api/cmd/api/main.go" {
			hasMain = true
		}
	}

	if writeCount == 0 {
		t.Fatal("expected at least one write op")
	}

	if !hasReadme {
		t.Fatal("expected README.md write op")
	}

	if !hasMain {
		t.Fatal("expected cmd/api/main.go write op")
	}
}

func TestPlanSkipsPostActionsWhenRequested(t *testing.T) {
	t.Parallel()

	planned, err := Plan(context.Background(), []string{"demo-api"}, testParams{values: map[string]string{
		"skip-git":  "true",
		"skip-tidy": "true",
	}})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	for _, op := range planned.Ops {
		if op.Type != plan.OpRun || len(op.Cmd) < 3 {
			continue
		}

		if op.Cmd[0] == "go" && op.Cmd[1] == "mod" && op.Cmd[2] == "tidy" {
			t.Fatal("expected go mod tidy to be omitted")
		}

		if op.Cmd[0] == "git" && op.Cmd[1] == "init" {
			t.Fatal("expected git init to be omitted")
		}
	}
}

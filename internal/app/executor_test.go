package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"goforge/internal/domain/command"
	"goforge/internal/domain/plan"
	"goforge/internal/infra/dbmigrate"
	infrafs "goforge/internal/infra/fs"
	"goforge/internal/infra/proc"
)

type fakeMigrateRunner struct {
	err error
}

func (r fakeMigrateRunner) Up(sourceURL, databaseURL string) error {
	return r.err
}

func (r fakeMigrateRunner) DownSteps(sourceURL, databaseURL string, steps int) error {
	return r.err
}

func (r fakeMigrateRunner) Force(sourceURL, databaseURL string, version int) error {
	return r.err
}

func staticMigrateCommand() command.Command {
	spec := command.Spec{ID: "db:migrate", Use: "db:migrate", Short: "migrate"}
	planner := func(context.Context, command.Input) (plan.Plan, error) {
		return plan.Plan{
			CommandID:   "db:migrate",
			Description: "Apply database migrations",
			Ops: []plan.Operation{{
				Type: plan.OpMigrateUp,
				Params: map[string]string{
					plan.MigrateParamSourceURL:   "file:///tmp/migrations",
					plan.MigrateParamDatabaseURL: "postgres://localhost:5432/demo?sslmode=disable",
				},
			}},
		}, nil
	}

	return command.NewStatic(spec, nil, planner)
}

func TestExecutorDryRun(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner(), fakeMigrateRunner{})
	result := exec.Execute(context.Background(), command.Input{
		CommandID: "new",
		Args:      []string{"demo-api"},
		Flags:     command.Flags{DryRun: true},
	})

	if result.Code != ExitOK {
		t.Fatalf("expected exit ok, got %d", result.Code)
	}

	if len(result.Entries) == 0 {
		t.Fatal("expected dry-run entries")
	}

	if got := result.Entries[0].Status; got != "INFO" {
		t.Fatalf("expected first entry INFO, got %q", got)
	}
}

func TestExecutorValidationError(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner(), fakeMigrateRunner{})
	result := exec.Execute(context.Background(), command.Input{CommandID: "new"})

	if result.Code != ExitValidation {
		t.Fatalf("expected validation exit code, got %d", result.Code)
	}
}

func TestExecutorMigrateNoChangeReturnsSkipEntry(t *testing.T) {
	reg := command.NewRegistry()
	if err := reg.Register(staticMigrateCommand()); err != nil {
		t.Fatalf("register migrate command: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner(), fakeMigrateRunner{err: dbmigrate.ErrNoChange})
	result := exec.Execute(context.Background(), command.Input{CommandID: "db:migrate"})

	if result.Code != ExitOK {
		t.Fatalf("expected exit ok, got %d", result.Code)
	}

	foundSkip := false
	for _, entry := range result.Entries {
		if entry.Status == "SKIP" {
			foundSkip = true
			break
		}
	}

	if !foundSkip {
		t.Fatalf("expected skip entry, got %+v", result.Entries)
	}
}

func TestExecutorMigrateFailureReturnsExecutionError(t *testing.T) {
	reg := command.NewRegistry()
	if err := reg.Register(staticMigrateCommand()); err != nil {
		t.Fatalf("register migrate command: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner(), fakeMigrateRunner{err: errors.New("boom")})
	result := exec.Execute(context.Background(), command.Input{CommandID: "db:migrate"})

	if result.Code != ExitExecution {
		t.Fatalf("expected execution exit code, got %d", result.Code)
	}
}

func TestExecutorMigrateDirtyErrorIncludesRecoveryHint(t *testing.T) {
	reg := command.NewRegistry()
	if err := reg.Register(staticMigrateCommand()); err != nil {
		t.Fatalf("register migrate command: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner(), fakeMigrateRunner{err: dbmigrate.ErrDirty{Version: 20260420034348}})
	result := exec.Execute(context.Background(), command.Input{CommandID: "db:migrate"})

	if result.Code != ExitConflict {
		t.Fatalf("expected conflict exit code, got %d", result.Code)
	}

	foundHint := false
	for _, entry := range result.Entries {
		if entry.Status == "ERROR" && strings.Contains(entry.Message, "db:migrate:force 20260420034348") {
			foundHint = true
			break
		}
	}

	if !foundHint {
		t.Fatalf("expected dirty recovery hint, got %+v", result.Entries)
	}
}

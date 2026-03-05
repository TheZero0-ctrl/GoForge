package app

import (
	"context"
	"testing"

	"goforge/internal/domain/command"
	infrafs "goforge/internal/infra/fs"
	"goforge/internal/infra/proc"
)

func TestExecutorDryRun(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner())
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

	exec := NewExecutor(reg, infrafs.NewOSFS(), proc.NewOSRunner())
	result := exec.Execute(context.Background(), command.Input{CommandID: "new"})

	if result.Code != ExitValidation {
		t.Fatalf("expected validation exit code, got %d", result.Code)
	}
}

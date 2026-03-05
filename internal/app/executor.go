package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"goforge/internal/domain/command"
	"goforge/internal/domain/plan"
	"goforge/internal/infra/fs"
	"goforge/internal/infra/proc"
)

type ExitCode int

const (
	ExitOK         ExitCode = 0
	ExitValidation ExitCode = 2
	ExitConflict   ExitCode = 3
	ExitExecution  ExitCode = 4

	defaultDirPerm  os.FileMode = 0o755
	defaultFilePerm os.FileMode = 0o644
)

type Entry struct {
	Status  string
	Message string
}

type Result struct {
	Code    ExitCode
	Entries []Entry
}

type Executor struct {
	registry *command.Registry
	fs       fs.FS
	runner   proc.Runner
}

func NewExecutor(registry *command.Registry, fileSystem fs.FS, runner proc.Runner) *Executor {
	return &Executor{registry: registry, fs: fileSystem, runner: runner}
}

func (e *Executor) Execute(ctx context.Context, input command.Input) Result {
	// Keep one execution pipeline for every command so flags like --dry-run,
	// --force, and --skip behave consistently across the CLI.
	cmd, ok := e.registry.Resolve(input.CommandID)
	if !ok {
		return Result{
			Code: ExitValidation,
			Entries: []Entry{
				{Status: "ERROR", Message: fmt.Sprintf("unknown command %q", input.CommandID)},
			},
		}
	}

	input.CommandID = cmd.Spec().ID

	if err := cmd.Validate(input); err != nil {
		return Result{
			Code: ExitValidation,
			Entries: []Entry{
				{Status: "ERROR", Message: err.Error()},
			},
		}
	}

	planned, err := cmd.Plan(ctx, input)
	if err != nil {
		return Result{Code: ExitExecution, Entries: []Entry{{Status: "ERROR", Message: err.Error()}}}
	}

	entries := make([]Entry, 0, len(planned.Ops)+1)

	if input.Flags.DryRun {
		entries = append(entries, Entry{Status: "INFO", Message: fmt.Sprintf("DRY-RUN %s", planned.Description)})
		for _, op := range planned.Ops {
			entries = append(entries, Entry{Status: "PLAN", Message: describeOp(op)})
		}
		return Result{Code: ExitOK, Entries: entries}
	}

	for _, op := range planned.Ops {
		entry, opErr := e.executeOp(ctx, op, input.Flags)
		entries = append(entries, entry)
		if opErr != nil {
			var conflict conflictError
			if errors.As(opErr, &conflict) {
				return Result{Code: ExitConflict, Entries: entries}
			}
			entries = append(entries, Entry{Status: "ERROR", Message: opErr.Error()})
			return Result{Code: ExitExecution, Entries: entries}
		}
	}

	entries = append(entries, Entry{Status: "DONE", Message: fmt.Sprintf("%s finished", planned.CommandID)})
	return Result{Code: ExitOK, Entries: entries}
}

type conflictError struct {
	message string
}

func (e conflictError) Error() string { return e.message }

func (e *Executor) executeOp(ctx context.Context, op plan.Operation, flags command.Flags) (Entry, error) {
	switch op.Type {
	case plan.OpNote:
		return Entry{Status: "INFO", Message: op.Message}, nil
	case plan.OpMkdir:
		perm := op.Perm
		if perm == 0 {
			perm = defaultDirPerm
		}
		if err := e.fs.MkdirAll(op.Path, perm); err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("MKDIR %s", op.Path)}, err
		}
		return Entry{Status: "CREATE", Message: op.Path}, nil
	case plan.OpWriteFile:
		exists, err := e.fs.Exists(op.Path)
		if err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("CHECK %s", op.Path)}, err
		}

		if exists {
			if flags.Skip {
				return Entry{Status: "SKIP", Message: fmt.Sprintf("%s (exists)", op.Path)}, nil
			}
			if !flags.Force {
				return Entry{Status: "ERROR", Message: fmt.Sprintf("%s already exists", op.Path)}, conflictError{message: fmt.Sprintf("conflict: %s already exists", op.Path)}
			}
		}

		if err := e.fs.MkdirAll(filepath.Dir(op.Path), defaultDirPerm); err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("MKDIR %s", filepath.Dir(op.Path))}, err
		}

		perm := op.Perm
		if perm == 0 {
			perm = defaultFilePerm
		}

		if err := e.fs.WriteFile(op.Path, op.Data, perm); err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("WRITE %s", op.Path)}, err
		}

		if exists && flags.Force {
			return Entry{Status: "UPDATE", Message: op.Path}, nil
		}
		return Entry{Status: "CREATE", Message: op.Path}, nil
	case plan.OpRun:
		if len(op.Cmd) == 0 {
			return Entry{Status: "ERROR", Message: "RUN <empty>"}, errors.New("run operation has empty command")
		}
		if err := e.runner.Run(ctx, op.Cmd[0], op.Cmd[1:]...); err != nil {
			return Entry{Status: "ERROR", Message: "RUN failed"}, err
		}
		return Entry{Status: "RUN", Message: fmt.Sprintf("%s", op.Cmd)}, nil
	default:
		return Entry{Status: "ERROR", Message: fmt.Sprintf("unknown op %q", op.Type)}, fmt.Errorf("unknown operation type %q", op.Type)
	}
}

func describeOp(op plan.Operation) string {
	switch op.Type {
	case plan.OpNote:
		return op.Message
	case plan.OpMkdir:
		return fmt.Sprintf("MKDIR %s", op.Path)
	case plan.OpWriteFile:
		return fmt.Sprintf("WRITE %s", op.Path)
	case plan.OpRun:
		return fmt.Sprintf("RUN %v", op.Cmd)
	default:
		return fmt.Sprintf("UNKNOWN %q", op.Type)
	}
}

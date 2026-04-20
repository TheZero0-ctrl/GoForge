package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"goforge/internal/domain/command"
	"goforge/internal/domain/plan"
	"goforge/internal/infra/dbmigrate"
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
	migrator dbmigrate.Runner
}

func NewExecutor(registry *command.Registry, fileSystem fs.FS, runner proc.Runner, migrator dbmigrate.Runner) *Executor {
	return &Executor{registry: registry, fs: fileSystem, runner: runner, migrator: migrator}
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
	case plan.OpEnsureEmptyDir:
		exists, err := e.fs.Exists(op.Path)

		if err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("CHECK %s", op.Path)}, err
		}

		if !exists {
			return Entry{Status: "INFO", Message: fmt.Sprintf("target %s does not exist (ok)", op.Path)}, nil
		}

		empty, err := e.fs.IsDirEmpty(op.Path)
		if err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("CHECK %s", op.Path)}, err
		}

		if empty {
			return Entry{Status: "INFO", Message: fmt.Sprintf("target %s is empty (ok)", op.Path)}, nil
		}

		return Entry{Status: "ERROR", Message: fmt.Sprintf("%s is non-empty", op.Path)}, conflictError{message: fmt.Sprintf("conflict: target directory %s is non-empty (use --force)", op.Path)}
	case plan.OpEnsureExists:
		exists, err := e.fs.Exists(op.Path)

		if err != nil {
			return Entry{Status: "ERROR", Message: fmt.Sprintf("CHECK %s", op.Path)}, err
		}

		if !exists {
			msg := op.Message
			if msg == "" {
				msg = fmt.Sprintf("required path %s not found", op.Path)
			}
			return Entry{Status: "ERROR", Message: msg}, conflictError{message: msg}
		}

		return Entry{Status: "INFO", Message: fmt.Sprintf("found %s", op.Path)}, nil
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
		if err := e.runner.Run(ctx, op.Path, op.Cmd[0], op.Cmd[1:]...); err != nil {
			return Entry{Status: "ERROR", Message: "RUN failed"}, err
		}
		return Entry{Status: "RUN", Message: fmt.Sprintf("%s", op.Cmd)}, nil
	case plan.OpMigrateUp:
		if e.migrator == nil {
			return Entry{Status: "ERROR", Message: "MIGRATE UP <unavailable>"}, errors.New("migrate runner is not configured")
		}

		sourceURL, databaseURL, opErr := migrateURLs(op)
		if opErr != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE UP <invalid>"}, opErr
		}

		err := e.migrator.Up(sourceURL, databaseURL)
		if err != nil {
			if errors.Is(err, dbmigrate.ErrNoChange) {
				return Entry{Status: "SKIP", Message: "database already up to date"}, nil
			}
			var dirty dbmigrate.ErrDirty
			if errors.As(err, &dirty) {
				hint := fmt.Sprintf("database is dirty at version %d; run `goforge db:migrate:force %d` then rerun `goforge db:migrate`", dirty.Version, dirty.Version)
				return Entry{Status: "ERROR", Message: hint}, conflictError{message: hint}
			}
			return Entry{Status: "ERROR", Message: "MIGRATE UP failed"}, err
		}

		return Entry{Status: "RUN", Message: fmt.Sprintf("MIGRATE UP %s", sourceURL)}, nil
	case plan.OpMigrateDown:
		if e.migrator == nil {
			return Entry{Status: "ERROR", Message: "MIGRATE DOWN <unavailable>"}, errors.New("migrate runner is not configured")
		}

		sourceURL, databaseURL, opErr := migrateURLs(op)
		if opErr != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE DOWN <invalid>"}, opErr
		}

		steps, opErr := migratePositiveInt(op, plan.MigrateParamSteps)
		if opErr != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE DOWN <invalid>"}, opErr
		}

		err := e.migrator.DownSteps(sourceURL, databaseURL, steps)
		if err != nil {
			if errors.Is(err, dbmigrate.ErrNoChange) {
				return Entry{Status: "SKIP", Message: "database already at base migration"}, nil
			}
			var dirty dbmigrate.ErrDirty
			if errors.As(err, &dirty) {
				hint := fmt.Sprintf("database is dirty at version %d; run `goforge db:migrate:force %d` then rerun `goforge db:rollback`", dirty.Version, dirty.Version)
				return Entry{Status: "ERROR", Message: hint}, conflictError{message: hint}
			}
			return Entry{Status: "ERROR", Message: "MIGRATE DOWN failed"}, err
		}

		return Entry{Status: "RUN", Message: fmt.Sprintf("MIGRATE DOWN %d %s", steps, sourceURL)}, nil
	case plan.OpMigrateForce:
		if e.migrator == nil {
			return Entry{Status: "ERROR", Message: "MIGRATE FORCE <unavailable>"}, errors.New("migrate runner is not configured")
		}

		sourceURL, databaseURL, opErr := migrateURLs(op)
		if opErr != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE FORCE <invalid>"}, opErr
		}

		version, opErr := migrateInt(op, plan.MigrateParamVersion)
		if opErr != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE FORCE <invalid>"}, opErr
		}

		err := e.migrator.Force(sourceURL, databaseURL, version)
		if err != nil {
			return Entry{Status: "ERROR", Message: "MIGRATE FORCE failed"}, err
		}

		return Entry{Status: "RUN", Message: fmt.Sprintf("MIGRATE FORCE %d %s", version, sourceURL)}, nil
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
	case plan.OpMigrateUp:
		sourceURL := strings.TrimSpace(op.Params[plan.MigrateParamSourceURL])
		if sourceURL == "" {
			return "MIGRATE UP <missing-source>"
		}
		return fmt.Sprintf("MIGRATE UP %s", sourceURL)
	case plan.OpMigrateDown:
		sourceURL := strings.TrimSpace(op.Params[plan.MigrateParamSourceURL])
		steps := strings.TrimSpace(op.Params[plan.MigrateParamSteps])
		if sourceURL == "" || steps == "" {
			return "MIGRATE DOWN <missing-params>"
		}
		return fmt.Sprintf("MIGRATE DOWN %s %s", steps, sourceURL)
	case plan.OpMigrateForce:
		sourceURL := strings.TrimSpace(op.Params[plan.MigrateParamSourceURL])
		version := strings.TrimSpace(op.Params[plan.MigrateParamVersion])
		if sourceURL == "" || version == "" {
			return "MIGRATE FORCE <missing-params>"
		}
		return fmt.Sprintf("MIGRATE FORCE %s %s", version, sourceURL)
	case plan.OpEnsureEmptyDir:
		return fmt.Sprintf("CHECK EMPTY %s", op.Path)
	case plan.OpEnsureExists:
		return fmt.Sprintf("CHECK EXISTS %s", op.Path)
	default:
		return fmt.Sprintf("UNKNOWN %q", op.Type)
	}
}

func migrateURLs(op plan.Operation) (string, string, error) {
	sourceURL := strings.TrimSpace(op.Params[plan.MigrateParamSourceURL])
	databaseURL := strings.TrimSpace(op.Params[plan.MigrateParamDatabaseURL])
	if sourceURL == "" || databaseURL == "" {
		return "", "", errors.New("migrate operation requires source and database URLs")
	}
	return sourceURL, databaseURL, nil
}

func migratePositiveInt(op plan.Operation, key string) (int, error) {
	value, err := migrateInt(op, key)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("migrate operation requires %s to be positive", key)
	}
	return value, nil
}

func migrateInt(op plan.Operation, key string) (int, error) {
	raw := strings.TrimSpace(op.Params[key])
	if raw == "" {
		return 0, fmt.Errorf("migrate operation requires param %s", key)
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("migrate operation param %s must be an integer", key)
	}
	return value, nil
}

package db

import (
	"context"
	"fmt"
	"goforge/internal/domain/params"
	"goforge/internal/domain/plan"
	"net/url"
	"path/filepath"
	"strconv"
)

func ValidateCreate(args []string, p params.Params) error {
	if len(args) != 0 {
		return fmt.Errorf("db:create does not accept positional arguments")
	}

	_, err := ParseConfig(p)

	return err
}

func ValidateDrop(args []string, p params.Params) error {
	if len(args) != 0 {
		return fmt.Errorf("db:drop does not accept positional arguments")
	}

	_, err := ParseConfig(p)

	return err
}

func ValidateMigrate(args []string, p params.Params) error {
	if len(args) != 0 {
		return fmt.Errorf("db:migrate does not accept positional arguments")
	}

	_, err := ParseConfig(p)

	return err
}

func ValidateRollback(args []string, p params.Params) error {
	if len(args) > 1 {
		return fmt.Errorf("db:rollback accepts at most one positional argument: [steps]")
	}

	if len(args) == 1 {
		steps, err := strconv.Atoi(args[0])
		if err != nil || steps <= 0 {
			return fmt.Errorf("rollback steps must be a positive integer")
		}
	}

	_, err := ParseConfig(p)

	return err
}

func ValidateMigrateForce(args []string, p params.Params) error {
	if len(args) != 1 {
		return fmt.Errorf("db:migrate:force requires one positional argument: <version>")
	}

	if _, err := strconv.Atoi(args[0]); err != nil {
		return fmt.Errorf("force version must be an integer")
	}

	_, err := ParseConfig(p)

	return err
}

func PlanCreate(_ context.Context, _ []string, p params.Params) (plan.Plan, error) {
	cfg, err := ParseConfig(p)

	if err != nil {
		return plan.Plan{}, err
	}

	ops := make([]plan.Operation, 0, 1)

	create := fmt.Sprintf("CREATE DATABASE \"%s\";", cfg.DatabaseName)
	ops = append(ops, plan.Operation{Type: plan.OpRun, Cmd: []string{"psql", cfg.AdminDSN, "-v", "ON_ERROR_STOP=1", "-c", create}})

	return plan.Plan{
		CommandID:   "db:create",
		Description: "Create database",
		Ops:         ops,
	}, nil
}

func PlanDrop(_ context.Context, _ []string, p params.Params) (plan.Plan, error) {
	cfg, err := ParseConfig(p)

	if err != nil {
		return plan.Plan{}, err
	}

	ops := make([]plan.Operation, 0, 1)

	drop := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\";", cfg.DatabaseName)
	ops = append(ops, plan.Operation{Type: plan.OpRun, Cmd: []string{"psql", cfg.AdminDSN, "-v", "ON_ERROR_STOP=1", "-c", drop}})

	return plan.Plan{
		CommandID:   "db:drop",
		Description: "Drop database",
		Ops:         ops,
	}, nil
}

func PlanMigrate(_ context.Context, _ []string, p params.Params) (plan.Plan, error) {
	migrationSource, cfg, err := resolveMigrateContext(p)
	if err != nil {
		return plan.Plan{}, err
	}

	ops := migratePrechecks()

	ops = append(ops, plan.Operation{Type: plan.OpMigrateUp, Params: map[string]string{
		plan.MigrateParamSourceURL:   migrationSource,
		plan.MigrateParamDatabaseURL: cfg.DSN,
	}})

	return plan.Plan{
		CommandID:   "db:migrate",
		Description: "Apply database migrations",
		Ops:         ops,
	}, nil
}

func PlanRollback(_ context.Context, args []string, p params.Params) (plan.Plan, error) {
	migrationSource, cfg, err := resolveMigrateContext(p)
	if err != nil {
		return plan.Plan{}, err
	}

	steps := 1
	if len(args) == 1 {
		steps, err = strconv.Atoi(args[0])
		if err != nil || steps <= 0 {
			return plan.Plan{}, fmt.Errorf("rollback steps must be a positive integer")
		}
	}

	ops := migratePrechecks()
	ops = append(ops, plan.Operation{Type: plan.OpMigrateDown, Params: map[string]string{
		plan.MigrateParamSourceURL:   migrationSource,
		plan.MigrateParamDatabaseURL: cfg.DSN,
		plan.MigrateParamSteps:       strconv.Itoa(steps),
	}})

	return plan.Plan{
		CommandID:   "db:rollback",
		Description: "Rollback database migrations",
		Ops:         ops,
	}, nil
}

func PlanMigrateForce(_ context.Context, args []string, p params.Params) (plan.Plan, error) {
	migrationSource, cfg, err := resolveMigrateContext(p)
	if err != nil {
		return plan.Plan{}, err
	}

	version, err := strconv.Atoi(args[0])
	if err != nil {
		return plan.Plan{}, fmt.Errorf("force version must be an integer")
	}

	ops := migratePrechecks()
	ops = append(ops, plan.Operation{Type: plan.OpMigrateForce, Params: map[string]string{
		plan.MigrateParamSourceURL:   migrationSource,
		plan.MigrateParamDatabaseURL: cfg.DSN,
		plan.MigrateParamVersion:     strconv.Itoa(version),
	}})

	return plan.Plan{
		CommandID:   "db:migrate:force",
		Description: "Force migration version",
		Ops:         ops,
	}, nil
}

func resolveMigrateContext(p params.Params) (string, Config, error) {
	cfg, err := ParseConfig(p)

	if err != nil {
		return "", Config{}, err
	}

	absMigrations, err := filepath.Abs("migrations")
	if err != nil {
		return "", Config{}, fmt.Errorf("resolve migrations path: %w", err)
	}

	migrationSource := (&url.URL{Scheme: "file", Path: filepath.ToSlash(absMigrations)}).String()
	return migrationSource, cfg, nil
}

func migratePrechecks() []plan.Operation {
	return []plan.Operation{
		{Type: plan.OpEnsureExists, Path: "cmd/api/main.go", Message: "missing cmd/api/main.go; run from GoForge app root"},
		{Type: plan.OpEnsureExists, Path: "migrations", Message: "missing migrations directory"},
	}
}

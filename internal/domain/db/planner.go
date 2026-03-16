package db

import (
	"context"
	"fmt"
	"goforge/internal/domain/params"
	"goforge/internal/domain/plan"
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
		return fmt.Errorf("db:create does not accept positional arguments")
	}

	_, err := ParseConfig(p)

	return err
}

func PlanCreate(_ context.Context, _ []string, p params.Params) (plan.Plan, error) {
	cfg, err := ParseConfig(p)

	if err != nil {
		return plan.Plan{}, err
	}

	ops := []plan.Operation{
		{Type: plan.OpEnsureExists, Path: "config/database.toml", Message: "missing config/database.toml"},
	}

	create := fmt.Sprintf("CREATE DATABASE \"%s\";", cfg.DatabaseName)
	ops = append(ops, plan.Operation{Type: plan.OpRun, Cmd: []string{"psql", cfg.AdminDSN, "-v", "ON_ERROR_STOP=1", "-c", create}})

	return plan.Plan{
		CommandID:   "db:create",
		Description: "Create database",
		Ops:         ops,
	}, nil
}

func PlanDrop(ctx context.Context, args []string, p params.Params) (plan.Plan, error) {
	cfg, err := ParseConfig(p)

	if err != nil {
		return plan.Plan{}, err
	}

	ops := []plan.Operation{
		{Type: plan.OpEnsureExists, Path: "config/database.toml", Message: "missing config/database.toml"},
	}

	drop := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\";", cfg.DatabaseName)
	ops = append(ops, plan.Operation{Type: plan.OpRun, Cmd: []string{"psql", cfg.AdminDSN, "-v", "ON_ERROR_STOP=1", "-c", drop}})

	return plan.Plan{
		CommandID:   "db:drop",
		Description: "Drop database",
		Ops:         ops,
	}, nil
}

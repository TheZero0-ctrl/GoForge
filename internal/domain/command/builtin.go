package command

import (
	"context"

	"goforge/internal/domain/db"
	"goforge/internal/domain/generate/migration"
	"goforge/internal/domain/generate/newapp"
	"goforge/internal/domain/generate/resource"
	"goforge/internal/domain/plan"
)

func NewNewCommand() Command {
	spec := Spec{
		ID:    "new",
		Use:   "new <app-name>",
		Short: "Create a new Go API app",
	}

	validate := func(input Input) error {
		return newapp.Validate(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return newapp.Plan(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewGenerateCommand() Command {
	spec := Spec{
		ID:      "generate",
		Use:     "generate",
		Short:   "Generate code using a named generator",
		Aliases: []string{"g"},
	}

	planner := func(_ context.Context, _ Input) (plan.Plan, error) {
		return plan.Plan{
			CommandID:   spec.ID,
			Description: "Generate command namespace",
			Ops: []plan.Operation{
				{Type: plan.OpNote, Message: "phase 0: generate namespace is wired; generators arrive in later phases"},
			},
		}, nil
	}

	return NewStatic(spec, nil, planner)
}

func NewDestroyCommand() Command {
	spec := Spec{
		ID:      "destroy",
		Use:     "destroy",
		Short:   "Reverse artifacts produced by generators",
		Aliases: []string{"d"},
	}

	planner := func(_ context.Context, _ Input) (plan.Plan, error) {
		return plan.Plan{
			CommandID:   spec.ID,
			Description: "Destroy command namespace",
			Ops: []plan.Operation{
				{Type: plan.OpNote, Message: "phase 0: destroy namespace is wired; destroy actions arrive later"},
			},
		}, nil
	}

	return NewStatic(spec, nil, planner)
}

func NewGenerateMigrationCommand() Command {
	spec := Spec{
		ID:    "generate:migration",
		Use:   "migration <name> [field:type...]",
		Short: "Generate migration files",
	}

	validate := func(input Input) error {
		return migration.Validate(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return migration.Plan(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewGenerateResourceCommand() Command {
	spec := Spec{
		ID:    "generate:resource",
		Use:   "resource <name> <field:type>...",
		Short: "Generate resource files, wiring, and migration",
	}

	validate := func(input Input) error {
		return resource.Validate(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return resource.Plan(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewDBCreateCommand() Command {
	spec := Spec{
		ID:    "db:create",
		Use:   "db:create",
		Short: "Create database",
	}

	validate := func(input Input) error {
		return db.ValidateCreate(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return db.PlanCreate(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewDBDropCommand() Command {
	spec := Spec{
		ID:    "db:drop",
		Use:   "db:drop",
		Short: "Drop database",
	}

	validate := func(input Input) error {
		return db.ValidateDrop(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return db.PlanDrop(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewDBMigrateCommand() Command {
	spec := Spec{
		ID:    "db:migrate",
		Use:   "db:migrate",
		Short: "Apply database migrations",
	}

	validate := func(input Input) error {
		return db.ValidateMigrate(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return db.PlanMigrate(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewDBRollbackCommand() Command {
	spec := Spec{
		ID:    "db:rollback",
		Use:   "db:rollback [steps]",
		Short: "Rollback database migrations",
	}

	validate := func(input Input) error {
		return db.ValidateRollback(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return db.PlanRollback(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

func NewDBMigrateForceCommand() Command {
	spec := Spec{
		ID:    "db:migrate:force",
		Use:   "db:migrate:force <version>",
		Short: "Force database migration version",
	}

	validate := func(input Input) error {
		return db.ValidateMigrateForce(input.Args, input)
	}

	planner := func(ctx context.Context, input Input) (plan.Plan, error) {
		return db.PlanMigrateForce(ctx, input.Args, input)
	}

	return NewStatic(spec, validate, planner)
}

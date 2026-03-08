package command

import (
	"context"

	"goforge/internal/domain/generate/newapp"
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

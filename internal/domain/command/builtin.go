package command

import (
	"context"
	"errors"

	"goforge/internal/domain/plan"
)

func NewNewCommand() Command {
	spec := Spec{
		ID:    "new",
		Use:   "new <app-name>",
		Short: "Create a new GoForge API app",
	}

	validate := func(input Input) error {
		if len(input.Args) != 1 {
			return errors.New("new requires exactly one argument: <app-name>")
		}
		return nil
	}

	planner := func(_ context.Context, _ Input) (plan.Plan, error) {
		return plan.Plan{
			CommandID:   spec.ID,
			Description: "Create new app scaffold",
			Ops: []plan.Operation{
				{Type: plan.OpNote, Message: "phase 0: new command planning is wired; implementation comes in phase 1"},
			},
		}, nil
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

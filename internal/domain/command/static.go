package command

import (
	"context"

	"goforge/internal/domain/plan"
)

type ValidateFunc func(input Input) error
type PlanFunc func(ctx context.Context, input Input) (plan.Plan, error)

type StaticCommand struct {
	spec      Spec
	validate  ValidateFunc
	plannerFn PlanFunc
}

func NewStatic(spec Spec, validate ValidateFunc, planner PlanFunc) *StaticCommand {
	if validate == nil {
		validate = func(Input) error { return nil }
	}

	if planner == nil {
		planner = func(context.Context, Input) (plan.Plan, error) {
			return plan.Plan{CommandID: spec.ID}, nil
		}
	}

	return &StaticCommand{spec: spec, validate: validate, plannerFn: planner}
}

func (c *StaticCommand) Spec() Spec {
	return c.spec
}

func (c *StaticCommand) Validate(input Input) error {
	return c.validate(input)
}

func (c *StaticCommand) Plan(ctx context.Context, input Input) (plan.Plan, error) {
	return c.plannerFn(ctx, input)
}

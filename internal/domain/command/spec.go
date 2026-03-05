package command

import (
	"context"

	"goforge/internal/domain/plan"
)

type Flags struct {
	DryRun bool
	Force  bool
	Skip   bool
}

type Input struct {
	CommandID string
	Args      []string
	Flags     Flags
}

type Spec struct {
	ID      string
	Use     string
	Short   string
	Aliases []string
}

type Command interface {
	Spec() Spec
	Validate(input Input) error
	Plan(ctx context.Context, input Input) (plan.Plan, error)
}

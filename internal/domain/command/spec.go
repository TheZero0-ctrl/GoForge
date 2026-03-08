package command

import (
	"context"
	"strings"

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
	Params    map[string]string
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

func (i Input) Param(key string) string {
	if i.Params == nil {
		return ""
	}
	return strings.TrimSpace(i.Params[key])
}

func (i Input) BoolParam(key string) bool {
	return strings.EqualFold(i.Param(key), "true")
}

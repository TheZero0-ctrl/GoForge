package newapp

import (
	"context"
	"path/filepath"

	"goforge/internal/domain/params"
	"goforge/internal/domain/plan"
)

func Validate(args []string, p params.Params) error {
	_, err := ParseConfig(args, p)
	return err
}

func Plan(_ context.Context, args []string, p params.Params) (plan.Plan, error) {
	cfg, err := ParseConfig(args, p)

	if err != nil {
		return plan.Plan{}, err
	}

	rendered, err := renderFiles(cfg)

	if err != nil {
		return plan.Plan{}, err
	}

	ops := make([]plan.Operation, 0, len(rendered)+5)
	ops = append(
		ops,
		plan.Operation{Type: plan.OpEnsureEmptyDir, Path: cfg.AppName},
		plan.Operation{Type: plan.OpMkdir, Path: cfg.AppName},
		plan.Operation{Type: plan.OpRun, Path: cfg.AppName, Cmd: []string{"go", "mod", "init", cfg.ModulePath}},
	)

	for _, f := range rendered {
		ops = append(ops, plan.Operation{
			Type: plan.OpWriteFile,
			Path: filepath.Join(cfg.AppName, f.Path),
			Data: f.Data,
		})
	}

	if !cfg.SkipTidy {
		ops = append(ops, plan.Operation{Type: plan.OpRun, Path: cfg.AppName, Cmd: []string{"go", "mod", "tidy"}})
	}

	if !cfg.SkipGit {
		ops = append(ops, plan.Operation{Type: plan.OpRun, Path: cfg.AppName, Cmd: []string{"git", "init"}})
	}

	return plan.Plan{
		CommandID:   "new",
		Description: "Create a new Go API app",
		Ops:         ops,
	}, nil
}

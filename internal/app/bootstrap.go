package app

import (
	"goforge/internal/domain/command"
	"goforge/internal/infra/fs"
	"goforge/internal/infra/proc"
)

func NewDefaultRegistry() (*command.Registry, error) {
	reg := command.NewRegistry()

	for _, cmd := range []command.Command{
		command.NewNewCommand(),
		command.NewGenerateCommand(),
		command.NewDestroyCommand(),
	} {
		if err := reg.Register(cmd); err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func NewDefaultExecutor(reg *command.Registry) *Executor {
	return NewExecutor(reg, fs.NewOSFS(), proc.NewOSRunner())
}

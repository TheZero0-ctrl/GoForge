package command

import (
	"fmt"
	"sort"
)

type Registry struct {
	commands map[string]Command
	aliases  map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

func (r *Registry) Register(cmd Command) error {
	spec := cmd.Spec()
	if spec.ID == "" {
		return fmt.Errorf("command id cannot be empty")
	}

	if _, exists := r.commands[spec.ID]; exists {
		return fmt.Errorf("command %q already registered", spec.ID)
	}

	r.commands[spec.ID] = cmd

	for _, alias := range spec.Aliases {
		if alias == "" {
			continue
		}
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %q already registered", alias)
		}
		r.aliases[alias] = spec.ID
	}

	return nil
}

func (r *Registry) Resolve(idOrAlias string) (Command, bool) {
	if cmd, ok := r.commands[idOrAlias]; ok {
		return cmd, true
	}

	id, ok := r.aliases[idOrAlias]
	if !ok {
		return nil, false
	}

	cmd, ok := r.commands[id]
	return cmd, ok
}

func (r *Registry) List() []Command {
	keys := make([]string, 0, len(r.commands))
	for k := range r.commands {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	out := make([]Command, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.commands[k])
	}
	return out
}

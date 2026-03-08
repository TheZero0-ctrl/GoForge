package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"goforge/internal/app"
	"goforge/internal/domain/command"
)

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return "command failed"
}

func (e exitError) ExitCode() int {
	return e.code
}

type rootOptions struct {
	DryRun bool
	Force  bool
	Skip   bool
}

func Run(ctx context.Context, stdout, stderr io.Writer) int {
	registry, err := app.NewDefaultRegistry()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return int(app.ExitExecution)
	}

	executor := app.NewDefaultExecutor(registry)
	root := newRootCommand(executor, registry, stdout, stderr)

	err = root.ExecuteContext(ctx)
	if err == nil {
		return int(app.ExitOK)
	}

	var coded exitError
	if errors.As(err, &coded) {
		return coded.code
	}

	fmt.Fprintln(stderr, err)
	return int(app.ExitExecution)
}

func newRootCommand(executor *app.Executor, registry *command.Registry, stdout, stderr io.Writer) *cobra.Command {
	opts := &rootOptions{}

	root := &cobra.Command{
		Use:           "goforge",
		Short:         "Rails-inspired CLI for Go API projects",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false, "Plan changes without writing files")
	root.PersistentFlags().BoolVar(&opts.Force, "force", false, "Overwrite existing files when possible")
	root.PersistentFlags().BoolVar(&opts.Skip, "skip", false, "Skip files that already exist")

	for _, cmd := range registry.List() {
		spec := cmd.Spec()

		module := ""
		skipGit := false
		skipTidy := false

		cobraCmd := &cobra.Command{
			Use:     spec.Use,
			Short:   spec.Short,
			Aliases: spec.Aliases,
			RunE: func(c *cobra.Command, args []string) error {
				params := map[string]string{}

				if spec.ID == "new" {
					params["module"] = module
					params["skip-git"] = fmt.Sprintf("%t", skipGit)
					params["skip-tidy"] = fmt.Sprintf("%t", skipTidy)
				}

				input := command.Input{
					CommandID: spec.ID,
					Args:      args,
					Flags: command.Flags{
						DryRun: opts.DryRun,
						Force:  opts.Force,
						Skip:   opts.Skip,
					},
					Params: params,
				}

				result := executor.Execute(c.Context(), input)
				printResult(stdout, stderr, result)

				if result.Code != app.ExitOK {
					return exitError{code: int(result.Code)}
				}

				return nil
			},
		}

		if spec.ID == "new" {
			cobraCmd.Flags().StringVar(&module, "module", "", "Explicit Go module path")
			cobraCmd.Flags().BoolVar(&skipGit, "skip-git", false, "Skip git init")
			cobraCmd.Flags().BoolVar(&skipTidy, "skip-tidy", false, "Skip go mod tidy")
		}
		root.AddCommand(cobraCmd)
	}

	return root
}

func printResult(stdout, stderr io.Writer, result app.Result) {
	for _, entry := range result.Entries {
		line := fmt.Sprintf("%s %s", entry.Status, entry.Message)
		if entry.Status == "ERROR" {
			fmt.Fprintln(stderr, line)
			continue
		}
		fmt.Fprintln(stdout, line)
	}
}

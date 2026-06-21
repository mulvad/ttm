package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/mulvad/ttm/internal/resolver"
	"github.com/spf13/cobra"
)

// NewResolveCmd creates the resolve command.
func NewResolveCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Show the full resolution chain without applying",
		Long: `Resolve displays the complete resolution chain from project profile
to terminal profile without actually applying any changes.

This is useful for debugging and understanding how ttm resolves
the terminal profile for the current project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResolve(configPath, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

func runResolve(configPath string, deps *Deps, w io.Writer) error {
	if deps == nil {
		deps = &Deps{}
	}
	if deps.ConfigLoader == nil {
		deps.ConfigLoader = DefaultConfigLoader{}
	}
	if deps.ProfileFinder == nil {
		deps.ProfileFinder = resolver.NewFinder(nil)
	}
	if deps.Getwd == nil {
		deps.Getwd = os.Getwd
	}

	// Load global config
	if configPath == "" {
		var err error
		configPath, err = deps.ConfigLoader.DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	cfg, err := deps.ConfigLoader.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find project profile
	cwd, err := deps.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	project, err := deps.ProfileFinder.FindAndLoadProfile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load project profile: %w", err)
	}

	if project == nil {
		return fmt.Errorf("no .terminal-profile found in %s or any parent directory", cwd)
	}

	// Resolve and show chain
	res := resolver.NewResolver(cfg)
	resolution, err := res.Resolve(project)
	if err != nil {
		return fmt.Errorf("failed to resolve profile: %w", err)
	}

	_, _ = fmt.Fprintln(w, "Resolution chain:")
	_, _ = fmt.Fprintln(w)

	for i, step := range resolution.Steps {
		prefix := "├──"
		if i == len(resolution.Steps)-1 {
			prefix = "└──"
		}
		_, _ = fmt.Fprintf(w, "%s [%s] %s → %s\n", prefix, step.Type, step.Key, step.Value)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Final profile: %s\n", resolution.Profile)

	return nil
}

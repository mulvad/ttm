package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mulvad/ttm/internal/resolver"
	"github.com/mulvad/ttm/internal/terminal"
	"github.com/spf13/cobra"
)

// NewCurrentCmd creates the current command.
func NewCurrentCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current terminal and project profile status",
		Long: `Current displays information about:
- The active terminal backend
- The current terminal profile
- The project profile (if any)
- The resolved theme`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent(cmd.Context(), configPath, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

func runCurrent(ctx context.Context, configPath string, deps *Deps, w io.Writer) error {
	if deps == nil {
		deps = &Deps{}
	}
	if deps.Backend == nil {
		deps.Backend = terminal.NewAppleTerminal()
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

	// Get terminal backend info
	_, _ = fmt.Fprintf(w, "Terminal: %s\n", deps.Backend.Name())
	_, _ = fmt.Fprintf(w, "Available: %v\n", deps.Backend.Available())

	if deps.Backend.Available() {
		currentProfile, err := deps.Backend.CurrentProfile(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(w, "Current profile: <error: %v>\n", err)
		} else {
			_, _ = fmt.Fprintf(w, "Current profile: %s\n", currentProfile)
		}
	}

	_, _ = fmt.Fprintln(w)

	// Find project profile
	cwd, err := deps.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	project, err := deps.ProfileFinder.FindAndLoadProfile(cwd)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Project profile: <error: %v>\n", err)
		return nil
	}

	if project == nil {
		_, _ = fmt.Fprintln(w, "Project profile: none")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Project profile: %s\n", project.Path)
	if project.UsesEnvironment() {
		_, _ = fmt.Fprintf(w, "  environment: %s\n", project.Environment)
	} else {
		_, _ = fmt.Fprintf(w, "  theme: %s\n", project.Theme)
	}

	// Try to resolve theme
	if configPath == "" {
		configPath, err = deps.ConfigLoader.DefaultConfigPath()
		if err != nil {
			_, _ = fmt.Fprintf(w, "\nResolved profile: <error: %v>\n", err)
			return nil
		}
	}

	cfg, err := deps.ConfigLoader.LoadConfig(configPath)
	if err != nil {
		_, _ = fmt.Fprintf(w, "\nResolved profile: <error loading config: %v>\n", err)
		return nil
	}

	res := resolver.NewResolver(cfg)
	resolution, err := res.Resolve(project)
	if err != nil {
		_, _ = fmt.Fprintf(w, "\nResolved profile: <error: %v>\n", err)
		return nil
	}

	_, _ = fmt.Fprintf(w, "\nResolved profile: %s\n", resolution.Profile)

	return nil
}

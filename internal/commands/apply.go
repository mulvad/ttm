// Package commands implements the CLI commands for ttm.
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

// NewApplyCmd creates the apply command.
func NewApplyCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terminal profile for the current project",
		Long: `Apply resolves the terminal profile for the current directory
based on the .terminal-profile file and global configuration,
then applies it to the terminal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd.Context(), configPath, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

// runApply executes the apply command. If deps is nil, uses defaults.
func runApply(ctx context.Context, configPath string, deps *Deps, w io.Writer) error {
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
		// No .terminal-profile found - clear any existing badge
		_ = deps.Backend.SetWindowTitle(ctx, "")
		return nil
	}

	// Resolve to terminal profile
	res := resolver.NewResolver(cfg)
	resolution, err := res.Resolve(project)
	if err != nil {
		return fmt.Errorf("failed to resolve profile: %w", err)
	}

	// Apply via backend
	if !deps.Backend.Available() {
		return fmt.Errorf("terminal backend not available")
	}

	if err := deps.Backend.ApplyProfile(ctx, resolution.Profile); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Applied profile: %s\n", resolution.Profile)

	// Set or clear window title badge
	if err := deps.Backend.SetWindowTitle(ctx, resolution.Badge); err != nil {
		_, _ = fmt.Fprintf(w, "Warning: failed to set window title: %v\n", err)
	} else if resolution.Badge != "" {
		_, _ = fmt.Fprintf(w, "Set badge: %s\n", resolution.Badge)
	}

	return nil
}

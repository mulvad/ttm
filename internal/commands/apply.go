// Package commands implements the CLI commands for ttm.
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/mulvad/ttm/internal/config"
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
			return runApply(cmd.Context(), configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

func runApply(ctx context.Context, configPath string) error {
	// Load global config
	if configPath == "" {
		var err error
		configPath, err = config.DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find project profile
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	finder := resolver.NewFinder(nil)
	project, err := finder.FindAndLoadProfile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load project profile: %w", err)
	}

	if project == nil {
		return fmt.Errorf("no .terminal-profile found in %s or any parent directory", cwd)
	}

	// Resolve to terminal profile
	res := resolver.NewResolver(cfg)
	profile, err := res.ResolveProfile(project)
	if err != nil {
		return fmt.Errorf("failed to resolve profile: %w", err)
	}

	// Apply via backend
	backend := terminal.NewAppleTerminal()
	if !backend.Available() {
		return fmt.Errorf("Apple Terminal backend not available")
	}

	if err := backend.ApplyProfile(ctx, profile); err != nil {
		return err
	}

	fmt.Printf("Applied profile: %s\n", profile)
	return nil
}

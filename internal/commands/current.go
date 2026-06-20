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
			return runCurrent(cmd.Context(), configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

func runCurrent(ctx context.Context, configPath string) error {
	// Get terminal backend info
	backend := terminal.NewAppleTerminal()
	fmt.Printf("Terminal: %s\n", backend.Name())
	fmt.Printf("Available: %v\n", backend.Available())

	if backend.Available() {
		currentProfile, err := backend.CurrentProfile(ctx)
		if err != nil {
			fmt.Printf("Current profile: <error: %v>\n", err)
		} else {
			fmt.Printf("Current profile: %s\n", currentProfile)
		}
	}

	fmt.Println()

	// Find project profile
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	finder := resolver.NewFinder(nil)
	project, err := finder.FindAndLoadProfile(cwd)
	if err != nil {
		fmt.Printf("Project profile: <error: %v>\n", err)
		return nil
	}

	if project == nil {
		fmt.Println("Project profile: none")
		return nil
	}

	fmt.Printf("Project profile: %s\n", project.Path)
	if project.UsesEnvironment() {
		fmt.Printf("  environment: %s\n", project.Environment)
	} else {
		fmt.Printf("  theme: %s\n", project.Theme)
	}

	// Try to resolve theme
	if configPath == "" {
		configPath, err = config.DefaultConfigPath()
		if err != nil {
			fmt.Printf("\nResolved theme: <error: %v>\n", err)
			return nil
		}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("\nResolved theme: <error loading config: %v>\n", err)
		return nil
	}

	res := resolver.NewResolver(cfg)
	resolution, err := res.Resolve(project)
	if err != nil {
		fmt.Printf("\nResolved theme: <error: %v>\n", err)
		return nil
	}

	fmt.Printf("\nResolved profile: %s\n", resolution.Profile)

	return nil
}

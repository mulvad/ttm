package commands

import (
	"fmt"
	"os"

	"github.com/mulvad/ttm/internal/config"
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
			return runResolve(configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

func runResolve(configPath string) error {
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

	// Resolve and show chain
	res := resolver.NewResolver(cfg)
	resolution, err := res.Resolve(project)
	if err != nil {
		return fmt.Errorf("failed to resolve profile: %w", err)
	}

	fmt.Println("Resolution chain:")
	fmt.Println()

	for i, step := range resolution.Steps {
		prefix := "├──"
		if i == len(resolution.Steps)-1 {
			prefix = "└──"
		}
		fmt.Printf("%s [%s] %s → %s\n", prefix, step.Type, step.Key, step.Value)
	}

	fmt.Println()
	fmt.Printf("Final profile: %s\n", resolution.Profile)

	return nil
}

package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// NewEnvironmentsCmd creates the environments command.
func NewEnvironmentsCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "environments",
		Short: "List all available environments",
		Long:  `List all environments defined in your global configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvironments(configPath, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

// runEnvironments executes the environments command. If deps is nil, uses defaults.
func runEnvironments(configPath string, deps *Deps, w io.Writer) error {
	if deps == nil {
		deps = &Deps{}
	}
	if deps.ConfigLoader == nil {
		deps.ConfigLoader = DefaultConfigLoader{}
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

	if len(cfg.Environments) == 0 {
		_, _ = fmt.Fprintln(w, "No environments configured")
		return nil
	}

	_, _ = fmt.Fprintln(w, "Available environments:")
	for name, env := range cfg.Environments {
		badge := ""
		if env.Badge != "" {
			badge = fmt.Sprintf(" [%s]", env.Badge)
		}
		_, _ = fmt.Fprintf(w, "  %s → %s%s\n", name, env.Theme, badge)
	}

	return nil
}

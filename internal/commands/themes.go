package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// NewThemesCmd creates the themes command.
func NewThemesCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "themes",
		Short: "List all available themes",
		Long:  `List all themes defined in your global configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThemes(configPath, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")

	return cmd
}

// runThemes executes the themes command. If deps is nil, uses defaults.
func runThemes(configPath string, deps *Deps, w io.Writer) error {
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

	if len(cfg.Themes) == 0 {
		_, _ = fmt.Fprintln(w, "No themes configured")
		return nil
	}

	_, _ = fmt.Fprintln(w, "Available themes:")
	for name, theme := range cfg.Themes {
		_, _ = fmt.Fprintf(w, "  %s → %s\n", name, theme.Profile)
	}

	return nil
}

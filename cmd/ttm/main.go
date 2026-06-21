// Package main provides the CLI entry point for ttm.
package main

import (
	"fmt"
	"os"

	"github.com/mulvad/ttm/internal/commands"
	"github.com/spf13/cobra"
)

// Version is set at build time.
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := &cobra.Command{
		Use:   "ttm",
		Short: "ttm (Terminal Theme Manager) - manage terminal themes based on project context",
		Long: `ttm (Terminal Theme Manager) manages terminal themes based on project context
using a three-layer architecture:

1. Project metadata (.terminal-profile) - Per-project YAML config
2. Semantic environment mapping - Maps environments to themes
3. Terminal-specific implementation - Apple Terminal profiles via AppleScript

Example usage:
  ttm apply    # Apply the terminal profile for the current project
  ttm current  # Show current terminal and project status
  ttm resolve  # Show the full resolution chain without applying
  ttm export   # Export terminal profiles to a file
  ttm import   # Import terminal profiles from a file`,
		Version: Version,
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewApplyCmd())
	rootCmd.AddCommand(commands.NewCurrentCmd())
	rootCmd.AddCommand(commands.NewResolveCmd())
	rootCmd.AddCommand(commands.NewExportCmd())
	rootCmd.AddCommand(commands.NewImportCmd())
	rootCmd.AddCommand(commands.NewVerifyCmd())

	return rootCmd.Execute()
}

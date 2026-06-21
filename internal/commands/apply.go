// Package commands implements the CLI commands for ttm.
package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mulvad/ttm/internal/resolver"
	"github.com/mulvad/ttm/internal/terminal"
	"github.com/spf13/cobra"
)

// NewApplyCmd creates the apply command.
func NewApplyCmd() *cobra.Command {
	var configPath string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terminal theme for the current project",
		Long: `Apply resolves the terminal theme for the current directory
based on the .terminal-profile file and global configuration,
then applies it to the terminal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(cmd.Context(), configPath, verbose, nil, os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file (default: ~/.ttm/config.yaml)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed output")

	return cmd
}

// runApply executes the apply command. If deps is nil, uses defaults.
func runApply(ctx context.Context, configPath string, verbose bool, deps *Deps, w io.Writer) error {
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
		// No .terminal-profile found - restore original profile if we have one
		originalProfile, err := loadOriginalProfile()
		if err == nil && originalProfile != "" {
			if err := deps.Backend.ApplyProfile(ctx, originalProfile); err != nil {
				return fmt.Errorf("failed to restore original theme: %w", err)
			}
			if verbose {
				_, _ = fmt.Fprintf(w, "Restored default theme\n")
			}
		} else if cfg.DefaultProfile != "" {
			// Fall back to configured default profile
			if err := deps.Backend.ApplyProfile(ctx, cfg.DefaultProfile); err != nil {
				return fmt.Errorf("failed to apply default theme: %w", err)
			}
			if verbose {
				_, _ = fmt.Fprintf(w, "Applied default theme\n")
			}
		}
		// Clear any existing badge
		_ = deps.Backend.SetWindowTitle(ctx, "")
		return nil
	}

	// Store original profile before applying a new one (if not already stored)
	if err := saveOriginalProfileIfNeeded(ctx, deps.Backend); err != nil {
		if verbose {
			_, _ = fmt.Fprintf(w, "Warning: failed to save original theme: %v\n", err)
		}
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

	if verbose {
		if resolution.Environment != "" {
			_, _ = fmt.Fprintf(w, "Applied environment: %s\n", resolution.Environment)
		} else {
			_, _ = fmt.Fprintf(w, "Applied theme\n")
		}
	}

	// Set or clear window title badge
	if err := deps.Backend.SetWindowTitle(ctx, resolution.Badge); err != nil {
		if verbose {
			_, _ = fmt.Fprintf(w, "Warning: failed to set badge: %v\n", err)
		}
	}

	return nil
}

// originalProfilePath returns the path to the file storing the original profile.
func originalProfilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ttm", "original-profile"), nil
}

// loadOriginalProfile loads the stored original profile name.
func loadOriginalProfile() (string, error) {
	path, err := originalProfilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// saveOriginalProfileIfNeeded saves the current profile as the original if not already saved.
func saveOriginalProfileIfNeeded(ctx context.Context, backend Backend) error {
	path, err := originalProfilePath()
	if err != nil {
		return err
	}

	// Check if already saved (and non-empty)
	if data, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(data)) != "" {
		return nil // Already exists with content, don't overwrite
	}

	// Get current profile from terminal
	currentProfile, err := backend.CurrentProfile(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	// Save it
	if err := os.WriteFile(path, []byte(currentProfile), 0644); err != nil {
		return fmt.Errorf("failed to save original profile: %w", err)
	}

	return nil
}

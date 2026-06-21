package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mulvad/ttm/internal/terminal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewExportCmd creates the export command.
func NewExportCmd() *cobra.Command {
	var outputPath string
	var profileNames []string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export terminal profiles to a file",
		Long: `Export terminal profiles to a YAML file for backup or sharing.

By default, exports all profiles. Use --profile to export specific profiles.

Examples:
  ttm export -o profiles.yaml              # Export all profiles
  ttm export -o my.yaml -p "Pro" -p "Basic" # Export specific profiles`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cmd.Context(), outputPath, profileNames, nil, os.Stdout, nil)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "profiles.yaml", "output file path")
	cmd.Flags().StringArrayVarP(&profileNames, "profile", "p", nil, "profile name to export (can be repeated)")

	return cmd
}

// FileWriter writes data to a file. Used for testing.
type FileWriter func(path string, data []byte, perm os.FileMode) error

func runExport(ctx context.Context, outputPath string, profileNames []string, deps *Deps, w io.Writer, writeFile FileWriter) error {
	if deps == nil {
		deps = &Deps{}
	}
	if deps.Backend == nil {
		deps.Backend = terminal.NewAppleTerminal()
	}
	if writeFile == nil {
		writeFile = os.WriteFile
	}

	if !deps.Backend.Available() {
		return fmt.Errorf("terminal backend not available")
	}

	var profiles []*terminal.Profile
	var err error

	if len(profileNames) == 0 {
		// Export all profiles
		profiles, err = deps.Backend.ExportAllProfiles(ctx)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "Exporting %d profiles...\n", len(profiles))
	} else {
		// Export specific profiles
		profiles = make([]*terminal.Profile, 0, len(profileNames))
		for _, name := range profileNames {
			profile, err := deps.Backend.ExportProfile(ctx, name)
			if err != nil {
				return err
			}
			profiles = append(profiles, profile)
		}
		_, _ = fmt.Fprintf(w, "Exporting %d profiles...\n", len(profiles))
	}

	// Write to file
	data, err := yaml.Marshal(map[string][]*terminal.Profile{"profiles": profiles})
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := writeFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Exported to %s\n", outputPath)
	return nil
}

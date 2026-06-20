package commands

import (
	"context"
	"fmt"
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
			return runExport(cmd.Context(), outputPath, profileNames)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "profiles.yaml", "output file path")
	cmd.Flags().StringArrayVarP(&profileNames, "profile", "p", nil, "profile name to export (can be repeated)")

	return cmd
}

func runExport(ctx context.Context, outputPath string, profileNames []string) error {
	backend := terminal.NewAppleTerminal()
	if !backend.Available() {
		return fmt.Errorf("Apple Terminal backend not available")
	}

	var profiles []*terminal.Profile
	var err error

	if len(profileNames) == 0 {
		// Export all profiles
		profiles, err = backend.ExportAllProfiles(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Exporting %d profiles...\n", len(profiles))
	} else {
		// Export specific profiles
		profiles = make([]*terminal.Profile, 0, len(profileNames))
		for _, name := range profileNames {
			profile, err := backend.ExportProfile(ctx, name)
			if err != nil {
				return err
			}
			profiles = append(profiles, profile)
		}
		fmt.Printf("Exporting %d profiles...\n", len(profiles))
	}

	// Write to file
	data, err := yaml.Marshal(map[string][]*terminal.Profile{"profiles": profiles})
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported to %s\n", outputPath)
	return nil
}

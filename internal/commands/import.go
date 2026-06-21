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

// ProfilesFile represents the structure of an exported profiles file.
type ProfilesFile struct {
	Profiles []*terminal.Profile `yaml:"profiles"`
}

// NewImportCmd creates the import command.
func NewImportCmd() *cobra.Command {
	var inputPath string
	var profileNames []string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import terminal profiles from a file",
		Long: `Import terminal profiles from a YAML file.

Creates new profiles or updates existing ones with the same name.

Examples:
  ttm import -i profiles.yaml              # Import all profiles from file
  ttm import -i my.yaml -p "Pro" -p "Basic" # Import specific profiles`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(cmd.Context(), inputPath, profileNames, nil, os.Stdout, nil)
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "profiles.yaml", "input file path")
	cmd.Flags().StringArrayVarP(&profileNames, "profile", "p", nil, "profile name to import (can be repeated)")

	return cmd
}

// FileReader reads data from a file. Used for testing.
type FileReader func(path string) ([]byte, error)

func runImport(ctx context.Context, inputPath string, profileNames []string, deps *Deps, w io.Writer, readFile FileReader) error {
	if deps == nil {
		deps = &Deps{}
	}
	if deps.Backend == nil {
		deps.Backend = terminal.NewAppleTerminal()
	}
	if readFile == nil {
		readFile = os.ReadFile
	}

	if !deps.Backend.Available() {
		return fmt.Errorf("terminal backend not available")
	}

	// Read file
	data, err := readFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var file ProfilesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if len(file.Profiles) == 0 {
		return fmt.Errorf("no profiles found in %s", inputPath)
	}

	// Filter profiles if specific names requested
	profilesToImport := file.Profiles
	if len(profileNames) > 0 {
		nameSet := make(map[string]bool)
		for _, name := range profileNames {
			nameSet[name] = true
		}

		filtered := make([]*terminal.Profile, 0)
		for _, p := range file.Profiles {
			if nameSet[p.Name] {
				filtered = append(filtered, p)
			}
		}
		profilesToImport = filtered

		if len(profilesToImport) == 0 {
			return fmt.Errorf("none of the specified profiles found in %s", inputPath)
		}
	}

	_, _ = fmt.Fprintf(w, "Importing %d profiles...\n", len(profilesToImport))

	for _, profile := range profilesToImport {
		if err := deps.Backend.ImportProfile(ctx, profile); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "  Imported: %s\n", profile.Name)
	}

	_, _ = fmt.Fprintln(w, "Done!")
	return nil
}

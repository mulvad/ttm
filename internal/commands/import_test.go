package commands

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mulvad/ttm/internal/terminal"
)

// mockBackendWithImport extends mockBackend with import functionality.
type mockBackendWithImport struct {
	mockBackend
	importedProfiles []*terminal.Profile
	importErr        error
}

func (m *mockBackendWithImport) ImportProfile(ctx context.Context, profile *terminal.Profile) error {
	if m.importErr != nil {
		return m.importErr
	}
	m.importedProfiles = append(m.importedProfiles, profile)
	return nil
}

func TestRunImport(t *testing.T) {
	validYAML := `profiles:
  - name: Basic
    font_name: Menlo
    font_size: 12
  - name: Pro
    font_name: Monaco
    font_size: 14
`

	tests := []struct {
		name         string
		inputPath    string
		profileNames []string
		deps         *Deps
		fileContent  string
		readErr      error
		wantErr      bool
		errContains  string
		wantOutput   []string
		wantImported int
	}{
		{
			name:         "import all profiles",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			fileContent:  validYAML,
			wantErr:      false,
			wantOutput:   []string{"Importing 2 profiles", "Imported: Basic", "Imported: Pro", "Done!"},
			wantImported: 2,
		},
		{
			name:         "import specific profile",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: []string{"Pro"},
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			fileContent:  validYAML,
			wantErr:      false,
			wantOutput:   []string{"Importing 1 profiles", "Imported: Pro"},
			wantImported: 1,
		},
		{
			name:         "backend not available",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: false}},
			},
			fileContent: validYAML,
			wantErr:     true,
			errContains: "not available",
		},
		{
			name:         "file read error",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			readErr:     errors.New("file not found"),
			wantErr:     true,
			errContains: "failed to read",
		},
		{
			name:         "invalid yaml",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			fileContent: "invalid: yaml: content:",
			wantErr:     true,
			errContains: "failed to parse",
		},
		{
			name:         "empty profiles",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			fileContent: "profiles: []",
			wantErr:     true,
			errContains: "no profiles found",
		},
		{
			name:         "specified profile not found",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: []string{"NonExistent"},
			deps: &Deps{
				Backend: &mockBackendWithImport{mockBackend: mockBackend{available: true}},
			},
			fileContent: validYAML,
			wantErr:     true,
			errContains: "none of the specified profiles",
		},
		{
			name:         "import error",
			inputPath:    "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithImport{
					mockBackend: mockBackend{available: true},
					importErr:   errors.New("import failed"),
				},
			},
			fileContent: validYAML,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			readFile := func(path string) ([]byte, error) {
				if tt.readErr != nil {
					return nil, tt.readErr
				}
				return []byte(tt.fileContent), nil
			}

			err := runImport(context.Background(), tt.inputPath, tt.profileNames, tt.deps, &buf, readFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("runImport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errContains != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}

			if tt.wantImported > 0 {
				backend := tt.deps.Backend.(*mockBackendWithImport)
				if len(backend.importedProfiles) != tt.wantImported {
					t.Errorf("imported %d profiles, want %d", len(backend.importedProfiles), tt.wantImported)
				}
			}
		})
	}
}

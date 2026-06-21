package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/mulvad/ttm/internal/terminal"
)

// mockBackendWithExport extends mockBackend with export functionality.
type mockBackendWithExport struct {
	mockBackend
	exportedProfiles []*terminal.Profile
	exportErr        error
}

func (m *mockBackendWithExport) ExportProfile(ctx context.Context, name string) (*terminal.Profile, error) {
	if m.exportErr != nil {
		return nil, m.exportErr
	}
	return &terminal.Profile{Name: name, FontName: "Menlo", FontSize: 12}, nil
}

func (m *mockBackendWithExport) ExportAllProfiles(ctx context.Context) ([]*terminal.Profile, error) {
	if m.exportErr != nil {
		return nil, m.exportErr
	}
	return m.exportedProfiles, nil
}

func TestRunExport(t *testing.T) {
	tests := []struct {
		name         string
		outputPath   string
		profileNames []string
		deps         *Deps
		writeErr     error
		wantErr      bool
		errContains  string
		wantOutput   []string
		checkWritten func(t *testing.T, data []byte)
	}{
		{
			name:         "export all profiles",
			outputPath:   "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithExport{
					mockBackend:      mockBackend{available: true},
					exportedProfiles: []*terminal.Profile{{Name: "Basic"}, {Name: "Pro"}},
				},
			},
			wantErr:    false,
			wantOutput: []string{"Exporting 2 profiles", "Exported to"},
			checkWritten: func(t *testing.T, data []byte) {
				if !bytes.Contains(data, []byte("Basic")) {
					t.Error("written data should contain 'Basic'")
				}
			},
		},
		{
			name:         "export specific profiles",
			outputPath:   "/tmp/profiles.yaml",
			profileNames: []string{"Pro"},
			deps: &Deps{
				Backend: &mockBackendWithExport{
					mockBackend: mockBackend{available: true},
				},
			},
			wantErr:    false,
			wantOutput: []string{"Exporting 1 profiles"},
		},
		{
			name:         "backend not available",
			outputPath:   "/tmp/profiles.yaml",
			profileNames: nil,
			deps: &Deps{
				Backend: &mockBackendWithExport{
					mockBackend: mockBackend{available: false},
				},
			},
			wantErr:     true,
			errContains: "not available",
		},
		{
			name:         "export error",
			outputPath:   "/tmp/profiles.yaml",
			profileNames: []string{"Pro"},
			deps: &Deps{
				Backend: &mockBackendWithExport{
					mockBackend: mockBackend{available: true},
					exportErr:   errors.New("export failed"),
				},
			},
			wantErr: true,
		},
		{
			name:         "write error",
			outputPath:   "/tmp/profiles.yaml",
			profileNames: []string{"Pro"},
			deps: &Deps{
				Backend: &mockBackendWithExport{
					mockBackend: mockBackend{available: true},
				},
			},
			writeErr:    errors.New("write failed"),
			wantErr:     true,
			errContains: "failed to write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var writtenData []byte

			writeFile := func(path string, data []byte, perm os.FileMode) error {
				writtenData = data
				return tt.writeErr
			}

			err := runExport(context.Background(), tt.outputPath, tt.profileNames, tt.deps, &buf, writeFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("runExport() error = %v, wantErr %v", err, tt.wantErr)
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

			if tt.checkWritten != nil && writtenData != nil {
				tt.checkWritten(t, writtenData)
			}
		})
	}
}

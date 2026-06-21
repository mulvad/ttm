package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mulvad/ttm/internal/config"
)

func TestRunResolve(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		deps        *Deps
		wantErr     bool
		errContains string
		wantOutput  []string
	}{
		{
			name:       "successful resolve with environment",
			configPath: "/test/config.yaml",
			deps: &Deps{
				ConfigLoader: &mockConfigLoader{
					config: &config.Config{
						Environments: map[string]config.EnvironmentConfig{
							"production": {Theme: "prod"},
						},
						Themes: map[string]config.ThemeConfig{
							"prod": {Profile: "Red Sands"},
						},
					},
				},
				ProfileFinder: &mockProfileFinder{
					profile: &config.ProjectProfile{
						Environment: "production",
						Path:        "/project/.terminal-profile",
					},
				},
				Getwd: func() (string, error) { return "/project", nil },
			},
			wantErr:    false,
			wantOutput: []string{"Resolution chain:", "environment", "production", "Red Sands", "Final profile:"},
		},
		{
			name:       "successful resolve with direct theme",
			configPath: "/test/config.yaml",
			deps: &Deps{
				ConfigLoader: &mockConfigLoader{
					config: &config.Config{
						Themes: map[string]config.ThemeConfig{
							"dark": {Profile: "Pro"},
						},
					},
				},
				ProfileFinder: &mockProfileFinder{
					profile: &config.ProjectProfile{
						Theme: "dark",
						Path:  "/project/.terminal-profile",
					},
				},
				Getwd: func() (string, error) { return "/project", nil },
			},
			wantErr:    false,
			wantOutput: []string{"theme", "dark", "Pro"},
		},
		{
			name:       "no project profile",
			configPath: "/test/config.yaml",
			deps: &Deps{
				ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
				ProfileFinder: &mockProfileFinder{profile: nil},
				Getwd:         func() (string, error) { return "/project", nil },
			},
			wantErr:     true,
			errContains: "no .terminal-profile found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runResolve(tt.configPath, tt.deps, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("runResolve() error = %v, wantErr %v", err, tt.wantErr)
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
		})
	}
}

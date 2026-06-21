package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/mulvad/ttm/internal/config"
)

func TestRunCurrent(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		deps       *Deps
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "full output with project",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend: &mockBackend{available: true, currentProfile: "Ocean"},
				ConfigLoader: &mockConfigLoader{
					config: &config.Config{
						Environments: map[string]config.EnvironmentConfig{
							"staging": {Theme: "stage"},
						},
						Themes: map[string]config.ThemeConfig{
							"stage": {Profile: "Ocean"},
						},
					},
				},
				ProfileFinder: &mockProfileFinder{
					profile: &config.ProjectProfile{
						Environment: "staging",
						Path:        "/project/.terminal-profile",
					},
				},
				Getwd: func() (string, error) { return "/project", nil },
			},
			wantErr:    false,
			wantOutput: []string{"Terminal: Mock", "Available: true", "Current profile: Ocean", "environment: staging", "Resolved profile: Ocean"},
		},
		{
			name:       "no project profile",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend:       &mockBackend{available: true, currentProfile: "Basic"},
				ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
				ProfileFinder: &mockProfileFinder{profile: nil},
				Getwd:         func() (string, error) { return "/project", nil },
			},
			wantErr:    false,
			wantOutput: []string{"Terminal: Mock", "Project profile: none"},
		},
		{
			name:       "backend not available",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend:       &mockBackend{available: false},
				ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
				ProfileFinder: &mockProfileFinder{profile: nil},
				Getwd:         func() (string, error) { return "/project", nil },
			},
			wantErr:    false,
			wantOutput: []string{"Available: false"},
		},
		{
			name:       "direct theme",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend: &mockBackend{available: true, currentProfile: "Pro"},
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
			wantOutput: []string{"theme: dark", "Resolved profile: Pro"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runCurrent(context.Background(), tt.configPath, tt.deps, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("runCurrent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

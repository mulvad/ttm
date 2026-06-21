package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mulvad/ttm/internal/config"
	"github.com/mulvad/ttm/internal/terminal"
)

// mockBackend implements Backend for testing.
type mockBackend struct {
	available      bool
	currentProfile string
	appliedProfile string
	applyErr       error
	profiles       []string
	windowTitle    string
	setTitleErr    error
}

func (m *mockBackend) Name() string                    { return "Mock" }
func (m *mockBackend) Available() bool                 { return m.available }
func (m *mockBackend) CurrentProfile(ctx context.Context) (string, error) {
	return m.currentProfile, nil
}
func (m *mockBackend) ApplyProfile(ctx context.Context, profile string) error {
	m.appliedProfile = profile
	return m.applyErr
}
func (m *mockBackend) ListProfiles(ctx context.Context) ([]string, error) {
	return m.profiles, nil
}
func (m *mockBackend) ExportProfile(ctx context.Context, name string) (*terminal.Profile, error) {
	return &terminal.Profile{Name: name}, nil
}
func (m *mockBackend) ExportAllProfiles(ctx context.Context) ([]*terminal.Profile, error) {
	return nil, nil
}
func (m *mockBackend) ImportProfile(ctx context.Context, profile *terminal.Profile) error {
	return nil
}
func (m *mockBackend) SetWindowTitle(ctx context.Context, title string) error {
	m.windowTitle = title
	return m.setTitleErr
}
func (m *mockBackend) GetWindowTitle(ctx context.Context) (string, error) {
	return m.windowTitle, nil
}

// mockConfigLoader implements ConfigLoader for testing.
type mockConfigLoader struct {
	config    *config.Config
	configErr error
	path      string
	pathErr   error
}

func (m *mockConfigLoader) LoadConfig(path string) (*config.Config, error) {
	return m.config, m.configErr
}

func (m *mockConfigLoader) DefaultConfigPath() (string, error) {
	return m.path, m.pathErr
}

// mockProfileFinder implements ProfileFinder for testing.
type mockProfileFinder struct {
	profile *config.ProjectProfile
	err     error
}

func (m *mockProfileFinder) FindAndLoadProfile(startDir string) (*config.ProjectProfile, error) {
	return m.profile, m.err
}

func TestRunApply(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		deps       *Deps
		wantErr    bool
		errContains string
		wantProfile string
	}{
		{
			name:       "successful apply",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend: &mockBackend{available: true},
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
			wantErr:     false,
			wantProfile: "Red Sands",
		},
		{
			name:       "backend not available",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend: &mockBackend{available: false},
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
			wantErr:     true,
			errContains: "not available",
		},
		{
			name:       "no project profile found - clears badge silently",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend:       &mockBackend{available: true},
				ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
				ProfileFinder: &mockProfileFinder{profile: nil},
				Getwd:         func() (string, error) { return "/project", nil },
			},
			wantErr: false,
		},
		{
			name:       "config load error",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend:       &mockBackend{available: true},
				ConfigLoader:  &mockConfigLoader{configErr: errors.New("config not found")},
				ProfileFinder: &mockProfileFinder{},
				Getwd:         func() (string, error) { return "/project", nil },
			},
			wantErr:     true,
			errContains: "failed to load config",
		},
		{
			name:       "getwd error",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend:       &mockBackend{available: true},
				ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
				ProfileFinder: &mockProfileFinder{},
				Getwd:         func() (string, error) { return "", errors.New("no cwd") },
			},
			wantErr:     true,
			errContains: "failed to get current directory",
		},
		{
			name:       "apply error",
			configPath: "/test/config.yaml",
			deps: &Deps{
				Backend: &mockBackend{available: true, applyErr: errors.New("apply failed")},
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runApply(context.Background(), tt.configPath, false, tt.deps, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("runApply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errContains != "" && err != nil {
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
			}

			if tt.wantProfile != "" {
				backend := tt.deps.Backend.(*mockBackend)
				if backend.appliedProfile != tt.wantProfile {
					t.Errorf("applied profile = %q, want %q", backend.appliedProfile, tt.wantProfile)
				}
			}
		})
	}
}

func TestRunApply_SetsBadge(t *testing.T) {
	backend := &mockBackend{available: true}
	deps := &Deps{
		Backend: backend,
		ConfigLoader: &mockConfigLoader{
			config: &config.Config{
				Environments: map[string]config.EnvironmentConfig{
					"production": {Theme: "prod", Badge: "PROD"},
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
	}

	var buf bytes.Buffer
	err := runApply(context.Background(), "/test/config.yaml", false, deps, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend.windowTitle != "PROD" {
		t.Errorf("window title = %q, want %q", backend.windowTitle, "PROD")
	}
}

func TestRunApply_ClearsBadgeWhenNoBadge(t *testing.T) {
	backend := &mockBackend{available: true, windowTitle: "OLD_BADGE"}
	deps := &Deps{
		Backend: backend,
		ConfigLoader: &mockConfigLoader{
			config: &config.Config{
				Themes: map[string]config.ThemeConfig{
					"dev": {Profile: "Grass"},
				},
			},
		},
		ProfileFinder: &mockProfileFinder{
			profile: &config.ProjectProfile{
				Theme: "dev", // Using theme directly, no badge
				Path:  "/project/.terminal-profile",
			},
		},
		Getwd: func() (string, error) { return "/project", nil },
	}

	var buf bytes.Buffer
	err := runApply(context.Background(), "/test/config.yaml", false, deps, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Badge should be cleared (empty string)
	if backend.windowTitle != "" {
		t.Errorf("window title should be cleared, got %q", backend.windowTitle)
	}
}

func TestRunApply_ClearsBadgeWhenNoProfile(t *testing.T) {
	backend := &mockBackend{available: true, windowTitle: "OLD_BADGE"}
	deps := &Deps{
		Backend:       backend,
		ConfigLoader:  &mockConfigLoader{config: &config.Config{}},
		ProfileFinder: &mockProfileFinder{profile: nil}, // No .terminal-profile found
		Getwd:         func() (string, error) { return "/project", nil },
	}

	var buf bytes.Buffer
	err := runApply(context.Background(), "/test/config.yaml", false, deps, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Badge should be cleared
	if backend.windowTitle != "" {
		t.Errorf("window title should be cleared, got %q", backend.windowTitle)
	}
}

func TestRunApply_Integration(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `environments:
  test:
    theme: testtheme
themes:
  testtheme:
    profile: "Test Profile"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create project profile
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	profilePath := filepath.Join(projectDir, ".terminal-profile")
	if err := os.WriteFile(profilePath, []byte("environment: test\n"), 0644); err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Run with mock backend but real config/finder
	backend := &mockBackend{available: true}
	var buf bytes.Buffer
	deps := &Deps{
		Backend:       backend,
		ConfigLoader:  DefaultConfigLoader{},
		ProfileFinder: nil, // Will use default
		Getwd:         func() (string, error) { return projectDir, nil },
	}

	err := runApply(context.Background(), configPath, false, deps, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend.appliedProfile != "Test Profile" {
		t.Errorf("applied profile = %q, want 'Test Profile'", backend.appliedProfile)
	}
}

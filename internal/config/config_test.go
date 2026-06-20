package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config",
			content: `environments:
  production:
    theme: prod
  staging:
    theme: stage
themes:
  prod:
    profile: "Red Sands"
  stage:
    profile: "Ocean"
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Environments) != 2 {
					t.Errorf("expected 2 environments, got %d", len(cfg.Environments))
				}
				if len(cfg.Themes) != 2 {
					t.Errorf("expected 2 themes, got %d", len(cfg.Themes))
				}
				if cfg.Environments["production"].Theme != "prod" {
					t.Errorf("expected production theme to be 'prod', got %q", cfg.Environments["production"].Theme)
				}
				if cfg.Themes["prod"].Profile != "Red Sands" {
					t.Errorf("expected prod profile to be 'Red Sands', got %q", cfg.Themes["prod"].Profile)
				}
			},
		},
		{
			name: "environment references undefined theme",
			content: `environments:
  production:
    theme: nonexistent
themes:
  prod:
    profile: "Basic"
`,
			wantErr: true,
		},
		{
			name: "theme without profile",
			content: `environments:
  production:
    theme: prod
themes:
  prod:
    profile: ""
`,
			wantErr: true,
		},
		{
			name: "environment without theme",
			content: `environments:
  production:
    theme: ""
themes:
  prod:
    profile: "Basic"
`,
			wantErr: true,
		},
		{
			name: "empty config is valid",
			content: `environments: {}
themes: {}
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if len(cfg.Environments) != 0 {
					t.Errorf("expected 0 environments, got %d", len(cfg.Environments))
				}
				if len(cfg.Themes) != 0 {
					t.Errorf("expected 0 themes, got %d", len(cfg.Themes))
				}
			},
		},
		{
			name:    "invalid yaml",
			content: `this is not: valid: yaml:`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			cfg, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestConfig_GetThemeProfile(t *testing.T) {
	cfg := &Config{
		Themes: map[string]ThemeConfig{
			"prod": {Profile: "Red Sands"},
		},
	}

	profile, err := cfg.GetThemeProfile("prod")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if profile != "Red Sands" {
		t.Errorf("expected 'Red Sands', got %q", profile)
	}

	_, err = cfg.GetThemeProfile("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent theme, got nil")
	}
}

func TestConfig_GetEnvironmentTheme(t *testing.T) {
	cfg := &Config{
		Environments: map[string]EnvironmentConfig{
			"production": {Theme: "prod"},
		},
	}

	theme, err := cfg.GetEnvironmentTheme("production")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if theme != "prod" {
		t.Errorf("expected 'prod', got %q", theme)
	}

	_, err = cfg.GetEnvironmentTheme("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent environment, got nil")
	}
}

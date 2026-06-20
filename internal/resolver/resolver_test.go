package resolver

import (
	"testing"

	"github.com/mulvad/ttm/internal/config"
)

func TestResolver_Resolve(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"production": {Theme: "prod"},
			"staging":    {Theme: "stage"},
			"development": {Theme: "dev"},
		},
		Themes: map[string]config.ThemeConfig{
			"prod":  {Profile: "Red Sands"},
			"stage": {Profile: "Ocean"},
			"dev":   {Profile: "Grass"},
			"custom": {Profile: "Novel"},
		},
	}

	tests := []struct {
		name        string
		project     *config.ProjectProfile
		wantProfile string
		wantSteps   int
		wantErr     bool
	}{
		{
			name: "resolve via environment",
			project: &config.ProjectProfile{
				Environment: "production",
				Path:        "/project/.terminal-profile",
			},
			wantProfile: "Red Sands",
			wantSteps:   4, // project -> environment -> theme -> profile
			wantErr:     false,
		},
		{
			name: "resolve via direct theme",
			project: &config.ProjectProfile{
				Theme: "custom",
				Path:  "/project/.terminal-profile",
			},
			wantProfile: "Novel",
			wantSteps:   3, // project -> theme -> profile
			wantErr:     false,
		},
		{
			name: "unknown environment",
			project: &config.ProjectProfile{
				Environment: "unknown",
				Path:        "/project/.terminal-profile",
			},
			wantErr: true,
		},
		{
			name: "unknown theme",
			project: &config.ProjectProfile{
				Theme: "unknown",
				Path:  "/project/.terminal-profile",
			},
			wantErr: true,
		},
		{
			name:    "nil project",
			project: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(cfg)
			res, err := resolver.Resolve(tt.project)

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if res.Profile != tt.wantProfile {
				t.Errorf("Resolve() profile = %q, want %q", res.Profile, tt.wantProfile)
			}

			if len(res.Steps) != tt.wantSteps {
				t.Errorf("Resolve() steps = %d, want %d", len(res.Steps), tt.wantSteps)
			}
		})
	}
}

func TestResolver_ResolveTheme(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"production": {Theme: "prod"},
		},
		Themes: map[string]config.ThemeConfig{
			"prod":   {Profile: "Red Sands"},
			"custom": {Profile: "Novel"},
		},
	}

	tests := []struct {
		name      string
		project   *config.ProjectProfile
		wantTheme string
		wantErr   bool
	}{
		{
			name: "resolve environment to theme",
			project: &config.ProjectProfile{
				Environment: "production",
			},
			wantTheme: "prod",
			wantErr:   false,
		},
		{
			name: "direct theme passthrough",
			project: &config.ProjectProfile{
				Theme: "custom",
			},
			wantTheme: "custom",
			wantErr:   false,
		},
		{
			name: "unknown environment",
			project: &config.ProjectProfile{
				Environment: "unknown",
			},
			wantErr: true,
		},
		{
			name:    "nil project",
			project: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(cfg)
			theme, err := resolver.ResolveTheme(tt.project)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTheme() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && theme != tt.wantTheme {
				t.Errorf("ResolveTheme() = %q, want %q", theme, tt.wantTheme)
			}
		})
	}
}

func TestResolver_ResolveProfile(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"production": {Theme: "prod"},
		},
		Themes: map[string]config.ThemeConfig{
			"prod": {Profile: "Red Sands"},
		},
	}

	resolver := NewResolver(cfg)
	project := &config.ProjectProfile{
		Environment: "production",
		Path:        "/project/.terminal-profile",
	}

	profile, err := resolver.ResolveProfile(project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile != "Red Sands" {
		t.Errorf("ResolveProfile() = %q, want %q", profile, "Red Sands")
	}
}

func TestResolution_StepDetails(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"production": {Theme: "prod"},
		},
		Themes: map[string]config.ThemeConfig{
			"prod": {Profile: "Red Sands"},
		},
	}

	resolver := NewResolver(cfg)
	project := &config.ProjectProfile{
		Environment: "production",
		Path:        "/project/.terminal-profile",
	}

	res, err := resolver.Resolve(project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify step types
	expectedTypes := []string{"project", "environment", "theme", "profile"}
	for i, step := range res.Steps {
		if step.Type != expectedTypes[i] {
			t.Errorf("Step %d type = %q, want %q", i, step.Type, expectedTypes[i])
		}
	}

	// Verify project step
	if res.Steps[0].Key != "/project/.terminal-profile" {
		t.Errorf("Project step key = %q, want '/project/.terminal-profile'", res.Steps[0].Key)
	}

	// Verify environment step
	if res.Steps[1].Key != "production" {
		t.Errorf("Environment step key = %q, want 'production'", res.Steps[1].Key)
	}
	if res.Steps[1].Value != "prod" {
		t.Errorf("Environment step value = %q, want 'prod'", res.Steps[1].Value)
	}
}

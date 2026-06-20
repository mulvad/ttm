package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProjectProfile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(t *testing.T, p *ProjectProfile)
	}{
		{
			name:    "environment only",
			content: "environment: production\n",
			wantErr: false,
			check: func(t *testing.T, p *ProjectProfile) {
				if p.Environment != "production" {
					t.Errorf("expected environment 'production', got %q", p.Environment)
				}
				if p.Theme != "" {
					t.Errorf("expected empty theme, got %q", p.Theme)
				}
				if !p.UsesEnvironment() {
					t.Error("expected UsesEnvironment() to be true")
				}
				if p.UsesTheme() {
					t.Error("expected UsesTheme() to be false")
				}
			},
		},
		{
			name:    "theme only",
			content: "theme: dark\n",
			wantErr: false,
			check: func(t *testing.T, p *ProjectProfile) {
				if p.Theme != "dark" {
					t.Errorf("expected theme 'dark', got %q", p.Theme)
				}
				if p.Environment != "" {
					t.Errorf("expected empty environment, got %q", p.Environment)
				}
				if p.UsesEnvironment() {
					t.Error("expected UsesEnvironment() to be false")
				}
				if !p.UsesTheme() {
					t.Error("expected UsesTheme() to be true")
				}
			},
		},
		{
			name: "both environment and theme",
			content: `environment: production
theme: dark
`,
			wantErr: true,
		},
		{
			name:    "neither environment nor theme",
			content: "other: value\n",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			content: "invalid: yaml: content:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			profilePath := filepath.Join(tmpDir, ProfileFileName)
			if err := os.WriteFile(profilePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test profile: %v", err)
			}

			profile, err := LoadProjectProfile(profilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadProjectProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, profile)
			}
		})
	}
}

func TestLoadProjectProfile_FileNotFound(t *testing.T) {
	_, err := LoadProjectProfile("/nonexistent/path/.terminal-profile")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadProjectProfile_StoresPath(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, ProfileFileName)
	if err := os.WriteFile(profilePath, []byte("environment: test\n"), 0644); err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	profile, err := LoadProjectProfile(profilePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Path != profilePath {
		t.Errorf("expected Path to be %q, got %q", profilePath, profile.Path)
	}
}

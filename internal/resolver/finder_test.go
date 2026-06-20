package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mulvad/ttm/internal/config"
)

// MockFileSystem is a mock implementation of FileSystem for testing.
type MockFileSystem struct {
	files map[string]bool
}

func (m *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if m.files[path] {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

func TestFinder_FindProfile_MockFS(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]bool
		startDir string
		want     string
	}{
		{
			name: "profile in current directory",
			files: map[string]bool{
				"/project/.terminal-profile": true,
			},
			startDir: "/project",
			want:     "/project/.terminal-profile",
		},
		{
			name: "profile in parent directory",
			files: map[string]bool{
				"/project/.terminal-profile": true,
			},
			startDir: "/project/subdir",
			want:     "/project/.terminal-profile",
		},
		{
			name: "profile in grandparent directory",
			files: map[string]bool{
				"/project/.terminal-profile": true,
			},
			startDir: "/project/subdir/nested",
			want:     "/project/.terminal-profile",
		},
		{
			name: "closest profile wins",
			files: map[string]bool{
				"/project/.terminal-profile":        true,
				"/project/subdir/.terminal-profile": true,
			},
			startDir: "/project/subdir",
			want:     "/project/subdir/.terminal-profile",
		},
		{
			name:     "no profile found",
			files:    map[string]bool{},
			startDir: "/project/subdir",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder(&MockFileSystem{files: tt.files})
			got, err := finder.FindProfile(tt.startDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("FindProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFinder_FindProfile_RealFS(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	// Create profile at middle level
	profilePath := filepath.Join(tmpDir, "a", config.ProfileFileName)
	if err := os.WriteFile(profilePath, []byte("environment: test\n"), 0644); err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	finder := NewFinder(nil) // Uses real filesystem

	// Find from deepest directory
	found, err := finder.FindProfile(subDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != profilePath {
		t.Errorf("FindProfile() = %q, want %q", found, profilePath)
	}

	// Find from directory containing profile
	found, err = finder.FindProfile(filepath.Join(tmpDir, "a"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != profilePath {
		t.Errorf("FindProfile() = %q, want %q", found, profilePath)
	}
}

func TestFinder_FindAndLoadProfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create profile
	profilePath := filepath.Join(tmpDir, config.ProfileFileName)
	if err := os.WriteFile(profilePath, []byte("environment: production\n"), 0644); err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	finder := NewFinder(nil)
	profile, err := finder.FindAndLoadProfile(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.Environment != "production" {
		t.Errorf("expected environment 'production', got %q", profile.Environment)
	}
	if profile.Path != profilePath {
		t.Errorf("expected path %q, got %q", profilePath, profile.Path)
	}
}

func TestFinder_FindAndLoadProfile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	finder := NewFinder(nil)
	profile, err := finder.FindAndLoadProfile(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile != nil {
		t.Errorf("expected nil profile, got %v", profile)
	}
}

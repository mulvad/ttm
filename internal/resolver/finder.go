// Package resolver handles finding and resolving terminal profiles.
package resolver

import (
	"os"
	"path/filepath"

	"github.com/mulvad/ttm/internal/config"
)

// FileSystem abstracts filesystem operations for testability.
type FileSystem interface {
	// Stat returns file info for the given path.
	Stat(path string) (os.FileInfo, error)
}

// OSFileSystem implements FileSystem using the real filesystem.
type OSFileSystem struct{}

// Stat returns file info for the given path.
func (OSFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Finder locates project profile files in the directory hierarchy.
type Finder struct {
	fs FileSystem
}

// NewFinder creates a new Finder with the given filesystem.
func NewFinder(fs FileSystem) *Finder {
	if fs == nil {
		fs = OSFileSystem{}
	}
	return &Finder{fs: fs}
}

// FindProfile searches for a .terminal-profile file starting from the given
// directory and walking up the directory tree until one is found or the
// filesystem root is reached.
func (f *Finder) FindProfile(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		profilePath := filepath.Join(dir, config.ProfileFileName)
		if _, err := f.fs.Stat(profilePath); err == nil {
			return profilePath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", nil
		}
		dir = parent
	}
}

// FindAndLoadProfile finds and loads a project profile starting from the given directory.
func (f *Finder) FindAndLoadProfile(startDir string) (*config.ProjectProfile, error) {
	profilePath, err := f.FindProfile(startDir)
	if err != nil {
		return nil, err
	}
	if profilePath == "" {
		return nil, nil
	}
	return config.LoadProjectProfile(profilePath)
}

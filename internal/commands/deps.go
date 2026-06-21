package commands

import (
	"context"

	"github.com/mulvad/ttm/internal/config"
	"github.com/mulvad/ttm/internal/terminal"
)

// Backend defines the terminal backend interface used by commands.
type Backend interface {
	Name() string
	Available() bool
	ApplyProfile(ctx context.Context, profile string) error
	CurrentProfile(ctx context.Context) (string, error)
	ListProfiles(ctx context.Context) ([]string, error)
	ExportProfile(ctx context.Context, name string) (*terminal.Profile, error)
	ExportAllProfiles(ctx context.Context) ([]*terminal.Profile, error)
	ImportProfile(ctx context.Context, profile *terminal.Profile) error
}

// ConfigLoader loads configuration files.
type ConfigLoader interface {
	LoadConfig(path string) (*config.Config, error)
	DefaultConfigPath() (string, error)
}

// ProfileFinder finds project profiles.
type ProfileFinder interface {
	FindAndLoadProfile(startDir string) (*config.ProjectProfile, error)
}

// Deps holds command dependencies for testing.
type Deps struct {
	Backend      Backend
	ConfigLoader ConfigLoader
	ProfileFinder ProfileFinder
	Getwd        func() (string, error)
}

// DefaultConfigLoader implements ConfigLoader using the config package.
type DefaultConfigLoader struct{}

func (DefaultConfigLoader) LoadConfig(path string) (*config.Config, error) {
	return config.LoadConfig(path)
}

func (DefaultConfigLoader) DefaultConfigPath() (string, error) {
	return config.DefaultConfigPath()
}

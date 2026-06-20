// Package terminal provides interfaces and implementations for terminal emulator backends.
package terminal

import "context"

// Backend defines the interface for terminal emulator integrations.
// Implementations handle applying themes/profiles to specific terminal emulators.
type Backend interface {
	// Name returns the human-readable name of the terminal backend.
	Name() string

	// Available returns true if this backend can be used in the current environment.
	Available() bool

	// ApplyProfile applies the given terminal profile/settings.
	ApplyProfile(ctx context.Context, profile string) error

	// CurrentProfile returns the name of the currently active profile.
	CurrentProfile(ctx context.Context) (string, error)

	// ListProfiles returns a list of available profile names.
	ListProfiles(ctx context.Context) ([]string, error)
}

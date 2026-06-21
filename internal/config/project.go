package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProjectProfile represents a project-specific terminal profile configuration.
// It is stored in a .terminal-profile file at the project root.
type ProjectProfile struct {
	// Environment specifies the semantic environment (e.g., "production", "staging").
	// Use "auto" to detect from the configured environment variable.
	// When used alone, the theme is resolved from the environment config.
	// Can be combined with Theme to get environment badge + explicit theme.
	Environment string `yaml:"environment,omitempty"`

	// Theme specifies the theme directly.
	// When used alone, bypasses environment resolution.
	// Can be combined with Environment to use explicit theme + environment badge.
	Theme string `yaml:"theme,omitempty"`

	// Path stores the file path where this profile was loaded from.
	Path string `yaml:"-"`
}

// UsesAutoEnvironment returns true if environment is set to "auto".
func (p *ProjectProfile) UsesAutoEnvironment() bool {
	return p.Environment == "auto"
}

// ProfileFileName is the name of the project profile file.
const ProfileFileName = ".terminal-profile"

// LoadProjectProfile loads a project profile from the given path.
func LoadProjectProfile(path string) (*ProjectProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read project profile: %w", err)
	}

	var profile ProjectProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		content := strings.TrimSpace(string(data))
		// Check if it looks like a plain string (no colon = no YAML key)
		if !strings.Contains(content, ":") {
			return nil, fmt.Errorf("invalid format in %s: found %q\n\nExpected format:\n  environment: <name>   # use semantic environment\n  theme: <name>         # use theme directly", path, content)
		}
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	profile.Path = path

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project profile %s: %w", path, err)
	}

	return &profile, nil
}

// Validate checks that the project profile is valid.
// It ensures at least one of Environment or Theme is set.
// Both can be set together (environment for badge, theme for styling).
func (p *ProjectProfile) Validate() error {
	hasEnv := p.Environment != ""
	hasTheme := p.Theme != ""

	if !hasEnv && !hasTheme {
		return errors.New("project profile must specify 'environment', 'theme', or both")
	}

	return nil
}

// UsesEnvironment returns true if this profile uses environment-based resolution.
func (p *ProjectProfile) UsesEnvironment() bool {
	return p.Environment != ""
}

// UsesTheme returns true if this profile uses direct theme specification.
func (p *ProjectProfile) UsesTheme() bool {
	return p.Theme != ""
}

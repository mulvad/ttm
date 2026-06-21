// Package config handles loading and validation of TTM configuration files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EnvironmentConfig defines the theme mapping for a semantic environment.
type EnvironmentConfig struct {
	Theme string `yaml:"theme"`
	Badge string `yaml:"badge,omitempty"` // Window title badge/prefix (e.g., "🔴 PROD")
}

// ThemeConfig defines a terminal profile mapping.
type ThemeConfig struct {
	Profile string `yaml:"profile"`
}

// Config represents the global TTM configuration stored at ~/.ttm/config.yaml.
type Config struct {
	// EnvironmentVariable is the name of the env var to check for auto-detection (e.g., "NODE_ENV")
	EnvironmentVariable string `yaml:"environment_variable,omitempty"`

	// DefaultProfile is the terminal profile to apply when no .terminal-profile is found
	DefaultProfile string `yaml:"default_profile,omitempty"`

	Environments map[string]EnvironmentConfig `yaml:"environments"`
	Themes       map[string]ThemeConfig       `yaml:"themes"`
}

// DefaultConfigPath returns the default path for the global config file.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ttm", "config.yaml"), nil
}

// LoadConfig loads and validates the global configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is valid.
// It ensures all environment theme references point to defined themes.
func (c *Config) Validate() error {
	if c.Environments == nil {
		c.Environments = make(map[string]EnvironmentConfig)
	}
	if c.Themes == nil {
		c.Themes = make(map[string]ThemeConfig)
	}

	var errs []error
	for envName, env := range c.Environments {
		if env.Theme == "" {
			errs = append(errs, fmt.Errorf("environment %q has no theme defined", envName))
			continue
		}
		if _, exists := c.Themes[env.Theme]; !exists {
			errs = append(errs, fmt.Errorf("environment %q references undefined theme %q", envName, env.Theme))
		}
	}

	for themeName, theme := range c.Themes {
		if theme.Profile == "" {
			errs = append(errs, fmt.Errorf("theme %q has no profile defined", themeName))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// GetThemeProfile resolves a theme name to its terminal profile.
func (c *Config) GetThemeProfile(themeName string) (string, error) {
	theme, exists := c.Themes[themeName]
	if !exists {
		return "", fmt.Errorf("theme %q not found", themeName)
	}
	return theme.Profile, nil
}

// GetEnvironmentTheme resolves an environment name to its theme.
func (c *Config) GetEnvironmentTheme(envName string) (string, error) {
	env, exists := c.Environments[envName]
	if !exists {
		return "", fmt.Errorf("environment %q not found", envName)
	}
	return env.Theme, nil
}

// GetEnvironment returns the full environment config.
func (c *Config) GetEnvironment(envName string) (*EnvironmentConfig, error) {
	env, exists := c.Environments[envName]
	if !exists {
		return nil, fmt.Errorf("environment %q not found", envName)
	}
	return &env, nil
}

// DetectEnvironment checks the configured environment variable and returns the matching environment name.
// Returns empty string if no env var is configured or the value doesn't match any environment.
func (c *Config) DetectEnvironment() string {
	if c.EnvironmentVariable == "" {
		return ""
	}

	value := os.Getenv(c.EnvironmentVariable)
	if value == "" {
		return ""
	}

	// Check if the env var value matches any environment name
	if _, exists := c.Environments[value]; exists {
		return value
	}

	return ""
}

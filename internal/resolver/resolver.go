package resolver

import (
	"fmt"

	"github.com/mulvad/ttm/internal/config"
)

// ResolutionStep represents a single step in the theme resolution chain.
type ResolutionStep struct {
	Type  string // "project", "environment", "theme", "profile"
	Key   string // The key/name at this step
	Value string // The resolved value
}

// Resolution contains the complete resolution chain from project to terminal profile.
type Resolution struct {
	Steps       []ResolutionStep
	Profile     string
	Badge       string // Optional badge/title from environment
	Environment string // Environment name if applicable
}

// Resolver handles theme resolution using the three-layer architecture.
type Resolver struct {
	config *config.Config
}

// NewResolver creates a new Resolver with the given global config.
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{config: cfg}
}

// Resolve performs full resolution from a project profile to a terminal profile.
// It returns the resolution chain and the final terminal profile name.
// Supports four modes:
// 1. Environment only: theme and badge from environment
// 2. Theme only: theme directly specified, no badge
// 3. Both: theme from project, badge from environment
// 4. Auto: environment detected from configured env var (e.g., NODE_ENV)
func (r *Resolver) Resolve(project *config.ProjectProfile) (*Resolution, error) {
	if project == nil {
		return nil, fmt.Errorf("no project profile provided")
	}

	res := &Resolution{
		Steps: make([]ResolutionStep, 0, 4),
	}

	// Handle auto environment detection
	envName := project.Environment
	if project.UsesAutoEnvironment() {
		detected := r.config.DetectEnvironment()
		if detected == "" {
			if r.config.EnvironmentVariable == "" {
				return nil, fmt.Errorf("environment set to 'auto' but no environment_variable configured in global config")
			}
			return nil, fmt.Errorf("environment set to 'auto' but %s is not set or doesn't match any environment", r.config.EnvironmentVariable)
		}
		envName = detected
	}

	hasEnv := envName != ""
	hasTheme := project.Theme != ""

	// Step 1: Record project profile
	var projectValue string
	if project.UsesAutoEnvironment() {
		if hasTheme {
			projectValue = fmt.Sprintf("environment: auto → %s, theme: %s", envName, project.Theme)
		} else {
			projectValue = fmt.Sprintf("environment: auto → %s", envName)
		}
	} else if hasEnv && hasTheme {
		projectValue = fmt.Sprintf("environment: %s, theme: %s", envName, project.Theme)
	} else if hasEnv {
		projectValue = fmt.Sprintf("environment: %s", envName)
	} else {
		projectValue = fmt.Sprintf("theme: %s", project.Theme)
	}
	res.Steps = append(res.Steps, ResolutionStep{
		Type:  "project",
		Key:   project.Path,
		Value: projectValue,
	})

	var themeName string
	var err error

	// Step 2: Resolve environment (for badge and optionally theme)
	if hasEnv {
		res.Environment = envName
		env, err := r.config.GetEnvironment(envName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve environment: %w", err)
		}
		res.Badge = env.Badge

		// If theme is explicitly set, use it; otherwise use environment's theme
		if hasTheme {
			themeName = project.Theme
			res.Steps = append(res.Steps, ResolutionStep{
				Type:  "environment",
				Key:   envName,
				Value: fmt.Sprintf("badge: %s", env.Badge),
			})
		} else {
			themeName = env.Theme
			res.Steps = append(res.Steps, ResolutionStep{
				Type:  "environment",
				Key:   envName,
				Value: fmt.Sprintf("theme: %s, badge: %s", themeName, env.Badge),
			})
		}
	} else {
		themeName = project.Theme
	}

	// Step 3: Record theme
	res.Steps = append(res.Steps, ResolutionStep{
		Type:  "theme",
		Key:   themeName,
		Value: themeName,
	})

	// Step 4: Resolve theme to terminal profile
	profile, err := r.config.GetThemeProfile(themeName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve theme: %w", err)
	}

	res.Steps = append(res.Steps, ResolutionStep{
		Type:  "profile",
		Key:   themeName,
		Value: profile,
	})

	res.Profile = profile
	return res, nil
}

// ResolveTheme resolves just the theme name from a project profile.
func (r *Resolver) ResolveTheme(project *config.ProjectProfile) (string, error) {
	if project == nil {
		return "", fmt.Errorf("no project profile provided")
	}

	if project.UsesTheme() {
		return project.Theme, nil
	}

	return r.config.GetEnvironmentTheme(project.Environment)
}

// ResolveProfile resolves directly to the terminal profile name.
func (r *Resolver) ResolveProfile(project *config.ProjectProfile) (string, error) {
	res, err := r.Resolve(project)
	if err != nil {
		return "", err
	}
	return res.Profile, nil
}

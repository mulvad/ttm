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
	Steps   []ResolutionStep
	Profile string
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
func (r *Resolver) Resolve(project *config.ProjectProfile) (*Resolution, error) {
	if project == nil {
		return nil, fmt.Errorf("no project profile provided")
	}

	res := &Resolution{
		Steps: make([]ResolutionStep, 0, 4),
	}

	// Step 1: Record project profile
	if project.UsesEnvironment() {
		res.Steps = append(res.Steps, ResolutionStep{
			Type:  "project",
			Key:   project.Path,
			Value: fmt.Sprintf("environment: %s", project.Environment),
		})
	} else {
		res.Steps = append(res.Steps, ResolutionStep{
			Type:  "project",
			Key:   project.Path,
			Value: fmt.Sprintf("theme: %s", project.Theme),
		})
	}

	var themeName string
	var err error

	// Step 2: Resolve environment to theme (if using environment)
	if project.UsesEnvironment() {
		themeName, err = r.config.GetEnvironmentTheme(project.Environment)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve environment: %w", err)
		}
		res.Steps = append(res.Steps, ResolutionStep{
			Type:  "environment",
			Key:   project.Environment,
			Value: themeName,
		})
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

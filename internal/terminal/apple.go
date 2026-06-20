package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ScriptRunner executes AppleScript commands.
type ScriptRunner interface {
	Run(ctx context.Context, script string) (string, error)
}

// OSAScriptRunner runs AppleScript using osascript.
type OSAScriptRunner struct{}

// Run executes an AppleScript and returns its output.
func (OSAScriptRunner) Run(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("osascript failed: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// AppleTerminal implements Backend for macOS Terminal.app.
type AppleTerminal struct {
	runner ScriptRunner
}

// NewAppleTerminal creates a new AppleTerminal backend.
func NewAppleTerminal() *AppleTerminal {
	return &AppleTerminal{runner: OSAScriptRunner{}}
}

// NewAppleTerminalWithRunner creates an AppleTerminal with a custom script runner (for testing).
func NewAppleTerminalWithRunner(runner ScriptRunner) *AppleTerminal {
	return &AppleTerminal{runner: runner}
}

// Name returns the backend name.
func (a *AppleTerminal) Name() string {
	return "Apple Terminal"
}

// Available returns true if running on macOS with Terminal.app.
func (a *AppleTerminal) Available() bool {
	// Check if we're on macOS
	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" {
		return true
	}
	// Also check if osascript is available (for testing on macOS outside Terminal)
	_, err := exec.LookPath("osascript")
	return err == nil
}

// escapeForAppleScript escapes a string for use in AppleScript.
func escapeForAppleScript(s string) string {
	// Escape backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// ApplyProfile applies the given terminal profile to the current window.
func (a *AppleTerminal) ApplyProfile(ctx context.Context, profile string) error {
	escapedProfile := escapeForAppleScript(profile)
	script := fmt.Sprintf(`
tell application "Terminal"
	set current settings of front window to settings set "%s"
end tell
`, escapedProfile)

	_, err := a.runner.Run(ctx, script)
	if err != nil {
		return fmt.Errorf("failed to apply profile %q: %w", profile, err)
	}
	return nil
}

// CurrentProfile returns the name of the current terminal profile.
func (a *AppleTerminal) CurrentProfile(ctx context.Context) (string, error) {
	script := `
tell application "Terminal"
	name of current settings of front window
end tell
`
	output, err := a.runner.Run(ctx, script)
	if err != nil {
		return "", fmt.Errorf("failed to get current profile: %w", err)
	}
	return output, nil
}

// ListProfiles returns all available terminal profiles.
func (a *AppleTerminal) ListProfiles(ctx context.Context) ([]string, error) {
	script := `
tell application "Terminal"
	set profileNames to {}
	repeat with s in settings sets
		set end of profileNames to name of s
	end repeat
	set AppleScript's text item delimiters to linefeed
	profileNames as text
end tell
`
	output, err := a.runner.Run(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	profiles := strings.Split(output, "\n")
	// Trim any whitespace from each profile name
	for i, p := range profiles {
		profiles[i] = strings.TrimSpace(p)
	}
	return profiles, nil
}

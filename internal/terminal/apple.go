package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
	// osascript adds a newline. Preserve whitespace inside values such as custom titles.
	return strings.TrimSuffix(string(output), "\n"), nil
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
// The profile's title preferences are preserved.
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

// SetWindowTitle sets the window/tab title via AppleScript.
// It also writes the badge to ~/.ttm/badge for shell integration scripts.
func (a *AppleTerminal) SetWindowTitle(ctx context.Context, title string) error {
	// Write badge to file for shell integration scripts that may want it
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	badgePath := home + "/.ttm/badge"
	if title == "" {
		_ = os.Remove(badgePath)
	} else {
		if err := os.WriteFile(badgePath, []byte(title), 0644); err != nil {
			return fmt.Errorf("failed to write badge file: %w", err)
		}
	}

	// Badges override the active tab's title, without changing the profile.
	// Clearing a badge restores the profile's custom title and visibility.
	script := `
tell application "Terminal"
	tell selected tab of front window
		set custom title to custom title of current settings
		set title displays custom title to title displays custom title of current settings
	end tell
end tell
`
	if title != "" {
		script = fmt.Sprintf(`
tell application "Terminal"
	tell selected tab of front window
		set custom title to "%s"
		set title displays custom title to true
	end tell
end tell
`, escapeForAppleScript(title))
	}

	_, err = a.runner.Run(ctx, script)
	if err != nil {
		return fmt.Errorf("failed to set window title: %w", err)
	}

	return nil
}

// GetWindowTitle returns the current window/tab title.
func (a *AppleTerminal) GetWindowTitle(ctx context.Context) (string, error) {
	script := `
tell application "Terminal"
	custom title of front window
end tell
`
	output, err := a.runner.Run(ctx, script)
	if err != nil {
		return "", fmt.Errorf("failed to get window title: %w", err)
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

// Color represents an RGB color with values from 0-65535 (Apple's format).
type Color struct {
	Red   int `yaml:"red"`
	Green int `yaml:"green"`
	Blue  int `yaml:"blue"`
}

// Profile represents an exportable terminal profile.
type Profile struct {
	Name            string          `yaml:"name"`
	BackgroundColor Color           `yaml:"background_color"`
	TextColor       Color           `yaml:"text_color"`
	BoldTextColor   Color           `yaml:"bold_text_color"`
	CursorColor     Color           `yaml:"cursor_color"`
	FontName        string          `yaml:"font_name"`
	FontSize        int             `yaml:"font_size"`
	Window          *WindowSettings `yaml:"window,omitempty"`
	Title           *TitleSettings  `yaml:"title,omitempty"`
}

// WindowSettings contains optional profile dimensions. Nil fields leave settings unchanged.
type WindowSettings struct {
	Rows    *int `yaml:"rows,omitempty"`
	Columns *int `yaml:"columns,omitempty"`
}

// TitleSettings controls the title components exposed by Terminal's scripting API.
// Pointers distinguish omitted values from explicit false or an empty custom title.
type TitleSettings struct {
	Custom          *string `yaml:"custom,omitempty"`
	ShowCustom      *bool   `yaml:"show_custom,omitempty"`
	ShowDeviceName  *bool   `yaml:"show_device_name,omitempty"`
	ShowShellPath   *bool   `yaml:"show_shell_path,omitempty"`
	ShowWindowSize  *bool   `yaml:"show_window_size,omitempty"`
	ShowProfileName *bool   `yaml:"show_profile_name,omitempty"`
}

// ExportProfile exports a single profile's settings.
func (a *AppleTerminal) ExportProfile(ctx context.Context, name string) (*Profile, error) {
	escapedName := escapeForAppleScript(name)
	script := fmt.Sprintf(`
tell application "Terminal"
	set theSettings to settings set "%s"
	set bgColor to background color of theSettings
	set fgColor to normal text color of theSettings
	set boldColor to bold text color of theSettings
	set curColor to cursor color of theSettings
	set theFont to font name of theSettings
	set theSize to font size of theSettings

	set theOutput to ""
	set theOutput to theOutput & "bg:" & (item 1 of bgColor) & "," & (item 2 of bgColor) & "," & (item 3 of bgColor) & linefeed
	set theOutput to theOutput & "fg:" & (item 1 of fgColor) & "," & (item 2 of fgColor) & "," & (item 3 of fgColor) & linefeed
	set theOutput to theOutput & "bold:" & (item 1 of boldColor) & "," & (item 2 of boldColor) & "," & (item 3 of boldColor) & linefeed
	set theOutput to theOutput & "cur:" & (item 1 of curColor) & "," & (item 2 of curColor) & "," & (item 3 of curColor) & linefeed
	set theOutput to theOutput & "font:" & theFont & linefeed
	set theOutput to theOutput & "size:" & theSize & linefeed
	set theOutput to theOutput & "rows:" & (number of rows of theSettings) & linefeed
	set theOutput to theOutput & "columns:" & (number of columns of theSettings) & linefeed
	set theOutput to theOutput & "show_custom:" & (title displays custom title of theSettings) & linefeed
	set theOutput to theOutput & "show_device_name:" & (title displays device name of theSettings) & linefeed
	set theOutput to theOutput & "show_shell_path:" & (title displays shell path of theSettings) & linefeed
	set theOutput to theOutput & "show_window_size:" & (title displays window size of theSettings) & linefeed
	set theOutput to theOutput & "show_profile_name:" & (title displays settings name of theSettings) & linefeed
	-- Custom title is last: it can contain colons, whitespace, and newlines.
	set theOutput to theOutput & "custom:" & (custom title of theSettings)
	theOutput
end tell
`, escapedName)

	output, err := a.runner.Run(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to export profile %q: %w", name, err)
	}

	profile := &Profile{Name: name}
	// Separate the final, free-form field before parsing the line-based values.
	if before, custom, ok := strings.Cut(output, "\ncustom:"); ok {
		profile.Title = &TitleSettings{Custom: &custom}
		output = before
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		switch key {
		case "bg":
			profile.BackgroundColor = parseColor(value)
		case "fg":
			profile.TextColor = parseColor(value)
		case "bold":
			profile.BoldTextColor = parseColor(value)
		case "cur":
			profile.CursorColor = parseColor(value)
		case "font":
			profile.FontName = value
		case "size":
			if size, err := strconv.Atoi(value); err == nil {
				profile.FontSize = size
			}
		case "rows", "columns":
			n, err := strconv.Atoi(value)
			if err != nil || n <= 0 {
				return nil, fmt.Errorf("invalid %s in exported profile %q: %q", key, name, value)
			}
			if profile.Window == nil {
				profile.Window = &WindowSettings{}
			}
			if key == "rows" {
				profile.Window.Rows = &n
			} else {
				profile.Window.Columns = &n
			}
		case "show_custom", "show_device_name", "show_shell_path", "show_window_size", "show_profile_name":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("invalid %s in exported profile %q: %q", key, name, value)
			}
			if profile.Title == nil {
				profile.Title = &TitleSettings{}
			}
			switch key {
			case "show_custom":
				profile.Title.ShowCustom = &b
			case "show_device_name":
				profile.Title.ShowDeviceName = &b
			case "show_shell_path":
				profile.Title.ShowShellPath = &b
			case "show_window_size":
				profile.Title.ShowWindowSize = &b
			case "show_profile_name":
				profile.Title.ShowProfileName = &b
			}
		}
	}

	return profile, nil
}

// parseColor parses a "r,g,b" string into a Color.
func parseColor(s string) Color {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return Color{}
	}
	r, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	g, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	b, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	return Color{Red: r, Green: g, Blue: b}
}

// ExportAllProfiles exports all terminal profiles.
func (a *AppleTerminal) ExportAllProfiles(ctx context.Context) ([]*Profile, error) {
	names, err := a.ListProfiles(ctx)
	if err != nil {
		return nil, err
	}

	profiles := make([]*Profile, 0, len(names))
	for _, name := range names {
		profile, err := a.ExportProfile(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to export profile %q: %w", name, err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// ImportProfile imports a profile, creating or updating it.
func (a *AppleTerminal) ImportProfile(ctx context.Context, profile *Profile) error {
	if profile == nil {
		return fmt.Errorf("profile must not be null")
	}
	extra, err := profileSettingsScript(profile)
	if err != nil {
		return err
	}
	escapedName := escapeForAppleScript(profile.Name)
	escapedFont := escapeForAppleScript(profile.FontName)

	// First, try to get existing profile or create new one
	script := fmt.Sprintf(`
tell application "Terminal"
	try
		set s to settings set "%s"
	on error
		-- Create a fresh profile when the name does not exist
		set s to make new settings set with properties {name:"%s"}
	end try

	set background color of s to {%d, %d, %d}
	set normal text color of s to {%d, %d, %d}
	set bold text color of s to {%d, %d, %d}
	set cursor color of s to {%d, %d, %d}
	set font name of s to "%s"
	set font size of s to %d
%s
end tell
`,
		escapedName, escapedName,
		profile.BackgroundColor.Red, profile.BackgroundColor.Green, profile.BackgroundColor.Blue,
		profile.TextColor.Red, profile.TextColor.Green, profile.TextColor.Blue,
		profile.BoldTextColor.Red, profile.BoldTextColor.Green, profile.BoldTextColor.Blue,
		profile.CursorColor.Red, profile.CursorColor.Green, profile.CursorColor.Blue,
		escapedFont, profile.FontSize, extra,
	)

	_, err = a.runner.Run(ctx, script)
	if err != nil {
		return fmt.Errorf("failed to import profile %q: %w", profile.Name, err)
	}

	return nil
}

// profileSettingsScript validates and emits only explicitly supplied optional settings.
func profileSettingsScript(profile *Profile) (string, error) {
	var script strings.Builder
	if w := profile.Window; w != nil {
		for _, field := range []struct {
			name  string
			value *int
		}{{"rows", w.Rows}, {"columns", w.Columns}} {
			if field.value == nil {
				continue
			}
			if *field.value <= 0 {
				return "", fmt.Errorf("window.%s must be positive", field.name)
			}
			fmt.Fprintf(&script, "\tset number of %s of s to %d\n", field.name, *field.value)
		}
	}
	if t := profile.Title; t != nil {
		if t.Custom != nil {
			fmt.Fprintf(&script, "\tset custom title of s to \"%s\"\n", escapeForAppleScript(*t.Custom))
		}
		for _, field := range []struct {
			name  string
			value *bool
		}{
			{"custom title", t.ShowCustom}, {"device name", t.ShowDeviceName},
			{"shell path", t.ShowShellPath}, {"window size", t.ShowWindowSize}, {"settings name", t.ShowProfileName},
		} {
			if field.value != nil {
				fmt.Fprintf(&script, "\tset title displays %s of s to %t\n", field.name, *field.value)
			}
		}
	}
	return script.String(), nil
}

package terminal

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// MockScriptRunner is a mock implementation of ScriptRunner for testing.
type MockScriptRunner struct {
	scripts []string
	outputs map[string]string
	errors  map[string]error
}

func NewMockScriptRunner() *MockScriptRunner {
	return &MockScriptRunner{
		outputs: make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (m *MockScriptRunner) Run(ctx context.Context, script string) (string, error) {
	m.scripts = append(m.scripts, script)

	// Check for exact matches first
	if err, ok := m.errors[script]; ok {
		return "", err
	}
	if output, ok := m.outputs[script]; ok {
		return output, nil
	}

	// Check for partial matches (script contains key)
	for key, err := range m.errors {
		if strings.Contains(script, key) {
			return "", err
		}
	}
	for key, output := range m.outputs {
		if strings.Contains(script, key) {
			return output, nil
		}
	}

	return "", nil
}

func (m *MockScriptRunner) SetOutput(scriptContains, output string) {
	m.outputs[scriptContains] = output
}

func (m *MockScriptRunner) SetError(scriptContains string, err error) {
	m.errors[scriptContains] = err
}

func (m *MockScriptRunner) LastScript() string {
	if len(m.scripts) == 0 {
		return ""
	}
	return m.scripts[len(m.scripts)-1]
}

func TestAppleTerminal_Name(t *testing.T) {
	terminal := NewAppleTerminal()
	if name := terminal.Name(); name != "Apple Terminal" {
		t.Errorf("Name() = %q, want 'Apple Terminal'", name)
	}
}

func TestAppleTerminal_ApplyProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		wantErr     bool
		errContains string
		checkScript func(t *testing.T, script string)
	}{
		{
			name:    "simple profile name",
			profile: "Basic",
			wantErr: false,
			checkScript: func(t *testing.T, script string) {
				if !strings.Contains(script, `settings set "Basic"`) {
					t.Errorf("script should contain profile name, got: %s", script)
				}
			},
		},
		{
			name:    "profile with spaces",
			profile: "Red Sands",
			wantErr: false,
			checkScript: func(t *testing.T, script string) {
				if !strings.Contains(script, `settings set "Red Sands"`) {
					t.Errorf("script should contain profile name with spaces, got: %s", script)
				}
			},
		},
		{
			name:    "profile with quotes",
			profile: `My "Special" Profile`,
			wantErr: false,
			checkScript: func(t *testing.T, script string) {
				if !strings.Contains(script, `settings set "My \"Special\" Profile"`) {
					t.Errorf("script should contain escaped quotes, got: %s", script)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockScriptRunner()
			terminal := NewAppleTerminalWithRunner(runner)

			err := terminal.ApplyProfile(context.Background(), tt.profile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkScript != nil {
				tt.checkScript(t, runner.LastScript())
			}
		})
	}
}

func TestAppleTerminal_ApplyProfile_Error(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetError("settings set", errors.New("profile not found"))
	terminal := NewAppleTerminalWithRunner(runner)

	err := terminal.ApplyProfile(context.Background(), "NonExistent")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to apply profile") {
		t.Errorf("error should mention failure, got: %v", err)
	}
}

func TestAppleTerminal_CurrentProfile(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetOutput("current settings", "Ocean")
	terminal := NewAppleTerminalWithRunner(runner)

	profile, err := terminal.CurrentProfile(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile != "Ocean" {
		t.Errorf("CurrentProfile() = %q, want 'Ocean'", profile)
	}
}

func TestAppleTerminal_CurrentProfile_Error(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetError("current settings", errors.New("no window"))
	terminal := NewAppleTerminalWithRunner(runner)

	_, err := terminal.CurrentProfile(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAppleTerminal_ListProfiles(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetOutput("settings sets", "Basic\nGrass\nHomebrew\nMan Page\nNovel\nOcean\nPro\nRed Sands\nSilver Aerogel\nSolid Colors")
	terminal := NewAppleTerminalWithRunner(runner)

	profiles, err := terminal.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Basic", "Grass", "Homebrew", "Man Page", "Novel", "Ocean", "Pro", "Red Sands", "Silver Aerogel", "Solid Colors"}
	if len(profiles) != len(expected) {
		t.Errorf("ListProfiles() returned %d profiles, want %d", len(profiles), len(expected))
	}

	for i, p := range profiles {
		if p != expected[i] {
			t.Errorf("profile[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestAppleTerminal_ListProfiles_Empty(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetOutput("settings sets", "")
	terminal := NewAppleTerminalWithRunner(runner)

	profiles, err := terminal.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected empty list, got %v", profiles)
	}
}

func TestEscapeForAppleScript(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Basic", "Basic"},
		{"Red Sands", "Red Sands"},
		{`My "Profile"`, `My \"Profile\"`},
		{`Path\To\Profile`, `Path\\To\\Profile`},
		{`Both "quotes" and \backslashes\`, `Both \"quotes\" and \\backslashes\\`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeForAppleScript(tt.input)
			if got != tt.want {
				t.Errorf("escapeForAppleScript(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input string
		want  Color
	}{
		{"65535,0,0", Color{Red: 65535, Green: 0, Blue: 0}},
		{"0, 65535, 0", Color{Red: 0, Green: 65535, Blue: 0}},
		{"10000,20000,30000", Color{Red: 10000, Green: 20000, Blue: 30000}},
		{"invalid", Color{}},
		{"", Color{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseColor(tt.input)
			if got != tt.want {
				t.Errorf("parseColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAppleTerminal_ExportProfile(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetOutput("settings set", `bg:5866,5866,5866
fg:65535,65535,65535
bold:65535,65535,65535
cur:35700,35700,35700
font:Menlo-Regular
size:12`)
	terminal := NewAppleTerminalWithRunner(runner)

	profile, err := terminal.ExportProfile(context.Background(), "Basic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.Name != "Basic" {
		t.Errorf("Name = %q, want 'Basic'", profile.Name)
	}
	if profile.BackgroundColor.Red != 5866 {
		t.Errorf("BackgroundColor.Red = %d, want 5866", profile.BackgroundColor.Red)
	}
	if profile.TextColor.Red != 65535 {
		t.Errorf("TextColor.Red = %d, want 65535", profile.TextColor.Red)
	}
	if profile.FontName != "Menlo-Regular" {
		t.Errorf("FontName = %q, want 'Menlo-Regular'", profile.FontName)
	}
	if profile.FontSize != 12 {
		t.Errorf("FontSize = %d, want 12", profile.FontSize)
	}
}

func TestAppleTerminal_ExportProfile_Error(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetError("settings set", errors.New("profile not found"))
	terminal := NewAppleTerminalWithRunner(runner)

	_, err := terminal.ExportProfile(context.Background(), "NonExistent")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAppleTerminal_ImportProfile(t *testing.T) {
	runner := NewMockScriptRunner()
	terminal := NewAppleTerminalWithRunner(runner)

	profile := &Profile{
		Name:            "TestProfile",
		BackgroundColor: Color{Red: 10000, Green: 20000, Blue: 30000},
		TextColor:       Color{Red: 65535, Green: 65535, Blue: 65535},
		BoldTextColor:   Color{Red: 65535, Green: 65535, Blue: 0},
		CursorColor:     Color{Red: 65535, Green: 0, Blue: 0},
		FontName:        "Menlo-Regular",
		FontSize:        14,
	}

	err := terminal.ImportProfile(context.Background(), profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	script := runner.LastScript()
	if !strings.Contains(script, "TestProfile") {
		t.Errorf("script should contain profile name, got: %s", script)
	}
	if !strings.Contains(script, "10000, 20000, 30000") {
		t.Errorf("script should contain background color, got: %s", script)
	}
	if !strings.Contains(script, "Menlo-Regular") {
		t.Errorf("script should contain font name, got: %s", script)
	}
}

func TestAppleTerminal_ImportProfile_Error(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetError("settings set", errors.New("failed"))
	terminal := NewAppleTerminalWithRunner(runner)

	profile := &Profile{
		Name:     "TestProfile",
		FontName: "Menlo-Regular",
		FontSize: 12,
	}

	err := terminal.ImportProfile(context.Background(), profile)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAppleTerminal_ExportAllProfiles(t *testing.T) {
	runner := NewMockScriptRunner()
	// First call lists profiles
	runner.SetOutput("settings sets", "Basic\nPro")
	// Subsequent calls export each profile
	runner.SetOutput("theSettings", `bg:0,0,0
fg:65535,65535,65535
bold:65535,65535,65535
cur:35700,35700,35700
font:Menlo-Regular
size:12`)
	terminal := NewAppleTerminalWithRunner(runner)

	profiles, err := terminal.ExportAllProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestAppleTerminal_ExportAllProfiles_ListError(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetError("settings sets", errors.New("failed to list"))
	terminal := NewAppleTerminalWithRunner(runner)

	_, err := terminal.ExportAllProfiles(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

package terminal

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestProfileSettingsRoundTrip(t *testing.T) {
	runner := NewMockScriptRunner()
	runner.SetOutput("theSettings", "size:12\nrows:40\ncolumns:120\nshow_custom:true\nshow_device_name:false\nshow_shell_path:true\nshow_window_size:false\nshow_profile_name:true\ncustom:  My \"title\": \\test\nshow_custom:false\n  ")
	backend := NewAppleTerminalWithRunner(runner)
	profile, err := backend.ExportProfile(context.Background(), "Test")
	if err != nil {
		t.Fatal(err)
	}
	data, err := yaml.Marshal(profile)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Profile
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(profile, &decoded) {
		t.Fatalf("YAML round trip changed profile: %s", data)
	}
	if profile.Window == nil || *profile.Window.Rows != 40 || *profile.Window.Columns != 120 {
		t.Fatalf("unexpected dimensions: %+v", profile.Window)
	}
	if profile.Title == nil || *profile.Title.Custom != "  My \"title\": \\test\nshow_custom:false\n  " {
		t.Fatalf("custom title did not survive export: %+v", profile.Title)
	}
	if err := backend.ImportProfile(context.Background(), &decoded); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"set number of rows of s to 40", "set number of columns of s to 120",
		"set title displays custom title of s to true", "set title displays device name of s to false",
		"set title displays shell path of s to true", "set title displays window size of s to false",
		"set title displays settings name of s to true",
		"set custom title of s to \"  My \\\"title\\\": \\\\test\nshow_custom:false\n  \"",
	} {
		if !strings.Contains(runner.LastScript(), want) {
			t.Errorf("import missing %q: %s", want, runner.LastScript())
		}
	}
}

func TestImportOptionalProfileSettings(t *testing.T) {
	for _, tc := range []struct {
		name, input, want, absent string
		invalid                   bool
	}{
		{name: "legacy", input: "name: Test", absent: " of s to false"},
		{name: "empty sections", input: "window: {}\ntitle: {}", absent: "set custom title"},
		{name: "explicit false", input: "title:\n  show_custom: false", want: "set title displays custom title of s to false", absent: "set title displays device name"},
		{name: "empty title", input: "title:\n  custom: ''", want: `set custom title of s to ""`, absent: "set title displays"},
		{name: "partial dimensions", input: "window:\n  columns: 100", want: "set number of columns of s to 100", absent: "set number of rows"},
		{name: "zero rows", input: "window:\n  rows: 0", invalid: true},
		{name: "negative columns", input: "window:\n  columns: -1", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var profile Profile
			if err := yaml.Unmarshal([]byte(tc.input), &profile); err != nil {
				t.Fatal(err)
			}
			runner := NewMockScriptRunner()
			err := NewAppleTerminalWithRunner(runner).ImportProfile(context.Background(), &profile)
			if tc.invalid {
				if err == nil || len(runner.scripts) != 0 {
					t.Fatalf("invalid dimensions reached Terminal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(runner.LastScript(), tc.want) || (tc.absent != "" && strings.Contains(runner.LastScript(), tc.absent)) {
				t.Fatalf("unexpected import script: %s", runner.LastScript())
			}
			if tc.name == "legacy" && strings.Contains(runner.LastScript(), "set number of") {
				t.Fatal("legacy import changed dimensions")
			}
		})
	}
}

func TestApplyPreservesProfileTitleSettings(t *testing.T) {
	runner := NewMockScriptRunner()
	if err := NewAppleTerminalWithRunner(runner).ApplyProfile(context.Background(), "Test"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(runner.LastScript(), "set title") || strings.Contains(runner.LastScript(), "set custom title") {
		t.Fatalf("apply overwrites title preferences: %s", runner.LastScript())
	}
}

func TestBadgeRestoresProfileTitle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(os.Getenv("HOME")+"/.ttm", 0755); err != nil {
		t.Fatal(err)
	}
	runner := NewMockScriptRunner()
	backend := NewAppleTerminalWithRunner(runner)
	if err := backend.SetWindowTitle(context.Background(), "PROD"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(runner.LastScript(), "tell selected tab of front window") || !strings.Contains(runner.LastScript(), "set title displays custom title to true") {
		t.Fatalf("badge not enabled for active tab: %s", runner.LastScript())
	}
	if err := backend.SetWindowTitle(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(runner.LastScript(), "set custom title to custom title of current settings") || !strings.Contains(runner.LastScript(), "set title displays custom title to title displays custom title of current settings") {
		t.Fatalf("profile title not restored: %s", runner.LastScript())
	}
}

// Opt-in: creates and deletes a temporary Terminal profile, without opening windows.
func TestProfileSettingsTerminalIntegration(t *testing.T) {
	if runtime.GOOS != "darwin" || os.Getenv("TTM_TERMINAL_INTEGRATION") != "1" {
		t.Skip("set TTM_TERMINAL_INTEGRATION=1 on macOS to test against Terminal.app")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	backend := NewAppleTerminal()
	name := "TTM integration " + time.Now().Format("20060102-150405.000000000")
	var profile Profile
	if err := yaml.Unmarshal([]byte(`name: `+name+`
font_name: Menlo-Regular
font_size: 12
window:
  rows: 43
  columns: 117
title:
  custom: '  Test "title": \\path  '
  show_custom: true
  show_device_name: false
  show_shell_path: true
  show_window_size: false
  show_profile_name: true
`), &profile); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, err := backend.runner.Run(cleanupCtx, `tell application "Terminal" to delete settings set "`+escapeForAppleScript(name)+`"`)
		if err != nil {
			t.Errorf("remove temporary profile %q: %v", name, err)
		}
	})
	if err := backend.ImportProfile(ctx, &profile); err != nil {
		t.Fatal(err)
	}
	got, err := backend.ExportProfile(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(profile.Window, got.Window) || !reflect.DeepEqual(profile.Title, got.Title) {
		t.Fatalf("Terminal round trip differs: want %+v / %+v, got %+v / %+v", profile.Window, profile.Title, got.Window, got.Title)
	}
	// A partial update preserves dimensions and unrelated title components.
	// Terminal's show_custom setter also clears the custom title when disabled.
	showCustom := false
	partial := *got
	partial.Window = nil
	partial.Title = &TitleSettings{ShowCustom: &showCustom}
	if err := backend.ImportProfile(ctx, &partial); err != nil {
		t.Fatal(err)
	}
	profile.Title.ShowCustom = &showCustom
	emptyTitle := ""
	profile.Title.Custom = &emptyTitle
	got, err = backend.ExportProfile(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(profile.Window, got.Window) || !reflect.DeepEqual(profile.Title, got.Title) {
		wantYAML, _ := yaml.Marshal(profile)
		gotYAML, _ := yaml.Marshal(got)
		t.Fatalf("partial import changed omitted settings or failed to disable the custom title:\nwant:\n%sgot:\n%s", wantYAML, gotYAML)
	}
	// Read Terminal's persisted preferences rather than only its scripting objects.
	for {
		data, err := exec.CommandContext(ctx, "/usr/bin/defaults", "export", "com.apple.Terminal", "-").Output()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "<key>"+name+"</key>") {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("temporary profile was not persisted in Terminal preferences")
		case <-time.After(200 * time.Millisecond):
		}
	}
}

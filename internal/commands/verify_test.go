package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVerify(t *testing.T) {
	var buf bytes.Buffer
	err := runVerify(&buf)
	if err != nil {
		t.Fatalf("runVerify() error = %v", err)
	}

	output := buf.String()

	// Should contain header
	if !strings.Contains(output, "=== TTM Title Verification ===") {
		t.Error("output should contain header")
	}

	// Should contain badge file section
	if !strings.Contains(output, "1. Badge file:") {
		t.Error("output should contain badge file section")
	}

	// Should contain DISABLE_AUTO_TITLE section
	if !strings.Contains(output, "2. DISABLE_AUTO_TITLE") {
		t.Error("output should contain DISABLE_AUTO_TITLE section")
	}

	// Should contain shell integration section
	if !strings.Contains(output, "3. Shell integration:") {
		t.Error("output should contain shell integration section")
	}

	// Should contain summary
	if !strings.Contains(output, "=== Summary ===") {
		t.Error("output should contain summary")
	}
}

func TestRunVerify_WithBadgeFile(t *testing.T) {
	// Create temp directory to use as home
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	// Create badge file
	ttmDir := filepath.Join(tmpDir, ".ttm")
	if err := os.MkdirAll(ttmDir, 0755); err != nil {
		t.Fatalf("failed to create .ttm dir: %v", err)
	}
	badgePath := filepath.Join(ttmDir, "badge")
	if err := os.WriteFile(badgePath, []byte("TEST_BADGE"), 0644); err != nil {
		t.Fatalf("failed to write badge: %v", err)
	}

	var buf bytes.Buffer
	err := runVerify(&buf)
	if err != nil {
		t.Fatalf("runVerify() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "TEST_BADGE") {
		t.Error("output should contain badge content")
	}
}

func TestRunVerify_DisableAutoTitle(t *testing.T) {
	// Set DISABLE_AUTO_TITLE env var
	origVal := os.Getenv("DISABLE_AUTO_TITLE")
	_ = os.Setenv("DISABLE_AUTO_TITLE", "true")
	defer func() { _ = os.Setenv("DISABLE_AUTO_TITLE", origVal) }()

	var buf bytes.Buffer
	err := runVerify(&buf)
	if err != nil {
		t.Fatalf("runVerify() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OK: DISABLE_AUTO_TITLE is exported") {
		t.Errorf("output should show DISABLE_AUTO_TITLE is exported, got: %s", output)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		want     bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "foo", false},
		{"foo", "", true},
		{"ttm_chpwd", "ttm_chpwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.haystack+"/"+tt.needle, func(t *testing.T) {
			got := containsString(tt.haystack, tt.needle)
			if got != tt.want {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.want)
			}
		})
	}
}

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewVerifyCmd creates the verify command.
func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify that the terminal title setup is working correctly",
		Long: `Verify checks the TTM title setup and reports any issues:

1. Checks if ~/.ttm/badge exists and shows its contents
2. Checks if DISABLE_AUTO_TITLE is set (required for oh-my-zsh users)
3. Provides suggestions to fix any issues found`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(os.Stdout)
		},
	}

	return cmd
}

func runVerify(w io.Writer) error {
	_, _ = fmt.Fprintln(w, "=== TTM Title Verification ===")
	_, _ = fmt.Fprintln(w)

	allOK := true

	// 1. Check badge file
	_, _ = fmt.Fprintln(w, "1. Badge file:")
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(w, "   Error: could not determine home directory: %v\n", err)
		allOK = false
	} else {
		badgePath := filepath.Join(home, ".ttm", "badge")
		if data, err := os.ReadFile(badgePath); err == nil {
			badge := string(data)
			if badge != "" {
				_, _ = fmt.Fprintf(w, "   OK: badge file exists with content: %s\n", badge)
			} else {
				_, _ = fmt.Fprintln(w, "   OK: badge file exists but is empty")
			}
		} else if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(w, "   Info: no badge file (this is normal if not in a project)")
		} else {
			_, _ = fmt.Fprintf(w, "   Error: could not read badge file: %v\n", err)
			allOK = false
		}
	}

	_, _ = fmt.Fprintln(w)

	// 2. Check DISABLE_AUTO_TITLE env var (should be exported by shell integration)
	_, _ = fmt.Fprintln(w, "2. DISABLE_AUTO_TITLE:")
	if os.Getenv("DISABLE_AUTO_TITLE") == "true" {
		_, _ = fmt.Fprintln(w, "   OK: DISABLE_AUTO_TITLE is exported and set to true")
	} else {
		_, _ = fmt.Fprintln(w, "   Warning: DISABLE_AUTO_TITLE is not exported")
		_, _ = fmt.Fprintln(w, "   Add this to your shell config (BEFORE sourcing oh-my-zsh if applicable):")
		_, _ = fmt.Fprintln(w, "   export DISABLE_AUTO_TITLE=\"true\"")
		allOK = false
	}

	_, _ = fmt.Fprintln(w)

	// 3. Check if shell integration is configured (ttm_chpwd for zsh, ttm_prompt_command for bash)
	_, _ = fmt.Fprintln(w, "3. Shell integration:")
	if home != "" {
		zshrcPath := filepath.Join(home, ".zshrc")
		if data, err := os.ReadFile(zshrcPath); err == nil {
			content := string(data)
			if containsString(content, "ttm_chpwd") {
				_, _ = fmt.Fprintln(w, "   OK: ttm_chpwd found in ~/.zshrc")
			} else {
				_, _ = fmt.Fprintln(w, "   Warning: ttm_chpwd not found in ~/.zshrc")
				_, _ = fmt.Fprintln(w, "   Add the shell integration from the README")
				allOK = false
			}
		} else if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(w, "   Info: ~/.zshrc does not exist (using bash?)")
			// Check bashrc
			bashrcPath := filepath.Join(home, ".bashrc")
			if data, err := os.ReadFile(bashrcPath); err == nil {
				content := string(data)
				if containsString(content, "ttm_prompt_command") {
					_, _ = fmt.Fprintln(w, "   OK: ttm_prompt_command found in ~/.bashrc")
				} else {
					_, _ = fmt.Fprintln(w, "   Warning: ttm_prompt_command not found in ~/.bashrc")
					_, _ = fmt.Fprintln(w, "   Add the shell integration from the README")
					allOK = false
				}
			}
		} else {
			_, _ = fmt.Fprintf(w, "   Error: could not read ~/.zshrc: %v\n", err)
		}
	}

	_, _ = fmt.Fprintln(w)

	// Summary
	_, _ = fmt.Fprintln(w, "=== Summary ===")
	if allOK {
		_, _ = fmt.Fprintln(w, "All checks passed! Your TTM title setup should be working.")
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "If titles still don't appear correctly:")
		_, _ = fmt.Fprintln(w, "  1. Open a NEW terminal window (existing windows may have stale state)")
		_, _ = fmt.Fprintln(w, "  2. Navigate to a project with a .terminal-profile file")
		_, _ = fmt.Fprintln(w, "  3. Run 'ttm apply' and check if the badge file is created")
	} else {
		_, _ = fmt.Fprintln(w, "Some issues were found. Please address the warnings above.")
	}

	return nil
}

func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

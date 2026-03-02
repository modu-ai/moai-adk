package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestInitCmd creates a fresh cobra.Command with the same flags as initCmd
// for isolated flag-validation tests that don't mutate the global initCmd.
func newTestInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init-test",
	}
	cmd.Flags().String("mode", "", "Development mode")
	cmd.Flags().String("git-mode", "", "Git workflow mode")
	cmd.Flags().String("git-provider", "", "Git provider")
	return cmd
}

// --- validateInitFlags tests ---

func TestValidateInitFlags_InvalidGitProvider(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"bitbucket", "bitbucket"},
		{"svn", "svn"},
		{"random", "random_provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestInitCmd()
			if err := cmd.Flags().Set("git-provider", tt.value); err != nil {
				t.Fatal(err)
			}
			err := validateInitFlags(cmd, nil)
			if err == nil {
				t.Errorf("expected error for git-provider=%q, got nil", tt.value)
			}
			if !strings.Contains(err.Error(), "invalid --git-provider") {
				t.Errorf("error should mention 'invalid --git-provider', got: %v", err)
			}
		})
	}
}

func TestValidateInitFlags_ValidGitProvider(t *testing.T) {
	for _, val := range []string{"github", "gitlab"} {
		t.Run(val, func(t *testing.T) {
			cmd := newTestInitCmd()
			if err := cmd.Flags().Set("git-provider", val); err != nil {
				t.Fatal(err)
			}
			if err := validateInitFlags(cmd, nil); err != nil {
				t.Errorf("unexpected error for git-provider=%q: %v", val, err)
			}
		})
	}
}

func TestValidateInitFlags_MultipleInvalidFlags_ReturnsFirstError(t *testing.T) {
	// When multiple flags are invalid, the function returns the first error it encounters.
	// mode is validated first in the function.
	cmd := newTestInitCmd()
	_ = cmd.Flags().Set("mode", "bad-mode")
	_ = cmd.Flags().Set("git-mode", "bad-git")

	err := validateInitFlags(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should report the mode error first since it's validated first.
	if !strings.Contains(err.Error(), "invalid --mode") {
		t.Errorf("expected mode error first, got: %v", err)
	}
}

func TestValidateInitFlags_AllEmpty(t *testing.T) {
	cmd := newTestInitCmd()
	// All flags are at their zero values (empty strings).
	if err := validateInitFlags(cmd, nil); err != nil {
		t.Errorf("all-empty flags should be valid, got: %v", err)
	}
}

func TestValidateInitFlags_AllValid(t *testing.T) {
	cmd := newTestInitCmd()
	_ = cmd.Flags().Set("mode", "tdd")
	_ = cmd.Flags().Set("git-mode", "team")
	_ = cmd.Flags().Set("git-provider", "github")

	if err := validateInitFlags(cmd, nil); err != nil {
		t.Errorf("all-valid flags should pass, got: %v", err)
	}
}

// --- runInitWizard error path tests ---

func TestRunInitWizard_ProjectNotInitialized(t *testing.T) {
	// Use a temp dir that has no .moai directory.
	tmpDir := t.TempDir()

	// runInitWizard calls os.Getwd() to find the project root,
	// so we need to chdir to our temp dir.
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	cmd := &cobra.Command{}
	cmd.SetOut(&strings.Builder{})

	err = runInitWizard(cmd, false)
	if err == nil {
		t.Fatal("expected error for uninitialized project, got nil")
	}
	if !strings.Contains(err.Error(), "project not initialized") {
		t.Errorf("expected 'project not initialized' in error, got: %v", err)
	}
}

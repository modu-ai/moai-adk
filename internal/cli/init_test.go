package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for init command behavior ---

func TestInitCmd_Exists(t *testing.T) {
	if initCmd == nil {
		t.Fatal("initCmd should not be nil")
	}
}

func TestInitCmd_Use(t *testing.T) {
	if initCmd.Use != "init [project-name]" {
		t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init [project-name]")
	}
}

func TestInitCmd_Short(t *testing.T) {
	if initCmd.Short == "" {
		t.Error("initCmd.Short should not be empty")
	}
}

func TestInitCmd_Long(t *testing.T) {
	if initCmd.Long == "" {
		t.Error("initCmd.Long should not be empty")
	}
}

func TestInitCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		// Use Name() which returns the command name without arguments
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("init should be registered as a subcommand of root")
	}
}

func TestInitCmd_HasFlags(t *testing.T) {
	flags := []string{"root", "name", "language", "framework", "username", "conv-lang", "mode", "non-interactive", "force"}
	for _, name := range flags {
		if initCmd.Flags().Lookup(name) == nil {
			t.Errorf("init command should have --%s flag", name)
		}
	}
}

func TestInitCmd_NonInteractiveExecution(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset flags to default before setting
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("set root flag: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("set non-interactive flag: %v", err)
	}
	if err := initCmd.Flags().Set("name", "test-project"); err != nil {
		t.Fatalf("set name flag: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("set language flag: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "ddd"); err != nil {
		t.Fatalf("set mode flag: %v", err)
	}

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}

	// Verify .moai/ was created
	moaiDir := filepath.Join(root, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created")
	}

	// Verify CLAUDE.md was created
	claudeMD := filepath.Join(root, "CLAUDE.md")
	if _, statErr := os.Stat(claudeMD); os.IsNotExist(statErr) {
		t.Error("expected CLAUDE.md to be created")
	}
}

// TestInitCmd_PositionalArgCreatesDirectory tests that positional argument creates a new directory
func TestInitCmd_PositionalArgCreatesDirectory(t *testing.T) {
	// Create a temp directory to work in
	workDir := t.TempDir()

	// Change to work directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset all flags to defaults
	_ = initCmd.Flags().Set("root", "")
	_ = initCmd.Flags().Set("non-interactive", "true")
	_ = initCmd.Flags().Set("name", "")
	_ = initCmd.Flags().Set("language", "Go")
	_ = initCmd.Flags().Set("mode", "ddd")

	// Run with positional argument
	err = initCmd.RunE(initCmd, []string{"my-new-project"})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify the directory was created
	projectDir := filepath.Join(workDir, "my-new-project")
	if _, statErr := os.Stat(projectDir); os.IsNotExist(statErr) {
		t.Error("expected my-new-project/ directory to be created")
	}

	// Verify .moai/ was created inside the new directory
	moaiDir := filepath.Join(projectDir, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created inside project folder")
	}

	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}
}

// TestInitCmd_DotArgUsesCurrentDirectory tests that "." argument uses current directory
func TestInitCmd_DotArgUsesCurrentDirectory(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	// Reset flags
	_ = initCmd.Flags().Set("root", "")
	_ = initCmd.Flags().Set("non-interactive", "true")
	_ = initCmd.Flags().Set("name", "dot-test")
	_ = initCmd.Flags().Set("language", "Go")
	_ = initCmd.Flags().Set("mode", "ddd")

	// Run with "." argument
	err = initCmd.RunE(initCmd, []string{"."})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify .moai/ was created in current directory (not in a new "." folder)
	moaiDir := filepath.Join(root, ".moai")
	if _, statErr := os.Stat(moaiDir); os.IsNotExist(statErr) {
		t.Error("expected .moai/ directory to be created in current directory")
	}

	// Verify initialization worked in current directory
	output := buf.String()
	if !strings.Contains(output, "MoAI project initialized") {
		t.Errorf("expected success message in output, got: %q", output)
	}
}

func TestGetStringFlag(t *testing.T) {
	// Flag exists but may have been set in previous test; just verify no panic
	_ = getStringFlag(initCmd, "name")

	// Non-existent flag returns empty
	if got := getStringFlag(initCmd, "nonexistent-flag-xyz"); got != "" {
		t.Errorf("getStringFlag for nonexistent flag = %q, want empty", got)
	}
}

func TestGetBoolFlag(t *testing.T) {
	// Non-existent flag returns false
	if got := getBoolFlag(initCmd, "nonexistent-flag-xyz"); got {
		t.Error("getBoolFlag for nonexistent flag should return false")
	}
}

// --- DDD PRESERVE: Characterization tests for flag validation ---

func TestValidateInitFlags_ValidMode(t *testing.T) {
	validModes := []string{"ddd", "tdd"}

	for _, mode := range validModes {
		t.Run(mode, func(t *testing.T) {
			// Set the mode flag
			if err := initCmd.Flags().Set("mode", mode); err != nil {
				t.Fatal(err)
			}

			// Validate should pass
			err := validateInitFlags(initCmd, []string{})
			if err != nil {
				t.Errorf("validateInitFlags with mode=%q should not error, got: %v", mode, err)
			}
		})
	}
}

func TestValidateInitFlags_InvalidMode(t *testing.T) {
	invalidModes := []string{"invalid", "test", "unknown", ""}

	for _, mode := range invalidModes {
		if mode == "" {
			continue // Empty mode is valid (uses default)
		}
		t.Run(mode, func(t *testing.T) {
			// Set the invalid mode flag
			if err := initCmd.Flags().Set("mode", mode); err != nil {
				t.Fatal(err)
			}

			// Validate should fail
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Errorf("validateInitFlags with mode=%q should error, got nil", mode)
			}

			// Error message should mention the invalid value
			if !strings.Contains(err.Error(), "invalid --mode") {
				t.Errorf("error should mention 'invalid --mode', got: %v", err)
			}
		})
	}
}

func TestValidateInitFlags_ValidGitMode(t *testing.T) {
	validGitModes := []string{"manual", "personal", "team"}

	for _, gitMode := range validGitModes {
		t.Run(gitMode, func(t *testing.T) {
			// Reset mode flag first (in case previous test left invalid value)
			if err := initCmd.Flags().Set("mode", ""); err != nil {
				t.Fatal(err)
			}

			// Set the git-mode flag
			if err := initCmd.Flags().Set("git-mode", gitMode); err != nil {
				t.Fatal(err)
			}

			// Validate should pass
			err := validateInitFlags(initCmd, []string{})
			if err != nil {
				t.Errorf("validateInitFlags with git-mode=%q should not error, got: %v", gitMode, err)
			}
		})
	}
}

func TestValidateInitFlags_InvalidGitMode(t *testing.T) {
	invalidGitModes := []string{"invalid", "auto", "unknown"}

	for _, gitMode := range invalidGitModes {
		t.Run(gitMode, func(t *testing.T) {
			// Reset mode flag first (in case previous test left invalid value)
			if err := initCmd.Flags().Set("mode", ""); err != nil {
				t.Fatal(err)
			}

			// Set the invalid git-mode flag
			if err := initCmd.Flags().Set("git-mode", gitMode); err != nil {
				t.Fatal(err)
			}

			// Validate should fail
			err := validateInitFlags(initCmd, []string{})
			if err == nil {
				t.Errorf("validateInitFlags with git-mode=%q should error, got nil", gitMode)
			}

			// Error message should mention the invalid value
			if !strings.Contains(err.Error(), "invalid --git-mode") {
				t.Errorf("error should mention 'invalid --git-mode', got: %v", err)
			}
		})
	}
}

func TestValidateInitFlags_EmptyFlags(t *testing.T) {
	// Reset all flags to empty
	if err := initCmd.Flags().Set("mode", ""); err != nil {
		t.Fatal(err)
	}
	if err := initCmd.Flags().Set("git-mode", ""); err != nil {
		t.Fatal(err)
	}

	// Validate should pass with empty flags (uses defaults)
	err := validateInitFlags(initCmd, []string{})
	if err != nil {
		t.Errorf("validateInitFlags with empty flags should not error, got: %v", err)
	}
}

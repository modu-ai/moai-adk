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
	if initCmd.Use != "init" {
		t.Errorf("initCmd.Use = %q, want %q", initCmd.Use, "init")
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
		if cmd.Use == "init" {
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
	if !strings.Contains(output, "MoAI project initialized successfully") {
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

func TestGetStringFlag(t *testing.T) {
	if got := getStringFlag(initCmd, "name"); got == "" {
		// Flag exists but may have been set in previous test; just verify no panic
	}

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

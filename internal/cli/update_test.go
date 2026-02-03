package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestUpdateCmd_Exists(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update" {
		t.Errorf("updateCmd.Use = %q, want %q", updateCmd.Use, "update")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("updateCmd.Short should not be empty")
	}
}

func TestUpdateCmd_HasFlags(t *testing.T) {
	flags := []string{"check", "force", "templates-only", "yes"}
	for _, name := range flags {
		if updateCmd.Flags().Lookup(name) == nil {
			t.Errorf("update command should have --%s flag", name)
		}
	}
}

func TestUpdateCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "update" {
			found = true
			break
		}
	}
	if !found {
		t.Error("update should be registered as a subcommand of root")
	}
}

func TestUpdateCmd_CheckOnly_NoDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Reset flags before test
	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Fatalf("update --check should not error with nil deps, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Current version") {
		t.Errorf("output should contain 'Current version', got %q", output)
	}
}

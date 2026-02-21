package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// TAG-004: moai init Mode Selection Tests

// TestInitCmd_HasInstallModeFlag tests that init command has --install-mode flag
func TestInitCmd_HasInstallModeFlag(t *testing.T) {
	if initCmd.Flags().Lookup("install-mode") == nil {
		t.Error("init command should have --install-mode flag")
	}
}

// TestValidateInstallModeFlag_Valid tests valid install-mode values
func TestValidateInstallModeFlag_Valid(t *testing.T) {
	validModes := []string{"global", "local", ""} // empty uses default

	for _, mode := range validModes {
		t.Run(mode, func(t *testing.T) {
			cmd := newTestInitCmd()
			cmd.Flags().String("install-mode", "", "Installation mode")
			if mode != "" {
				if err := cmd.Flags().Set("install-mode", mode); err != nil {
					t.Fatal(err)
				}
			}

			err := validateInitFlags(cmd, nil)
			if err != nil {
				t.Errorf("validateInitFlags with install-mode=%q should not error, got: %v", mode, err)
			}
		})
	}
}

// TestValidateInstallModeFlag_Invalid tests invalid install-mode values
func TestValidateInstallModeFlag_Invalid(t *testing.T) {
	invalidModes := []string{"auto", "hybrid", "none", "mixed"}

	for _, mode := range invalidModes {
		t.Run(mode, func(t *testing.T) {
			cmd := newTestInitCmd()
			cmd.Flags().String("install-mode", "", "Installation mode")
			if err := cmd.Flags().Set("install-mode", mode); err != nil {
				t.Fatal(err)
			}

			err := validateInitFlags(cmd, nil)
			if err == nil {
				t.Errorf("validateInitFlags with install-mode=%q should error", mode)
			}
			if !strings.Contains(err.Error(), "invalid --install-mode") {
				t.Errorf("error should mention 'invalid --install-mode', got: %v", err)
			}
		})
	}
}

// TestInitCmd_InstallModeSavedToSystemYAML tests that --install-mode is saved to system.yaml
func TestInitCmd_InstallModeSavedToSystemYAML(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset flags to default
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("set root flag: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("set non-interactive flag: %v", err)
	}
	if err := initCmd.Flags().Set("name", "test-mode-project"); err != nil {
		t.Fatalf("set name flag: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("set language flag: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "tdd"); err != nil {
		t.Fatalf("set mode flag: %v", err)
	}
	if err := initCmd.Flags().Set("install-mode", "global"); err != nil {
		t.Fatalf("set install-mode flag: %v", err)
	}

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify system.yaml contains installation_mode: global
	systemYAMLPath := filepath.Join(root, ".moai", "config", "sections", "system.yaml")
	data, err := os.ReadFile(systemYAMLPath)
	if err != nil {
		t.Fatalf("read system.yaml: %v", err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse system.yaml: %v", err)
	}

	system, ok := config["system"].(map[string]any)
	if !ok {
		t.Fatal("system.yaml should have 'system' key")
	}

	mode, ok := system["installation_mode"].(string)
	if !ok {
		t.Fatal("system.yaml should have 'installation_mode' key")
	}

	if mode != "global" {
		t.Errorf("installation_mode = %q, want %q", mode, "global")
	}
}

// TestInitCmd_DefaultInstallModeIsLocal tests that default install-mode is "local"
func TestInitCmd_DefaultInstallModeIsLocal(t *testing.T) {
	root := t.TempDir()

	buf := new(bytes.Buffer)
	initCmd.SetOut(buf)
	initCmd.SetErr(buf)

	// Reset flags to default (no --install-mode specified)
	if err := initCmd.Flags().Set("root", root); err != nil {
		t.Fatalf("set root flag: %v", err)
	}
	if err := initCmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("set non-interactive flag: %v", err)
	}
	if err := initCmd.Flags().Set("name", "test-default-mode"); err != nil {
		t.Fatalf("set name flag: %v", err)
	}
	// Explicitly reset install-mode to empty (test isolation from previous tests)
	if err := initCmd.Flags().Set("install-mode", ""); err != nil {
		t.Fatalf("reset install-mode flag: %v", err)
	}
	if err := initCmd.Flags().Set("language", "Go"); err != nil {
		t.Fatalf("set language flag: %v", err)
	}
	if err := initCmd.Flags().Set("mode", "tdd"); err != nil {
		t.Fatalf("set mode flag: %v", err)
	}
	// Don't set install-mode, should default to "local"

	err := initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatalf("init command RunE error = %v", err)
	}

	// Verify system.yaml contains installation_mode: local
	systemYAMLPath := filepath.Join(root, ".moai", "config", "sections", "system.yaml")
	data, err := os.ReadFile(systemYAMLPath)
	if err != nil {
		t.Fatalf("read system.yaml: %v", err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse system.yaml: %v", err)
	}

	system, ok := config["system"].(map[string]any)
	if !ok {
		t.Fatal("system.yaml should have 'system' key")
	}

	mode, ok := system["installation_mode"].(string)
	if !ok {
		t.Fatal("system.yaml should have 'installation_mode' key")
	}

	if mode != "local" {
		t.Errorf("default installation_mode = %q, want %q", mode, "local")
	}
}

// Helper to create a test init command with install-mode flag
func newTestInitCmdWithInstallMode() *cobra.Command {
	cmd := &cobra.Command{
		Use: "init-test",
	}
	cmd.Flags().String("mode", "", "Development mode")
	cmd.Flags().String("git-mode", "", "Git workflow mode")
	cmd.Flags().String("git-provider", "", "Git provider")
	cmd.Flags().String("model-policy", "", "Agent model policy")
	cmd.Flags().String("conv-lang", "", "Conversation language")
	cmd.Flags().String("install-mode", "", "Installation mode")
	return cmd
}

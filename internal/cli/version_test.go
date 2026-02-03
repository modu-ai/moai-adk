package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for version command behavior ---

func TestVersionCmd_Exists(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}
}

func TestVersionCmd_Use(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %q, want %q", versionCmd.Use, "version")
	}
}

func TestVersionCmd_Short(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}

func TestVersionCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version should be registered as a subcommand of root")
	}
}

func TestVersionCmd_OutputFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)

	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Fatalf("version command RunE error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "moai-adk") {
		t.Errorf("version output should contain 'moai-adk', got %q", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("version output should contain 'commit:', got %q", output)
	}
	if !strings.Contains(output, "built:") {
		t.Errorf("version output should contain 'built:', got %q", output)
	}
}

func TestVersionCmd_HasRunE(t *testing.T) {
	if versionCmd.RunE == nil {
		t.Error("versionCmd.RunE should not be nil")
	}
}

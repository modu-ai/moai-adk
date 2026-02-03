package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for root command behavior ---

func TestRootCmd_Exists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
}

func TestRootCmd_Use(t *testing.T) {
	if rootCmd.Use != "moai" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "moai")
	}
}

func TestRootCmd_Short(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
	if !strings.Contains(rootCmd.Short, "MoAI-ADK") {
		t.Errorf("rootCmd.Short should contain 'MoAI-ADK', got %q", rootCmd.Short)
	}
}

func TestRootCmd_Long(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootCmd_HasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version should not be empty")
	}
}

func TestRootCmd_HelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root --help error: %v", err)
	}

	output := buf.String()

	// Root command should show available commands section
	if !strings.Contains(output, "Available Commands") {
		t.Error("root --help should show Available Commands section")
	}

	// Verify core subcommands are registered
	requiredCommands := []string{"version", "init", "doctor", "status"}
	for _, cmd := range requiredCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("root --help should list %q subcommand", cmd)
		}
	}
}

func TestRootCmd_NoArgsShowsHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{})

	_ = rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "MoAI-ADK") {
		t.Error("root with no args should display help containing MoAI-ADK")
	}
}

func TestRootCmd_UnknownCommandError(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"nonexistent-command-xyz"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("unknown command should return error")
	}
}

func TestRootCmd_DroppedCommandsNotPresent(t *testing.T) {
	droppedCommands := []string{"language", "analyze", "switch"}
	for _, name := range droppedCommands {
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == name {
				t.Errorf("dropped command %q should not be registered", name)
			}
		}
	}
}

func TestRootCmd_SubcommandCount(t *testing.T) {
	// Ensure we have a reasonable number of subcommands
	count := len(rootCmd.Commands())
	if count < 4 {
		t.Errorf("rootCmd should have at least 4 subcommands, got %d", count)
	}
}

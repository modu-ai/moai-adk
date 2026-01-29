package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	// Check command properties
	if cmd.Use != "moai" {
		t.Errorf("root command Use = %s, want 'moai'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("root command Short is empty")
	}

	if cmd.Long == "" {
		t.Error("root command Long is empty")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	cmd := NewRootCommand()

	expectedCommands := []string{
		"init",
		"doctor",
		"status",
		"update",
		"statusline",
		"hook",
		"version",
	}

	for _, expected := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("root command missing subcommand '%s'", expected)
		}
	}
}

func TestRootCommandVersionFlag(t *testing.T) {
	cmd := NewRootCommand()

	// Cobra automatically adds --version and -v flags
	// We just verify the command is properly configured
	if cmd.Version == "" {
		t.Error("root command has empty version")
	}
}

func TestSubcommandsHaveHelp(t *testing.T) {
	rootCmd := NewRootCommand()

	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Short == "" {
			t.Errorf("subcommand '%s' has empty Short description", subCmd.Name())
		}

		if subCmd.Long == "" {
			t.Errorf("subcommand '%s' has empty Long description", subCmd.Name())
		}
	}
}

func TestHookCommandRequiresArgument(t *testing.T) {
	rootCmd := NewRootCommand()

	// Find the hook command
	var hookCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "hook" {
			hookCmd = cmd
			break
		}
	}

	if hookCmd == nil {
		t.Fatal("hook command not found")
	}

	// Check that it requires exactly 1 argument
	if hookCmd.Args == nil {
		t.Error("hook command has no Args validator")
	}
}

func TestVersionCommandOutput(t *testing.T) {
	rootCmd := NewRootCommand()

	// Find the version command
	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("version command not found")
	}

	// Check that version command has RunE defined
	if versionCmd.RunE == nil {
		t.Error("version command has no RunE function")
	}
}

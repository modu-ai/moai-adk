package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHookCmd_Exists(t *testing.T) {
	if hookCmd == nil {
		t.Fatal("hookCmd should not be nil")
	}
}

func TestHookCmd_Use(t *testing.T) {
	if hookCmd.Use != "hook" {
		t.Errorf("hookCmd.Use = %q, want %q", hookCmd.Use, "hook")
	}
}

func TestHookCmd_Short(t *testing.T) {
	if hookCmd.Short == "" {
		t.Error("hookCmd.Short should not be empty")
	}
}

func TestHookCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "hook" {
			found = true
			break
		}
	}
	if !found {
		t.Error("hook should be registered as a subcommand of root")
	}
}

func TestHookCmd_HasSubcommands(t *testing.T) {
	expected := []string{
		"session-start", "pre-tool", "post-tool", "session-end", "stop", "compact",
		"list", "agent", "pre-push",
		"post-tool-failure", "notification", "subagent-start", "user-prompt-submit",
		"permission-request", "teammate-idle", "task-completed", "subagent-stop",
	}
	for _, name := range expected {
		found := false
		for _, cmd := range hookCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("hook should have %q subcommand", name)
		}
	}
}

func TestHookCmd_SubcommandCount(t *testing.T) {
	count := len(hookCmd.Commands())
	if count != 19 {
		t.Errorf("hook should have 19 subcommands, got %d", count)
	}
}

func TestHookCmd_ListWithNilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	// Find the list subcommand
	var listCmd *bytes.Buffer
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "list" {
			listCmd = new(bytes.Buffer)
			cmd.SetOut(listCmd)
			cmd.SetErr(listCmd)
			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("hook list should not error, got: %v", err)
			}
			break
		}
	}

	if listCmd == nil {
		t.Fatal("list subcommand not found")
	}

	output := listCmd.String()
	if !strings.Contains(output, "not initialized") {
		t.Errorf("output should indicate not initialized, got %q", output)
	}
}

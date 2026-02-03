package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatuslineCmd_Exists(t *testing.T) {
	if statuslineCmd == nil {
		t.Fatal("statuslineCmd should not be nil")
	}
}

func TestStatuslineCmd_Use(t *testing.T) {
	if statuslineCmd.Use != "statusline" {
		t.Errorf("statuslineCmd.Use = %q, want %q", statuslineCmd.Use, "statusline")
	}
}

func TestStatuslineCmd_Hidden(t *testing.T) {
	if !statuslineCmd.Hidden {
		t.Error("statusline command should be hidden")
	}
}

func TestStatuslineCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "statusline" {
			found = true
			break
		}
	}
	if !found {
		t.Error("statusline should be registered as a subcommand of root")
	}
}

func TestStatuslineCmd_Execution_NoDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	statuslineCmd.SetOut(buf)
	statuslineCmd.SetErr(buf)

	err := statuslineCmd.RunE(statuslineCmd, []string{})
	if err != nil {
		t.Fatalf("statusline should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "moai") {
		t.Errorf("output should contain 'moai', got %q", output)
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatuslineCmd_Exists(t *testing.T) {
	if StatuslineCmd == nil {
		t.Fatal("StatuslineCmd should not be nil")
	}
}

func TestStatuslineCmd_Use(t *testing.T) {
	if StatuslineCmd.Use != "statusline" {
		t.Errorf("StatuslineCmd.Use = %q, want %q", StatuslineCmd.Use, "statusline")
	}
}

func TestStatuslineCmd_Hidden(t *testing.T) {
	if !StatuslineCmd.Hidden {
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
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)

	err := StatuslineCmd.RunE(StatuslineCmd, []string{})
	if err != nil {
		t.Fatalf("statusline should not error, got: %v", err)
	}

	output := buf.String()
	// Statusline should produce some output (git status, version, branch, or fallback)
	output = strings.TrimSpace(output)
	if output == "" {
		t.Errorf("output should not be empty")
	}
	// If output doesn't contain expected sections, it should at least be a valid fallback
	// The new independent collection always shows git status or version when available
}

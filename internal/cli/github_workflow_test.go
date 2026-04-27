// Package cli는 GitHub workflow 명령에 대한 테스트를 제공합니다.
// Package cli provides tests for GitHub workflow command.
package cli

import (
	"testing"
)

func TestNewWorkflowCmd(t *testing.T) {
	cmd := newWorkflowCmd()

	if cmd == nil {
		t.Fatal("newWorkflowCmd returned nil")
	}

	if cmd.Use != "workflow" {
		t.Errorf("expected Use 'workflow', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	expectedSubcmds := []string{"sync", "validate"}
	subcmds := make(map[string]bool)
	for _, c := range cmd.Commands() {
		subcmds[c.Name()] = true
	}

	for _, expected := range expectedSubcmds {
		if !subcmds[expected] {
			t.Errorf("missing subcommand %q", expected)
		}
	}
}

func TestNewWorkflowSyncCmd(t *testing.T) {
	cmd := newWorkflowSyncCmd()

	if cmd == nil {
		t.Fatal("newWorkflowSyncCmd returned nil")
	}

	if cmd.Use != "sync" {
		t.Errorf("expected Use 'sync', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewWorkflowValidateCmd(t *testing.T) {
	cmd := newWorkflowValidateCmd()

	if cmd == nil {
		t.Fatal("newWorkflowValidateCmd returned nil")
	}

	if cmd.Use != "validate" {
		t.Errorf("expected Use 'validate', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

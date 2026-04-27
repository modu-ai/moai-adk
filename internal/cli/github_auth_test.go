// Package cli는 GitHub auth 명령에 대한 테스트를 제공합니다.
// Package cli provides tests for GitHub auth command.
package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewAuthCmd(t *testing.T) {
	cmd := newAuthCmd()

	if cmd == nil {
		t.Fatal("newAuthCmd returned nil")
	}

	if cmd.Use != "auth" {
		t.Errorf("expected Use 'auth', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	expectedSubcmds := []string{"claude", "codex", "gemini", "glm"}
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

func TestNewAuthClaudeCmd(t *testing.T) {
	cmd := newAuthClaudeCmd()

	if cmd == nil {
		t.Fatal("newAuthClaudeCmd returned nil")
	}

	if cmd.Use != "claude <token>" {
		t.Errorf("expected Use 'claude <token>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestAuthCommandsStructure(t *testing.T) {
	tests := []struct {
		name       string
		cmdFactory func() *cobra.Command
		use        string
	}{
		{"claude", newAuthClaudeCmd, "claude <token>"},
		{"codex", newAuthCodexCmd, "codex <token>"},
		{"gemini", newAuthGeminiCmd, "gemini <token>"},
		{"glm", newAuthGLMCmd, "glm <token>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFactory()

			if cmd == nil {
				t.Fatal("command factory returned nil")
			}

			if cmd.Use != tt.use {
				t.Errorf("expected Use %q, got %q", tt.use, cmd.Use)
			}

			if cmd.Args == nil {
				t.Error("Args validator should be set")
			}
		})
	}
}

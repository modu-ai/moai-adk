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

// --- DDD PRESERVE: Characterization tests for statusline utility functions ---

func TestRenderSimpleFallback(t *testing.T) {
	result := renderSimpleFallback()

	if result == "" {
		t.Error("renderSimpleFallback should not return empty string")
	}

	if result != "moai" {
		t.Errorf("renderSimpleFallback() = %q, want %q", result, "moai")
	}
}

func TestRenderSimpleFallback_NotEmpty(t *testing.T) {
	result := renderSimpleFallback()

	if len(result) == 0 {
		t.Fatal("renderSimpleFallback should return non-empty string")
	}

	// Should be a simple string without special characters
	if strings.Contains(result, "\n") {
		t.Error("renderSimpleFallback should not contain newlines")
	}
}

func TestRenderSimpleFallback_ConsistentOutput(t *testing.T) {
	// Should return consistent output across multiple calls
	first := renderSimpleFallback()
	second := renderSimpleFallback()

	if first != second {
		t.Errorf("renderSimpleFallback should be consistent, got %q and %q", first, second)
	}
}

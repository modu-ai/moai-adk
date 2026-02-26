package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCCCmd_Exists(t *testing.T) {
	if ccCmd == nil {
		t.Fatal("ccCmd should not be nil")
	}
}

func TestCCCmd_Use(t *testing.T) {
	if ccCmd.Use != "cc" {
		t.Errorf("ccCmd.Use = %q, want %q", ccCmd.Use, "cc")
	}
}

func TestCCCmd_Short(t *testing.T) {
	if ccCmd.Short == "" {
		t.Error("ccCmd.Short should not be empty")
	}
}

func TestCCCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("cc should be registered as a subcommand of root")
	}
}

func TestCCCmd_Execution_NoDeps(t *testing.T) {
	// Use a temporary project root to prevent any mutation of real project files.
	// The project root finder is overridden via findProjectRootFn.
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{})
	if err != nil {
		t.Fatalf("cc command should not error with nil deps, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude") {
		t.Errorf("output should mention Claude, got %q", output)
	}
}

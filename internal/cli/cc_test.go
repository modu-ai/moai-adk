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

// --- Tests for project-level .env.glm removal (issue #384) ---

func TestRemoveProjectEnvGLM(t *testing.T) {
	projectRoot := t.TempDir()
	moaiDir := filepath.Join(projectRoot, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .env.glm file
	envPath := filepath.Join(moaiDir, ".env.glm")
	if err := os.WriteFile(envPath, []byte("export ANTHROPIC_AUTH_TOKEN=\"key\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Verify file exists before removal
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("test setup: .env.glm should exist")
	}

	err := removeProjectEnvGLM(projectRoot)
	if err != nil {
		t.Fatalf("removeProjectEnvGLM should succeed, got: %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		t.Error(".env.glm should be removed after removeProjectEnvGLM")
	}
}

func TestRemoveProjectEnvGLM_NoFileExists(t *testing.T) {
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Should succeed gracefully when file doesn't exist
	err := removeProjectEnvGLM(projectRoot)
	if err != nil {
		t.Fatalf("removeProjectEnvGLM should succeed when file doesn't exist, got: %v", err)
	}
}

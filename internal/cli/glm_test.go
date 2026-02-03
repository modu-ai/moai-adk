package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGLMCmd_Exists(t *testing.T) {
	if glmCmd == nil {
		t.Fatal("glmCmd should not be nil")
	}
}

func TestGLMCmd_Use(t *testing.T) {
	if !strings.HasPrefix(glmCmd.Use, "glm") {
		t.Errorf("glmCmd.Use should start with 'glm', got %q", glmCmd.Use)
	}
}

func TestGLMCmd_Short(t *testing.T) {
	if glmCmd.Short == "" {
		t.Error("glmCmd.Short should not be empty")
	}
}

func TestGLMCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "glm" {
			found = true
			break
		}
	}
	if !found {
		t.Error("glm should be registered as a subcommand of root")
	}
}

func TestGLMCmd_NoArgs(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GLM") {
		t.Errorf("output should mention GLM, got %q", output)
	}
	if !strings.Contains(output, "settings.local.json") {
		t.Errorf("output should mention settings.local.json, got %q", output)
	}
}

func TestGLMCmd_InjectsEnv(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}

	// Verify settings.local.json was created with env
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("should create settings.local.json: %v", err)
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("settings.local.json should be valid JSON: %v", err)
	}

	if settings.Env == nil {
		t.Fatal("env should be set")
	}
	if _, ok := settings.Env["ANTHROPIC_BASE_URL"]; !ok {
		t.Error("ANTHROPIC_BASE_URL should be set")
	}
	if _, ok := settings.Env["ANTHROPIC_AUTH_TOKEN"]; !ok {
		t.Error("ANTHROPIC_AUTH_TOKEN should be set")
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot should succeed: %v", err)
	}

	// Normalize paths for comparison
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualRoot, _ := filepath.EvalSymlinks(root)
	if actualRoot != expectedRoot {
		t.Errorf("findProjectRoot returned %q, expected %q", actualRoot, expectedRoot)
	}
}

func TestFindProjectRoot_NotInProject(t *testing.T) {
	// Create temp dir without .moai
	tmpDir := t.TempDir()

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err := findProjectRoot()
	if err == nil {
		t.Error("findProjectRoot should error when not in a MoAI project")
	}
}

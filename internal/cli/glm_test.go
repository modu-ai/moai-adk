package cli

import (
	"bytes"
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

func TestGLMCmd_AcceptsOptionalArg(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	// Without API key
	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm without args should not error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "GLM") {
		t.Errorf("output should mention GLM, got %q", buf.String())
	}

	// With API key
	buf.Reset()
	err = glmCmd.RunE(glmCmd, []string{"test-key"})
	if err != nil {
		t.Fatalf("glm with API key should not error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "API key") {
		t.Errorf("output should mention API key, got %q", buf.String())
	}
}

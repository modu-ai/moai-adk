package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// --- endsWith / endsWithAny tests ---

func TestEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		suffixes []string
		want     bool
	}{
		{
			name:     "single suffix match",
			s:        "ddd-pre-transformation",
			suffixes: []string{"-pre-transformation"},
			want:     true,
		},
		{
			name:     "single suffix no match",
			s:        "ddd-pre-transformation",
			suffixes: []string{"-validation"},
			want:     false,
		},
		{
			name:     "multiple suffixes first match",
			s:        "backend-validation",
			suffixes: []string{"-validation", "-pre-transformation"},
			want:     true,
		},
		{
			name:     "multiple suffixes second match",
			s:        "backend-pre-transformation",
			suffixes: []string{"-validation", "-pre-transformation"},
			want:     true,
		},
		{
			name:     "multiple suffixes no match",
			s:        "backend-completion",
			suffixes: []string{"-validation", "-pre-transformation"},
			want:     false,
		},
		{
			name:     "empty string",
			s:        "",
			suffixes: []string{"-validation"},
			want:     false,
		},
		{
			name:     "empty suffixes",
			s:        "backend-validation",
			suffixes: []string{},
			want:     false,
		},
		{
			name:     "suffix longer than string",
			s:        "abc",
			suffixes: []string{"longer-than-input"},
			want:     false,
		},
		{
			name:     "exact match",
			s:        "-completion",
			suffixes: []string{"-completion"},
			want:     true,
		},
		{
			name:     "empty suffix matches any string",
			s:        "anything",
			suffixes: []string{""},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := endsWith(tt.s, tt.suffixes...)
			if got != tt.want {
				t.Errorf("endsWith(%q, %v) = %v, want %v", tt.s, tt.suffixes, got, tt.want)
			}
		})
	}
}

func TestEndsWithAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		suffixes []string
		want     bool
	}{
		{
			name:     "validation suffix",
			s:        "backend-validation",
			suffixes: []string{"-validation", "-pre-transformation", "-pre-implementation"},
			want:     true,
		},
		{
			name:     "pre-implementation suffix",
			s:        "frontend-pre-implementation",
			suffixes: []string{"-validation", "-pre-transformation", "-pre-implementation"},
			want:     true,
		},
		{
			name:     "no match",
			s:        "backend-completion",
			suffixes: []string{"-validation", "-pre-transformation", "-pre-implementation"},
			want:     false,
		},
		{
			name:     "verification suffix",
			s:        "quality-verification",
			suffixes: []string{"-verification", "-post-transformation", "-post-implementation"},
			want:     true,
		},
		{
			name:     "completion suffix",
			s:        "agent-completion",
			suffixes: []string{"-completion"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := endsWithAny(tt.s, tt.suffixes...)
			if got != tt.want {
				t.Errorf("endsWithAny(%q, %v) = %v, want %v", tt.s, tt.suffixes, got, tt.want)
			}
		})
	}
}

// --- runAgentHook tests ---

func TestRunAgentHook_NilDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("runAgentHook should error with nil deps")
			}
			if !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("error should mention not initialized, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_NilProtocol(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookRegistry: &mockHookRegistry{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("runAgentHook should error with nil protocol")
			}
			if !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("error should mention not initialized, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_NilRegistry(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("runAgentHook should error with nil registry")
			}
			if !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("error should mention not initialized, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_ValidationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"backend-validation"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPreToolUse {
				t.Errorf("validation action should dispatch as PreToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_VerificationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"quality-verification"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPostToolUse {
				t.Errorf("verification action should dispatch as PostToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_CompletionAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"agent-completion"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventSubagentStop {
				t.Errorf("completion action should dispatch as SubagentStop, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_DefaultAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"unknown-action"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPreToolUse {
				t.Errorf("unknown action should default to PreToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_ReadInputError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return nil, io.ErrUnexpectedEOF
			},
		},
		HookRegistry: &mockHookRegistry{},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("should error on ReadInput failure")
			}
			if !strings.Contains(err.Error(), "read hook input") {
				t.Errorf("error should mention read hook input, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_DispatchError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, _ hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				return nil, io.ErrUnexpectedEOF
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("should error on dispatch failure")
			}
			if !strings.Contains(err.Error(), "dispatch agent hook") {
				t.Errorf("error should mention dispatch agent hook, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_WriteOutputError(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
			writeOutputFunc: func(_ io.Writer, _ *hook.HookOutput) error {
				return io.ErrShortWrite
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, _ hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"test-validation"})
			if err == nil {
				t.Error("should error on WriteOutput failure")
			}
			if !strings.Contains(err.Error(), "write hook output") {
				t.Errorf("error should mention write hook output, got %v", err)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_PreTransformationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"ddd-pre-transformation"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPreToolUse {
				t.Errorf("pre-transformation action should dispatch as PreToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_PostTransformationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"ddd-post-transformation"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPostToolUse {
				t.Errorf("post-transformation action should dispatch as PostToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_PreImplementationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"backend-pre-implementation"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPreToolUse {
				t.Errorf("pre-implementation action should dispatch as PreToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

func TestRunAgentHook_PostImplementationAction(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	var capturedEvent hook.EventType
	deps = &Dependencies{
		HookProtocol: &mockHookProtocol{
			readInputFunc: func(_ io.Reader) (*hook.HookInput, error) {
				return &hook.HookInput{SessionID: "test"}, nil
			},
		},
		HookRegistry: &mockHookRegistry{
			dispatchFunc: func(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
				capturedEvent = event
				return hook.NewAllowOutput(), nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "agent" {
			cmd.SetContext(context.Background())
			err := cmd.RunE(cmd, []string{"backend-post-implementation"})
			if err != nil {
				t.Fatalf("runAgentHook error: %v", err)
			}
			if capturedEvent != hook.EventPostToolUse {
				t.Errorf("post-implementation action should dispatch as PostToolUse, got %v", capturedEvent)
			}
			return
		}
	}
	t.Error("agent subcommand not found")
}

// --- resolveConventionName additional coverage ---

func TestResolveConventionName_EnvVariousValues(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    string
	}{
		{"conventional", "conventional", "conventional"},
		{"angular", "angular", "angular"},
		{"none", "none", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MOAI_GIT_CONVENTION", tt.envVal)
			got := resolveConventionName()
			if got != tt.want {
				t.Errorf("resolveConventionName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- isEnforceOnPushEnabled additional coverage ---

func TestIsEnforceOnPushEnabled_EnvVariousValues(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true string", "true", true},
		{"1 string", "1", true},
		{"false string", "false", false},
		{"0 string", "0", false},
		{"random string", "random", false},
		{"yes string", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MOAI_ENFORCE_ON_PUSH", tt.envVal)
			got := isEnforceOnPushEnabled()
			if got != tt.want {
				t.Errorf("isEnforceOnPushEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- runPrePush tests ---

func TestRunPrePush_DisabledByDefault(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	t.Setenv("MOAI_ENFORCE_ON_PUSH", "")

	buf := new(bytes.Buffer)
	prePushCmd.SetOut(buf)
	prePushCmd.SetErr(buf)

	err := prePushCmd.RunE(prePushCmd, []string{})
	if err != nil {
		t.Fatalf("runPrePush should not error when enforcement disabled, got: %v", err)
	}

	// When disabled, output should be empty (early return)
	output := buf.String()
	if output != "" {
		t.Errorf("runPrePush should produce no output when disabled, got %q", output)
	}
}

// --- PrintWelcomeMessage tests ---

func TestPrintWelcomeMessage_OutputFormat(t *testing.T) {
	output, err := captureStdout(func() {
		PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(output) == 0 {
		t.Error("PrintWelcomeMessage should produce output")
	}

	expectedStrings := []string{
		"Welcome",
		"MoAI-ADK",
		"Initialization",
		"Ctrl+C",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintWelcomeMessage output should contain %q, got:\n%s", expected, output)
		}
	}
}

func TestPrintWelcomeMessage_MultipleCallsConsistent(t *testing.T) {
	output1, err := captureStdout(func() {
		PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	output2, err := captureStdout(func() {
		PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	if output1 != output2 {
		t.Error("PrintWelcomeMessage should produce consistent output across calls")
	}
}

// --- removeGLMEnv tests ---

func TestRemoveGLMEnv_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	err := removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv should succeed when file does not exist, got: %v", err)
	}
}

func TestRemoveGLMEnv_WithGLMVars(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	settings := SettingsLocal{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN":          "test-token",
			"ANTHROPIC_BASE_URL":            "https://test.example.com",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "test-haiku",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "test-sonnet",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "test-opus",
			"CUSTOM_VAR":                     "keep-this",
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}

	// Read back and verify
	result, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var resultSettings SettingsLocal
	if err := json.Unmarshal(result, &resultSettings); err != nil {
		t.Fatal(err)
	}

	// GLM vars should be removed
	for _, key := range []string{
		"ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
	} {
		if _, exists := resultSettings.Env[key]; exists {
			t.Errorf("GLM env var %q should be removed", key)
		}
	}

	// Custom var should remain
	if resultSettings.Env["CUSTOM_VAR"] != "keep-this" {
		t.Errorf("CUSTOM_VAR should be preserved, got %q", resultSettings.Env["CUSTOM_VAR"])
	}
}

func TestRemoveGLMEnv_OnlyGLMVars(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	settings := SettingsLocal{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "test-token",
			"ANTHROPIC_BASE_URL":  "https://test.example.com",
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}

	// Read back and verify env is nil (all vars removed)
	result, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var resultSettings SettingsLocal
	if err := json.Unmarshal(result, &resultSettings); err != nil {
		t.Fatal(err)
	}

	if resultSettings.Env != nil {
		t.Errorf("Env should be nil when all GLM vars removed, got %v", resultSettings.Env)
	}
}

func TestRemoveGLMEnv_EmptyEnv(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	settings := SettingsLocal{
		Env: map[string]string{},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}
}

func TestRemoveGLMEnv_NoEnvKey(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	settings := SettingsLocal{}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}
}

func TestRemoveGLMEnv_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	if err := os.WriteFile(settingsPath, []byte("not valid json{"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := removeGLMEnv(settingsPath)
	if err == nil {
		t.Error("removeGLMEnv should error on invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse settings.local.json") {
		t.Errorf("error should mention parse, got %v", err)
	}
}

func TestRemoveGLMEnv_PreservesOtherFields(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.local.json")

	settings := SettingsLocal{
		EnabledMcpjsonServers: []string{"server1", "server2"},
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "test-token",
			"CUSTOM_VAR":          "keep",
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = removeGLMEnv(settingsPath)
	if err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}

	result, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var resultSettings SettingsLocal
	if err := json.Unmarshal(result, &resultSettings); err != nil {
		t.Fatal(err)
	}

	if len(resultSettings.EnabledMcpjsonServers) != 2 {
		t.Errorf("EnabledMcpjsonServers should be preserved, got %v", resultSettings.EnabledMcpjsonServers)
	}
}

// --- readStdinWithTimeout tests ---

func TestReadStdinWithTimeout_ReturnsSomething(t *testing.T) {
	// readStdinWithTimeout returns an io.Reader, either os.Stdin or an empty MultiReader.
	// We mainly verify it does not panic and returns a non-nil reader.
	reader := readStdinWithTimeout()
	if reader == nil {
		t.Error("readStdinWithTimeout should return a non-nil reader")
	}
}

// --- buildAutoUpdateFunc tests ---

func TestBuildAutoUpdateFunc_ReturnsNonNil(t *testing.T) {
	fn := buildAutoUpdateFunc()
	if fn == nil {
		t.Fatal("buildAutoUpdateFunc should return a non-nil function")
	}
}

func TestBuildAutoUpdateFunc_DevBuildSkipped(t *testing.T) {
	// The function should detect dev builds and skip the update.
	// Since we're running in a test environment, the version is likely "dev"
	// or contains "dirty"/"none", which should be skipped.
	fn := buildAutoUpdateFunc()
	result, err := fn(context.Background())
	if err != nil {
		t.Fatalf("buildAutoUpdateFunc should not error for dev build, got: %v", err)
	}
	if result == nil {
		t.Fatal("buildAutoUpdateFunc should return a non-nil result for dev build")
	}
	if result.Updated {
		t.Error("dev build should not be updated")
	}
}

// --- EnsureGit additional coverage ---

func TestEnsureGit_ValidGitRepo(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create minimal git config
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte("[core]\n\tbare = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &Dependencies{}
	err := d.EnsureGit(tmpDir)
	// This may succeed or fail depending on git library requirements.
	// We just want to exercise the code path.
	if err == nil {
		if d.Git == nil {
			t.Error("EnsureGit should set Git on success")
		}
		if d.GitBranch == nil {
			t.Error("EnsureGit should set GitBranch on success")
		}
		if d.GitWorktree == nil {
			t.Error("EnsureGit should set GitWorktree on success")
		}
	}
}

// --- EnsureUpdate additional coverage ---

func TestEnsureUpdate_LocalSource(t *testing.T) {
	t.Setenv("MOAI_UPDATE_SOURCE", "local")
	t.Setenv("MOAI_RELEASES_DIR", t.TempDir())

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err != nil {
		t.Fatalf("EnsureUpdate with local source error: %v", err)
	}
	if d.UpdateChecker == nil {
		t.Error("EnsureUpdate should set UpdateChecker for local source")
	}
	if d.UpdateOrch == nil {
		t.Error("EnsureUpdate should set UpdateOrch for local source")
	}
}

func TestEnsureUpdate_CustomURL(t *testing.T) {
	t.Setenv("MOAI_UPDATE_SOURCE", "")
	t.Setenv("MOAI_UPDATE_URL", "https://api.example.com/releases")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err != nil {
		t.Fatalf("EnsureUpdate with custom URL error: %v", err)
	}
	if d.UpdateChecker == nil {
		t.Error("EnsureUpdate should set UpdateChecker for custom URL")
	}
}

// --- Hook list with multiple handlers ---

func TestRunHookList_MultipleHandlers(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		HookRegistry: &mockHookRegistry{
			handlersFunc: func(event hook.EventType) []hook.Handler {
				if event == hook.EventSessionStart {
					return []hook.Handler{
						&mockHandler{eventType: hook.EventSessionStart},
						&mockHandler{eventType: hook.EventSessionStart},
					}
				}
				return nil
			},
		},
	}

	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runHookList error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "2 handlers") {
				t.Errorf("output should show plural handlers, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestSessionStartHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewSessionStartHandler(cfg)

	if got := h.EventType(); got != EventSessionStart {
		t.Errorf("EventType() = %q, want %q", got, EventSessionStart)
	}
}

func TestSessionStartHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cfg          *config.Config
		input        *HookInput
		wantDecision string
		wantDataKeys []string
	}{
		{
			name: "normal session initialization with project config",
			cfg: func() *config.Config {
				c := newTestConfig()
				c.Project = models.ProjectConfig{
					Name:     "moai-adk-go",
					Type:     models.ProjectTypeCLI,
					Language: "go",
				}
				return c
			}(),
			input: &HookInput{
				SessionID:     "sess-abc-123",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
				ProjectDir:    t.TempDir(),
			},
			wantDecision: DecisionAllow,
			wantDataKeys: []string{"project_name"},
		},
		{
			name: "session start with nil config returns allow",
			cfg:  nil,
			input: &HookInput{
				SessionID:     "sess-nil-cfg",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "session start with empty project config returns allow",
			cfg:  newTestConfig(),
			input: &HookInput{
				SessionID:     "sess-empty",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
			},
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: tt.cfg}
			h := NewSessionStartHandler(cfg)

			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// SessionStart does NOT use hookSpecificOutput per Claude Code protocol
			if got.HookSpecificOutput != nil {
				t.Errorf("HookSpecificOutput should be nil for SessionStart, got %+v", got.HookSpecificOutput)
			}

			if len(tt.wantDataKeys) > 0 && got.Data != nil {
				var data map[string]any
				if err := json.Unmarshal(got.Data, &data); err != nil {
					t.Fatalf("unmarshal data: %v", err)
				}
				for _, key := range tt.wantDataKeys {
					if _, ok := data[key]; !ok {
						t.Errorf("data missing key %q", key)
					}
				}
			}
		})
	}
}

func TestEnsureGLMCredentials(t *testing.T) {
	// Not parallel: subtests use t.Setenv which requires non-parallel parent

	t.Run("no settings file", func(t *testing.T) {
		t.Parallel()
		msg := ensureGLMCredentials(t.TempDir())
		if msg != "" {
			t.Errorf("expected empty msg for missing file, got %q", msg)
		}
	})

	t.Run("no GLM models in settings", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		_ = os.MkdirAll(claudeDir, 0o755)
		settings := `{"env":{"SOME_VAR":"value"}}`
		_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

		msg := ensureGLMCredentials(dir)
		if msg != "" {
			t.Errorf("expected empty msg for non-GLM settings, got %q", msg)
		}
	})

	t.Run("GLM models with AUTH_TOKEN already present", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		_ = os.MkdirAll(claudeDir, 0o755)
		settings := `{"env":{"ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-5.1","ANTHROPIC_AUTH_TOKEN":"existing-key"}}`
		_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

		msg := ensureGLMCredentials(dir)
		if msg != "" {
			t.Errorf("expected empty msg when AUTH_TOKEN exists, got %q", msg)
		}
	})

	t.Run("GLM models without AUTH_TOKEN auto-injects from env.glm", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		_ = os.MkdirAll(claudeDir, 0o755)
		settings := `{"env":{"ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-5.1","ANTHROPIC_DEFAULT_SONNET_MODEL":"glm-4.7"}}`
		_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

		// Create fake ~/.moai/.env.glm
		fakeHome := t.TempDir()
		moaiDir := filepath.Join(fakeHome, ".moai")
		_ = os.MkdirAll(moaiDir, 0o755)
		_ = os.WriteFile(filepath.Join(moaiDir, ".env.glm"), []byte("GLM_API_KEY=\"test-glm-key-123\"\n"), 0o600)
		t.Setenv("HOME", fakeHome)
		t.Setenv("USERPROFILE", fakeHome) // Windows: os.UserHomeDir reads USERPROFILE

		msg := ensureGLMCredentials(dir)
		if msg == "" {
			t.Error("expected non-empty msg when credentials auto-injected")
		}

		// Verify settings.local.json was updated
		data, err := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
		if err != nil {
			t.Fatalf("read updated settings: %v", err)
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("parse updated settings: %v", err)
		}

		var env map[string]string
		if err := json.Unmarshal(raw["env"], &env); err != nil {
			t.Fatalf("parse env: %v", err)
		}

		if env["ANTHROPIC_AUTH_TOKEN"] != "test-glm-key-123" {
			t.Errorf("AUTH_TOKEN = %q, want %q", env["ANTHROPIC_AUTH_TOKEN"], "test-glm-key-123")
		}
		if env["ANTHROPIC_BASE_URL"] != "https://api.z.ai/api/anthropic" {
			t.Errorf("BASE_URL = %q, want Z.AI endpoint", env["ANTHROPIC_BASE_URL"])
		}
		if env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] != "1" {
			t.Error("DISABLE_EXPERIMENTAL_BETAS should be set to 1")
		}
	})
}

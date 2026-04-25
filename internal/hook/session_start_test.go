package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// testFixedNow is a stable reference time for staleness tests.
var testFixedNow = time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

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
			} else if got.HookSpecificOutput != nil {
				// SessionStart does NOT use hookSpecificOutput per Claude Code protocol
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

// T-016: Windows CLAUDE_ENV_FILE injection

// TestInjectCLAUDEEnvFile_WithEnvFile verifies that injectCLAUDEEnvFile sets
// CLAUDE_ENV_FILE in settings.local.json when a .env file exists in the project.
func TestInjectCLAUDEEnvFile_WithEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a .env file in the project root
	envFilePath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFilePath, []byte("SOME_VAR=value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Error("injectCLAUDEEnvFile: expected non-empty message when .env exists")
	}

	// Verify settings.local.json was written with CLAUDE_ENV_FILE
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.local.json not created: %v", err)
	}
	var settings struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("unmarshal settings.local.json: %v", err)
	}
	if settings.Env["CLAUDE_ENV_FILE"] != envFilePath {
		t.Errorf("CLAUDE_ENV_FILE = %q, want %q", settings.Env["CLAUDE_ENV_FILE"], envFilePath)
	}
}

// TestInjectCLAUDEEnvFile_NoEnvFile verifies that injectCLAUDEEnvFile is a
// no-op when no .env file exists (returns empty string).
func TestInjectCLAUDEEnvFile_NoEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	msg := injectCLAUDEEnvFile(dir)
	if msg != "" {
		t.Errorf("injectCLAUDEEnvFile: expected empty msg when no .env, got %q", msg)
	}
}

// TestSessionStart_MemoryStaleWrap verifies AC-EXT001-07:
// when a memory file is older than 24h, it is wrapped in <system-reminder>.
func TestSessionStart_MemoryStaleWrap(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// Create .claude/agent-memory/expert-backend/ with a stale memory file.
	agentMemDir := filepath.Join(projectDir, ".claude", "agent-memory", "expert-backend")
	if err := os.MkdirAll(agentMemDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	memPath := filepath.Join(agentMemDir, "note.md")
	content := "---\nname: test\ndescription: d\ntype: user\n---\nbody\n"
	if err := os.WriteFile(memPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Set mtime to 25h ago.
	staleTime := testFixedNow.Add(-25 * time.Hour)
	if err := os.Chtimes(memPath, staleTime, staleTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	result := detectAndWrapStaleMemories(projectDir, testFixedNow)
	if result == "" {
		t.Fatal("detectAndWrapStaleMemories returned empty string; expected staleness caveat")
	}
	if !strings.Contains(result, "<system-reminder>") {
		t.Error("result does not contain <system-reminder>")
	}
	if !strings.Contains(result, "verify against current state") {
		t.Error("result does not contain staleness caveat phrase")
	}
}

// TestSessionStart_AuditDisabled verifies MOAI_MEMORY_AUDIT=0 skips staleness detection.
// Not parallel: uses t.Setenv.
func TestSessionStart_AuditDisabled(t *testing.T) {
	t.Setenv("MOAI_MEMORY_AUDIT", "0")

	projectDir := t.TempDir()
	agentMemDir := filepath.Join(projectDir, ".claude", "agent-memory", "expert-backend")
	if err := os.MkdirAll(agentMemDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	memPath := filepath.Join(agentMemDir, "note.md")
	if err := os.WriteFile(memPath, []byte("---\nname: t\ndescription: d\ntype: user\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	staleTime := testFixedNow.Add(-25 * time.Hour)
	if err := os.Chtimes(memPath, staleTime, staleTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	result := detectAndWrapStaleMemories(projectDir, testFixedNow)
	if result != "" {
		t.Errorf("detectAndWrapStaleMemories with MOAI_MEMORY_AUDIT=0 = %q, want empty", result)
	}
}

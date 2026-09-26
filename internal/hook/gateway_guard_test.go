package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func TestGatewayHooksPreserveRoutingSettings(t *testing.T) {
	for _, provider := range []string{"claude", "glm"} {
		t.Run(provider, func(t *testing.T) {
			t.Setenv(config.EnvMoaiLaunchProvider, provider)
			t.Setenv("TMUX", "/fixture/no-server,1,0")
			t.Setenv("PATH", "")
			root := t.TempDir()
			path := filepath.Join(root, ".claude", "settings.local.json")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			before := []byte(`{"teammateMode":"tmux","env":{"ANTHROPIC_BASE_URL":"https://fixture.invalid","ANTHROPIC_DEFAULT_OPUS_MODEL":"glm-fixture","ANTHROPIC_AUTH_TOKEN":"fixture-token","CLAUDE_CODE_TEAMMATE_DISPLAY":"preserve"},"permissions":{"allow":["Read"]}}`)
			if err := os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			if got := ensureGLMCredentials(root); got != "" {
				t.Fatalf("gateway credential injection %q", got)
			}
			if got := ensureTmuxGLMEnv(root); got != "" {
				t.Fatalf("gateway tmux injection %q", got)
			}
			clearTmuxSessionEnv(context.Background())
			cleanupGLMSettingsLocal(root)
			after, err := os.ReadFile(path)
			if err != nil || string(after) != string(before) {
				t.Fatalf("gateway hook modified routing settings: %s %v", after, err)
			}
			if got := ensureTeammateMode(root); got != "in-process" {
				t.Fatalf("gateway teammate mode %q", got)
			}
			after, _ = os.ReadFile(path)
			var doc struct {
				Env         map[string]string
				Permissions map[string]any
			}
			if err := json.Unmarshal(after, &doc); err != nil {
				t.Fatal(err)
			}
			if doc.Env["ANTHROPIC_AUTH_TOKEN"] != "fixture-token" || doc.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] != "preserve" || doc.Permissions["allow"] == nil {
				t.Fatalf("teammate guard changed unrelated settings: %s", after)
			}
		})
	}
}

func TestGatewayRecordUsesInitialProviderAndKeepsExistingRecord(t *testing.T) {
	root := newMoaiProjectRoot(t)
	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "run")
	t.Setenv(config.EnvMoaiKanbanBackend, "claude")
	t.Setenv(config.EnvMoaiLaunchProvider, "glm")
	input := &HookInput{SessionID: "gateway-initial", ProjectDir: root, CWD: root, Source: "startup"}
	writeKanbanSessionRecord(input)
	rec, err := kanban.Read(root, input.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Backend != kanban.BackendGLM {
		t.Fatalf("backend %q", rec.Backend)
	}
	t.Setenv(config.EnvMoaiLaunchProvider, "claude")
	writeKanbanSessionRecord(input)
	rec, err = kanban.Read(root, input.SessionID)
	if err != nil || rec.Backend != kanban.BackendGLM {
		t.Fatalf("initial record changed: %+v %v", rec, err)
	}
}

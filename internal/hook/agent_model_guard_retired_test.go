package hook

// The workflow.agent_model_guard key no longer has an effect
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-010, AC-AMI-010): subagents inherit the
// main session's model, so a spawn whose declared model differs from the old
// profile cell is never denied, and a project whose workflow.yaml still carries
// `agent_model_guard: {enabled: true}` loads and behaves exactly like one
// without the key.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func loadProjectConfig(t *testing.T, workflowYAML string) *config.Config {
	t.Helper()
	root := t.TempDir()
	if workflowYAML != "" {
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(workflowYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := config.NewConfigManager().Load(root)
	if err != nil {
		t.Fatalf("config with workflow.yaml %q failed to load: %v", workflowYAML, err)
	}
	return cfg
}

func TestAgentModelGuardKey_LoadsAndHasNoEffect(t *testing.T) {
	withKey := loadProjectConfig(t, "workflow:\n  agent_model_guard:\n    enabled: true\n")
	withoutKey := loadProjectConfig(t, "workflow: {}\n")

	var llm config.LLMConfig
	_, resolved := resolveAgentModel(llm, "Explore")
	for _, model := range []string{"haiku", resolved, ""} {
		var decisions [2]string
		for i, cfg := range []*config.Config{withKey, withoutKey} {
			h := &preToolHandler{cfg: &auditConfigProvider{cfg: cfg}, policy: DefaultSecurityPolicy(), projectDir: t.TempDir()}
			out, err := h.Handle(context.Background(), agentInput("Explore", model))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			assertNotDeny(t, out)
			decisions[i] = decisionOf(out) + "|" + reasonOf(out)
		}
		if decisions[0] != decisions[1] {
			t.Errorf("model %q: with key %q, without key %q — the key must have no effect", model, decisions[0], decisions[1])
		}
	}
}

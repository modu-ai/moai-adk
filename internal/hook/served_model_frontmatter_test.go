package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// frontmatterProject builds a MoAI-shaped project root whose agent file for
// agent carries the given frontmatter model line ("" omits the line), and
// exports it as the project root the way the runtime does.
func frontmatterProject(t *testing.T, agent, modelLine string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll .moai: %v", err)
	}
	if agent != "" {
		dir := filepath.Join(root, ".claude", "agents", "moai")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll agents: %v", err)
		}
		body := "---\nname: " + agent + "\n" + modelLine + "effort: high\n---\n\nbody\n"
		if err := os.WriteFile(filepath.Join(dir, agent+".md"), []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile agent: %v", err)
		}
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
}

// A bare plan-auditor spawn passes no model argument, so the sibling meta.json
// carries none; the declared model the observation compares against must then
// come from the agent definition's frontmatter, which is the canonical
// declaration (agent-common-protocol.md § Subagent Model and Effort).
func TestServedModel_FrontmatterDeclaresWhenSpawnPassesNone(t *testing.T) {
	cases := []struct {
		name      string
		agent     string
		modelLine string
		meta      map[string]string
		rows      []string
		want      string
		wantExp   string
	}{
		{"a_frontmatter_opus_served_opus_is_ok", "plan-auditor", "model: opus\n",
			map[string]string{"agentType": "plan-auditor"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictOK, "opus"},
		{"b_frontmatter_opus_served_glm_is_drift", "plan-auditor", "model: opus\n",
			map[string]string{"agentType": "plan-auditor"}, repeatRows(assistantRow("glm-5.3-flash"), 2), ServedVerdictDrift, "opus"},
		{"c_spawn_argument_wins_over_frontmatter", "sync-auditor", "model: opus\n",
			map[string]string{"agentType": "sync-auditor", "model": "sonnet"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictDrift, "sonnet"},
		{"d_frontmatter_inherit_stays_unknown", "plan-auditor", "model: inherit\n",
			map[string]string{"agentType": "plan-auditor"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictUnknown, ""},
		{"e_frontmatter_without_model_stays_unknown", "plan-auditor", "",
			map[string]string{"agentType": "plan-auditor"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictUnknown, ""},
		{"f_no_agent_file_stays_unknown", "", "",
			map[string]string{"agentType": "plan-auditor"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictUnknown, ""},
		{"g_path_like_agent_type_is_not_looked_up", "plan-auditor", "model: opus\n",
			map[string]string{"agentType": "../moai/plan-auditor"}, repeatRows(assistantRow("claude-opus-5-5"), 2), ServedVerdictUnknown, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frontmatterProject(t, tc.agent, tc.modelLine)
			path := writeSubagentTranscript(t, t.TempDir(), "a1", tc.rows, tc.meta)
			obs := ObserveServedModel(path, "", &config.Config{})
			if obs.Verdict != tc.want {
				t.Fatalf("verdict = %q, want %q (obs=%+v)", obs.Verdict, tc.want, obs)
			}
			if obs.ExpectedModel != tc.wantExp {
				t.Fatalf("expected model = %q, want %q", obs.ExpectedModel, tc.wantExp)
			}
		})
	}
}

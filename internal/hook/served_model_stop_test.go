package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// newServedRoot creates a project tree carrying .moai/ so the SubagentStop
// observer has a project root to write its audit row under.
func newServedRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	return root
}

// readServedAuditRows returns every row of the agent-model audit log as a
// generic map, plus the raw lines.
func readServedAuditRows(t *testing.T, root string) ([]map[string]any, []string) {
	t.Helper()
	f, err := os.Open(filepath.Join(root, ".moai", "logs", agentModelAuditFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		t.Fatalf("open audit log: %v", err)
	}
	defer func() { _ = f.Close() }()
	var rows []map[string]any
	var raw []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		raw = append(raw, sc.Text())
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("audit row is not JSON: %q", sc.Text())
		}
		rows = append(rows, m)
	}
	return rows, raw
}

// runServedStop runs the SubagentStop handler with CLAUDE_PROJECT_DIR pointed at
// the input's tree — the root the runtime exports and the served row lands
// under — so the row never escapes into the package directory.
func runServedStop(t *testing.T, cfg ConfigProvider, input *HookInput) *HookOutput {
	t.Helper()
	if input.CWD != "" {
		t.Setenv(config.EnvClaudeProjectDir, input.CWD)
	}
	out, err := NewSubagentStopHandlerWithConfig(cfg).Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("SubagentStop Handle returned an error: %v", err)
	}
	return out
}

// AC-SMA-003 (M5 shape) — the expectation is DECLARATION-ONLY: the transcript's
// declared model is the expected model, and the configuration provider no
// longer participates. The control inverts with the design: three config
// shapes (claude-mode, glm-mode, zero) must produce the SAME declaration-based
// answer — the provider is provably not consulted. (The former profile /
// agent-override shapes retired with the per-agent profile matrix.)
func TestServedModel_ResolvedFromActiveProfile(t *testing.T) {
	claudeMode := &config.Config{LLM: config.LLMConfig{
		Harness: "claude", GLMEnvVar: "GLM_API_KEY",
	}}
	glmMode := &config.Config{LLM: config.LLMConfig{
		Mode: "glm", TeamMode: "glm", ClaudeBin: "/nonexistent/moai-test-bin",
	}}
	zero := &config.Config{}

	cases := []struct {
		name         string
		cfg          *config.Config
		wantResolved string
		wantVerdict  string
	}{
		{"a_claude_mode_config_same_as_declaration", claudeMode, "claude-sonnet-5", ServedVerdictOK},
		{"b_glm_mode_config_same_as_declaration", glmMode, "claude-sonnet-5", ServedVerdictOK},
		{"control_zero_config_same_as_declaration", zero, "claude-sonnet-5", ServedVerdictOK},
	}
	got := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newServedRoot(t)
			tr := writeSubagentTranscript(t, filepath.Join(t.TempDir(), "subagents"), "a3",
				repeatRows(assistantRow("claude-sonnet-5"), 2),
				map[string]string{"agentType": "manager-develop", "model": "claude-sonnet-5"})
			runServedStop(t, staticConfigProvider{cfg: tc.cfg}, &HookInput{
				CWD: root, SessionID: "s-3", AgentID: "a3", AgentType: "manager-develop",
				AgentTranscriptPath: tr, HookEventName: string(EventSubagentStop),
			})
			rows, _ := readServedAuditRows(t, root)
			if len(rows) != 1 {
				t.Fatalf("audit rows = %d, want 1", len(rows))
			}
			if rows[0]["resolved_model"] != tc.wantResolved {
				t.Fatalf("resolved_model = %v, want %q", rows[0]["resolved_model"], tc.wantResolved)
			}
			if rows[0]["verdict"] != tc.wantVerdict {
				t.Fatalf("verdict = %v, want %q", rows[0]["verdict"], tc.wantVerdict)
			}
			got[tc.name] = tc.wantVerdict + "/" + tc.wantResolved
		})
	}
	if got["a_claude_mode_config_same_as_declaration"] != got["control_zero_config_same_as_declaration"] {
		t.Fatalf("the claude-mode result (%q) differs from the zero-config result (%q) — the provider must not participate", got["a_claude_mode_config_same_as_declaration"], got["control_zero_config_same_as_declaration"])
	}
}

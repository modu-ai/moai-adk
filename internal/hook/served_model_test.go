package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// assistantRow renders one transcript row of type assistant carrying model.
func assistantRow(model string) string {
	b, _ := json.Marshal(map[string]any{
		"type":    "assistant",
		"message": map[string]any{"model": model, "role": "assistant"},
	})
	return string(b)
}

// assistantRowNoModel renders an assistant row whose message has no model field.
func assistantRowNoModel() string {
	return `{"type":"assistant","message":{"role":"assistant"}}`
}

// userRow renders a non-assistant row that also mentions a model, so a reader
// that does not filter on the row type would pick it up.
func userRow() string {
	return `{"type":"user","message":{"model":"glm-5.3-flash","role":"user"}}`
}

// writeSubagentTranscript writes <dir>/agent-<id>.jsonl and, when meta is
// non-nil, the sibling agent-<id>.meta.json. It returns the transcript path.
func writeSubagentTranscript(t *testing.T, dir, agentID string, rows []string, meta map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "agent-"+agentID+".jsonl")
	body := strings.Join(rows, "\n")
	if len(rows) > 0 {
		body += "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile transcript: %v", err)
	}
	if meta != nil {
		b, _ := json.Marshal(meta)
		if err := os.WriteFile(filepath.Join(dir, "agent-"+agentID+".meta.json"), b, 0o600); err != nil {
			t.Fatalf("WriteFile meta: %v", err)
		}
	}
	return path
}

func repeatRows(row string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = row
	}
	return out
}

// AC-SMA-001 — verdict table: ok / served_drift / mixed drift / unmapped.
func TestServedModel_Classify(t *testing.T) {
	zero := &config.Config{}
	cases := []struct {
		name        string
		agentType   string
		meta        map[string]string
		rows        []string
		cfg         *config.Config
		wantVerdict string
		wantServed  []string
	}{
		{
			name:        "a_declared_opus_served_opus_is_ok",
			agentType:   "manager-develop",
			meta:        map[string]string{"agentType": "manager-develop", "model": "opus"},
			rows:        repeatRows(assistantRow("claude-opus-5-5"), 3),
			cfg:         zero,
			wantVerdict: ServedVerdictOK,
			wantServed:  []string{"claude-opus-5-5"},
		},
		{
			name:        "b_declared_opus_served_glm_is_drift",
			agentType:   "manager-develop",
			meta:        map[string]string{"agentType": "manager-develop", "model": "opus"},
			rows:        repeatRows(assistantRow("glm-5.3-flash"), 5),
			cfg:         zero,
			wantVerdict: ServedVerdictDrift,
			wantServed:  []string{"glm-5.3-flash"},
		},
		{
			name:        "c_declared_opus_mixed_served_is_drift_with_both_members",
			agentType:   "manager-develop",
			meta:        map[string]string{"agentType": "manager-develop", "model": "opus"},
			rows:        []string{assistantRow("claude-opus-5"), assistantRow("glm-5.3"), assistantRow("claude-opus-5")},
			cfg:         zero,
			wantVerdict: ServedVerdictDrift,
			wantServed:  []string{"claude-opus-5", "glm-5.3"},
		},
		{
			name:        "d_undeclared_agent_is_unknown_no_expectation",
			agentType:   "custom-harness-specialist",
			meta:        map[string]string{"agentType": "custom-harness-specialist"},
			rows:        []string{assistantRow("claude-opus-5-5")},
			cfg:         zero,
			wantVerdict: ServedVerdictUnknown,
			wantServed:  []string{"claude-opus-5-5"},
		},
		{
			name:        "e_declared_context_suffix_id_served_bare_id_is_ok",
			agentType:   "manager-develop",
			meta:        map[string]string{"agentType": "manager-develop", "model": "claude-opus-5[1m]"},
			rows:        []string{assistantRow("claude-opus-5")},
			cfg:         zero,
			wantVerdict: ServedVerdictOK,
			wantServed:  []string{"claude-opus-5"},
		},
		{
			name:        "f_declared_context_suffix_alias_served_family_is_ok",
			agentType:   "plan-auditor",
			meta:        map[string]string{"agentType": "plan-auditor", "model": "opus[1m]"},
			rows:        []string{assistantRow("claude-opus-5-5")},
			cfg:         zero,
			wantVerdict: ServedVerdictOK,
			wantServed:  []string{"claude-opus-5-5"},
		},
		{
			name:        "g_control_context_suffix_alias_served_glm_stays_drift",
			agentType:   "plan-auditor",
			meta:        map[string]string{"agentType": "plan-auditor", "model": "opus[1m]"},
			rows:        []string{assistantRow("glm-5.3-flash")},
			cfg:         zero,
			wantVerdict: ServedVerdictDrift,
			wantServed:  []string{"glm-5.3-flash"},
		},
		{
			name:        "h_declared_inherit_is_no_declaration_unknown",
			agentType:   "fork",
			meta:        map[string]string{"agentType": "fork", "model": "inherit"},
			rows:        []string{assistantRow("claude-sonnet-5")},
			cfg:         zero,
			wantVerdict: ServedVerdictUnknown,
			wantServed:  []string{"claude-sonnet-5"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := writeSubagentTranscript(t, t.TempDir(), "a1", tc.rows, tc.meta)
			obs := ObserveServedModel(path, tc.agentType, tc.cfg)
			if obs.Verdict != tc.wantVerdict {
				t.Fatalf("verdict = %q, want %q (obs=%+v)", obs.Verdict, tc.wantVerdict, obs)
			}
			if !slices.Equal(obs.ServedModels, tc.wantServed) {
				t.Fatalf("served = %v, want %v", obs.ServedModels, tc.wantServed)
			}
		})
	}
}

// AC-SMA-002 — every uncertain state is `unknown`, never `ok`; the positive
// controls show the reader is not simply returning `unknown` for everything.
func TestServedModel_UnknownNeverOK(t *testing.T) {
	zero := &config.Config{}
	opusMeta := map[string]string{"agentType": "manager-develop", "model": "opus"}
	tiny := servedReadBudget{MaxBytes: 64, MaxLines: 2}
	matching := repeatRows(assistantRow("claude-opus-5-5"), 10)

	cases := []struct {
		name        string
		rows        []string
		meta        map[string]string
		absent      bool
		cfg         *config.Config
		budget      *servedReadBudget
		wantVerdict string
	}{
		{name: "a_absent_transcript", absent: true, meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictUnknown},
		{name: "b_empty_model_value", rows: []string{assistantRow("")}, meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictUnknown},
		{name: "c_model_field_absent", rows: []string{assistantRowNoModel(), userRow()}, meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictUnknown},
		{name: "d_synthetic_only", rows: repeatRows(assistantRow("<synthetic>"), 2), meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictUnknown},
		{name: "e_over_injected_budget", rows: matching, meta: opusMeta, cfg: zero, budget: &tiny, wantVerdict: ServedVerdictUnknown},
		{name: "f_undeclared_and_no_configuration", rows: matching, meta: map[string]string{"agentType": "manager-develop"}, cfg: nil, wantVerdict: ServedVerdictUnknown},
		{name: "g_control_synthetic_plus_real_is_ok", rows: []string{assistantRow("<synthetic>"), assistantRow("claude-opus-5-5")}, meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictOK},
		{name: "h_control_same_fixture_default_budget_is_ok", rows: matching, meta: opusMeta, cfg: zero, wantVerdict: ServedVerdictOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "agent-missing.jsonl")
			if !tc.absent {
				path = writeSubagentTranscript(t, dir, "a1", tc.rows, tc.meta)
			}
			budget := defaultServedReadBudget
			if tc.budget != nil {
				budget = *tc.budget
			}
			obs := observeServedModel(path, "manager-develop", tc.cfg, budget)
			if obs.Verdict != tc.wantVerdict {
				t.Fatalf("verdict = %q, want %q (obs=%+v)", obs.Verdict, tc.wantVerdict, obs)
			}
			if tc.wantVerdict == ServedVerdictUnknown && obs.Cause == "" {
				t.Fatalf("an unknown verdict must name its cause (obs=%+v)", obs)
			}
			if slices.Contains(obs.ServedModels, "<synthetic>") {
				t.Fatalf("served set must never carry the synthetic marker: %v", obs.ServedModels)
			}
		})
	}
}

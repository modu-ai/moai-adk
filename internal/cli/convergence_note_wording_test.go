package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestConvergenceNoteMatchesBlockDecision pins the residual-risk note wording to
// what the convergence actually decides: a split that includes a required FAIL
// fails the overall verdict (and the multi-review gate blocks on it), so the note
// must not call it advisory; a split that stays advisory must keep saying so.
func TestConvergenceNoteMatchesBlockDecision(t *testing.T) {
	cases := []struct {
		name        string
		verdicts    []PerBackendVerdict
		wantOverall string
		wantIn      []string
		wantNotIn   []string
	}{
		{
			name: "required split blocks",
			verdicts: []PerBackendVerdict{
				{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "fail"},
				{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "pass"},
			},
			wantOverall: overallVerdictFail,
			wantIn:      []string{"required-backend FAIL: claude", "pass=[codex(required)]", "fail=[claude(required)]"},
			wantNotIn:   []string{"NOT a block", "advisory"},
		},
		{
			name: "required fail against advisory pass blocks",
			verdicts: []PerBackendVerdict{
				{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "fail"},
				{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "pass"},
			},
			wantOverall: overallVerdictFail,
			wantIn:      []string{"required-backend FAIL: codex"},
			wantNotIn:   []string{"NOT a block"},
		},
		{
			name: "advisory-only conflict does not block",
			verdicts: []PerBackendVerdict{
				{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
				{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "fail"},
			},
			wantOverall: overallVerdictPass,
			wantIn:      []string{"advisory, NOT a block"},
			wantNotIn:   []string{"required-backend FAIL"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := converge(tc.verdicts)
			if r.OverallVerdict != tc.wantOverall {
				t.Fatalf("overall_verdict = %q, want %q", r.OverallVerdict, tc.wantOverall)
			}
			for _, s := range tc.wantIn {
				if !strings.Contains(r.ResidualRiskNote, s) {
					t.Errorf("residual_risk_note = %q, want it to contain %q", r.ResidualRiskNote, s)
				}
			}
			for _, s := range tc.wantNotIn {
				if strings.Contains(r.ResidualRiskNote, s) {
					t.Errorf("residual_risk_note = %q, must not contain %q", r.ResidualRiskNote, s)
				}
			}
		})
	}
}

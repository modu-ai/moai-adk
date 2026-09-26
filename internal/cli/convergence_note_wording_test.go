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
	// Gate-unmet path: an explicitly-required backend returned no verdict, so
	// enforceRequiredGateUnmet flips overall to fail after converge() wrote an
	// advisory note. The flipped result must not still say "NOT a block".
	t.Run("required gate unmet flips advisory note", func(t *testing.T) {
		verdicts := []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "fail"},
		}
		r := enforceRequiredGateUnmet(converge(verdicts), verdicts, config.AuditGates{Codex: config.AuditGateRequired})
		if r.OverallVerdict != overallVerdictFail {
			t.Fatalf("overall_verdict = %q, want %q", r.OverallVerdict, overallVerdictFail)
		}
		if !strings.Contains(r.ResidualRiskNote, "required gate unmet") {
			t.Errorf("residual_risk_note = %q, want the unmet gate named", r.ResidualRiskNote)
		}
		if strings.Contains(r.ResidualRiskNote, "NOT a block") {
			t.Errorf("residual_risk_note = %q, must not contain %q on a failing verdict", r.ResidualRiskNote, "NOT a block")
		}
	})

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

package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestSynthesizeReviewOutput_RecordsSignalDivergence is the first half of
// AC-CVS-004.
//
// Adopting the conservative verdict is not enough on its own. A verdict that
// came out of two signals contradicting each other is a different fact from one
// the review stated plainly, and if the result carries only the value, the next
// person diagnosing an incident has no way to tell them apart — which is how the
// review that produced this card took as long as it did.
func TestSynthesizeReviewOutput_RecordsSignalDivergence(t *testing.T) {
	const body = "Verdict: pass — nothing to block on.\n\n- [P1] SQL injection at db.go:12"

	out := synthesizeReviewOutput(body, codexMethodTurnStart)
	if out.Verdict != "fail" {
		t.Errorf("adopted Verdict = %q, want \"fail\" (a finding outranks a stated pass)", out.Verdict)
	}
	if out.SynthesisNote == "" {
		t.Error("SynthesisNote is empty; a divergence between signals must be recorded, not only resolved")
	}
}

// TestSynthesizeReviewOutput_NoNoteWhenSignalsAgree keeps the record meaningful.
// A note attached to every review says nothing; it has to mark the reviews where
// the signals actually disagreed.
func TestSynthesizeReviewOutput_NoNoteWhenSignalsAgree(t *testing.T) {
	cases := map[string]string{
		"agreeing signals": "Verdict: pass — nothing to block on.\n\nPASS 0.99 / 1.00",
		"single signal":    "Verdict: fail — merge blocked.",
		"no signal":        "I walked the diff and moved on.",
	}
	for name, body := range cases {
		if note := synthesizeReviewOutput(body, codexMethodTurnStart).SynthesisNote; note != "" {
			t.Errorf("%s: SynthesisNote = %q, want empty", name, note)
		}
	}
}

// TestConverge_SurfacesSignalDivergence_WithoutBlocking is the second half of
// AC-CVS-004.
//
// The divergence has to reach the convergence layer, and it has to arrive as
// INFORMATION. disagreement_flag is not a block category (mcp_convergence.go
// invariant C3), so the assertion pins the overall verdict against the same
// convergence run with the note cleared: whatever the existing policy produced,
// it still produces.
func TestConverge_SurfacesSignalDivergence_WithoutBlocking(t *testing.T) {
	withNote := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, VerdictInconclusive),
	}
	withNote[1].SynthesisNote = "codex signals diverged: stated verdict label=pass, scored verdict line=inconclusive; adopted inconclusive"

	without := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, VerdictInconclusive),
	}

	got := converge(withNote)
	baseline := converge(without)

	if !got.DisagreementFlag {
		t.Error("disagreement_flag = false; a backend whose own signals diverged must be surfaced")
	}
	if !strings.Contains(got.ResidualRiskNote, "diverged") {
		t.Errorf("residual_risk_note = %q; want the divergence named in it", got.ResidualRiskNote)
	}
	if got.OverallVerdict != baseline.OverallVerdict {
		t.Errorf("overall_verdict = %q with the note, %q without it; recording a divergence must not become a new block category",
			got.OverallVerdict, baseline.OverallVerdict)
	}
}

// TestRunMultiAudit_ForwardsSynthesisNoteToPerBackendVerdict closes the wiring
// gap the field would otherwise leave: a note that lives only on ReviewOutput
// and is dropped when the fan-out assembles PerBackendVerdict is a field that
// converge can never see. The struct would look correct and the feature would
// not exist.
func TestRunMultiAudit_ForwardsSynthesisNoteToPerBackendVerdict(t *testing.T) {
	const note = "codex signals diverged: stated verdict label=pass, severity-tagged finding bullet=fail; adopted fail"

	prev := backendCall
	backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		out := ReviewOutput{Verdict: "pass", Summary: backend, Findings: []Finding{}, NextSteps: []string{}}
		if backend == BackendCodex {
			out.Verdict = "fail"
			out.SynthesisNote = note
		}
		return out
	}
	t.Cleanup(func() { backendCall = prev })

	r := runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", MultiAuditConfig{
		Gates: config.AuditGates{
			Claude: config.AuditGateRequired,
			Codex:  config.AuditGateRequired,
			GLM:    config.AuditGateOff,
		},
	}, nil)

	for _, v := range r.PerBackendVerdicts {
		if v.Backend != BackendCodex {
			continue
		}
		if v.SynthesisNote != note {
			t.Errorf("codex PerBackendVerdict.SynthesisNote = %q, want it forwarded verbatim from the ReviewOutput", v.SynthesisNote)
		}
		return
	}
	t.Fatal("no codex entry in per_backend_verdicts")
}

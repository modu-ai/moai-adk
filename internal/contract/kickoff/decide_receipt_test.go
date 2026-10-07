package kickoff_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// TestKickoffEvaluatorRequiredBackendRefusal is AC-ACE-022's kickoff seam arm
// (SPEC-AUDIT-CEILING-001 REQ-ACE-009): with a required backend configured on
// the tree, the evaluator's planAuditCheck refuses a verdict carrying that
// backend's fail receipt although the auditor's own label, score, must-pass,
// blocking, and hash all pass. The control arm — the same tree and verdict
// with a pass receipt — decides, proving the refusal comes from the receipt
// check and not from the mere presence of the configuration.
func TestKickoffEvaluatorRequiredBackendRefusal(t *testing.T) {
	build := func(receiptLine string) func(t *testing.T) *dfx {
		return func(t *testing.T) *dfx {
			f := newDecide(t, "manager-spec", auditReport{name: "plan-audit-1.md", verdict: "PASS", score: "0.90"})
			// Configure exactly one required backend on the fixture tree.
			f.p.WriteFile(".moai/config/sections/workflow.yaml", "workflow:\n  audit:\n    gates:\n      claude: required\n")
			// Attach the convergence receipt to the exported verdict file.
			path := ".moai/reports/" + signtest.Card + "/plan-audit-1.md"
			f.p.WriteFile(path, string(f.p.ReadFile(path))+"convergence_overall: pass\n"+receiptLine+"\n")
			return f
		}
	}

	t.Run("fail_receipt_refuses", func(t *testing.T) {
		f := build("required_backend: claude fail")(t)
		r0, e0 := f.counts()
		res := f.mustDecide(f.input("llm+jev", "approve", false))
		f.assertNoDecision(res, "precondition:a", r0, e0)
	})

	t.Run("pass_receipt_decides", func(t *testing.T) {
		f := build("required_backend: claude pass")(t)
		r0, e0 := f.counts()
		res := f.mustDecide(f.input("llm+jev", "approve", false))
		if f.jevCalls != 1 {
			t.Fatalf("control arm stalled at precondition (a): outcome %q reason %q, Jev calls %d", res.Outcome, res.Reason, f.jevCalls)
		}
		f.assertReceipt(r0, e0)
	})
}

package kickoff_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestKickoffEvaluatorCeilingRefusal is AC-ACE-015's kickoff seam arm
// (SPEC-AUDIT-CEILING-001 REQ-ACE-003): a ceiling-hit round whose verdict
// fails admission refuses at the production kickoff evaluator — the refusal
// reason names the ceiling outcome and the outcome persists to the audit
// trail, so a mutant that skips the engine wiring cannot pass either arm.
// The control arm — an admission-clean verdict at the same ceiling state —
// decides (REQ-ACE-013's pass-through), proving the refusal comes from the
// ceiling engine and not from the label alone.
func TestKickoffEvaluatorCeilingRefusal(t *testing.T) {
	build := func(verdict string) func(t *testing.T) *dfx {
		return func(t *testing.T) *dfx {
			// Pass the explicit report so newDecide writes no default
			// plan-audit-1.md — its equal N would race the highest-N pick.
			f := newDecide(t, "manager-spec", auditReport{name: "plan-audit-iter1.md", verdict: verdict, score: "0.90"})
			// Fixture ceiling: Tier L = 1, no delta rounds — one iteration
			// is already the final hit.
			f.p.WriteFile(".moai/config/sections/harness.yaml", "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 0\n    on_final_hit: hold-and-split\n")
			// Rewrite the report with the SPEC named in its header: the
			// counter's convention attribution reads it (REQ-ACE-001).
			body := "# SPEC Review Report: " + signtest.SpecID + "\nVerdict: " + verdict + "\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: " + f.hash() + "\n"
			f.p.WriteFile(".moai/reports/"+signtest.Card+"/plan-audit-iter1.md", body)
			return f
		}
	}

	t.Run("failing_verdict_refuses_at_ceiling", func(t *testing.T) {
		f := build("FAIL")(t)
		r0, e0 := f.counts()
		res := f.mustDecide(f.input("llm+jev", "approve", false))
		f.assertNoDecision(res, "precondition:a", r0, e0)
		if len(res.Preconditions) == 0 || res.Preconditions[0].Name != "a" ||
			!strings.Contains(res.Preconditions[0].Detail, "ceiling refusal") {
			t.Fatalf("precondition a detail %v, want a ceiling refusal reason", res.Preconditions)
		}
		// The refusal persisted to the machine-local audit trail.
		trail, err := os.ReadFile(filepath.Join(f.p.Root, ".moai", "state", "audit-enforcement.log"))
		if err != nil {
			t.Fatalf("ceiling trail not written: %v", err)
		}
		if !strings.Contains(string(trail), "ceiling-refusal") || !strings.Contains(string(trail), signtest.SpecID) {
			t.Fatalf("trail line incomplete: %s", trail)
		}
	})

	t.Run("clean_verdict_passes_through", func(t *testing.T) {
		f := build("PASS")(t)
		r0, e0 := f.counts()
		res := f.mustDecide(f.input("llm+jev", "approve", false))
		if f.jevCalls != 1 {
			t.Fatalf("control arm stalled: outcome %q reason %q detail %q", res.Outcome, res.Reason, res.Preconditions[0].Detail)
		}
		f.assertReceipt(r0, e0)
		// The pass-through outcome carries no refusal record.
		if trail, err := os.ReadFile(filepath.Join(f.p.Root, ".moai", "state", "audit-enforcement.log")); err == nil &&
			strings.Contains(string(trail), "ceiling-refusal") {
			t.Fatalf("pass-through wrote a refusal record: %s", trail)
		}
	})
}

// TestResolveRequiredBackendsIsImportable documents the gate-set resolver's
// home package surface the seam shares with the homestate card transition.
func TestResolveRequiredBackendsIsImportable(t *testing.T) {
	gs, err := runtime.ResolveRequiredBackends(t.TempDir())
	if err != nil || len(gs.Required) != 0 {
		t.Fatalf("absent config: %+v %v", gs, err)
	}
}

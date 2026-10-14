// mcp_convergence_not_converged_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card
// t1632, REQ-MN-022): convergence does NOT pass on the caller's own submitted
// verdict when a REQUIRED backend is inconclusive — the outcome is a distinct
// not-converged result (overall inconclusive), never a pass.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestConverge_RequiredInconclusive_NeverPassesOnSubmittedVerdict pins
// REQ-MN-022: a required backend left inconclusive means the participant
// evidence threshold is unmet — the claude participant's pass (the caller's
// own submitted verdict) cannot rescue the convergence to pass. The overall
// is inconclusive (the distinct not-converged outcome), and the fail-open
// backends are still named for the orchestrator.
func TestConverge_RequiredInconclusive_NeverPassesOnSubmittedVerdict(t *testing.T) {
	verdicts := []PerBackendVerdict{
		{Backend: BackendClaude, Gate: config.AuditGateAdvisory, Verdict: "pass"},
		{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
	}
	r := converge(verdicts)
	if r.OverallVerdict == overallVerdictPass {
		t.Fatalf("required-inconclusive converged to pass on the submitted claude verdict: %+v", r)
	}
	if r.OverallVerdict != VerdictInconclusive {
		t.Errorf("overall = %q, want %q (the distinct not-converged outcome)", r.OverallVerdict, VerdictInconclusive)
	}
	// The codex backend's true inconclusive verdict is preserved for the
	// audit trail.
	found := false
	for _, v := range r.PerBackendVerdicts {
		if v.Backend == BackendCodex && v.Verdict == VerdictInconclusive {
			found = true
		}
	}
	if !found {
		t.Error("the codex backend's inconclusive verdict was lost from per_backend_verdicts")
	}
}

// TestConverge_RequiredPassStillPasses pins the positive space: a required
// backend that PASSes still converges to pass — the fail-closed fix narrows
// only the inconclusive-rescue branch.
func TestConverge_RequiredPassStillPasses(t *testing.T) {
	verdicts := []PerBackendVerdict{
		{Backend: BackendClaude, Gate: config.AuditGateAdvisory, Verdict: "pass"},
		{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "pass"},
	}
	r := converge(verdicts)
	if r.OverallVerdict != overallVerdictPass {
		t.Fatalf("all-required-pass = %q, want pass", r.OverallVerdict)
	}
}

// TestConverge_RequiredFailStillFails pins the conservative half: a required
// FAIL still fails.
func TestConverge_RequiredFailStillFails(t *testing.T) {
	verdicts := []PerBackendVerdict{
		{Backend: BackendClaude, Gate: config.AuditGateAdvisory, Verdict: "pass"},
		{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "fail"},
	}
	r := converge(verdicts)
	if r.OverallVerdict != overallVerdictFail {
		t.Fatalf("required fail = %q, want fail", r.OverallVerdict)
	}
}

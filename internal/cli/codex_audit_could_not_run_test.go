// codex_audit_could_not_run_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-021): a review that could not be performed (transport loss, unparsable
// reply, missing binary) keeps a NOT-PERFORMED verdict distinct from a content
// failure — the required gate does not convert it to verdict `fail`, the
// backend's stderr is preserved for diagnosis, and the leg retries exactly once
// before reporting.
package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestCodexAudit_CouldNotRunNotConvertedToFail pins the REQ-MN-021 split at the
// single-backend surface: a transport-level could-not-run under an explicit
// required gate stays inconclusive (not-performed) — the required-gate
// enforcement no longer converts it into a reviewed failure.
func TestCodexAudit_CouldNotRunNotConvertedToFail(t *testing.T) {
	root := newProbeProject(t, "SPEC-CNR-911")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	sess := withCodexSession(t, nil)
	sess.startErr = errors.New("transport dropped mid-handshake")

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError — could-not-run stays a structured result")
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v == "fail" {
		t.Errorf("verdict = fail — a could-not-run was converted into a reviewed failure (REQ-MN-021)")
	}
	g, _ := got["gate_unmet"].(string)
	if !strings.Contains(g, "could-not-run") {
		t.Errorf("gate_unmet = %q, want it to carry could-not-run", g)
	}
}

// TestCodexAudit_StderrPreservedOnCouldNotRun pins the diagnostic half: when
// the codex turn fails, the backend's stderr rides the structured result. The
// fake session supplies the stderr a real process would have written.
func TestCodexAudit_StderrPreservedOnCouldNotRun(t *testing.T) {
	root := newProbeProject(t, "SPEC-CNR-912")
	withCodexSessionStderr(t, nil, "codex: fatal: quota exhausted")

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	if res.IsError {
		t.Fatalf("unexpected IsError result")
	}
	got := structuredMap(t, res)
	if se, _ := got["stderr"].(string); !strings.Contains(se, "quota exhausted") {
		t.Errorf("stderr = %q, want the backend's stderr preserved", se)
	}
}

// TestCodexAudit_CouldNotRunRetriesOnce pins the retry contract: a
// could-not-run leg retries EXACTLY once — a second call that succeeds is
// honored, and the two calls are the ceiling (no loop).
func TestCodexAudit_CouldNotRunRetriesOnce(t *testing.T) {
	root := newProbeProject(t, "SPEC-CNR-913")
	writeCodexAuditGate(t, root, config.AuditGateRequired)

	// Card t1632 CI repair: the audit handler resolves the codex binary BEFORE
	// the codexReviewRPC seam — on CI no codex is installed, so the real
	// LookPath failed-closed there before the stubbed seam was ever reached
	// (green locally only where an operator codex exists). Point the seam at
	// the stub path the rest of this suite uses.
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	t.Cleanup(func() { codexLookPath = prevLook })

	calls := 0
	prev := codexReviewRPC
	codexReviewRPC = func(ctx context.Context, binaryPath, method string, params map[string]any) (ReviewOutput, error) {
		calls++
		if calls == 1 {
			return inconclusiveReview("transport dropped mid-turn"), errors.New("transport dropped mid-turn")
		}
		return ReviewOutput{Verdict: "pass", Summary: "clean after retry", Findings: []Finding{}, NextSteps: []string{}}, nil
	}
	t.Cleanup(func() { codexReviewRPC = prev })

	res := callToolCodexAudit(t, map[string]any{"project_root": root})
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != "pass" {
		t.Errorf("verdict = %q, want pass (the retry succeeded)", v)
	}
	if calls != 2 {
		t.Errorf("codexReviewRPC calls = %d, want exactly 2 (original + one retry)", calls)
	}
}

// TestCodexAudit_SuccessVerdictRetriesNever pins the positive space: a leg
// that produced a verdict never retries.
func TestCodexAudit_SuccessVerdictRetriesNever(t *testing.T) {
	root := newProbeProject(t, "SPEC-CNR-914")
	// Card t1632 CI repair: same binary-resolution seam stub as the retry-once
	// test — the handler's LookPath must not depend on an installed codex.
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	t.Cleanup(func() { codexLookPath = prevLook })

	calls := 0
	prev := codexReviewRPC
	codexReviewRPC = func(ctx context.Context, binaryPath, method string, params map[string]any) (ReviewOutput, error) {
		calls++
		return ReviewOutput{Verdict: "pass", Summary: "clean", Findings: []Finding{}, NextSteps: []string{}}, nil
	}
	t.Cleanup(func() { codexReviewRPC = prev })

	_ = callToolCodexAudit(t, map[string]any{"project_root": root})
	if calls != 1 {
		t.Errorf("codexReviewRPC calls = %d, want 1 (no retry on a real verdict)", calls)
	}
}

// TestEnforceRequiredGateUnmet_CouldNotRunNotConvergedToFail pins the
// convergence half of REQ-MN-021: a required backend whose verdict is
// inconclusive (could not run) leaves the overall at the distinct
// not-converged outcome — never converted to overall fail, never pass.
func TestEnforceRequiredGateUnmet_CouldNotRunNotConvergedToFail(t *testing.T) {
	gates := config.AuditGates{Codex: config.AuditGateRequired}
	r := ConvergenceResult{OverallVerdict: VerdictInconclusive}
	verdicts := []PerBackendVerdict{
		{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
	}
	got := enforceRequiredGateUnmet(r, verdicts, gates)
	if got.OverallVerdict == overallVerdictFail {
		t.Errorf("overall = fail — could-not-run converted into a converged failure (REQ-MN-021)")
	}
	if got.OverallVerdict != VerdictInconclusive {
		t.Errorf("overall = %q, want %q (the distinct not-converged outcome)", got.OverallVerdict, VerdictInconclusive)
	}
	if got.GateUnmet == "" {
		t.Error("gate_unmet is empty — the unmet required gate must still be named")
	}
}

// TestEnforceRequiredGateUnmet_KeepsAReportedRequiredFail pins the l57-r2
// audit finding: a required backend that REPORTED fail is a reviewed failure,
// not a could-not-run — a second required leg left inconclusive must NOT
// overwrite the converged FAIL with the not-converged outcome (that would
// reclassify a found defect as never-run). The unmet leg is still named.
func TestEnforceRequiredGateUnmet_KeepsAReportedRequiredFail(t *testing.T) {
	gates := config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired}
	verdicts := []PerBackendVerdict{
		{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "fail"},
		{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
	}
	r := converge(verdicts)
	if r.OverallVerdict != overallVerdictFail {
		t.Fatalf("precondition: converge overall = %q, want fail (pinned by TestConverge_RequiredFailWithInconclusive_Case2)", r.OverallVerdict)
	}
	got := enforceRequiredGateUnmet(r, verdicts, gates)
	if got.OverallVerdict != overallVerdictFail {
		t.Errorf("overall = %q, want the reported FAIL kept — a required-inconclusive leg reclassified a found failure as not-run (l57-r2 finding 2)", got.OverallVerdict)
	}
	if !strings.Contains(got.GateUnmet, BackendCodex) {
		t.Errorf("gate_unmet = %q, want it to name the inconclusive codex leg", got.GateUnmet)
	}
}

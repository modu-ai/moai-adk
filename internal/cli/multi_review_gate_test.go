// Package cli — tests for the `moai hook multi-review-gate` Stop-hook logic
// (SPEC-AUDIT-MULTI-MODEL-001 M5, REQ-AMM-013 / REQ-AMM-014 / REQ-AMM-015 /
// AC-AMM-018 / AC-AMM-019 / AC-AMM-020 / AC-AMM-021).
//
// These tests pin the gate logic BEFORE the implementation exists (RED). The
// gate mirrors the codex-review-gate self-gate + opt-in + 900s-timeout pattern
// but consumes the pre-computed ConvergenceResult persisted by
// persistConvergenceResult (.moai/state/audit-multi/<session>.json, DQ-1)
// instead of re-invoking the convergence engine. Disagreement is advisory and
// NEVER blocks; the only BLOCK path is a required-backend FAIL (the gate's
// enforcement of REQ-AMM-006 #2/#3, surfaced by the convergence engine).
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// withMultiChangeDetector swaps the multi-review-gate change detector seam so
// the tests drive the self-gate deterministically without a real git repo.
func withMultiChangeDetector(t *testing.T, hasChanges bool) {
	t.Helper()
	prev := reviewGateChangeDetector
	reviewGateChangeDetector = func(string) bool { return hasChanges }
	t.Cleanup(func() { reviewGateChangeDetector = prev })
}

// writeMultiState writes a ConvergenceResult JSON to
// <tmpDir>/.moai/state/audit-multi/<session>.json (the DQ-1 path the gate
// reads). Returns the project root (tmpDir) so each test gets an isolated tree.
func writeMultiState(t *testing.T, result ConvergenceResult, sessionID string) string {
	t.Helper()
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".moai", "state", "audit-multi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir audit-multi state dir: %v", err)
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	path := filepath.Join(dir, sessionID+".json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}
	return tmp
}

// allPassResult is a clean ConvergenceResult: all required PASS, no disagreement.
func allPassResult() ConvergenceResult {
	return ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "pass"},
		},
		OverallVerdict:   overallVerdictPass,
		DisagreementFlag: boolPtr(false),
	}
}

// requiredFailResult: codex (required) FAILs while claude (required) PASSes.
// overall = fail (conservative — the required-gate contract holds per backend);
// disagreement_flag is set (a required split).
func requiredFailResult() ConvergenceResult {
	return ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "fail"},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "pass"},
		},
		OverallVerdict:   overallVerdictFail,
		DisagreementFlag: boolPtr(true),
		ResidualRiskNote: "required split",
	}
}

// advisoryOnlyFailResult: glm (advisory) FAILs while both required PASS — overall
// PASS, disagreement_flag set, advisory (MUST NOT block per AC-AMM-009).
func advisoryOnlyFailResult() ConvergenceResult {
	return ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: "fail"},
		},
		OverallVerdict:   overallVerdictPass,
		DisagreementFlag: boolPtr(true),
		ResidualRiskNote: "advisory-only conflict",
	}
}

// --- AC-AMM-018 / AC-AMM-019 ---

// TestMultiReviewGate_DisabledAllows proves the gate is opt-in: when the config
// toggle is off, the hook ALLOWs immediately without reading any state file
// (distributed default is OFF per the BranchGuard pattern).
func TestMultiReviewGate_DisabledAllows(t *testing.T) {
	withMultiChangeDetector(t, true) // even with changes present...
	tmp := writeMultiState(t, requiredFailResult(), "sess-disabled")
	// Even a required-FAIL state must NOT block when the gate is disabled.
	out, err := HandleMultiReviewGate(gateInput(false), false /* enabled */, tmp, "sess-disabled")
	if err != nil {
		t.Fatalf("disabled gate must not error; got %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("disabled gate must ALLOW (empty output); got %+v", out)
	}
}

// TestMultiReviewGate_LoopPreventionAllows proves stop_hook_active ALLOWs
// (mandatory CC loop-prevention protocol; never re-block an already-continuing
// turn — same heuristic as codex-review-gate).
func TestMultiReviewGate_LoopPreventionAllows(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := writeMultiState(t, requiredFailResult(), "sess-loop")

	out, _ := HandleMultiReviewGate(gateInput(true) /* stop_hook_active */, true /* enabled */, tmp, "sess-loop")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("stop_hook_active must ALLOW (loop prevention); got %+v", out)
	}
}

// TestMultiReviewGate_NoEditTurnAllows proves the self-gate (AC-AMM-019): a turn
// that produced no reviewable change (status report / review-result / no-op)
// ALLOWs immediately without reading the convergence state. No false block.
func TestMultiReviewGate_NoEditTurnAllows(t *testing.T) {
	withMultiChangeDetector(t, false) // clean working tree
	tmp := writeMultiState(t, requiredFailResult(), "sess-noedit")

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-noedit")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("no-edit turn must ALLOW (self-gate); got %+v", out)
	}
}

// --- AC-AMM-020 ---

// TestMultiReviewGate_AllRequiredPassAllows proves the gate's ALLOW contract: an
// edit turn where the most-recent ConvergenceResult has overall=pass (all
// required PASS) ALLOWs the session to end.
func TestMultiReviewGate_AllRequiredPassAllows(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := writeMultiState(t, allPassResult(), "sess-pass")

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-pass")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("all-required-PASS must ALLOW; got %+v", out)
	}
}

// TestMultiReviewGate_RequiredFailBlocks proves the gate's BLOCK contract: an
// edit turn where the convergence result has overall=fail (a required backend
// FAILed) BLOCKs the session end conservatively (AC-AMM-020 + AC-AMM-008).
func TestMultiReviewGate_RequiredFailBlocks(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := writeMultiState(t, requiredFailResult(), "sess-fail")

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-fail")
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Errorf("required-backend FAIL must BLOCK; got %+v", out)
	}
	if out != nil && out.Reason == "" {
		t.Error("BLOCK must carry a reason naming the failing backend(s)")
	}
}

// TestMultiReviewGate_AdvisoryDisagreementNeverBlocks proves the policy
// invariant (AC-AMM-009 / EC-3): an advisory-only backend FAIL with all required
// PASS yields overall=PASS + disagreement_flag=true. The gate MUST NOT block on
// advisory disagreement — the user-policy fixed term.
func TestMultiReviewGate_AdvisoryDisagreementNeverBlocks(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := writeMultiState(t, advisoryOnlyFailResult(), "sess-adv")

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-adv")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("advisory-only disagreement must NEVER block (AC-AMM-009); got %+v", out)
	}
}

// --- AC-AMM-021 (fail-open to claude when all non-Claude backends missing) ---

// TestMultiReviewGate_AllSecondariesInconclusive_FailOpenClaude proves that
// when every non-Claude backend is inconclusive (missing/unauthenticated), the
// gate falls back to the claude verdict — the autonomous flow is NEVER
// hard-blocked on a missing optional dependency (C2).
func TestMultiReviewGate_AllSecondariesInconclusive_FailOpenClaude(t *testing.T) {
	withMultiChangeDetector(t, true)
	allInconclusive := ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: VerdictInconclusive},
		},
		// overall pass (fail-open to claude — no required FAIL, claude anchor pass).
		OverallVerdict:   overallVerdictPass,
		FailOpenBackends: []string{BackendCodex, BackendGLM},
	}
	tmp := writeMultiState(t, allInconclusive, "sess-fo")

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-fo")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("all-secondaries-inconclusive (claude pass) must ALLOW (fail-open to claude); got %+v", out)
	}
}

// --- fail-open: missing state file, malformed JSON ---

// TestMultiReviewGate_MissingStateFileAllows proves the gate fail-OPENs when no
// state file exists yet (the convergence engine hasn't run for this session).
// A missing file is evidence-of-absence, not evidence-of-failure (C2).
func TestMultiReviewGate_MissingStateFileAllows(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := t.TempDir() // no audit-multi state written

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-nostate")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("missing state file must ALLOW (fail-open); got %+v", out)
	}
}

// TestMultiReviewGate_MalformedStateFileAllows proves a corrupted state file
// does not trap the session — fail-open ALLOW.
func TestMultiReviewGate_MalformedStateFileAllows(t *testing.T) {
	withMultiChangeDetector(t, true)
	tmp := t.TempDir()
	dir := filepath.Join(tmp, ".moai", "state", "audit-multi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sess-bad.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, _ := HandleMultiReviewGate(gateInput(false), true, tmp, "sess-bad")
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("malformed state file must ALLOW (fail-open); got %+v", out)
	}
}

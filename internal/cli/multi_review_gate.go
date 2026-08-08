// Package cli — `moai hook multi-review-gate` Stop-hook logic
// (SPEC-AUDIT-MULTI-MODEL-001 M5, REQ-AMM-013 / REQ-AMM-014 / REQ-AMM-015 /
// AC-AMM-018 / AC-AMM-019 / AC-AMM-020 / AC-AMM-021).
//
// multi_review_gate.go holds the PURE gate logic. The CLI subcommand wiring
// (stdin read → project-root resolve → config-gate read → this handler →
// stdout) lives in hook.go runMultiReviewGate. Keeping the logic pure + in the
// cli package lets it reuse the same reviewGateChangeDetector self-gate seam
// the codex-review-gate uses (AC-AMM-019 — the mandatory self-gate, same
// heuristic). It returns the standard ALLOW/BLOCK HookOutput and NEVER invokes
// AskUserQuestion (subagent boundary — the orchestrator translates a BLOCK).
//
// The gate consumes the ConvergenceResult persisted by persistConvergenceResult
// (mcp_convergence.go DQ-1) at .moai/state/audit-multi/<session>.json rather
// than re-invoking the convergence engine — the engine is the audit_multi MCP
// tool's job (M3); the gate reads its most-recent output. This separation is
// the Path C (fully-autonomous goal convergence) plumbing: the orchestrator
// arms a goal, the audit_multi tool produces the ConvergenceResult during the
// run, and this Stop hook blocks turn-end only if a required backend FAIL is
// unresolved in that result. Advisory disagreement NEVER blocks (AC-AMM-009).
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"


	"github.com/modu-ai/moai-adk/internal/hook"
)

// HandleMultiReviewGate is the Stop-hook gate logic. ALLOW/BLOCK contract:
//   - empty HookOutput ({})                            = ALLOW (Claude may stop)
//   - {Decision: "block", Reason: "..."}               = BLOCK (keep working)
//
// Decision order (AC-AMM-019 self-gate + AC-AMM-018 opt-in + REQ-AMM-015
// fail-open + AC-AMM-020 verdict policy + AC-AMM-009 advisory-never-blocks):
//  1. gate disabled (config off)             → ALLOW (opt-in default-off)
//  2. stop_hook_active (loop prevention)     → ALLOW (mandatory CC protocol)
//  3. no reviewable uncommitted change       → ALLOW (self-gate; no false block)
//  4. no session id / missing state file     → ALLOW (fail-open; no result yet)
//  5. malformed state file                   → ALLOW (fail-open; corrupt JSON)
//  6. overall_verdict = pass                 → ALLOW (all required PASS, OR
//                                              advisory-only conflict — never block)
//  7. overall_verdict = fail (required FAIL) → BLOCK (the gate's ONLY block path)
//
// `enabled` is read by the caller (runMultiReviewGate via
// readMultiReviewGateEnabled) and passed in so this function stays free of
// config I/O (testable as pure logic). It is re-checked here as defense-in-depth.
//
// The gate does NOT distinguish "fail because all required agree on fail" from
// "fail because of a required split" — both surface as overall=fail in the
// ConvergenceResult (the convergence engine resolves them identically per
// REQ-AMM-006 #2/#3). The residual_risk_note in the result already names which
// backend(s) failed; this gate's BLOCK reason echoes that note so the operator
// sees the same trail at the Stop surface.
func HandleMultiReviewGate(input *hook.HookInput, enabled bool, projectDir, sessionID string) (*hook.HookOutput, error) {
	allow := &hook.HookOutput{}
	if !enabled {
		return allow, nil // (1) opt-in default-off
	}
	if input != nil && input.StopHookActive {
		return allow, nil // (2) loop prevention
	}
	if !reviewGateChangeDetector(projectDir) {
		return allow, nil // (3) self-gate — nothing reviewable ⇒ no false block
	}

	result, ok := loadConvergenceResult(projectDir, sessionID)
	if !ok {
		return allow, nil // (4)+(5) fail-open: missing/malformed state ⇒ ALLOW
	}

	// (6)+(7): the convergence engine already derived overall_verdict per the
	// REQ-AMM-006 policy table (advisory-never-blocks baked in at AC-AMM-009).
	// A required FAIL produces overall=fail; advisory-only conflict yields
	// overall=pass + disagreement_flag=true. The gate trusts that derivation.
	if result.OverallVerdict == overallVerdictFail {
		return &hook.HookOutput{
			Decision: hook.DecisionBlock, // (7) the gate's ONLY BLOCK path
			Reason:   blockReason(result),
		}, nil
	}
	return allow, nil // pass / advisory conflict / fail-open to claude ⇒ ALLOW
}

// blockReason builds the BLOCK decision's reason from the ConvergenceResult.
// Echoes the residual_risk_note (which already names which backend(s) failed)
// so the operator sees one consistent trail across the convergence result and
// the Stop surface.
func blockReason(r ConvergenceResult) string {
	note := strings.TrimSpace(r.ResidualRiskNote)
	if note == "" {
		note = "required-backend FAIL (see per_backend_verdicts for details)"
	}
	return "multi-review-gate: " + note
}

// loadConvergenceResult reads + decodes the ConvergenceResult JSON at
// <projectDir>/.moai/state/audit-multi/<sessionID>.json (the DQ-1 path written
// by persistConvergenceResult). Returns (zero, false) on any error path
// (missing file, malformed JSON, empty session id) so the caller fail-OPENs.
// The fail-open direction is load-bearing: a missing optional backend's
// evidence-of-absence must NOT trap the session.
func loadConvergenceResult(projectDir, sessionID string) (ConvergenceResult, bool) {
	if projectDir == "" || sessionID == "" {
		return ConvergenceResult{}, false
	}
	path := filepath.Join(projectDir, ".moai", "state", "audit-multi", sessionID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return ConvergenceResult{}, false
	}
	var r ConvergenceResult
	if err := json.Unmarshal(b, &r); err != nil {
		return ConvergenceResult{}, false
	}
	return r, true
}

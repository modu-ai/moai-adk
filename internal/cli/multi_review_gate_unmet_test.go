package cli

// Card t1632, sync-audit r9 delta F1: REQ-MN-021 turned "a required gate that
// could not run" from overall=fail into overall=inconclusive carrying
// gate_unmet. The multi-review Stop gate (Claude hook and Codex chain member 7)
// blocked only on overall=fail, so the unmet required gate stopped blocking the
// turn end. Both consumers must block on a required-gate-unmet result, and
// must still allow an inconclusive result that no explicit gate made required.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

func requiredGateUnmetResult() ConvergenceResult {
	return ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendCodex, Gate: config.AuditGateRequired, Verdict: VerdictInconclusive},
		},
		OverallVerdict:   VerdictInconclusive,
		GateUnmet:        BackendCodex,
		ResidualRiskNote: "required gate unmet (explicitly configured required, could-not-run, no verdict): codex",
		FailOpenBackends: []string{BackendCodex},
	}
}

func inconclusiveNoGateResult() ConvergenceResult {
	return ConvergenceResult{
		PerBackendVerdicts: []PerBackendVerdict{
			{Backend: BackendClaude, Gate: config.AuditGateRequired, Verdict: "pass"},
			{Backend: BackendGLM, Gate: config.AuditGateAdvisory, Verdict: VerdictInconclusive},
		},
		OverallVerdict:   overallVerdictPass,
		FailOpenBackends: []string{BackendGLM},
	}
}

func TestMultiReviewGate_RequiredGateUnmetBlocks(t *testing.T) {
	withMultiChangeDetector(t, true)

	blocked := writeMultiState(t, requiredGateUnmetResult(), "sess-unmet")
	out, err := HandleMultiReviewGate(gateInput(false), true, blocked, "sess-unmet")
	if err != nil {
		t.Fatalf("HandleMultiReviewGate: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Fatalf("a required gate left unmet (overall=inconclusive, gate_unmet=codex) must block the stop, got %+v", out)
	}

	// Control: an inconclusive advisory backend that no explicit gate made
	// required stays a fail-open allow — the fix must not over-block.
	allowed := writeMultiState(t, inconclusiveNoGateResult(), "sess-adv-inc")
	out, err = HandleMultiReviewGate(gateInput(false), true, allowed, "sess-adv-inc")
	if err != nil {
		t.Fatalf("HandleMultiReviewGate: %v", err)
	}
	if out != nil && out.Decision == hook.DecisionBlock {
		t.Fatalf("an inconclusive advisory backend must not block, got %+v", out)
	}
}

func TestCodexStopChain_MultiReviewMemberBlocksOnRequiredGateUnmet(t *testing.T) {
	withMultiChangeDetector(t, true)
	root := writeMultiState(t, requiredGateUnmetResult(), "sess-codex-unmet")
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte("workflow:\n  multi:\n    review_gate:\n      enabled: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := newCodexStopChain(root, stopInput("sess-codex-unmet", false)).multiReviewMember(context.Background())
	if got.Class != reasonGateFailed {
		t.Fatalf("the Codex multi-review member must deny on a required gate left unmet, got decision=%v class=%v status=%v", got.Decision, got.Class, got.Status)
	}
}

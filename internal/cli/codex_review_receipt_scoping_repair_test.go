package cli

// SPEC-CODEX-GATE-SCOPING-001 — the card-review repair round (card t1404),
// finding R2: the runtime-drift reclassification was wired into the Claude
// Stop path only, while the receipt producer recorded a fail the Codex Stop
// chain's member 6 DENYs on for the same config-only findings. REQ-CGSC-008's
// intent is BOTH automatic arms taking the one reclassification decision
// (decision-index Q3: both arms kept).
//
// This file sits beside the producer it repairs so the repair's two commits
// stay independently compiling: the gate-surface findings (R1/R3/R4/R5) land
// first, this producer arm second.

import (
	"strings"
	"testing"
)

// TestProduceCodexReviewReceipt_TreeDriftFindingsReclassified is the R2 repro:
// the receipt producer rides the same reclassification decision the Claude
// Stop path takes (codex_review_gate.go step 7-pre) — a tree-scope review
// whose every finding targets only the runtime-config surfaces records the
// gate's outcome (pass) and the reclassification row, never a fail receipt
// the Codex Stop chain would DENY on.
func TestProduceCodexReviewReceipt_TreeDriftFindingsReclassified(t *testing.T) {
	f := newStopFixture(t)
	f.dirty(t, "reviewable")
	fakeCodexVersion(t, "codex-cli 0.0.0-repair")
	withCodexSession(t, codexSessionScript(runtimeDriftReviewText))

	read := captureGateDiagnostics(t)
	got, err := runVerifyCodexReview(t, f.root)
	if err != nil {
		t.Fatalf("producer: %v", err)
	}
	diagnostics := read()
	if got["verdict"] != codexReviewVerdictPass {
		t.Fatalf("a config-only-findings fail must be reclassified on the receipt arm too, got verdict %v (%v)", got["verdict"], got)
	}
	if !strings.Contains(diagnostics, "runtime-managed") {
		t.Errorf("the reclassification must be recorded on the diagnostic channel (REQ-CGSC-011); diagnostics: %q", diagnostics)
	}
}

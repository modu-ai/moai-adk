// audit_multi_bind_timing_test.go — card t1632 sync-audit r3 repair F13: the
// audit_multi receipt binds the tree state captured at CONVERGENCE START, not
// at write time — the exact F5 class on the convergence surface.
package cli

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// TestAuditMulti_ReceiptBindsStartOfConvergenceState pins the F13 fix: the
// tree state MOVES WHILE THE BACKENDS RUN (the fan-out stub flips a flag),
// and the receipt must carry the state captured BEFORE the fan-out — the
// state the convergence actually judged. A write-time bind records the
// post-edit state (the F13 laundering path).
func TestAuditMulti_ReceiptBindsStartOfConvergenceState(t *testing.T) {
	root := newProbeProject(t, "SPEC-BINDTIMING-952")
	seedWorkflowGates(t, root, map[string]string{"codex": config.AuditGateRequired})

	// The backends run, THEN the tree state moves: the fan-out stub flips the
	// flag its own invocation completed. Atomic (card t1632 sync-audit r4 F16):
	// the backends run on goroutines of the errgroup, so the flag is written
	// there and read in TreeState concurrently.
	var postEdit atomic.Bool
	rc := &recordingCallerMulti{verdictBy: map[string]ReviewOutput{BackendCodex: inconclusiveVerdict("codex unavailable")}}
	orig := backendCall
	backendCall = func(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput {
		out := rc.call(ctx, backend, target, focus, projectRoot)
		postEdit.Store(true)
		return out
	}
	t.Cleanup(func() { backendCall = orig })

	prev := auditreceipt.TreeState
	auditreceipt.TreeState = func(treeRoot string) (string, string, error) {
		if postEdit.Load() {
			return "post-edit-head", "post-edit-digest", nil
		}
		return "start-head", "start-digest", nil
	}
	t.Cleanup(func() { auditreceipt.TreeState = prev })

	withCodexSession(t, codexSessionScript(realCleanReview))
	res := runMultiAudit(context.Background(), claudePassAnchor(), "uncommittedChanges", "", MultiAuditConfig{
		Gates:       auditGatesDefault(),
		SessionID:   "bind-timing-session",
		ProjectRoot: root,
	}, nil)
	if res.AuditReceipt == "" {
		t.Fatal("the required codex gate exposed no receipt id — the bind-timing test needs the receipt")
	}
	rec, err := auditreceipt.ReadReceipt(auditreceipt.Canonical(root), res.AuditReceipt)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if rec.HeadCommit != "start-head" || rec.TreeDigest != "start-digest" {
		t.Errorf("receipt binds head=%q digest=%q, want the START state — a write-time bind on the convergence surface is the F13 laundering path", rec.HeadCommit, rec.TreeDigest)
	}
}

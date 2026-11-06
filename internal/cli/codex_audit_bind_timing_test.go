// codex_audit_bind_timing_test.go — card t1632 sync-audit repair F5: the
// audit receipt binds the tree state captured at AUDIT START, not at write
// time — a tree edit during a minutes-long audit must not make the receipt
// attest the post-edit state (that is the PASS laundering hole).
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

// TestCodexAudit_ReceiptBindsStartOfAuditState pins the F5 fix: with the tree
// state MOVING between the audit start and the receipt write, the recorded
// receipt carries the START state — the state the audit actually judged.
func TestCodexAudit_ReceiptBindsStartOfAuditState(t *testing.T) {
	root := newProbeProject(t, "SPEC-BINDTIMING-951")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript(realCleanReview))

	// The tree state MOVES once, between the audit's start capture and the
	// receipt write: the first TreeState read (start capture) returns state A,
	// every later read (the write-time bind) returns state B.
	calls := 0
	prev := auditreceipt.TreeState
	auditreceipt.TreeState = func(treeRoot string) (string, string, error) {
		calls++
		if calls == 1 {
			return "start-head", "start-digest", nil
		}
		return "post-edit-head", "post-edit-digest", nil
	}
	t.Cleanup(func() { auditreceipt.TreeState = prev })

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	rid, _ := got["audit_receipt"].(string)
	if rid == "" {
		t.Fatal("the required gate exposed no receipt id — the bind-timing test needs the receipt")
	}
	rec, err := auditreceipt.ReadReceipt(auditreceipt.Canonical(root), rid)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if rec.HeadCommit != "start-head" || rec.TreeDigest != "start-digest" {
		t.Errorf("receipt binds head=%q digest=%q, want the START state (start-head/start-digest) — a post-audit bind launders a PASS across a tree edit (sync-audit F5)", rec.HeadCommit, rec.TreeDigest)
	}
	if calls != 1 {
		// The FIX removes the write-time read entirely: the only TreeState
		// read is the start capture. A second read would mean the receipt is
		// bound at write time again (the F5 defect).
		t.Errorf("TreeState reads = %d, want exactly 1 (the start capture) — a write-time bind would launder a PASS across a tree edit (sync-audit F5)", calls)
	}
}

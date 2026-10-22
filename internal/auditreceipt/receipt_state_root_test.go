// receipt_state_root_test.go — card t1632 card-review repair P1-2: the reuse
// check compares the receipt's binding against the state of the tree the
// receipt JUDGED (r.TreeRoot), never the store root the caller handed in — a
// hook in a linked worktree passes the primary checkout as the store root,
// and the receipt was recorded against the worktree.
package auditreceipt

import (
	"testing"
	"time"
)

// TestCheckCitedReceipts_StateComparesTheJudgedTree pins the P1-2 fix: the
// TreeState read for the freshness comparison runs against r.TreeRoot (the
// audited tree), not the caller's store root.
func TestCheckCitedReceipts_StateComparesTheJudgedTree(t *testing.T) {
	storeRoot := t.TempDir()     // what a hook in a linked worktree hands in
	judged := "/linked-worktree" // what the receipt actually recorded

	var readFor string
	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) {
		readFor = treeRoot
		return "head-1", "digest-1", nil
	}
	t.Cleanup(func() { TreeState = prev })

	start := testStart(judged, time.Now().Add(-time.Minute))
	id, err := WriteReceipt(storeRoot, &Receipt{
		Tool: ToolCodexAudit, TreeRoot: judged, CreatedAt: time.Now(), CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}

	// The receipt must QUALIFY: the judged tree's current state matches its
	// binding (the stub returns the same values for whatever root it reads).
	ok, cause := CheckCitedReceiptsSince(storeRoot, start, time.Time{}, []string{id})
	if !ok {
		t.Fatalf("a matching-state receipt was refused: %s", cause)
	}
	// P1-2: the comparison read the JUDGED tree, not the store root.
	if readFor != judged {
		t.Errorf("TreeState read for %q, want %q — the state comparison must run on the tree the receipt judged (card-review P1-2)", readFor, judged)
	}
}

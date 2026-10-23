// receipt_integrity_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-020): an audit receipt binds the head commit and a tree digest of
// the state it judged, a receipt whose verdict is FAIL never corroborates a
// PASS, and a receipt whose head/digest differs from the current tree is not
// reusable as evidence.
package auditreceipt

import (
	"testing"
	"time"
)

// stubTreeState installs a deterministic tree-state seam: HEAD and digest the
// tests control, never a real git read.
func stubTreeState(t *testing.T, head, digest string) {
	t.Helper()
	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) { return head, digest, nil }
	t.Cleanup(func() { TreeState = prev })
}

func testStart(treeRoot string, at time.Time) *StartMarker {
	return &StartMarker{AgentID: "agent-a", TreeRoot: treeRoot, StartedAt: at}
}

// TestReceipt_CarriesHeadAndTreeDigest pins the binding: a receipt written for
// a tree carries the head commit and the tree digest of the state it judged.
func TestReceipt_CarriesHeadAndTreeDigest(t *testing.T) {
	root := t.TempDir()
	stubTreeState(t, "abc123def456", "digest-77")

	r := Receipt{Tool: ToolCodexAudit, TreeRoot: root}
	id, err := WriteReceipt(root, &r)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	got, err := ReadReceipt(root, id)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if got.HeadCommit != "abc123def456" {
		t.Errorf("head_commit = %q, want the judged head commit", got.HeadCommit)
	}
	if got.TreeDigest != "digest-77" {
		t.Errorf("tree_digest = %q, want the judged tree digest", got.TreeDigest)
	}
}

// TestCheckCitedReceipts_RefusesFailVerdict pins the contradiction refusal: a
// receipt whose codex verdict is FAIL does not corroborate a PASS — the audit
// ran and found a blocking failure; a PASS citing it is a contradiction.
func TestCheckCitedReceipts_RefusesFailVerdict(t *testing.T) {
	root := t.TempDir()
	stubTreeState(t, "head-1", "digest-1")
	start := testStart(root, time.Now().Add(-time.Minute))
	id, err := WriteReceipt(root, &Receipt{
		Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: time.Now(), CodexVerdict: "fail",
	})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}

	ok, cause := CheckCitedReceipts(root, start, []string{id})
	if ok {
		t.Fatal("a FAIL-verdict receipt corroborated a PASS — REQ-MN-020 refuses the contradiction")
	}
	if cause != CauseReceiptVerdictFail {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptVerdictFail)
	}
}

// TestCheckCitedReceipts_RefusesTreeStateChanged pins the freshness binding: a
// receipt whose head or tree digest differs from the tree's CURRENT state is
// not reusable evidence for that state.
func TestCheckCitedReceipts_RefusesTreeStateChanged(t *testing.T) {
	root := t.TempDir()
	stubTreeState(t, "head-1", "digest-1")
	start := testStart(root, time.Now().Add(-time.Minute))
	id, err := WriteReceipt(root, &Receipt{
		Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: time.Now(), CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}

	// The tree moved on after the audit.
	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) { return "head-2", "digest-2", nil }
	t.Cleanup(func() { TreeState = prev })

	ok, cause := CheckCitedReceipts(root, start, []string{id})
	if ok {
		t.Fatal("a receipt for a changed tree corroborated the current state — REQ-MN-020 refuses the reuse")
	}
	if cause != CauseReceiptTreeStateChanged {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptTreeStateChanged)
	}
}

// TestCheckCitedReceipts_MatchingStateStillQualifies pins the positive space:
// a receipt whose head/digest matches the current state corroborates as
// before — the binding adds a refusal arm, not a blanket refusal.
func TestCheckCitedReceipts_MatchingStateStillQualifies(t *testing.T) {
	root := t.TempDir()
	stubTreeState(t, "head-1", "digest-1")
	start := testStart(root, time.Now().Add(-time.Minute))
	id, err := WriteReceipt(root, &Receipt{
		Tool: ToolCodexAudit, TreeRoot: root, CreatedAt: time.Now(), CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}

	ok, cause := CheckCitedReceipts(root, start, []string{id})
	if !ok {
		t.Fatalf("a matching-state receipt was refused: %s", cause)
	}
}

// codex_audit_capture_failure_test.go — card t1632 sync-audit r3 repair F14:
// a START capture that FAILS must not fall back to the write-time bind — the
// fallback re-lands the F5 laundering through the error path. The receipt
// records an explicit capture-failed marker and the reuse check refuses it.
package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
)

func timeNowForTest() time.Time { return auditreceipt.Now() }

// startMarkerFor builds the start marker the citation check needs, judged in
// the same tree as the receipt.
func startMarkerFor(treeRoot string) *auditreceipt.StartMarker {
	return &auditreceipt.StartMarker{AgentID: "agent-capture", TreeRoot: treeRoot, StartedAt: auditreceipt.Now().Add(-time.Minute)}
}

// TestCodexAudit_CaptureFailureNeverFallsBackToWriteTimeBind pins the F14
// fix: with the start capture FAILING and the tree moving during the audit,
// the receipt records the capture-failed marker — it never carries a
// write-time binding the audit never judged.
func TestCodexAudit_CaptureFailureNeverFallsBackToWriteTimeBind(t *testing.T) {
	root := newProbeProject(t, "SPEC-BINDFAIL-953")
	writeCodexAuditGate(t, root, config.AuditGateRequired)
	withCodexSession(t, codexSessionScript(realCleanReview))

	prev := auditreceipt.TreeState
	auditreceipt.TreeState = func(treeRoot string) (string, string, error) {
		// The start capture fails (a broken symlink, a wedged git); every
		// later read would succeed — the fallback would then bind the
		// post-audit state, the exact laundering F14 names.
		return "", "", errors.New("broken symlink at capture time")
	}
	t.Cleanup(func() { auditreceipt.TreeState = prev })

	got := structuredMap(t, callToolCodexAudit(t, map[string]any{"project_root": root}))
	rid, _ := got["audit_receipt"].(string)
	if rid == "" {
		t.Fatal("the required gate exposed no receipt id — the capture-failure test needs the receipt")
	}
	rec, err := auditreceipt.ReadReceipt(auditreceipt.Canonical(root), rid)
	if err != nil {
		t.Fatalf("ReadReceipt: %v", err)
	}
	if rec.HeadCommit != "" || rec.TreeDigest != "" {
		t.Errorf("receipt carries a binding (%s/%s) despite the start capture failing — the write-time fallback re-lands the F5 laundering (sync-audit r3 F14)", rec.HeadCommit, rec.TreeDigest)
	}
	if !rec.BindCaptureFailed {
		t.Error("bind_capture_failed = false, want true — the capture failure must be an explicit marker, never silent")
	}
}

// TestCheckCitedReceipts_CaptureFailedMarkerRefused pins the consumer half: a
// receipt whose binding capture failed is NOT reusable as PASS evidence — a
// binding whose freshness is unknowable proves nothing.
func TestCheckCitedReceipts_CaptureFailedMarkerRefused(t *testing.T) {
	root := t.TempDir()
	prev := auditreceipt.TreeState
	auditreceipt.TreeState = func(treeRoot string) (string, string, error) {
		return "head-1", "digest-1", nil
	}
	t.Cleanup(func() { auditreceipt.TreeState = prev })

	r := auditreceipt.Receipt{
		Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: timeNowForTest(),
		CodexVerdict: "pass", BindCaptureFailed: true,
	}
	id, err := auditreceipt.WriteReceipt(root, &r)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	start := startMarkerFor(root)
	ok, cause := auditreceipt.CheckCitedReceipts(root, start, []string{id})
	if ok {
		t.Fatalf("a capture-failed receipt corroborated a PASS (cause=%s) — its freshness is unknowable (sync-audit r3 F14)", cause)
	}
}

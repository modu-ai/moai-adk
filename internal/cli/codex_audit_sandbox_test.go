// codex_audit_sandbox_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-024): every codex app-server the audit/review/convergence legs spawn
// runs under a read-only sandbox policy and a scrubbed child environment —
// the reviewer never needs to write, and the child inherits neither the
// session's identity keys nor the lane's launch keys.
package cli

import (
	"context"
	"strings"
	"testing"
)

// fakeCodexSessionArgs records the argv the session start received, so a test
// can pin the spawn contract (REQ-MN-024).
func withCodexSessionArgs(t *testing.T, lines []string) (*fakeCodexSession, *[]string) {
	t.Helper()
	sess := withCodexSession(t, lines)
	return sess, &sess.startArgs
}

// TestCodexAudit_SessionSpawnsUnderReadOnlySandbox pins the sandbox policy:
// the app-server argv carries the read-only sandbox override.
func TestCodexAudit_SessionSpawnsUnderReadOnlySandbox(t *testing.T) {
	root := newProbeProject(t, "SPEC-SANDBOX-941")
	_, args := withCodexSessionArgs(t, codexSessionScript(realCleanReview))

	_ = callToolCodexAudit(t, map[string]any{"project_root": root})
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "sandbox_mode=read-only") {
		t.Errorf("session argv = %v, want the read-only sandbox override (REQ-MN-024)", *args)
	}
}

// TestCodexAudit_NoProjectRootStillContainsChild (card t1632 sync-audit r4):
// an audit leg that names NO project_root (the audit_multi codex leg's
// default) must STILL spawn a contained child. Before the repair the
// containment block keyed on params["cwd"], so an omitted root skipped the
// foreign-server disable entirely — the child resolved the operator's own MCP
// servers, write tools included. The listing here carries one foreign server,
// so the argv must name it whatever the request carried.
func TestCodexAudit_NoProjectRootStillContainsChild(t *testing.T) {
	_, args := withCodexSessionArgs(t, codexSessionScript(realCleanReview))
	prevList := codexAuditMCPList
	codexAuditMCPList = func(context.Context, string, string) (string, error) {
		return `[{"name":"context7","transport":{"type":"streamable_http","url":"https://mcp.context7.example"}}]`, nil
	}
	t.Cleanup(func() { codexAuditMCPList = prevList })

	_ = performCodexAudit(context.Background(), codexTargetUncommitted, "", "")
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "mcp_servers.context7") {
		t.Errorf("no-root audit child argv = %v, want the foreign-server disable (REQ-MN-009 containment is unconditional)", *args)
	}
}

// TestCodexReview_SessionSpawnsUnderReadOnlySandbox pins the same policy on
// the review gate's session path (the Stop hook review leg).
func TestCodexReview_SessionSpawnsUnderReadOnlySandbox(t *testing.T) {
	root := newProbeProject(t, "SPEC-SANDBOX-942")
	_, args := withCodexSessionArgs(t, codexSessionScript(realCleanReview))

	_, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("produceCodexReviewReceipt: %v", err)
	}
	joined := strings.Join(*args, " ")
	if !strings.Contains(joined, "sandbox_mode=read-only") {
		t.Errorf("review session argv = %v, want the read-only sandbox override (REQ-MN-024)", *args)
	}
}

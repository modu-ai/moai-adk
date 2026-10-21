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
		t.Errorf("no-root audit child argv = %v, want the foreign-server disable (REQ-MN-009 containment is unconditional ON THE GATE/AUDIT PATH — l57-r8 finding 1 scoped it there; ordinary codex_task spawns stay neutral)", *args)
	}
}

// TestCodexAudit_SessionChildCwdMatchesListingRoot (card t1632 sync-audit
// l57-r5 finding 1): the app-server child must run AT the root the containment
// listing read. The disable list names the MCP servers configured at that
// root; a child that inherits the SERVER process's cwd (A) while the request
// named a different root (B) resolves A's config.toml instead, and a foreign
// server registered only in A escapes the disable list and reaches the child.
func TestCodexAudit_SessionChildCwdMatchesListingRoot(t *testing.T) {
	root := newProbeProject(t, "SPEC-CHILD-CWD-943")
	sess, _ := withCodexSessionArgs(t, codexSessionScript(realCleanReview))

	_ = callToolCodexAudit(t, map[string]any{"project_root": root})
	if sess.startDir != root {
		t.Errorf("session child dir = %q, want the containment listing root %q — a child at another cwd resolves a different config.toml and escapes the disable list (l57-r5 finding 1)", sess.startDir, root)
	}
}

// TestCodexTask_SessionChildNotContained pins the l57-r8 finding 1 split:
// containment belongs to GATE/AUDIT sessions only. An ordinary codex_task
// delegation rides the same session-open path, and the r4 round's
// unconditional containment leaked the foreign-server disable (and the
// session-level read-only sandbox) into EVERY task — the child lost all of
// the operator's configured MCP servers regardless of allow_write. The task
// child's argv carries neither; the task's own per-turn sandboxPolicy
// governs its sandbox.
func TestCodexTask_SessionChildNotContained(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	_, args := withCodexSessionArgs(t, codexTaskScript("trn-nogate", "task output"))
	// A foreign server the operator configured: the stub makes the leak
	// visible in the argv whenever the task path still runs the containment.
	prevList := codexAuditMCPList
	codexAuditMCPList = func(context.Context, string, string) (string, error) {
		return `[{"name":"context7","transport":{"type":"streamable_http","url":"https://mcp.context7.example"}}]`, nil
	}
	t.Cleanup(func() { codexAuditMCPList = prevList })

	res := callCodexTask(t, map[string]any{"prompt": "refactor the parser"})
	if res.IsError {
		t.Fatalf("unexpected IsError result: %+v", res)
	}
	joined := strings.Join(*args, " ")
	if strings.Contains(joined, "mcp_servers.context7") {
		t.Errorf("ordinary codex_task child argv = %v, want NO foreign-server disable args — the containment belongs to gate/audit spawns (l57-r8 finding 1)", *args)
	}
	if strings.Contains(joined, "sandbox_mode=read-only") {
		t.Errorf("ordinary codex_task child argv = %v, want NO session-level read-only sandbox — the task's per-turn sandboxPolicy governs it (l57-r8 finding 1)", *args)
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

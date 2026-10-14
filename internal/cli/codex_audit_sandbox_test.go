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

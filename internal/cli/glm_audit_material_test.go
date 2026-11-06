// glm_audit_material_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-024): the audit result marks truncated material and excluded
// untracked files, and an empty material never invokes the backend.
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGLMAudit_MarksTruncatedMaterial pins the truncation mark: a change
// bigger than the review size limit sets truncated on the result.
func TestGLMAudit_MarksTruncatedMaterial(t *testing.T) {
	root := newGLMReviewTree(t, false)
	// A tracked change far beyond reviewDiffMaxBytes: committed first, then
	// rewritten, so `git diff HEAD` carries it.
	big := make([]byte, reviewDiffMaxBytes+4096)
	for i := range big {
		big[i] = 'x'
	}
	path := filepath.Join(root, "big.go")
	if err := os.WriteFile(path, []byte("package p\n"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	cardScopeGit(t, root, "add", "-A")
	cardScopeGit(t, root, "commit", "-q", "-m", "seed big file")
	if err := os.WriteFile(path, big, 0o644); err != nil {
		t.Fatalf("write big file: %v", err)
	}
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Findings: []Finding{}, NextSteps: []string{}})}
	withGLMSeams(t, "stub-key", stub)

	got := structuredMap(t, callToolGLMAudit(t, map[string]any{"project_root": root}))
	if tr, _ := got["truncated"].(bool); !tr {
		t.Errorf("truncated = %v, want true — the material exceeded the limit (REQ-MN-024)", got["truncated"])
	}
}

// TestGLMAudit_MarksExcludedUntracked pins the untracked mark: untracked
// files the diff cannot carry are named on the result.
func TestGLMAudit_MarksExcludedUntracked(t *testing.T) {
	root := newGLMReviewTree(t, true)
	if err := os.WriteFile(filepath.Join(root, "untracked_secret.go"), []byte("package p\n"), 0o644); err != nil {
		t.Fatalf("write untracked: %v", err)
	}
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Findings: []Finding{}, NextSteps: []string{}})}
	withGLMSeams(t, "stub-key", stub)

	got := structuredMap(t, callToolGLMAudit(t, map[string]any{"project_root": root}))
	list, _ := got["excluded_untracked"].([]any)
	found := false
	for _, e := range list {
		if s, _ := e.(string); s == "untracked_secret.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("excluded_untracked = %v, want it to name untracked_secret.go (REQ-MN-024)", list)
	}
}

// TestGLMAudit_EmptyMaterialNeverInvokesBackend pins the empty-material guard:
// a clean tree produces an inconclusive with ZERO backend calls.
func TestGLMAudit_EmptyMaterialNeverInvokesBackend(t *testing.T) {
	root := newGLMReviewTree(t, false) // committed clean
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Findings: []Finding{}, NextSteps: []string{}})}
	withGLMSeams(t, "stub-key", stub)

	got := structuredMap(t, callToolGLMAudit(t, map[string]any{"project_root": root}))
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want inconclusive on empty material", v)
	}
	if stub.calls != 0 {
		t.Errorf("backend invoked %d time(s) on empty material — the guard must call neither backend (REQ-MN-024)", stub.calls)
	}
}

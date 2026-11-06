// receipt_state_error_test.go — card t1632 sync-audit repair F3: a state read
// that ERRORS refuses the reuse — the freshness check must not silently
// vanish exactly when the tree state cannot be judged. Only the narrow
// absent-git carve-out (the command itself is unavailable) leaves the check
// undecided, and a Korean filename survives the state read end-to-end.
package auditreceipt

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupCitedReceipt(t *testing.T, judged string) (storeRoot, id string, start *StartMarker) {
	t.Helper()
	storeRoot = t.TempDir()
	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) {
		return "head-1", "digest-1", nil
	}
	t.Cleanup(func() { TreeState = prev })
	start = testStart(judged, time.Now().Add(-time.Minute))
	id, err := WriteReceipt(storeRoot, &Receipt{
		Tool: ToolCodexAudit, TreeRoot: judged, CreatedAt: time.Now(), CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	return storeRoot, id, start
}

// TestCheckCitedReceipts_StateReadErrorRefuses pins the F3 fix: a state read
// that errors (a Korean filename that broke the old quoting parser, any other
// read failure) REFUSES the reuse instead of accepting the receipt.
func TestCheckCitedReceipts_StateReadErrorRefuses(t *testing.T) {
	judged := t.TempDir()
	storeRoot, id, start := setupCitedReceipt(t, judged)

	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) {
		return "", "", errors.New("status parse failed")
	}
	t.Cleanup(func() { TreeState = prev })

	ok, cause := CheckCitedReceipts(storeRoot, start, []string{id})
	if ok {
		t.Fatal("a receipt was ACCEPTED while the tree state could not be read — the freshness check vanished (card-review sync-audit F3)")
	}
	if cause != CauseReceiptStateUnreadable {
		t.Errorf("cause = %q, want %q", cause, CauseReceiptStateUnreadable)
	}
}

// TestCheckCitedReceipts_AbsentGitCarveOut pins the narrow carve-out: the git
// command being ABSENT (exec.ErrNotFound) leaves the check undecided rather
// than refusing — the same permissive reading the old code gave every failure,
// narrowed to the one case that cannot be a state change.
func TestCheckCitedReceipts_AbsentGitCarveOut(t *testing.T) {
	judged := t.TempDir()
	storeRoot, id, start := setupCitedReceipt(t, judged)

	prev := TreeState
	TreeState = func(treeRoot string) (string, string, error) {
		// The real shape: gitIn wraps the exec error with %w, so the chain
		// carries exec.ErrNotFound when the git command is absent.
		return "", "", fmt.Errorf(`git status in /x: %w`, exec.ErrNotFound)
	}
	t.Cleanup(func() { TreeState = prev })

	ok, cause := CheckCitedReceipts(storeRoot, start, []string{id})
	if !ok {
		t.Fatalf("an absent-git state read refused the receipt (cause=%s) — the carve-out should keep the permissive reading", cause)
	}
}

// TestTreeState_KoreanFilenameSurvives pins the -z parse end-to-end: an
// untracked file with a Korean name is read (no C-quoted misparse), so the
// state read succeeds and the file's content moves the digest.
func TestTreeState_KoreanFilenameSurvives(t *testing.T) {
	root := initContentRepo(t)
	korean := filepath.Join(root, "설계서-초안.txt")
	if err := os.WriteFile(korean, []byte("초안 v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest1, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState with a Korean filename: %v", err)
	}
	if err := os.WriteFile(korean, []byte("초안 v2 — 수정\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, digest2, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState after the edit: %v", err)
	}
	if digest1 == digest2 {
		t.Fatal("digest unchanged after a Korean-named file's content changed — the -z parse did not reach the file")
	}
}

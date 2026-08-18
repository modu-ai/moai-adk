// f1_traversal_test.go — RED probe for review finding F1 (specID path
// traversal, HIGH/security): a traversal-shaped specID must be REFUSED by the
// board read/persistence surface and yield no out-of-tree read, routed
// through the repo's shared sanitizer internal/cli/specid.ValidateSpecID.
package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadCardStatus_RejectsTraversalSpecID — F1: a specID carrying `..`, a
// path separator, or an absolute path is refused at the board read entry
// point, never interpolated into a worktree path or a git-show ref.
func TestReadCardStatus_RejectsTraversalSpecID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bases := WorktreeBases{Claude: filepath.Join(root, "wt"), MoAI: filepath.Join(root, "moai")}
	for _, bad := range []string{
		"../sibling",
		"..",
		"foo/../bar",
		"foo/bar",
		`foo\bar`,
		"/etc/passwd",
	} {
		if _, err := ReadCardStatus(root, bad, bases); err == nil {
			t.Errorf("ReadCardStatus(%q) err = nil, want refusal — a traversal specID must not reach a path or ref join", bad)
		}
	}
}

// TestTransitionIntoRun_RejectsTraversalSpecID — F1's persistence half: a
// traversal value is not persistable. The write is refused before any file is
// touched and the board is not created.
func TestTransitionIntoRun_RejectsTraversalSpecID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "leader-sess")
	for _, bad := range []string{"../sibling", "..", "a/../b", "a/b", `/etc/x`} {
		err := TransitionIntoRun(root, "leader-sess", bad)
		if err == nil {
			t.Errorf("TransitionIntoRun(%q) err = nil, want refusal — a traversal specID must not be persisted", bad)
		}
	}
	if _, statErr := os.Stat(BoardPath(root)); statErr == nil {
		t.Fatal("board state file appeared despite traversal-specID refusal")
	}
}

// TestReadCardStatus_AcceptsCanonicalSpecID — positive control: a canonical
// SPEC-ID is accepted (the guard is conditional on shape, not unconditional).
func TestReadCardStatus_AcceptsCanonicalSpecID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bases := WorktreeBases{}
	for _, ok := range []string{"SPEC-NAV-001", "SPEC-KANBAN-BOARD-001"} {
		if _, err := ReadCardStatus(root, ok, bases); err != nil {
			t.Errorf("ReadCardStatus(%q) error = %v, want acceptance — canonical ids must pass", ok, err)
		}
	}
}

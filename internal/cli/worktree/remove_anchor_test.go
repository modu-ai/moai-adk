package worktree

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// runRemoveCmd executes `moai worktree remove` with the given args against
// the installed mock provider and returns (stdout+stderr, error).
func runRemoveCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRemoveCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String() + errOut.String(), err
}

// withAnchoredMockProvider installs a mock provider whose single worktree
// sits at tree, which carries a live anchored session in its tree-local
// registry. Returns the mock.
func withAnchoredMockProvider(t *testing.T, tree string) *mockWorktreeProvider {
	t.Helper()
	writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	t.Cleanup(func() { WorktreeProvider = origProvider })
	return mock
}

// TestRunRemove_RefusesWhenAnchoredSessionLive (t73 surface 1): `moai
// worktree remove <path>` must refuse with the ANCHORED_SESSIONS_PRESENT
// sentinel while a live session is anchored in the tree.
func TestRunRemove_RefusesWhenAnchoredSessionLive(t *testing.T) {
	tree := t.TempDir()
	mock := withAnchoredMockProvider(t, tree)

	_, err := runRemoveCmd(t, tree)
	if err == nil {
		t.Fatal("remove must refuse (non-zero exit) while a live session is anchored in the worktree")
	}
	if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
		t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
	}
	if mock.removeCalled {
		t.Error("Remove() must not be called on refusal")
	}
}

// TestRunRemove_ForceOverridesAnchorWarning: --force removes the tree but
// warns about the anchored session it is cutting down.
func TestRunRemove_ForceOverridesAnchorWarning(t *testing.T) {
	tree := t.TempDir()
	mock := withAnchoredMockProvider(t, tree)

	combined, err := runRemoveCmd(t, "--force", tree)
	if err != nil {
		t.Fatalf("--force must remove despite the anchor, got error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("Remove() must be called under --force")
	}
	// Assert on the distinctive warning phrase, NOT the bare word "Warning":
	// t.TempDir() embeds the test name in the path, and this test's own name
	// would make a bare "Warning" match vacuously.
	if !strings.Contains(combined, "force removing") {
		t.Errorf("--force must warn about the anchored session(s), output: %q", combined)
	}
}

// TestRunRemove_RemovesWhenNoAnchoredSession: a tree without a tree-local
// registry disposes as before (fail-open on absent registry).
func TestRunRemove_RemovesWhenNoAnchoredSession(t *testing.T) {
	tree := t.TempDir()
	mock := withAnchoredMockProvider(t, tree)
	// Drop the registry the helper wrote — no anchor evidence remains.
	if err := os.RemoveAll(tree + "/.moai"); err != nil {
		t.Fatalf("remove registry: %v", err)
	}

	if _, err := runRemoveCmd(t, tree); err != nil {
		t.Fatalf("no anchor present — removal must proceed, got error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("Remove() must be called when no anchor is present")
	}
}

package worktree

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

// TestRunDone_RefusesWhenAnchoredSessionInCallerRegistry reproduces the
// MEASURED lane topology (2026-08-17): a `moai cc -w` lane registers in the
// LAUNCHER checkout's registry with cwd=<worktree> — the worktree itself
// carries no local registry. The anchor guard must therefore consult the
// caller's project registry (CLAUDE_PROJECT_DIR, else CWD), not only the
// tree-local one, or it misses the exact incident class t46 was built for.
func TestRunDone_RefusesWhenAnchoredSessionInCallerRegistry(t *testing.T) {
	tree := t.TempDir()
	launcher := t.TempDir() // the checkout `moai cc -w` launched from
	t.Setenv("CLAUDE_PROJECT_DIR", launcher)

	// cc-w lane shape: entry lives in the LAUNCHER registry, cwd = tree.
	// The tree itself has no .moai/state registry at all.
	writeTreeRegistry(t, launcher, []session.Entry{anchoredEntry(t, tree, os.Getpid())})

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: "feature/SPEC-ANCHOR-001", Path: tree}},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"feature/SPEC-ANCHOR-001"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("runDone must refuse while a live cc-w lane session (registered in the caller registry) is anchored in the worktree")
	}
	if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
		t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
	}
	if mock.removeCalled {
		t.Error("Remove() must not be called on refusal")
	}
}

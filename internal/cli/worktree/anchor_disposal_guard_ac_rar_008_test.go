package worktree

// anchor_disposal_guard_ac_rar_008_test.go — t1058 /
// SPEC-SESSION-REGISTRY-READ-ANCHOR-001 AC-RAR-008.
//
// One fixture per CLI consumer coordinate of the disposal guard, each in BOTH
// directions:
//
//   - internal/cli/worktree/remove.go:51  — `moai worktree remove <path>`
//   - internal/cli/worktree/done.go:77    — auto-mode cleanup
//   - internal/cli/worktree/done.go:176   — interactive `moai worktree done`
//
// (The fourth coordinate, internal/session/anchor_lock.go:111, is covered by
// the sibling fixture in internal/session — AnchorDecision is not reachable
// from this package's surface.)
//
// [HARD] Both directions, and the converse is the load-bearing one. A fixture
// that only measures "refuses disposal while a live session is anchored" is
// satisfied by an implementation that refuses EVERYTHING — the empty-green
// shape this SPEC exists to prevent. Each case below therefore pairs the
// refusal with a converse control on the same coordinate: a worktree whose
// only registry entry is dead is reported FREE and disposal proceeds.
//
// What these fixtures establish, stated at its actual grade: the
// live-anchored-worktree judgement is MEASURED unchanged at S1 landing, and
// they pre-place the regression guard that will catch S2 (anchoring
// LiveAnchoredSessions) breaking the disposal path. They do NOT verify the
// criterion under S2 — S2 is not in this run, so its behaviour is unverified.

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/session"
)

const acRAR008Branch = "feature/SPEC-ANCHOR-001"

// deadEntry is the converse control's registry row: an entry that IS present
// in the registry but describes no live session — PID 0, which
// LiveAnchoredSessions rejects at its `e.PID > 0` guard without probing the
// OS, and a heartbeat two hours old, well past DefaultStaleMinutes.
//
// A present-but-dead entry is a stronger converse than an absent registry: it
// shows the guard evaluates LIVENESS rather than mere presence.
func deadEntry(t *testing.T, cwd string) session.Entry {
	t.Helper()
	e := anchoredEntry(t, cwd, 0)
	e.SessionID = "99999999-8888-7777-6666-555555555555"
	old := time.Now().UTC().Add(-2 * time.Hour)
	e.StartedAt, e.LastHeartbeat = old, old
	return e
}

// isolateCallerRegistry points CLAUDE_PROJECT_DIR at an empty directory so
// LiveAnchoredSessions' second root (the caller's project registry) cannot
// contribute a real entry from the machine this test runs on. Without it the
// converse cases would read whatever registry the working directory resolves
// to, and a passing result would not be attributable to the fixture.
func isolateCallerRegistry(t *testing.T) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
}

// installProvider installs a mock provider carrying one worktree at tree and
// restores the original on cleanup.
func installProvider(t *testing.T, tree string) *mockWorktreeProvider {
	t.Helper()
	orig := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{{Branch: acRAR008Branch, Path: tree}},
	}
	WorktreeProvider = mock
	t.Cleanup(func() { WorktreeProvider = orig })
	return mock
}

// TestACRAR008_RemoveCoordinate covers internal/cli/worktree/remove.go:51.
func TestACRAR008_RemoveCoordinate(t *testing.T) {
	t.Run("live anchored session refuses disposal", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})
		mock := installProvider(t, tree)

		_, err := runRemoveCmd(t, tree)
		if err == nil {
			t.Fatal("remove must refuse while a live session is anchored in the worktree")
		}
		if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
			t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
		}
		if mock.removeCalled {
			t.Error("Remove() must not be called on refusal")
		}
	})

	t.Run("converse control: no live session reports free", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{deadEntry(t, tree)})
		mock := installProvider(t, tree)

		if _, err := runRemoveCmd(t, tree); err != nil {
			t.Fatalf("no live session is anchored — removal must proceed, got error: %v", err)
		}
		if !mock.removeCalled {
			t.Error("Remove() must be called when no live session is anchored")
		}
	})
}

// TestACRAR008_DoneAutoCoordinate covers internal/cli/worktree/done.go:77 —
// the auto-mode path, which degrades to a not-done report rather than an
// error.
func TestACRAR008_DoneAutoCoordinate(t *testing.T) {
	t.Run("live anchored session skips removal", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})
		mock := installProvider(t, tree)

		success, err := runDoneWorktreeCleanup(acRAR008Branch, false, false)
		if err != nil {
			t.Fatalf("auto mode must degrade gracefully, got error: %v", err)
		}
		if success {
			t.Error("auto mode must report not-done while a live session is anchored")
		}
		if mock.removeCalled {
			t.Error("Remove() must not be called while a live session is anchored")
		}
	})

	t.Run("converse control: no live session reports free", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{deadEntry(t, tree)})
		mock := installProvider(t, tree)

		success, err := runDoneWorktreeCleanup(acRAR008Branch, false, false)
		if err != nil {
			t.Fatalf("auto mode must proceed with no live anchor, got error: %v", err)
		}
		if !success {
			t.Error("auto mode must report done when no live session is anchored")
		}
		if !mock.removeCalled {
			t.Error("Remove() must be called when no live session is anchored")
		}
	})
}

// TestACRAR008_DoneInteractiveCoordinate covers
// internal/cli/worktree/done.go:176 — the interactive command path, which
// refuses with a non-zero exit.
func TestACRAR008_DoneInteractiveCoordinate(t *testing.T) {
	t.Run("live anchored session refuses disposal", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{anchoredEntry(t, tree, os.Getpid())})
		mock := installProvider(t, tree)

		cmd := newDoneCmd()
		cmd.SetOut(new(strings.Builder))
		cmd.SetErr(new(strings.Builder))
		cmd.SetArgs([]string{acRAR008Branch})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("done must refuse while a live session is anchored in the worktree")
		}
		if !strings.Contains(err.Error(), "ANCHORED_SESSIONS_PRESENT") {
			t.Errorf("refusal must carry the ANCHORED_SESSIONS_PRESENT sentinel, got: %v", err)
		}
		if mock.removeCalled {
			t.Error("Remove() must not be called on refusal")
		}
	})

	t.Run("converse control: no live session reports free", func(t *testing.T) {
		isolateCallerRegistry(t)
		tree := t.TempDir()
		writeTreeRegistry(t, tree, []session.Entry{deadEntry(t, tree)})
		mock := installProvider(t, tree)

		cmd := newDoneCmd()
		cmd.SetOut(new(strings.Builder))
		cmd.SetErr(new(strings.Builder))
		cmd.SetArgs([]string{acRAR008Branch})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("no live session is anchored — removal must proceed, got error: %v", err)
		}
		if !mock.removeCalled {
			t.Error("Remove() must be called when no live session is anchored")
		}
	})
}

package worktree

// migrate_review_t2053_test.go — card t2053, codex review findings.
//
//	F1  a `git worktree list` failure (reachable now that the call is bounded by
//	    gitCommandTimeout) used to be swallowed by the planner, so the apply
//	    printed "0 moved, 0 skipped, 0 failed" as if there were no candidates.
//	    The failure must reach the caller — the verb and the update
//	    consolidation, dry run and apply alike.
//	F2  a reverse `git worktree move` that errors AFTER the filesystem move
//	    completed (a timeout kills git late) left the tree at the source while
//	    the re-key and the rebind still named the destination. The location is
//	    re-checked by identity before the move is called failed.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t2053ListTimeout makes `git worktree list` hang until the shortened bound
// kills it.
func t2053ListTimeout(t *testing.T) {
	t.Helper()
	t2053BlockingGit(t, "list")
	orig := gitCommandTimeout
	gitCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { gitCommandTimeout = orig })
}

// TestMigrationEngine_WorktreeListTimeoutIsAnError — F1: the engine returns an
// error naming the listing failure; it never returns a clean empty report.
func TestMigrationEngine_WorktreeListTimeoutIsAnError(t *testing.T) {
	for _, apply := range []bool{false, true} {
		t.Run(fmt.Sprintf("apply=%v", apply), func(t *testing.T) {
			w := t2053TwoTreeWorld(t)
			t2053ListTimeout(t)
			if apply {
				if err := os.MkdirAll(filepath.Dir(migrationLockPath(w.repo)), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
			}
			report, err := runMigrationEngine(w.repo, apply, "migrate")
			if err == nil {
				t.Fatalf("a failed worktree listing returned success (report=%+v)", report)
			}
			if !strings.Contains(err.Error(), "timed out") {
				t.Errorf("error %q does not carry the timeout reason", err)
			}
		})
	}
}

// TestMigrateVerb_WorktreeListTimeoutFails — F1: the verb exits non-zero with
// the reason and does not print the success summary.
func TestMigrateVerb_WorktreeListTimeoutFails(t *testing.T) {
	for _, args := range [][]string{{"--apply"}, {}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_ = t2053TwoTreeWorld(t)
			t2053ListTimeout(t)
			out, err := m4RunMigrate(t, args...)
			if err == nil {
				t.Fatalf("the verb succeeded on a failed worktree listing:\n%s", out)
			}
			if !strings.Contains(err.Error(), "timed out") {
				t.Errorf("error %q does not carry the timeout reason", err)
			}
			if strings.Contains(out, "moved 0, skipped 0, failed 0") {
				t.Errorf("the output reads as a clean zero-target success:\n%s", out)
			}
		})
	}
}

// TestUpdateConsolidation_WorktreeListTimeoutFails — F1: the consolidation
// returns the error (the update turns it into its warning row); the success
// summary line is not printed.
func TestUpdateConsolidation_WorktreeListTimeoutFails(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(fmt.Sprintf("dryRun=%v", dryRun), func(t *testing.T) {
			w := t2053TwoTreeWorld(t)
			t2053ListTimeout(t)
			buf := &strings.Builder{}
			err := runUpdateConsolidation(w.repo, dryRun, buf)
			if err == nil {
				t.Fatalf("the consolidation succeeded on a failed worktree listing:\n%s", buf.String())
			}
			if !strings.Contains(err.Error(), "timed out") {
				t.Errorf("error %q does not carry the timeout reason", err)
			}
			if strings.Contains(buf.String(), "0 moved, 0 skipped, 0 failed") {
				t.Errorf("the output reads as a clean zero-target success:\n%s", buf.String())
			}
		})
	}
}

// t2053ReverseWorld performs the forward move by hand (the state m6ReverseMove
// starts from) and returns the plan, identity and move-list path.
func t2053ReverseWorld(t *testing.T) (*m6World, migratePlan, TreeIdentity, string) {
	t.Helper()
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	dest := filepath.Join(w.homeRoot, "legacy6")
	id, err := EnsureInstanceToken(w.movedTree)
	if err != nil {
		t.Fatalf("instance token: %v", err)
	}
	m6Git(t, w.repo, "worktree", "move", w.movedTree, dest)
	mlPath := filepath.Join(w.repo, ".moai", "state", "worktree-migrate", "move-list.jsonl")
	if err := os.MkdirAll(filepath.Dir(mlPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return w, migratePlan{source: w.movedTree, destination: dest}, id, mlPath
}

// TestReverseMove_ErrorAfterMoveCompleted_IsTreatedAsMovedBack — F2: the seam
// performs the move and THEN reports a timeout. The tree stands at the source
// under its own identity and the destination is gone, so the reverse order
// completes: the card is rebound onto the source and a `reversed` line lands.
func TestReverseMove_ErrorAfterMoveCompleted_IsTreatedAsMovedBack(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)

	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	dest := p.destination
	if _, perr := db.RecordPicked(context.Background(), "run-t2053", "card-t2053", homestate.CardFields{WorktreePath: &dest}, "t2053-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked: %v", perr)
	}
	_ = db.Close()

	orig := reverseGitMove
	reverseGitMove = func(root, from, to string) error {
		m6Git(t, root, "worktree", "move", from, to) // the move completes ...
		return fmt.Errorf("%w after 1ms", ErrCommandTimeout)
	}
	t.Cleanup(func() { reverseGitMove = orig })

	movedBack, mlErr := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if mlErr != nil {
		t.Fatalf("move list: %v", mlErr)
	}
	if !movedBack {
		t.Fatalf("a reverse move that completed before git reported the timeout was not treated as moved back")
	}
	if !m6Registered(t, w.repo, p.source) || m6Registered(t, w.repo, p.destination) {
		t.Errorf("registration does not name the source only")
	}
	raw, rerr := os.ReadFile(mlPath)
	if rerr != nil {
		t.Fatalf("read move list: %v", rerr)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if last := lines[len(lines)-1]; !strings.Contains(last, `"result":"reversed"`) {
		t.Errorf("last move-list line is not `reversed`: %s", last)
	}
	db2, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("reopen factory: %v", err)
	}
	defer func() { _ = db2.Close() }()
	card, lerr := db2.LoadCard(context.Background(), "run-t2053", "card-t2053")
	if lerr != nil {
		t.Fatalf("load card: %v", lerr)
	}
	if card.WorktreePath != p.source {
		t.Errorf("card worktree = %q, want it rebound to the source %q", card.WorktreePath, p.source)
	}
}

// TestReverseMove_ErrorWhileTreeStaysAtDestination_StaysPending — F2's other
// half: the move did not happen, the destination still carries the identity,
// so the tree stays pending and not moved back.
func TestReverseMove_ErrorWhileTreeStaysAtDestination_StaysPending(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	orig := reverseGitMove
	reverseGitMove = func(root, from, to string) error { return fmt.Errorf("%w after 1ms", ErrCommandTimeout) }
	t.Cleanup(func() { reverseGitMove = orig })

	movedBack, mlErr := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if mlErr != nil {
		t.Fatalf("move list: %v", mlErr)
	}
	if movedBack {
		t.Errorf("a tree that never left the destination was reported moved back")
	}
	if !MovedBy(w.repo, p.destination) {
		t.Errorf("no pending protection at the destination")
	}
}

// TestReverseMove_AmbiguousLocation_StaysPending — F2: when neither location
// carries the identity the state is ambiguous and fails safe.
func TestReverseMove_AmbiguousLocation_StaysPending(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	if err := os.RemoveAll(p.destination); err != nil {
		t.Fatalf("remove destination: %v", err)
	}
	orig := reverseGitMove
	reverseGitMove = func(root, from, to string) error { return fmt.Errorf("%w after 1ms", ErrCommandTimeout) }
	t.Cleanup(func() { reverseGitMove = orig })

	movedBack, _ := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if movedBack {
		t.Errorf("an ambiguous location (identity at neither path) was reported moved back")
	}
}

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
	"encoding/json"
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

// ---- second review round (G1, G2) -------------------------------------------
//
//	G1  identity at the source is not enough: after an interrupted reverse move
//	    git's worktrees/<id>/gitdir may still name the destination. The tree is
//	    "moved back" only when git's registration names the source — after one
//	    bounded `git worktree repair` if it did not.
//	G2  --rollback shares the post-error re-check (and the registration check).

// t2053RenameOnlyReverse is a reverse-move seam that renames the directory the
// way a killed `git worktree move` leaves it — the files at the target, git's
// registration still naming the origin — and then reports the timeout.
func t2053RenameOnlyReverse(t *testing.T) {
	t.Helper()
	orig := reverseGitMove
	reverseGitMove = func(root, from, to string) error {
		if err := os.Rename(from, to); err != nil {
			return err
		}
		return fmt.Errorf("%w after 1ms", ErrCommandTimeout)
	}
	t.Cleanup(func() { reverseGitMove = orig })
}

// t2053LastMoveListResult returns the `result` of the last move-list line.
func t2053LastMoveListResult(t *testing.T, mlPath string) string {
	t.Helper()
	raw, err := os.ReadFile(mlPath)
	if err != nil {
		t.Fatalf("read move list: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	var rec struct {
		Result string `json:"result"`
	}
	if jerr := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); jerr != nil {
		t.Fatalf("parse %q: %v", lines[len(lines)-1], jerr)
	}
	return rec.Result
}

// TestReverseMove_StaleRegistration_RepairConverges — G1: the directory is back
// at the source but git still registers the destination. One bounded
// `git worktree repair <source>` converges it; only then is the tree reversed.
func TestReverseMove_StaleRegistration_RepairConverges(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	t2053RenameOnlyReverse(t)

	movedBack, mlErr := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if mlErr != nil {
		t.Fatalf("move list: %v", mlErr)
	}
	if !movedBack {
		t.Fatalf("a repairable stale registration was not converged")
	}
	if !m6Registered(t, w.repo, p.source) || m6Registered(t, w.repo, p.destination) {
		t.Errorf("registration does not name the source only after the repair")
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "reversed" {
		t.Errorf("last move-list result = %q, want reversed", got)
	}
}

// TestReverseMove_StaleRegistration_UnrepairableNeverReversed — G1: when the
// repair cannot run (a hung `git worktree repair`), the registration still
// names the destination, so the tree is NOT reported reversed.
func TestReverseMove_StaleRegistration_UnrepairableNeverReversed(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	t2053RenameOnlyReverse(t)
	t2053BlockingGit(t, "repair")
	orig := gitCommandTimeout
	gitCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { gitCommandTimeout = orig })

	movedBack, _ := m6ReverseMove(w.repo, p, id, mlPath, "migrate")
	if movedBack {
		t.Fatalf("reported moved back while git still registers the destination")
	}
	if !m6Registered(t, w.repo, p.destination) {
		t.Fatalf("fixture error: the registration no longer names the destination")
	}
	if got := t2053LastMoveListResult(t, mlPath); got == "reversed" {
		t.Errorf("a `reversed` line was written while the registration names the destination")
	}
}

// t2053RollbackJournal writes a one-step rollback journal for the plan and
// returns its path.
func t2053RollbackJournal(t *testing.T, p migratePlan, id TreeIdentity) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: p.source, NewPath: p.destination, Outcome: "moved", Ident: id.Ident(),
	}}})
	if err != nil {
		t.Fatalf("marshal journal: %v", err)
	}
	path := filepath.Join(t.TempDir(), "journal.json")
	if werr := os.WriteFile(path, body, 0o644); werr != nil {
		t.Fatalf("write journal: %v", werr)
	}
	return path
}

// TestRollback_ReverseErrorAfterMoveCompleted_RecordsRestored — G2: the seam
// performs the move and then reports a timeout; the rollback re-keys, rebinds
// the card and records `restored` instead of leaving the step pending.
func TestRollback_ReverseErrorAfterMoveCompleted_RecordsRestored(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	journal := t2053RollbackJournal(t, p, id)

	db, err := homestate.OpenFactory(w.repo)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	dest := p.destination
	if _, perr := db.RecordPicked(context.Background(), "run-t2053-rb", "card-t2053-rb", homestate.CardFields{WorktreePath: &dest}, "t2053-test", time.Now().UTC()); perr != nil {
		t.Fatalf("record picked: %v", perr)
	}
	_ = db.Close()

	orig := reverseGitMove
	reverseGitMove = func(root, from, to string) error {
		m6Git(t, root, "worktree", "move", from, to)
		return fmt.Errorf("%w after 1ms", ErrCommandTimeout)
	}
	t.Cleanup(func() { reverseGitMove = orig })

	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr != nil {
		t.Fatalf("rollback failed although the move had completed: %v\n%s", rerr, out)
	}
	if !strings.Contains(out, "restored 1 tree(s)") {
		t.Errorf("rollback summary: %q", out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "restored" {
		t.Errorf("last move-list result = %q, want restored", got)
	}
	db2, derr := homestate.OpenFactory(w.repo)
	if derr != nil {
		t.Fatalf("reopen factory: %v", derr)
	}
	defer func() { _ = db2.Close() }()
	card, lerr := db2.LoadCard(context.Background(), "run-t2053-rb", "card-t2053-rb")
	if lerr != nil {
		t.Fatalf("load card: %v", lerr)
	}
	if card.WorktreePath != p.source {
		t.Errorf("card worktree = %q, want it rebound to %q", card.WorktreePath, p.source)
	}
}

// TestRollback_StaleRegistrationUnrepairable_StaysPending — G2 + G1: a rollback
// whose reverse move errors with the registration still naming the destination
// records the step pending and fails the run; it never records `restored`.
func TestRollback_StaleRegistrationUnrepairable_StaysPending(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t)
	journal := t2053RollbackJournal(t, p, id)
	t2053RenameOnlyReverse(t)
	t2053BlockingGit(t, "repair")
	orig := gitCommandTimeout
	gitCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { gitCommandTimeout = orig })

	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil {
		t.Fatalf("rollback reported success while the registration names the destination:\n%s", out)
	}
	if raw, err := os.ReadFile(mlPath); err == nil && strings.Contains(string(raw), `"result":"restored"`) {
		t.Errorf("a `restored` line was written:\n%s", raw)
	}
	_ = w
}

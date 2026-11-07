package worktree

// migrate_hardening_e_t2053_test.go — card t2053, sixth review round.
//
//	L1  an interrupted-migration check that depends on where the tree stands lets
//	    a rerun move a tree whose earlier step is still pending
//	L2  verify reads a source it cannot inspect as a gone one
//	class sweep: an error that is not fs.ErrNotExist never means "absent",
//	"not this tree", "nothing to do" or "converged"

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func t2053SkipIfRoot(t *testing.T) {
	t.Helper()
	t2085SkipOnWindows(t, t2085ReasonPosixMode)
	if os.Geteuid() == 0 {
		t.Skip("root ignores file modes")
	}
}

// t2053Chmod sets mode on path and restores 0755/0644 when the test ends.
func t2053Chmod(t *testing.T, path string, mode os.FileMode, restore os.FileMode) {
	t.Helper()
	t2085SkipOnWindows(t, t2085ReasonPosixMode)
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod %s: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, restore) })
}

// t2053PendingIntentWorld is the state a run killed after writing its journal
// and move-list `pending` lines but BEFORE `git worktree move` leaves: the
// tree intact and registered at its source, the destination absent.
func t2053PendingIntentWorld(t *testing.T, ident string) (*m6World, migratePlan, TreeIdentity, string) {
	t.Helper()
	w := t2053TwoTreeWorld(t)
	dest := filepath.Join(w.homeRoot, "legacy6")
	id, err := EnsureInstanceToken(w.movedTree)
	if err != nil {
		t.Fatalf("instance token: %v", err)
	}
	if ident == "" {
		ident = id.Ident()
	}
	if merr := os.MkdirAll(t2053StateDir(w.repo), 0o755); merr != nil {
		t.Fatalf("mkdir: %v", merr)
	}
	if aerr := appendMoveListLine(moveListPath(w.repo), m6MoveLine("pending", w.movedTree, dest, ident)); aerr != nil {
		t.Fatalf("pending line: %v", aerr)
	}
	body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: w.movedTree, NewPath: dest, Outcome: "pending", Ident: ident,
	}}, "trigger": "migrate"})
	src := filepath.Join(t.TempDir(), "run.json")
	if werr := os.WriteFile(src, body, 0o644); werr != nil {
		t.Fatalf("write journal: %v", werr)
	}
	return w, migratePlan{source: w.movedTree, destination: dest}, id, t2053PlaceJournal(t, w.repo, src, "20260101T000000.json")
}

func t2053ResultsFor(t *testing.T, repo, ident string) []string {
	t.Helper()
	raw, err := os.ReadFile(moveListPath(repo))
	if err != nil {
		t.Fatalf("read move list: %v", err)
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var rec mlRecord
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Ident == ident {
			out = append(out, rec.Result)
		}
	}
	return out
}

func t2053ReadJournalSteps(t *testing.T, path string) []journalStep {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	var j struct {
		Steps []journalStep `json:"steps"`
	}
	if jerr := json.Unmarshal(raw, &j); jerr != nil {
		t.Fatalf("parse journal: %v", jerr)
	}
	return j.Steps
}

// TestApply_PendingIntentButNeverMoved_IsSettledThenMoved — L1: the source is
// intact and registered and the destination is absent, so the earlier step
// provably never moved. The rerun settles it (journal `failed`, a releasing
// move-list line), moves the tree, and a later --rollback of the OLD journal
// must not reverse the new move.
func TestApply_PendingIntentButNeverMoved_IsSettledThenMoved(t *testing.T) {
	w, p, id, journal := t2053PendingIntentWorld(t, "")

	if out, err := m4RunMigrate(t, "--apply"); err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	if _, serr := os.Stat(p.destination); serr != nil {
		t.Fatalf("the tree was not moved to %s: %v", p.destination, serr)
	}
	steps := t2053ReadJournalSteps(t, journal)
	if len(steps) != 1 || steps[0].Outcome != "failed" || !strings.Contains(steps[0].Reason, "interrupted before the move") {
		t.Errorf("the old journal step was not settled: %+v", steps)
	}
	got := strings.Join(t2053ResultsFor(t, w.repo, id.Ident()), ",")
	if got != "pending,reversed,pending,moved" {
		t.Errorf("move-list results = %s, want pending,reversed,pending,moved", got)
	}

	out, err := m4RunMigrate(t, "--rollback", journal)
	if err != nil {
		t.Fatalf("rollback of the settled old journal: %v\n%s", err, out)
	}
	if _, serr := os.Stat(p.destination); serr != nil {
		t.Errorf("rollback of the OLD journal reversed the new move: %v", serr)
	}
	if !m6Registered(t, w.repo, p.destination) {
		t.Errorf("the new move lost its registration")
	}
	if !MovedBy(w.repo, p.destination) {
		t.Errorf("the moved tree is no longer protected")
	}
}

// TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted — L1: anything but
// the provable case stays interrupted and moves nothing.
func TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted(t *testing.T) {
	cases := []struct {
		name  string
		ident string
		setup func(t *testing.T, w *m6World, p migratePlan)
	}{
		{name: "the source carries another identity", ident: "other:ident"},
		{name: "the destination cannot be inspected", setup: func(t *testing.T, w *m6World, p migratePlan) {
			t2053SkipIfRoot(t)
			t2053Chmod(t, filepath.Dir(p.destination), 0o000, 0o755)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, p, _, journal := t2053PendingIntentWorld(t, tc.ident)
			if tc.setup != nil {
				tc.setup(t, w, p)
			}
			journalBefore, _ := os.ReadFile(journal)
			listBefore, _ := os.ReadFile(moveListPath(w.repo))

			out, err := m4RunMigrate(t)
			if err == nil {
				t.Fatalf("a not-provably-unmoved pending step exited clean:\n%s", out)
			}
			if !strings.Contains(out, "interrupted") || !strings.Contains(out, "--rollback") || !strings.Contains(out, journal) {
				t.Errorf("output does not name the interrupted migration and its rollback:\n%s", out)
			}
			if _, serr := os.Lstat(w.movedTree); serr != nil {
				t.Errorf("the source was touched: %v", serr)
			}
			journalAfter, _ := os.ReadFile(journal)
			if string(journalAfter) != string(journalBefore) {
				t.Errorf("the journal was rewritten")
			}
			listAfter, _ := os.ReadFile(moveListPath(w.repo))
			if string(listAfter) != string(listBefore) {
				t.Errorf("the move list changed:\n%s", listAfter)
			}
		})
	}
}

// TestPlan_PendingIntent_DryRunWritesNothing — L1: the converge is an apply
// act; the dry run of the provable case reports the tree as a candidate and
// leaves the journal and the move list alone.
func TestPlan_PendingIntent_DryRunWritesNothing(t *testing.T) {
	w, _, _, journal := t2053PendingIntentWorld(t, "")
	journalBefore, _ := os.ReadFile(journal)
	listBefore, _ := os.ReadFile(moveListPath(w.repo))

	out, err := m4RunMigrate(t)
	if err != nil {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	journalAfter, _ := os.ReadFile(journal)
	listAfter, _ := os.ReadFile(moveListPath(w.repo))
	if string(journalAfter) != string(journalBefore) || string(listAfter) != string(listBefore) {
		t.Errorf("the dry run wrote to the journal or the move list")
	}
}

// TestRollback_StepSettledAsNeverMoved_IsANoOp — L1: a step the planner settled
// as "interrupted before the move" is not a move; rolling it back does nothing
// even when the same identity now stands at the destination.
func TestRollback_StepSettledAsNeverMoved_IsANoOp(t *testing.T) {
	w, p, id, mlPath := t2053ReverseWorld(t) // the tree really stands at the destination
	body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
		OldPath: p.source, NewPath: p.destination, Outcome: "failed", Ident: id.Ident(),
		Reason: "interrupted before the move",
	}}})
	journal := filepath.Join(t.TempDir(), "journal.json")
	if err := os.WriteFile(journal, body, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}
	listBefore, _ := os.ReadFile(mlPath)

	out, err := m4RunMigrate(t, "--rollback", journal)
	if err != nil {
		t.Fatalf("rollback: %v\n%s", err, out)
	}
	if _, serr := os.Stat(p.destination); serr != nil {
		t.Errorf("the rollback reversed a tree its step never moved: %v", serr)
	}
	if !m6Registered(t, w.repo, p.destination) {
		t.Errorf("registration no longer names the destination")
	}
	listAfter, _ := os.ReadFile(mlPath)
	if string(listAfter) != string(listBefore) {
		t.Errorf("a no-op rollback wrote move-list lines:\n%s", listAfter)
	}
}

func t2053VerifyRows(t *testing.T) ([]map[string]any, error) {
	t.Helper()
	out, err := m4RunMigrate(t, "--verify", "--json")
	var rows []map[string]any
	firstJSONValue(t, out, &rows)
	return rows, err
}

// TestVerify_UnreadableEnds_AreFailuresNotGone — L2: a source that cannot be
// inspected, and a destination whose identity cannot be read, are unknown —
// verify reports them as failures with the error, it does not drop the step.
func TestVerify_UnreadableEnds_AreFailuresNotGone(t *testing.T) {
	t2053SkipIfRoot(t)
	t.Run("pending step, source cannot be inspected", func(t *testing.T) {
		w, p, id, _ := t2053PendingIntentWorld(t, "")
		_ = id
		t2053Chmod(t, filepath.Dir(p.source), 0o000, 0o755)
		rows, err := t2053VerifyRows(t)
		if err == nil {
			t.Fatalf("verify succeeded over a source it cannot inspect: %v", rows)
		}
		found := false
		for _, r := range rows {
			if r["old_path"] == w.movedTree && strings.Contains(r["reason"].(string), "cannot be inspected") {
				found = true
			}
		}
		if !found {
			t.Errorf("no row names the uninspectable source: %v", rows)
		}
	})
	t.Run("pending step, destination cannot be inspected", func(t *testing.T) {
		_, p, _, _ := t2053PendingIntentWorld(t, "")
		t2053Chmod(t, filepath.Dir(p.destination), 0o000, 0o755)
		rows, err := t2053VerifyRows(t)
		if err == nil {
			t.Fatalf("verify succeeded over a destination it cannot inspect: %v", rows)
		}
		found := false
		for _, r := range rows {
			if r["kind"] == "unverifiable" && strings.Contains(r["reason"].(string), "destination cannot be inspected") {
				found = true
			}
		}
		if !found {
			t.Errorf("no unverifiable row for the uninspectable destination: %v", rows)
		}
	})
	t.Run("pending step, move list cannot be read", func(t *testing.T) {
		w, _, _, _ := t2053PendingIntentWorld(t, "")
		t2053Chmod(t, moveListPath(w.repo), 0o000, 0o644)
		rows, err := t2053VerifyRows(t)
		if err == nil {
			t.Fatalf("verify succeeded without reading the move list: %v", rows)
		}
		found := false
		for _, r := range rows {
			if r["kind"] == "unverifiable" && strings.Contains(r["reason"].(string), "move list") {
				found = true
			}
		}
		if !found {
			t.Errorf("no unverifiable row naming the move list: %v", rows)
		}
	})
	for _, outcome := range []string{"moved", "failed"} {
		t.Run(outcome+" step, destination identity unreadable", func(t *testing.T) {
			w, p, id, _ := t2053ReverseWorld(t) // the tree stands at the destination
			body, _ := json.Marshal(map[string]any{"steps": []journalStep{{
				OldPath: p.source, NewPath: p.destination, Outcome: outcome, Ident: id.Ident(),
			}}})
			src := filepath.Join(t.TempDir(), "run.json")
			_ = os.WriteFile(src, body, 0o644)
			t2053PlaceJournal(t, w.repo, src, "20260101T000000.json")
			t2053Chmod(t, filepath.Join(p.destination, ".git"), 0o000, 0o644)
			rows, err := t2053VerifyRows(t)
			if err == nil {
				t.Fatalf("verify succeeded over a destination whose identity it cannot read: %v", rows)
			}
			found := false
			for _, r := range rows {
				if r["kind"] == "unverifiable" && r["new_path"] == p.destination {
					found = true
				}
			}
			if !found {
				t.Errorf("no unverifiable row for the unreadable destination: %v", rows)
			}
		})
	}
}

// TestTreeIdent_UnreadableToken_IsAnError — sweep: an instance token that
// exists but cannot be read is not "never tokened".
func TestTreeIdent_UnreadableToken_IsAnError(t *testing.T) {
	t2053SkipIfRoot(t)
	_, p, _, _ := t2053ReverseWorld(t)
	gitdir, err := gitDirPointer(p.destination)
	if err != nil {
		t.Fatalf("gitdir: %v", err)
	}
	t2053Chmod(t, filepath.Join(gitdir, tokenFileName), 0o000, 0o644)
	if id, terr := TreeIdent(p.destination); terr == nil {
		t.Errorf("TreeIdent read an unreadable token as an empty one: %+v", id)
	}
}

// TestEnsureInstanceToken_UnreadableToken_IsNotReplaced — sweep: a token that
// cannot be read must not be overwritten with a fresh one (that would orphan
// every move-list line and journal step carrying the old identity).
func TestEnsureInstanceToken_UnreadableToken_IsNotReplaced(t *testing.T) {
	t2053SkipIfRoot(t)
	_, p, id, _ := t2053ReverseWorld(t)
	gitdir, _ := gitDirPointer(p.destination)
	tokenFile := filepath.Join(gitdir, tokenFileName)
	t2053Chmod(t, tokenFile, 0o200, 0o644) // writable but unreadable: the overwrite would succeed
	if got, err := EnsureInstanceToken(p.destination); err == nil {
		t.Errorf("EnsureInstanceToken answered %+v for a token it cannot read", got)
	}
	_ = os.Chmod(tokenFile, 0o644)
	body, _ := os.ReadFile(tokenFile)
	if strings.TrimSpace(string(body)) != id.Token {
		t.Errorf("the token was overwritten: %q, want %q", strings.TrimSpace(string(body)), id.Token)
	}
}

// TestMovedBy_UnknownIsProtected — sweep: the cleanup guard answers "not moved"
// only when it can show the tree was not; everything it cannot read protects.
func TestMovedBy_UnknownIsProtected(t *testing.T) {
	t.Run("control: a moved, registered tree is protected", func(t *testing.T) {
		w, p, id, mlPath := t2053ReverseWorld(t)
		_ = appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
		if !MovedBy(w.repo, p.destination) {
			t.Fatalf("fixture error: a moved tree is not protected")
		}
	})
	t.Run("control: a path that is not a linked worktree is not protected", func(t *testing.T) {
		w, _, _, _ := t2053ReverseWorld(t)
		if MovedBy(w.repo, t.TempDir()) {
			t.Errorf("a plain directory is reported as moved by the migration")
		}
		if MovedBy(w.repo, filepath.Join(t.TempDir(), "missing")) {
			t.Errorf("a missing path is reported as moved by the migration")
		}
	})
	t.Run("token cannot be read", func(t *testing.T) {
		t2053SkipIfRoot(t)
		w, p, id, mlPath := t2053ReverseWorld(t)
		_ = appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
		gitdir, _ := gitDirPointer(p.destination)
		t2053Chmod(t, filepath.Join(gitdir, tokenFileName), 0o000, 0o644)
		if !MovedBy(w.repo, p.destination) {
			t.Errorf("an unreadable identity released the protection")
		}
	})
	t.Run("move list cannot be read", func(t *testing.T) {
		t2053SkipIfRoot(t)
		w, p, id, mlPath := t2053ReverseWorld(t)
		_ = appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
		t2053Chmod(t, mlPath, 0o000, 0o644)
		if !MovedBy(w.repo, p.destination) {
			t.Errorf("an unreadable move list released the protection")
		}
	})
	t.Run("git cannot list the worktrees", func(t *testing.T) {
		w, p, id, mlPath := t2053ReverseWorld(t)
		_ = appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
		t2053FailingGit(t, "list")
		if !MovedBy(w.repo, p.destination) {
			t.Errorf("a failed worktree listing released the protection")
		}
	})
}

// TestMovedTo_UnreadableMoveList_GivesNoAnswer — sweep: with the move list
// unreadable the lookup cannot vouch for a candidate, so it answers nothing.
func TestMovedTo_UnreadableMoveList_GivesNoAnswer(t *testing.T) {
	t2053SkipIfRoot(t)
	w, p, id, mlPath := t2053ReverseWorld(t)
	_ = appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
	t2053Chmod(t, mlPath, 0o000, 0o644)
	if got, ok := MovedTo(w.repo, p.source, MoveRequester{}, p.destination); ok {
		t.Errorf("MovedTo answered %s over an unreadable move list", got)
	}
}

// TestAppendMoveListLine_UnreadableList_IsAnError — sweep: the dedupe and the
// newline repair both read the list; a list that cannot be read is an error, not
// an empty one.
func TestAppendMoveListLine_UnreadableList_IsAnError(t *testing.T) {
	t2053SkipIfRoot(t)
	dir := t.TempDir()
	mlPath := filepath.Join(dir, "move-list.jsonl")
	if err := os.WriteFile(mlPath, []byte("{\"result\":\"pending\"}"), 0o644); err != nil { // no trailing newline
		t.Fatalf("seed: %v", err)
	}
	t2053Chmod(t, mlPath, 0o200, 0o644) // appendable, unreadable
	if err := appendMoveListLine(mlPath, `{"result":"moved"}`); err == nil {
		t.Errorf("appendMoveListLine appended to a list it could not read")
	}
	_ = os.Chmod(mlPath, 0o644)
	if body, _ := os.ReadFile(mlPath); strings.Contains(string(body), "moved") {
		t.Errorf("the line was appended anyway:\n%s", body)
	}
}

// TestTreeMovedBack_UnreadableDestination_IsNotMovedBack — sweep: a destination
// whose identity cannot be read does not prove the tree left it.
func TestTreeMovedBack_UnreadableDestination_IsNotMovedBack(t *testing.T) {
	t2053SkipIfRoot(t)
	w, p, id, _ := t2053ReverseWorld(t)
	m6Git(t, w.repo, "worktree", "move", p.destination, p.source) // really back, registered at the source
	if err := os.MkdirAll(p.destination, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gitFile := filepath.Join(p.destination, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: "+filepath.Join(w.repo, ".git", "worktrees", "elsewhere")+"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t2053Chmod(t, gitFile, 0o000, 0o644)
	if treeMovedBack(w.repo, p.source, p.destination, id.Ident()) {
		t.Errorf("treeMovedBack concluded from an unreadable destination identity")
	}
}

// TestReverseMove_UnreadableDestinationIdentity_StaysPending — sweep: the apply
// time reverse, like the rollback, holds a tree whose identity it cannot read
// (pending, with the error) instead of writing a `refused` verdict about it.
func TestReverseMove_UnreadableDestinationIdentity_StaysPending(t *testing.T) {
	t2053SkipIfRoot(t)
	w, p, id, mlPath := t2053ReverseWorld(t)
	t2053Chmod(t, filepath.Join(p.destination, ".git"), 0o000, 0o644)
	moved, why, mlErr := m6ReverseMoveWhy(w.repo, p, id, mlPath, "migrate")
	if mlErr != nil {
		t.Fatalf("move list: %v", mlErr)
	}
	if moved {
		t.Fatalf("reported moved back")
	}
	if got := t2053LastMoveListResult(t, mlPath); got != "pending" {
		t.Errorf("last result = %q, want pending (an unreadable identity is unknown, not a mismatch); why=%s", got, why)
	}
}

// TestRollback_UnreadableIdentity_HoldsInsteadOfRefusing — sweep: both rollback
// branches (tree at the destination, tree back at the source) hold a step whose
// tree identity cannot be read, naming the error; `refused: identity mismatch`
// is reserved for an identity that was read and differs.
func TestRollback_UnreadableIdentity_HoldsInsteadOfRefusing(t *testing.T) {
	t2053SkipIfRoot(t)
	rowFor := func(t *testing.T, out string) map[string]any {
		t.Helper()
		var rows []map[string]any
		firstJSONValue(t, out, &rows)
		if len(rows) != 1 {
			t.Fatalf("rows = %v", rows)
		}
		return rows[0]
	}
	t.Run("tree at the destination", func(t *testing.T) {
		_, p, id, _ := t2053ReverseWorld(t)
		journal := t2053RollbackJournal(t, p, id)
		t2053Chmod(t, filepath.Join(p.destination, ".git"), 0o000, 0o644)
		out, err := m4RunMigrate(t, "--rollback", journal, "--json")
		if err == nil {
			t.Fatalf("rollback exited clean over an unreadable identity:\n%s", out)
		}
		if r := rowFor(t, out); r["outcome"] != "pending" || strings.Contains(r["reason"].(string), "identity mismatch") {
			t.Errorf("row = %v, want pending naming the read error", r)
		}
	})
	t.Run("tree back at the source", func(t *testing.T) {
		w, p, id, _ := t2053ReverseWorld(t)
		m6Git(t, w.repo, "worktree", "move", p.destination, p.source)
		journal := t2053RollbackJournal(t, p, id)
		t2053Chmod(t, filepath.Join(p.source, ".git"), 0o000, 0o644)
		out, err := m4RunMigrate(t, "--rollback", journal, "--json")
		if err == nil {
			t.Fatalf("rollback exited clean over an unreadable identity:\n%s", out)
		}
		if r := rowFor(t, out); r["outcome"] != "pending" || strings.Contains(r["reason"].(string), "identity mismatch") {
			t.Errorf("row = %v, want pending naming the read error", r)
		}
	})
}

// TestRebindCardForPath_UnstatableStore_IsAnError — sweep: a factory store whose
// location cannot be inspected is not a repository without a store.
func TestRebindCardForPath_UnstatableStore_IsAnError(t *testing.T) {
	t2053SkipIfRoot(t)
	w := m6NewWorld(t)
	dbPath, err := homestate.FactoryDBPath(w.repo)
	if err != nil {
		t.Skipf("no factory path: %v", err)
	}
	if merr := os.MkdirAll(filepath.Dir(dbPath), 0o755); merr != nil {
		t.Fatalf("mkdir: %v", merr)
	}
	t2053Chmod(t, filepath.Dir(dbPath), 0o000, 0o755)
	if rerr := rebindCardForPath(w.repo, "/a", "/b"); rerr == nil {
		t.Errorf("a store behind an uninspectable directory was read as absent")
	}
}

// TestRemovalTargetLock_UnresolvablePath_IsAnError — remove.go sweep: when the
// symlink resolution of the target fails for any reason but absence, the lock
// state is unknown, not "unlocked".
func TestRemovalTargetLock_UnresolvablePath_IsAnError(t *testing.T) {
	t2053SkipIfRoot(t)
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	parent := filepath.Join(w.repo, ".moai", "worktrees")
	target := filepath.Join(parent, "not-registered")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t2053Chmod(t, parent, 0o000, 0o755)
	if _, err := removalTargetLock(target); err == nil {
		t.Errorf("an unresolvable target read as unlocked")
	}
	// Control: a target that does not exist is unlocked, not unknown.
	_ = os.Chmod(parent, 0o755)
	if _, err := removalTargetLock(filepath.Join(parent, "never-created")); err != nil {
		t.Errorf("a missing target is an error: %v", err)
	}
}

// TestHoist_UnreadableDestination_IsNotOverwritten — hoist.go sweep: an existing
// destination file that cannot be read is neither absent nor identical; the hoist
// stops with the error instead of overwriting it (REQ-RLC-006).
func TestHoist_UnreadableDestination_IsNotOverwritten(t *testing.T) {
	t2053SkipIfRoot(t)
	mainRoot, tree := t.TempDir(), filepath.Join(t.TempDir(), "WT-x")
	if err := os.MkdirAll(filepath.Join(tree, ".moai", "reports"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tree, ".moai", "reports", "a.txt"), []byte("incoming"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	dst := filepath.Join(mainRoot, ".moai", "reports", "worktrees", "WT-x", "a.txt")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dst, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t2053Chmod(t, dst, 0o200, 0o644)
	if _, err := hoistWorktreeReports(tree, mainRoot); err == nil {
		t.Errorf("the hoist went past an unreadable destination file")
	}
	_ = os.Chmod(dst, 0o644)
	if body, _ := os.ReadFile(dst); string(body) != "existing" {
		t.Errorf("the destination was overwritten: %q", body)
	}
}

// TestList_UnreadableMarkerDirectory_IsAnError — list.go sweep: a home projects
// directory that cannot be read is not a home with no markers; every moved tree
// would list as unclassified without a word.
func TestList_UnreadableMarkerDirectory_IsAnError(t *testing.T) {
	t2053SkipIfRoot(t)
	w := m6NewWorld(t)
	projects := homeProjectsDir()
	t2053Chmod(t, projects, 0o000, 0o755)
	if _, err := buildListRows(w.repo); err == nil {
		t.Errorf("buildListRows classified the trees without the markers it could not read")
	}
}

// TestResolveTreePath_OnlyAbsenceKeepsTheSpelling — planner sweep: a path that
// is absent resolves to its cleaned spelling; one that cannot be resolved for any
// other reason is an error (the planner then skips the tree saying so, instead of
// comparing an unresolved spelling). The planner-level reproduction is shadowed by
// git's own "prunable" verdict for a tree behind an unsearchable directory, so
// the helper is tested directly.
func TestResolveTreePath_OnlyAbsenceKeepsTheSpelling(t *testing.T) {
	t2053SkipIfRoot(t)
	root := t.TempDir()
	parent := filepath.Join(root, "p")
	if err := os.MkdirAll(filepath.Join(parent, "tree"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Run("present", func(t *testing.T) {
		got, err := resolveTreePath(filepath.Join(parent, "tree"))
		if err != nil || filepath.Base(got) != "tree" {
			t.Errorf("got %q, %v", got, err)
		}
	})
	t.Run("absent keeps the cleaned spelling", func(t *testing.T) {
		missing := filepath.Join(root, "nope", "..", "gone")
		got, err := resolveTreePath(missing)
		if err != nil || got != filepath.Join(root, "gone") {
			t.Errorf("got %q, %v", got, err)
		}
	})
	t.Run("unsearchable parent is an error", func(t *testing.T) {
		t2053Chmod(t, parent, 0o000, 0o755)
		if got, err := resolveTreePath(filepath.Join(parent, "tree")); err == nil {
			t.Errorf("resolved to %q instead of failing", got)
		}
	})
}

// TestManifest_UnreadableReportsTree_IsNotACount — sweep: the reports walk
// ignored its errors, so a subdirectory it could not read counted as empty.
// Both manifest algorithms now refuse to produce a digest over it; an absent
// reports directory is still just zero files.
func TestManifest_UnreadableReportsTree_IsNotACount(t *testing.T) {
	t2053SkipIfRoot(t)
	w := m6NewWorld(t)
	tree := w.movedTree
	if _, err := captureManifestChecked(tree); err != nil {
		t.Fatalf("control (no reports directory): %v", err)
	}
	sub := filepath.Join(tree, ".moai", "reports", "card")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := captureManifestChecked(tree); err != nil {
		t.Fatalf("control (readable reports): %v", err)
	}
	t2053Chmod(t, sub, 0o000, 0o755)
	if _, err := captureManifestChecked(tree); err == nil {
		t.Errorf("captureManifestChecked digested a reports tree it could not read")
	}
	if got := captureLegacyManifest(tree); got != "" {
		t.Errorf("captureLegacyManifest produced %q over a reports tree it could not read", got)
	}
}

// TestPlan_UnreadableMoveList_IsARunError — planner sweep: a move list that
// exists but cannot be read means no tree can be judged interrupted or not; the
// run fails with the error instead of planning against an empty list.
func TestPlan_UnreadableMoveList_IsARunError(t *testing.T) {
	t2053SkipIfRoot(t)
	w := t2053TwoTreeWorld(t)
	if err := os.MkdirAll(t2053StateDir(w.repo), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(moveListPath(w.repo), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t2053Chmod(t, moveListPath(w.repo), 0o000, 0o644)
	out, err := m4RunMigrate(t)
	if err == nil || !strings.Contains(err.Error(), "move list") {
		t.Errorf("err = %v, want a move-list read error\n%s", err, out)
	}
}

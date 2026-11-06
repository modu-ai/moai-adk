package worktree

// migrate_hardening_d_t2053_test.go — card t2053, fifth review round.
//
//	K1  journals written in the same second sort by creation order, not by name
//	K2  protection is released only when the tree's location is proven

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestOrderedJournals_CollisionSuffixSortsAfterTheBareName — K1: `<stamp>-1`
// was created after `<stamp>`, whatever the lexical order says.
func TestOrderedJournals_CollisionSuffixSortsAfterTheBareName(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"20260101T000000-10.json", "20260101T000000-2.json", "20260101T000000.json", "20260101T000001.json", "move-list.jsonl", "lock"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}
	got, err := orderedJournals(dir)
	if err != nil {
		t.Fatalf("orderedJournals: %v", err)
	}
	var names []string
	for _, p := range got {
		names = append(names, filepath.Base(p))
	}
	want := "20260101T000000.json,20260101T000000-2.json,20260101T000000-10.json,20260101T000001.json"
	if strings.Join(names, ",") != want {
		t.Errorf("order = %s\nwant    %s", strings.Join(names, ","), want)
	}
}

func t2053WriteStepJournal(t *testing.T, path string, steps []journalStep) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"steps": steps})
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}
}

// TestJournalRecording_PicksTheSecondJournalOfTheSameSecond — K1.
func TestJournalRecording_PicksTheSecondJournalOfTheSameSecond(t *testing.T) {
	w, p, id, _ := t2053ReverseWorld(t)
	dir := t2053StateDir(w.repo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	step := []journalStep{{OldPath: p.source, NewPath: p.destination, Outcome: "pending", Ident: id.Ident()}}
	first, second := filepath.Join(dir, "20260101T000000.json"), filepath.Join(dir, "20260101T000000-1.json")
	t2053WriteStepJournal(t, first, step)
	t2053WriteStepJournal(t, second, step)

	if got := journalRecording(w.repo, id.Ident()); got != second {
		t.Errorf("journalRecording = %s, want the newer %s", got, second)
	}
}

// TestVerify_PicksTheSecondJournalOfTheSameSecond — K1: the older journal
// still says `pending`; the newer one says the move settled. The newer wins.
func TestVerify_PicksTheSecondJournalOfTheSameSecond(t *testing.T) {
	w, p, id, _ := t2053ReverseWorld(t)
	dir := t2053StateDir(w.repo)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t2053WriteStepJournal(t, filepath.Join(dir, "20260101T000000.json"),
		[]journalStep{{OldPath: p.source, NewPath: p.destination, Outcome: "pending", Ident: id.Ident()}})
	t2053WriteStepJournal(t, filepath.Join(dir, "20260101T000000-1.json"),
		[]journalStep{{OldPath: p.source, NewPath: p.destination, Outcome: "moved", Ident: id.Ident()}})

	out, err := m4RunMigrate(t, "--verify")
	if err != nil {
		t.Fatalf("verify judged the older journal: %v\n%s", err, out)
	}
}

// TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending —
// K2: the forward move errors after git renamed and registered the tree, and
// the destination's identity cannot be read at that moment (an unresolved
// relative pointer). The source is absent, so the location is uncertain: the
// tree stays pending and keeps its protection.
func TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending(t *testing.T) {
	w := t2053TwoTreeWorld(t)
	orig := forwardGitMove
	// The engine hands a batch of trees to this seam at the same time (card
	// t2081), so the list the seam appends to is guarded.
	var mu sync.Mutex
	var restore []func()
	forwardGitMove = func(root, from, to string) error {
		if err := orig(root, from, to); err != nil {
			return err
		}
		gitFile := filepath.Join(to, ".git")
		body, _ := os.ReadFile(gitFile)
		_ = os.WriteFile(gitFile, []byte("gitdir: \n"), 0o644)
		mu.Lock()
		restore = append(restore, func() { _ = os.WriteFile(gitFile, body, 0o644) })
		mu.Unlock()
		return errors.New("killed right after the rename")
	}
	t.Cleanup(func() { forwardGitMove = orig })

	report, err := t2053EngineApply(t, w.repo)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	for _, f := range restore {
		f()
	}
	if report.failed != 2 {
		t.Fatalf("failed = %d, want 2: %+v", report.failed, report.rows)
	}
	dest := filepath.Join(w.homeRoot, "legacy6")
	if got := t2053LastMoveListResult(t, moveListPath(w.repo)); got == "reversed" {
		t.Errorf("a `reversed` line was written for a tree whose location is unknown")
	}
	if !MovedBy(w.repo, dest) {
		t.Errorf("MovedBy released the protection of a tree whose location is unknown")
	}
}

// TestRollback_UnstatableDestination_IsNotTreatedAsAbsent — K2's audit: a
// destination that cannot be inspected is not an absent destination, so the
// rollback must not write `restored` on that reading.
func TestRollback_UnstatableDestination_IsNotTreatedAsAbsent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory modes")
	}
	w, p, id, mlPath := t2053ReverseWorld(t)
	m6Git(t, w.repo, "worktree", "move", p.destination, p.source) // really back, really registered
	journal := t2053RollbackJournal(t, p, id)
	parent := filepath.Dir(p.destination)
	if err := os.Chmod(parent, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil {
		t.Fatalf("rollback treated an uninspectable destination as absent:\n%s", out)
	}
	if got := t2053LastMoveListResult(t, mlPath); got == "restored" {
		t.Errorf("a `restored` line was written without knowing where the tree is")
	}
}

// TestRollback_UnstatableSource_IsNotTreatedAsGone — K2's audit: with the
// destination absent, a source that cannot be inspected is not a source that is
// gone; the step must not be skipped as "nothing to do".
func TestRollback_UnstatableSource_IsNotTreatedAsGone(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory modes")
	}
	w, p, id, _ := t2053ReverseWorld(t)
	m6Git(t, w.repo, "worktree", "move", p.destination, p.source)
	journal := t2053RollbackJournal(t, p, id)
	parent := filepath.Dir(p.source)
	if err := os.Chmod(parent, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

	out, rerr := m4RunMigrate(t, "--rollback", journal)
	if rerr == nil {
		t.Fatalf("rollback skipped a step whose source it could not inspect:\n%s", out)
	}
}

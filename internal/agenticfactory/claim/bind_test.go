package claim

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// two lanes are two stores over the same directory (two processes in practice)
func twoStores(t *testing.T) (*Store, *Store) {
	t.Helper()
	dir := t.TempDir()
	a, err := OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	b, err := OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	return a, b
}

// Two writers binding one path at the same moment: exactly one wins. The scan of
// the other records cannot exclude this alone (both scans run before either
// write), so the path is reserved through a create-only document write, which
// the store already makes atomic.
func TestConcurrentBindsOfOnePathHaveExactlyOneWinner(t *testing.T) {
	wt_shared := realWT(t, "shared")
	sa, sb := twoStores(t)
	ra, err := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	rb, err := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	// both writers stand between "no other record uses the path" and the write
	var arrive sync.WaitGroup
	arrive.Add(2)
	gate := make(chan struct{})
	hold := func() { arrive.Done(); <-gate }
	sa.afterScan, sb.afterScan = hold, hold
	go func() { arrive.Wait(); close(gate) }()

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); _, errs[0] = sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, wt_shared) }()
	go func() { defer wg.Done(); _, errs[1] = sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_shared) }()
	wg.Wait()

	won, lost := 0, 0
	for _, e := range errs {
		switch {
		case e == nil:
			won++
		case errors.Is(e, core.ErrWorktreeInUse):
			lost++
		default:
			t.Errorf("unexpected: %v", e)
		}
	}
	if won != 1 || lost != 1 {
		t.Fatalf("winners=%d losers=%d (%v)", won, lost, errs)
	}
	recs, _ := sa.List()
	bound := 0
	for _, r := range recs {
		if r.Worktree == canonOf(t, wt_shared) {
			bound++
		}
	}
	if bound != 1 {
		t.Errorf("%d records ended up bound to one path", bound)
	}
}

func TestRebindingTheSameIssueIsIdempotentAndFreesTheOldPath(t *testing.T) {
	wt_one := realWT(t, "one")
	wt_two := realWT(t, "two")
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, wt_one); err != nil {
		t.Fatal(err)
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, wt_one); err != nil {
		t.Errorf("binding the same path again: %v", err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_one); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("a bound path: %v", err)
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, wt_two); err != nil {
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_one); err != nil {
		t.Errorf("a path its Issue moved off is free again: %v", err)
	}
}

func TestAReservationOfADoneIssueIsTakenOver(t *testing.T) {
	wt_one := realWT(t, "one")
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, wt_one); err != nil {
		t.Fatal(err)
	}
	if _, err := sa.Update(ref(t, 1), "w1", ra.Epoch, func(r *Record) error { r.State = core.StateDone; r.Closed = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_one); err != nil {
		t.Errorf("the reservation of a finished Issue must not hold the path: %v", err)
	}
}

// Between the reservation and the write of the record the Issue is not yet bound
// to the path, yet it must still hold it: otherwise a second lane would read the
// reservation as stale and take it over.
func TestAReservationNotYetWrittenToItsRecordStillHoldsThePath(t *testing.T) {
	wt_mid := realWT(t, "mid")
	sa, sb := twoStores(t)
	if _, err := sa.Acquire(ref(t, 1), "w1", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if err := sa.reserveWorktree(ref(t, 1), wt_mid); err != nil { // issue 1 reserved, record not updated
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_mid); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("the half-done bind lost its path: %v", err)
	}
}

// A handoff reserves the receipt's worktree through the same path as a bind, so a
// handoff and a bind of one path at the same moment have exactly one winner.
func TestAHandoffReservationAndABindOfOnePathHaveExactlyOneWinner(t *testing.T) {
	wt_pr := realWT(t, "pr")
	sa, sb := twoStores(t)
	if _, err := sa.Acquire(ref(t, 1), "w1", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	var arrive sync.WaitGroup
	arrive.Add(2)
	gate := make(chan struct{})
	hold := func() { arrive.Done(); <-gate }
	sa.afterScan, sb.afterScan = hold, hold
	go func() { arrive.Wait(); close(gate) }()
	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); errs[0] = sa.ReserveWorktree(ref(t, 1), wt_pr) }()
	go func() { defer wg.Done(); _, errs[1] = sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, wt_pr) }()
	wg.Wait()
	won, lost := 0, 0
	for _, e := range errs {
		switch {
		case e == nil:
			won++
		case errors.Is(e, core.ErrWorktreeInUse):
			lost++
		default:
			t.Errorf("unexpected: %v", e)
		}
	}
	if won != 1 || lost != 1 {
		t.Fatalf("winners=%d losers=%d (%v)", won, lost, errs)
	}
}

// An Issue works in one worktree and hands its PR off from another: both paths
// are in use, and only a finished Issue lets either go.
func TestAReceiptPathIsInUseEvenWhenTheIssueIsBoundElsewhere(t *testing.T) {
	wt_a := realWT(t, "a")
	wt_b := realWT(t, "b")
	wt_other := realWT(t, "other")
	s := open(t)
	rec, err := s.Acquire(ref(t, 1), "w1", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ref(t, 1), "w1", rec.Epoch, func(r *Record) error {
		r.Worktree = wt_a
		r.Receipt = &Receipt{HandoffID: "ho-1", PR: 1, Head: core.HeadSHA("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"), Branch: "WT-x", Worktree: wt_b}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !s.usesWorktree(worktreeReservation{Path: canonOf(t, wt_b), Issue: 1}) || !s.usesWorktree(worktreeReservation{Path: canonOf(t, wt_a), Issue: 1}) {
		t.Error("both the bound and the receipt path are in use")
	}
	if s.usesWorktree(worktreeReservation{Path: canonOf(t, wt_other), Issue: 1}) {
		t.Error("a path the Issue neither uses nor reserved is not in use")
	}
}

// realWT makes a worktree directory that exists, as a bind requires.
func realWT(t *testing.T, name string) string {
	t.Helper()
	d := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func canonOf(t *testing.T, p string) string {
	t.Helper()
	c, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Two spellings of one directory are one worktree: a symlink, a trailing slash, a
// dot segment. A path that cannot be resolved, or is not absolute, is refused and
// reserves nothing.
func TestOneWorktreeInDifferentSpellingsIsOneWorktree(t *testing.T) {
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	real := realWT(t, "tree")
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, real); err != nil {
		t.Fatal(err)
	}
	for name, p := range map[string]string{
		"symlink":        link,
		"trailing slash": real + string(filepath.Separator),
		"dot segment":    filepath.Join(real, "sub", ".."),
	} {
		if name == "dot segment" {
			if err := os.MkdirAll(filepath.Join(real, "sub"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, p); !errors.Is(err, core.ErrWorktreeInUse) {
			t.Errorf("%s spelling of a bound worktree: %v", name, err)
		}
	}
	rec, _ := sa.Read(ref(t, 1))
	if rec.Worktree != canonOf(t, real) {
		t.Errorf("the record keeps the canonical path: %q", rec.Worktree)
	}
}

func TestAPathThatCannotBeResolvedIsRefusedAndReservesNothing(t *testing.T) {
	s := open(t)
	rec, _ := s.Acquire(ref(t, 1), "w1", "d1", "run-1")
	missing := filepath.Join(t.TempDir(), "gone")
	for name, p := range map[string]string{"missing": missing, "relative": "wt/rel", "dot": "."} {
		if _, err := s.BindWorktree(ref(t, 1), "w1", rec.Epoch, p); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := s.GetDoc("worktree", WorktreeKey(missing)); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("a refused path left a reservation: %v", err)
	}
	if cur, _ := s.Read(ref(t, 1)); cur.Worktree != "" {
		t.Errorf("a refused path was recorded: %q", cur.Worktree)
	}
}

// On a case-insensitive filesystem (the macOS default) "Tree" and "tree" are one
// directory. The tree is identified by the file itself, not by its spelling.
func TestCaseOnlyDifferentSpellingsOfOneTreeAreOneWorktree(t *testing.T) {
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	real := realWT(t, "CaseTree")
	alt := filepath.Join(filepath.Dir(real), "casetree")
	a, errA := os.Stat(real)
	b, errB := os.Stat(alt)
	if errA != nil || errB != nil || !os.SameFile(a, b) {
		t.Skip("this filesystem tells CaseTree and casetree apart; nothing to measure")
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, real); err != nil {
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, alt); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("the same directory spelled in another case: %v", err)
	}
	if WorktreeKey(real) != WorktreeKey(alt) {
		t.Errorf("one directory, two reservation keys: %s %s", WorktreeKey(real), WorktreeKey(alt))
	}
}

// sameWorktree decides by the file, not by a name: two directories that happen to
// share a base name are different trees, and one tree spelled two ways is one.
func TestSameWorktreeComparesTheFileNotTheName(t *testing.T) {
	a := realWT(t, "same")
	b := realWT(t, "same") // another parent, the same base name
	if sameWorktree(a, canonOf(t, b)) {
		t.Error("two different directories with one base name were taken for one tree")
	}
	if !sameWorktree(a, canonOf(t, a)) {
		t.Error("a directory is not the same tree as itself")
	}
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(a, link); err != nil {
		t.Fatal(err)
	}
	if !sameWorktree(link, canonOf(t, a)) {
		t.Error("a symlink to a tree is that tree")
	}
	if sameWorktree("", canonOf(t, a)) {
		t.Error("an empty stored path is no tree")
	}
}

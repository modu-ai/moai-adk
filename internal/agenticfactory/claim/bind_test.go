package claim

import (
	"errors"
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
	go func() { defer wg.Done(); _, errs[0] = sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, "/wt/shared") }()
	go func() { defer wg.Done(); _, errs[1] = sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/shared") }()
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
		if r.Worktree == "/wt/shared" {
			bound++
		}
	}
	if bound != 1 {
		t.Errorf("%d records ended up bound to one path", bound)
	}
}

func TestRebindingTheSameIssueIsIdempotentAndFreesTheOldPath(t *testing.T) {
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, "/wt/one"); err != nil {
		t.Fatal(err)
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, "/wt/one"); err != nil {
		t.Errorf("binding the same path again: %v", err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/one"); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("a bound path: %v", err)
	}
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, "/wt/two"); err != nil {
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/one"); err != nil {
		t.Errorf("a path its Issue moved off is free again: %v", err)
	}
}

func TestAReservationOfADoneIssueIsTakenOver(t *testing.T) {
	sa, sb := twoStores(t)
	ra, _ := sa.Acquire(ref(t, 1), "w1", "d1", "run-1")
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if _, err := sa.BindWorktree(ref(t, 1), "w1", ra.Epoch, "/wt/one"); err != nil {
		t.Fatal(err)
	}
	if _, err := sa.Update(ref(t, 1), "w1", ra.Epoch, func(r *Record) error { r.State = core.StateDone; r.Closed = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/one"); err != nil {
		t.Errorf("the reservation of a finished Issue must not hold the path: %v", err)
	}
}

// Between the reservation and the write of the record the Issue is not yet bound
// to the path, yet it must still hold it: otherwise a second lane would read the
// reservation as stale and take it over.
func TestAReservationNotYetWrittenToItsRecordStillHoldsThePath(t *testing.T) {
	sa, sb := twoStores(t)
	if _, err := sa.Acquire(ref(t, 1), "w1", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	rb, _ := sb.Acquire(ref(t, 2), "w2", "d2", "run-1")
	if err := sa.reserveWorktree(ref(t, 1), "/wt/mid"); err != nil { // issue 1 reserved, record not updated
		t.Fatal(err)
	}
	if _, err := sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/mid"); !errors.Is(err, core.ErrWorktreeInUse) {
		t.Errorf("the half-done bind lost its path: %v", err)
	}
}

// A handoff reserves the receipt's worktree through the same path as a bind, so a
// handoff and a bind of one path at the same moment have exactly one winner.
func TestAHandoffReservationAndABindOfOnePathHaveExactlyOneWinner(t *testing.T) {
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
	go func() { defer wg.Done(); errs[0] = sa.ReserveWorktree(ref(t, 1), "/wt/pr") }()
	go func() { defer wg.Done(); _, errs[1] = sb.BindWorktree(ref(t, 2), "w2", rb.Epoch, "/wt/pr") }()
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
	s := open(t)
	rec, err := s.Acquire(ref(t, 1), "w1", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ref(t, 1), "w1", rec.Epoch, func(r *Record) error {
		r.Worktree = "/wt/a"
		r.Receipt = &Receipt{HandoffID: "ho-1", PR: 1, Head: core.HeadSHA("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"), Branch: "WT-x", Worktree: "/wt/b"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !s.usesWorktree(worktreeReservation{Path: "/wt/b", Issue: 1}) || !s.usesWorktree(worktreeReservation{Path: "/wt/a", Issue: 1}) {
		t.Error("both the bound and the receipt path are in use")
	}
	if s.usesWorktree(worktreeReservation{Path: "/wt/other", Issue: 1}) {
		t.Error("a path the Issue neither uses nor reserved is not in use")
	}
}

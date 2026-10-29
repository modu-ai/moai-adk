package e2e

import (
	"errors"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Scenarios 1 and 2 bound to the real lane-18 claim store (not the model).

func realStore(t *testing.T) (*claim.Store, core.IssueRef, core.RunID) {
	t.Helper()
	st, err := claim.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ref, err := core.NewIssueRef(1948)
	if err != nil {
		t.Fatal(err)
	}
	run, err := core.ParseRunID("run-1")
	if err != nil {
		t.Fatal(err)
	}
	return st, ref, run
}

// Scenario 1 (real): concurrent claimants, exactly one record is created and
// every other claimant gets Held or Busy (retryable), never a second claim.
func TestReal_CompetingClaim_ExactlyOneHolder(t *testing.T) {
	st, ref, run := realStore(t)
	const n = 8
	errs := make([]error, n)
	recs := make([]claim.Record, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			recs[i], errs[i] = st.Acquire(ref, "s-"+string(rune('a'+i)), "d-"+string(rune('a'+i)), run)
		})
	}
	wg.Wait()

	wins := 0
	for i, err := range errs {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, core.ErrHeld), errors.Is(err, core.ErrBusy):
		default:
			t.Errorf("claimant %d: unexpected error %v", i, err)
		}
	}
	if wins != 1 {
		t.Fatalf("winners = %d, want 1", wins)
	}
	cur, err := st.Load(ref)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Epoch != 1 {
		t.Errorf("epoch = %d, want 1", cur.Epoch)
	}
	// A later, uncontended claimant is Held, never Busy.
	if _, err := st.Acquire(ref, "late", "d-late", run); !errors.Is(err, core.ErrHeld) {
		t.Errorf("late claimant: %v, want ErrHeld", err)
	}
}

// Scenario 2 (real): the same dispatch id again is a duplicate, not a new claim.
func TestReal_DuplicateDispatch_NotReclaimed(t *testing.T) {
	st, ref, run := realStore(t)
	first, err := st.Acquire(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if _, err := st.Acquire(ref, "w1", "d-1", run); !errors.Is(err, core.ErrDuplicateDispatch) {
			t.Fatalf("repeat dispatch: %v, want ErrDuplicateDispatch", err)
		}
	}
	cur, err := st.Load(ref)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Epoch != first.Epoch || cur.Seq != first.Seq {
		t.Errorf("record changed by duplicates: epoch %d->%d seq %d->%d", first.Epoch, cur.Epoch, first.Seq, cur.Seq)
	}
}

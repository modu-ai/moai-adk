package e2e

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// DELTA4 additions on the real lane-18 code.

// Leader restart: a new Store over the same directory resumes the claim and
// its state; nothing lives only in the old process.
func TestReal_LeaderRestart_StateSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	st, err := claim.OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	ref, _ := core.NewIssueRef(1948)
	run, _ := core.ParseRunID("run-1")
	r := wired(st)
	tk, err := r.Claim(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	upTo(t, r, tk, core.StateReady)
	before, err := st.Load(ref)
	if err != nil {
		t.Fatal(err)
	}

	reopened, err := claim.OpenWithLimit(dir, 100)
	if err != nil {
		t.Fatal(err)
	}
	after, err := reopened.Load(ref)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Errorf("record changed across reopen: %+v -> %+v", before, after)
	}
	// The original ticket keeps working against the reopened store.
	rn := wired(reopened)
	if _, err := rn.PreflightPR(tk, headA, repoA, "Refs #"+issueNumber(tk)); err != nil {
		t.Errorf("preflight after reopen: %v", err)
	}
	if _, err := rn.Land(tk, headA); err != nil {
		t.Errorf("land after reopen: %v", err)
	}
}

// Duplicate handoff: completing twice never double-completes.
func TestReal_DuplicateHandoff_SecondCompleteRefused(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	l := fullLanding
	first, err := r.Complete(tk, l)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Complete(tk, l); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("second complete: %v, want ErrTransitionInvalid", err)
	}
	if got := state(t, r, tk); got.Seq != first.Seq {
		t.Errorf("seq moved on a refused duplicate: %d -> %d", first.Seq, got.Seq)
	}
}

// Repair contention: after the claim is released and re-acquired, the old
// ticket cannot write; only the current epoch's holder can.
func TestReal_RepairContention_StaleTicketRefused(t *testing.T) {
	st, ref, run := realStore(t)
	r := wired(st)
	old, err := r.Claim(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Release(ref, "w1"); err != nil {
		t.Fatal(err)
	}
	cur, err := r.Claim(ref, "w2", "d-2", run)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Epoch <= old.Epoch {
		t.Fatalf("epoch did not increase: %d -> %d", old.Epoch, cur.Epoch)
	}
	if _, err := r.StartAttempt(old, "att-old", core.DeliveryExecuted); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("stale ticket: %v, want ErrOwnershipUnverified", err)
	}
	if _, err := r.StartAttempt(cur, "att-new", core.DeliveryExecuted); err != nil {
		t.Errorf("current ticket refused: %v", err)
	}
}

// Repair contention, simultaneous: two writers on the same ticket never both
// advance the same transition.
func TestReal_RepairContention_ConcurrentSameTransition_OneWins(t *testing.T) {
	r, tk := newRunner(t)
	const n = 6
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			_, errs[i] = r.StartAttempt(tk, "att-1", core.DeliveryExecuted)
		})
	}
	wg.Wait()
	wins := 0
	for _, err := range errs {
		if err == nil {
			wins++
		}
	}
	if wins != 1 {
		t.Errorf("successful StartAttempt calls = %d, want 1 (errors: %v)", wins, errs)
	}
}

// PR issued is not done: reaching landing (the PR is out) never completes the
// Issue; only the three records do.
func TestReal_PRIssued_IsNotDone(t *testing.T) {
	r, tk := newRunner(t)
	upTo(t, r, tk, core.StateLanding)
	if got := state(t, r, tk).State; got == core.StateDone {
		t.Fatal("landing reached done without any landing record")
	}
	if _, err := r.Complete(tk, runner.Landing{}); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("complete with no records: %v", err)
	}
}

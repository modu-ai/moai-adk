package claim

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func ref(t *testing.T, n int) core.IssueRef {
	t.Helper()
	r, err := core.NewIssueRef(n)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAcquireStartsQueuedToClaimed(t *testing.T) {
	s := open(t)
	rec, err := s.Acquire(ref(t, 1948), "sess-a", "disp-1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Holder != "sess-a" || rec.State != core.StateClaimed || rec.Epoch != 1 || rec.Seq != 1 {
		t.Errorf("unexpected record: %+v", rec)
	}
	got, err := s.Load(ref(t, 1948))
	if err != nil || got.Dispatch != "disp-1" {
		t.Errorf("load: %+v %v", got, err)
	}
}

func TestAcquireByOtherIsHeld(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Acquire(ref(t, 1), "b", "d2", "run-2")
	if !errors.Is(err, core.ErrHeld) {
		t.Errorf("want ErrHeld, got %v", err)
	}
}

func TestSameDispatchTwiceIsDuplicate(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Acquire(ref(t, 1), "a", "d1", "run-1")
	if !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("want ErrDuplicateDispatch, got %v", err)
	}
	// Same holder, different dispatch: still the holder's, not a new claim.
	_, err = s.Acquire(ref(t, 1), "a", "d2", "run-1")
	if !errors.Is(err, core.ErrHeld) {
		t.Errorf("want ErrHeld for new dispatch on owned issue, got %v", err)
	}
}

func TestConcurrentAcquireExactlyOneWins(t *testing.T) {
	s := open(t)
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			holder := string(rune('a' + i))
			if _, err := s.Acquire(ref(t, 7), holder, "d-"+holder, "run-1"); err == nil {
				wins.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Errorf("want exactly 1 winner, got %d", wins.Load())
	}
}

func TestVerifyRefusesWrongHolderEpochAndMissing(t *testing.T) {
	s := open(t)
	if _, err := s.Verify(ref(t, 1), "a", 1); !errors.Is(err, core.ErrOwnershipUnverified) || !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing claim: %v", err)
	}
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Verify(ref(t, 1), "b", 1); !errors.Is(err, core.ErrOwnershipUnverified) || !errors.Is(err, core.ErrForeign) {
		t.Errorf("wrong holder: %v", err)
	}
	if _, err := s.Verify(ref(t, 1), "a", 2); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("wrong epoch: %v", err)
	}
	if _, err := s.Verify(ref(t, 1), "a", 1); err != nil {
		t.Errorf("good holder: %v", err)
	}
}

func TestUpdateRefusedWithoutOwnershipAndLeavesRecordUntouched(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Update(ref(t, 1), "intruder", 1, func(r *Record) error {
		r.State = core.StateImplementing
		return nil
	})
	if !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("want ownership refusal, got %v", err)
	}
	got, _ := s.Load(ref(t, 1))
	if got.State != core.StateClaimed || got.Seq != 1 {
		t.Errorf("record changed by refused write: %+v", got)
	}
}

func TestUpdateBumpsSeqAndFnErrorAborts(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	rec, err := s.Update(ref(t, 1), "a", 1, func(r *Record) error { r.State = core.StateImplementing; return nil })
	if err != nil || rec.Seq != 2 || rec.State != core.StateImplementing {
		t.Fatalf("update: %+v %v", rec, err)
	}
	boom := errors.New("boom")
	if _, err := s.Update(ref(t, 1), "a", 1, func(r *Record) error { r.State = core.StateDone; return boom }); !errors.Is(err, boom) {
		t.Errorf("want boom, got %v", err)
	}
	got, _ := s.Load(ref(t, 1))
	if got.State != core.StateImplementing || got.Seq != 2 {
		t.Errorf("aborted update leaked: %+v", got)
	}
}

func TestReleaseOnlyByHolderThenReacquireBumpsEpoch(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(ref(t, 1), "b"); !errors.Is(err, core.ErrForeign) {
		t.Errorf("foreign release: %v", err)
	}
	if err := s.Release(ref(t, 1), "a"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(ref(t, 1), "a"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("double release: %v", err)
	}
	rec, err := s.Acquire(ref(t, 1), "b", "d2", "run-2")
	if err != nil || rec.Epoch != 2 {
		t.Errorf("reacquire epoch want 2: %+v %v", rec, err)
	}
}

func TestStoreUnavailableMapsToBlockReason(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	if _, err := s.Acquire(ref(t, 1), "a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issue-1.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.Verify(ref(t, 1), "a", 1)
	if !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("want ownership refusal on corrupt store, got %v", err)
	}
	if !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("corrupt store must say store-unavailable: %v", err)
	}
}

func TestBusyWhenLockHeld(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	if err := os.WriteFile(filepath.Join(dir, "issue-1.lock"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := s.Acquire(ref(t, 1), "a", "d1", "run-1")
	if !errors.Is(err, core.ErrBusy) {
		t.Errorf("want ErrBusy with held lock, got %v", err)
	}
}

func TestOpenRejectsEmptyDir(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Error("empty dir must be rejected")
	}
}

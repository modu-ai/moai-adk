package factory

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// SPEC-FACTORY-ATOMIC-LEASE-001 WM2 — the lock-scope primitive. These tests pin
// BacklogStore.WithLock and the LockedBacklog handle it lends: two mutations
// inside one section see each other, the lock is held across the callback and
// released on every exit, a relocated queue is refused as Mutate refuses it,
// and the handle's read never adopts (plan D1, mutant MU8).

var errWithLockProbe = errors.New("withlock probe: callback refused")

// wlItem returns a queued card for the fixtures below.
func wlItem(id, text string) BacklogItem {
	return BacklogItem{ID: id, Text: text, State: BacklogStateQueued}
}

// wlRunMutate runs the PUBLIC Mutate on a second store value in a goroutine and
// returns a channel carrying its result, so a test can ask whether it finished.
func wlRunMutate(path string, mutate func(*BacklogRecord) error) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- NewBacklogStore(path).Mutate(mutate)
	}()
	return done
}

func TestWithLockMutationsSeeEachOther(t *testing.T) {
	s := NewBacklogStore(filepath.Join(t.TempDir(), backlogFileName))
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	err := s.WithLock(func(l *LockedBacklog) error {
		if err := l.Mutate(func(rec *BacklogRecord) error {
			rec.Items = append(rec.Items, wlItem("t2", "first inside"))
			return nil
		}); err != nil {
			return err
		}
		mid, err := l.LoadPure()
		if err != nil {
			return err
		}
		if len(mid.Items) != 2 || mid.Items[1].Text != "first inside" {
			t.Errorf("LoadPure after the first Mutate = %+v; want the write visible", mid.Items)
		}
		return l.Mutate(func(rec *BacklogRecord) error {
			if len(rec.Items) != 2 {
				t.Errorf("second Mutate saw %d items; want the first write visible (2)", len(rec.Items))
			}
			rec.Items = append(rec.Items, wlItem("t3", "second inside"))
			return nil
		})
	})
	if err != nil {
		t.Fatalf("WithLock: %v", err)
	}
	got, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 3 || got.Items[2].Text != "second inside" {
		t.Fatalf("after WithLock: %+v; want three items with both writes applied", got.Items)
	}
}

func TestWithLockHoldsLockUntilFnReturns(t *testing.T) {
	path := filepath.Join(t.TempDir(), backlogFileName)
	s := NewBacklogStore(path)
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	// Positive control for the window: the same contender with no lock held
	// finishes well inside it.
	ctrl := wlRunMutate(path, func(rec *BacklogRecord) error { return nil })
	select {
	case err := <-ctrl:
		if err != nil {
			t.Fatalf("control Mutate with no lock held: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("control Mutate with no lock held did not finish in 2s; the 300ms window below would not discriminate")
	}

	var contender <-chan error
	err := s.WithLock(func(l *LockedBacklog) error {
		contender = wlRunMutate(path, func(rec *BacklogRecord) error {
			rec.Items = append(rec.Items, wlItem("t2", "contender"))
			return nil
		})
		select {
		case err := <-contender:
			t.Errorf("a public Mutate finished inside the section (err=%v); the lock must be held across fn", err)
		case <-time.After(300 * time.Millisecond):
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock: %v", err)
	}
	select {
	case err := <-contender:
		if err != nil {
			t.Fatalf("contender after release: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("contender did not proceed after the section returned")
	}
	got, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[1].Text != "contender" {
		t.Fatalf("contender's write not applied after release: %+v", got.Items)
	}
}

// wlAssertReleased fails when the queue lock is still held: a public Mutate
// must complete well inside the wait budget.
func wlAssertReleased(t *testing.T, path, after string) {
	t.Helper()
	select {
	case err := <-wlRunMutate(path, func(rec *BacklogRecord) error { return nil }):
		if err != nil {
			t.Fatalf("Mutate after %s: %v", after, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("lock still held 2s after %s", after)
	}
}

func TestWithLockReleasesOnErrorReturn(t *testing.T) {
	path := filepath.Join(t.TempDir(), backlogFileName)
	s := NewBacklogStore(path)
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	err := s.WithLock(func(l *LockedBacklog) error { return errWithLockProbe })
	if !errors.Is(err, errWithLockProbe) {
		t.Fatalf("WithLock err = %v; want the callback's error returned", err)
	}
	wlAssertReleased(t, path, "an error return")
}

func TestWithLockReleasesOnPanic(t *testing.T) {
	path := filepath.Join(t.TempDir(), backlogFileName)
	s := NewBacklogStore(path)
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("the callback's panic did not propagate out of WithLock")
			}
		}()
		_ = s.WithLock(func(l *LockedBacklog) error { panic("withlock probe: panic") })
	}()
	wlAssertReleased(t, path, "a panic")
}

func TestWithLockRefusesRelocatedQueue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, backlogFileName)
	s := NewBacklogStore(path)
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, backlogRetiredFileName), []byte("new-root"), 0o600); err != nil {
		t.Fatal(err)
	}
	ran := false
	err := s.WithLock(func(l *LockedBacklog) error {
		ran = true
		return nil
	})
	if !errors.Is(err, ErrBacklogRelocated) {
		t.Fatalf("WithLock on a relocated queue err = %v; want ErrBacklogRelocated, as Mutate refuses it", err)
	}
	if ran {
		t.Fatal("the callback ran against a relocated queue")
	}
	// The same refusal as the public Mutate, so the lease can tell it from `raced`.
	if mErr := s.Mutate(func(rec *BacklogRecord) error { return nil }); !errors.Is(mErr, ErrBacklogRelocated) {
		t.Fatalf("control: Mutate err = %v; want ErrBacklogRelocated", mErr)
	}
	lock, lErr := s.acquireLock()
	if lErr != nil {
		t.Fatalf("lock still held after a relocated refusal: %v", lErr)
	}
	if rErr := lock.Release(); rErr != nil {
		t.Fatal(rErr)
	}
}

func TestLockedBacklogMutateRefusalKeepsFileUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), backlogFileName)
	s := NewBacklogStore(path)
	if _, _, err := s.Add("seed"); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	wlErr := s.WithLock(func(l *LockedBacklog) error {
		return l.Mutate(func(rec *BacklogRecord) error {
			rec.Items = append(rec.Items, wlItem("t2", "refused"))
			return errWithLockProbe
		})
	})
	if !errors.Is(wlErr, errWithLockProbe) {
		t.Fatalf("err = %v; want the mutation's error wrapped", wlErr)
	}
	if !strings.Contains(wlErr.Error(), "mutation refused") {
		t.Fatalf("err = %q; want the Mutate refusal wording preserved", wlErr)
	}
	after, err := s.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.Items, after.Items) {
		t.Fatalf("a refused mutation changed the queue: before=%+v after=%+v", before.Items, after.Items)
	}
}

// TestLockedBacklogLoadIsPure pins AC-FAL-011's non-adopting half (mutant MU8):
// a queue in the legacy layout shows no layout change after a WithLock that
// only reads, because the handle offers LoadPure and no adopting Load.
func TestLockedBacklogLoadIsPure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, backlogFileName)
	raw, err := json.Marshal(&BacklogRecord{
		Version: backlogVersion,
		LastSeq: 1,
		Items:   []BacklogItem{wlItem("t1", "legacy card")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	s := NewBacklogStore(path)
	if layout := inspectBacklogLayout(path); layout.dbExists || !layout.jsonExists {
		t.Fatalf("fixture layout = %+v; want legacy JSON only", layout)
	}
	censusBefore := dirCensus(t, dir)

	var seen []BacklogItem
	if err := s.WithLock(func(l *LockedBacklog) error {
		rec, err := l.LoadPure()
		if err != nil {
			return err
		}
		seen = rec.Items
		return nil
	}); err != nil {
		t.Fatalf("WithLock: %v", err)
	}

	if len(seen) != 1 || seen[0].ID != "t1" {
		t.Fatalf("LoadPure inside the section = %+v; want the legacy card served", seen)
	}
	if layout := inspectBacklogLayout(path); layout.dbExists || !layout.jsonExists {
		t.Fatalf("layout after the section = %+v; the read adopted (migrated) the queue", layout)
	}
	// The lock sidecar the section took is the only permitted new entry.
	censusAfter := dirCensus(t, dir)
	for _, name := range censusAfter {
		if strings.HasSuffix(name, ".db") || strings.HasSuffix(name, "-wal") || strings.HasSuffix(name, "-shm") || name == backlogRetiredFileName {
			t.Fatalf("directory after the section = %v (before %v); a database or relocation artifact appeared", censusAfter, censusBefore)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, after) {
		t.Fatal("the legacy backlog.json bytes changed after a read-only section")
	}
}

func TestLockWaitBudgetIsTheQueueLockBudget(t *testing.T) {
	got := LockWaitBudget()
	if got <= 0 {
		t.Fatalf("LockWaitBudget() = %s; want the queue lock's positive wait budget", got)
	}
	if got != stateLockWaitBudget {
		t.Fatalf("LockWaitBudget() = %s; want stateLockWaitBudget %s (read, never copied)", got, stateLockWaitBudget)
	}
}

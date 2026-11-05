package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// releaseNow releases whatever epoch the holder currently has; for tests that are
// not about the epoch.
func releaseNow(t *testing.T, s *Store, i core.IssueRef, holder string) error {
	t.Helper()
	rec, err := s.Read(i)
	if err != nil {
		return err
	}
	return s.Release(i, holder, rec.Epoch)
}

// F14: a run of the same session that holds an older epoch must not release the
// newer claim that session took afterwards.
func TestReleaseNeedsTheCurrentEpochOfTheHolder(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	first, err := s.Acquire(i, "a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Release(i, "a", first.Epoch); err != nil {
		t.Fatal(err)
	}
	second, err := s.Acquire(i, "a", "d2", "run-2")
	if err != nil || second.Epoch <= first.Epoch {
		t.Fatalf("a new claim by the same session gets a later epoch: %+v %v", second, err)
	}
	// the older run, still holding epoch 1, tries to release
	if err := s.Release(i, "a", first.Epoch); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Fatalf("a stale epoch released the newer claim: %v", err)
	}
	if cur, _ := s.Read(i); cur.Holder != "a" || cur.Epoch != second.Epoch {
		t.Errorf("the claim moved: %+v", cur)
	}
	if err := s.Release(i, "a", second.Epoch+1); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("an epoch ahead of the record: %v", err)
	}
	if err := s.Release(i, "a", second.Epoch); err != nil {
		t.Errorf("the current epoch: %v", err)
	}
}

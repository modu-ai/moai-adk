package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestNewClaimRefusedWhileLimitIsUnset(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Acquire(ref(t, 1), "w", "d", "run-1")
	if !errors.Is(err, core.ErrBacklogUnset) {
		t.Fatalf("want ErrBacklogUnset, got %v", err)
	}
	if r, ok := core.ReasonFor(err); !ok || r != core.ReasonNotConfigured {
		t.Errorf("reason = %q ok=%v", r, ok)
	}
	if _, err := s.Load(ref(t, 1)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("refused claim left a record")
	}
	// Reading stays possible without a limit: a monitor opens the store this way.
	if recs, err := s.List(); err != nil || len(recs) != 0 {
		t.Errorf("list: %+v %v", recs, err)
	}
}

func TestLimitBelowOneIsRejectedNotClamped(t *testing.T) {
	s, _ := OpenWithLimit(t.TempDir(), 5)
	for _, bad := range []int{0, -1, -100} {
		if err := s.SetMaxOpenPRs(bad); !errors.Is(err, core.ErrLimitInvalid) {
			t.Errorf("limit %d: %v", bad, err)
		}
		if s.MaxOpenPRs() != 5 {
			t.Fatalf("a rejected value changed the limit to %d", s.MaxOpenPRs())
		}
	}
	if _, err := OpenWithLimit(t.TempDir(), 0); !errors.Is(err, core.ErrLimitInvalid) {
		t.Errorf("OpenWithLimit(0): %v", err)
	}
	// No upper bound is invented: a large value is stored exactly.
	if err := s.SetMaxOpenPRs(1 << 30); err != nil || s.MaxOpenPRs() != 1<<30 {
		t.Errorf("large limit: %d %v", s.MaxOpenPRs(), err)
	}
}

func TestRepairClaimDoesNotNeedALimit(t *testing.T) {
	full, _ := OpenWithLimit(t.TempDir(), 5)
	i := reopened(t, full, 5, "WT-b")
	// A store reopened read-only (no limit) can still admit the repair claim.
	ro, _ := Open(full.dir)
	if _, err := ro.AcquireRepair(i, "r", "dr", "run-2", "WT-b"); err != nil {
		t.Errorf("repair needs no backlog limit: %v", err)
	}
}

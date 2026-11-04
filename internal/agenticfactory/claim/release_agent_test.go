package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestReleaseAgentRefusedWhileHoldingAClaim(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	i := ref(t, 10)
	rec, _ := s.AcquireFor(i, a, "d", "run-1")
	if err := s.ReleaseAgent(a.SessionID, a.Number); !errors.Is(err, core.ErrHeld) {
		t.Fatalf("release with live claim: %v", err)
	}
	if _, _, err := s.ResumeAgent(a.SessionID, a.Number); err != nil {
		t.Errorf("identity must survive a refused release: %v", err)
	}
	if err := releaseNow(t, s, i, a.SessionID); err != nil {
		t.Fatal(err)
	}
	_ = rec
	if err := s.ReleaseAgent(a.SessionID, a.Number); err != nil {
		t.Errorf("release after claim ended: %v", err)
	}
}

func TestReleaseAgentOnlyByTheRegisteredSession(t *testing.T) {
	s := open(t)
	a, _ := s.RegisterAgent(sid(1))
	if err := s.ReleaseAgent(sid(2), a.Number); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("foreign release: %v", err)
	}
	if err := s.ReleaseAgent(a.SessionID, 9); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("unknown number: %v", err)
	}
	if _, _, err := s.ResumeAgent(a.SessionID, a.Number); err != nil {
		t.Errorf("identity damaged by refused release: %v", err)
	}
}

func TestReleasedNumberIsReusedAndOldSessionCannotResume(t *testing.T) {
	s := open(t)
	old, _ := s.RegisterAgent(sid(1))
	keep, _ := s.RegisterAgent(sid(2))
	if err := s.ReleaseAgent(old.SessionID, old.Number); err != nil {
		t.Fatal(err)
	}
	// Before anyone takes the number the old session has nothing to resume.
	if _, _, err := s.ResumeAgent(old.SessionID, old.Number); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("resume of released number: %v", err)
	}
	fresh, err := s.RegisterAgent(sid(3))
	if err != nil || fresh.Number != old.Number {
		t.Fatalf("lowest free number must be reused: %+v %v", fresh, err)
	}
	if _, _, err := s.ResumeAgent(old.SessionID, old.Number); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("old session resuming a reused number: %v", err)
	}
	if got, _, err := s.ResumeAgent(fresh.SessionID, fresh.Number); err != nil || got != fresh {
		t.Errorf("new owner resume: %+v %v", got, err)
	}
	if got, _, err := s.ResumeAgent(keep.SessionID, keep.Number); err != nil || got != keep {
		t.Errorf("untouched agent: %+v %v", got, err)
	}
	// The old session may register again and gets a different number.
	again, err := s.RegisterAgent(old.SessionID)
	if err != nil || again.Number == fresh.Number || again.Number == keep.Number {
		t.Errorf("re-register after release: %+v %v", again, err)
	}
}

func TestReleasedOrReusedIdentityCannotClaim(t *testing.T) {
	s := open(t)
	old, _ := s.RegisterAgent(sid(1))
	if err := s.ReleaseAgent(old.SessionID, old.Number); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireFor(ref(t, 1), old, "d1", "run-1"); !errors.Is(err, core.ErrIdentityMismatch) && !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("released identity claimed: %v", err)
	}
	fresh, _ := s.RegisterAgent(sid(2))
	if _, err := s.AcquireFor(ref(t, 1), old, "d1", "run-1"); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("stale identity after reuse: %v", err)
	}
	if _, err := s.Load(ref(t, 1)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("refused claim left a record")
	}
	if _, err := s.AcquireFor(ref(t, 1), fresh, "d2", "run-1"); err != nil {
		t.Errorf("new owner claim: %v", err)
	}
}

func TestReleasedIdentityCannotRepairEither(t *testing.T) {
	s := open(t)
	old, _ := s.RegisterAgent(sid(1))
	i := reopened(t, s, 5, "WT-b")
	_ = s.ReleaseAgent(old.SessionID, old.Number)
	if _, err := s.AcquireRepairFor(i, old, "dr", "run-2", "WT-b"); err == nil {
		t.Error("released identity took a repair claim")
	}
}

func TestManyReleaseAndReregisterKeepsNumbersUnique(t *testing.T) {
	s := open(t)
	var first []core.Agent
	for k := 1; k <= 8; k++ {
		a, _ := s.RegisterAgent(sid(k))
		first = append(first, a)
	}
	for _, a := range first {
		if err := s.ReleaseAgent(a.SessionID, a.Number); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[int]bool{}
	for k := 101; k <= 108; k++ {
		a, err := s.RegisterAgent(sid(k))
		if err != nil || a.Number < 1 || a.Number > 8 || seen[a.Number] {
			t.Fatalf("reuse: %+v %v seen=%v", a, err, seen)
		}
		seen[a.Number] = true
	}
}

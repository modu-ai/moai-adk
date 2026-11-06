package claim

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// With a control ref attached, an agent takes a new Issue only when every
// unfinished Issue it already holds has its responsibility on the ref (REQ C4).
// A local record the ref does not carry (written before the ref was attached, or
// held under another holder there) stops the new claim.
func TestANewClaimNeedsEveryHeldIssueRecordedOnTheRef(t *testing.T) {
	m := NewMemoryRef()
	s := open(t)
	if _, err := s.Acquire(ref(t, 1), "w", "d1", "run-1"); err != nil { // before the ref exists
		t.Fatal(err)
	}
	s.Attach(m)
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); !errors.Is(err, core.ErrResponsibilityNotRecorded) {
		t.Fatalf("issue 1 is held but not on the ref: %v", err)
	}
	if _, err := s.Read(ref(t, 2)); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a refused claim left a record")
	}
	// another holder's claim is not held back by w's unrecorded Issue
	if _, err := s.Acquire(ref(t, 3), "other", "d3", "run-1"); err != nil {
		t.Errorf("a different holder: %v", err)
	}
	// once the record is on the ref (any write through the attached store puts it there)
	cur, _ := s.Read(ref(t, 1))
	if _, err := s.Update(ref(t, 1), "w", cur.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); err != nil {
		t.Errorf("issue 1 is on the ref now: %v", err)
	}
}

func TestADoneIssueAndALocalOnlyStoreAreNotChecked(t *testing.T) {
	// no control ref: the check is skipped on purpose (there is nothing to record on)
	local := open(t)
	if _, err := local.Acquire(ref(t, 1), "w", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := local.Acquire(ref(t, 2), "w", "d2", "run-1"); err != nil {
		t.Errorf("a store without a control ref: %v", err)
	}
	// a finished Issue holds no responsibility
	s := open(t)
	rec, _ := s.Acquire(ref(t, 1), "w", "d1", "run-1")
	if _, err := s.Update(ref(t, 1), "w", rec.Epoch, func(r *Record) error { finish(r); return nil }); err != nil {
		t.Fatal(err)
	}
	s.Attach(NewMemoryRef())
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); err != nil {
		t.Errorf("a done Issue is not a responsibility: %v", err)
	}
}

// The ref must carry this holder and epoch; the same key held by someone else on
// the ref is not this agent's responsibility being recorded.
func TestTheRefMustNameThisHolder(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	if _, err := s.Acquire(ref(t, 1), "w", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	data, rev, _ := m.Read(context.Background(), "issue/1")
	var onRef Record
	_ = jsonUnmarshal(data, &onRef)
	onRef.Holder = "someone-else"
	moved, _ := json.Marshal(onRef)
	if _, err := m.CompareAndSwap(context.Background(), "issue/1", rev, moved); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); !errors.Is(err, core.ErrResponsibilityNotRecorded) {
		t.Errorf("the ref names another holder: %v", err)
	}
}

// A ref that cannot be read does not say yes: the claim fails.
type downRef struct{ ControlRef }

func (downRef) Read(context.Context, string) ([]byte, int64, error) {
	return nil, 0, errors.New("ref unreachable")
}

func TestAnUnreadableRefFailsTheClaim(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	if _, err := s.Acquire(ref(t, 1), "w", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	s.Attach(downRef{m})
	_, err := s.Acquire(ref(t, 2), "w", "d2", "run-1")
	if err == nil || errors.Is(err, core.ErrResponsibilityNotRecorded) {
		t.Fatalf("an unreadable ref must fail the claim with the store error, got %v", err)
	}
	if _, rerr := s.Read(ref(t, 2)); !errors.Is(rerr, core.ErrNotHeld) {
		t.Error("a failed claim left a record")
	}
}

// The same holder on the ref is not enough: the epoch is the claim. A ref record
// of an older epoch is a claim this cache has moved past, not the one recorded.
func TestTheRefMustCarryThisEpoch(t *testing.T) {
	m := NewMemoryRef()
	s := attached(t, m)
	if _, err := s.Acquire(ref(t, 1), "w", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	data, rev, _ := m.Read(context.Background(), "issue/1")
	var onRef Record
	_ = jsonUnmarshal(data, &onRef)
	onRef.Epoch--
	older, _ := json.Marshal(onRef)
	if _, err := m.CompareAndSwap(context.Background(), "issue/1", rev, older); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ref(t, 2), "w", "d2", "run-1"); !errors.Is(err, core.ErrResponsibilityNotRecorded) {
		t.Errorf("the ref carries an older epoch of the same holder: %v", err)
	}
}

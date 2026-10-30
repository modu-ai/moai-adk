package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// The write path refuses a record whose derived position no chain of S46 edges
// reaches from the stored one, and leaves the stored record as it was.
func TestAWriteTheTableCannotMakeIsRefusedAndNothingIsStored(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(r *Record) error { r.State = core.StateImplementing; return nil }); err != nil {
		t.Fatalf("claimed to implementing is E23: %v", err)
	}
	// BUILDING/IMPLEMENT back to BUILDING/SPEC: no edge goes backwards
	before, _ := s.Read(i)
	_, err := s.Update(i, "sess-a", rec.Epoch, func(r *Record) error { r.State = core.StateClaimed; return nil })
	if !errors.Is(err, core.ErrTransitionInvalid) {
		t.Fatalf("a sub stage going backwards: %v", err)
	}
	after, _ := s.Read(i)
	if after.Seq != before.Seq || after.State != before.State {
		t.Errorf("a refused write stored something: %+v -> %+v", before, after)
	}
	// finishing in one write is the collapse of E03, E08, E12 and E13 and is allowed
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(r *Record) error {
		r.State, r.Merged, r.Closed = core.StateDone, true, true
		return nil
	}); err != nil {
		t.Fatalf("implementing to done in one write: %v", err)
	}
	// a finished Issue does not open again through a write
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(r *Record) error { r.State, r.Merged = core.StateImplementing, false; return nil }); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("a done Issue reopening: %v", err)
	}
}

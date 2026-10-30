package claim

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestRecordMirrorSetsOnlyTheStateAndMovesSeq(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, err := s.Acquire(i, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RecordMirror(i, "refused"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Read(i)
	if got.Mirror != "refused" || got.Holder != "sess-a" || got.Epoch != rec.Epoch || got.State != rec.State || got.Seq != rec.Seq+1 {
		t.Errorf("only the mirror state may change: %+v from %+v", got, rec)
	}
	// the owner's ticket still works: the mirror took nothing from it
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(*Record) error { return nil }); err != nil {
		t.Errorf("owner write after the mirror's: %v", err)
	}
}

func TestRecordMirrorLeavesAnUnchangedStateAlone(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	if _, err := s.Acquire(i, "sess-a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordMirror(i, "refused"); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Read(i)
	if err := s.RecordMirror(i, "refused"); err != nil {
		t.Fatal(err)
	}
	after, _ := s.Read(i)
	if after.Seq != before.Seq {
		t.Errorf("recording the same state again moved Seq: %d -> %d", before.Seq, after.Seq)
	}
}

func TestRecordMirrorOfAnUnknownIssueIsNotHeld(t *testing.T) {
	s := open(t)
	if err := s.RecordMirror(ref(t, 4242), "refused"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("%v", err)
	}
}

// A monitor's write under the owner's holder and epoch (a reminder on a wait) moves
// Seq but is not the owner acting, so it leaves OwnerSeq and any safe stop alone.
func TestUpdateAsMonitorDoesNotCountAsTheOwnerActing(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	stop, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.UpdateAsMonitor(i, "sess-a", rec.Epoch, func(r *Record) error { r.Wait = monitorWait(); return nil })
	if err != nil || got.Wait == nil || got.Wait.NextAction != "ask again" || got.Seq != rec.Seq+1 || got.OwnerSeq != rec.OwnerSeq {
		t.Fatalf("a monitor write moves Seq and not OwnerSeq: %+v %v", got, err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, stop); err != nil {
		t.Errorf("the stop must survive a monitor write: %v", err)
	}
	if _, err := s.UpdateAsMonitor(i, "sess-b", rec.Epoch, func(*Record) error { return nil }); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("another holder: %v", err)
	}
	if _, err := s.UpdateAsMonitor(i, "sess-a", rec.Epoch+1, func(*Record) error { return nil }); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("another epoch: %v", err)
	}
}

// A monitor writes under the owner's ticket, so what it may change is narrowed to
// the wait itself: whatever else fn touches is put back. This limits the damage a
// monitor can do with the ticket; it does not make the write the owner's, and the
// monitor is still trusted to be the monitor (REQ-AF-011's identity lock closes that).
func TestUpdateAsMonitorChangesOnlyTheWait(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	got, err := s.UpdateAsMonitor(i, "sess-a", rec.Epoch, func(r *Record) error {
		r.State, r.Return, r.Head, r.Resume, r.Closed, r.Merged, r.Pending = core.StateDone, core.StateBlocked, "deadbeef", "x", true, true, "deploy"
		r.Wait = monitorWait()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Wait == nil || got.Wait.NextAction != "ask again" {
		t.Errorf("the wait is what a monitor may write: %+v", got.Wait)
	}
	if got.State != rec.State || got.Return != rec.Return || got.Head != rec.Head || got.Resume != rec.Resume ||
		got.Closed != rec.Closed || got.Merged != rec.Merged || got.Pending != rec.Pending {
		t.Errorf("a monitor changed more than the wait: %+v -> %+v", rec, got)
	}
	stored, _ := s.Read(i)
	if stored.State != rec.State || stored.Closed {
		t.Errorf("the stored record carries the monitor's extra changes: %+v", stored)
	}
}

func monitorWait() *core.Wait {
	return &core.Wait{Owner: "sess-a", NextAction: "ask again", RecheckAt: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC), Diagnosis: "d", MaxRetries: 2, Escalate: "leader"}
}

// fn gets a copy it may change in place: a pointer field's target or a History
// element changed through it must not reach the stored record, because only the
// wait is kept.
func TestUpdateAsMonitorCannotChangeSharedPartsInPlace(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(r *Record) error {
		r.Receipt = &Receipt{HandoffID: "h1", PR: 7, Branch: "b", Worktree: "w"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Read(i)
	if _, err := s.UpdateAsMonitor(i, "sess-a", rec.Epoch, func(r *Record) error {
		r.Receipt.PR = 0
		r.Receipt.LeaderAccepted = true
		r.History[0].Note = "rewritten"
		r.History = append(r.History, Event{Kind: "forged"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	after, _ := s.Read(i)
	if after.Receipt == nil || *after.Receipt != *before.Receipt {
		t.Errorf("the receipt changed in place: %+v -> %+v", before.Receipt, after.Receipt)
	}
	if len(after.History) != len(before.History) || after.History[0] != before.History[0] {
		t.Errorf("the history changed in place: %+v -> %+v", before.History, after.History)
	}
}

// The reference-typed fields of Record are the ones clone() must copy. A new one
// fails here until clone handles it and this list names it.
func TestRecordCloneKnowsEveryReferenceField(t *testing.T) {
	known := map[string]bool{"Receipt": true, "Park": true, "Gate": true, "Refusal": true, "Wait": true, "Repair": true, "History": true}
	typ := reflect.TypeOf(Record{})
	for n := 0; n < typ.NumField(); n++ {
		f := typ.Field(n)
		switch f.Type.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
			if !known[f.Name] {
				t.Errorf("Record.%s is a reference field clone() does not copy: add it to clone and to this list", f.Name)
			}
			delete(known, f.Name)
		}
	}
	for n := range known {
		t.Errorf("clone() lists %s but Record has no such reference field", n)
	}
}

// A clone shares no memory with its original: a change through the copy of any
// pointer target or of a History element is invisible in the original.
func TestRecordCloneSharesNoMemory(t *testing.T) {
	r := Record{
		Receipt: &Receipt{PR: 1}, Park: &core.Park{Reason: "a"}, Gate: &core.Gate{Detail: "a"},
		Refusal: &core.Refusal{Detail: "a"}, Wait: &core.Wait{Diagnosis: "a"}, Repair: &Repair{Decision: "a"},
		History: []Event{{Note: "a"}},
	}
	c := r.clone()
	c.Receipt.PR, c.Park.Reason, c.Gate.Detail, c.Refusal.Detail, c.Wait.Diagnosis, c.Repair.Decision, c.History[0].Note = 2, "b", "b", "b", "b", "b", "b"
	c.History = append(c.History, Event{Kind: "x"})
	if r.Receipt.PR != 1 || r.Park.Reason != "a" || r.Gate.Detail != "a" || r.Refusal.Detail != "a" ||
		r.Wait.Diagnosis != "a" || r.Repair.Decision != "a" || r.History[0].Note != "a" || len(r.History) != 1 {
		t.Errorf("the clone shares memory with its original: %+v", r)
	}
	if (Record{}).clone().History != nil {
		t.Error("a nil History stays nil")
	}
}

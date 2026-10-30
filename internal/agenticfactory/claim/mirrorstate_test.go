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
	if after.Receipt == nil || !reflect.DeepEqual(after.Receipt, before.Receipt) {
		t.Errorf("the receipt changed in place: %+v -> %+v", before.Receipt, after.Receipt)
	}
	if !reflect.DeepEqual(after.History, before.History) {
		t.Errorf("the history changed in place: %+v -> %+v", before.History, after.History)
	}
}

// clone() copies the targets of Record's own reference fields with `v := *p`, which
// is a full copy only while those targets hold values. This walks the type: a
// reference-typed field of Record that clone() does not name fails, and so does a
// reference-typed field anywhere inside what it points to (a slice, map, pointer or
// interface in Receipt, Event, Repair, Park, Gate, Refusal or Wait), because a
// shallow copy of that target would share it. A new one fails until clone() copies
// it deeply and the lists below name it. time.Time is a value: its location pointer
// is immutable and shared on purpose.
func TestRecordCloneKnowsEveryReferenceField(t *testing.T) {
	topKnown := map[string]bool{"Receipt": true, "Park": true, "Gate": true, "Refusal": true, "Wait": true, "Repair": true, "History": true, "Evidence": true}
	var nestedKnown = map[string]bool{} // "Receipt.Field" paths clone() copies deeply
	isRef := func(k reflect.Kind) bool {
		switch k {
		case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
			return true
		}
		return false
	}
	timeType := reflect.TypeOf(time.Time{})
	flag := func(p string) {
		if !nestedKnown[p] {
			t.Errorf("%s is a reference field inside a type clone() copies with a shallow `*p`: copy it deeply and add %q to nestedKnown", p, p)
		}
	}
	// walk looks inside typ for reference-typed members. The container itself (a
	// pointer, slice, array or map) is judged by the caller; its element is judged
	// here, and so are the members of an element that is a struct.
	var walk func(typ reflect.Type, path string)
	walk = func(typ reflect.Type, path string) {
		switch typ.Kind() {
		case reflect.Pointer:
			walk(typ.Elem(), path)
		case reflect.Slice:
			if isRef(typ.Elem().Kind()) {
				flag(path + "[]")
			}
			walk(typ.Elem(), path)
		case reflect.Array:
			if isRef(typ.Elem().Kind()) {
				flag(path + "[]")
			}
			walk(typ.Elem(), path+"[]")
		case reflect.Map:
			if isRef(typ.Elem().Kind()) {
				flag(path + "[value]")
			}
			walk(typ.Elem(), path+"[value]")
		case reflect.Struct:
			if typ == timeType {
				return
			}
			for n := 0; n < typ.NumField(); n++ {
				f := typ.Field(n)
				p := path + "." + f.Name
				if isRef(f.Type.Kind()) {
					flag(p)
				}
				walk(f.Type, p)
			}
		}
	}
	typ := reflect.TypeOf(Record{})
	for n := 0; n < typ.NumField(); n++ {
		f := typ.Field(n)
		if !isRef(f.Type.Kind()) {
			continue
		}
		if !topKnown[f.Name] {
			t.Errorf("Record.%s is a reference field clone() does not copy: add it to clone and to this list", f.Name)
		}
		delete(topKnown, f.Name)
		walk(f.Type, f.Name)
	}
	for n := range topKnown {
		t.Errorf("clone() lists %s but Record has no such reference field", n)
	}
}

// A clone shares no memory with its original: a change through the copy of any
// pointer target or of a History element is invisible in the original.
func TestRecordCloneSharesNoMemory(t *testing.T) {
	r := Record{
		Receipt: &Receipt{PR: 1}, Park: &core.Park{Reason: "a"}, Gate: &core.Gate{Detail: "a"},
		Refusal: &core.Refusal{Detail: "a"}, Wait: &core.Wait{Diagnosis: "a"}, Repair: &Repair{Decision: "a"},
		History: []Event{{Note: "a"}}, Evidence: []string{"a"},
	}
	c := r.clone()
	c.Receipt.PR, c.Park.Reason, c.Gate.Detail, c.Refusal.Detail, c.Wait.Diagnosis, c.Repair.Decision, c.History[0].Note = 2, "b", "b", "b", "b", "b", "b"
	c.History = append(c.History, Event{Kind: "x"})
	c.Evidence[0] = "b"
	if r.Receipt.PR != 1 || r.Park.Reason != "a" || r.Gate.Detail != "a" || r.Refusal.Detail != "a" ||
		r.Wait.Diagnosis != "a" || r.Repair.Decision != "a" || r.History[0].Note != "a" || len(r.History) != 1 || r.Evidence[0] != "a" {
		t.Errorf("the clone shares memory with its original: %+v", r)
	}
	if (Record{}).clone().History != nil {
		t.Error("a nil History stays nil")
	}
}

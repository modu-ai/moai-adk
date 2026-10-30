package claim

import (
	"errors"
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

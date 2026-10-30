package claim

import (
	"errors"
	"testing"

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
	got, err := s.UpdateAsMonitor(i, "sess-a", rec.Epoch, func(r *Record) error { r.Resume = "reminded"; return nil })
	if err != nil || got.Resume != "reminded" || got.Seq != rec.Seq+1 || got.OwnerSeq != rec.OwnerSeq {
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

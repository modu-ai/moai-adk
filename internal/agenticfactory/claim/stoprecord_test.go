package claim

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestASafeStopIsRecordedByTheCurrentHolderOnly(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, err := s.Acquire(i, "sess-a", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RecordSafeStop(i, "sess-b", rec.Epoch, "x"); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("another holder: %v", err)
	}
	if _, err := s.RecordSafeStop(i, "sess-a", rec.Epoch+1, "x"); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("a wrong epoch: %v", err)
	}
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stopped after the last commit")
	if err != nil || name == "" {
		t.Fatalf("record: %q %v", name, err)
	}
	again, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stopped after the last commit")
	if err != nil || again != name {
		t.Errorf("recording the same stop twice names the same record: %q %v", again, err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Errorf("the record it wrote must check out: %v", err)
	}
}

func TestCheckSafeStopRefusesAnythingButThisOwnersRecord(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	other := ref(t, 2000)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	orec, _ := s.Acquire(other, "sess-a", "d2", "run-1")
	mine, _ := s.RecordSafeStop(i, "sess-a", rec.Epoch, "x")
	theirs, _ := s.RecordSafeStop(other, "sess-a", orec.Epoch, "y")
	for name, c := range map[string]struct {
		holder string
		epoch  int64
		ref    string
	}{
		"not written":    {"sess-a", rec.Epoch, "issue-1948-e9"},
		"invented":       {"sess-a", rec.Epoch, "stop-record-7"},
		"not a key":      {"sess-a", rec.Epoch, "../x"},
		"empty":          {"sess-a", rec.Epoch, " "},
		"another Issue":  {"sess-a", rec.Epoch, theirs},
		"another holder": {"sess-b", rec.Epoch, mine},
		"another epoch":  {"sess-a", rec.Epoch + 1, mine},
	} {
		if err := s.CheckSafeStop(i, c.holder, c.epoch, c.ref); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

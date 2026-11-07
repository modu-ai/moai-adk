package claim

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
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

// A safe stop says "I stopped"; an owner that acts after it has not. The stop
// goes stale once the claim record is written again or the owner beats again.
func TestASafeStopGoesStaleWhenTheOwnerActsAfterIt(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")

	// an owner that beat before it stopped is stopped
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp", time.Hour); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "first stop")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Fatalf("a beat before the stop does not undo it: %v", err)
	}

	// a beat after it does
	now = now.Add(time.Second)
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("the owner beat after its stop: %v", err)
	}
	// stopping again refreshes the same record
	now = now.Add(time.Second)
	again, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "second stop")
	if err != nil || again != name {
		t.Fatalf("a fresh stop names the same record: %q %v", again, err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Errorf("the refreshed stop: %v", err)
	}

	// a write of the claim record after it does too
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(*Record) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("the owner wrote its record after its stop: %v", err)
	}
}

// An earlier holder's beat says nothing about the current holder's stop.
func TestAnEarlierHoldersBeatDoesNotStaleTheNextHoldersStop(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp", time.Hour); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	// The takeover re-reads the owner's process under the lock, so the first holder's
	// process is gone when it goes ahead; its presence file stays behind.
	s.probe = probeOf(homestate.ProcessIdentityDead, "fp")
	next, err := s.TakeOverUnverified(i, "sess-a", rec.Epoch, "sess-b", "d2", "run-2", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
	now = now.Add(time.Second)
	name, err := s.RecordSafeStop(i, "sess-b", next.Epoch, "second owner stops")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-b", next.Epoch, name); err != nil {
		t.Errorf("the first holder's presence file must not count: %v", err)
	}
}

// With a presence file that cannot be read, nothing can be said about whether the
// owner beat after its stop: the stop is refused, and the refusal says how to go on.
func TestADamagedPresenceRefusesTheStopAndSaysHowToRecover(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	if err := os.WriteFile(s.presencePath(i), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stop")
	if err != nil {
		t.Fatal(err)
	}
	err = s.CheckSafeStop(i, "sess-a", rec.Epoch, name)
	if !errors.Is(err, core.ErrEvidenceIncomplete) || !strings.Contains(err.Error(), "remove") || !strings.Contains(err.Error(), "issue-1948.presence") {
		t.Errorf("a damaged presence: %v", err)
	}
}

// When the beat and the stop carry the same instant their order is unknown, and an
// unknown order is not a stop.
func TestABeatAtTheSameInstantAsTheStopMakesItStale(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stop")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Heartbeat(i, "sess-a", rec.Epoch, 4242, "fp", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("a beat at the instant of the stop: %v", err)
	}
}

// Even an earlier holder's presence file stamped LATER than the stop (a clock that
// jumped, a file copied in) says nothing about the current holder's stop.
func TestAnotherHoldersLaterPresenceFileDoesNotStaleTheStop(t *testing.T) {
	s := open(t)
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stop")
	if err != nil {
		t.Fatal(err)
	}
	other := `{"holder":"sess-z","epoch":9,"pid":4242,"start":"fp","at":"2031-01-01T00:00:00Z","lease_until":"2031-01-02T00:00:00Z"}`
	if err := os.WriteFile(s.presencePath(i), []byte(other), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Errorf("another holder's presence file counted: %v", err)
	}
}

// Only what the OWNER writes undoes its stop. A record written by the mirror or a
// monitor (RecordMirror, a repair signal) is not the owner acting: a dead owner cannot
// record its stop again, so counting those writes would make every stop go stale.
func TestAStopSurvivesWritesThatAreNotTheOwners(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RecordMirror(i, "refused"); err != nil {
		t.Fatal(err)
	}
	if err := s.RecordMirror(i, "rejected(404)"); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Fatalf("the mirror's writes must not undo the owner's stop: %v", err)
	}
	// the owner writing its record does
	if _, err := s.Update(i, "sess-a", rec.Epoch, func(*Record) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("the owner wrote its record after its stop: %v", err)
	}
}

// And the same through the takeover: a stop that only the mirror wrote after still
// hands the Issue over.
func TestTakeOverOnStopSurvivesTheMirrorsWrites(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RecordMirror(i, "refused"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TakeOverOnStop(i, "sess-a", rec.Epoch, "sess-b", "d2", "run-2", "note", nil, name); err != nil {
		t.Errorf("%v", err)
	}
}

// A stop document that carries no owner_seq (the old "seq" key, or none) cannot be
// placed in the owner's history, so it proves nothing: fail closed. An owner that
// has never written since acquiring (OwnerSeq 0 on both sides) must not make a
// missing key read as a match.
func TestAStopWithoutOwnerSeqIsStale(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	rec, _ := s.Acquire(i, "sess-a", "d1", "run-1")
	name := stopName(i, rec.Epoch)
	old := fmt.Sprintf(`{"issue":%d,"holder":"sess-a","epoch":%d,"note":"old format","at":%q,"seq":%d}`,
		i.Number, rec.Epoch, s.clock().Format(time.RFC3339Nano), rec.OwnerSeq)
	if _, err := s.PutDoc("stop", name, 0, []byte(old)); err != nil {
		t.Fatal(err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("a stop with no owner_seq must be refused: %v", err)
	}
	// recording again replaces it with a record that carries owner_seq
	again, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "re-recorded")
	if err != nil || again != name {
		t.Fatalf("re-record: %q %v", again, err)
	}
	if err := s.CheckSafeStop(i, "sess-a", rec.Epoch, name); err != nil {
		t.Errorf("the re-recorded stop is fresh: %v", err)
	}
}

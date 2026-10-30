package claim

import (
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The safe stop is checked where the record changes hands, under the Issue lock, so
// nothing the owner writes in between can slip past a check made earlier.
func TestTakeOverOnStopChecksTheStopItself(t *testing.T) {
	setup := func(t *testing.T) (*Store, core.IssueRef, int64, string, *time.Time) {
		s := open(t)
		now := time.Date(2026, 10, 30, 11, 0, 0, 0, time.UTC)
		s.now = func() time.Time { return now }
		s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
		i := ref(t, 1948)
		rec, err := s.Acquire(i, "sess-a", "d1", "run-1")
		if err != nil {
			t.Fatal(err)
		}
		now = now.Add(time.Second)
		name, err := s.RecordSafeStop(i, "sess-a", rec.Epoch, "stop")
		if err != nil {
			t.Fatal(err)
		}
		return s, i, rec.Epoch, name, &now
	}

	t.Run("a fresh stop hands the Issue over", func(t *testing.T) {
		s, i, epoch, name, _ := setup(t)
		got, err := s.TakeOverOnStop(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil, name)
		if err != nil || got.Holder != "sess-b" || got.Epoch != epoch+1 {
			t.Fatalf("%+v %v", got, err)
		}
	})
	t.Run("a claim record written after the stop refuses it", func(t *testing.T) {
		s, i, epoch, name, _ := setup(t)
		if _, err := s.Update(i, "sess-a", epoch, func(*Record) error { return nil }); err != nil {
			t.Fatal(err)
		}
		if _, err := s.TakeOverOnStop(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Fatalf("a write after the stop: %v", err)
		}
		if cur, _ := s.Read(i); cur.Holder != "sess-a" {
			t.Errorf("the record moved: %+v", cur)
		}
	})
	t.Run("a beat after the stop refuses it", func(t *testing.T) {
		s, i, epoch, name, now := setup(t)
		*now = now.Add(time.Second)
		if err := s.Heartbeat(i, "sess-a", epoch, 4242, "fp", time.Hour); err != nil {
			t.Fatal(err)
		}
		if _, err := s.TakeOverOnStop(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil, name); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Fatalf("a beat after the stop: %v", err)
		}
	})
	t.Run("an invented reference refuses it", func(t *testing.T) {
		s, i, epoch, _, _ := setup(t)
		if _, err := s.TakeOverOnStop(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil, "issue-1948-e9"); !errors.Is(err, core.ErrEvidenceIncomplete) {
			t.Fatalf("an invented reference: %v", err)
		}
	})
	t.Run("a takeover that relies on a stop needs one", func(t *testing.T) {
		s, i, epoch, _, _ := setup(t)
		for _, ref := range []string{"", "  "} {
			if _, err := s.TakeOverOnStop(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil, ref); !errors.Is(err, core.ErrEvidenceIncomplete) {
				t.Errorf("an empty stop reference %q: %v", ref, err)
			}
		}
		if cur, _ := s.Read(i); cur.Holder != "sess-a" {
			t.Errorf("the record moved: %+v", cur)
		}
	})
	t.Run("the unverified takeover is a separate call that says so", func(t *testing.T) {
		s, i, epoch, _, _ := setup(t)
		if _, err := s.TakeOverUnverified(i, "sess-a", epoch, "sess-b", "d2", "run-2", "note", nil); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

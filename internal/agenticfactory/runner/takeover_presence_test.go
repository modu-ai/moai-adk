package runner

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func beat(t *testing.T, r *Runner, tk Ticket, state homestate.ProcessIdentityState) {
	t.Helper()
	r.Store.SetProcessProbe(func(int) (string, homestate.ProcessIdentityState) { return "fp-owner", homestate.ProcessIdentityLive })
	if err := r.Store.Heartbeat(tk.Issue, tk.Holder, tk.Epoch, 4242, "fp-owner", time.Hour); err != nil {
		t.Fatal(err)
	}
	r.Store.SetProcessProbe(func(int) (string, homestate.ProcessIdentityState) { return "fp-owner", state })
}

func notGone() Takeover {
	ev := fullTakeover()
	ev.OwnerGone = false
	return ev
}

// Where the owner's process is on record, whether it is gone is read from the
// process, not taken from the caller.
func TestTakeoverReadsTheRecordedProcessInsteadOfTheCallersWord(t *testing.T) {
	t.Run("a live owner is not taken over however the caller puts it", func(t *testing.T) {
		r, ref, tk := handedOffOwner(t)
		beat(t, r, tk, homestate.ProcessIdentityLive)
		if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover()); !errors.Is(err, core.ErrHeld) {
			t.Fatalf("a live owner: %v", err)
		}
		if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
			t.Errorf("the record moved: %+v", cur)
		}
	})
	t.Run("an owner that cannot be told apart is unconfirmed", func(t *testing.T) {
		r, ref, tk := handedOffOwner(t)
		beat(t, r, tk, homestate.ProcessIdentityIndeterminate)
		if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover()); !errors.Is(err, core.ErrOwnerUnconfirmed) {
			t.Fatalf("an indeterminate owner: %v", err)
		}
	})
	t.Run("a dead owner needs no claim from the caller", func(t *testing.T) {
		r, ref, tk := handedOffOwner(t)
		beat(t, r, tk, homestate.ProcessIdentityDead)
		rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", notGone())
		if err != nil || rec.Holder != "sess-new" {
			t.Fatalf("a dead owner: %+v %v", rec, err)
		}
		if note := rec.History[len(rec.History)-1].Note; !strings.Contains(note, "owner process 4242 is gone") {
			t.Errorf("history must say how the departure was established: %q", note)
		}
	})
	t.Run("a stop record the owner left stands in even for a live process", func(t *testing.T) {
		r, ref, tk := handedOffOwner(t)
		beat(t, r, tk, homestate.ProcessIdentityLive)
		stop, err := r.RecordSafeStop(tk, "owner stopped itself")
		if err != nil {
			t.Fatal(err)
		}
		ev := fullTakeover() // the caller also claims the owner is gone; the process says it is not
		ev.SafeStop = stop
		rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev)
		if err != nil {
			t.Fatalf("a safe stop: %v", err)
		}
		note := rec.History[len(rec.History)-1].Note
		if !strings.Contains(note, "safe stop "+stop) || strings.Contains(note, "is gone") {
			t.Errorf("history must not call a running owner gone: %q", note)
		}
	})
	t.Run("with no process on record the caller's word still counts, and says so", func(t *testing.T) {
		r, ref, tk := handedOffOwner(t)
		rec, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover())
		if err != nil {
			t.Fatal(err)
		}
		if note := rec.History[len(rec.History)-1].Note; !strings.Contains(note, "caller-asserted") {
			t.Errorf("an unverified departure must be marked: %q", note)
		}
	})
}

// A running owner is not taken over on the strength of an unverified string.
func TestARunningOwnerIsNotTakenOverOnAnInventedStopReference(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	beat(t, r, tk, homestate.ProcessIdentityLive)
	ev := notGone()
	ev.SafeStop = "stop-record-9"
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("an invented reference: %v", err)
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
		t.Errorf("the record moved: %+v", cur)
	}
}

// A presence file that cannot be read blocks the takeover and the refusal names
// the file and the way out.
func TestADamagedPresenceBlocksATakeoverAndNamesTheRecovery(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	beat(t, r, tk, homestate.ProcessIdentityLive)
	path := r.Store.PresenceFile(ref)
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", fullTakeover())
	if !errors.Is(err, core.ErrOwnerUnconfirmed) || !strings.Contains(err.Error(), "remove") || !strings.Contains(err.Error(), "issue-1948.presence") {
		t.Errorf("a damaged presence: %v", err)
	}
}

// The reverse of a stop recorded after the last beat: the owner beats AGAIN after
// recording its stop, so it has resumed and the stop no longer stands for it.
func TestARunningOwnerThatBeatsAfterItsStopIsNotTakenOver(t *testing.T) {
	r, ref, tk := handedOffOwner(t)
	beat(t, r, tk, homestate.ProcessIdentityLive)
	stop, err := r.RecordSafeStop(tk, "owner stopped")
	if err != nil {
		t.Fatal(err)
	}
	beat(t, r, tk, homestate.ProcessIdentityLive) // resumed
	ev := notGone()
	ev.SafeStop = stop
	if _, err := r.TakeOver(ref, tk, "sess-new", "disp-t", "run-2", ev); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("a stop the owner has beaten after: %v", err)
	}
	if cur, _ := r.Store.Read(ref); cur.Holder != "sess-a" {
		t.Errorf("the record moved: %+v", cur)
	}
}

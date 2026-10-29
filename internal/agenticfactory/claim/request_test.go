package claim

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var pol = Policy{MaxInFlight: 2, MaxParked: 1}

func TestPutRequestIsIdempotentAndConflictsOnOtherPayload(t *testing.T) {
	s := open(t)
	a, created, err := s.PutRequest("req-1", "hash-a", 5, pol)
	if err != nil || !created || a.Stage != StageACK || a.Seq != 1 || len(a.Evidence) != 1 || a.Evidence[0].Stage != StageACK {
		t.Fatalf("first put: %+v %v %v", a, created, err)
	}
	again, created, err := s.PutRequest("req-1", "hash-a", 5, pol)
	if err != nil || created || again.Seq != a.Seq {
		t.Errorf("same payload must return the record: %+v %v %v", again, created, err)
	}
	if _, _, err := s.PutRequest("req-1", "hash-b", 5, pol); !errors.Is(err, core.ErrConflict) {
		t.Errorf("other payload: %v", err)
	}
	if got, _ := s.GetRequest("req-1"); got.PayloadHash != "hash-a" {
		t.Errorf("conflict changed the record: %+v", got)
	}
	if _, err := s.GetRequest("nope"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing request: %v", err)
	}
	for _, bad := range []string{"", "Bad Key", "../x", "a/b"} {
		if _, _, err := s.PutRequest(bad, "h", 0, pol); !errors.Is(err, core.ErrIDInvalid) {
			t.Errorf("key %q: %v", bad, err)
		}
	}
}

func TestPutRequestAdmissionParksThenRejects(t *testing.T) {
	s := open(t)
	for _, k := range []string{"a", "b"} {
		r, _, err := s.PutRequest(k, "h", 1, pol)
		if err != nil || r.Parked != "" {
			t.Fatalf("%s: %+v %v", k, r, err)
		}
	}
	c, _, err := s.PutRequest("c", "h", 1, pol)
	if err != nil || c.Parked != ParkBackpressure {
		t.Fatalf("c must be parked: %+v %v", c, err)
	}
	if _, _, err := s.PutRequest("d", "h", 9, pol); !errors.Is(err, core.ErrIntakeFull) {
		t.Fatalf("d must be rejected: %v", err)
	}
	if _, err := s.GetRequest("d"); !errors.Is(err, core.ErrNotHeld) {
		t.Error("a rejected request left a record")
	}
	// A repeat of an already admitted request is never refused for capacity.
	if _, created, err := s.PutRequest("a", "h", 1, pol); err != nil || created {
		t.Errorf("repeat under full capacity: %v %v", created, err)
	}
}

func TestUpdateRequestKeepsIdentityAndStageOrder(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 3, pol)
	out, err := s.UpdateRequest("r", func(r *Request) error {
		r.Key, r.PayloadHash, r.Seq, r.Priority = "x", "y", 99, 99
		r.Stage = StageIssuePublished
		r.Issue = 7
		return nil
	})
	if err != nil || out.Key != "r" || out.PayloadHash != "h" || out.Seq != 1 || out.Priority != 3 || out.Stage != StageIssuePublished {
		t.Fatalf("immutables changed: %+v %v", out, err)
	}
	for _, bad := range []Stage{StageStarted, StageACK, "BOGUS"} {
		if _, err := s.UpdateRequest("r", func(r *Request) error { r.Stage = bad; return nil }); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("stage %q: %v", bad, err)
		}
	}
	boom := errors.New("boom")
	if _, err := s.UpdateRequest("r", func(*Request) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("fn error: %v", err)
	}
	if _, err := s.UpdateRequest("none", func(*Request) error { return nil }); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing: %v", err)
	}
}

// startViaClaim makes the request's Issue run: a claim whose first attempt began.
func startViaClaim(t *testing.T, s *Store, key string, issue int) {
	t.Helper()
	for _, st := range []Stage{StageIssuePublished, StageWTReady} {
		if _, err := s.UpdateRequest(key, func(r *Request) error {
			r.Stage, r.Issue = st, issue
			if st == StageWTReady {
				r.Ready = &Ready{Slot: "wt-1", Path: "/wt", Head: "h", WriterStopped: true}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	ref, _ := core.NewIssueRef(issue)
	rec, err := s.Acquire(ref, "holder-"+key, "d-"+key, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ref, rec.Holder, rec.Epoch, func(r *Record) error { r.State, r.AttemptSeq = core.StateImplementing, 1; return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestStartedIsDerivedNotStored(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	startViaClaim(t, s, "r", 11)
	raw, _ := s.GetRequest("r")
	if raw.Stage != StageWTReady {
		t.Fatalf("stored stage must stay WT_READY, got %v", raw.Stage)
	}
	v, err := s.ViewRequest("r")
	if err != nil || v.Stage != StageStarted {
		t.Fatalf("view: %+v %v", v, err)
	}
	last := v.Evidence[len(v.Evidence)-1]
	if last.Stage != StageStarted || last.Ref == "" {
		t.Errorf("derived evidence: %+v", last)
	}
	if again, _ := s.GetRequest("r"); len(again.Evidence) != len(raw.Evidence) {
		t.Error("deriving STARTED must not write evidence")
	}
}

func TestReadyGateOnlyOpensAtWTReady(t *testing.T) {
	s := open(t)
	if err := s.ReadyGate(42); err != nil {
		t.Errorf("an issue with no intake request is not gated: %v", err)
	}
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Stage, r.Issue = StageIssuePublished, 42; return nil })
	if err := s.ReadyGate(42); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("before WT_READY: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error {
		r.Stage = StageWTReady
		r.Ready = &Ready{Slot: "wt-1", Path: "/wt", Head: "h", WriterStopped: false}
		return nil
	})
	if err := s.ReadyGate(42); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("preparer still writing: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Ready.WriterStopped = true; return nil })
	if err := s.ReadyGate(42); err != nil {
		t.Errorf("ready: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Parked = ParkNoSlot; return nil })
	if err := s.ReadyGate(42); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("parked: %v", err)
	}
}

func TestPromoteParkedByPriorityThenArrival(t *testing.T) {
	s := open(t)
	p := Policy{MaxInFlight: 1, MaxParked: 5}
	_, _, _ = s.PutRequest("run", "h", 1, p)
	for _, k := range []struct {
		key string
		pri int
	}{{"low", 1}, {"high", 9}, {"high2", 9}, {"mid", 5}} {
		if r, _, _ := s.PutRequest(k.key, "h", k.pri, p); r.Parked != ParkBackpressure {
			t.Fatalf("%s not parked", k.key)
		}
	}
	// Capacity is still taken: nothing is promoted.
	if got, _ := s.PromoteParked(p, 0); len(got) != 0 {
		t.Fatalf("promoted under full capacity: %+v", got)
	}
	startViaClaim(t, s, "run", 21)
	got, err := s.PromoteParked(p, 0)
	if err != nil || len(got) != 1 || got[0].Key != "high" {
		t.Fatalf("first promotion: %+v %v", got, err)
	}
	// high is now in flight; the next free slot goes to high2 (same priority, later arrival).
	startViaClaim(t, s, "high", 22)
	if got, _ := s.PromoteParked(p, 0); len(got) != 1 || got[0].Key != "high2" {
		t.Errorf("second promotion: %+v", got)
	}
}

func TestPromoteNoSlotParkedNeedsAFreeSlot(t *testing.T) {
	s := open(t)
	p := Policy{MaxInFlight: 5, MaxParked: 5}
	_, _, _ = s.PutRequest("w", "h", 1, p)
	_, _ = s.UpdateRequest("w", func(r *Request) error { r.Parked = ParkNoSlot; return nil })
	if got, _ := s.PromoteParked(p, 0); len(got) != 0 {
		t.Errorf("promoted without a free slot: %+v", got)
	}
	if got, _ := s.PromoteParked(p, 1); len(got) != 1 || got[0].Parked != "" {
		t.Errorf("free slot must promote: %+v", got)
	}
}

func TestSlotsAreExclusive(t *testing.T) {
	s := open(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := s.AcquireSlot("wt-1", "owner-"+string(rune('a'+i))); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if wins != 1 {
		t.Errorf("slot won %d times", wins)
	}
	owner, _ := s.SlotOwner("wt-1")
	if owner == "" {
		t.Fatal("slot has no owner")
	}
	if err := s.AcquireSlot("wt-1", owner); err != nil {
		t.Errorf("owner re-acquire must be idempotent: %v", err)
	}
	if err := s.AcquireSlot("wt-1", "someone-else"); !errors.Is(err, core.ErrHeld) {
		t.Errorf("held slot: %v", err)
	}
	if err := s.ReleaseSlot("wt-1", "someone-else"); !errors.Is(err, core.ErrForeign) {
		t.Errorf("foreign release: %v", err)
	}
	if err := s.ReleaseSlot("wt-1", owner); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSlot("wt-1", owner); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("double release: %v", err)
	}
	if o, err := s.SlotOwner("wt-1"); err != nil || o != "" {
		t.Errorf("free slot owner = %q %v", o, err)
	}
	if err := s.AcquireSlot("Bad Slot", "o"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("slot name: %v", err)
	}
}

func TestViewRequestsAndCorruptRequestFile(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("a", "h", 1, pol)
	_, _, _ = s.PutRequest("b", "h", 1, pol)
	startViaClaim(t, s, "a", 31)
	views, err := s.ViewRequests()
	if err != nil || len(views) != 2 || views[0].Key != "a" || views[0].Stage != StageStarted || views[1].Stage != StageACK {
		t.Fatalf("views: %+v %v", views, err)
	}
	if all, _ := s.ListRequests(); all[0].Stage != StageWTReady {
		t.Errorf("the stored stage must stay raw: %v", all[0].Stage)
	}
	if err := os.WriteFile(s.requestPath("b"), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ListRequests(); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("list with corrupt request: %v", err)
	}
	if _, err := s.GetRequest("b"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("get corrupt request: %v", err)
	}
	if _, _, err := s.PutRequest("c", "h", 1, pol); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("admission must not guess over an unreadable store: %v", err)
	}
}

func TestSlotFileCorruptionAndEmptyOwner(t *testing.T) {
	s := open(t)
	if err := s.AcquireSlot("wt-1", ""); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("empty owner: %v", err)
	}
	if err := s.AcquireSlot("wt-1", "o"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.slotPath("wt-1"), []byte("{x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SlotOwner("wt-1"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("corrupt slot: %v", err)
	}
	if err := s.AcquireSlot("wt-1", "o2"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("acquire over a corrupt slot must not guess: %v", err)
	}
	if err := s.ReleaseSlot("wt-1", "o"); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("release of a corrupt slot: %v", err)
	}
}

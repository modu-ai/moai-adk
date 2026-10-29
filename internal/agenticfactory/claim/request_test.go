package claim

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var (
	pol   = Policy{MaxInFlight: 2, MaxParked: 1}
	slots = []string{"wt-1", "wt-2", "wt-3"}
	sha40 = "0123456789abcdef0123456789abcdef01234567"
)

// toStage walks a request through the stages up to and including want, writing
// each step as the intake code would. The agent is registered on demand.
func toStage(t *testing.T, s *Store, key string, issue int, want Stage) core.Agent {
	t.Helper()
	a, err := s.RegisterAgent(sid(issue))
	if err != nil {
		t.Fatal(err)
	}
	cur, err := s.GetRequest(key)
	if err != nil {
		t.Fatal(err)
	}
	// Only the steps after the request's current stage are applied.
	need := func(st Stage) bool { return stageRank[want] >= stageRank[st] && stageRank[cur.Stage] < stageRank[st] }
	step := func(st Stage, fn func(*Request)) {
		if _, err := s.UpdateRequest(key, func(r *Request) error { r.Stage = st; fn(r); return nil }); err != nil {
			t.Fatalf("step %s: %v", st, err)
		}
	}
	if need(StageIssueConfirmed) {
		step(StageIssueConfirmed, func(r *Request) { r.Issue = issue })
	}
	if need(StageBranchRegistered) {
		step(StageBranchRegistered, func(r *Request) { r.Branch = &Branch{Ref: "WT-x", BaseOID: sha40} })
	}
	if need(StageAgentReserved) {
		// Reservation goes through the store: it picks the slot and checks the agent.
		if _, err := s.ReserveRequest(key, a, slots); err != nil {
			t.Fatalf("reserve: %v", err)
		}
	}
	if need(StageWTPrepared) {
		step(StageWTPrepared, func(r *Request) {
			r.Ready = &Ready{Host: "host-1", Slot: r.Slot, Path: "/wt/" + key, Head: sha40, WriterStopped: true}
		})
	}
	if need(StageReadyAcked) {
		if _, err := s.AckRequestReady(key, a); err != nil {
			t.Fatalf("ack: %v", err)
		}
	}
	return a
}

// startViaClaim makes the request's Issue run: the agent claims it and begins an attempt.
func startViaClaim(t *testing.T, s *Store, key string, issue int) {
	t.Helper()
	a := toStage(t, s, key, issue, StageReadyAcked)
	ref, _ := core.NewIssueRef(issue)
	rec, err := s.AcquireFor(ref, a, "d-"+key, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ref, rec.Holder, rec.Epoch, func(r *Record) error { r.State, r.AttemptSeq = core.StateImplementing, 1; return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestPutRequestIsIdempotentAndConflictsOnOtherPayload(t *testing.T) {
	s := open(t)
	a, created, err := s.PutRequest("req-1", "hash-a", 5, pol)
	if err != nil || !created || a.Stage != StageAccepted || a.Seq != 1 || len(a.Evidence) != 1 || a.Evidence[0].Stage != StageAccepted {
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
	if _, created, err := s.PutRequest("a", "h", 1, pol); err != nil || created {
		t.Errorf("repeat under full capacity: %v %v", created, err)
	}
}

func TestUpdateRequestKeepsIdentityAndStageOrder(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 3, pol)
	out, err := s.UpdateRequest("r", func(r *Request) error {
		r.Key, r.PayloadHash, r.Seq, r.Priority = "x", "y", 99, 99
		r.Stage = StageIssueConfirmed
		r.Issue = 7
		return nil
	})
	if err != nil || out.Key != "r" || out.PayloadHash != "h" || out.Seq != 1 || out.Priority != 3 || out.Stage != StageIssueConfirmed {
		t.Fatalf("immutables changed: %+v %v", out, err)
	}
	for _, bad := range []Stage{StageReadyAcked, StageAccepted, StageAgentReserved, "STARTED", "ACK", "BOGUS"} {
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

func TestStartedIsDerivedNotStored(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	startViaClaim(t, s, "r", 11)
	raw, _ := s.GetRequest("r")
	if raw.Stage != StageReadyAcked || raw.Started || raw.StartedRef != "" {
		t.Fatalf("a started agent changes nothing in the stored request: %+v", raw)
	}
	v, err := s.ViewRequest("r")
	if err != nil || v.Stage != StageReadyAcked || !v.Started || v.StartedRef == "" {
		t.Fatalf("view: %+v %v", v, err)
	}
	if again, _ := s.GetRequest("r"); len(again.Evidence) != len(raw.Evidence) {
		t.Error("deriving Started must not write evidence")
	}
	// Updating through the store never persists the derived fields.
	if _, err := s.UpdateRequest("r", func(r *Request) error { r.Started, r.StartedRef = true, "x"; return nil }); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetRequest("r"); got.Started || got.StartedRef != "" {
		t.Errorf("derived fields were stored: %+v", got)
	}
}

func TestReserveNeedsABranchAnAgentAndASlot(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	a, _ := s.RegisterAgent(sid(1))
	if _, err := s.ReserveRequest("r", a, slots); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("reserve before the branch is registered: %v", err)
	}
	toStage(t, s, "r", 1, StageBranchRegistered)
	ghost := core.Agent{SessionID: sid(99), Number: 9}
	if _, err := s.ReserveRequest("r", ghost, slots); !errors.Is(err, core.ErrIdentityMismatch) {
		t.Errorf("unregistered agent: %v", err)
	}
	if _, err := s.ReserveRequest("none", a, slots); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing request: %v", err)
	}
	got, err := s.ReserveRequest("r", a, slots)
	if err != nil || got.Stage != StageAgentReserved || got.Agent != a.SessionID || got.AgentName != "agent-1" || got.Slot != "wt-1" {
		t.Fatalf("reserve: %+v %v", got, err)
	}
	if owner, _ := s.SlotOwner("wt-1"); owner != "r" {
		t.Errorf("slot owner = %q", owner)
	}
	if last := got.Evidence[len(got.Evidence)-1]; last.Stage != StageAgentReserved || last.Ref != "agent-1@wt-1" {
		t.Errorf("evidence = %+v", last)
	}
}

func TestOneAgentOneRequestAtATime(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r1", "h", 1, Policy{MaxInFlight: 5, MaxParked: 5})
	_, _, _ = s.PutRequest("r2", "h", 1, Policy{MaxInFlight: 5, MaxParked: 5})
	a := toStage(t, s, "r1", 1, StageAgentReserved)
	for _, st := range []Stage{StageIssueConfirmed, StageBranchRegistered} {
		_, _ = s.UpdateRequest("r2", func(r *Request) error {
			r.Stage, r.Issue, r.Branch = st, 2, &Branch{Ref: "WT-y", BaseOID: sha40}
			return nil
		})
	}
	if _, err := s.ReserveRequest("r2", a, slots); !errors.Is(err, core.ErrHeld) {
		t.Fatalf("the same agent for a second request: %v", err)
	}
	if got, _ := s.GetRequest("r2"); got.Stage != StageBranchRegistered || got.Agent != "" {
		t.Errorf("a refused reservation changed the request: %+v", got)
	}
	// Once the first request's agent handed its PR off, the agent is free again.
	_, _ = s.AckRequestReady("r1", a) // refused: not prepared yet
	toStage(t, s, "r1", 1, StageReadyAcked)
	ref1, _ := core.NewIssueRef(1)
	rec, err := s.AcquireFor(ref1, a, "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.Update(ref1, rec.Holder, rec.Epoch, func(r *Record) error { r.AttemptSeq, r.State = 1, core.StateImplementing; return nil })
	if _, err := s.ReserveRequest("r2", a, slots); !errors.Is(err, core.ErrHeld) {
		t.Errorf("an agent that is still building holds its request: %v", err)
	}
	if err := s.Release(ref1, a.SessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReserveRequest("r2", a, slots); err != nil {
		t.Errorf("after the claim was released the agent is free: %v", err)
	}
}

func TestReserveWithNoFreeSlotParksTheRequest(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	a := toStage(t, s, "r", 1, StageBranchRegistered)
	for _, n := range slots {
		if err := s.AcquireSlot(n, "other-"+n); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.ReserveRequest("r", a, slots); !errors.Is(err, core.ErrNoSlot) {
		t.Fatalf("no slot: %v", err)
	}
	got, _ := s.GetRequest("r")
	if got.Parked != ParkNoSlot || got.Stage != StageBranchRegistered || got.Agent != "" {
		t.Errorf("no half reservation: %+v", got)
	}
}

func TestAckIsOnlyForTheReservedAgentAfterPreparation(t *testing.T) {
	s := open(t)
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	a := toStage(t, s, "r", 1, StageAgentReserved)
	if _, err := s.AckRequestReady("r", a); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("ack before the worktree is prepared: %v", err)
	}
	toStage(t, s, "r", 1, StageWTPrepared)
	other, _ := s.RegisterAgent(sid(2))
	if _, err := s.AckRequestReady("r", other); !errors.Is(err, core.ErrForeign) {
		t.Errorf("ack by another agent: %v", err)
	}
	if _, err := s.UpdateRequest("r", func(r *Request) error { r.Ready.WriterStopped = false; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AckRequestReady("r", a); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("ack while the preparer may still write: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Ready.WriterStopped = true; return nil })
	got, err := s.AckRequestReady("r", a)
	if err != nil || got.Stage != StageReadyAcked || got.AckedBy != a.SessionID {
		t.Fatalf("ack: %+v %v", got, err)
	}
	if _, err := s.AckRequestReady("r", a); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("a second ack: %v", err)
	}
}

func TestReadyGateNeedsTheWholeChainAndTheSameAgent(t *testing.T) {
	s := open(t)
	if err := s.ReadyGate(42, "anyone"); err != nil {
		t.Errorf("an issue no request names is not gated: %v", err)
	}
	_, _, _ = s.PutRequest("r", "h", 1, pol)
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Issue = 42; return nil })
	for _, st := range []Stage{StageIssueConfirmed, StageBranchRegistered, StageAgentReserved, StageWTPrepared} {
		if st == StageAgentReserved {
			a, _ := s.RegisterAgent(sid(42))
			_, _ = s.ReserveRequest("r", a, slots)
		} else if st == StageWTPrepared {
			_, _ = s.UpdateRequest("r", func(r *Request) error {
				r.Stage, r.Ready = st, &Ready{Host: "h", Slot: r.Slot, Path: "/wt", Head: sha40, WriterStopped: true}
				return nil
			})
		} else {
			_, _ = s.UpdateRequest("r", func(r *Request) error { r.Stage = st; r.Branch = &Branch{Ref: "WT-x", BaseOID: sha40}; return nil })
		}
		if err := s.ReadyGate(42, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
			t.Errorf("before READY_ACKED (%s): %v", st, err)
		}
	}
	a, _ := s.RegisterAgent(sid(42))
	if _, err := s.AckRequestReady("r", a); err != nil {
		t.Fatal(err)
	}
	if err := s.ReadyGate(42, sid(42)); err != nil {
		t.Errorf("acked by the starting agent: %v", err)
	}
	if err := s.ReadyGate(42, sid(7)); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("another agent starting: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Blocked = &Blocked{Step: StageReadyAcked, Detail: "x"}; return nil })
	if err := s.ReadyGate(42, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("blocked: %v", err)
	}
	_, _ = s.UpdateRequest("r", func(r *Request) error { r.Blocked = nil; r.Parked = ParkNoSlot; return nil })
	if err := s.ReadyGate(42, sid(42)); !errors.Is(err, core.ErrTransitionInvalid) {
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
	if got, _ := s.PromoteParked(p, 0); len(got) != 0 {
		t.Fatalf("promoted under full capacity: %+v", got)
	}
	startViaClaim(t, s, "run", 21)
	got, err := s.PromoteParked(p, 0)
	if err != nil || len(got) != 1 || got[0].Key != "high" {
		t.Fatalf("first promotion: %+v %v", got, err)
	}
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
	if err != nil || len(views) != 2 || views[0].Key != "a" || !views[0].Started || views[1].Started {
		t.Fatalf("views: %+v %v", views, err)
	}
	if all, _ := s.ListRequests(); all[0].Started {
		t.Errorf("the stored request must stay raw")
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

// Records written by the first (S46) layout are never raised above the highest
// S51 stage whose evidence they carry.
func TestLegacyRecordsAreMigratedWithoutInventingEvidence(t *testing.T) {
	cases := []struct {
		name       string
		file       string
		wantStage  Stage
		wantIssue  int
		wantUnk    bool
		wantSlot   string
		wantReady  bool
		wantDetail string
	}{
		{"ACK", `{"key":"k","payload_hash":"h","priority":1,"seq":1,"stage":"ACK","evidence":[{"stage":"ACK","ref":"accepted"}]}`, StageAccepted, 0, false, "", false, ""},
		{"ISSUE_PUBLISHED", `{"key":"k","payload_hash":"h","priority":1,"seq":1,"stage":"ISSUE_PUBLISHED","issue":7,"evidence":[{"stage":"ACK","ref":"accepted"},{"stage":"ISSUE_PUBLISHED","ref":"r"}]}`, StageAccepted, 7, true, "", false, "without a confirming lookup"},
		{"WT_READY", `{"key":"k","payload_hash":"h","priority":1,"seq":1,"stage":"WT_READY","issue":7,"slot":"wt-2","ready":{"slot":"wt-2","path":"/wt","head":"` + sha40 + `","writer_stopped":true},"evidence":[{"stage":"WT_READY","ref":"/wt"}]}`, StageAccepted, 7, true, "wt-2", true, "no branch or agent evidence"},
		{"ACK with a failed publish", `{"key":"k","payload_hash":"h","priority":1,"seq":1,"stage":"ACK","attempts":1,"retryable":true,"last_error":"timeout","evidence":[]}`, StageAccepted, 0, true, "", false, "timeout"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := open(t)
			if err := os.MkdirAll(s.requestsDir(), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(s.requestPath("k"), []byte(c.file), 0o600); err != nil {
				t.Fatal(err)
			}
			r, err := s.GetRequest("k")
			if err != nil {
				t.Fatal(err)
			}
			unknown := r.Blocked != nil && r.Blocked.Unknown
			if r.Stage != c.wantStage || r.Issue != c.wantIssue || unknown != c.wantUnk || r.Slot != c.wantSlot || (r.Ready != nil) != c.wantReady {
				t.Fatalf("migrated: %+v blocked=%+v", r, r.Blocked)
			}
			if c.wantDetail != "" && (r.Blocked == nil || !contains(r.Blocked.Detail, c.wantDetail)) {
				t.Errorf("detail = %+v, want %q", r.Blocked, c.wantDetail)
			}
			if r.LegacyRetryable || r.LegacyLastError != "" {
				t.Error("legacy fields must be cleared once migrated")
			}
			if r.Branch != nil || r.Agent != "" || r.AckedBy != "" {
				t.Error("a legacy record has no branch, agent or ack evidence and must not gain any")
			}
			// The old evidence stays as written.
			if len(r.Evidence) != len(parsedEvidence(t, c.file)) {
				t.Errorf("evidence was rewritten: %+v", r.Evidence)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func parsedEvidence(t *testing.T, file string) []StageEvidence {
	t.Helper()
	var r struct {
		Evidence []StageEvidence `json:"evidence"`
	}
	if err := jsonUnmarshal([]byte(file), &r); err != nil {
		t.Fatal(err)
	}
	return r.Evidence
}

// A migrated record goes on from where its evidence allows: the lookup confirms
// the issue, and nothing is created a second time.
func TestMigratedRecordIsWrittenBackInTheNewLayout(t *testing.T) {
	s := open(t)
	_ = os.MkdirAll(s.requestsDir(), 0o700)
	_ = os.WriteFile(s.requestPath("k"), []byte(`{"key":"k","payload_hash":"h","priority":1,"seq":1,"stage":"ISSUE_PUBLISHED","issue":7,"evidence":[]}`), 0o600)
	out, err := s.UpdateRequest("k", func(r *Request) error {
		r.Stage, r.Blocked = StageIssueConfirmed, nil
		return nil
	})
	if err != nil || out.Stage != StageIssueConfirmed || out.Issue != 7 {
		t.Fatalf("update after migration: %+v %v", out, err)
	}
	b, _ := os.ReadFile(s.requestPath("k"))
	if contains(string(b), `"ISSUE_PUBLISHED"`) || contains(string(b), "last_error") || contains(string(b), `"retryable"`) {
		t.Errorf("legacy names persisted:\n%s", b)
	}
}

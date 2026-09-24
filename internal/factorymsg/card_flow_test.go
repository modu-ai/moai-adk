package factorymsg

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// cardFlowCombo is one lead/worker backend pairing of the deterministic card
// flow. The subtest name is "<lead>-<worker>".
type cardFlowCombo struct{ lead, worker string }

// newCardFlowFixture builds a dispatch fixture whose lead carries the given
// backend label instead of the default one.
func newCardFlowFixture(t *testing.T, leadBackend string) *dispatchFixture {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	f := &dispatchFixture{t: t, root: t.TempDir(), live: map[string]bool{}, slots: map[string]string{}}
	f.open()
	t.Cleanup(func() { _ = f.s.Close() })
	f.lead = registerBackend(f, leadBackend, "lead", "lead", "lead-s1", "start-lead")
	return f
}

// registerBackend registers a live peer for slot under the given backend and
// proves the broker recorded that backend on the current lane row.
func registerBackend(f *dispatchFixture, backend, slot, role, session, start string) Peer {
	f.t.Helper()
	f.live[start] = true
	p := Peer{ProjectKey: "project", RunID: "run", Backend: backend, Role: role, Slot: slot, SessionUUID: session, Generation: 1, PID: os.Getpid(), ProcessStart: start}
	got, err := f.s.RegisterPeer(context.Background(), p)
	if err != nil {
		f.t.Fatalf("register %s/%s: %v", backend, slot, err)
	}
	lane, err := f.s.ResolveLane(context.Background(), got.Slot)
	if err != nil || lane.Backend != backend {
		f.t.Fatalf("lane %s backend=%q err=%v, want %q", got.Slot, lane.Backend, err, backend)
	}
	f.slots[got.SessionUUID] = got.Slot
	return got
}

// TestFactoryCardFlowFourCombinations covers AC-DHR-017 (REQ-DHR-022): each
// backend pairing drives one dispatch through assigned -> delivered ->
// started -> result_recorded -> integrated against the real broker store,
// with a worker interruption after started, a reassignment to attempt 2, and
// late results from the interrupted attempt. Exactly one result is applied.
func TestFactoryCardFlowFourCombinations(t *testing.T) {
	for _, c := range []cardFlowCombo{
		{lead: "claude", worker: "claude"},
		{lead: "codex", worker: "codex"},
		{lead: "claude", worker: "codex"},
		{lead: "codex", worker: "claude"},
	} {
		t.Run(c.lead+"-"+c.worker, func(t *testing.T) { runCardFlow(t, c) })
	}
}

func runCardFlow(t *testing.T, c cardFlowCombo) {
	const id = "card-d1"
	ctx := context.Background()
	f := newCardFlowFixture(t, c.lead)
	first := registerBackend(f, c.worker, "lane-1", "worker", "w1-s1", "start-w1")
	second := registerBackend(f, c.worker, "lane-2", "worker", "w2-s1", "start-w2")

	// accepted counts every ApplyResult call that reported acceptance for
	// this dispatch; exactly-once means it ends at 1.
	accepted := 0
	applyReport := func(step string, r ResultReport, want string) {
		t.Helper()
		got := f.apply(r)
		if got == ApplyAccepted {
			accepted++
		}
		requireOutcome(t, step, got, want)
	}

	// Attempt 1: assign, deliver, start. A message receipt alone never
	// advances the record past what the explicit call set.
	if _, err := f.s.CreateDispatch(ctx, id, "t1100", first); err != nil {
		t.Fatalf("create: %v", err)
	}
	assign1, err := f.s.Send(ctx, SendRequest{From: f.lead, To: first, Kind: KindDispatchNotice, IdempotencyKey: AssignmentKey(id, 1), TaskRef: "t1100", CorrelationID: "a1", TTL: time.Hour, Payload: []byte("assign attempt 1")})
	if err != nil {
		t.Fatalf("send assignment 1: %v", err)
	}
	f.receiveAll(first, assign1.ID)
	requireState(t, f, id, "after assignment receipt", DispatchAssigned)
	if _, err := f.s.MarkDispatchDelivered(ctx, first, id, 1); err != nil {
		t.Fatalf("delivered 1: %v", err)
	}
	nudge, err := f.s.Send(ctx, SendRequest{From: f.lead, To: first, Kind: KindStatusRequest, IdempotencyKey: "nudge-1", TaskRef: "t1100", CorrelationID: "n1", TTL: time.Hour, Payload: []byte("status?")})
	if err != nil {
		t.Fatalf("send nudge: %v", err)
	}
	f.receiveAll(first, nudge.ID)
	requireState(t, f, id, "after a further message receipt", DispatchDelivered)
	if _, err := f.s.StartDispatch(ctx, first, id, 1); err != nil {
		t.Fatalf("start 1: %v", err)
	}
	requireState(t, f, id, "after start", DispatchStarted)

	// The first worker reports once, then is interrupted: its process is gone.
	late, err := f.sendResult(first, id, 1, "attempt 1 result")
	if err != nil {
		t.Fatalf("attempt 1 result: %v", err)
	}
	f.live[first.ProcessStart] = false

	// Reassignment to a new attempt on another lane of the same backend.
	d, err := f.s.ReassignDispatch(ctx, id, 1, second, false)
	if err != nil {
		t.Fatalf("reassign after interruption: %v", err)
	}
	if d.Attempt != 2 || d.LaneSlot != second.Slot || d.State != DispatchAssigned {
		t.Fatalf("reassigned record: %+v", d)
	}
	// The interrupted attempt's owner is fenced from every assignee transition.
	if _, err := f.s.StartDispatch(ctx, first, id, 1); !errors.Is(err, ErrDispatchFenced) {
		t.Fatalf("interrupted attempt 1 start: %v, want ErrDispatchFenced", err)
	}
	// Its lane comes back under a newer generation; the old endpoint is stale
	// at the broker boundary (verifyPeerOn inside the dispatch transaction).
	back := f.restart(first, "w1-s2", "start-w1b")
	if back.Backend != c.worker {
		t.Fatalf("restarted lane backend %q, want %q", back.Backend, c.worker)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, first, id, 2); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("pre-restart endpoint transition: %v, want ErrStalePeer", err)
	}

	assign2, err := f.s.Send(ctx, SendRequest{From: f.lead, To: second, Kind: KindDispatchNotice, IdempotencyKey: AssignmentKey(id, 2), TaskRef: "t1100", CorrelationID: "a2", TTL: time.Hour, Payload: []byte("assign attempt 2")})
	if err != nil {
		t.Fatalf("send assignment 2: %v", err)
	}
	f.receiveAll(second, assign2.ID)
	requireState(t, f, id, "after attempt 2 assignment receipt", DispatchAssigned)
	if _, err := f.s.MarkDispatchDelivered(ctx, second, id, 2); err != nil {
		t.Fatalf("delivered 2: %v", err)
	}
	if _, err := f.s.StartDispatch(ctx, second, id, 2); err != nil {
		t.Fatalf("start 2: %v", err)
	}

	// The interrupted attempt's result reaches the lead while attempt 2 is
	// started and has not reported: it must not be applied.
	lateClaim := f.claimOne(time.Minute)
	if lateClaim.ID != late.ID {
		t.Fatalf("lead claimed %s, want the late attempt 1 result %s", lateClaim.ID, late.ID)
	}
	lateReport := reportFromClaim(f, lateClaim, id, 1)
	before := f.dispatch(id)
	applyReport("late attempt 1 before attempt 2 reports", lateReport, ApplyStale)
	requireUnchanged(t, "late attempt 1 before attempt 2 reports", before, f.dispatch(id))
	f.settle(lateClaim, DispositionAccepted)

	// Attempt 2 reports. Its arrival and receipt alone record nothing.
	res2, err := f.sendResult(second, id, 2, "attempt 2 result")
	if err != nil {
		t.Fatalf("attempt 2 result: %v", err)
	}
	c2 := f.claimOne(time.Minute)
	if c2.ID != res2.ID {
		t.Fatalf("lead claimed %s, want attempt 2 result %s", c2.ID, res2.ID)
	}
	requireState(t, f, id, "after attempt 2 result arrival", DispatchStarted)
	report2 := reportFromClaim(f, c2, id, 2)
	applyReport("attempt 2 result", report2, ApplyAccepted)
	recorded := f.dispatch(id)
	requireRecorded(t, recorded, c2.ID)
	f.settle(c2, DispositionAccepted)

	// The same attempt 2 result redelivered, and the late attempt 1 result
	// replayed from its original envelope, both leave the record untouched.
	applyReport("attempt 2 redelivery", report2, ApplyDuplicate)
	applyReport("late attempt 1 after attempt 2 reports", lateReport, ApplyStale)
	requireUnchanged(t, "post-record replays", recorded, f.dispatch(id))

	// The lead integrates; nothing reopens the record afterwards.
	if d, err = f.s.IntegrateDispatch(ctx, id); err != nil || d.State != DispatchIntegrated {
		t.Fatalf("integrate: %+v %v", d, err)
	}
	integrated := f.dispatch(id)
	applyReport("attempt 2 replay after integrated", report2, ApplyDuplicate)
	applyReport("late attempt 1 after integrated", lateReport, ApplyStale)
	requireUnchanged(t, "post-integration replays", integrated, f.dispatch(id))

	if accepted != 1 {
		t.Fatalf("applied results for %s = %d, want exactly 1", id, accepted)
	}
	if integrated.ResultAttempt != 2 || integrated.ResultRef != c2.ID || integrated.ResultDigest != ResultDigest([]byte("attempt 2 result")) {
		t.Fatalf("applied result must be attempt 2's only: %+v", integrated)
	}
}

// reportFromClaim is the receiver's view of a claimed result: identity from
// the envelope, digest from the body read under the claim. Replays reuse it,
// as a redelivered envelope carries the same identity and body.
func reportFromClaim(f *dispatchFixture, cl Claim, id string, attempt int64) ResultReport {
	f.t.Helper()
	body, err := f.s.ReadBody(context.Background(), f.lead, cl.ID, cl.ClaimToken)
	if err != nil {
		f.t.Fatal(err)
	}
	if want := f.slots[cl.SenderSession]; cl.SenderSlot != want {
		f.t.Fatalf("envelope sender slot %q, want %q", cl.SenderSlot, want)
	}
	return ResultReport{DispatchID: id, Attempt: attempt, Slot: cl.SenderSlot, Generation: cl.SenderGeneration, Digest: ResultDigest(body), Ref: cl.ID}
}

func requireState(t *testing.T, f *dispatchFixture, id, step, want string) {
	t.Helper()
	if got := f.dispatch(id).State; got != want {
		t.Fatalf("%s: dispatch state %q, want %q", step, got, want)
	}
}

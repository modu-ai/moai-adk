package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"
)

// dispatchFixture drives the dispatch record through the real broker store.
// Liveness is controlled by the set of process-start values considered live,
// so a "restart" is a re-registration under a new start while the old one is
// marked dead.
type dispatchFixture struct {
	t     *testing.T
	root  string
	s     *Store
	live  map[string]bool
	slots map[string]string
	lead  Peer
}

func newDispatchFixture(t *testing.T) *dispatchFixture {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	f := &dispatchFixture{t: t, root: t.TempDir(), live: map[string]bool{}, slots: map[string]string{}}
	f.open()
	t.Cleanup(func() { _ = f.s.Close() })
	f.lead = f.register("lead", "lead", "lead-s1", "start-lead")
	return f
}

func (f *dispatchFixture) open() {
	f.t.Helper()
	s, err := Open(f.root, "run")
	if err != nil {
		f.t.Fatal(err)
	}
	s.ownerCurrent = func(pid int, start string) bool { return pid == os.Getpid() && f.live[start] }
	f.s = s
}

// reopen simulates a broker process restart: the store is closed and a new
// one is opened on the same database file.
func (f *dispatchFixture) reopen() {
	f.t.Helper()
	if err := f.s.Close(); err != nil {
		f.t.Fatal(err)
	}
	f.open()
}

func (f *dispatchFixture) register(slot, role, session, start string) Peer {
	f.t.Helper()
	f.live[start] = true
	p := Peer{ProjectKey: "project", RunID: "run", Backend: "codex", Role: role, Slot: slot, SessionUUID: session, Generation: 1, PID: os.Getpid(), ProcessStart: start}
	got, err := f.s.RegisterPeer(context.Background(), p)
	if err != nil {
		f.t.Fatalf("register %s: %v", slot, err)
	}
	f.slots[got.SessionUUID] = got.Slot
	return got
}

// restart marks the current owner of p dead and re-registers the same slot
// under a new session UUID and process start, advancing the generation.
func (f *dispatchFixture) restart(p Peer, session, start string) Peer {
	f.t.Helper()
	f.live[p.ProcessStart] = false
	next := p
	next.SessionUUID = session
	next.ProcessStart = start
	f.live[start] = true
	got, err := f.s.RegisterPeer(context.Background(), next)
	if err != nil {
		f.t.Fatalf("restart %s: %v", p.Slot, err)
	}
	if got.Generation <= p.Generation {
		f.t.Fatalf("restart did not advance generation: %d -> %d", p.Generation, got.Generation)
	}
	f.slots[got.SessionUUID] = got.Slot
	return got
}

func (f *dispatchFixture) dispatch(id string) Dispatch {
	f.t.Helper()
	d, err := f.s.Dispatch(context.Background(), id)
	if err != nil {
		f.t.Fatalf("read dispatch %s: %v", id, err)
	}
	return d
}

func (f *dispatchFixture) messageRows(key string) int {
	f.t.Helper()
	var n int
	if err := f.s.db.QueryRowContext(context.Background(), `SELECT count(*) FROM messages WHERE idem_key=?`, key).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

// assign creates the dispatch, delivers the assignment message to the worker,
// and walks it to started. It also proves a message receipt alone does not
// advance the record.
func (f *dispatchFixture) assign(id string, worker Peer) {
	f.t.Helper()
	ctx := context.Background()
	if _, err := f.s.CreateDispatch(ctx, id, "t1100", worker); err != nil {
		f.t.Fatalf("create dispatch: %v", err)
	}
	msg, err := f.s.Send(ctx, SendRequest{From: f.lead, To: worker, Kind: KindDispatchNotice, IdempotencyKey: AssignmentKey(id, 1), TaskRef: "t1100", CorrelationID: "a-" + id, TTL: time.Hour, Payload: []byte("assign " + id)})
	if err != nil {
		f.t.Fatalf("send assignment: %v", err)
	}
	f.receiveAll(worker, msg.ID)
	if got := f.dispatch(id).State; got != DispatchAssigned {
		f.t.Fatalf("a message receipt moved the dispatch to %q", got)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, worker, id, 1); err != nil {
		f.t.Fatalf("delivered: %v", err)
	}
	if _, err := f.s.StartDispatch(ctx, worker, id, 1); err != nil {
		f.t.Fatalf("start: %v", err)
	}
}

// receiveAll claims, disposes, and receipts every pending message for p and
// fails unless want is among them.
func (f *dispatchFixture) receiveAll(p Peer, want string) {
	f.t.Helper()
	ctx := context.Background()
	claims, err := f.s.Claim(ctx, p, MaxBatch, time.Minute)
	if err != nil {
		f.t.Fatal(err)
	}
	seen := false
	for _, c := range claims {
		seen = seen || c.ID == want
		if err := f.s.RecordDisposition(ctx, p, c.ID, c.ClaimToken, DispositionAccepted); err != nil {
			f.t.Fatal(err)
		}
		if err := f.s.Receipt(ctx, p, c.ID, c.ClaimToken); err != nil {
			f.t.Fatal(err)
		}
	}
	if !seen {
		f.t.Fatalf("message %s was not delivered to %s", want, p.Slot)
	}
}

func (f *dispatchFixture) sendResult(from Peer, id string, attempt int64, body string) (Envelope, error) {
	return f.s.Send(context.Background(), SendRequest{From: from, To: f.lead, Kind: KindStatusReport, IdempotencyKey: ResultKey(id, attempt), TaskRef: "t1100", CorrelationID: "r-" + id, TTL: time.Hour, Payload: []byte(body)})
}

// claimOne claims exactly one message for the lead.
func (f *dispatchFixture) claimOne(lease time.Duration) Claim {
	f.t.Helper()
	claims, err := f.s.Claim(context.Background(), f.lead, MaxBatch, lease)
	if err != nil || len(claims) != 1 {
		f.t.Fatalf("lead claim=%+v err=%v", claims, err)
	}
	return claims[0]
}

// applyFromClaim is the receiver's result application: the report identity is
// taken from the claimed envelope, the digest from the body it read.
func (f *dispatchFixture) applyFromClaim(c Claim, id string, attempt int64) string {
	f.t.Helper()
	body, err := f.s.ReadBody(context.Background(), f.lead, c.ID, c.ClaimToken)
	if err != nil {
		f.t.Fatal(err)
	}
	if want := f.slots[c.SenderSession]; c.SenderSlot != want {
		f.t.Fatalf("envelope sender slot %q, want %q", c.SenderSlot, want)
	}
	out, err := f.s.ApplyResult(context.Background(), ResultReport{DispatchID: id, Attempt: attempt, Slot: c.SenderSlot, Generation: c.SenderGeneration, Digest: ResultDigest(body), Ref: c.ID})
	if err != nil {
		f.t.Fatal(err)
	}
	return out
}

func (f *dispatchFixture) apply(r ResultReport) string {
	f.t.Helper()
	out, err := f.s.ApplyResult(context.Background(), r)
	if err != nil {
		f.t.Fatal(err)
	}
	return out
}

func (f *dispatchFixture) settle(c Claim, disposition string) {
	f.t.Helper()
	ctx := context.Background()
	if err := f.s.RecordDisposition(ctx, f.lead, c.ID, c.ClaimToken, disposition); err != nil {
		f.t.Fatal(err)
	}
	if err := f.s.Receipt(ctx, f.lead, c.ID, c.ClaimToken); err != nil {
		f.t.Fatal(err)
	}
}

func requireOutcome(t *testing.T, step, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: outcome %q, want %q", step, got, want)
	}
}

func requireUnchanged(t *testing.T, step string, before, after Dispatch) {
	t.Helper()
	if before != after {
		t.Fatalf("%s changed the dispatch record:\nbefore=%+v\nafter =%+v", step, before, after)
	}
}

func requireRecorded(t *testing.T, d Dispatch, ref string) {
	t.Helper()
	if d.State != DispatchResultRecorded || d.ResultAttempt != d.Attempt || d.ResultRef != ref {
		t.Fatalf("dispatch not recorded once with ref %s: %+v", ref, d)
	}
}

// TestDispatchResultExactlyOnce covers AC-DHR-014: redelivery, lost receipt,
// sender restart, and broker restart each leave exactly one applied result.
func TestDispatchResultExactlyOnce(t *testing.T) {
	const id = "d-1"
	const body = "result body"

	t.Run("duplicate_delivery", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Minute)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		first := f.dispatch(id)
		requireRecorded(t, first, c.ID)
		requireOutcome(t, "second", f.applyFromClaim(c, id, 1), ApplyDuplicate)
		requireUnchanged(t, "second delivery", first, f.dispatch(id))
		f.settle(c, DispositionAccepted)
	})

	t.Run("lost_receipt_redelivery", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Millisecond)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		first := f.dispatch(id)
		requireRecorded(t, first, c.ID)
		// The result is persisted while the message is still unacknowledged.
		var state string
		if err := f.s.db.QueryRowContext(context.Background(), `SELECT state FROM messages WHERE id=?`, c.ID).Scan(&state); err != nil || state != "claimed" {
			t.Fatalf("result must be recorded before the receipt: message state=%q err=%v", state, err)
		}
		time.Sleep(3 * time.Millisecond)
		again := f.claimOne(time.Minute)
		if again.ID != c.ID || again.ClaimToken == c.ClaimToken {
			t.Fatalf("lease redelivery: %+v", again)
		}
		requireOutcome(t, "redelivery", f.applyFromClaim(again, id, 1), ApplyDuplicate)
		requireUnchanged(t, "redelivery", first, f.dispatch(id))
		f.settle(again, DispositionDuplicate)
	})

	t.Run("sender_restart_after_result", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Minute)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		f.settle(c, DispositionAccepted)
		first := f.dispatch(id)
		restarted := f.restart(worker, "w-s2", "start-w2")
		if _, err := f.sendResult(restarted, id, 1, body); err != nil {
			t.Fatalf("same-request retry after restart: %v", err)
		}
		// Whatever the message layer delivers again is judged a duplicate.
		claims, err := f.s.Claim(context.Background(), f.lead, MaxBatch, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		for _, again := range claims {
			requireOutcome(t, "redelivered retry", f.applyFromClaim(again, id, 1), ApplyDuplicate)
			f.settle(again, DispositionDuplicate)
		}
		requireOutcome(t, "restarted reporter", f.apply(ResultReport{DispatchID: id, Attempt: 1, Slot: restarted.Slot, Generation: restarted.Generation, Digest: ResultDigest([]byte(body)), Ref: "retry"}), ApplyDuplicate)
		requireUnchanged(t, "retry after restart", first, f.dispatch(id))
	})

	t.Run("sender_restart_lane_scope", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		orig, err := f.sendResult(worker, id, 1, body)
		if err != nil {
			t.Fatal(err)
		}
		restarted := f.restart(worker, "w-s2", "start-w2")
		retry, err := f.sendResult(restarted, id, 1, body)
		if err != nil {
			t.Fatalf("same-scope retry after restart: %v", err)
		}
		if retry.ID != orig.ID || f.messageRows(ResultKey(id, 1)) != 1 {
			t.Fatalf("lane-scoped retry must return the original message: orig=%s retry=%s rows=%d", orig.ID, retry.ID, f.messageRows(ResultKey(id, 1)))
		}
		claims, err := f.s.Claim(context.Background(), f.lead, MaxBatch, time.Minute)
		if err != nil || len(claims) != 1 || claims[0].ID != orig.ID {
			t.Fatalf("recipient must see exactly the original message: %+v err=%v", claims, err)
		}
	})

	t.Run("broker_restart", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Millisecond)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		first := f.dispatch(id)
		f.reopen()
		time.Sleep(3 * time.Millisecond)
		again := f.claimOne(time.Minute)
		if again.ID != c.ID {
			t.Fatalf("redelivery after broker restart: %+v", again)
		}
		requireOutcome(t, "after broker restart", f.applyFromClaim(again, id, 1), ApplyDuplicate)
		requireUnchanged(t, "broker restart", first, f.dispatch(id))
		f.settle(again, DispositionDuplicate)
	})

	t.Run("old_generation_redelivery", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Minute)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		first := f.dispatch(id)
		restarted := f.restart(worker, "w-s2", "start-w2")
		requireOutcome(t, "restarted reporter", f.apply(ResultReport{DispatchID: id, Attempt: 1, Slot: restarted.Slot, Generation: restarted.Generation, Digest: ResultDigest([]byte(body)), Ref: "retry"}), ApplyDuplicate)
		// The generation-1 message processed again after the restart.
		requireOutcome(t, "old generation", f.applyFromClaim(c, id, 1), ApplyStale)
		requireUnchanged(t, "old generation", first, f.dispatch(id))
	})

	t.Run("collision", func(t *testing.T) {
		f := newDispatchFixture(t)
		worker := f.register("lane-1", "worker", "w-s1", "start-w1")
		f.assign(id, worker)
		if _, err := f.sendResult(worker, id, 1, body); err != nil {
			t.Fatal(err)
		}
		c := f.claimOne(time.Minute)
		requireOutcome(t, "first", f.applyFromClaim(c, id, 1), ApplyAccepted)
		first := f.dispatch(id)
		if _, err := f.sendResult(worker, id, 1, "different body"); err == nil {
			t.Fatal("message layer accepted a different body under the same key")
		}
		if n := f.messageRows(ResultKey(id, 1)); n != 1 {
			t.Fatalf("rejected send mutated the queue: rows=%d", n)
		}
		requireOutcome(t, "different digest", f.apply(ResultReport{DispatchID: id, Attempt: 1, Slot: worker.Slot, Generation: worker.Generation, Digest: ResultDigest([]byte("different body")), Ref: "other"}), ApplyCollision)
		requireUnchanged(t, "collision", first, f.dispatch(id))
	})
}

// TestDispatchResultOrderTable pins every row of the result-application order
// and that no outcome other than accepted touches the record.
func TestDispatchResultOrderTable(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	worker := f.register("lane-1", "worker", "w-s1", "start-w1")
	other := f.register("lane-2", "worker", "o-s1", "start-o1")
	digest := ResultDigest([]byte("body"))
	report := func(mut func(*ResultReport)) ResultReport {
		r := ResultReport{DispatchID: "d-t", Attempt: 1, Slot: worker.Slot, Generation: worker.Generation, Digest: digest, Ref: "ref-1"}
		if mut != nil {
			mut(&r)
		}
		return r
	}
	requireOutcome(t, "no record", f.apply(report(nil)), ApplyUnknown)
	if _, err := f.s.CreateDispatch(ctx, "d-t", "t1100", worker); err != nil {
		t.Fatal(err)
	}
	before := f.dispatch("d-t")
	requireOutcome(t, "not started", f.apply(report(nil)), ApplyInvalidState)
	requireUnchanged(t, "invalid-state", before, f.dispatch("d-t"))
	if _, err := f.s.MarkDispatchDelivered(ctx, worker, "d-t", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.StartDispatch(ctx, worker, "d-t", 1); err != nil {
		t.Fatal(err)
	}
	before = f.dispatch("d-t")
	for name, r := range map[string]ResultReport{
		"other attempt":     report(func(r *ResultReport) { r.Attempt = 2 }),
		"other lane":        report(func(r *ResultReport) { r.Slot = other.Slot }),
		"older generation":  report(func(r *ResultReport) { r.Generation = 0 }),
		"future generation": report(func(r *ResultReport) { r.Generation = worker.Generation + 1 }),
	} {
		requireOutcome(t, name, f.apply(r), ApplyStale)
		requireUnchanged(t, name, before, f.dispatch("d-t"))
	}
	requireOutcome(t, "accepted", f.apply(report(nil)), ApplyAccepted)
	for _, bad := range []ResultReport{
		report(func(r *ResultReport) { r.DispatchID = "../x" }),
		report(func(r *ResultReport) { r.Digest = "" }),
		report(func(r *ResultReport) { r.Ref = "" }),
	} {
		if _, err := f.s.ApplyResult(ctx, bad); err == nil {
			t.Fatalf("invalid report accepted: %+v", bad)
		}
	}
}

// TestDispatchLifecycleTransitions pins the explicit, fenced transitions.
func TestDispatchLifecycleTransitions(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	worker := f.register("lane-1", "worker", "w-s1", "start-w1")
	other := f.register("lane-2", "worker", "o-s1", "start-o1")
	d, err := f.s.CreateDispatch(ctx, "d-l", "t1100", worker)
	if err != nil {
		t.Fatal(err)
	}
	if d.Attempt != 1 || d.AssigneeGeneration != worker.Generation || d.LaneSlot != worker.Slot || d.State != DispatchAssigned || d.CardID != "t1100" {
		t.Fatalf("created dispatch: %+v", d)
	}
	if _, err := f.s.CreateDispatch(ctx, "d-l", "t1100", worker); err == nil {
		t.Fatal("duplicate dispatch id accepted")
	}
	if _, err := f.s.StartDispatch(ctx, worker, "d-l", 1); !errors.Is(err, ErrDispatchState) {
		t.Fatalf("start before delivered: %v", err)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, other, "d-l", 1); !errors.Is(err, ErrDispatchFenced) {
		t.Fatalf("delivered by another lane: %v", err)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, worker, "d-l", 2); !errors.Is(err, ErrDispatchFenced) {
		t.Fatalf("delivered for another attempt: %v", err)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, worker, "missing", 1); !errors.Is(err, ErrDispatchNotFound) {
		t.Fatalf("delivered for a missing dispatch: %v", err)
	}
	if _, err := f.s.IntegrateDispatch(ctx, "d-l"); !errors.Is(err, ErrDispatchState) {
		t.Fatalf("integrate before a result: %v", err)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, worker, "d-l", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.StartDispatch(ctx, worker, "d-l", 1); err != nil {
		t.Fatal(err)
	}
	requireOutcome(t, "result", f.apply(ResultReport{DispatchID: "d-l", Attempt: 1, Slot: worker.Slot, Generation: worker.Generation, Digest: ResultDigest([]byte("b")), Ref: "ref"}), ApplyAccepted)
	if _, err := f.s.AbandonDispatch(ctx, "d-l"); !errors.Is(err, ErrDispatchState) {
		t.Fatalf("abandon after a result: %v", err)
	}
	if d, err = f.s.IntegrateDispatch(ctx, "d-l"); err != nil || d.State != DispatchIntegrated {
		t.Fatalf("integrate: %+v %v", d, err)
	}
	if _, err := f.s.CreateDispatch(ctx, "d-a", "t1100", worker); err != nil {
		t.Fatal(err)
	}
	if d, err = f.s.AbandonDispatch(ctx, "d-a"); err != nil || d.State != DispatchAbandoned {
		t.Fatalf("abandon: %+v %v", d, err)
	}
	if _, err := f.s.ReassignDispatch(ctx, "d-a", 1, other, true); !errors.Is(err, ErrDispatchState) {
		t.Fatalf("reassign an abandoned dispatch: %v", err)
	}
	if AssignmentKey("d-1", 2) != "dispatch:d-1:2" || ResultKey("d-1", 2) != "result:d-1:2" {
		t.Fatalf("key shape: %s %s", AssignmentKey("d-1", 2), ResultKey("d-1", 2))
	}
}

// TestDispatchStaleFencingAndSuperseded covers AC-DHR-015.
func TestDispatchStaleFencingAndSuperseded(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	gen1 := f.register("lane-1", "worker", "w-s1", "start-w1")
	f.assign("d-s", gen1)
	pending, err := f.s.Send(ctx, SendRequest{From: f.lead, To: gen1, Kind: KindStatusRequest, IdempotencyKey: "left-pending", TaskRef: "t1100", CorrelationID: "c-pending", TTL: time.Hour, Payload: []byte("pending for generation 1")})
	if err != nil {
		t.Fatal(err)
	}
	claimedMsg, err := f.s.Send(ctx, SendRequest{From: f.lead, To: gen1, Kind: KindStatusRequest, IdempotencyKey: "left-claimed", TaskRef: "t1100", CorrelationID: "c-claimed", TTL: time.Hour, Payload: []byte("claimed by generation 1")})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := f.s.Claim(ctx, gen1, 1, time.Hour)
	if err != nil || len(claims) != 1 {
		t.Fatalf("generation 1 claim: %+v %v", claims, err)
	}
	held := claims[0]
	if held.ID != pending.ID && held.ID != claimedMsg.ID {
		t.Fatalf("unexpected claim %s", held.ID)
	}

	gen2 := f.restart(gen1, "w-s2", "start-w2")
	snapshot := func() string {
		rows, err := f.s.db.QueryContext(ctx, `SELECT id,state,claim_token,disposition,COALESCE(acknowledged_at,'') FROM messages ORDER BY id`)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()
		var out string
		for rows.Next() {
			var id, state, token, disp, ack string
			if err := rows.Scan(&id, &state, &token, &disp, &ack); err != nil {
				t.Fatal(err)
			}
			out += id + "|" + state + "|" + token + "|" + disp + "|" + ack + ";"
		}
		return out
	}
	beforeMsgs, beforeRec := snapshot(), f.dispatch("d-s")

	digest := ResultDigest([]byte("gen1 result"))
	requireOutcome(t, "generation 1 result", f.apply(ResultReport{DispatchID: "d-s", Attempt: 1, Slot: gen1.Slot, Generation: gen1.Generation, Digest: digest, Ref: "g1"}), ApplyStale)
	if _, err := f.s.Claim(ctx, gen1, MaxBatch, time.Minute); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("generation 1 claim: %v", err)
	}
	if _, err := f.s.ReadBody(ctx, gen1, held.ID, held.ClaimToken); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("generation 1 read: %v", err)
	}
	if err := f.s.RecordDisposition(ctx, gen1, held.ID, held.ClaimToken, DispositionAccepted); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("generation 1 dispose: %v", err)
	}
	if err := f.s.Receipt(ctx, gen1, held.ID, held.ClaimToken); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("generation 1 ack: %v", err)
	}
	if _, err := f.s.MarkDispatchDelivered(ctx, gen1, "d-s", 1); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("generation 1 dispatch transition: %v", err)
	}
	requireUnchanged(t, "generation 1 attempts", beforeRec, f.dispatch("d-s"))
	if after := snapshot(); after != beforeMsgs {
		t.Fatalf("generation 1 attempts mutated messages:\nbefore=%s\nafter =%s", beforeMsgs, after)
	}

	st, err := f.s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.Superseded != 2 || st.Pending != 0 || st.Claimed != 0 {
		t.Fatalf("superseded messages must not count as pending or claimed: %+v", st)
	}
	fresh, err := f.s.Claim(ctx, gen2, MaxBatch, time.Minute)
	if err != nil || len(fresh) != 0 {
		t.Fatalf("broker released superseded bodies to generation 2: %+v %v", fresh, err)
	}

	// Before the regrant the new generation holds no dispatch authority.
	requireOutcome(t, "generation 2 before regrant", f.apply(ResultReport{DispatchID: "d-s", Attempt: 1, Slot: gen2.Slot, Generation: gen2.Generation, Digest: digest, Ref: "g2"}), ApplyStale)
	requireUnchanged(t, "generation 2 before regrant", beforeRec, f.dispatch("d-s"))

	// A failing caller change rolls the regrant back with it.
	failure := errors.New("caller state change failed")
	if _, err := f.s.RegrantDispatch(ctx, "d-s", 1, gen2, func(*sql.Tx) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("regrant with failing caller change: %v", err)
	}
	requireUnchanged(t, "rolled-back regrant", beforeRec, f.dispatch("d-s"))
	if _, err := f.s.RegrantDispatch(ctx, "d-s", 2, gen2, nil); !errors.Is(err, ErrDispatchFenced) {
		t.Fatalf("regrant for another attempt: %v", err)
	}

	called := false
	regranted, err := f.s.RegrantDispatch(ctx, "d-s", 1, gen2, func(tx *sql.Tx) error {
		called = true
		var n int
		return tx.QueryRowContext(ctx, `SELECT count(*) FROM dispatches WHERE dispatch_id='d-s' AND assignee_generation=?`, gen2.Generation).Scan(&n)
	})
	if err != nil || !called {
		t.Fatalf("regrant: %v called=%v", err, called)
	}
	if regranted.Attempt != 1 || regranted.AssigneeGeneration != gen2.Generation || regranted.State != DispatchStarted {
		t.Fatalf("regrant must keep the attempt and state: %+v", regranted)
	}
	requireOutcome(t, "generation 2 after regrant", f.apply(ResultReport{DispatchID: "d-s", Attempt: 1, Slot: gen2.Slot, Generation: gen2.Generation, Digest: digest, Ref: "g2"}), ApplyAccepted)
	if d := f.dispatch("d-s"); d.Attempt != 1 || d.State != DispatchResultRecorded || d.ResultRef != "g2" {
		t.Fatalf("result after regrant: %+v", d)
	}

	key := AssignmentKey("d-s", 1)
	rows := f.messageRows(key)
	if _, err := f.s.Send(ctx, SendRequest{From: f.lead, To: gen2, Kind: KindDispatchNotice, IdempotencyKey: key, TaskRef: "t1100", CorrelationID: "a-d-s", TTL: time.Hour, Payload: []byte("assign d-s")}); err == nil {
		t.Fatal("same-key resend to a different recipient generation accepted")
	}
	if got := f.messageRows(key); got != rows {
		t.Fatalf("rejected resend mutated the queue: %d -> %d", rows, got)
	}
	if _, err := f.s.Send(ctx, SendRequest{From: f.lead, To: gen2, Kind: KindDispatchNotice, IdempotencyKey: key + ":regrant", TaskRef: "t1100", CorrelationID: "a-d-s", TTL: time.Hour, Payload: []byte("assign d-s")}); err != nil {
		t.Fatalf("new-key resend after regrant: %v", err)
	}
}

// TestDispatchReassignmentFencesLateResult covers AC-DHR-016.
func TestDispatchReassignmentFencesLateResult(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	first := f.register("lane-1", "worker", "a-s1", "start-a1")
	second := f.register("lane-2", "worker", "b-s1", "start-b1")
	f.assign("d-r", first)
	before := f.dispatch("d-r")

	if _, err := f.s.ReassignDispatch(ctx, "d-r", 1, second, false); !errors.Is(err, ErrDispatchOwnerLive) {
		t.Fatalf("reassignment away from a live owner without revoke: %v", err)
	}
	requireUnchanged(t, "refused reassignment", before, f.dispatch("d-r"))
	if _, err := f.s.ReassignDispatch(ctx, "d-r", 2, second, true); !errors.Is(err, ErrDispatchFenced) {
		t.Fatalf("reassignment with a stale attempt: %v", err)
	}

	// The first worker's late result message is already in flight.
	late, err := f.sendResult(first, "d-r", 1, "late result")
	if err != nil {
		t.Fatal(err)
	}
	f.live[first.ProcessStart] = false
	d, err := f.s.ReassignDispatch(ctx, "d-r", 1, second, false)
	if err != nil {
		t.Fatalf("reassignment after the owner is confirmed gone: %v", err)
	}
	if d.Attempt != 2 || d.LaneSlot != second.Slot || d.AssigneeGeneration != second.Generation || d.State != DispatchAssigned {
		t.Fatalf("reassigned record: %+v", d)
	}
	assign2, err := f.s.Send(ctx, SendRequest{From: f.lead, To: second, Kind: KindDispatchNotice, IdempotencyKey: AssignmentKey("d-r", 2), TaskRef: "t1100", CorrelationID: "a-d-r-2", TTL: time.Hour, Payload: []byte("assign d-r attempt 2")})
	if err != nil {
		t.Fatalf("attempt 2 assignment: %v", err)
	}
	f.receiveAll(second, assign2.ID)
	if _, err := f.s.MarkDispatchDelivered(ctx, second, "d-r", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.StartDispatch(ctx, second, "d-r", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := f.sendResult(second, "d-r", 2, "attempt 2 result"); err != nil {
		t.Fatal(err)
	}
	claims, err := f.s.Claim(ctx, f.lead, MaxBatch, time.Minute)
	if err != nil || len(claims) != 2 {
		t.Fatalf("lead inbox: %+v %v", claims, err)
	}
	var lateClaim, secondClaim Claim
	for _, c := range claims {
		if c.ID == late.ID {
			lateClaim = c
		} else {
			secondClaim = c
		}
	}
	requireOutcome(t, "attempt 2", f.applyFromClaim(secondClaim, "d-r", 2), ApplyAccepted)
	recorded := f.dispatch("d-r")
	requireOutcome(t, "late attempt 1", f.applyFromClaim(lateClaim, "d-r", 1), ApplyStale)
	requireUnchanged(t, "late attempt 1", recorded, f.dispatch("d-r"))
	if recorded.ResultAttempt != 2 || recorded.ResultRef != secondClaim.ID || recorded.ResultDigest != ResultDigest([]byte("attempt 2 result")) {
		t.Fatalf("applied result must be attempt 2's only: %+v", recorded)
	}

	// An explicit revoke fences a live owner as well.
	f.assign("d-v", first2(f, first))
	if d, err = f.s.ReassignDispatch(ctx, "d-v", 1, second, true); err != nil || d.Attempt != 2 || d.LaneSlot != second.Slot {
		t.Fatalf("revoked reassignment: %+v %v", d, err)
	}
}

// first2 brings the first lane back as a live owner under a new session so
// the revoke path is exercised against a live previous owner.
func first2(f *dispatchFixture, p Peer) Peer {
	f.t.Helper()
	return f.restart(p, "a-s2", "start-a2")
}

package factorymsg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// bindFixture is a broker with a bound lead and a bound lane-1, both seeded
// through the production launcher path (RegisterLaunchPending then
// BindLaunchPending), and a lane-1 handoff advanced through the production
// store transitions.
type bindFixture struct {
	s            *Store
	root, run    string
	lead, source Peer
	h            Handoff
	ownerStart   string
}

const bindSlot = "lane-1"

func currentOwnerStart(t *testing.T) string {
	t.Helper()
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("process-start identity of the test process is unavailable")
	}
	return start
}

func seedBoundPeer(t *testing.T, s *Store, p Peer, session string) Peer {
	t.Helper()
	ctx := context.Background()
	pending, err := s.RegisterLaunchPending(ctx, p)
	if err != nil {
		t.Fatalf("seed %s: %v", p.Slot, err)
	}
	pending.SessionUUID = session
	bound, ok, err := s.BindLaunchPending(ctx, pending)
	if err != nil || !ok {
		t.Fatalf("seed bind %s ok=%v err=%v", p.Slot, ok, err)
	}
	return bound
}

// newBindSeed opens a broker with the lead and the lane-1 source bound, and
// no handoff yet.
func newBindSeed(t *testing.T) *bindFixture {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	f := &bindFixture{root: filepath.Join(t.TempDir(), "project"), run: "run-bind"}
	var err error
	f.s, err = Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.s.Close() })
	f.ownerStart = currentOwnerStart(t)
	f.lead = seedBoundPeer(t, f.s, Peer{ProjectKey: "project", RunID: f.run, Backend: "claude", Role: "lead", Slot: "lead", PID: os.Getpid(), ProcessStart: f.ownerStart}, "lead-uuid")
	f.source = seedBoundPeer(t, f.s, Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: bindSlot, PID: os.Getpid(), ProcessStart: "fake-source-start"}, "src-uuid")
	return f
}

// reserveToPending drives the lane-1 handoff to its mode's SWITCH_PENDING
// state; headless also records official relocation evidence.
func (f *bindFixture) reserveToPending(t *testing.T, mode string) {
	t.Helper()
	ctx := context.Background()
	r := validReservation(f.root)
	r.Mode = mode
	h, err := f.s.ReserveHandoff(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if h, err = f.s.MarkHandoffWTReady(ctx, h); err != nil {
		t.Fatal(err)
	}
	if mode == HandoffModeInteractive {
		h, err = f.s.MarkHandoffSwitchPendingInteractive(ctx, h)
	} else {
		h, err = f.s.MarkHandoffSwitchPendingHeadless(ctx, h)
	}
	if err != nil {
		t.Fatal(err)
	}
	if mode == HandoffModeHeadless {
		if err := f.s.RecordHeadlessRelocation(ctx, h, validRelocation(h)); err != nil {
			t.Fatal(err)
		}
	}
	f.h = h
}

func newBindFixture(t *testing.T, mode string) *bindFixture {
	t.Helper()
	f := newBindSeed(t)
	f.reserveToPending(t, mode)
	return f
}

// newSession is the endpoint the mode's evidence names: the SessionStart
// session of the user's next normal turn, or the official returned thread id.
func (f *bindFixture) newSession() string {
	if f.h.Mode == HandoffModeHeadless {
		return validRelocation(f.h).ThreadID
	}
	return "post-cd-uuid"
}

// evidence is matching mode evidence for f.h, owned by a current process.
func (f *bindFixture) evidence() HandoffBindEvidence {
	return HandoffBindEvidence{
		Mode: f.h.Mode, Nonce: f.h.Nonce, CardID: f.h.CardID, SpecID: f.h.SpecID,
		SessionUUID: f.newSession(), PID: os.Getpid(), ProcessStart: f.ownerStart,
		Cwd: f.h.TargetPath, WorktreeRoot: f.h.TargetPath, Branch: f.h.TargetBranch, Head: f.h.DevelopPin,
	}
}

func (f *bindFixture) newPeer() Peer {
	p := f.source
	p.SessionUUID = f.newSession()
	p.Generation = f.source.Generation + 1
	p.PID = os.Getpid()
	p.ProcessStart = f.ownerStart
	return p
}

func (f *bindFixture) send(t *testing.T, to Peer, key string) Envelope {
	t.Helper()
	env, err := f.s.Send(context.Background(), SendRequest{
		From: f.lead, To: to, Kind: KindDispatchNotice, IdempotencyKey: key, TaskRef: "t1082",
		CorrelationID: "c-" + key, TTL: time.Hour, Payload: []byte("body-" + key),
	})
	if err != nil {
		t.Fatalf("send %s: %v", key, err)
	}
	return env
}

func (f *bindFixture) handoffState(t *testing.T) (string, string) {
	t.Helper()
	var state, reason string
	if err := f.s.db.QueryRow(`SELECT state,reason FROM lane_handoffs WHERE id=?`, f.h.ID).Scan(&state, &reason); err != nil {
		t.Fatal(err)
	}
	return state, reason
}

// bindCounts is every write the atomic rebind may make, counted.
type bindCounts struct{ Tombstones, Bound, Receipts, Releases, MessageReleases int }

func (f *bindFixture) counts(t *testing.T) bindCounts {
	t.Helper()
	return bindCounts{
		Tombstones:      f.s.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones`),
		Bound:           f.s.countRows(t, `SELECT count(*) FROM lane_handoff_events WHERE to_state='BOUND'`),
		Receipts:        f.s.countRows(t, `SELECT count(*) FROM lane_handoff_receipts`),
		Releases:        f.s.countRows(t, `SELECT count(*) FROM lane_dispatch_releases`),
		MessageReleases: f.s.countRows(t, `SELECT count(*) FROM lane_message_releases`),
	}
}

type messageRow struct {
	Recipient, State, Token, Disposition string
	Generation                           int64
}

func (f *bindFixture) messageRow(t *testing.T, id string) messageRow {
	t.Helper()
	var r messageRow
	if err := f.s.db.QueryRow(`SELECT recipient_session,recipient_generation,state,claim_token,disposition FROM messages WHERE id=?`, id).
		Scan(&r.Recipient, &r.Generation, &r.State, &r.Token, &r.Disposition); err != nil {
		t.Fatal(err)
	}
	return r
}

func requireStale(t *testing.T, err error, wantCode string, current Peer) {
	t.Helper()
	stale, ok := StaleEndpoint(err)
	if !ok {
		t.Fatalf("err=%v, want a stale-endpoint NACK %s", err, wantCode)
	}
	if stale.Code != wantCode {
		t.Fatalf("stale code = %s, want %s (err=%v)", stale.Code, wantCode, err)
	}
	if stale.Current.Slot != current.Slot || stale.Current.SessionUUID != current.SessionUUID || stale.Current.Generation != current.Generation {
		t.Fatalf("redirect = %+v, want slot=%s session=%s generation=%d", stale.Current, current.Slot, current.SessionUUID, current.Generation)
	}
}

// TestFactoryLaneHandoffAtomicModeEvidenceRebind is AC-FLH-005 (REQ-FLH-008):
// peer replacement, old tombstone, BOUND state, receipt, and release marker are
// all visible or all absent, for interactive and headless evidence alike.
func TestFactoryLaneHandoffAtomicModeEvidenceRebind(t *testing.T) {
	steps := []string{"locked", "tombstone", "peer", "bound", "receipt", "release"}
	for _, mode := range []string{HandoffModeInteractive, HandoffModeHeadless} {
		t.Run(mode+"/failpoint_at_every_write_boundary", func(t *testing.T) {
			f := newBindFixture(t, mode)
			f.send(t, f.source, "pre-bind")
			ctx := context.Background()
			before := readEndpointRow(t, f.s.db, bindSlot)
			for _, step := range steps {
				injected := errors.New("injected at " + step)
				f.s.bindStep = func(s string) error {
					if s == step {
						return injected
					}
					return nil
				}
				_, err := f.s.BindHandoff(ctx, f.h, f.evidence())
				if !errors.Is(err, injected) {
					t.Fatalf("%s: err=%v, want injected failure", step, err)
				}
				if got := f.counts(t); got != (bindCounts{}) {
					t.Fatalf("%s: partial rebind visible: %+v", step, got)
				}
				if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
					t.Fatalf("%s: endpoint moved without BOUND: before=%+v after=%+v", step, before, after)
				}
				if st, _ := f.handoffState(t); st != f.h.State {
					t.Fatalf("%s: handoff state = %s, want %s", step, st, f.h.State)
				}
			}
			var seen []string
			f.s.bindStep = func(s string) error { seen = append(seen, s); return nil }
			b, err := f.s.BindHandoff(ctx, f.h, f.evidence())
			if err != nil {
				t.Fatalf("rebind: %v", err)
			}
			if strings.Join(seen, ",") != strings.Join(steps, ",") {
				t.Fatalf("write boundaries = %v, want %v", seen, steps)
			}
			if got := f.counts(t); got != (bindCounts{Tombstones: 1, Bound: 1, Receipts: 1, Releases: 1, MessageReleases: 1}) {
				t.Fatalf("all-visible counts = %+v", got)
			}
			now := readEndpointRow(t, f.s.db, bindSlot)
			if now.Session != f.newSession() || now.Generation != f.source.Generation+1 || now.PID != os.Getpid() || now.ProcessStart != f.ownerStart {
				t.Fatalf("endpoint after rebind = %+v", now)
			}
			if st, _ := f.handoffState(t); st != HandoffBound {
				t.Fatalf("handoff state = %s, want BOUND", st)
			}
			var ts struct {
				session, by string
				gen, byGen  int64
			}
			if err := f.s.db.QueryRow(`SELECT session_uuid,generation,replaced_by_session,replaced_by_generation FROM lane_endpoint_tombstones WHERE handoff_id=?`, f.h.ID).
				Scan(&ts.session, &ts.gen, &ts.by, &ts.byGen); err != nil {
				t.Fatal(err)
			}
			if ts.session != "src-uuid" || ts.gen != f.source.Generation || ts.by != f.newSession() || ts.byGen != f.source.Generation+1 {
				t.Fatalf("tombstone = %+v", ts)
			}
			if b.ReceiptID == "" || b.HandoffID != f.h.ID || b.New.SessionUUID != f.newSession() || b.New.Generation != f.source.Generation+1 || b.Old.SessionUUID != "src-uuid" || b.Released != 1 {
				t.Fatalf("binding = %+v", b)
			}

			// A lost commit ACK is retried: the same receipt is redelivered and
			// nothing is written twice.
			again, err := f.s.BindHandoff(ctx, f.h, f.evidence())
			if err != nil || again.ReceiptID != b.ReceiptID {
				t.Fatalf("idempotent retry = %+v err=%v, want receipt %s", again, err, b.ReceiptID)
			}
			if got := f.counts(t); got != (bindCounts{Tombstones: 1, Bound: 1, Receipts: 1, Releases: 1, MessageReleases: 1}) {
				t.Fatalf("retry wrote again: %+v", got)
			}
		})

		mismatches := map[string]struct {
			mutate func(*bindFixture, *HandoffBindEvidence)
			reason string
		}{
			"cwd_subdirectory":  {func(f *bindFixture, e *HandoffBindEvidence) { e.Cwd = filepath.Join(f.h.TargetPath, "internal") }, NackTargetReadbackMismatch},
			"cwd_lookalike":     {func(f *bindFixture, e *HandoffBindEvidence) { e.Cwd = f.h.TargetPath + "-x"; e.WorktreeRoot = e.Cwd }, NackTargetReadbackMismatch},
			"root_not_target":   {func(f *bindFixture, e *HandoffBindEvidence) { e.WorktreeRoot = f.root }, NackTargetReadbackMismatch},
			"branch_mismatch":   {func(f *bindFixture, e *HandoffBindEvidence) { e.Branch = "WT-other" }, NackTargetReadbackMismatch},
			"head_mismatch":     {func(f *bindFixture, e *HandoffBindEvidence) { e.Head = strings.Repeat("b", 40) }, NackTargetReadbackMismatch},
			"uuid_not_new":      {func(f *bindFixture, e *HandoffBindEvidence) { e.SessionUUID = "src-uuid" }, NackBindingEvidenceInvalid},
			"uuid_in_use":       {func(f *bindFixture, e *HandoffBindEvidence) { e.SessionUUID = "lead-uuid" }, NackBindingEvidenceInvalid},
			"uuid_launch_token": {func(f *bindFixture, e *HandoffBindEvidence) { e.SessionUUID = launchPendingSessionPrefix + "x" }, NackBindingEvidenceInvalid},
			"owner_not_current": {func(f *bindFixture, e *HandoffBindEvidence) { e.ProcessStart = "fake-dead-start" }, NackBindingEvidenceInvalid},
			"owner_pid_invalid": {func(f *bindFixture, e *HandoffBindEvidence) { e.PID = 0 }, NackBindingEvidenceInvalid},
			"mode_crossed":      {func(f *bindFixture, e *HandoffBindEvidence) { e.Mode = otherMode(f.h.Mode) }, NackBindingEvidenceInvalid},
			"card_mismatch":     {func(f *bindFixture, e *HandoffBindEvidence) { e.CardID = "t9999" }, NackBindingEvidenceInvalid},
			"spec_mismatch":     {func(f *bindFixture, e *HandoffBindEvidence) { e.SpecID = "SPEC-OTHER-001" }, NackBindingEvidenceInvalid},
		}
		if mode == HandoffModeHeadless {
			mismatches["thread_not_relocated"] = struct {
				mutate func(*bindFixture, *HandoffBindEvidence)
				reason string
			}{func(f *bindFixture, e *HandoffBindEvidence) { e.SessionUUID = "thr-unrelated" }, NackRelocationEvidenceInvalid}
		}
		for name, tc := range mismatches {
			t.Run(mode+"/mismatch/"+name, func(t *testing.T) {
				f := newBindFixture(t, mode)
				before := readEndpointRow(t, f.s.db, bindSlot)
				ev := f.evidence()
				tc.mutate(f, &ev)
				_, err := f.s.BindHandoff(context.Background(), f.h, ev)
				if got, ok := HandoffNackReason(err); !ok || got != tc.reason {
					t.Fatalf("err=%v, want NACK %s", err, tc.reason)
				}
				if st, reason := f.handoffState(t); st != HandoffNack || reason != tc.reason {
					t.Fatalf("handoff = %s/%s, want NACK/%s", st, reason, tc.reason)
				}
				if got := f.counts(t); got != (bindCounts{}) {
					t.Fatalf("mismatched evidence wrote: %+v", got)
				}
				if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
					t.Fatalf("endpoint moved on mismatched evidence: before=%+v after=%+v", before, after)
				}
			})
		}

		// Evidence for another reservation is not this handoff's evidence: it
		// is refused as stale and writes nothing, not even a NACK.
		t.Run(mode+"/stale_nonce", func(t *testing.T) {
			f := newBindFixture(t, mode)
			ev := f.evidence()
			ev.Nonce = "not-this-reservation"
			_, err := f.s.BindHandoff(context.Background(), f.h, ev)
			requireStale(t, err, NackStaleGeneration, f.source)
			if st, _ := f.handoffState(t); st != f.h.State {
				t.Fatalf("handoff state = %s, want %s untouched", st, f.h.State)
			}
			if got := f.counts(t); got != (bindCounts{}) {
				t.Fatalf("stale nonce wrote: %+v", got)
			}
		})

		// The CAS on the reserved source tuple: once the row differs, the
		// rebind returns STALE_GENERATION and writes nothing.
		t.Run(mode+"/source_tuple_moved", func(t *testing.T) {
			f := newBindFixture(t, mode)
			moved := f.source
			moved.Generation++
			moved.ProcessStart = "fake-relaunch-start"
			if _, err := f.s.db.Exec(`UPDATE peers SET generation=?,process_start=? WHERE slot=?`, moved.Generation, moved.ProcessStart, bindSlot); err != nil {
				t.Fatal(err)
			}
			before := readEndpointRow(t, f.s.db, bindSlot)
			_, err := f.s.BindHandoff(context.Background(), f.h, f.evidence())
			requireStale(t, err, NackStaleGeneration, moved)
			if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
				t.Fatalf("CAS loser wrote the row: before=%+v after=%+v", before, after)
			}
			if got := f.counts(t); got != (bindCounts{}) {
				t.Fatalf("CAS loser wrote: %+v", got)
			}
		})
	}
}

func otherMode(m string) string {
	if m == HandoffModeInteractive {
		return HandoffModeHeadless
	}
	return HandoffModeInteractive
}

// TestFactoryLaneHandoffDispatchAfterBound is AC-FLH-006 (REQ-FLH-009): one
// dispatch sent before relocation and one during SWITCH_PENDING keep their
// metadata, stay unreadable before BOUND, and release exactly once to the new
// generation.
func TestFactoryLaneHandoffDispatchAfterBound(t *testing.T) {
	f := newBindSeed(t)
	ctx := context.Background()

	d1 := f.send(t, f.source, "d1-before-relocation")
	// The old endpoint claimed D1 before any handoff existed.
	pre, err := f.s.Claim(ctx, f.source, 1, time.Hour)
	if err != nil || len(pre) != 1 || pre[0].ID != d1.ID {
		t.Fatalf("pre-handoff claim = %+v err=%v", pre, err)
	}
	f.reserveToPending(t, HandoffModeInteractive)
	d2 := f.send(t, f.source, "d2-during-switch-pending")
	newPeer := f.newPeer()

	// Before BOUND: metadata is durable; body claim/read/ACK and code-write
	// authorization are denied to the old endpoint; the new endpoint is not an
	// endpoint yet.
	for _, id := range []string{d1.ID, d2.ID} {
		if r := f.messageRow(t, id); r.Recipient != "src-uuid" || r.Generation != f.source.Generation {
			t.Fatalf("pre-BOUND metadata of %s = %+v", id, r)
		}
	}
	if st, err := f.s.Status(ctx); err != nil || st.Pending+st.Claimed != 2 {
		t.Fatalf("pre-BOUND status = %+v err=%v", st, err)
	}
	requirePending := func(what string, err error) {
		t.Helper()
		if got, ok := HandoffNackReason(err); !ok || got != NackEndpointHandoffPending {
			t.Fatalf("%s before BOUND: err=%v, want NACK %s", what, err, NackEndpointHandoffPending)
		}
	}
	_, err = f.s.Claim(ctx, f.source, MaxBatch, time.Hour)
	requirePending("claim", err)
	_, err = f.s.ReadBody(ctx, f.source, d1.ID, pre[0].ClaimToken)
	requirePending("read", err)
	requirePending("disposition", f.s.RecordDisposition(ctx, f.source, d1.ID, pre[0].ClaimToken, DispositionAccepted))
	requirePending("ack", f.s.Receipt(ctx, f.source, d1.ID, pre[0].ClaimToken))
	requirePending("code write", f.s.AuthorizeCardWrite(ctx, f.source, f.h.CardID))
	if _, err := f.s.Claim(ctx, newPeer, MaxBatch, time.Hour); err == nil {
		t.Fatal("new endpoint claimed before BOUND")
	}
	if err := f.s.AuthorizeCardWrite(ctx, newPeer, f.h.CardID); err == nil {
		t.Fatal("new endpoint authorized to write before BOUND")
	}

	b, err := f.s.BindHandoff(ctx, f.h, f.evidence())
	if err != nil {
		t.Fatal(err)
	}
	if b.Released != 2 {
		t.Fatalf("released = %d, want 2", b.Released)
	}

	// After BOUND: the old endpoint is stale; the new generation receives both
	// bodies exactly once.
	_, err = f.s.Claim(ctx, f.source, MaxBatch, time.Hour)
	requireStale(t, err, NackStaleEndpoint, newPeer)
	if err := f.s.AuthorizeCardWrite(ctx, newPeer, f.h.CardID); err != nil {
		t.Fatalf("bound endpoint not authorized: %v", err)
	}
	claims, err := f.s.Claim(ctx, newPeer, MaxBatch, time.Hour)
	if err != nil || len(claims) != 2 {
		t.Fatalf("post-BOUND claim = %+v err=%v, want 2", claims, err)
	}
	executed := map[string]int{}
	for _, c := range claims {
		if c.RecipientSession != newPeer.SessionUUID || c.RecipientGeneration != newPeer.Generation {
			t.Fatalf("released to %s/%d, want %s/%d", c.RecipientSession, c.RecipientGeneration, newPeer.SessionUUID, newPeer.Generation)
		}
		body, err := f.s.ReadBody(ctx, newPeer, c.ID, c.ClaimToken)
		if err != nil {
			t.Fatal(err)
		}
		executed[string(body)]++
		if err := f.s.RecordDisposition(ctx, newPeer, c.ID, c.ClaimToken, DispositionAccepted); err != nil {
			t.Fatal(err)
		}
		if err := f.s.Receipt(ctx, newPeer, c.ID, c.ClaimToken); err != nil {
			t.Fatal(err)
		}
	}
	if executed["body-d1-before-relocation"] != 1 || executed["body-d2-during-switch-pending"] != 1 || len(executed) != 2 {
		t.Fatalf("bodies executed = %v, want each once", executed)
	}
	// Exactly once: a retried rebind releases nothing again, and nothing is
	// claimable afterwards.
	if again, err := f.s.BindHandoff(ctx, f.h, f.evidence()); err != nil || again.ReceiptID != b.ReceiptID {
		t.Fatalf("retry = %+v err=%v", again, err)
	}
	if more, err := f.s.Claim(ctx, newPeer, MaxBatch, time.Hour); err != nil || len(more) != 0 {
		t.Fatalf("second release: claims=%+v err=%v", more, err)
	}
	if n := f.s.countRows(t, `SELECT count(*) FROM messages WHERE state='acknowledged'`); n != 2 {
		t.Fatalf("acknowledged = %d, want 2", n)
	}
}

// TestFactoryLaneHandoffStaleEndpointRejected is AC-FLH-007 (REQ-FLH-001/010).
func TestFactoryLaneHandoffStaleEndpointRejected(t *testing.T) {
	f := newBindSeed(t)
	ctx := context.Background()
	d1 := f.send(t, f.source, "d1")
	pre, err := f.s.Claim(ctx, f.source, 1, time.Hour)
	if err != nil || len(pre) != 1 {
		t.Fatalf("pre-handoff claim = %+v err=%v", pre, err)
	}
	f.reserveToPending(t, HandoffModeInteractive)
	if _, err := f.s.BindHandoff(ctx, f.h, f.evidence()); err != nil {
		t.Fatal(err)
	}
	current := f.newPeer()
	oldGen := current
	oldGen.Generation = f.source.Generation

	type snapshot struct {
		row      endpointRow
		messages string
	}
	snap := func() snapshot {
		var dump []string
		rows, err := f.s.db.Query(`SELECT id,recipient_session,recipient_generation,state,claim_token,disposition FROM messages ORDER BY id`)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id, rs, st, tok, d string
			var g int64
			if err := rows.Scan(&id, &rs, &g, &st, &tok, &d); err != nil {
				t.Fatal(err)
			}
			dump = append(dump, strings.Join([]string{id, rs, st, tok, d}, "|"))
		}
		return snapshot{row: readEndpointRow(t, f.s.db, bindSlot), messages: strings.Join(dump, ";")}
	}
	sendFrom := func(from, to Peer, key string) error {
		_, err := f.s.Send(ctx, SendRequest{From: from, To: to, Kind: KindStatusReport, IdempotencyKey: key, TaskRef: "t1082", CorrelationID: "c-" + key, TTL: time.Hour, Payload: []byte("x")})
		return err
	}

	stale := []struct {
		name, code string
		op         func() error
	}{
		{"old_endpoint_send", NackStaleEndpoint, func() error { return sendFrom(f.source, f.lead, "old-send") }},
		{"send_to_old_endpoint", NackStaleEndpoint, func() error { return sendFrom(f.lead, f.source, "to-old") }},
		{"old_endpoint_claim", NackStaleEndpoint, func() error { _, err := f.s.Claim(ctx, f.source, 1, time.Hour); return err }},
		{"old_endpoint_read", NackStaleEndpoint, func() error { _, err := f.s.ReadBody(ctx, f.source, d1.ID, pre[0].ClaimToken); return err }},
		{"old_endpoint_disposition", NackStaleEndpoint, func() error {
			return f.s.RecordDisposition(ctx, f.source, d1.ID, pre[0].ClaimToken, DispositionAccepted)
		}},
		{"old_endpoint_ack", NackStaleEndpoint, func() error { return f.s.Receipt(ctx, f.source, d1.ID, pre[0].ClaimToken) }},
		{"stale_generation_claim", NackStaleGeneration, func() error { _, err := f.s.Claim(ctx, oldGen, 1, time.Hour); return err }},
		{"pre_handoff_claim_token_read", NackStaleGeneration, func() error { _, err := f.s.ReadBody(ctx, current, d1.ID, pre[0].ClaimToken); return err }},
		{"pre_handoff_claim_token_ack", NackStaleGeneration, func() error { return f.s.Receipt(ctx, current, d1.ID, pre[0].ClaimToken) }},
		{"stale_reservation", NackStaleGeneration, func() error {
			ev := f.evidence()
			ev.Nonce = "old-reservation-nonce"
			_, err := f.s.BindHandoff(ctx, f.h, ev)
			return err
		}},
		{"bound_reservation_replayed_with_other_endpoint", NackStaleGeneration, func() error {
			ev := f.evidence()
			ev.SessionUUID = "another-uuid"
			_, err := f.s.BindHandoff(ctx, f.h, ev)
			return err
		}},
	}
	for _, tc := range stale {
		t.Run(tc.name, func(t *testing.T) {
			before := snap()
			err := tc.op()
			requireStale(t, err, tc.code, current)
			if strings.Contains(err.Error(), "body-") || strings.Contains(err.Error(), pre[0].ClaimToken) {
				t.Fatalf("redirect leaks a body or a claim token: %v", err)
			}
			if after := snap(); after != before {
				t.Fatalf("stale operation mutated state:\nbefore=%+v\nafter=%+v", before, after)
			}
		})
	}

	// The tombstone survives a restart of the broker handle.
	if err := f.s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	f.s = reopened
	t.Cleanup(func() { _ = reopened.Close() })
	requireStale(t, sendFrom(f.source, f.lead, "after-restart"), NackStaleEndpoint, current)
	if n := reopened.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid='src-uuid'`); n != 1 {
		t.Fatalf("tombstones after restart = %d, want 1", n)
	}

	// Current operations succeed.
	claims, err := reopened.Claim(ctx, current, MaxBatch, time.Hour)
	if err != nil || len(claims) != 1 || claims[0].ID != d1.ID {
		t.Fatalf("current claim = %+v err=%v", claims, err)
	}
	if _, err := reopened.ReadBody(ctx, current, d1.ID, claims[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if err := reopened.RecordDisposition(ctx, current, d1.ID, claims[0].ClaimToken, DispositionAccepted); err != nil {
		t.Fatal(err)
	}
	if err := reopened.Receipt(ctx, current, d1.ID, claims[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if err := sendFrom(f.lead, current, "to-current"); err != nil {
		t.Fatalf("send to current endpoint: %v", err)
	}
	if err := sendFrom(current, f.lead, "from-current"); err != nil {
		t.Fatalf("send from current endpoint: %v", err)
	}

	// Tombstoned session UUID through the t1074 UserPromptSubmit path
	// (RegisterPeer, not launch-pending): a resume with the recorded owner, a
	// restart with a new owner, and the same after the handle is reopened.
	durable := func(s *Store) string {
		return dumpRows(t, s, `SELECT slot,session_uuid,generation,replaced_by_session,replaced_by_generation,handoff_id,bound_at FROM lane_endpoint_tombstones ORDER BY session_uuid,generation`) +
			"#" + dumpRows(t, s, `SELECT id,handoff_id,old_session,old_generation,session_uuid,generation,pid,process_start,created_at FROM lane_handoff_receipts ORDER BY id`)
	}
	turnRegistration := func(s *Store, name string, pid int, start string) {
		t.Helper()
		row, rows := readEndpointRow(t, s.db, bindSlot), durable(s)
		_, err := s.RegisterPeer(ctx, Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: bindSlot, SessionUUID: "src-uuid", Generation: 1, PID: pid, ProcessStart: start})
		requireStale(t, err, NackStaleEndpoint, current)
		if after := readEndpointRow(t, s.db, bindSlot); after != row {
			t.Fatalf("%s: endpoint row changed: before=%+v after=%+v", name, row, after)
		}
		if after := durable(s); after != rows {
			t.Fatalf("%s: tombstone or receipt rows changed", name)
		}
	}
	turnRegistration(reopened, "resume_recorded_owner", f.source.PID, f.source.ProcessStart)
	turnRegistration(reopened, "restart_new_owner", 999_983, "restart-start")
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
	again, err := Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = again.Close() })
	turnRegistration(again, "after_reopen", f.source.PID, f.source.ProcessStart)

	// BOUND row, before any launcher registration: a SessionStart bind of the
	// tombstoned UUID is refused, not silently passed over.
	row, rows := readEndpointRow(t, again.db, bindSlot), durable(again)
	_, ok, err := again.BindLaunchPending(ctx, Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: bindSlot, SessionUUID: "src-uuid", Generation: 1, PID: current.PID, ProcessStart: current.ProcessStart})
	if ok {
		t.Fatal("BOUND row: tombstoned session bound")
	}
	requireStale(t, err, NackStaleEndpoint, current)
	if after := readEndpointRow(t, again.db, bindSlot); after != row {
		t.Fatalf("BOUND row: endpoint row changed: before=%+v after=%+v", row, after)
	}
	if after := durable(again); after != rows {
		t.Fatal("BOUND row: tombstone or receipt rows changed")
	}

	t.Run("launcher_resume_leg", func(t *testing.T) { staleEndpointLauncherResumeLeg(t) })
}

// dumpRows renders every row of query as one comparable string.
func dumpRows(t *testing.T, s *Store, query string) string {
	t.Helper()
	rows, err := s.db.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		out = append(out, fmt.Sprint(vals...))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(out, ";")
}

// staleEndpointLauncherResumeLeg is the AC-FLH-007 launcher resume and
// positive legs (REQ-FLH-010, lead decision D1=(b)). The handoff-bound owner
// is a live helper process so the test can make it not current.
func staleEndpointLauncherResumeLeg(t *testing.T) {
	f := newBindSeed(t)
	ctx := context.Background()
	f.reserveToPending(t, HandoffModeInteractive)
	boundOwner := startStoppableOwner(t)
	ev := f.evidence()
	ev.PID, ev.ProcessStart = boundOwner.PID, boundOwner.Start
	b, err := f.s.BindHandoff(ctx, f.h, ev)
	if err != nil {
		t.Fatal(err)
	}
	tombGen := f.source.Generation
	const bindInputGeneration = 1 // the production SessionStart hook's bind input
	if bindInputGeneration == tombGen {
		t.Fatalf("bind input generation %d equals the tombstoned generation; the leg would not discriminate a pair match", tombGen)
	}
	boundOwner.Stop(t)

	durable := func() string {
		return dumpRows(t, f.s, `SELECT slot,session_uuid,generation,replaced_by_session,replaced_by_generation,handoff_id,bound_at FROM lane_endpoint_tombstones ORDER BY session_uuid,generation`) +
			"#" + dumpRows(t, f.s, `SELECT id,handoff_id,old_session,old_generation,session_uuid,generation,pid,process_start,created_at FROM lane_handoff_receipts ORDER BY id`)
	}
	rowsBefore := durable()
	launcher := startStoppableOwner(t)
	lane := Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: bindSlot, PID: launcher.PID, ProcessStart: launcher.Start}
	pending, err := f.s.RegisterLaunchPending(ctx, lane)
	if err != nil {
		t.Fatalf("launcher registration after the bound owner stopped: %v", err)
	}
	if pending.Generation <= b.New.Generation {
		t.Fatalf("launch-pending generation %d not above the handoff-bound generation %d", pending.Generation, b.New.Generation)
	}
	committed := dumpRows(t, f.s, `SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot='`+bindSlot+`'`)
	redirect := Peer{Slot: bindSlot, SessionUUID: "", Generation: pending.Generation}

	resume := lane
	resume.SessionUUID, resume.Generation = "src-uuid", bindInputGeneration
	_, ok, err := f.s.BindLaunchPending(ctx, resume)
	if ok {
		t.Fatal("launcher resume: tombstoned session bound to the launch-pending row")
	}
	requireStale(t, err, NackStaleEndpoint, redirect)
	if got := dumpRows(t, f.s, `SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot='`+bindSlot+`'`); got != committed {
		t.Fatalf("launch-pending row changed by the refused bind:\nbefore=%s\nafter=%s", committed, got)
	}
	if got := durable(); got != rowsBefore {
		t.Fatal("launcher resume: tombstone or receipt rows changed")
	}

	turn := resume
	if _, err := f.s.RegisterPeer(ctx, turn); err == nil {
		t.Fatal("turn registration of the tombstoned session accepted")
	} else {
		requireStale(t, err, NackStaleEndpoint, redirect)
	}
	if got := dumpRows(t, f.s, `SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot='`+bindSlot+`'`); got != committed {
		t.Fatal("launch-pending row changed by the refused turn registration")
	}

	revived := resume
	revived.Generation = pending.Generation + 1
	send := func(from, to Peer, key string) error {
		_, err := f.s.Send(ctx, SendRequest{From: from, To: to, Kind: KindStatusReport, IdempotencyKey: key, TaskRef: "t1082", CorrelationID: "c-" + key, TTL: time.Hour, Payload: []byte("x")})
		return err
	}
	if err := send(revived, f.lead, "from-revived"); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("send from the refused bind's identity: err=%v, want ErrStalePeer class", err)
	}
	if err := send(f.lead, revived, "to-revived"); !errors.Is(err, ErrStalePeer) {
		t.Fatalf("send to the refused bind's identity: err=%v, want ErrStalePeer class", err)
	}

	// Positive leg: once the launch-pending owner is not current, a fresh
	// launcher registration and an untombstoned bind succeed.
	launcher.Stop(t)
	fresh := lane
	fresh.PID, fresh.ProcessStart = os.Getpid(), f.ownerStart
	pending2, err := f.s.RegisterLaunchPending(ctx, fresh)
	if err != nil {
		t.Fatalf("fresh launcher registration: %v", err)
	}
	if pending2.Generation <= pending.Generation {
		t.Fatalf("fresh launch-pending generation %d not above %d", pending2.Generation, pending.Generation)
	}
	fresh.SessionUUID, fresh.Generation = "fresh-uuid", bindInputGeneration
	bound, ok, err := f.s.BindLaunchPending(ctx, fresh)
	if err != nil || !ok {
		t.Fatalf("untombstoned bind ok=%v err=%v", ok, err)
	}
	if bound.Generation <= pending2.Generation {
		t.Fatalf("bound generation %d not above %d", bound.Generation, pending2.Generation)
	}
	if row := readEndpointRow(t, f.s.db, bindSlot); row.Session != "fresh-uuid" || row.Generation != bound.Generation {
		t.Fatalf("current endpoint = %+v, want fresh-uuid at %d", row, bound.Generation)
	}
}

// TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch is AC-FLH-008
// (REQ-FLH-009). No assertion relies on which fields form the idempotency
// basis: every check below holds whether the key is scoped by sender session
// or by sender lane slot (t1100).
func TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch(t *testing.T) {
	f := newBindSeed(t)
	ctx := context.Background()
	k1Req := SendRequest{From: f.lead, To: f.source, Kind: KindDispatchNotice, IdempotencyKey: "K1", TaskRef: "t1082", CorrelationID: "c-k1", TTL: time.Hour, Payload: []byte("k1-body")}

	// Sender layer, same recipient generation: the identical repeat, by the
	// same sender with no restart in between, returns the existing envelope.
	k1, err := f.s.Send(ctx, k1Req)
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := f.s.Send(ctx, k1Req)
	if err != nil || repeat.ID != k1.ID {
		t.Fatalf("identical K1 repeat = %+v err=%v, want envelope %s", repeat, err, k1.ID)
	}
	if n := f.s.countRows(t, `SELECT count(*) FROM messages WHERE idem_key='K1'`); n != 1 {
		t.Fatalf("K1 rows = %d, want 1", n)
	}

	f.reserveToPending(t, HandoffModeInteractive)
	if _, err := f.s.BindHandoff(ctx, f.h, f.evidence()); err != nil {
		t.Fatal(err)
	}
	current := f.newPeer()

	// K1 reused after BOUND with a different recipient session/generation is
	// not merged into the K1 envelope. Rejected or stored separately — the
	// basis decides, and the test asserts neither.
	k1Before := f.messageRow(t, k1.ID)
	reuse := k1Req
	reuse.To = current
	if env, err := f.s.Send(ctx, reuse); err == nil && env.ID == k1.ID {
		t.Fatalf("K1 with a different recipient was merged into envelope %s", k1.ID)
	}
	if after := f.messageRow(t, k1.ID); after != k1Before {
		t.Fatalf("K1 reuse changed the K1 envelope: before=%+v after=%+v", k1Before, after)
	}

	// The resend after BOUND uses a key the sender has not used before.
	k2Req := k1Req
	k2Req.To, k2Req.IdempotencyKey, k2Req.CorrelationID = current, "K2", "c-k2"
	k2, err := f.s.Send(ctx, k2Req)
	if err != nil || k2.ID == k1.ID || k2.RecipientSession != current.SessionUUID || k2.RecipientGeneration != current.Generation {
		t.Fatalf("K2 resend = %+v err=%v", k2, err)
	}

	// Stale-generation layer: a retry addressed to the old generation NACKs.
	_, err = f.s.Send(ctx, k1Req)
	requireStale(t, err, NackStaleEndpoint, current)

	// Recipient layer: the K1 body executes once; a repeated delivery of the
	// same envelope in the same generation is acknowledged as a duplicate.
	executed := map[string]int{}
	var receipts int
	deliver := func(c Claim) {
		t.Helper()
		if executed[c.ID] > 0 {
			if err := f.s.RecordDisposition(ctx, current, c.ID, c.ClaimToken, DispositionDuplicate); err != nil {
				t.Fatal(err)
			}
		} else {
			if _, err := f.s.ReadBody(ctx, current, c.ID, c.ClaimToken); err != nil {
				t.Fatal(err)
			}
			executed[c.ID]++
			if err := f.s.RecordDisposition(ctx, current, c.ID, c.ClaimToken, DispositionAccepted); err != nil {
				t.Fatal(err)
			}
		}
	}
	first, err := f.s.Claim(ctx, current, MaxBatch, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	var k1First Claim
	for _, c := range first {
		if c.ID == k1.ID {
			k1First = c
		}
	}
	if k1First.ID == "" {
		t.Fatalf("K1 not delivered to the rebound endpoint: %+v", first)
	}
	deliver(k1First) // executes; the receipt is lost with the lease
	time.Sleep(5 * time.Millisecond)
	second, err := f.s.Claim(ctx, current, MaxBatch, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var k1Again Claim
	for _, c := range second {
		if c.ID == k1.ID {
			k1Again = c
		}
	}
	if k1Again.ID == "" || k1Again.ClaimToken == k1First.ClaimToken || k1Again.RecipientGeneration != current.Generation {
		t.Fatalf("repeated K1 delivery = %+v, want same envelope, new token, same generation", k1Again)
	}
	deliver(k1Again)
	if err := f.s.Receipt(ctx, current, k1.ID, k1Again.ClaimToken); err != nil {
		t.Fatal(err)
	}
	receipts++
	if err := f.s.Receipt(ctx, current, k1.ID, k1First.ClaimToken); err == nil {
		t.Fatal("receipt with the superseded delivery token accepted")
	}
	if executed[k1.ID] != 1 || receipts != 1 {
		t.Fatalf("K1 executed=%d receipts=%d, want 1/1", executed[k1.ID], receipts)
	}
	if r := f.messageRow(t, k1.ID); r.State != "acknowledged" || r.Disposition != DispositionDuplicate {
		t.Fatalf("K1 after repeated delivery = %+v, want acknowledged/duplicate", r)
	}
}

// TestLaneHandoffBindRefusalsAndRedirects pins the refusal shapes the rebind
// and the stale classifier produce outside the named AC paths: a finalized
// handoff, a vanished or provisional lane row (whose private token is never
// echoed), code-write refusal without a BOUND handoff, and the unchanged t1074
// errors for plain mismatches.
func TestLaneHandoffBindRefusalsAndRedirects(t *testing.T) {
	ctx := context.Background()

	t.Run("finalized_handoff_is_stale", func(t *testing.T) {
		f := newBindFixture(t, HandoffModeInteractive)
		if _, err := f.s.NackHandoff(ctx, f.h, NackStaleGeneration); err != nil {
			t.Fatal(err)
		}
		before := readEndpointRow(t, f.s.db, bindSlot)
		_, err := f.s.BindHandoff(ctx, f.h, f.evidence())
		requireStale(t, err, NackStaleGeneration, f.source)
		if after := readEndpointRow(t, f.s.db, bindSlot); after != before {
			t.Fatalf("rebind after finalization wrote the row: %+v -> %+v", before, after)
		}
		if got := f.counts(t); got != (bindCounts{}) {
			t.Fatalf("rebind after finalization wrote: %+v", got)
		}
	})

	t.Run("vanished_lane_row_is_stale", func(t *testing.T) {
		f := newBindFixture(t, HandoffModeInteractive)
		if _, err := f.s.db.Exec(`DELETE FROM peers WHERE slot=?`, bindSlot); err != nil {
			t.Fatal(err)
		}
		_, err := f.s.BindHandoff(ctx, f.h, f.evidence())
		stale, ok := StaleEndpoint(err)
		if !ok || stale.Code != NackStaleGeneration || stale.Current != (EndpointRef{Slot: bindSlot}) {
			t.Fatalf("err=%v, want STALE_GENERATION with an empty redirect", err)
		}
	})

	t.Run("provisional_row_token_is_not_echoed", func(t *testing.T) {
		f := newBindFixture(t, HandoffModeInteractive)
		token := launchPendingSessionPrefix + "private"
		if _, err := f.s.db.Exec(`UPDATE peers SET session_uuid=?,generation=generation+1 WHERE slot=?`, token, bindSlot); err != nil {
			t.Fatal(err)
		}
		_, err := f.s.BindHandoff(ctx, f.h, f.evidence())
		stale, ok := StaleEndpoint(err)
		if !ok || stale.Current.SessionUUID != "" || stale.Current.Generation != f.source.Generation+1 {
			t.Fatalf("err=%v, want a redirect with the provisional token redacted", err)
		}
		if strings.Contains(err.Error(), "private") {
			t.Fatalf("redirect leaks the provisional token: %v", err)
		}
	})

	t.Run("code_write_needs_a_bound_handoff_for_the_card", func(t *testing.T) {
		f := newBindSeed(t)
		for _, card := range []string{"t1082", "t9999"} {
			if got, ok := HandoffNackReason(f.s.AuthorizeCardWrite(ctx, f.source, card)); !ok || got != NackEndpointHandoffPending {
				t.Fatalf("card %s without handoff: reason=%s ok=%v", card, got, ok)
			}
		}
		f.reserveToPending(t, HandoffModeInteractive)
		if _, err := f.s.NackHandoff(ctx, f.h, NackTargetDirty); err != nil {
			t.Fatal(err)
		}
		if got, ok := HandoffNackReason(f.s.AuthorizeCardWrite(ctx, f.source, "t1082")); !ok || got != NackEndpointHandoffPending {
			t.Fatalf("NACKed handoff authorized code writes: reason=%s ok=%v", got, ok)
		}
	})

	t.Run("plain_mismatches_keep_t1074_errors", func(t *testing.T) {
		f := newBindSeed(t)
		env := f.send(t, f.source, "plain")
		claims, err := f.s.Claim(ctx, f.source, 1, time.Hour)
		if err != nil || len(claims) != 1 {
			t.Fatalf("claim = %+v err=%v", claims, err)
		}
		if _, err := f.s.ReadBody(ctx, f.source, env.ID, "wrong-token"); err == nil || err.Error() != "claim identity mismatch" {
			t.Fatalf("read with a wrong token: %v", err)
		}
		if err := f.s.RecordDisposition(ctx, f.source, env.ID, "wrong-token", DispositionAccepted); err == nil || err.Error() != "claim identity mismatch" {
			t.Fatalf("disposition with a wrong token: %v", err)
		}
		if err := f.s.Receipt(ctx, f.source, env.ID, "wrong-token"); err == nil || err.Error() != "receipt requires matching claim and persisted disposition" {
			t.Fatalf("receipt with a wrong token: %v", err)
		}
		stranger := f.source
		stranger.SessionUUID = "never-registered"
		_, strangerErr := f.s.Claim(ctx, stranger, 1, time.Hour)
		if strangerErr == nil || strangerErr.Error() != "stale or unregistered peer" {
			t.Fatalf("unregistered peer: %v", strangerErr)
		}
		if _, ok := StaleEndpoint(strangerErr); ok {
			t.Fatal("unregistered peer classified as stale")
		}
	})
}

// A broker created before the handoff tables existed is opened on the hook hot
// path without schema setup. The pre-BOUND gate must read "no handoff" there,
// not fail every claim.
func TestClaimOnBrokerWithoutHandoffTables(t *testing.T) {
	f := newBindSeed(t)
	f.send(t, f.source, "legacy-broker")
	for _, table := range []string{"lane_handoffs", "lane_endpoint_tombstones", "lane_message_releases"} {
		if _, err := f.s.db.Exec("DROP TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}
	legacy, err := OpenExistingWithDeadline(f.root, f.run, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = legacy.Close() })
	claims, err := legacy.Claim(context.Background(), f.source, MaxBatch, time.Hour)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim on a pre-handoff broker = %+v err=%v", claims, err)
	}
}

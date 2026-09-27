package factorymsg

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// endpointRow is the byte-level snapshot of one peers row that the race
// criteria compare before and after a racer returns.
type endpointRow struct {
	Session, ProcessStart, UpdatedAt string
	Generation                       int64
	PID                              int
}

func readEndpointRow(t *testing.T, db *sql.DB, slot string) endpointRow {
	t.Helper()
	var r endpointRow
	if err := db.QueryRow(`SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot=?`, slot).
		Scan(&r.Session, &r.Generation, &r.PID, &r.ProcessStart, &r.UpdatedAt); err != nil {
		t.Fatalf("read endpoint row %q: %v", slot, err)
	}
	return r
}

func countLaneHandoffs(t *testing.T, db *sql.DB, slot string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM lane_handoffs WHERE slot=?`, slot).Scan(&n); err != nil {
		t.Fatalf("count lane handoffs: %v", err)
	}
	return n
}

// requireDistinctRacerHandles is the separate-handle guard of AC-FLH-018/019:
// two racers sharing one Store (and so one SetMaxOpenConns(1) pool) serialize
// in the Go pool and never reach the SQLite BEGIN IMMEDIATE boundary.
func requireDistinctRacerHandles(a, b *Store) error {
	if a == nil || b == nil {
		return errors.New("racer handle is nil")
	}
	if a == b {
		return errors.New("racers share one *Store")
	}
	if a.db == b.db {
		return errors.New("racers share one *sql.DB")
	}
	return nil
}

// proveDistinctHandles runs parts (1) and (3) of the separate-handle proof:
// the guard must first go red on a deliberately shared pair, then pass on the
// real racers. Part (2) is observeBlocked on the waiting racer's own handle.
func proveDistinctHandles(t *testing.T, a, b *Store) {
	t.Helper()
	if err := requireDistinctRacerHandles(a, a); err == nil {
		t.Fatal("distinct-handle guard accepted a shared pair")
	}
	if err := requireDistinctRacerHandles(a, b); err != nil {
		t.Fatalf("racer handles not distinct: %v", err)
	}
}

// stepGate holds a racer inside its write transaction at one named step until
// the test releases it.
type stepGate struct {
	name             string
	reached, release chan struct{}
	once             sync.Once
}

func newStepGate(name string) *stepGate {
	return &stepGate{name: name, reached: make(chan struct{}), release: make(chan struct{})}
}

func (g *stepGate) hook(step string) error {
	if step == g.name {
		g.once.Do(func() { close(g.reached) })
		<-g.release
	}
	return nil
}

func (g *stepGate) ctx() context.Context { return WithStepHook(context.Background(), g.hook) }

func (g *stepGate) await(t *testing.T, who string) {
	t.Helper()
	select {
	case <-g.reached:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s never reached step %q", who, g.name)
	}
}

// racerResult is one racer's return, stamped when it returned.
type racerResult struct {
	peer Peer
	h    Handoff
	b    HandoffBinding
	err  error
	at   time.Time
}

// observeBlocked is part (2) of the separate-handle proof: the waiting racer's
// own handle holds its own connection (InUse==1) without waiting in the pool
// (WaitCount==0), and its call has not returned — it waits on the SQLite lock.
func observeBlocked(t *testing.T, waiter *Store, done <-chan racerResult) bool {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		st := waiter.db.Stats()
		if st.InUse == 1 && st.WaitCount == 0 {
			time.Sleep(150 * time.Millisecond)
			select {
			case r := <-done:
				t.Fatalf("waiting racer returned while the holder kept its transaction: err=%v", r.err)
			default:
			}
			st = waiter.db.Stats()
			return st.InUse == 1 && st.WaitCount == 0
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func receive(t *testing.T, who string, done <-chan racerResult) racerResult {
	t.Helper()
	select {
	case r := <-done:
		return r
	case <-time.After(5 * time.Second):
		t.Fatalf("%s did not return", who)
		return racerResult{}
	}
}

// orderLog records the observed order of racer events.
type orderLog struct {
	mu  sync.Mutex
	seq []string
}

func (l *orderLog) add(ev string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq = append(l.seq, ev)
}

func (l *orderLog) require(t *testing.T, want ...string) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	if strings.Join(l.seq, ",") != strings.Join(want, ",") {
		t.Fatalf("observed order = %v, want the forced order %v", l.seq, want)
	}
}

// raceFixture is one broker with a bound lead and a bound lane-1 source, both
// seeded through the production launcher path, and (optionally) a lane-1
// handoff driven to SWITCH_PENDING_INTERACTIVE through the store transitions.
type raceFixture struct {
	root, run string
	seed      *Store
	lead      Peer
	source    Peer
	h         Handoff
	pid       int
	start     string // live process-start of the test process
}

const raceSlot = "lane-1"

// newRaceFixture seeds the source owned by the live test process (pid/start),
// or, with notCurrentSource, by a fake process-start that is never current.
func newRaceFixture(t *testing.T, notCurrentSource bool) *raceFixture {
	t.Helper()
	f := &raceFixture{root: filepath.Join(t.TempDir(), "project"), run: "run-race", pid: os.Getpid(), start: currentOwnerStart(t)}
	var err error
	if f.seed, err = Open(f.root, f.run); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.seed.Close() })
	f.lead = seedBoundPeer(t, f.seed, Peer{ProjectKey: "project", RunID: f.run, Backend: "claude", Role: "lead", Slot: "lead", PID: f.pid, ProcessStart: f.start}, "lead-uuid")
	sourceStart := f.start
	if notCurrentSource {
		sourceStart = "fake-source-start"
	}
	f.source = seedBoundPeer(t, f.seed, Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: raceSlot, PID: f.pid, ProcessStart: sourceStart}, "src-uuid")
	return f
}

func (f *raceFixture) open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func (f *raceFixture) reservation() HandoffReservation {
	r := validReservation(f.root)
	r.Slot = raceSlot
	r.Mode = HandoffModeInteractive
	return r
}

func (f *raceFixture) toSwitchPending(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	h, err := f.seed.ReserveHandoff(ctx, f.reservation())
	if err != nil {
		t.Fatal(err)
	}
	if h, err = f.seed.MarkHandoffWTReady(ctx, h); err != nil {
		t.Fatal(err)
	}
	if h, err = f.seed.MarkHandoffSwitchPendingInteractive(ctx, h); err != nil {
		t.Fatal(err)
	}
	f.h = h
}

// postCD is racer H's peer: the same owner as the source, a new session UUID —
// the UserPromptSubmit registration of the turn after an interactive /cd.
func (f *raceFixture) postCD(session string) Peer {
	return Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: raceSlot, SessionUUID: session, Generation: 1, PID: f.pid, ProcessStart: f.start}
}

// relaunch is racer A's launcher provisional registration for a relaunched
// process owned by the live test process.
func (f *raceFixture) relaunch() Peer {
	return Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: raceSlot, PID: f.pid, ProcessStart: f.start}
}

// evidence is racer R's matching interactive evidence for session.
func (f *raceFixture) evidence(session string) HandoffBindEvidence {
	return HandoffBindEvidence{
		Mode: HandoffModeInteractive, Nonce: f.h.Nonce, CardID: f.h.CardID, SpecID: f.h.SpecID,
		SessionUUID: session, PID: f.pid, ProcessStart: f.start,
		Cwd: f.h.TargetPath, WorktreeRoot: f.h.TargetPath, Branch: f.h.TargetBranch, Head: f.h.DevelopPin,
	}
}

func (f *raceFixture) handoffState(t *testing.T, id string) (string, string) {
	t.Helper()
	var state, reason string
	if err := f.seed.db.QueryRow(`SELECT state,reason FROM lane_handoffs WHERE id=?`, id).Scan(&state, &reason); err != nil {
		t.Fatal(err)
	}
	return state, reason
}

type raceCounts struct{ Tombstones, Receipts, Releases int }

func (f *raceFixture) counts(t *testing.T) raceCounts {
	t.Helper()
	return raceCounts{
		Tombstones: f.seed.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones`),
		Receipts:   f.seed.countRows(t, `SELECT count(*) FROM lane_handoff_receipts`),
		Releases:   f.seed.countRows(t, `SELECT count(*) FROM lane_dispatch_releases`),
	}
}

func requireNackReason(t *testing.T, err error, want string) {
	t.Helper()
	if reason, ok := HandoffNackReason(err); !ok || reason != want {
		t.Fatalf("err=%v, want NACK %s", err, want)
	}
}

// requireBoundEndState is the shared end state of AC-FLH-019 orders (i)/(ii).
func (f *raceFixture) requireBoundEndState(t *testing.T) {
	t.Helper()
	row := readEndpointRow(t, f.seed.db, raceSlot)
	if row.Session != "post-cd-uuid" || row.Generation != f.source.Generation+1 || row.PID != f.pid || row.ProcessStart != f.start {
		t.Fatalf("endpoint = %+v, want post-cd-uuid at generation %d", row, f.source.Generation+1)
	}
	if st, _ := f.handoffState(t, f.h.ID); st != HandoffBound {
		t.Fatalf("handoff state = %s, want BOUND", st)
	}
	if n := f.seed.countRows(t, `SELECT count(*) FROM lane_endpoint_tombstones WHERE session_uuid='src-uuid' AND generation=?`, f.source.Generation); n != 1 {
		t.Fatalf("tombstones naming src-uuid/g = %d, want 1", n)
	}
	if c := f.counts(t); c.Tombstones != 1 || c.Receipts != 1 {
		t.Fatalf("counts = %+v, want exactly one tombstone and one BOUND receipt", c)
	}
}

// TestFactoryLaneHandoffRebindVsUserPromptRegisterRace is the AC-FLH-019
// named test: forced orders (i)-(vii) and 200 unforced H-versus-R iterations.
func TestFactoryLaneHandoffRebindVsUserPromptRegisterRace(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Log("AC_FLH_019_ORDERS_COVERED=i,ii,iii,iv,v,vi,vii,unforced200")

	t.Run("order_i_registration_holds_first_then_rebind", func(t *testing.T) {
		f := newRaceFixture(t, false)
		f.toSwitchPending(t)
		hStore, rStore := f.open(t), f.open(t)
		proveDistinctHandles(t, hStore, rStore)
		var log orderLog
		before := readEndpointRow(t, f.seed.db, raceSlot)

		hGate := newStepGate(StepRegisterHandoffRead)
		hDone := make(chan racerResult, 1)
		go func() {
			p, err := hStore.RegisterPeer(hGate.ctx(), f.postCD("post-cd-uuid"))
			hDone <- racerResult{peer: p, err: err, at: time.Now()}
		}()
		hGate.await(t, "H")
		log.add("H:held")

		rGate := newStepGate("locked")
		rDone := make(chan racerResult, 1)
		go func() {
			b, err := rStore.BindHandoff(rGate.ctx(), f.h, f.evidence("post-cd-uuid"))
			rDone <- racerResult{b: b, err: err, at: time.Now()}
		}()
		if !observeBlocked(t, rStore, rDone) {
			t.Fatal("R_blocked_observed=false: R was not observed waiting on the SQLite lock")
		}
		t.Log("R_blocked_observed=true")
		log.add("R:blocked")
		close(hGate.release)
		h := receive(t, "H", hDone)
		log.add("H:returned")
		requireNackReason(t, h.err, NackEndpointHandoffPending)

		rGate.await(t, "R")
		log.add("R:locked")
		if between := readEndpointRow(t, f.seed.db, raceSlot); between != before {
			t.Fatalf("row changed between H's return and R's commit: before=%+v between=%+v", before, between)
		}
		close(rGate.release)
		r := receive(t, "R", rDone)
		log.add("R:committed")
		if r.err != nil {
			t.Fatalf("R rebind: %v", r.err)
		}
		log.require(t, "H:held", "R:blocked", "H:returned", "R:locked", "R:committed")
		f.requireBoundEndState(t)
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_ii_rebind_holds_first_then_registrations", func(t *testing.T) {
		f := newRaceFixture(t, false)
		f.toSwitchPending(t)
		hStore, rStore := f.open(t), f.open(t)
		proveDistinctHandles(t, rStore, hStore)
		var log orderLog

		rGate := newStepGate("locked")
		rDone := make(chan racerResult, 1)
		go func() {
			b, err := rStore.BindHandoff(rGate.ctx(), f.h, f.evidence("post-cd-uuid"))
			rDone <- racerResult{b: b, err: err, at: time.Now()}
		}()
		rGate.await(t, "R")
		log.add("R:held")

		hDone := make(chan racerResult, 1)
		go func() {
			p, err := hStore.RegisterPeer(context.Background(), f.postCD("post-cd-uuid"))
			hDone <- racerResult{peer: p, err: err, at: time.Now()}
		}()
		if !observeBlocked(t, hStore, hDone) {
			t.Fatal("H_blocked_observed=false: H was not observed waiting on the SQLite lock")
		}
		t.Log("H_blocked_observed=true")
		log.add("H:blocked")
		close(rGate.release)
		r := receive(t, "R", rDone)
		log.add("R:committed")
		if r.err != nil {
			t.Fatalf("R rebind: %v", r.err)
		}
		h := receive(t, "H", hDone)
		log.add("H:returned")
		log.require(t, "R:held", "H:blocked", "R:committed", "H:returned")
		if !r.at.Before(h.at) {
			t.Fatalf("R commit %s is not earlier than H return %s", r.at, h.at)
		}
		f.requireBoundEndState(t)
		bound := readEndpointRow(t, f.seed.db, raceSlot)
		if h.err != nil {
			t.Fatalf("H post-cd-uuid after BOUND: %v", h.err)
		}
		if after := readEndpointRow(t, f.seed.db, raceSlot); after.Session != bound.Session || after.Generation != bound.Generation || after.PID != bound.PID || after.ProcessStart != bound.ProcessStart {
			t.Fatalf("H post-cd-uuid moved the endpoint: bound=%+v after=%+v", bound, after)
		}

		// One further H call carrying the tombstoned source session.
		beforeStale := readEndpointRow(t, f.seed.db, raceSlot)
		stale := f.postCD("src-uuid")
		_, err := hStore.RegisterPeer(context.Background(), stale)
		requireNackReason(t, err, NackStaleEndpoint)
		if after := readEndpointRow(t, f.seed.db, raceSlot); after != beforeStale {
			t.Fatalf("tombstoned src-uuid changed the row: before=%+v after=%+v", beforeStale, after)
		}
		f.requireBoundEndState(t)
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_iii_rebind_nacks_then_registration_follows_t1074", func(t *testing.T) {
		f := newRaceFixture(t, false)
		msg, err := f.seed.Send(context.Background(), SendRequest{
			From: f.lead, To: f.source, Kind: KindDispatchNotice, IdempotencyKey: "dispatch-iii", TaskRef: "t1082",
			CorrelationID: "c-iii", TTL: time.Hour, Payload: []byte("body-iii"),
		})
		if err != nil {
			t.Fatal(err)
		}
		f.toSwitchPending(t)
		hStore, rStore := f.open(t), f.open(t)
		proveDistinctHandles(t, rStore, hStore)

		ev := f.evidence("post-cd-uuid")
		ev.Branch = "WT-some-other-branch"
		_, err = rStore.BindHandoff(context.Background(), f.h, ev)
		requireNackReason(t, err, NackTargetReadbackMismatch)
		if st, _ := f.handoffState(t, f.h.ID); st != HandoffNack {
			t.Fatalf("handoff state = %s, want NACK", st)
		}
		if c := f.counts(t); c != (raceCounts{}) {
			t.Fatalf("counts after a failed rebind = %+v, want all 0", c)
		}
		if row := readEndpointRow(t, f.seed.db, raceSlot); row.Session != "src-uuid" || row.Generation != f.source.Generation {
			t.Fatalf("endpoint moved before H: %+v", row)
		}
		p, err := hStore.RegisterPeer(context.Background(), f.postCD("post-cd-uuid"))
		if err != nil {
			t.Fatalf("H after a final NACK must follow t1074: %v", err)
		}
		if row := readEndpointRow(t, f.seed.db, raceSlot); row.Session != "post-cd-uuid" || row.Generation != f.source.Generation+1 {
			t.Fatalf("endpoint after H = %+v, want post-cd-uuid at %d", row, f.source.Generation+1)
		}
		claims, err := f.seed.Claim(context.Background(), p, MaxBatch, 30*time.Second)
		if err != nil || len(claims) != 0 {
			t.Fatalf("new endpoint claimed %+v err=%v; the dispatch must stay unreleased", claims, err)
		}
		if _, err := f.seed.ReadBody(context.Background(), p, msg.ID, ""); err == nil {
			t.Fatal("dispatch body readable without a BOUND handoff")
		}
		if c := f.counts(t); c != (raceCounts{}) {
			t.Fatalf("counts after H = %+v, want all 0", c)
		}
	})

	// Orders (iv)-(vii): the adversary opens BEGIN IMMEDIATE, writes, and holds
	// at a barrier; the subject is observed blocked on its own handle.
	t.Run("order_iv_registration_reads_handoff_inside_its_transaction", func(t *testing.T) {
		f := newRaceFixture(t, false)
		vStore, hStore := f.open(t), f.open(t)
		proveDistinctHandles(t, vStore, hStore)
		reserved := readEndpointRow(t, f.seed.db, raceSlot)

		vGate := newStepGate(StepReserveInserted)
		vDone := make(chan racerResult, 1)
		go func() {
			h, err := vStore.ReserveHandoff(vGate.ctx(), f.reservation())
			vDone <- racerResult{h: h, err: err, at: time.Now()}
		}()
		vGate.await(t, "V")

		hDone := make(chan racerResult, 1)
		go func() {
			p, err := hStore.RegisterPeer(context.Background(), f.postCD("post-cd-uuid"))
			hDone <- racerResult{peer: p, err: err, at: time.Now()}
		}()
		if !observeBlocked(t, hStore, hDone) {
			t.Fatal("H_blocked_observed=false: H was not observed waiting on the SQLite lock")
		}
		t.Log("H_blocked_observed=true")
		close(vGate.release)
		v := receive(t, "V", vDone)
		if v.err != nil {
			t.Fatalf("V reservation: %v", v.err)
		}
		h := receive(t, "H", hDone)
		t.Logf("V_commit_at=%s H_return_at=%s", v.at.Format(time.RFC3339Nano), h.at.Format(time.RFC3339Nano))
		if !v.at.Before(h.at) {
			t.Fatalf("V commit %s is not earlier than H return %s", v.at, h.at)
		}
		requireNackReason(t, h.err, NackEndpointHandoffPending)
		if after := readEndpointRow(t, f.seed.db, raceSlot); after != reserved {
			t.Fatalf("row is not the reserved source tuple: reserved=%+v after=%+v", reserved, after)
		}
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_v_launcher_registration_finalizes_handoff_in_its_commit", func(t *testing.T) {
		f := newRaceFixture(t, true)
		vStore, aStore := f.open(t), f.open(t)
		proveDistinctHandles(t, vStore, aStore)

		vGate := newStepGate(StepReserveInserted)
		vDone := make(chan racerResult, 1)
		go func() {
			h, err := vStore.ReserveHandoff(vGate.ctx(), f.reservation())
			vDone <- racerResult{h: h, err: err, at: time.Now()}
		}()
		vGate.await(t, "V")

		aDone := make(chan racerResult, 1)
		go func() {
			p, err := aStore.RegisterLaunchPending(context.Background(), f.relaunch())
			aDone <- racerResult{peer: p, err: err, at: time.Now()}
		}()
		if !observeBlocked(t, aStore, aDone) {
			t.Fatal("A_blocked_observed=false: A was not observed waiting on the SQLite lock")
		}
		t.Log("A_blocked_observed=true")
		close(vGate.release)
		v := receive(t, "V", vDone)
		if v.err != nil {
			t.Fatalf("V reservation: %v", v.err)
		}
		a := receive(t, "A", aDone)
		t.Logf("V_commit_at=%s A_return_at=%s", v.at.Format(time.RFC3339Nano), a.at.Format(time.RFC3339Nano))
		if !v.at.Before(a.at) {
			t.Fatalf("V commit %s is not earlier than A return %s", v.at, a.at)
		}
		if a.err != nil {
			t.Fatalf("A launcher registration: %v", a.err)
		}
		row := readEndpointRow(t, f.seed.db, raceSlot)
		if !isLaunchPendingSession(row.Session) {
			t.Fatalf("row after A = %+v, want launch-pending", row)
		}
		if st, reason := f.handoffState(t, v.h.ID); st != HandoffNack || reason != NackStaleGeneration {
			t.Fatalf("handoff = %s/%s while the row is launch-pending, want NACK/%s from A's commit", st, reason, NackStaleGeneration)
		}
		if c := f.counts(t); c != (raceCounts{}) {
			t.Fatalf("launcher finalize wrote %+v, want no tombstone, receipt, or release", c)
		}
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_vi_control_registration_completes_before_reservation", func(t *testing.T) {
		f := newRaceFixture(t, false)
		hStore, vStore := f.open(t), f.open(t)
		proveDistinctHandles(t, hStore, vStore)
		p, err := hStore.RegisterPeer(context.Background(), f.postCD("post-cd-uuid"))
		if err != nil {
			t.Fatalf("H with no handoff follows t1074: %v", err)
		}
		row := readEndpointRow(t, f.seed.db, raceSlot)
		if row.Session != "post-cd-uuid" || row.Generation != f.source.Generation+1 || p.Generation != row.Generation {
			t.Fatalf("row after H = %+v peer=%+v", row, p)
		}
		h, err := vStore.ReserveHandoff(context.Background(), f.reservation())
		if err != nil {
			t.Fatalf("V after H: %v", err)
		}
		if h.Source.SessionUUID != row.Session || h.Source.Generation != row.Generation || h.Source.PID != row.PID || h.Source.ProcessStart != row.ProcessStart {
			t.Fatalf("V reserved %+v against row %+v", h.Source, row)
		}
	})

	t.Run("order_vii_reservation_reads_source_inside_its_transaction", func(t *testing.T) {
		f := newRaceFixture(t, true)
		aStore, vStore := f.open(t), f.open(t)
		proveDistinctHandles(t, aStore, vStore)

		// Adversary A: the production launcher provisional registration, held
		// at the REQ-FLH-017 read-and-finalize step after its slot-row write.
		aGate := newStepGate(StepRegisterFinalize)
		aDone := make(chan racerResult, 1)
		go func() {
			p, err := aStore.RegisterLaunchPending(aGate.ctx(), f.relaunch())
			aDone <- racerResult{peer: p, err: err, at: time.Now()}
		}()
		aGate.await(t, "A")

		// Subject V, pinned by its caller to the source tuple it resolved
		// before A began.
		source := f.source
		vDone := make(chan racerResult, 1)
		go func() {
			r := f.reservation()
			r.ExpectedSource = &source
			h, err := vStore.ReserveHandoff(context.Background(), r)
			vDone <- racerResult{h: h, err: err, at: time.Now()}
		}()
		if !observeBlocked(t, vStore, vDone) {
			t.Fatal("V_blocked_observed=false: subject was not observed waiting on the SQLite lock")
		}
		t.Log("V_blocked_observed=true")
		close(aGate.release)
		a := receive(t, "A", aDone)
		if a.err != nil {
			t.Fatalf("A launcher registration: %v", a.err)
		}
		committed := readEndpointRow(t, f.seed.db, raceSlot)
		v := receive(t, "V", vDone)
		t.Logf("A_commit_at=%s V_return_at=%s", a.at.Format(time.RFC3339Nano), v.at.Format(time.RFC3339Nano))
		if !a.at.Before(v.at) {
			t.Fatalf("timestamp order reversed: A commit %s is not earlier than V return %s", a.at, v.at)
		}
		requireNackReason(t, v.err, NackEndpointLaunchPending)
		if n := countLaneHandoffs(t, f.seed.db, raceSlot); n != 0 {
			t.Fatalf("reservation rows for lane = %d, want 0 (stale-tuple reservation committed)", n)
		}
		if after := readEndpointRow(t, f.seed.db, raceSlot); after != committed {
			t.Fatalf("slot row changed after subject: committed=%+v after=%+v", committed, after)
		}
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("unforced_200_iterations", func(t *testing.T) {
		pending, bound := 0, 0
		for i := 0; i < 200; i++ {
			outcome := runUnforcedRegisterVsRebind(t, i)
			if outcome == "i" {
				pending++
			} else {
				bound++
			}
		}
		t.Logf("UNFORCED_OUTCOMES i=%d ii=%d", pending, bound)
	})
}

// runUnforcedRegisterVsRebind runs one unforced H-versus-R race on a fresh
// broker and returns which forced end state it matched: "i" (H refused while
// pending, R bound) or "ii" (R bound first, H left the endpoint unchanged).
func runUnforcedRegisterVsRebind(t *testing.T, i int) string {
	t.Helper()
	f := newRaceFixture(t, false)
	f.toSwitchPending(t)
	hStore, rStore := f.seed, f.open(t)
	if err := requireDistinctRacerHandles(hStore, rStore); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var h, r racerResult
	runH := func() {
		defer wg.Done()
		<-start
		p, err := hStore.RegisterPeer(context.Background(), f.postCD("post-cd-uuid"))
		h = racerResult{peer: p, err: err, at: time.Now()}
	}
	runR := func() {
		defer wg.Done()
		<-start
		b, err := rStore.BindHandoff(context.Background(), f.h, f.evidence("post-cd-uuid"))
		r = racerResult{b: b, err: err, at: time.Now()}
	}
	wg.Add(2)
	// Alternate the spawn order so goroutine wake-up order does not decide
	// every iteration the same way; the commit order stays unforced.
	if i%2 == 0 {
		go runH()
		go runR()
	} else {
		go runR()
		go runH()
	}
	close(start)
	wg.Wait()
	if r.err != nil {
		t.Fatalf("iteration %d: R rebind: %v", i, r.err)
	}
	f.requireBoundEndState(t)
	if h.err == nil {
		if h.peer.SessionUUID != "post-cd-uuid" || h.peer.Generation != f.source.Generation+1 {
			t.Fatalf("iteration %d: H succeeded with %+v, want the bound endpoint unchanged", i, h.peer)
		}
		return "ii"
	}
	if reason, ok := HandoffNackReason(h.err); !ok || reason != NackEndpointHandoffPending {
		t.Fatalf("iteration %d: H err=%v, want nil or %s", i, h.err, NackEndpointHandoffPending)
	}
	return "i"
}

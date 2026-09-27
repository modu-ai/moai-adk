package hook

import (
	"context"
	"database/sql"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const ownerHelperEnv = "MOAI_T1082_OWNER_HELPER"

// TestFactoryHandoffOwnerHelperProcess is not a test. Re-executed as a child of
// the race test, it is a live process distinct from both the test process and
// the session owner the hook resolves, so the handoff's new endpoint owner is
// current under the t1074 live-owner rule. It lives until its stdin closes.
func TestFactoryHandoffOwnerHelperProcess(t *testing.T) {
	if os.Getenv(ownerHelperEnv) != "1" {
		return
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

// startLiveOwner starts the helper child and returns its PID and process-start.
// The child is killed and reaped by the test's cleanup on every path.
func startLiveOwner(t *testing.T) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestFactoryHandoffOwnerHelperProcess$", "-test.timeout=300s")
	cmd.Env = append(os.Environ(), ownerHelperEnv+"=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start live owner helper: %v", err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
		if state == homestate.ProcessIdentityLive && start != "" {
			return cmd.Process.Pid, start
		}
		if time.Now().After(deadline) {
			t.Fatalf("live owner helper %d never became probeable (state=%v)", cmd.Process.Pid, state)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// hookStepGate holds a racer inside its write transaction at one named step.
type hookStepGate struct {
	name             string
	reached, release chan struct{}
	once             sync.Once
}

func newHookStepGate(name string) *hookStepGate {
	return &hookStepGate{name: name, reached: make(chan struct{}), release: make(chan struct{})}
}

func (g *hookStepGate) ctx() context.Context {
	return factorymsg.WithStepHook(context.Background(), func(step string) error {
		if step == g.name {
			g.once.Do(func() { close(g.reached) })
			<-g.release
		}
		return nil
	})
}

func (g *hookStepGate) await(t *testing.T, who string) {
	t.Helper()
	select {
	case <-g.reached:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s never reached step %q", who, g.name)
	}
}

type hookRacerResult struct {
	peer factorymsg.Peer
	b    factorymsg.HandoffBinding
	err  error
	at   time.Time
}

func receiveHookRacer(t *testing.T, who string, done <-chan hookRacerResult) hookRacerResult {
	t.Helper()
	select {
	case r := <-done:
		return r
	case <-time.After(5 * time.Second):
		t.Fatalf("%s did not return", who)
		return hookRacerResult{}
	}
}

// observeHookRacerBlocked is part (2) of the separate-handle proof on the
// waiting racer's own handle.
func observeHookRacerBlocked(t *testing.T, waiter *factorymsg.Store, done <-chan hookRacerResult) bool {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		st := waiter.HandleStats()
		if st.InUse == 1 && st.WaitCount == 0 {
			time.Sleep(150 * time.Millisecond)
			select {
			case r := <-done:
				t.Fatalf("waiting racer returned while the holder kept its transaction: err=%v", r.err)
			default:
			}
			st = waiter.HandleStats()
			return st.InUse == 1 && st.WaitCount == 0
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// proveHookHandlesDistinct runs parts (1) and (3) of the separate-handle proof.
func proveHookHandlesDistinct(t *testing.T, a, b *factorymsg.Store) {
	t.Helper()
	if !factorymsg.SharesHandle(a, a) {
		t.Fatal("distinct-handle guard accepted a shared pair")
	}
	if a == b || factorymsg.SharesHandle(a, b) {
		t.Fatal("racer handles are not distinct")
	}
}

type launchRaceRow struct {
	Session, ProcessStart, UpdatedAt string
	Generation                       int64
	PID                              int
}

// launchRaceFixture is one broker with a bound lead and a lane-1 source bound
// through the production launcher path with a fake process-start (not
// current), a dispatch waiting for the source, and a headless handoff in
// SWITCH_PENDING_HEADLESS with official relocation evidence recorded.
type launchRaceFixture struct {
	root, run  string
	seed       *factorymsg.Store
	db         *sql.DB
	lead       factorymsg.Peer
	source     factorymsg.Peer
	h          factorymsg.Handoff
	dispatchID string
	ownerPID   int
	ownerStart string
	newPID     int
	newStart   string
}

const launchRaceSlot = "lane-1"

// newLaunchRaceFixture builds the fixture; hookReady also records the active
// factory run the production hook path validates (orders (iv)/(v) only).
func newLaunchRaceFixture(t *testing.T, hookReady bool, ownerPID int, ownerStart string, newPID int, newStart string) *launchRaceFixture {
	t.Helper()
	ctx := context.Background()
	f := &launchRaceFixture{root: t.TempDir(), run: "run-launch-race", ownerPID: ownerPID, ownerStart: ownerStart, newPID: newPID, newStart: newStart}
	if hookReady {
		recordActiveFactoryRun(t, f.root, f.run)
	}
	var err error
	if f.seed, err = factorymsg.Open(f.root, f.run); err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "seed broker", f.seed)
	key := homestate.ProjectKey(f.root)
	f.lead = bindThroughLauncher(t, f.seed, factorymsg.Peer{ProjectKey: key, RunID: f.run, Backend: "claude", Role: "lead", Slot: "lead", PID: ownerPID, ProcessStart: ownerStart}, "lead-uuid")
	f.source = bindThroughLauncher(t, f.seed, factorymsg.Peer{ProjectKey: key, RunID: f.run, Backend: "codex", Role: "worker", Slot: launchRaceSlot, PID: os.Getpid(), ProcessStart: "fake-source-start"}, "src-uuid")
	env, err := f.seed.Send(ctx, factorymsg.SendRequest{
		From: f.lead, To: f.source, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: "dispatch-launch-race", TaskRef: "t1082",
		CorrelationID: "c-launch-race", TTL: time.Hour, Payload: []byte("dispatch body"),
	})
	if err != nil {
		t.Fatal(err)
	}
	f.dispatchID = env.ID
	h, err := f.seed.ReserveHandoff(ctx, factorymsg.HandoffReservation{
		Slot: launchRaceSlot, CardID: "t1082", SpecID: "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001", Mode: factorymsg.HandoffModeHeadless,
		DevelopPin: strings.Repeat("a", 40), TargetPath: filepath.Join(f.root, ".claude", "worktrees", "t1082"), TargetBranch: "WT-lane-handoff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if h, err = f.seed.MarkHandoffWTReady(ctx, h); err != nil {
		t.Fatal(err)
	}
	if h, err = f.seed.MarkHandoffSwitchPendingHeadless(ctx, h); err != nil {
		t.Fatal(err)
	}
	if err := f.seed.RecordHeadlessRelocation(ctx, h, factorymsg.HeadlessRelocation{
		Method: factorymsg.RelocationMethodThreadFork, SourceThreadID: "thr-source", ThreadID: "thr-forked", ForkedFromID: "thr-source", ThreadStarted: true,
		RequestCwd: h.TargetPath, ResponseCwd: h.TargetPath, ReadbackCwd: h.TargetPath, ReadbackBranch: h.TargetBranch, ReadbackHead: h.DevelopPin,
	}); err != nil {
		t.Fatal(err)
	}
	f.h = h
	path, err := factorymsg.BrokerPath(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if f.db, err = sql.Open("sqlite", "file:"+path); err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "read-only observer", f.db)
	return f
}

func bindThroughLauncher(t *testing.T, s *factorymsg.Store, p factorymsg.Peer, session string) factorymsg.Peer {
	t.Helper()
	ctx := context.Background()
	pending, err := s.RegisterLaunchPending(ctx, p)
	if err != nil {
		t.Fatalf("launcher registration %s: %v", p.Slot, err)
	}
	pending.SessionUUID = session
	bound, ok, err := s.BindLaunchPending(ctx, pending)
	if err != nil || !ok {
		t.Fatalf("launcher bind %s ok=%v err=%v", p.Slot, ok, err)
	}
	return bound
}

func (f *launchRaceFixture) open(t *testing.T) *factorymsg.Store {
	t.Helper()
	s, err := factorymsg.Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "racer broker", s)
	return s
}

// relaunch is racer A's launcher registration: the relaunched process owned
// by the identity the production hook path resolves in this test process.
func (f *launchRaceFixture) relaunch() factorymsg.Peer {
	return factorymsg.Peer{ProjectKey: homestate.ProjectKey(f.root), RunID: f.run, Backend: "codex", Role: "worker", Slot: launchRaceSlot, PID: f.ownerPID, ProcessStart: f.ownerStart}
}

// evidence is racer B's matching headless evidence: the returned thread id,
// owned by the live helper process.
func (f *launchRaceFixture) evidence() factorymsg.HandoffBindEvidence {
	return factorymsg.HandoffBindEvidence{
		Mode: factorymsg.HandoffModeHeadless, Nonce: f.h.Nonce, CardID: f.h.CardID, SpecID: f.h.SpecID,
		SessionUUID: "thr-forked", PID: f.newPID, ProcessStart: f.newStart,
		Cwd: f.h.TargetPath, WorktreeRoot: f.h.TargetPath, Branch: f.h.TargetBranch, Head: f.h.DevelopPin,
	}
}

func (f *launchRaceFixture) rows(t *testing.T) []launchRaceRow {
	t.Helper()
	rs, err := f.db.Query(`SELECT session_uuid,generation,pid,process_start,updated_at FROM peers WHERE slot=?`, launchRaceSlot)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rs.Close() }()
	var out []launchRaceRow
	for rs.Next() {
		var r launchRaceRow
		if err := rs.Scan(&r.Session, &r.Generation, &r.PID, &r.ProcessStart, &r.UpdatedAt); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

func (f *launchRaceFixture) row(t *testing.T) launchRaceRow {
	t.Helper()
	rs := f.rows(t)
	if len(rs) != 1 {
		t.Fatalf("slot rows = %d, want exactly 1: %+v", len(rs), rs)
	}
	return rs[0]
}

func (f *launchRaceFixture) handoff(t *testing.T) (string, string) {
	t.Helper()
	var state, reason string
	if err := f.db.QueryRow(`SELECT state,reason FROM lane_handoffs WHERE id=?`, f.h.ID).Scan(&state, &reason); err != nil {
		t.Fatal(err)
	}
	return state, reason
}

type launchRaceCounts struct{ Tombstones, Receipts, Releases int }

func (f *launchRaceFixture) counts(t *testing.T) launchRaceCounts {
	t.Helper()
	var c launchRaceCounts
	for q, dst := range map[string]*int{
		`SELECT count(*) FROM lane_endpoint_tombstones`: &c.Tombstones,
		`SELECT count(*) FROM lane_handoff_receipts`:    &c.Receipts,
		`SELECT count(*) FROM lane_dispatch_releases`:   &c.Releases,
	} {
		if err := f.db.QueryRow(q).Scan(dst); err != nil {
			t.Fatal(err)
		}
	}
	return c
}

// requireNoOpenMismatch: no non-final handoff whose reserved source tuple
// differs from the row, and no launch-pending row.
func (f *launchRaceFixture) requireInvariants(t *testing.T) launchRaceRow {
	t.Helper()
	row := f.row(t)
	if strings.HasPrefix(row.Session, "launch-pending:") {
		t.Fatalf("orphan launch-pending row: %+v", row)
	}
	if row.Generation <= f.source.Generation {
		t.Fatalf("generation %d is not greater than g=%d", row.Generation, f.source.Generation)
	}
	var n int
	if err := f.db.QueryRow(`SELECT count(*) FROM lane_handoffs WHERE slot=? AND state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS') AND (source_session!=? OR source_generation!=? OR source_pid!=? OR source_process_start!=?)`,
		launchRaceSlot, row.Session, row.Generation, row.PID, row.ProcessStart).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("%d non-final handoff(s) whose source tuple differs from the row %+v", n, row)
	}
	return row
}

// requireLauncherWon is the end state when A's registration committed first.
func (f *launchRaceFixture) requireLauncherWon(t *testing.T, session string, wantGen int64) {
	t.Helper()
	row := f.requireInvariants(t)
	if row.Session != session || row.Generation != wantGen || row.PID != f.ownerPID || row.ProcessStart != f.ownerStart {
		t.Fatalf("row = %+v, want %s at generation %d owned by the relaunched process", row, session, wantGen)
	}
	if st, reason := f.handoff(t); st != factorymsg.HandoffNack || reason != factorymsg.NackStaleGeneration {
		t.Fatalf("handoff = %s/%s, want NACK/%s", st, reason, factorymsg.NackStaleGeneration)
	}
	if c := f.counts(t); c != (launchRaceCounts{}) {
		t.Fatalf("counts = %+v, want no tombstone, receipt, or release", c)
	}
}

// requireRebindWon is the end state when B's rebind committed first.
func (f *launchRaceFixture) requireRebindWon(t *testing.T) {
	t.Helper()
	row := f.requireInvariants(t)
	if row.Session != "thr-forked" || row.Generation != f.source.Generation+1 || row.PID != f.newPID || row.ProcessStart != f.newStart {
		t.Fatalf("row = %+v, want the handoff endpoint thr-forked at generation %d", row, f.source.Generation+1)
	}
	if st, _ := f.handoff(t); st != factorymsg.HandoffBound {
		t.Fatalf("handoff = %s, want BOUND", st)
	}
	if c := f.counts(t); c.Tombstones != 1 || c.Receipts != 1 {
		t.Fatalf("counts = %+v, want exactly one tombstone and one BOUND receipt", c)
	}
}

func requireHookNack(t *testing.T, err error, want string) {
	t.Helper()
	if reason, ok := factorymsg.HandoffNackReason(err); !ok || reason != want {
		t.Fatalf("err=%v, want NACK %s", err, want)
	}
}

// generationTrail asserts the generations observed across commits never
// decreased.
func generationTrail(t *testing.T, gens ...int64) {
	t.Helper()
	for i := 1; i < len(gens); i++ {
		if gens[i] < gens[i-1] {
			t.Fatalf("generation decreased across observed commits: %v", gens)
		}
	}
}

// TestFactoryLaneHandoffRebindVsLaunchBindRace is the AC-FLH-018 named test:
// the t1074 launcher registration/bind (A) against the handoff rebind (B) on
// one slot, forced orders (i)-(v) and 200 unforced iterations.
func TestFactoryLaneHandoffRebindVsLaunchBindRace(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(config.EnvMoaiKanbanID, "run-launch-race")
	t.Setenv(config.EnvMoaiFactoryWorker, launchRaceSlot)
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")
	ownerPID, ownerStart := factoryHookOwnerIdentity(t)
	newPID, newStart := startLiveOwner(t)
	if newPID == ownerPID || newPID == os.Getpid() || newStart == ownerStart {
		t.Fatalf("handoff owner %d/%s is not distinct from the relaunch owner %d/%s", newPID, newStart, ownerPID, ownerStart)
	}
	fixture := func(t *testing.T) *launchRaceFixture {
		return newLaunchRaceFixture(t, true, ownerPID, ownerStart, newPID, newStart)
	}
	storeFixture := func(t *testing.T) *launchRaceFixture {
		return newLaunchRaceFixture(t, false, ownerPID, ownerStart, newPID, newStart)
	}

	t.Run("order_i_A_register_then_B_then_A_bind", func(t *testing.T) {
		f := fixture(t)
		aStore, bStore := f.open(t), f.open(t)
		proveHookHandlesDistinct(t, aStore, bStore)
		g := f.source.Generation

		aGate := newHookStepGate(factorymsg.StepRegisterFinalize)
		aDone := make(chan hookRacerResult, 1)
		go func() {
			p, err := aStore.RegisterLaunchPending(aGate.ctx(), f.relaunch())
			aDone <- hookRacerResult{peer: p, err: err, at: time.Now()}
		}()
		aGate.await(t, "A")

		bGate := newHookStepGate(factorymsg.StepBindBegun)
		bDone := make(chan hookRacerResult, 1)
		go func() {
			b, err := bStore.BindHandoff(bGate.ctx(), f.h, f.evidence())
			bDone <- hookRacerResult{b: b, err: err, at: time.Now()}
		}()
		if !observeHookRacerBlocked(t, bStore, bDone) {
			t.Fatal("B_blocked_observed=false: B was not observed waiting on the SQLite lock")
		}
		t.Log("B_blocked_observed=true")
		close(aGate.release)
		a := receiveHookRacer(t, "A", aDone)
		if a.err != nil {
			t.Fatalf("A registration: %v", a.err)
		}
		// B has taken the lock but read nothing yet: the handoff A's commit
		// left behind is already final.
		bGate.await(t, "B")
		if st, reason := f.handoff(t); st != factorymsg.HandoffNack || reason != factorymsg.NackStaleGeneration {
			t.Fatalf("handoff after A's commit, before B reads = %s/%s, want NACK/%s", st, reason, factorymsg.NackStaleGeneration)
		}
		afterA := f.row(t)
		close(bGate.release)
		b := receiveHookRacer(t, "B", bDone)
		if !a.at.Before(b.at) {
			t.Fatalf("A commit %s is not earlier than B return %s", a.at, b.at)
		}
		requireHookNack(t, b.err, factorymsg.NackStaleGeneration)
		if f.row(t) != afterA {
			t.Fatal("B changed the row after returning STALE_GENERATION")
		}
		pending := a.peer
		pending.SessionUUID = "relaunch-session"
		bound, ok, err := aStore.BindLaunchPending(context.Background(), pending)
		if err != nil || !ok {
			t.Fatalf("A bind ok=%v err=%v", ok, err)
		}
		f.requireLauncherWon(t, "relaunch-session", g+2)
		generationTrail(t, g, afterA.Generation, bound.Generation)
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_ii_B_then_A", func(t *testing.T) {
		f := fixture(t)
		aStore, bStore := f.open(t), f.open(t)
		proveHookHandlesDistinct(t, bStore, aStore)
		g := f.source.Generation

		bGate := newHookStepGate("locked")
		bDone := make(chan hookRacerResult, 1)
		go func() {
			b, err := bStore.BindHandoff(bGate.ctx(), f.h, f.evidence())
			bDone <- hookRacerResult{b: b, err: err, at: time.Now()}
		}()
		bGate.await(t, "B")
		aDone := make(chan hookRacerResult, 1)
		go func() {
			p, err := aStore.RegisterLaunchPending(context.Background(), f.relaunch())
			aDone <- hookRacerResult{peer: p, err: err, at: time.Now()}
		}()
		if !observeHookRacerBlocked(t, aStore, aDone) {
			t.Fatal("A_blocked_observed=false: A was not observed waiting on the SQLite lock")
		}
		t.Log("A_blocked_observed=true")
		close(bGate.release)
		b := receiveHookRacer(t, "B", bDone)
		if b.err != nil {
			t.Fatalf("B rebind: %v", b.err)
		}
		a := receiveHookRacer(t, "A", aDone)
		if !b.at.Before(a.at) {
			t.Fatalf("B commit %s is not earlier than A return %s", b.at, a.at)
		}
		if a.err == nil || !strings.Contains(a.err.Error(), "live owner") {
			t.Fatalf("A registration err=%v, want the t1074 live-owner rejection", a.err)
		}
		f.requireRebindWon(t)
		afterB := f.row(t)
		want := f.relaunch()
		want.SessionUUID = "relaunch-session"
		want.Generation = 1
		if _, bound, err := aStore.BindLaunchPending(context.Background(), want); err != nil || bound {
			t.Fatalf("A bind after a rejected registration bound=%v err=%v", bound, err)
		}
		if f.row(t) != afterB {
			t.Fatal("A changed the handoff-bound row")
		}
		f.requireRebindWon(t)
		generationTrail(t, g, afterB.Generation)
		t.Log("RACER_HANDLES_DISTINCT=2")
	})

	t.Run("order_iii_A_then_B", func(t *testing.T) {
		f := fixture(t)
		aStore, bStore := f.open(t), f.open(t)
		proveHookHandlesDistinct(t, aStore, bStore)
		g := f.source.Generation
		pending, err := aStore.RegisterLaunchPending(context.Background(), f.relaunch())
		if err != nil {
			t.Fatal(err)
		}
		if st, reason := f.handoff(t); st != factorymsg.HandoffNack || reason != factorymsg.NackStaleGeneration {
			t.Fatalf("handoff right after A's commit = %s/%s, want NACK/%s", st, reason, factorymsg.NackStaleGeneration)
		}
		afterRegister := f.row(t)
		pending.SessionUUID = "relaunch-session"
		bound, ok, err := aStore.BindLaunchPending(context.Background(), pending)
		if err != nil || !ok {
			t.Fatalf("A bind ok=%v err=%v", ok, err)
		}
		before := f.row(t)
		_, err = bStore.BindHandoff(context.Background(), f.h, f.evidence())
		requireHookNack(t, err, factorymsg.NackStaleGeneration)
		if f.row(t) != before {
			t.Fatal("B changed the row after returning STALE_GENERATION")
		}
		f.requireLauncherWon(t, "relaunch-session", g+2)
		generationTrail(t, g, afterRegister.Generation, bound.Generation)
	})

	t.Run("order_iv_session_start_before_A_register", func(t *testing.T) {
		f := fixture(t)
		ctx := context.Background()
		g := f.source.Generation
		input := &HookInput{SessionID: "relaunch-session", ProjectDir: f.root, CWD: f.root}
		before := f.row(t)
		if notice := registerFactorySessionStartPeer(ctx, input); notice != "" {
			t.Fatalf("SessionStart on the still-bound source row emitted %q, want no bind", notice)
		}
		if f.row(t) != before {
			t.Fatal("SessionStart changed the still-bound source row")
		}
		if st, _ := f.handoff(t); st != factorymsg.HandoffSwitchPendingHeadless {
			t.Fatalf("handoff after SessionStart = %s, want still pending", st)
		}
		aStore := f.open(t)
		if _, err := aStore.RegisterLaunchPending(ctx, f.relaunch()); err != nil {
			t.Fatal(err)
		}
		if st, reason := f.handoff(t); st != factorymsg.HandoffNack || reason != factorymsg.NackStaleGeneration {
			t.Fatalf("handoff after A's commit = %s/%s, want NACK/%s", st, reason, factorymsg.NackStaleGeneration)
		}
		afterRegister := f.row(t)
		input.Prompt = "Continue the assigned factory work."
		if notice := registerFactoryUserPromptPeer(ctx, input); !strings.Contains(notice, "factory messaging bound") {
			t.Fatalf("UserPromptSubmit did not bind the provisional row: %q", notice)
		}
		afterPrompt := f.row(t)
		bStore := f.open(t)
		_, err := bStore.BindHandoff(ctx, f.h, f.evidence())
		requireHookNack(t, err, factorymsg.NackStaleGeneration)
		f.requireLauncherWon(t, "relaunch-session", g+2)
		generationTrail(t, g, afterRegister.Generation, afterPrompt.Generation)

		current, err := f.seed.ResolveLane(ctx, launchRaceSlot)
		if err != nil {
			t.Fatal(err)
		}
		if claims, err := f.seed.Claim(ctx, current, factorymsg.MaxBatch, 30*time.Second); err != nil || len(claims) != 0 {
			t.Fatalf("new endpoint claimed %+v err=%v; the dispatch was never released", claims, err)
		}
		if _, err := f.seed.Claim(ctx, f.source, factorymsg.MaxBatch, 30*time.Second); err == nil {
			t.Fatal("the replaced source endpoint could still claim its dispatch body")
		}
		if _, err := f.seed.ReadBody(ctx, current, f.dispatchID, ""); err == nil {
			t.Fatal("dispatch body readable without a BOUND handoff")
		}
	})

	t.Run("order_v_alias_correction_after_A_register", func(t *testing.T) {
		f := fixture(t)
		ctx := context.Background()
		g := f.source.Generation
		aStore := f.open(t)
		if _, err := aStore.RegisterLaunchPending(ctx, f.relaunch()); err != nil {
			t.Fatal(err)
		}
		if st, reason := f.handoff(t); st != factorymsg.HandoffNack || reason != factorymsg.NackStaleGeneration {
			t.Fatalf("handoff after A's commit = %s/%s, want NACK/%s", st, reason, factorymsg.NackStaleGeneration)
		}
		alias := &HookInput{SessionID: "session-start-alias", ProjectDir: f.root, CWD: f.root}
		if notice := registerFactorySessionStartPeer(ctx, alias); !strings.Contains(notice, "factory messaging bound") {
			t.Fatalf("SessionStart did not bind the alias: %q", notice)
		}
		if row := f.row(t); row.Session != alias.SessionID || row.Generation != g+2 {
			t.Fatalf("alias row = %+v, want %s at %d", row, alias.SessionID, g+2)
		}
		real := &HookInput{SessionID: "real-session", ProjectDir: f.root, CWD: f.root, Prompt: "Bind the real session."}
		if notice := registerFactoryUserPromptPeer(ctx, real); !strings.Contains(notice, "factory messaging bound") {
			t.Fatalf("UserPromptSubmit did not correct the alias: %q", notice)
		}
		f.requireLauncherWon(t, "real-session", g+3)
	})

	t.Run("unforced_200_iterations", func(t *testing.T) {
		launcher, rebind := 0, 0
		for i := 0; i < 200; i++ {
			if runUnforcedLaunchVsRebind(t, i, storeFixture) {
				launcher++
			} else {
				rebind++
			}
		}
		t.Logf("UNFORCED_OUTCOMES launcher_first=%d rebind_first=%d", launcher, rebind)
	})
}

// runUnforcedLaunchVsRebind runs A (register then bind) against B on a fresh
// broker and reports whether the launcher registration committed first.
func runUnforcedLaunchVsRebind(t *testing.T, i int, fixture func(*testing.T) *launchRaceFixture) bool {
	t.Helper()
	f := fixture(t)
	aStore, bStore := f.seed, f.open(t)
	if factorymsg.SharesHandle(aStore, bStore) {
		t.Fatal("racer handles are not distinct")
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var regErr, bindErr, rebindErr error
	// Both racers' arguments are built before the start signal: f.relaunch()
	// resolves the project key, which took ~37ms and, inside the goroutine,
	// handed B the lock in every iteration.
	relaunch, evidence := f.relaunch(), f.evidence()
	runA := func() {
		defer wg.Done()
		<-start
		pending, err := aStore.RegisterLaunchPending(context.Background(), relaunch)
		regErr = err
		if err != nil {
			return
		}
		pending.SessionUUID = "relaunch-session"
		_, _, bindErr = aStore.BindLaunchPending(context.Background(), pending)
	}
	runB := func() {
		defer wg.Done()
		<-start
		_, rebindErr = bStore.BindHandoff(context.Background(), f.h, evidence)
	}
	wg.Add(2)
	// Alternate which racer is spawned first so the runtime's goroutine
	// wake-up order does not decide every iteration the same way. Neither
	// racer waits for the other; the order of their commits is unforced.
	if i%2 == 0 {
		go runA()
		go runB()
	} else {
		go runB()
		go runA()
	}
	close(start)
	wg.Wait()
	if bindErr != nil {
		t.Fatalf("iteration %d: A bind: %v", i, bindErr)
	}
	if rebindErr == nil {
		if regErr == nil || !strings.Contains(regErr.Error(), "live owner") {
			t.Fatalf("iteration %d: B bound but A registration err=%v, want the live-owner rejection", i, regErr)
		}
		f.requireRebindWon(t)
		return false
	}
	if regErr != nil {
		t.Fatalf("iteration %d: A registration %v and B %v both lost", i, regErr, rebindErr)
	}
	requireHookNack(t, rebindErr, factorymsg.NackStaleGeneration)
	f.requireLauncherWon(t, "relaunch-session", f.source.Generation+2)
	return true
}

// TestFactoryUserPromptRegistrationSurfacesHandoffRefusals pins the hook
// surface of REQ-FLH-018: a UserPromptSubmit registration refused because the
// lane's handoff is not final, or because its session was replaced by a
// handoff rebind, returns a named notice (fail-open) and never rotates the
// endpoint.
func TestFactoryUserPromptRegistrationSurfacesHandoffRefusals(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	run := "run-prompt-refusal"
	recordActiveFactoryRun(t, root, run)
	t.Setenv(config.EnvMoaiKanbanID, run)
	t.Setenv(config.EnvMoaiFactoryWorker, launchRaceSlot)
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")
	ctx := context.Background()
	owner, start := factoryHookOwnerIdentity(t)
	s, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", s)
	source := bindThroughLauncher(t, s, factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "codex", Role: "worker", Slot: launchRaceSlot, PID: owner, ProcessStart: start}, "src-uuid")
	h, err := s.ReserveHandoff(ctx, factorymsg.HandoffReservation{
		Slot: launchRaceSlot, CardID: "t1082", SpecID: "SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001", Mode: factorymsg.HandoffModeInteractive,
		DevelopPin: strings.Repeat("a", 40), TargetPath: filepath.Join(root, ".claude", "worktrees", "t1082"), TargetBranch: "WT-lane-handoff",
	})
	if err != nil {
		t.Fatal(err)
	}
	if h, err = s.MarkHandoffWTReady(ctx, h); err != nil {
		t.Fatal(err)
	}
	if h, err = s.MarkHandoffSwitchPendingInteractive(ctx, h); err != nil {
		t.Fatal(err)
	}
	lane := func() factorymsg.Peer {
		p, err := s.ResolveLane(ctx, launchRaceSlot)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}

	post := &HookInput{SessionID: "post-cd-uuid", ProjectDir: root, CWD: root, Prompt: "Continue after /cd."}
	notice := registerFactoryUserPromptPeer(ctx, post)
	if !strings.Contains(notice, "factory handoff pending") || !strings.Contains(notice, factorymsg.NackEndpointHandoffPending) {
		t.Fatalf("pending-handoff notice = %q", notice)
	}
	if got := lane(); got != source {
		t.Fatalf("UserPromptSubmit moved the endpoint during a pending handoff: before=%+v after=%+v", source, got)
	}

	if _, err := s.BindHandoff(ctx, h, factorymsg.HandoffBindEvidence{
		Mode: factorymsg.HandoffModeInteractive, Nonce: h.Nonce, CardID: h.CardID, SpecID: h.SpecID,
		SessionUUID: "post-cd-uuid", PID: owner, ProcessStart: start,
		Cwd: h.TargetPath, WorktreeRoot: h.TargetPath, Branch: h.TargetBranch, Head: h.DevelopPin,
	}); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	bound := lane()
	stale := &HookInput{SessionID: "src-uuid", ProjectDir: root, CWD: root, Prompt: "The old session speaks again."}
	notice = registerFactoryUserPromptPeer(ctx, stale)
	if !strings.Contains(notice, "factory endpoint replaced") || !strings.Contains(notice, factorymsg.NackStaleEndpoint) || !strings.Contains(notice, "post-cd-uuid") {
		t.Fatalf("stale-endpoint notice = %q", notice)
	}
	if got := lane(); got != bound {
		t.Fatalf("the tombstoned session moved the endpoint: before=%+v after=%+v", bound, got)
	}
}

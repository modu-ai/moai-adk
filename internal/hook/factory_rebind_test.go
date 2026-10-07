package hook

// factory_rebind_test.go — SPEC-FACTORY-STALE-RUN-HEAL-001 M2: the hook's
// single-measurement current-vocabulary path and the lane rebind state machine
// (REQ-SRH-004..010, AC-SRH-008..014).
//
// No assertion here depends on machine load: every budget the path reads is a
// package variable the fixture pins generously (rebindEnv), except the two
// tests that exhaust a budget on purpose (one nanosecond) or bound the rebind
// path's elapsed time against the production bind budget.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// rebindEnv seeds the launch environment of a factory session naming run X
// under the given lane label and backend, scrubs every variable a sibling test
// could leave behind, and pins every budget so no verdict depends on load. The
// hook's owner identity is this test process (MOAI_SESSION_PID names itself).
func rebindEnv(t *testing.T, run, label, backend string) {
	t.Helper()
	m3ScrubEnv(t)
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(config.EnvFactoryRunID, run)
	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	t.Setenv(config.EnvMoaiFactoryWorker, label)
	t.Setenv(config.EnvFactoryBackend, backend)
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(os.Getpid()))
	pinFactoryGateBudget(t)
	prevBind, prevInspect := factoryBindBudget, factoryHookInspectionDeadline
	factoryBindBudget, factoryHookInspectionDeadline = time.Minute, 30*time.Second
	t.Cleanup(func() { factoryBindBudget, factoryHookInspectionDeadline = prevBind, prevInspect })
}

// rebindRoot seeds a project root whose factory database holds the given runs
// ("<id>=<status>"; status active is recorded through the standard helper).
func rebindRoot(t *testing.T, runs ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, spec := range runs {
		id, status, _ := strings.Cut(spec, "=")
		if status == "active" {
			recordActiveFactoryRun(t, root, id)
		} else {
			recordFactoryRunWithStatus(t, root, id, status)
		}
	}
	return root
}

// rebindPrompt drives one UserPromptSubmit turn of session sid and returns the
// additionalContext the hook produced.
func rebindPrompt(t *testing.T, root, sid string) string {
	t.Helper()
	out, err := NewUserPromptSubmitHandler(nil).Handle(context.Background(), &HookInput{SessionID: sid, ProjectDir: root, CWD: root, Prompt: "please continue the card"})
	if err != nil {
		t.Fatalf("the hook must fail open, not error: %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		return ""
	}
	return out.HookSpecificOutput.AdditionalContext
}

// brokerLanes reads the peer rows of a run's broker without creating it:
// exists is false when the broker file was never created.
func brokerLanes(t *testing.T, root, run string) (lanes []factorymsg.LaneStatus, exists bool) {
	t.Helper()
	path, err := factorymsg.BrokerPath(root, run)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		return nil, false
	}
	s, err := factorymsg.OpenExistingWithDeadline(root, run, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer closeFactoryHookStore(s)
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, lane := range status.Lanes {
		lane.ObservedAt = time.Time{}
		lanes = append(lanes, lane)
	}
	return lanes, true
}

// laneRows renders a broker's peer rows as "slot|session|generation".
func laneRows(t *testing.T, root, run string) string {
	t.Helper()
	lanes, _ := brokerLanes(t, root, run)
	rows := make([]string, 0, len(lanes))
	for _, l := range lanes {
		rows = append(rows, fmt.Sprintf("%s|%s|%d", l.Slot, l.SessionUUID, l.Generation))
	}
	return strings.Join(rows, ",")
}

func brokerExists(t *testing.T, root, run string) bool {
	t.Helper()
	_, exists := brokerLanes(t, root, run)
	return exists
}

// rebindMarker reads the notice marker the hook wrote for a session identity.
func rebindMarker(t *testing.T, root, sid string) factoryNoticeMarker {
	t.Helper()
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	return readFactoryNoticeMarker(dbPath, sid)
}

// workersView is the leader-visible slot state a rebind must not touch (DP11):
// the workers registry rows and the free-slot view.
func workersView(t *testing.T, root string) string {
	t.Helper()
	reg := factory.LoadFactoryRegistry(factory.FactoryRegistryPath(root))
	free := factory.FactoryFreeSlots(root, 4, func(int) bool { return true })
	return fmt.Sprintf("%v | free=%v", reg, free)
}

func noFactoryNotice(t *testing.T, label, got string) {
	t.Helper()
	if strings.Contains(got, "factory lane") || strings.Contains(got, "degraded") {
		t.Fatalf("%s repeated a notice: %q", label, got)
	}
}

func TestLaneRebindsIntoSoleActiveRun(t *testing.T) { // AC-SRH-008
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=retired", "runY=active")
	if _, err := factory.ClaimFactoryLane(root, "lane-3", false, os.Getpid(), "runX", func(int) bool { return true }); err != nil {
		t.Fatalf("seed the lane's workers row: %v", err)
	}
	viewBefore := workersView(t, root)
	if !strings.Contains(viewBefore, "lane-3") {
		t.Fatalf("precondition: the workers registry holds no lane-3 row: %s", viewBefore)
	}

	first := rebindPrompt(t, root, "s1")
	for _, want := range []string{"factory lane rebound:", "runX", "runY", "lane-3", "generation 1"} {
		if !strings.Contains(first, want) {
			t.Fatalf("prompt 1 lacks %q: %q", want, first)
		}
	}
	if strings.Contains(first, "degraded") {
		t.Fatalf("prompt 1 still says degraded: %q", first)
	}
	if got := laneRows(t, root, "runY"); got != "lane-3|s1|1" {
		t.Fatalf("runY peers = %q, want lane-3|s1|1", got)
	}
	if brokerExists(t, root, "runX") {
		t.Fatal("the retired run's broker was created")
	}
	noFactoryNotice(t, "prompt 2", rebindPrompt(t, root, "s1"))
	if got := rebindMarker(t, root, "s1"); got.State != "rebound:runY" || got.PrescriptionEmittedAt != "" || got.UnbindEmittedAt != "" {
		t.Fatalf("marker = %+v, want only state rebound:runY", got)
	}
	if viewAfter := workersView(t, root); viewAfter != viewBefore {
		t.Fatalf("the rebind touched the workers registry (DP11):\nbefore %s\nafter  %s", viewBefore, viewAfter)
	}
}

// A healthy lane — its own run active, a second run active as well — must bind
// into its own run and never rebind: the not-active guard is the whole point.
func TestHealthyLaneNeverRebinds(t *testing.T) { // AC-SRH-008 control, kills "rebinds without the guard"
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active", "runY=active")
	first := rebindPrompt(t, root, "s1")
	if !strings.Contains(first, "factory messaging bound: run=runX slot=lane-3 generation=1") {
		t.Fatalf("healthy lane did not bind into its own run: %q", first)
	}
	if got := laneRows(t, root, "runX"); got != "lane-3|s1|1" {
		t.Fatalf("runX peers = %q", got)
	}
	if brokerExists(t, root, "runY") {
		t.Fatal("a healthy lane opened the other run's broker")
	}
}

func TestUnboundThenRebound(t *testing.T) { // AC-SRH-009
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=retired")

	first := rebindPrompt(t, root, "s1")
	for _, want := range []string{"factory lane unbound:", "lane-3", "runX"} {
		if !strings.Contains(first, want) {
			t.Fatalf("prompt 1 lacks %q: %q", want, first)
		}
	}
	if strings.Contains(first, "degraded") || len(relaunchLines(first)) != 0 {
		t.Fatalf("prompt 1 carries degraded or a command with no active run: %q", first)
	}
	noFactoryNotice(t, "prompt 2", rebindPrompt(t, root, "s1"))
	if got := rebindMarker(t, root, "s1").State; got != "unbound:runX" {
		t.Fatalf("marker state = %q, want unbound:runX", got)
	}

	recordActiveFactoryRun(t, root, "runY")
	third := rebindPrompt(t, root, "s1")
	if !strings.Contains(third, "factory lane rebound:") || !strings.Contains(third, "runY") {
		t.Fatalf("prompt 3 did not rebind once runY became active (unbound is not final): %q", third)
	}
	if got := laneRows(t, root, "runY"); got != "lane-3|s1|1" {
		t.Fatalf("runY peers = %q, want lane-3|s1|1", got)
	}
}

func TestRebindAmbiguousActiveRuns(t *testing.T) { // AC-SRH-010
	const verb = relaunchVerbPrefix + " --provider "
	t.Run("two candidates", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active", "runZ=active")
		first := rebindPrompt(t, root, "s1")
		want := []string{verb + "cc --lane lane-3 --run runY", verb + "cc --lane lane-3 --run runZ"}
		if got := relaunchLines(first); !reflect.DeepEqual(got, want) {
			t.Fatalf("command lines = %q, want %q\nnotice: %q", got, want, first)
		}
		if strings.Contains(first, "degraded") || !strings.Contains(first, "factory lane unbound:") {
			t.Fatalf("ambiguity notice malformed: %q", first)
		}
		if brokerExists(t, root, "runY") || brokerExists(t, root, "runZ") {
			t.Fatal("a peer was registered although two runs are active")
		}
		noFactoryNotice(t, "prompt 2", rebindPrompt(t, root, "s1"))
		if got := rebindMarker(t, root, "s1").State; got != "ambiguous:runY,runZ" {
			t.Fatalf("marker state = %q", got)
		}
		// A different candidate set is a different state: the notice fires again.
		recordActiveFactoryRun(t, root, "runW")
		if again := rebindPrompt(t, root, "s1"); len(relaunchLines(again)) != 3 {
			t.Fatalf("a changed candidate set stayed silent: %q", again)
		}
	})
	t.Run("four candidates are capped at three lines and a count", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runA=active", "runB=active", "runC=active", "runD=active")
		first := rebindPrompt(t, root, "s1")
		if got := relaunchLines(first); len(got) != 3 {
			t.Fatalf("command lines = %q, want three", got)
		}
		if !strings.Contains(first, "+1 more active factory runs") {
			t.Fatalf("no candidate count: %q", first)
		}
	})
	t.Run("codex prints the one bare line", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "gpt")
		root := rebindRoot(t, "runX=retired", "runY=active", "runZ=active")
		got := relaunchLines(rebindPrompt(t, root, "s1"))
		if want := []string{verb + "codex"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("command lines = %q, want %q", got, want)
		}
	})
}

func TestRebindEligibleSessionStartSilent(t *testing.T) { // AC-SRH-012
	cases := []struct {
		name string
		runs []string
	}{
		{"rebind-eligible: one other run active", []string{"runX=retired", "runY=active"}},
		{"ineligible: no active run", []string{"runX=retired"}},
		{"ineligible: two active runs", []string{"runX=retired", "runY=active", "runZ=active"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rebindEnv(t, "runX", "lane-3", "claude")
			root := rebindRoot(t, c.runs...)
			if got := registerFactorySessionStartPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}); got != "" {
				t.Fatalf("SessionStart said %q for a current-vocabulary lane on a not-active run, want nothing", got)
			}
			for _, run := range []string{"runX", "runY", "runZ"} {
				if brokerExists(t, root, run) {
					t.Fatalf("SessionStart created the broker of %s", run)
				}
			}
			if got := rebindMarker(t, root, "s1"); got != (factoryNoticeMarker{}) {
				t.Fatalf("SessionStart wrote a notice marker: %+v", got)
			}
		})
	}
}

// A leader is not a lane: on a not-active run it keeps today's degraded answer
// verbatim, on both bind surfaces, and never lists runs.
func TestLeaderOnNotActiveRunKeepsDegradedString(t *testing.T) { // REQ-SRH-009 preservation
	rebindEnv(t, "runX", "", "claude")
	root := rebindRoot(t, "runX=retired", "runY=active")
	lists := 0
	prev := factoryHookActiveRuns
	factoryHookActiveRuns = func(ctx context.Context, dbPath string) ([]string, error) { lists++; return prev(ctx, dbPath) }
	t.Cleanup(func() { factoryHookActiveRuns = prev })

	const want = "factory messaging degraded: NO_ACTIVE_FACTORY"
	if got := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, factoryPeerBindUserPrompt); got != want {
		t.Fatalf("leader prompt = %q, want %q", got, want)
	}
	if got := registerFactorySessionStartPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}); got != want {
		t.Fatalf("leader SessionStart = %q, want %q", got, want)
	}
	if lists != 0 {
		t.Fatalf("a leader listed active runs %d times", lists)
	}
}

func TestLegacyLabelNeverRebindsAndLiveOwnerNotDisplaced(t *testing.T) { // AC-SRH-011
	const verb = relaunchVerbPrefix + " --provider "
	t.Run("legacy label is never registered", func(t *testing.T) {
		rebindEnv(t, "runX", "worker-69", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		got := rebindPrompt(t, root, "s1")
		if want := []string{verb + "cc"}; !reflect.DeepEqual(relaunchLines(got), want) {
			t.Fatalf("legacy notice lines = %q, want %q (R8)", relaunchLines(got), want)
		}
		if lanes, _ := brokerLanes(t, root, "runY"); len(lanes) != 0 {
			t.Fatalf("a legacy session was registered into runY: %+v", lanes)
		}
	})
	t.Run("a live owner of the slot is not displaced", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		other := os.Getppid()
		otherStart, state := homestate.ProbeProcessIdentity(other)
		if state != homestate.ProcessIdentityLive || otherStart == "" {
			t.Fatalf("precondition: the parent process %d has no live identity (%v)", other, state)
		}
		s, err := factorymsg.Open(root, "runY")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.RegisterPeer(context.Background(), factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "runY", Backend: "claude", Role: "lane", Slot: "lane-3", SessionUUID: "other", Generation: 1, PID: other, ProcessStart: otherStart}); err != nil {
			t.Fatal(err)
		}
		closeFactoryHookStore(s)
		before, _ := brokerLanes(t, root, "runY")

		got := rebindPrompt(t, root, "s1")
		if !strings.Contains(got, "factory lane rebind refused:") {
			t.Fatalf("no refusal notice: %q", got)
		}
		if want := []string{verb + "cc --run runY"}; !reflect.DeepEqual(relaunchLines(got), want) {
			t.Fatalf("refusal lines = %q, want %q (R3)", relaunchLines(got), want)
		}
		if after, _ := brokerLanes(t, root, "runY"); !reflect.DeepEqual(before, after) {
			t.Fatalf("the existing peer row changed:\nbefore %+v\nafter  %+v", before, after)
		}
		noFactoryNotice(t, "prompt 2", rebindPrompt(t, root, "s1"))
		if got := rebindMarker(t, root, "s1").State; got != "refused:runY" {
			t.Fatalf("marker state = %q, want refused:runY", got)
		}
	})
}

func TestReboundClaimReadsRebindRunOnly(t *testing.T) { // AC-SRH-013
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=retired", "runY=active")
	owner, start := factoryHookOwnerIdentity(t)
	ctx := context.Background()
	key := homestate.ProjectKey(root)

	// runX's broker holds a pending message for session s1 (the environment run).
	sx, err := factorymsg.Open(root, "runX")
	if err != nil {
		t.Fatal(err)
	}
	defer closeFactoryHookStore(sx)
	laneX, err := sx.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: key, RunID: "runX", Backend: "claude", Role: "lane", Slot: "lane-3", SessionUUID: "s1", Generation: 1, PID: owner, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	leadX, err := sx.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: key, RunID: "runX", Backend: "claude", Role: "leader", Slot: "leader", SessionUUID: "leadX", Generation: 1, PID: owner, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	msgX, err := sx.Send(ctx, factorymsg.SendRequest{From: leadX, To: laneX, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "kx", TaskRef: "t1345", CorrelationID: "cx", TTL: time.Hour, Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}

	// The first prompt rebinds s1 into runY (empty inbox there).
	first := rebindPrompt(t, root, "s1")
	if !strings.Contains(first, "factory lane rebound:") {
		t.Fatalf("precondition: the first prompt did not rebind: %q", first)
	}
	if strings.Contains(first, msgX.ID) {
		t.Fatalf("the rebound claim read the environment run's broker: %q", first)
	}
	sy, err := factorymsg.Open(root, "runY")
	if err != nil {
		t.Fatal(err)
	}
	defer closeFactoryHookStore(sy)
	laneY, err := sy.Peer(ctx, "s1")
	if err != nil {
		t.Fatalf("precondition: s1 is not registered in runY: %v", err)
	}
	leadY, err := sy.RegisterPeer(ctx, factorymsg.Peer{ProjectKey: key, RunID: "runY", Backend: "claude", Role: "leader", Slot: "leader", SessionUUID: "leadY", Generation: 1, PID: owner, ProcessStart: start})
	if err != nil {
		t.Fatal(err)
	}
	msgY, err := sy.Send(ctx, factorymsg.SendRequest{From: leadY, To: laneY, Kind: factorymsg.KindStatusRequest, IdempotencyKey: "ky", TaskRef: "t1345", CorrelationID: "cy", TTL: time.Hour, Payload: []byte("y")})
	if err != nil {
		t.Fatal(err)
	}

	// The next prompt is silent about the rebind but its claim names runY and
	// lists only runY's message.
	second := rebindPrompt(t, root, "s1")
	if !strings.Contains(second, "Factory inbox run=runY") || !strings.Contains(second, msgY.ID) {
		t.Fatalf("the rebound claim does not list runY's message: %q", second)
	}
	if strings.Contains(second, "run=runX") || strings.Contains(second, msgX.ID) {
		t.Fatalf("the rebound claim leaked the environment run's message: %q", second)
	}

	// A Stop event reads the environment run's broker and rebinds nothing.
	peersBefore := laneRows(t, root, "runY")
	stopCtx, _, state := factoryHookBatch(ctx, &HookInput{SessionID: "s1", ProjectDir: root}, EventStop)
	if !strings.Contains(stopCtx, "Factory inbox run=runX") || !strings.Contains(stopCtx, msgX.ID) {
		t.Fatalf("Stop did not read the environment run's broker: state=%s ctx=%q", state, stopCtx)
	}
	if after := laneRows(t, root, "runY"); after != peersBefore {
		t.Fatalf("Stop changed runY's peers: %q -> %q", peersBefore, after)
	}
}

func TestFactoryHookBatchForRunOpensTheNamedRun(t *testing.T) { // REQ-SRH-008 unit
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	// An override naming a run with no broker is degraded, not a fall-back to the
	// environment run.
	_, _, state := factoryHookBatchForRun(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, EventUserPromptSubmit, "runNoBroker")
	if !strings.HasPrefix(state, "degraded:") {
		t.Fatalf("state = %q, want a degraded open of the named run's broker", state)
	}
	// An empty override is the environment run, exactly as factoryHookBatch.
	_, _, state = factoryHookBatchForRun(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, EventUserPromptSubmit, "")
	if state == "disabled" {
		t.Fatalf("an empty override must read the environment run, got state %q", state)
	}
}

// countingSeams wraps the three measurement seams with counters and restores
// them on cleanup.
type seamCounts struct{ resolve, probe, list int }

func countingSeams(t *testing.T) *seamCounts {
	t.Helper()
	c := &seamCounts{}
	prevPath, prevProbe, prevList := factoryHookDBPath, factoryHookProbeRun, factoryHookActiveRuns
	factoryHookDBPath = func(root string) (string, error) { c.resolve++; return prevPath(root) }
	factoryHookProbeRun = func(ctx context.Context, dbPath, run string) (factorymsg.RunState, string, error) {
		c.probe++
		return prevProbe(ctx, dbPath, run)
	}
	factoryHookActiveRuns = func(ctx context.Context, dbPath string) ([]string, error) { c.list++; return prevList(ctx, dbPath) }
	t.Cleanup(func() { factoryHookDBPath, factoryHookProbeRun, factoryHookActiveRuns = prevPath, prevProbe, prevList })
	return c
}

func TestHealthyLanePathUnchangedAndFailOpen(t *testing.T) { // AC-SRH-014, REQ-SRH-009/-010
	t.Run("healthy lane: same output, one resolution, one query, no listing", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=active", "runY=active")
		c := countingSeams(t)

		const bound = "factory messaging bound: run=runX slot=lane-3 generation=1; messages arrive at turn boundaries, not idle wake"
		if got := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, factoryPeerBindUserPrompt); got != bound {
			t.Fatalf("prompt output = %q, want %q", got, bound)
		}
		if *c != (seamCounts{resolve: 1, probe: 1, list: 0}) {
			t.Fatalf("prompt path counts = %+v, want one resolution, one query, no listing", *c)
		}
		*c = seamCounts{}
		if got := registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}, factoryPeerBindUserPrompt); got != "" {
			t.Fatalf("an already-bound prompt said %q", got)
		}
		if *c != (seamCounts{resolve: 1, probe: 1, list: 0}) {
			t.Fatalf("repeat prompt counts = %+v", *c)
		}
		*c = seamCounts{}
		if got := registerFactorySessionStartPeer(context.Background(), &HookInput{SessionID: "s1", ProjectDir: root}); got != "" {
			t.Fatalf("SessionStart said %q", got)
		}
		if *c != (seamCounts{resolve: 1, probe: 1, list: 0}) {
			t.Fatalf("SessionStart counts = %+v", *c)
		}
	})

	t.Run("unavailable measurement degrades without rebinding", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		prevProbe := factoryHookProbeRun
		factoryHookProbeRun = func(context.Context, string, string) (factorymsg.RunState, string, error) {
			return factorymsg.RunStateUnavailable, "", errors.New("measure boom")
		}
		t.Cleanup(func() { factoryHookProbeRun = prevProbe })
		got := rebindPrompt(t, root, "s1")
		if !strings.Contains(got, "factory messaging degraded: measure boom") || strings.Contains(got, "factory lane") {
			t.Fatalf("answer = %q, want the degraded string and no lane notice", got)
		}
		if brokerExists(t, root, "runY") {
			t.Fatal("a failed measurement registered a peer")
		}
	})

	t.Run("failed listing degrades without a notice or a marker", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		prevList := factoryHookActiveRuns
		factoryHookActiveRuns = func(context.Context, string) ([]string, error) { return nil, errors.New("list boom") }
		t.Cleanup(func() { factoryHookActiveRuns = prevList })
		got := rebindPrompt(t, root, "s1")
		if !strings.Contains(got, "factory messaging degraded: list boom") || strings.Contains(got, "factory lane") {
			t.Fatalf("answer = %q, want the degraded string and no lane notice", got)
		}
		if brokerExists(t, root, "runY") || rebindMarker(t, root, "s1") != (factoryNoticeMarker{}) {
			t.Fatal("a failed listing registered a peer or wrote a marker")
		}
	})

	t.Run("corrupt factory database stays degraded and writes no run record", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		dbPath, err := homestate.FactoryDBPath(root)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dbPath, []byte("this is not a sqlite database, only garbage bytes"), 0o600); err != nil {
			t.Fatal(err)
		}
		got := rebindPrompt(t, root, "s1")
		if !strings.Contains(got, "factory messaging degraded") || strings.Contains(got, "factory lane") {
			t.Fatalf("answer = %q, want the degraded string and no lane notice", got)
		}
		if brokerExists(t, root, "runY") {
			t.Fatal("a corrupt database registered a peer")
		}
	})

	t.Run("hook leaves the runs and events tables untouched", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		count := func() string {
			db, err := homestate.OpenFactory(root)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			var runs, events int
			if err := db.DB.QueryRow(`SELECT (SELECT count(*) FROM runs), (SELECT count(*) FROM events)`).Scan(&runs, &events); err != nil {
				t.Fatal(err)
			}
			return fmt.Sprintf("runs=%d events=%d", runs, events)
		}
		before := count()
		rebindPrompt(t, root, "s1")
		if after := count(); after != before {
			t.Fatalf("the hook wrote run records: %s -> %s", before, after)
		}
	})
}

func TestRebindPathStaysInsideBindBudget(t *testing.T) { // AC-SRH-014, REQ-SRH-010
	t.Run("production bind budget", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		// The production bind budget, not the generous fixture pin.
		factoryBindBudget = 2 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), factoryBindBudget)
		defer cancel()
		started := time.Now()
		got := registerFactoryHookPeer(ctx, &HookInput{SessionID: "s1", ProjectDir: root}, factoryPeerBindUserPrompt)
		elapsed := time.Since(started)
		t.Logf("rebind path elapsed %v against the %v bind budget", elapsed, factoryBindBudget)
		if !strings.Contains(got, "factory lane rebound:") {
			t.Fatalf("the rebind did not complete inside the production bind budget: %q (elapsed %v)", got, elapsed)
		}
		if elapsed >= factoryBindBudget {
			t.Fatalf("rebind path took %v, not below the %v bind budget", elapsed, factoryBindBudget)
		}
	})
	t.Run("one-nanosecond budget fails open", func(t *testing.T) {
		rebindEnv(t, "runX", "lane-3", "claude")
		root := rebindRoot(t, "runX=retired", "runY=active")
		factoryBindBudget = time.Nanosecond
		got := rebindPrompt(t, root, "s1") // Handle returns no error even with the budget spent
		if got != "" && !strings.Contains(got, "factory messaging degraded") {
			t.Fatalf("answer = %q, want silence or the degraded string", got)
		}
		if strings.Contains(got, "factory lane") {
			t.Fatalf("a spent budget produced a lane notice: %q", got)
		}
		if brokerExists(t, root, "runY") || rebindMarker(t, root, "s1") != (factoryNoticeMarker{}) {
			t.Fatal("a spent budget registered a peer or wrote a marker")
		}
	})
}

// A busy broker must not replace the hook caller's budget with Open's default.
func TestRebindBrokerInitializationHonorsCallerBudget(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=retired", "runY=active")
	path, err := factorymsg.BrokerPath(root, "runY")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	lock, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = lock.Exec("ROLLBACK")
		_ = lock.Close()
	})
	if _, err := lock.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatal(err)
	}
	if _, err := lock.Exec("BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	notice, run := registerReboundLane(ctx, laneRebindRequest{root: root}, "runY", "claude")
	elapsed := time.Since(started)
	t.Logf("locked initialization elapsed=%v notice=%q", elapsed, notice)
	if run != "" || !strings.Contains(notice, "factory messaging degraded:") {
		t.Fatalf("unexpected registration run=%q notice=%q", run, notice)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("broker initialization exceeded caller budget allowance: %v", elapsed)
	}
	var peers int
	if err := lock.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='peers'").Scan(&peers); err != nil {
		t.Fatal(err)
	}
	if peers != 0 {
		t.Fatalf("failed initialization registered %d peers", peers)
	}
	if _, err := lock.Exec("ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	retryCtx, retryCancel := context.WithTimeout(context.Background(), time.Second)
	defer retryCancel()
	retry, err := factorymsg.OpenWithContext(retryCtx, root, "runY")
	if err != nil {
		t.Fatalf("initialization was not retryable after releasing the lock: %v", err)
	}
	if err := retry.Close(); err != nil {
		t.Fatal(err)
	}
	if retry.HandleStats().OpenConnections != 0 {
		t.Fatal("closed retry retained broker connections")
	}
}

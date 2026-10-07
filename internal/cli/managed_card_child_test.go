package cli

// The launcher now opens one parent session. These fixtures exercise the
// retained managed owner directly, independently of the retired boot loop.
import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const cardChildWatchdog = 90 * time.Second

func strPtr(s string) *string { return &s }

type cardChildCall struct{ dir string }
type cardChildOpts struct {
	cards     int
	realOwner bool
	source    io.Reader
}
type cardChildLoop struct {
	t       *testing.T
	root    string
	managed []cardChildCall
	cards   int
	logPath string
}

func newCardChildLoop(t *testing.T, opts cardChildOpts) *cardChildLoop {
	t.Helper()
	root, store := fcFixture(t)
	sdScrubLauncherEnv(t)
	states := make([]factory.BacklogState, opts.cards)
	for i := range states {
		states[i] = factory.BacklogStatePicked
	}
	fcQueue(t, store, states...)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	t.Setenv(config.EnvFactoryRunID, fcRun)
	h := &cardChildLoop{t: t, root: root, cards: opts.cards}
	prevLook, prevSource := codexLookPath, managedLaneOperatorSource
	t.Cleanup(func() {
		endManagedLanePump()
		codexLookPath, managedLaneOperatorSource = prevLook, prevSource
	})
	if opts.source != nil {
		managedLaneOperatorSource = opts.source
	}
	if opts.realOwner {
		if runtime.GOOS == "windows" {
			t.Skip("POSIX sh fixture backend")
		}
		h.logPath = filepath.Join(t.TempDir(), "card-child.log")
		t.Setenv(fakeAppServerRoleEnv, "appserver")
		t.Setenv(fakeAppServerLogEnv, h.logPath)
		bin := fakeAppServerScript(t)
		codexLookPath = func(string) (string, error) { return bin, nil }
	}
	return h
}

// run supplies a fixed number of explicit sessions to the retained owner.
// Queue consumption and lane boot are tested separately.
func (h *cardChildLoop) run() (string, string, error) {
	t := h.t
	t.Helper()
	restoreRun, err := enterFactoryLaneRun(h.root, fcRun, "", nil)
	if err != nil {
		return "", "", err
	}
	defer restoreRun()
	restoreLane := enterFactoryLaneMode(factory.FactoryLaneLabel(1), 0, "", config.FactoryDispatchAuto)
	defer restoreLane()
	restoreBackend := exportFactoryLaunchFacts("", factory.BackendGPT)
	defer restoreBackend()
	defer endManagedLanePump()
	bin, err := codexLookPath(codexBinaryName)
	if err != nil {
		return "", "", err
	}
	var stderr strings.Builder
	for i := 0; i < h.cards; i++ {
		cardID := fmt.Sprintf("t%d", i+1)
		wt := t.TempDir()
		env := append(codexChildEnv(),
			config.EnvFactoryRole+"="+config.FactoryRoleLane,
			config.EnvMoaiFactoryWorker+"="+factory.FactoryLaneLabel(1),
			config.EnvFactoryBackend+"="+factory.BackendGPT,
			config.EnvFactoryCard+"="+cardID,
			config.EnvFactoryRunID+"="+fcRun)
		h.managed = append(h.managed, cardChildCall{dir: wt})
		if err := defaultManagedCodexCardLaunch(bin, []string{bin}, env, wt); err != nil {
			fmt.Fprintf(&stderr, "managed card %s: %v\n", cardID, err)
		}
	}
	return "", stderr.String(), nil
}

// The retained owner must keep refusing a CLI cwd override: its launch
// directory is supplied explicitly to the App Server owner.
func TestManagedCardChildOwnerRefusesCwdOverride(t *testing.T) {
	_, _, err := managedCodexOptions([]string{"codex", "-C", "/x"})
	if err == nil || !strings.Contains(err.Error(), "-C") {
		t.Fatalf("managedCodexOptions(-C) error = %v, want a refusal naming -C", err)
	}
}

// brokerLanes reads the run's lane roster through a fresh broker handle.
func brokerLanes(t *testing.T, root string) []factorymsg.LaneStatus {
	t.Helper()
	store, err := factorymsg.Open(root, fcRun)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()
	roster, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return roster.Lanes
}

func fakeLogCount(t *testing.T, logPath, line string) int {
	t.Helper()
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("fake app server log: %v", err)
	}
	n := 0
	for _, l := range strings.Split(string(raw), "\n") {
		if l == line {
			n++
		}
	}
	return n
}

func runCardChildLoopBounded(t *testing.T, h *cardChildLoop) error {
	t.Helper()
	errCh := make(chan error, 1)
	go func() {
		_, _, err := h.run()
		errCh <- err
	}()
	select {
	case err := <-errCh:
		return err
	case <-time.After(cardChildWatchdog):
		t.Fatalf("lane loop did not finish within %s", cardChildWatchdog)
		return nil
	}
}

// AC-CC-007 — with the lane seam's DEFAULT body over the real broker store and
// a fake App Server, two successive card sessions of one launcher process
// register, bind and replace the lane endpoint; a failed start leaves no row.
func TestManagedCardChildSecondCardRebinds(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	t.Run("sequential_cards_rebind", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 2, realOwner: true, source: strings.NewReader("/exit\n/exit\n")})
		if err := runCardChildLoopBounded(t, h); err != nil {
			t.Fatalf("codex lane: %v", err)
		}
		lanes := brokerLanes(t, h.root)
		if len(lanes) != 1 {
			t.Fatalf("roster lanes = %+v, want exactly one lane endpoint", lanes)
		}
		if lanes[0].BindingState != factorymsg.BindingBound || lanes[0].SessionUUID != fakeAppServerThreadID {
			t.Errorf("endpoint binding = %q session %q, want %q/%q", lanes[0].BindingState, lanes[0].SessionUUID, factorymsg.BindingBound, fakeAppServerThreadID)
		}
		if got := fakeLogCount(t, h.logPath, "method thread/start"); got != 2 {
			t.Errorf("thread/start logged %d times, want 2 (one per card)", got)
		}
	})
	t.Run("start_failure_leaves_no_pending_row", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 2, realOwner: true, source: strings.NewReader("")})
		codexLookPath = func(string) (string, error) { return filepath.Join(t.TempDir(), "no-such-codex"), nil }
		if err := runCardChildLoopBounded(t, h); err != nil {
			t.Fatalf("codex lane: %v (a failed session must not stop the loop)", err)
		}
		if len(h.managed) != 2 {
			t.Errorf("owner launches = %d, want 2 (both cards attempted)", len(h.managed))
		}
		if lanes := brokerLanes(t, h.root); len(lanes) != 0 {
			t.Errorf("roster lanes after failed starts = %+v, want none", lanes)
		}
	})
	t.Run("positive_control_pending_row_visible", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 1})
		start := homestate.CurrentProcessFingerprint()
		env := []string{
			config.EnvFactoryRunID + "=" + fcRun,
			config.EnvFactoryBackend + "=" + factory.BackendGPT,
			config.EnvMoaiFactoryWorker + "=" + factory.FactoryLaneLabel(1),
		}
		if _, err := registerFactoryLaunchPending(context.Background(), h.root, env, os.Getpid(), start); err != nil {
			t.Fatal(err)
		}
		if lanes := brokerLanes(t, h.root); len(lanes) != 1 {
			t.Fatalf("roster lanes right after launch-pending registration = %+v, want one pending row", lanes)
		}
	})
}

// AC-CC-008 — a broker message sent to the lane is delivered by the same owner
// code through the loop: the fake App Server sees a second turn/start after
// the priming turn.
func TestManagedCardChildDeliversInboxThroughLoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	pr, pw := io.Pipe()
	t.Cleanup(func() { _ = pw.Close() })
	h := newCardChildLoop(t, cardChildOpts{cards: 1, realOwner: true, source: pr})
	errCh := make(chan error, 1)
	go func() {
		_, _, err := h.run()
		errCh <- err
	}()

	var lane factorymsg.LaneStatus
	deadline := time.Now().Add(cardChildWatchdog)
	for time.Now().Before(deadline) {
		if lanes := brokerLanes(t, h.root); len(lanes) == 1 && lanes[0].BindingState == factorymsg.BindingBound {
			lane = lanes[0]
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if lane.Slot == "" {
		t.Fatal("the lane endpoint never bound within the watchdog")
	}

	store, err := factorymsg.Open(h.root, fcRun)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()
	senderStart, state := homestate.ProbeProcessIdentity(os.Getppid())
	if state != homestate.ProcessIdentityLive || senderStart == "" {
		t.Skip("test parent process identity unavailable")
	}
	to := factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(h.root), RunID: fcRun, Backend: lane.Backend, Role: lane.Role,
		Slot: lane.Slot, SessionUUID: lane.SessionUUID, Generation: lane.Generation, PID: lane.PID, ProcessStart: lane.ProcessStart,
	}
	from, err := store.RegisterPeer(context.Background(), factorymsg.Peer{
		ProjectKey: to.ProjectKey, RunID: fcRun, Backend: factory.BackendClaude, Role: factory.RoleLeader, Slot: factory.RoleLeader,
		SessionUUID: "card-child-leader", Generation: 1, PID: os.Getppid(), ProcessStart: senderStart,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Send(context.Background(), factorymsg.SendRequest{
		From: from, To: to, Kind: factorymsg.KindStatusRequest,
		IdempotencyKey: "card-child-once", TaskRef: "t1", CorrelationID: "c1", TTL: time.Minute, Payload: []byte("status please"),
	}); err != nil {
		t.Fatal(err)
	}

	delivered := false
	for end := time.Now().Add(10 * time.Second); time.Now().Before(end); time.Sleep(100 * time.Millisecond) {
		if fakeLogCount(t, h.logPath, "method turn/start") >= 2 {
			delivered = true
			break
		}
	}
	if _, werr := pw.Write([]byte("/exit\n")); werr != nil {
		t.Logf("write /exit: %v", werr)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("codex lane: %v", err)
		}
	case <-time.After(cardChildWatchdog):
		t.Fatal("lane loop did not finish after /exit")
	}
	if !delivered {
		t.Error("the fake App Server never saw a second turn/start (the inbox message was not delivered through the loop)")
	}
}

// AC-CC-010 (loop level) — over the real owner, one chunk carrying two end
// tokens ends the first card's session on the first and the second card's on
// the second: input after the ending line goes to the next session.
func TestManagedCardChildOperatorInputReachesNextSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	for _, tc := range []struct{ name, chunk string }{
		{"exit_exit", "/exit\n/exit\n"},
		{"quit_quit", "/quit\n/quit\n"},
		{"exit_quit", "/exit\n/quit\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newCardChildLoop(t, cardChildOpts{cards: 2, realOwner: true, source: strings.NewReader(tc.chunk)})
			if err := runCardChildLoopBounded(t, h); err != nil {
				t.Fatalf("codex lane: %v", err)
			}
			if len(h.managed) != 2 {
				t.Errorf("owner launches = %d, want 2", len(h.managed))
			}
			if got := fakeLogCount(t, h.logPath, "method thread/start"); got != 2 {
				t.Errorf("thread/start logged %d times, want 2", got)
			}
		})
	}
}

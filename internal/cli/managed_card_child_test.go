package cli

// managed_card_child_test.go — SPEC-FACTORY-MANAGED-CARD-CHILD-001 (card
// t1440): with the managed opt-in on, `moai codex -f lane` starts each leased
// card's child session through the managed Codex owner; with it off, the card
// child keeps the direct door byte for byte.
//
// Every test here builds its own fixture and must run under the lane
// environment scrub (see acceptance.md): sdScrubLauncherEnv sits BEFORE the
// queue is touched, because a lane-stamped process refuses queue mutation.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	cardChildSentinelBin = "/sentinel/codex"
	cardChildLocalBody   = "local rule\n"
	// 90s: the 30s binding watchdog expired on CI (run 37316159953), where the
	// -race shard ran this loop under full-shard load — the endpoint bound
	// locally in 7.3s and on main's run, but not within 30.75s there. 3x the
	// observed breach point, still far under the job's 35m -timeout.
	cardChildWatchdog = 90 * time.Second
)

// cardChildCall is one launch observed at the managed lane seam.
type cardChildCall struct {
	bin  string
	args []string
	env  []string
	dir  string
}

// cardChildOpts shapes one lane-loop fixture.
type cardChildOpts struct {
	cards int
	// managed is the MOAI_FACTORY_MANAGED value; nil leaves the variable unset.
	managed *string
	// localInstruction, when set, is committed as AGENTS.local.md so every card
	// worktree carries it.
	localInstruction string
	// realOwner keeps the lane seam's DEFAULT body (wrapped only to move the
	// card forward after each session) over a fake App Server binary.
	realOwner bool
	// noWork leaves the card where it is after a real-owner session.
	noWork bool
	// source feeds the lane loop's operator input (real-owner fixtures).
	source io.Reader
}

// cardChildLoop is one `moai codex -f lane` fixture with every launch door
// replaced by a recorder.
type cardChildLoop struct {
	t       *testing.T
	root    string
	managed []cardChildCall
	plain   atomic.Int64
	direct  []*exec.Cmd
	// managedResult and directResult decide what the Nth launch returns (nil
	// means success). The substituted session moves its own card forward first,
	// the way the real session's work would.
	managedResult func(i int) error
	directResult  func(i int) error
	// anchorCalls counts worktree anchor-lock placements.
	anchorCalls atomic.Int64
	// logPath is the fake App Server's evidence log (real-owner fixtures).
	logPath string
}

func strPtr(s string) *string { return &s }

// setManagedSwitch sets or clears the opt-in for the test.
func setManagedSwitch(t *testing.T, value *string) {
	t.Helper()
	t.Setenv(config.EnvMoaiFactoryManaged, "x")
	if value == nil {
		_ = os.Unsetenv(config.EnvMoaiFactoryManaged)
		return
	}
	t.Setenv(config.EnvMoaiFactoryManaged, *value)
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
	if opts.localInstruction != "" {
		if err := os.WriteFile(filepath.Join(root, "AGENTS.local.md"), []byte(opts.localInstruction), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, root, "add", "AGENTS.local.md")
		fcGit(t, root, "commit", "-q", "-m", "local instructions")
	}
	t.Chdir(root)
	setManagedSwitch(t, opts.managed)
	t.Setenv(config.EnvFactoryRunID, fcRun)
	t.Setenv(config.EnvMoaiFactoryWorkers, "1")

	h := &cardChildLoop{t: t, root: root}
	bin := cardChildSentinelBin
	if opts.realOwner {
		if runtime.GOOS == "windows" {
			t.Skip("POSIX sh fixture backend")
		}
		h.logPath = filepath.Join(t.TempDir(), "card-child.log")
		t.Setenv(fakeAppServerRoleEnv, "appserver")
		t.Setenv(fakeAppServerLogEnv, h.logPath)
		bin = fakeAppServerScript(t)
	}

	prevLook, prevDirect, prevPlain, prevCard := codexLookPath, codexDirectLaunchFn, managedFactoryCodexLaunchFunc, managedCodexCardLaunchFunc
	prevAnchor, prevSource := codexWorktreeAnchorLock, managedLaneOperatorSource
	t.Cleanup(func() {
		codexLookPath, codexDirectLaunchFn, managedFactoryCodexLaunchFunc, managedCodexCardLaunchFunc = prevLook, prevDirect, prevPlain, prevCard
		codexWorktreeAnchorLock, managedLaneOperatorSource = prevAnchor, prevSource
		endManagedLanePump() // a skip or fatal path may leave the process-global pump set
	})
	codexLookPath = func(string) (string, error) { return bin, nil }
	codexWorktreeAnchorLock = func(string, int, string) error { h.anchorCalls.Add(1); return nil }
	if opts.source != nil {
		managedLaneOperatorSource = opts.source
	}
	managedFactoryCodexLaunchFunc = func(string, []string, []string, string) error { h.plain.Add(1); return nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		i := len(h.direct)
		h.direct = append(h.direct, c)
		sdCodexSessionWork(t, root, sdEnvOf(t, c.Env)[config.EnvFactoryCard])
		if h.directResult != nil {
			return h.directResult(i)
		}
		return nil
	}
	if opts.realOwner {
		orig := managedCodexCardLaunchFunc
		managedCodexCardLaunchFunc = func(bin string, args, env []string, dir string) error {
			h.managed = append(h.managed, cardChildCall{bin: bin, args: args, env: env, dir: dir})
			err := orig(bin, args, env, dir)
			if !opts.noWork {
				sdCodexSessionWork(t, root, launchEnvValue(env, config.EnvFactoryCard))
			}
			return err
		}
		return h
	}
	managedCodexCardLaunchFunc = func(bin string, args, env []string, dir string) error {
		i := len(h.managed)
		h.managed = append(h.managed, cardChildCall{
			bin: bin, args: append([]string(nil), args...), env: append([]string(nil), env...), dir: dir,
		})
		sdCodexSessionWork(t, root, launchEnvValue(env, config.EnvFactoryCard))
		if h.managedResult != nil {
			return h.managedResult(i)
		}
		return nil
	}
	return h
}

// run drives `moai codex -f lane` and returns stdout, stderr and the error.
func (h *cardChildLoop) run() (string, string, error) {
	h.t.Helper()
	return runCodexCmd(h.t, "-l")
}

// AC-CC-001 — switch on + lane stamps: each leased card's child goes through
// the lane loop's managed seam exactly once; the plain divert seam and the
// direct door are never reached.
func TestManagedCardChildLaneLoopUsesManagedOwner(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"switch_1", "1"},
		{"switch_TRUE_padded", " TRUE "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr(tc.value)})
			if _, _, err := h.run(); err != nil {
				t.Fatalf("codex lane: %v", err)
			}
			if len(h.managed) != 2 || h.plain.Load() != 0 || len(h.direct) != 0 {
				t.Fatalf("lane seam calls=%d plain seam calls=%d direct door calls=%d, want 2/0/0",
					len(h.managed), h.plain.Load(), len(h.direct))
			}
			for i, id := range []string{"t1", "t2"} {
				card := fcCard(t, h.root, id)
				if card.WorktreePath == "" || h.managed[i].dir != card.WorktreePath {
					t.Errorf("call %d: dir = %q, want card %s's worktree %q", i, h.managed[i].dir, id, card.WorktreePath)
				}
				if got := launchEnvValue(h.managed[i].env, config.EnvFactoryCard); got != id {
					t.Errorf("call %d: %s = %q, want %q", i, config.EnvFactoryCard, got, id)
				}
			}
		})
	}
}

// AC-CC-002 — switch off (unset, empty, 0, yes): the direct door receives the
// same *exec.Cmd as before this SPEC, compared to LITERAL expectations over a
// fully controlled process environment, and neither managed seam nor the
// operator-input pump is touched.
func TestManagedCardChildSwitchOffKeepsDirectDoor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *string
	}{
		{"unset", nil},
		{"empty", strPtr("")},
		{"zero", strPtr("0")},
		{"yes", strPtr("yes")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: tc.value, localInstruction: cardChildLocalBody})
			reads := &countingReader{}
			managedLaneOperatorSource = reads
			pumpsBefore := managedOperatorPumpsCreated.Load()
			label := sdNextFreeLaneLabel(t, h.root)
			path, gitCfg := os.Getenv("PATH"), os.Getenv("GIT_CONFIG_GLOBAL")
			codexHome := t.TempDir()
			inherited := []string{
				"PATH=" + path,
				"HOME=" + t.TempDir(),
				"GIT_CONFIG_GLOBAL=" + gitCfg,
				"GIT_CONFIG_NOSYSTEM=1",
				config.EnvClaudeProjectDir + "=" + h.root,
				config.EnvHome + "=",
				"CODEX_HOME=" + codexHome,
			}
			if tc.value != nil {
				inherited = append(inherited, config.EnvMoaiFactoryManaged+"="+*tc.value)
			}
			controlProcessEnv(t, inherited...)

			if _, _, err := h.run(); err != nil {
				t.Fatalf("codex lane: %v", err)
			}
			if len(h.direct) != 2 || len(h.managed) != 0 || h.plain.Load() != 0 {
				t.Fatalf("direct door calls=%d lane seam calls=%d plain seam calls=%d, want 2/0/0",
					len(h.direct), len(h.managed), h.plain.Load())
			}
			for i, id := range []string{"t1", "t2"} {
				wt := fcCard(t, h.root, id).WorktreePath
				c := h.direct[i]
				if c.Path != cardChildSentinelBin {
					t.Errorf("card %s: Path = %q, want %q", id, c.Path, cardChildSentinelBin)
				}
				wantArgs := []string{cardChildSentinelBin, "-C", wt, "-c", "developer_instructions=" + cardChildLocalInstructionLiteral()}
				if !equalStrings(c.Args, wantArgs) {
					t.Errorf("card %s: Args = %q, want %q", id, c.Args, wantArgs)
				}
				if c.Dir != wt {
					t.Errorf("card %s: Dir = %q, want %q", id, c.Dir, wt)
				}
				// The loop publishes these on top of the controlled inherited
				// environment; the child sees them (they are not lane keys).
				wantEnv := append([]string(nil), inherited...)
				wantEnv = append(wantEnv,
					config.EnvAutonomyTier+"=fully-autonomous",
					config.EnvClaudeCodeMaxConcurrentSubagents+"=10",
					config.EnvFactoryClearPolicy+"=",
					config.EnvFactoryAutoDispatch+"=auto",
					"CODEX_HOME="+codexHome,
					config.EnvFactoryRole+"=lane",
					config.EnvMoaiFactoryWorker+"="+label,
					config.EnvFactoryBackend+"=gpt",
					config.EnvFactoryCard+"="+id,
				)
				if !equalStrings(c.Env, wantEnv) {
					t.Errorf("card %s: Env = %q\nwant %q", id, c.Env, wantEnv)
				}
				if c.Stdin != io.Reader(os.Stdin) || c.Stdout != io.Writer(os.Stdout) || c.Stderr != io.Writer(os.Stderr) {
					t.Errorf("card %s: stdio is not the parent's own (stdin %v stdout %v stderr %v)", id, c.Stdin, c.Stdout, c.Stderr)
				}
			}
			if reads.n.Load() != 0 {
				t.Errorf("operator-input source read %d times on the switch-off path, want 0", reads.n.Load())
			}
			if got := managedOperatorPumpsCreated.Load() - pumpsBefore; got != 0 {
				t.Errorf("%d operator-input pump(s) created on the switch-off path, want 0", got)
			}
		})
	}
}

// cardChildLocalInstructionLiteral is the JSON string the local-instruction
// producer must emit for the AGENTS.local.md fixture, written out by hand: the
// source header, a newline, the file body — JSON-encoded the way encoding/json
// does (angle brackets as unicode escapes).
func cardChildLocalInstructionLiteral() string {
	const bs = "\\"
	return `"` + bs + "u003c!-- source: AGENTS.local.md --" + bs + "u003e" + bs + "n" + "local rule" + bs + "n" + `"`
}

type countingReader struct{ n atomic.Int64 }

func (r *countingReader) Read([]byte) (int, error) {
	r.n.Add(1)
	return 0, io.EOF
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// controlProcessEnv replaces the whole process environment with pairs, in
// order, for the rest of the test; cleanup restores the original.
func controlProcessEnv(t *testing.T, pairs ...string) {
	t.Helper()
	saved := os.Environ()
	os.Clearenv()
	for _, kv := range pairs {
		k, v, _ := strings.Cut(kv, "=")
		_ = os.Setenv(k, v)
	}
	t.Cleanup(func() {
		os.Clearenv()
		for _, kv := range saved {
			k, v, _ := strings.Cut(kv, "=")
			_ = os.Setenv(k, v)
		}
	})
}

// AC-CC-004 — the managed card-child argv carries no -C (the owner refuses it),
// the local developer-instruction pair is the producer's, and the owner is
// launched in the card worktree.
func TestManagedCardChildLaunchShape(t *testing.T) {
	t.Run("no_dash_C_and_pair_and_dir", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 1, managed: strPtr("1"), localInstruction: cardChildLocalBody})
		if _, _, err := h.run(); err != nil {
			t.Fatalf("codex lane: %v", err)
		}
		if len(h.managed) != 1 {
			t.Fatalf("lane seam calls=%d, want 1", len(h.managed))
		}
		call := h.managed[0]
		wt := fcCard(t, h.root, "t1").WorktreePath
		if len(call.args) == 0 || call.args[0] != cardChildSentinelBin || call.bin != cardChildSentinelBin {
			t.Errorf("bin %q args[0] %v, want the codexLookPath binary %q", call.bin, call.args, cardChildSentinelBin)
		}
		for _, a := range call.args {
			if a == "-C" {
				t.Errorf("argv %q carries -C; the managed owner refuses it", call.args)
			}
		}
		pair, err := codexLocalDeveloperInstructionArgs(wt)
		if err != nil || len(pair) != 2 {
			t.Fatalf("local instruction pair for %s = %v, %v", wt, pair, err)
		}
		if !equalStrings(call.args, []string{cardChildSentinelBin, pair[0], pair[1]}) {
			t.Errorf("argv = %q, want [bin %q %q]", call.args, pair[0], pair[1])
		}
		if call.dir != wt {
			t.Errorf("dir = %q, want the card worktree %q", call.dir, wt)
		}
	})
	t.Run("oversize_instruction_skips_owner", func(t *testing.T) {
		huge := strings.Repeat("x", config.DefaultCodexInstructionArgBytes)
		h := newCardChildLoop(t, cardChildOpts{cards: 1, managed: strPtr("1"), localInstruction: huge})
		_, stderr, err := h.run()
		if err != nil {
			t.Fatalf("codex lane: %v (a failed session must not stop the loop)", err)
		}
		if len(h.managed) != 0 {
			t.Errorf("owner launched %d times despite an oversize instruction token", len(h.managed))
		}
		if !strings.Contains(stderr, "codex lane: card t1 session:") {
			t.Errorf("stderr lacks the session error line:\n%s", stderr)
		}
	})
	t.Run("owner_refuses_dash_C", func(t *testing.T) {
		_, _, err := managedCodexOptions([]string{cardChildSentinelBin, "-C", "/x"})
		if err == nil || !strings.Contains(err.Error(), "-C") {
			t.Fatalf("managedCodexOptions(-C) error = %v, want a refusal naming -C", err)
		}
	})
}

// AC-CC-005 — the managed child's environment carries the card child's
// identity plus the run id the owner requires, and no lane-count stamp.
func TestManagedCardChildEnvCarriesIdentity(t *testing.T) {
	t.Setenv(config.EnvClaudeCodeSessionID, "foreign-claude-session")
	h := newCardChildLoop(t, cardChildOpts{cards: 1, managed: strPtr("1")})
	label := sdNextFreeLaneLabel(t, h.root)
	if _, _, err := h.run(); err != nil {
		t.Fatalf("codex lane: %v", err)
	}
	if len(h.managed) != 1 {
		t.Fatalf("lane seam calls=%d, want 1", len(h.managed))
	}
	env := h.managed[0].env
	for key, want := range map[string]string{
		config.EnvFactoryRole:       config.FactoryRoleLane,
		config.EnvMoaiFactoryWorker: label,
		config.EnvFactoryBackend:    factory.BackendGPT,
		config.EnvFactoryCard:       "t1",
		config.EnvFactoryRunID:      fcRun,
	} {
		if got := launchEnvValue(env, key); got != want {
			t.Errorf("env %s = %q, want %q", key, got, want)
		}
	}
	for _, key := range []string{config.EnvMoaiFactoryWorkers, config.EnvClaudeCodeSessionID} {
		if envHasKey(env, key) {
			t.Errorf("env carries %s; the child scrub removes it", key)
		}
	}
}

// AC-CC-006 — the managed card-child launch places no worktree anchor lock and
// leaves the lane claim on the launcher's own pid, before and after.
func TestManagedCardChildKeepsLauncherClaim(t *testing.T) {
	h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("1")})
	var claimed []int
	orig := managedCodexCardLaunchFunc
	managedCodexCardLaunchFunc = func(bin string, args, env []string, dir string) error {
		entry := loadFactoryRegistry(factoryRegistryPath(h.root))[launchEnvValue(env, config.EnvMoaiFactoryWorker)]
		claimed = append(claimed, entry.PID)
		return orig(bin, args, env, dir)
	}
	if _, _, err := h.run(); err != nil {
		t.Fatalf("codex lane: %v", err)
	}
	if h.anchorCalls.Load() != 0 {
		t.Errorf("anchor lock placed %d times on the managed card-child path, want 0", h.anchorCalls.Load())
	}
	if len(claimed) != 2 {
		t.Fatalf("claim observed at %d owner launches, want 2", len(claimed))
	}
	for i, pid := range claimed {
		if pid != os.Getpid() {
			t.Errorf("launch %d: lane claim pid = %d, want the launcher's own %d", i, pid, os.Getpid())
		}
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
		h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("1"), realOwner: true, source: strings.NewReader("/exit\n/exit\n")})
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
		h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("1"), realOwner: true, source: strings.NewReader("")})
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
		h := newCardChildLoop(t, cardChildOpts{cards: 1, managed: strPtr("1")})
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
	h := newCardChildLoop(t, cardChildOpts{cards: 1, managed: strPtr("1"), realOwner: true, noWork: true, source: pr})
	done := make(chan cardChildResult, 1)
	go func() {
		stdout, stderr, err := h.run()
		done <- cardChildResult{stdout: stdout, stderr: stderr, err: err}
	}()

	var lane factorymsg.LaneStatus
	deadline := time.Now().Add(cardChildWatchdog)
	for time.Now().Before(deadline) {
		if err := cardChildEarlyExit(done); err != nil {
			t.Fatal(err)
		}
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
	case result := <-done:
		if result.err != nil {
			t.Errorf("codex lane: %v; stdout=%q stderr=%q", result.err, result.stdout, result.stderr)
		}
	case <-time.After(cardChildWatchdog):
		t.Fatal("lane loop did not finish after /exit")
	}
	if !delivered {
		t.Error("the fake App Server never saw a second turn/start (the inbox message was not delivered through the loop)")
	}
}

func TestManagedCardChildEarlyExitReportsLaunchEvidence(t *testing.T) {
	for _, launchErr := range []error{nil, errors.New("fixture startup refused")} {
		done := make(chan cardChildResult, 1)
		done <- cardChildResult{stdout: "fixture stdout", stderr: "fixture stderr", err: launchErr}
		err := cardChildEarlyExit(done)
		if err == nil || !strings.Contains(err.Error(), "fixture stdout") || !strings.Contains(err.Error(), "fixture stderr") {
			t.Fatalf("early completion = %v, want launch output even when the loop returned nil", err)
		}
		if launchErr != nil && !errors.Is(err, launchErr) {
			t.Fatalf("early completion = %v, want wrapped launch error %v", err, launchErr)
		}
	}
}

type cardChildResult struct {
	stdout, stderr string
	err            error
}

func cardChildEarlyExit(done <-chan cardChildResult) error {
	select {
	case result := <-done:
		return errors.Join(fmt.Errorf("lane loop ended before binding; stdout=%q stderr=%q", result.stdout, result.stderr), result.err)
	default:
		return nil
	}
}

// AC-CC-009 — a managed session's end, normal or with an error, lets the loop
// continue; an error is one stderr line in the same form the direct door uses.
func TestManagedCardChildSessionEndContinuesLoop(t *testing.T) {
	t.Run("owner_error_logged_and_loop_continues", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("1")})
		h.managedResult = func(i int) error {
			if i == 0 {
				return errors.New("boom")
			}
			return nil
		}
		_, stderr, err := h.run()
		if err != nil {
			t.Fatalf("codex lane: %v", err)
		}
		if len(h.managed) != 2 {
			t.Errorf("owner launches = %d, want 2 (the loop must not stop at the first error)", len(h.managed))
		}
		if got := strings.Count(stderr, "codex lane: card t1 session: boom\n"); got != 1 {
			t.Errorf("session error line count = %d, want 1; stderr:\n%s", got, stderr)
		}
	})
	t.Run("same_line_form_as_direct_door", func(t *testing.T) {
		h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("0")})
		h.directResult = func(i int) error {
			if i == 0 {
				return errors.New("boom")
			}
			return nil
		}
		_, stderr, err := h.run()
		if err != nil {
			t.Fatalf("codex lane: %v", err)
		}
		if got := strings.Count(stderr, "codex lane: card t1 session: boom\n"); got != 1 {
			t.Errorf("direct-door session error line count = %d, want 1; stderr:\n%s", got, stderr)
		}
	})
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
			h := newCardChildLoop(t, cardChildOpts{cards: 2, managed: strPtr("1"), realOwner: true, source: strings.NewReader(tc.chunk)})
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

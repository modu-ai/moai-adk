package cli

// codex_lane_preflight_test.go — card t1779: `moai codex -l` must refuse a
// launch environment that cannot host a Codex session BEFORE any launch side
// effect, and must refuse launching a lane from inside a Codex session. The
// operator-measured failure: launches into a TERM=dumb sandbox burned 45s and
// 19s each before dying INSIDE the session, after the run join, the
// lane-label claim, and the launch-pending registration had already
// committed; and a `moai codex -l` from inside a Codex session recursed.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// armCodexLaneEnvironmentGate re-arms the real t1779 environment gate for the
// calling test: TestMain pins the seam open for this package's launch-
// mechanics tests (no test may depend on this process's terminal state), so
// the gate's own tests restore the real body for their own duration.
func armCodexLaneEnvironmentGate(t *testing.T) {
	t.Helper()
	prev := codexLaneEnvironmentRefusalFn
	codexLaneEnvironmentRefusalFn = codexLaneLaunchEnvironmentRefusal
	t.Cleanup(func() { codexLaneEnvironmentRefusalFn = prev })
}

// codexLaneRefusalFixture wires the lane fixture the existing -l tests use
// (a picked queue plus the leader run the join resolves), scrubs the ambient
// factory environment, and pins both launch seams to record instead of exec —
// so a refusal can be observed launching NOTHING.
func codexLaneRefusalFixture(t *testing.T) (root string, launches *int) {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
	count := 0
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return sentinelCodexBinaryPath, nil }
	codexDirectLaunchFn = func(*exec.Cmd) error { count++; return nil }
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })
	return root, &count
}

// codexLaneLaunchPendingRows counts the factorymsg launch-pending
// registration rows the broker roster carries for the fixture run.
func codexLaneLaunchPendingRows(t *testing.T, root string) int {
	t.Helper()
	s, err := factorymsg.Open(root, fcRun)
	if err != nil {
		t.Fatalf("open factory message broker: %v", err)
	}
	defer closeOnCleanup(t, "factory message broker", s)
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatalf("factory message broker status: %v", err)
	}
	n := 0
	for _, lane := range status.Lanes {
		if lane.BindingState == factorymsg.BindingLaunchPending {
			n++
		}
	}
	return n
}

// assertCodexLaneRefusal is the shared refusal contract (card t1779): ONE
// stderr line carrying the expected refusal, a non-zero exit, zero launches
// through any seam, and zero launch-pending registration rows.
func assertCodexLaneRefusal(t *testing.T, stderr string, err error, launches int, wantSubstring, root string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want a pre-execution refusal, got success (launches recorded: %d)", launches)
	}
	if code, ok := ResolveExitCode(err); !ok || code == 0 {
		t.Fatalf("refusal exit code = (%d, %v), want a non-zero code", code, ok)
	}
	lines := make([]string, 0, 1)
	for _, line := range strings.Split(stderr, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) != 1 {
		t.Fatalf("stderr carries %d non-empty lines (%q), want exactly one refusal line", len(lines), stderr)
	}
	if !strings.Contains(lines[0], wantSubstring) {
		t.Fatalf("refusal line %q does not name %q", lines[0], wantSubstring)
	}
	if launches != 0 {
		t.Fatalf("launches recorded through the seams = %d, want 0 — the refusal must precede every seam", launches)
	}
	if n := codexLaneLaunchPendingRows(t, root); n != 0 {
		t.Fatalf("factorymsg launch_pending rows = %d, want 0 — nothing registered, nothing to roll back", n)
	}
}

// TestCodexLaneLaunchRefusesDumbTerm — TERM=dumb cannot host a Codex session;
// the launch must refuse with one stderr line naming the interactive-terminal
// need BEFORE the join, the lane-label claim, the stamps, and the
// launch-pending registration (card t1779).
func TestCodexLaneLaunchRefusesDumbTerm(t *testing.T) {
	root, launches := codexLaneRefusalFixture(t)
	armCodexLaneEnvironmentGate(t)
	t.Setenv("TERM", "dumb")

	_, stderr, err := runCodexCmd(t, "-l")

	assertCodexLaneRefusal(t, stderr, err, *launches, "interactive terminal", root)
}

// TestCodexLaneLaunchRefusesRecursive — a Codex session carries the factory
// backend marker; a `moai codex -l` from inside one would recurse. The
// comparison mirrors factoryLaneRefusal's established pattern
// (internal/cli/factory_card.go): the marker is
// config.EnvFactoryBackend == factory.BackendGPT (card t1779).
func TestCodexLaneLaunchRefusesRecursive(t *testing.T) {
	root, launches := codexLaneRefusalFixture(t)
	armCodexLaneEnvironmentGate(t)
	t.Setenv(config.EnvFactoryBackend, factory.BackendGPT)

	_, stderr, err := runCodexCmd(t, "-l")

	assertCodexLaneRefusal(t, stderr, err, *launches, "recurs", root)
}

// TestCodexLaneLaunchRefusesPlainSessionRecursion — a plain `moai codex`
// launch scrubs the factory keys from its child (codexChildEnv drops the
// lane keys), so the gpt-backend marker cannot see a plain-session nest.
// The launcher's self-identification marker closes the gap: the fixture env
// carries ONLY config.EnvMoaiCodexSession — no gpt backend key — with TTY
// and a clean ps pinned, and the refusal still fires with zero side effects
// (card t1779, cold sync-audit F1). The writer half pins that codexChildEnv
// stamps the marker, which is what makes the production nest reachable.
func TestCodexLaneLaunchRefusesPlainSessionRecursion(t *testing.T) {
	root, launches := codexLaneRefusalFixture(t)
	armCodexLaneEnvironmentGate(t)
	prevProbe, prevStdio := codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn
	codexLaneSandboxProbeFn = func() error { return nil }
	codexLaneStdioIsTerminalFn = func() bool { return true }
	t.Cleanup(func() { codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn = prevProbe, prevStdio })
	t.Setenv(config.EnvMoaiCodexSession, "1")

	_, stderr, err := runCodexCmd(t, "-l")

	assertCodexLaneRefusal(t, stderr, err, *launches, "recurs", root)
	if got := launchEnvValue(codexChildEnv(), config.EnvMoaiCodexSession); got != "1" {
		t.Fatalf("codexChildEnv %s = %q, want \"1\"", config.EnvMoaiCodexSession, got)
	}
}

// TestCodexSpawnCommandCarriesSessionMarker — the BOTH-DOORS
// self-identification contract: the direct child env (codexChildEnv) and the
// spawn pane's command env (buildCodexSpawnCommandWithEnv) must carry the
// marker identically. The spawn door assembles its pane env on the tmux
// command line, never from req.Env, so a door that drops the stamp revives
// the F1 recursion class on a real-TTY pane (card t1779, sync-audit
// round-2 F4).
func TestCodexSpawnCommandCarriesSessionMarker(t *testing.T) {
	if got := launchEnvValue(codexChildEnv(), config.EnvMoaiCodexSession); got != "1" {
		t.Fatalf("direct door %s = %q, want \"1\"", config.EnvMoaiCodexSession, got)
	}
	command := buildCodexSpawnCommand("/test/codex", nil)
	if !strings.Contains(command, config.EnvMoaiCodexSession+"=1") {
		t.Fatalf("spawn command %q lacks %s=1 — the pane carries no session marker", command, config.EnvMoaiCodexSession)
	}
}

// TestCodexChildEnvStampsSessionMarkerWithoutAmbient — the writer half
// without ambient shadowing: with the marker absent from this process's
// environment, the stamp must still land, proving the stamp rather than
// t.Setenv inheritance (card t1779, sync-audit round-2 F3).
func TestCodexChildEnvStampsSessionMarkerWithoutAmbient(t *testing.T) {
	t.Setenv(config.EnvMoaiCodexSession, "")
	_ = os.Unsetenv(config.EnvMoaiCodexSession)
	if got := launchEnvValue(codexChildEnv(), config.EnvMoaiCodexSession); got != "1" {
		t.Fatalf("codexChildEnv %s = %q, want the stamp \"1\" without ambient inheritance", config.EnvMoaiCodexSession, got)
	}
}

// TestCodexLaneLaunchEnvironmentRefusalSandboxSignature — a sandbox whose
// process-listing probe fails with "operation not permitted" reads as the
// sandbox signature and earns the same refusal as a terminal-less
// environment, because the session cannot run there either; a probe failure
// of any other class is NOT the signature and refuses nothing, so a missing
// ps keeps today's late-failure behavior (card t1779).
func TestCodexLaneLaunchEnvironmentRefusalSandboxSignature(t *testing.T) {
	prevProbe, prevStdio := codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn
	codexLaneStdioIsTerminalFn = func() bool { return true } // isolate the probe branch
	t.Cleanup(func() { codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn = prevProbe, prevStdio })

	codexLaneSandboxProbeFn = func() error { return errors.New("ps: Operation not permitted") }
	if got := codexLaneLaunchEnvironmentRefusal([]string{"TERM=xterm-256color"}); !strings.Contains(got, "forbids process introspection") {
		t.Fatalf("sandbox-signature refusal = %q, want the sandbox line", got)
	}

	codexLaneSandboxProbeFn = func() error { return exec.ErrNotFound }
	if got := codexLaneLaunchEnvironmentRefusal([]string{"TERM=xterm-256color"}); got != "" {
		t.Fatalf("non-permission probe failure refusal = %q, want no refusal", got)
	}
}

// TestCodexLaneLaunch_MirrorsSessionMarkerIntoShellEnvPolicy — under
// shell_environment_policy.inherit=core/none the child-env stamp never
// reaches the session's shell, so the marker rides the SAME
// shell_environment_policy.set mirror the lane keys traverse (card t1819's
// path; sync-audit round-3 P2-A). A nested `moai codex -l` reads the shell's
// copy.
func TestCodexLaneLaunch_MirrorsSessionMarkerIntoShellEnvPolicy(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	var argv []string
	codexLookPath = func(string) (string, error) { return sentinelCodexBinaryPath, nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error { argv = c.Args; return nil }
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })
	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex -l: %v", err)
	}
	got := codexShellEnvSetValues(argv)
	if got[config.EnvMoaiCodexSession] != `"1"` {
		t.Fatalf("shell env override %s = %q, want %q (argv %q)", config.EnvMoaiCodexSession, got[config.EnvMoaiCodexSession], `"1"`, argv)
	}
}

// TestLaneAmbientScrubCoversCodexSessionMarker — the self-identification
// marker is scrubbed from this binary's ambient environment by BOTH helpers,
// so codex-hosting tests re-arming the real gate see a marker-free seat
// instead of refusal-wording mismatches (card t1779, round-3 P2-B).
func TestLaneAmbientScrubCoversCodexSessionMarker(t *testing.T) {
	t.Run("netScrubLaneEnv", func(t *testing.T) {
		t.Setenv(config.EnvMoaiCodexSession, "1")
		netScrubLaneEnv(t)
		if os.Getenv(config.EnvMoaiCodexSession) != "" {
			t.Fatalf("%s = %q survived netScrubLaneEnv", config.EnvMoaiCodexSession, os.Getenv(config.EnvMoaiCodexSession))
		}
	})
	t.Run("clearFactoryAmbientEnv", func(t *testing.T) {
		t.Setenv(config.EnvMoaiCodexSession, "1")
		clearFactoryAmbientEnv()
		if os.Getenv(config.EnvMoaiCodexSession) != "" {
			t.Fatalf("%s = %q survived clearFactoryAmbientEnv", config.EnvMoaiCodexSession, os.Getenv(config.EnvMoaiCodexSession))
		}
	})
}

// TestCodexLaneLaunchEnvironmentRefusalNonTerminalStdio — the session
// receives this process's stdin and stdout verbatim, so a pipe or a file
// there refuses the launch REGARDLESS of TERM: an inherited TERM=xterm-256color
// with piped stdio is exactly the scripted/agent-harness context the
// pre-execution refusal exists for (card t1779 card-review P2-2). Real
// terminals stay admitted whatever TERM carries.
func TestCodexLaneLaunchEnvironmentRefusalNonTerminalStdio(t *testing.T) {
	prevProbe, prevStdio := codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn
	codexLaneSandboxProbeFn = func() error { return nil } // isolate from the seat's ps state
	t.Cleanup(func() { codexLaneSandboxProbeFn, codexLaneStdioIsTerminalFn = prevProbe, prevStdio })

	codexLaneStdioIsTerminalFn = func() bool { return false }
	if got := codexLaneLaunchEnvironmentRefusal([]string{"TERM=xterm-256color"}); !strings.Contains(got, "interactive terminal") {
		t.Fatalf("piped-stdio refusal = %q, want the interactive-terminal line", got)
	}

	codexLaneStdioIsTerminalFn = func() bool { return true }
	if got := codexLaneLaunchEnvironmentRefusal([]string{"TERM=xterm-256color"}); got != "" {
		t.Fatalf("terminal-stdio refusal = %q, want none", got)
	}
	if got := codexLaneLaunchEnvironmentRefusal(nil); got != "" {
		t.Fatalf("TERM unset over real terminals refusal = %q, want none", got)
	}
}

// TestCodexSpawnStampFailureRollsBackPendingRegistration — card t1779's
// third required behavior: the EXISTING spawn-door rollback stays. When the
// lane-claim stamp fails AFTER the launch-pending registration (a lane whose
// workers row is gone — the stale-claim shape), the launch is refused, the
// pane cleaned up, and the launch's own pending row rolled back, while rows
// the launch did not register stay untouched.
func TestCodexSpawnStampFailureRollsBackPendingRegistration(t *testing.T) {
	root := codexLedRun(t, "rollback-run", BackendCodex)
	// Positive control on the roster read: a directly registered leader row
	// is visible, so the lane-1 zero assertion after the rollback is not an
	// empty sweep (mirrors TestManagedLaunchPendingRollback's control).
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		t.Fatal("test process identity unavailable")
	}
	if _, err := registerFactoryLaunchPending(context.Background(), root, []string{
		config.EnvFactoryRunID + "=rollback-run",
		config.EnvFactoryBackend + "=codex",
		config.EnvMoaiFactoryWorkers + "=1",
	}, os.Getpid(), start); err != nil {
		t.Fatal(err)
	}
	if got := peerCount(t, root, "rollback-run", "leader"); got != 1 {
		t.Fatalf("control leader rows = %d, want 1", got)
	}

	env := []string{
		config.EnvFactoryRunID + "=rollback-run",
		config.EnvFactoryBackend + "=codex",
		config.EnvMoaiFactoryWorker + "=lane-1",
		config.EnvMoaiFactoryWorkers + "=0",
	}
	prevSpawn, prevIdentity, prevCleanup := tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnCleanupPaneFn
	tmuxSpawnFn = func(string, string) (string, error) { return "%42", nil }
	codexSpawnPaneIdentityFn = func(string) (int, string, error) { return os.Getpid(), start, nil }
	cleanups := 0
	codexSpawnCleanupPaneFn = func(string) error { cleanups++; return nil }
	t.Cleanup(func() {
		tmuxSpawnFn, codexSpawnPaneIdentityFn, codexSpawnCleanupPaneFn = prevSpawn, prevIdentity, prevCleanup
	})

	// No resolveFactoryLaneName pre-claim: the workers row for lane-1 does
	// not exist, so the lane-claim stamp fails after the registration.
	err := defaultCodexSpawnLaunch(root, "/test/codex", nil, env)
	if err == nil || !strings.Contains(err.Error(), "factory lane claim") {
		t.Fatalf("launch error = %v, want the lane-claim refusal", err)
	}
	if cleanups != 1 {
		t.Fatalf("pane cleanups = %d, want 1 — the pane does not outlive a refused launch", cleanups)
	}
	if got := peerCount(t, root, "rollback-run", "lane-1"); got != 0 {
		t.Fatalf("lane-1 pending rows after rollback = %d, want 0 — the launch's own registration is rolled back", got)
	}
	if got := peerCount(t, root, "rollback-run", "leader"); got != 1 {
		t.Fatalf("leader control rows after rollback = %d, want 1 — the rollback removes only the launch's own row", got)
	}
}

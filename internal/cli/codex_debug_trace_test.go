package cli

// codex_debug_trace_test.go — SPEC-CODEX-DEBUG-MODE-001 M2: the launcher
// debug trace engine and the child-env debug linkage (AC-008, AC-009,
// AC-010, AC-011, AC-015, and the AC-014 non-gating pin).
//
// Every cell drives the built launcher path with the capture harness — no
// real codex process. Assertions are anchored on the literal prefix string
// and step vocabulary so they pin the operator-visible contract, not the
// implementation's identifiers.

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// The operator-visible trace vocabulary under test (REQ-005 / REQ-013).
const (
	testDebugTracePrefix = "moai-launcher-debug:"
	testExecSeamSentinel = "<<EXEC-SEAM>>"
)

// debugTraceStepLines counts the trace lines a step produced.
func debugTraceStepLines(stderr, step string) int {
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, testDebugTracePrefix+" "+step) {
			n++
		}
	}
	return n
}

// runCodexInto runs the codex entry with the trace sink bound to buf so a
// cell can assert the relative order of trace writes against another writer
// of the same buffer (the AC-015 stub-seam-relative form).
func runCodexInto(t *testing.T, buf io.Writer, args ...string) error {
	t.Helper()
	c := &cobra.Command{Use: "codex"}
	c.SetContext(context.Background())
	c.SetOut(io.Discard)
	c.SetErr(buf)
	return runCodex(c, args)
}

// envValues returns every value env carries for key, in order.
func envValues(env []string, key string) []string {
	var out []string
	for _, entry := range env {
		if k, v, _ := strings.Cut(entry, "="); k == key {
			out = append(out, v)
		}
	}
	return out
}

// unsetRustLog removes RUST_LOG for the test and restores the prior state.
func unsetRustLog(t *testing.T) {
	t.Helper()
	prev, had := os.LookupEnv(config.EnvRustLog)
	if !had {
		return
	}
	if err := os.Unsetenv(config.EnvRustLog); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Setenv(config.EnvRustLog, prev) })
}

// AC-008: one trace line per executed pre-exec step, each under the single
// prefix constant. A bare launch (no -f, no -w) executes exactly the six
// steps below; the lane claim and worktree steps must be absent.
func TestCodexDebugTraceStepLines(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	_, stderr, err := runCodexCmd(t, "-d")
	if err != nil {
		t.Fatalf("runCodex(-d): %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	for _, step := range []string{
		"binary resolution",
		"project-root resolution",
		"init-offer gate",
		"local-instruction load",
		"child-env assembly",
		"exec handoff",
	} {
		if got := debugTraceStepLines(stderr, step); got != 1 {
			t.Errorf("step %q produced %d trace lines, want exactly 1:\n%s", step, got, stderr)
		}
	}
	for _, absent := range []string{"lane claim", "worktree materialization"} {
		if got := debugTraceStepLines(stderr, absent); got != 0 {
			t.Errorf("step %q produced %d trace lines on a bare launch, want 0:\n%s", absent, got, stderr)
		}
	}
}

// AC-009: the trace names environment keys with their presence and never
// emits a value. Every lane axis the trace reads is pinned with a sentinel
// value (t1350: a test reading lane env pins every axis); the sentinel
// strings must appear nowhere in the captured stderr. The chain, SPEC and
// lane-label markers are no longer lane keys: they left the scrub list with
// their last publisher (SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a).
func TestCodexDebugTraceEnvKeysOnly(t *testing.T) {
	sentinels := map[string]string{
		config.EnvFactoryRunID:            "run-sentinel",
		config.EnvFactoryLeadAddr:         "leader-secret-sentinel",
		config.EnvFactoryLeadName:         "leadname-sentinel",
		config.EnvFactoryBackend:          "backend-sentinel",
		config.EnvFactoryCard:             "card-sentinel",
		config.EnvFactorySettingsInjected: "settings-sentinel",
		config.EnvMoaiFactoryWorker:       "worker-sentinel",
		config.EnvMoaiFactoryWorkers:      "workers-sentinel",
		config.EnvFactoryRole:             "role-sentinel",
	}
	for key, value := range sentinels {
		t.Setenv(key, value)
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	_, stderr, err := runCodexCmd(t, "-d")
	if err != nil {
		t.Fatalf("runCodex(-d): %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	if !strings.Contains(stderr, config.EnvFactoryLeadAddr) {
		t.Errorf("trace does not name the lane key %s with its presence:\n%s", config.EnvFactoryLeadAddr, stderr)
	}
	for key, value := range sentinels {
		if strings.Contains(stderr, value) {
			t.Errorf("trace leaked the VALUE of %s (sentinel %q found) — REQ-006 violation:\n%s", key, value, stderr)
		}
	}
	// The child env stays scrubbed: the sentinel values reach neither the
	// child argv nor the child environment under debug.
	childEnv := strings.Join(cap.records[0].Env, "\n")
	for key, value := range sentinels {
		if strings.Contains(childEnv, value) {
			t.Errorf("child env leaked the VALUE of %s (sentinel %q found)", key, value)
		}
	}
}

// AC-010: a -w entry traces the resolved directory, the writer-check
// outcome, and the anchor-lock outcome.
func TestCodexDebugTraceWorktreeEntry(t *testing.T) {
	root := t.TempDir()
	wt := filepath.Join(root, ".moai", "worktrees", "debug-fixture")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	_, stderr, err := runCodexCmd(t, "-w", "debug-fixture", "-d")
	if err != nil {
		t.Fatalf("runCodex(-w debug-fixture -d): %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	if filepath.Clean(cap.records[0].Dir) != filepath.Clean(wt) {
		t.Fatalf("launch dir = %q, want the fixture tree %q", cap.records[0].Dir, wt)
	}
	if got := debugTraceStepLines(stderr, "worktree materialization"); got != 1 {
		t.Fatalf("worktree materialization produced %d trace lines, want 1:\n%s", got, stderr)
	}
	for _, want := range []string{wt, "writer-check ok"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("worktree trace lacks %q:\n%s", want, stderr)
		}
	}
	if !strings.Contains(stderr, "anchor-lock ok") {
		t.Errorf("trace lacks the anchor-lock outcome:\n%s", stderr)
	}
}

// AC-011: RUST_LOG is injected when absent (REQ-010) and an operator-supplied
// value is preserved unmodified (REQ-011).
func TestCodexDebugRustLogInjection(t *testing.T) {
	t.Run("injected when absent (REQ-010)", func(t *testing.T) {
		unsetRustLog(t)
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		if _, _, err := runCodexCmd(t, "-d"); err != nil {
			t.Fatalf("runCodex(-d): %v", err)
		}
		values := envValues(cap.records[0].Env, config.EnvRustLog)
		if len(values) == 0 || values[len(values)-1] != "debug" {
			t.Errorf("child RUST_LOG = %v, want the last entry to be debug (last-wins append)", values)
		}
	})
	t.Run("operator value preserved (REQ-011)", func(t *testing.T) {
		t.Setenv(config.EnvRustLog, "info")
		cap := withCodexLaunchCapture(t)
		withCodexProjectRoot(t, t.TempDir())
		if _, _, err := runCodexCmd(t, "-d"); err != nil {
			t.Fatalf("runCodex(-d): %v", err)
		}
		values := envValues(cap.records[0].Env, config.EnvRustLog)
		if len(values) != 1 || values[0] != "info" {
			t.Errorf("child RUST_LOG = %v, want exactly the operator value [info] — REQ-011", values)
		}
	})
}

// AC-015: every trace line is written before the exec seam is invoked.
// codexDirectLaunchFn is the single launch seam BOTH doors route through
// (the POSIX door's syscall.Exec and the Windows Start/wait twin sit behind
// it), so the stub-seam-relative assertion covers both doors; the
// GOOS=windows build pins the compile axis.
func TestCodexDebugTracePrecedesExecSeam(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	// Re-wrap the harness's recording stub so the seam writes a sentinel
	// into the SAME buffer the trace sink uses — the relative order of the
	// writes is the assertion, never wall-clock timing.
	recordingStub := codexDirectLaunchFn
	var buf bytes.Buffer
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		if err := recordingStub(c); err != nil {
			return err
		}
		buf.WriteString(testExecSeamSentinel + "\n")
		return nil
	}
	if err := runCodexInto(t, &buf, "-d"); err != nil {
		t.Fatalf("runCodex(-d): %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	content := buf.String()
	seamIdx := strings.Index(content, testExecSeamSentinel)
	if seamIdx < 0 {
		t.Fatalf("exec seam sentinel never written:\n%s", content)
	}
	pre, post := content[:seamIdx], content[seamIdx:]
	steps := 0
	for _, line := range strings.Split(pre, "\n") {
		if strings.HasPrefix(line, testDebugTracePrefix) {
			steps++
		}
	}
	if steps < 6 {
		t.Errorf("only %d trace lines preceded the seam, want the full pre-exec vocabulary (>= 6):\n%s", steps, pre)
	}
	for _, line := range strings.Split(post, "\n") {
		if strings.HasPrefix(line, testDebugTracePrefix) {
			t.Errorf("trace line printed after the exec seam — REQ-009 violation:\n%s", line)
		}
	}
}

// AC-014: MOAI_LOG_LEVEL neither enables nor suppresses the trace — two
// separate axes (REQ-008). The pin runs a traced launch with the level at
// its most suppressive value and requires the full trace.
func TestCodexDebugTraceIgnoresLogLevel(t *testing.T) {
	t.Setenv(config.EnvLogLevel, "error")
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, t.TempDir())
	_, stderr, err := runCodexCmd(t, "-d")
	if err != nil {
		t.Fatalf("runCodex(-d) under MOAI_LOG_LEVEL=error: %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("launches = %d, want 1", cap.count())
	}
	for _, step := range []string{"binary resolution", "init-offer gate", "exec handoff"} {
		if got := debugTraceStepLines(stderr, step); got != 1 {
			t.Errorf("MOAI_LOG_LEVEL=error suppressed the %q trace line (%d found) — REQ-008:\n%s", step, got, stderr)
		}
	}
}

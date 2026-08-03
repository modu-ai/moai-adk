package cli

// hook_flush_test.go — SPEC-HOOK-TRACE-FLUSH-001 regression guards for the
// trace flush barrier at the hook CLI boundary.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repoRootFromCLITest returns the repository root relative to this package's
// test working directory.
func repoRootFromCLITest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root %s has no go.mod: %v", root, err)
	}
	return root
}

// observabilityFixture is the minimum project config that turns hook tracing on.
const observabilityFixture = `observability:
  enabled: true
  trace_dir: .moai/logs
  report_dir: .moai/reports
`

// newTracedProject creates a temp project whose only content is the
// observability config and the log directory the trace writer writes into.
func newTracedProject(t *testing.T) string {
	t.Helper()
	proj := t.TempDir()
	cfgDir := filepath.Join(proj, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "observability.yaml"), []byte(observabilityFixture), 0o644); err != nil {
		t.Fatalf("write observability.yaml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(proj, ".moai", "logs"), 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}
	return proj
}

// TestHookCommandFlushesLastHandlerEntry is the end-to-end judgment for
// REQ-HTF-002: it crosses a real process boundary, which is the only way to
// exercise the flush barrier wired into the hook CLI. The in-package registry
// guard cannot reach this code, because internal/hook does not depend on
// internal/cli.
//
// It automates the controlled experiment recorded in spec.md §1.1: dispatching
// SessionStart runs three handlers, and before the fix only the first one's
// entry survived — the later handlers contributed 0 of 20 expected entries
// across ten runs. The defect is a race against process exit, so the hook is
// executed repeatedly rather than once.
//
// The guard asserts STRICT per-run retention (every run must retain all
// wantDistinctHandlers), not a quorum. The drain barrier is signal-confirmed
// (CloseWithTimeout blocks on the writer's done channel, with the timer only
// as a hang safety cap), so once DefaultTraceFlushTimeout was raised to 2s to
// absorb ubuntu CI's -race + coverage syscall inflation, every run reliably
// completes the drain. A quorum that tolerates the trailing handlers dying in
// a majority of runs would defeat the SPEC's core guard (REQ-HTF-008: the
// last handler's entry MUST be asserted to exist).
func TestHookCommandFlushesLastHandlerEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the moai binary; skipped under -short")
	}

	root := repoRootFromCLITest(t)
	bin := filepath.Join(t.TempDir(), "moai")
	// The binary is stamped with a dev version so the SessionStart auto-update
	// handler (buildAutoUpdateFunc) takes its isDevBuild early-return and never
	// reaches the network. Without this, the binary reports the hardcoded
	// default "v3.0.0"; on CI the handler's GitHub checker (unauthenticated,
	// whose API call a fresh runner IP succeeds at where a saturated local IP
	// gets HTTP 403) finds the latest release v3.0.1 > v3.0.0, downloads it,
	// and the updater atomically replaces os.Executable() — this very binary —
	// mid-test (internal/update/updater.go renames onto binaryPath, which deps
	// resolves to os.Executable()). Run 0 finishes on the original image, but
	// runs 1..N exec the released v3.0.1, which predates this SPEC and has NO
	// defer rs.Shutdown(), so the trailing handlers' trace entries race process
	// exit and lose exactly as spec.md §1.1 measured. Stamping "dev" keeps every
	// run on the image under test.
	build := exec.Command("go", "build",
		"-ldflags=-X github.com/modu-ai/moai-adk/pkg/version.Version=dev",
		"-o", bin, "./cmd/moai")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build moai binary: %v\n%s", err, out)
	}

	// spec.md §1.1 measured three registered SessionStart handlers. Asserting
	// the floor rather than the exact count keeps the guard alive if a handler
	// is added, while still failing loudly if the trailing handlers vanish —
	// which is exactly the symptom this SPEC removes.
	const wantDistinctHandlers = 3

	// Strict per-run assertion (REQ-HTF-008): every process-boundary crossing
	// MUST retain all wantDistinctHandlers. This guard previously used a quorum
	// (4/10) to tolerate the slow-disk residual loss the async design accepts
	// under a tight flush budget — but the budget is the root cause, not the
	// assertion. DefaultTraceFlushTimeout was raised from 200ms to 2s so the
	// drain barrier (signal-confirmed on the writer's done channel) reliably
	// completes even under ubuntu CI's -race + coverage instrumentation, where
	// per-syscall overhead inflates 5-20x. With the root cause corrected, the
	// strict per-run assertion is restored: a quorum that tolerates the
	// trailing handlers dying in 60% of runs defeats the SPEC's core guard.
	const runs = 5

	for i := range runs {
		proj := newTracedProject(t)
		sessionID := fmt.Sprintf("flush-e2e-%d", i)
		stdin := fmt.Sprintf(
			`{"session_id":%q,"hook_event_name":"SessionStart","cwd":%q}`,
			sessionID, proj,
		)

		cmd := exec.Command(bin, "hook", "session-start")
		cmd.Dir = proj
		cmd.Stdin = strings.NewReader(stdin)
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+proj)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("run %d: moai hook session-start: %v\nstderr: %s", i, err, stderr.String())
		}

		tracePath := filepath.Join(proj, ".moai", "logs", "trace-"+sessionID+".jsonl")
		// A missing or empty trace file is the hard-failure case: it means the
		// writer never drained even one entry before the process exited — the
		// exact defect the flush barrier exists to prevent. countDistinctHandlersQuiet
		// returns 0 for a missing file, which fails the strict floor below.
		handlerCount := countDistinctHandlersQuiet(tracePath)
		if handlerCount < wantDistinctHandlers {
			t.Errorf("run %d: trace retained %d distinct handlers, want at least %d — "+
				"the trailing handlers' entries died with the process (stderr: %s)",
				i, handlerCount, wantDistinctHandlers, stderr.String())
		}
	}
}

// countDistinctHandlersQuiet returns the number of distinct handler values in
// the trace file, or 0 on any error (missing file, read error, unparseable
// line — including partial-write residue). Non-fatal line parsing keeps a
// half-written final JSON line from aborting the count; a missing/empty file
// yields 0, which fails the strict per-run floor in the caller — that is the
// intended hard failure (the writer never drained even one entry).
func countDistinctHandlersQuiet(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry struct {
			Handler string `json:"handler"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // skip unparseable lines (partial-write residue)
		}
		if entry.Handler != "" {
			seen[entry.Handler] = true
		}
	}
	return len(seen)
}

// flushBarrierCall matches a call to either flush-barrier entry point.
var flushBarrierCall = regexp.MustCompile(`CloseWithTimeout\(|\.Shutdown\(\)`)

// TestFlushBarrierHasProductionCaller asserts that the flush barrier is reached
// from production code rather than from tests alone (AC-HTF-009 — REQ-HTF-010).
//
// This guard proves only that the call exists in source; it says nothing about
// whether the call works. The behavioral judgments are
// TestRegistryShutdownFlushesLastHandlerEntry and
// TestHookCommandFlushesLastHandlerEntry, and this guard does not substitute
// for them. Its value is narrow: it catches a teardown call being deleted
// outright, which was the original defect — Close had zero production callers.
//
// The search is scoped to the hook and CLI packages, and comment lines are
// excluded. A wider scan loses discrimination: it matches an unrelated comment
// in internal/web and passes on a tree with no teardown wiring at all.
func TestFlushBarrierHasProductionCaller(t *testing.T) {
	t.Parallel()

	root := repoRootFromCLITest(t)
	var callers []string

	for _, pkgDir := range []string{"internal/hook", "internal/cli"} {
		walkRoot := filepath.Join(root, pkgDir)
		err := filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for n, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "//") {
					continue
				}
				if flushBarrierCall.MatchString(line) {
					rel, _ := filepath.Rel(root, path)
					callers = append(callers, fmt.Sprintf("%s:%d", rel, n+1))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", pkgDir, err)
		}
	}

	if len(callers) == 0 {
		t.Error("no production caller of the trace flush barrier in internal/hook or internal/cli; " +
			"an async trace writer with no teardown caller silently loses entries at process exit")
	}
	t.Logf("flush barrier production callers: %v", callers)
}

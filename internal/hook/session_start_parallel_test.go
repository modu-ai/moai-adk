package hook

// Characterization tests for the SessionStart input-lag reduction
// (errgroup parallelization + deferred heavy scanning).
//
// These tests pin the OBSERVABLE contract of the refactored Handle():
//   1. Heavy advisory scanning (drift detection et al.) is deferred off the
//      synchronous critical path — Handle() returns promptly even when the
//      drift computation is slow. Previously drift ran synchronously up to
//      its time-box and blocked Handle() for the full computation when the
//      time-box was large.
//   2. Turn-visible synchronous side effects are preserved: AdditionalContext
//      attribution is injected, the multi-session registry is written, and
//      the synchronous return's Data map still carries the markers produced
//      by steps that must complete before the hook returns.
//
// Behavior preservation is the goal: these tests do NOT pin step ORDER, only
// that the side effects reach a developer-observable state.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/session"
)

// TestSessionStart_DeferredScanDoesNotBlockReturn is the RED→GREEN pin for
// change (b): the heavy advisory scanning MUST NOT block Handle()'s return.
// We inject a drift computation that blocks longer than the assertion budget
// and a time-box larger than the block, then assert Handle() returns promptly.
//
// Against the pre-refactor code this test FAILS: detectStatusDrift runs the
// injected slow fn synchronously and blocks Handle for the full duration.
// After deferral, the slow scan runs in a background goroutine and Handle
// returns immediately.
func TestSessionStart_DeferredScanDoesNotBlockReturn(t *testing.T) {
	// Isolate GLM PROCESS env so glmGuardrailReminder() returns "" on a
	// developer machine running CG mode (same isolation pattern as
	// TestSessionStartHandler_Handle). Non-parallel: mutates package seams.
	t.Setenv("ANTHROPIC_BASE_URL", "")
	stageLap := observeStageLaps(t)

	// SPEC-DRIFT-CACHE-FILL-001 REQ-DCF-011: the deferred step's drift work is
	// now the HEAD-SHA cache RESOLVE (driftCachedCountFn), not the in-band
	// compute (driftCountFn). The slow scan is injected at the seam the
	// deferred step actually consults; injecting it at driftCountFn would leave
	// this test green against any implementation, because nothing on the
	// deferred path calls that seam any more.
	origFn := driftCachedCountFn
	origTimeout := sessionStartDriftTimeout
	t.Cleanup(func() {
		driftCachedCountFn = origFn
		sessionStartDriftTimeout = origTimeout
	})

	sessionStartDriftTimeout = 30 * time.Second
	// driftAbort lets the test release the deferred scan after the synchronous
	// assertions pass, so the deferred goroutine can exit and join via
	// completedCh (goleak hygiene).
	driftAbort := make(chan struct{})
	driftCachedCountFn = func(string) (int, bool) {
		select {
		case <-driftAbort:
			return 7, true // test released the scan; value irrelevant
		case <-time.After(deferredScanSafetyNet):
			return 99, true
		}
	}
	completedCh := registerDeferredScanSeam(t)

	projectDir := t.TempDir()
	mkStateDir(t, projectDir)
	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:     "sess-deferred-scan-pin",
		CWD:           t.TempDir(),
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	// Deferral is read from two facts that do not depend on how long Handle's
	// unrelated synchronous steps took (card t2048): the scan was still pending
	// when Handle returned, and the deferred join alone stayed within its bound.
	// A total-elapsed threshold here (1s, from t662) failed correct runs whose
	// synchronous git and file-lock steps ran slow (1.046s observed), while a
	// scan moved onto the synchronous path or an unbounded join still fails both
	// checks. Total elapsed time is not asserted here: the input-lag contract
	// belongs to TestSessionStart_HandleInputLagBudget.
	assertDeferralOffSyncPath(t, completedCh, stageLap)

	// Turn-visible attribution MUST remain synchronous.
	if out.HookSpecificOutput == nil || !strings.Contains(out.HookSpecificOutput.AdditionalContext, "source_session_id=") {
		t.Errorf("AdditionalContext attribution missing or malformed: %+v", out.HookSpecificOutput)
	}

	// Release the deferred drift fn and join the goroutine so goleak sees a
	// clean exit. If the join does not complete within 3s the goroutine never
	// reached the drift fn (or never respected driftAbort), which would itself
	// be a defect.
	close(driftAbort)
	waitDeferred(t, completedCh, 3*time.Second)
}

// TestSessionStart_HandleInputLagBudget pins the OTHER contract that used to
// ride on TestSessionStart_DeferredScanDoesNotBlockReturn's 500ms threshold:
// SessionStart must not add unbounded input lag under NORMAL conditions — a
// drift scan that completes promptly, so the bounded join contributes ~0 and
// the elapsed time is Handle's own synchronous work.
//
// It is a separate test because it answers a separate question. Deferral is a
// structural property (did the scan run on the critical path?); input lag is a
// wall-clock property (how long did the user wait?). Riding both on one
// threshold is what let a slow-but-correct run fail with a message blaming
// deferral (t662).
//
// Threshold derivation (t662):
//
//	observed max, this condition   539.75ms   measured, n=300
//	* 2.5                        = 1349ms     -> 1.5s
//
// The 2.5x covers two axes the measurement did NOT reach: the CI machine class
// (all figures are from one darwin developer machine) and a real repository's
// scan cost (the measurement ran against an empty t.TempDir).
//
// This is deliberately a GROSS-REGRESSION guard — roughly 7x the 218ms median —
// NOT a latency SLO. A tight wall-clock budget on a contended machine is
// precisely the defect t662 diagnosed: it fires on load rather than on code.
// Tightening it requires re-deriving the bound from a fresh measurement that
// covers the two axes above.
//
// Measurement: .moai/reports/t662/verdict.md
func TestSessionStart_HandleInputLagBudget(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origFn := driftCachedCountFn
	t.Cleanup(func() { driftCachedCountFn = origFn })
	// The NORMAL case at the seam the deferred step consults (REQ-DCF-011): a
	// cache HIT, so the join returns immediately, no fill is started, and
	// elapsed is Handle's synchronous work.
	driftCachedCountFn = func(string) (int, bool) { return 0, true }
	completedCh := registerDeferredScanSeam(t)

	projectDir := t.TempDir()
	mkStateDir(t, projectDir)
	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:     "sess-input-lag-budget",
		CWD:           t.TempDir(),
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	}

	start := time.Now()
	out, err := h.Handle(context.Background(), input)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	const inputLagBudget = 1500 * time.Millisecond
	if elapsed > inputLagBudget {
		t.Fatalf("SessionStart added %v of input lag with a prompt advisory scan (budget %v); "+
			"the synchronous path regressed — see .moai/reports/t662/verdict.md for the derivation",
			elapsed, inputLagBudget)
	}

	waitDeferred(t, completedCh, 3*time.Second)
}

// TestSessionStart_SynchronousSideEffectsPreserved pins change (a): the
// independent synchronous steps run concurrently (errgroup) but every
// developer-observable side effect still lands before Handle() returns.
// This characterizes the merged behavior, not the order of execution.
func TestSessionStart_SynchronousSideEffectsPreserved(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origFn := driftCachedCountFn
	t.Cleanup(func() { driftCachedCountFn = origFn })
	// A deferred scan that would block if it ran on the synchronous path,
	// injected at the seam the deferred step consults (REQ-DCF-011).
	driftAbort := make(chan struct{})
	driftCachedCountFn = func(string) (int, bool) {
		select {
		case <-driftAbort:
			return 7, true
		case <-time.After(5 * time.Second):
			return 7, true
		}
	}
	completedCh := registerDeferredScanSeam(t)

	projectDir := t.TempDir()
	mkStateDir(t, projectDir)

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:     "sess-sync-effects-pin",
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	// Registry write completed synchronously — the entry is on disk.
	entries, err := session.NewRegistry(
		filepath.Join(projectDir, session.DefaultRegistryPath), nil,
	).Query("")
	if err != nil {
		t.Fatalf("registry query: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.SessionID == input.SessionID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("session %q not registered before Handle returned", input.SessionID)
	}

	// Synchronous Data markers MUST still be present in the return payload.
	if len(out.Data) == 0 {
		t.Fatal("Data payload empty")
	}

	// Release the deferred drift fn and join the goroutine (goleak hygiene).
	close(driftAbort)
	waitDeferred(t, completedCh, 3*time.Second)
}

// TestSessionStart_DeferredScanJoinsWithinBound pins the bounded-join
// contract added in change 1: a deferred scan that completes WITHIN
// deferredScanJoinBound lands its advisory key in this session's Data map,
// while a scan that EXCEEDS the bound is dropped (advisory absent).
//
// This complements TestSessionStart_DeferredScanDoesNotBlockReturn (which
// proves the bound caps input lag) by proving the join is EFFECTIVE — fast
// scans are not unconditionally dropped the way pure fire-and-forget would.
func TestSessionStart_DeferredScanJoinsWithinBound(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	// Injected at driftCachedCountFn: the deferred step resolves drift from the
	// HEAD-SHA cache now (REQ-DCF-011), so this is the seam whose speed decides
	// whether the advisory lands inside the bound.
	origFn := driftCachedCountFn
	t.Cleanup(func() { driftCachedCountFn = origFn })

	// Subcase A — fast scan completes within the bound: the cache resolve
	// returns immediately with a count at the warning threshold, so its
	// status_drift_warning advisory MUST land in the synchronous Data return.
	t.Run("fast scan lands advisory", func(t *testing.T) {
		driftAbort := make(chan struct{})
		driftCachedCountFn = func(string) (int, bool) {
			select {
			case <-driftAbort:
				return driftWarningThreshold, true
			default:
				return driftWarningThreshold, true
			}
		}
		completedCh := registerDeferredScanSeam(t)

		projectDir := t.TempDir()
		mkStateDir(t, projectDir)
		h := NewSessionStartHandler(nil)
		out, err := h.Handle(context.Background(), &HookInput{
			SessionID:     "sess-join-fast",
			CWD:           projectDir,
			ProjectDir:    projectDir,
			HookEventName: "SessionStart",
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(out.Data, &payload); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
		if _, ok := payload["status_drift_warning"]; !ok {
			t.Errorf("fast scan advisory was DROPPED; expected status_drift_warning to land within bound. payload=%v", payload)
		}

		close(driftAbort)
		waitDeferred(t, completedCh, 3*time.Second)
	})

	// Subcase B — slow scan exceeds the bound: the deferred step blocks longer
	// than deferredScanJoinBound, so its advisory MUST be dropped from the
	// synchronous Data return (non-blocking), and Handle MUST return within
	// a small envelope above the bound.
	t.Run("slow scan drops advisory", func(t *testing.T) {
		driftAbort := make(chan struct{})
		driftCachedCountFn = func(string) (int, bool) {
			select {
			case <-driftAbort:
				return driftWarningThreshold, true
			case <-time.After(deferredScanSafetyNet):
				return driftWarningThreshold, true
			}
		}
		completedCh := registerDeferredScanSeam(t)

		projectDir := t.TempDir()
		mkStateDir(t, projectDir)
		h := NewSessionStartHandler(nil)
		stageLap := observeStageLaps(t)

		out, err := h.Handle(context.Background(), &HookInput{
			SessionID:     "sess-join-slow",
			CWD:           projectDir,
			ProjectDir:    projectDir,
			HookEventName: "SessionStart",
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(out.Data, &payload); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
		if _, ok := payload["status_drift_warning"]; ok {
			t.Errorf("slow scan advisory was NOT dropped; expected absence since scan exceeds bound. payload=%v", payload)
		}

		// Handle must return with the slow scan still pending and its join
		// within the bound; it must NOT wait for the scan to finish.
		assertDeferralOffSyncPath(t, completedCh, stageLap)

		close(driftAbort)
		waitDeferred(t, completedCh, 3*time.Second)
	})
}

// deferredScanSafetyNet is how long an injected slow deferred scan blocks when
// the test never releases it. The assertions do not wait on it: a correct
// Handle returns after the 250ms join, long before this. It exists so that a
// defect which puts the scan on the synchronous path or removes the join bound
// shows up as a deferred_join lap of this length (far above the limit below)
// and a scan already finished at return, instead of hanging the test.
const deferredScanSafetyNet = 5 * time.Second

// observeStageLaps installs the SessionStart stage observer for the test and
// returns a reader of the accumulated time per stage name.
func observeStageLaps(t *testing.T) func(stage string) time.Duration {
	t.Helper()
	orig := handleStageObserver
	var mu sync.Mutex
	got := map[string]time.Duration{}
	handleStageObserver = func(stage string, d time.Duration) {
		mu.Lock()
		got[stage] += d
		mu.Unlock()
	}
	t.Cleanup(func() { handleStageObserver = orig })
	return func(stage string) time.Duration {
		mu.Lock()
		defer mu.Unlock()
		return got[stage]
	}
}

// assertDeferralOffSyncPath asserts, from facts that do not depend on the cost
// of Handle's other synchronous steps, that the deferred advisory scan is off
// the synchronous path and its join is bounded: (1) the scan had not finished
// when Handle returned, and (2) the deferred_join stage alone stayed within
// twice deferredScanJoinBound (the join waits one bound; the second bound is
// scheduling headroom).
func assertDeferralOffSyncPath(t *testing.T, completed <-chan struct{}, stageLap func(string) time.Duration) {
	t.Helper()
	select {
	case <-completed:
		t.Fatal("the deferred scan had already finished when Handle returned: it ran on the synchronous path")
	default:
	}
	if l := stageLap("deferred_join"); l > 2*deferredScanJoinBound {
		t.Fatalf("deferred_join took %v, want at most %v: the join must be bounded by %v", l, 2*deferredScanJoinBound, deferredScanJoinBound)
	}
}

// mkStateDir ensures .moai/state exists under the project dir so the
// multi-session registry has a writable home.
func mkStateDir(t *testing.T, projectDir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
}

// registerDeferredScanSeam installs the deferred-advisory-scan completion
// seam (test-only) so the test can join the background goroutine, AND opts
// this test back into the production async path (deferredScansAsync=true).
// TestMain defaults the package to sync (false) so the ~50 Handle-calling
// tests never leak the goroutine; any test that installs this seam is by
// definition exercising the async path, so the two toggles belong together.
// Both are cleared automatically on test completion. Returns the completion
// channel the goroutine closes on exit.
func registerDeferredScanSeam(t *testing.T) chan struct{} {
	t.Helper()
	ch := make(chan struct{})
	deferredScanSeamMu.Lock()
	deferredScanCompletedCh = ch
	origAsync := deferredScansAsync
	deferredScansAsync = true
	deferredScanSeamMu.Unlock()
	t.Cleanup(func() {
		deferredScanSeamMu.Lock()
		deferredScanCompletedCh = nil
		deferredScansAsync = origAsync
		deferredScanSeamMu.Unlock()
	})
	return ch
}

// waitDeferred blocks until the deferred goroutine signals completion or the
// timeout expires (a timeout indicates the goroutine never reached the drift
// fn or never respected the test's abort — itself a defect).
func waitDeferred(t *testing.T, completed <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-completed:
		return
	case <-time.After(timeout):
		t.Fatalf("deferred advisory scan goroutine did not exit within %v", timeout)
	}
}

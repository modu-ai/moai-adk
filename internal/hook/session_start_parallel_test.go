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

	origFn := driftCountFn
	origTimeout := sessionStartDriftTimeout
	t.Cleanup(func() {
		driftCountFn = origFn
		sessionStartDriftTimeout = origTimeout
	})

	// A time-box generous enough that the synchronous code path would wait
	// for the full block (proving the block, not the time-box, governs).
	sessionStartDriftTimeout = 30 * time.Second
	// driftAbort lets the test release the deferred drift fn after the
	// synchronous assertions pass, so the deferred goroutine can exit and
	// join via completedCh (goleak hygiene).
	driftAbort := make(chan struct{})
	driftCountFn = func(ctx context.Context, _ string) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-driftAbort:
			return 7, nil // test released the scan; value irrelevant
		case <-time.After(2 * time.Second):
			return 99, nil
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

	start := time.Now()
	out, err := h.Handle(context.Background(), input)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	// The deferred scan must NOT block the synchronous return. A 500ms budget
	// is well under the 2s injected block, so a passing run proves the scan
	// was deferred rather than run synchronously.
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Handle blocked %v waiting for advisory scan; expected deferred (non-blocking) return", elapsed)
	}

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

// TestSessionStart_SynchronousSideEffectsPreserved pins change (a): the
// independent synchronous steps run concurrently (errgroup) but every
// developer-observable side effect still lands before Handle() returns.
// This characterizes the merged behavior, not the order of execution.
func TestSessionStart_SynchronousSideEffectsPreserved(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "")

	origFn := driftCountFn
	t.Cleanup(func() { driftCountFn = origFn })
	// A drift fn that would block if it ran on the synchronous path.
	driftAbort := make(chan struct{})
	driftCountFn = func(ctx context.Context, _ string) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-driftAbort:
			return 7, nil
		case <-time.After(5 * time.Second):
			return 7, nil
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

	origFn := driftCountFn
	t.Cleanup(func() { driftCountFn = origFn })

	// Subcase A — fast scan completes within the bound: drift returns
	// immediately with a count at the warning threshold, so its
	// status_drift_warning advisory MUST land in the synchronous Data return.
	t.Run("fast scan lands advisory", func(t *testing.T) {
		driftAbort := make(chan struct{})
		driftCountFn = func(ctx context.Context, _ string) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-driftAbort:
				return driftWarningThreshold, nil
			default:
				return driftWarningThreshold, nil
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

	// Subcase B — slow scan exceeds the bound: drift blocks longer than
	// deferredScanJoinBound, so its advisory MUST be dropped from the
	// synchronous Data return (non-blocking), and Handle MUST return within
	// a small envelope above the bound.
	t.Run("slow scan drops advisory", func(t *testing.T) {
		driftAbort := make(chan struct{})
		driftCountFn = func(ctx context.Context, _ string) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-driftAbort:
				return driftWarningThreshold, nil
			case <-time.After(deferredScanJoinBound + 200*time.Millisecond):
				return driftWarningThreshold, nil
			}
		}
		completedCh := registerDeferredScanSeam(t)

		projectDir := t.TempDir()
		mkStateDir(t, projectDir)
		h := NewSessionStartHandler(nil)

		start := time.Now()
		out, err := h.Handle(context.Background(), &HookInput{
			SessionID:     "sess-join-slow",
			CWD:           projectDir,
			ProjectDir:    projectDir,
			HookEventName: "SessionStart",
		})
		elapsed := time.Since(start)
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

		// Handle must return within a small envelope above the bound (the
		// bound itself plus modest headroom for the synchronous errgroup
		// steps). It must NOT approach the slow scan's full duration.
		if elapsed > deferredScanJoinBound+500*time.Millisecond {
			t.Errorf("Handle blocked %v; expected return near the %v bound", elapsed, deferredScanJoinBound)
		}

		close(driftAbort)
		waitDeferred(t, completedCh, 3*time.Second)
	})
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

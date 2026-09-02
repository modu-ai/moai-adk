package hook

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/mx"
)

// SPEC-GRAPH-REPORT-001 REQ-GR-010 — deferred edges refresh step mechanics.
//
// The handler-level contract under test: the injected DeferredEdgesRefresh
// seam is invoked (exactly once, with the session's project dir) only when
// the exported staleness predicates say the edges layer moved, is NOT invoked
// when they say fresh, and is skipped entirely (probe included) when the seam
// is nil — the pre-M4 construction. A failing seam never fails Handle
// (REQ-GR-011 fail-open).
//
// All cases run in synchronous-deferred mode (TestMain flips the async seam
// off for this test binary), so the step completes inside Handle and the
// goleak hygiene of the package applies unchanged.

// freshEdgesFixture stamps a tree whose edges layer probes FRESH: an mx sidecar
// with matching provenance plus an edges meta stamped with the tree's current
// source fingerprints (the same prep the cli package's selected-artifact test
// uses — graph-side builds never write the sidecar themselves).
func freshEdgesFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if _, err := mx.RefreshIndex(filepath.Join(root, ".moai", "state"), root, nil); err != nil {
		t.Fatalf("refresh mx index: %v", err)
	}
	metaDir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := graph.WriteEdgesMeta(filepath.Join(metaDir, graph.MetaFileName),
		root, graph.SourceFingerprintsForEdges(root), 0); err != nil {
		t.Fatal(err)
	}
	return root
}

func runSessionStartHandle(t *testing.T, h Handler, projectDir string) error {
	t.Helper()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "deferred-edges-test",
		CWD:           projectDir,
		ProjectDir:    projectDir,
		HookEventName: "SessionStart",
	})
	return err
}

// edgesSeamRecorder records the project dirs the injected seam saw.
type edgesSeamRecorder struct {
	mu   sync.Mutex
	dirs []string
}

func (r *edgesSeamRecorder) fn(projectDir string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dirs = append(r.dirs, projectDir)
	return nil
}

func (r *edgesSeamRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.dirs)
}

// A stale tree (no meta sidecar, no mx sidecar — both predicates red) invokes
// the injected seam exactly once with the session's project dir.
func TestSessionStartDeferredEdgesRefresh_StaleInvokesSeam(t *testing.T) {
	root := t.TempDir() // bare: EdgesSourcesMoved && MXIndexNeedsRefresh both true
	if !graph.EdgesSourcesMoved(root) {
		t.Fatal("precondition: bare fixture must probe stale (no meta sidecar)")
	}

	rec := &edgesSeamRecorder{}
	h := NewSessionStartHandler(nil,
		WithSynchronousDeferredScans(),
		WithDeferredEdgesRefresh(rec.fn),
	)
	if err := runSessionStartHandle(t, h, root); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got := rec.count(); got != 1 {
		t.Errorf("stale tree must invoke the seam exactly once, got %d (dirs: %v)", got, rec.dirs)
	}
	if rec.count() == 1 && rec.dirs[0] != root {
		t.Errorf("seam must receive the session's project dir %q, got %q", root, rec.dirs[0])
	}
}

// A fresh tree (matching fingerprints, in-sync mx sidecar) does not invoke the
// seam — the deferred step is a no-op, not a unconditional rebuild.
func TestSessionStartDeferredEdgesRefresh_FreshSkipsSeam(t *testing.T) {
	root := freshEdgesFixture(t)
	if graph.EdgesSourcesMoved(root) || graph.MXIndexNeedsRefresh(root, graph.DefaultThresholds().MXIndexChangedFiles) {
		t.Fatal("precondition: fixture must probe fresh on both predicates")
	}

	rec := &edgesSeamRecorder{}
	h := NewSessionStartHandler(nil,
		WithSynchronousDeferredScans(),
		WithDeferredEdgesRefresh(rec.fn),
	)
	if err := runSessionStartHandle(t, h, root); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got := rec.count(); got != 0 {
		t.Errorf("fresh tree must not invoke the seam, got %d invocation(s)", got)
	}
}

// Backward compatibility: a handler constructed WITHOUT the seam (every
// pre-M4 call site) skips the step entirely — no probe side effects, no
// writes, Handle unchanged.
func TestSessionStartDeferredEdgesRefresh_NilSeamSkipsStep(t *testing.T) {
	root := t.TempDir() // stale on both predicates

	h := NewSessionStartHandler(nil, WithSynchronousDeferredScans())
	if err := runSessionStartHandle(t, h, root); err != nil {
		t.Fatalf("Handle (nil seam, stale tree): %v", err)
	}
	// The step must not have materialized any edges-layer artifact: the
	// probe has no side effects and the refresh (the only writer) is absent.
	for _, name := range []string{"edges.jsonl", graph.MetaFileName} {
		if _, err := os.Stat(filepath.Join(root, ".moai", "project", "graph", name)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("nil seam must skip the step, but %s exists (stat err: %v)", name, err)
		}
	}
}

// Fail-open (REQ-GR-011): a seam error is logged inside Handle and never
// fails the session start.
func TestSessionStartDeferredEdgesRefresh_FailOpenOnError(t *testing.T) {
	root := t.TempDir() // stale

	calls := 0
	h := NewSessionStartHandler(nil,
		WithSynchronousDeferredScans(),
		WithDeferredEdgesRefresh(func(string) error {
			calls++
			return errors.New("boom: rebuild refused")
		}),
	)
	if err := runSessionStartHandle(t, h, root); err != nil {
		t.Fatalf("a seam failure must not fail Handle (fail-open), got: %v", err)
	}
	if calls != 1 {
		t.Errorf("stale tree must still have invoked the failing seam once, got %d", calls)
	}
}

// The helper's own nil-guard is a no-op even when called directly — the
// defensive double-guard behind the edgesStale gating (safe from any future
// call site).
func TestSessionStartDeferredEdgesRefresh_NilGuardIsNoOp(t *testing.T) {
	h, ok := NewSessionStartHandler(nil).(*sessionStartHandler)
	if !ok {
		t.Fatalf("constructor must return *sessionStartHandler, got %T", h)
	}
	// Must neither panic nor write anything.
	h.runDeferredEdgesRefresh(t.TempDir())
}

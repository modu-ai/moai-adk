package hook

// navigator_detect_hardening_test.go — SPEC-NAVIGATOR-SYNC-002 M1.4 (fail-open
// + concurrency hardening). TDD coverage for:
//   - AC-NS2-004 — fail-open across 5 error modes (table-driven: 004a graph
//     absent, 004b unparseable JSON, 004c schema-invalid, 004d traversal error,
//     004e timeout).
//   - AC-NS2-006 — atomic read during regen (the reader observes the prior
//     committed graph while the M0 writer is held at the
//     NAVIGATOR_PRE_RENAME_BARRIER mid-rename).
//   - AC-NS2-012 — PostToolUse never blocks (panic + bounded-deadline safety
//     via runNavigatorDetectSafe).
//
// Concurrency note (CLAUDE.local.md §2 [WARN] + internal/hook/CLAUDE.md): the
// atomic-read test uses the process-global NAVIGATOR_PRE_RENAME_BARRIER env
// var (unconditionally unset inside M0's atomicWrite at
// internal/navigator/sync/write.go:42), so it MUST NOT run in parallel with
// any other test that touches the barrier — it is deliberately serial.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// --- AC-NS2-004 — fail-open across 5 error modes (table-driven) ---
//
// Each row asserts: (1) the Detect branch returns cleanly (nil or empty
// Result), (2) no panic propagates, (3) the surrounding tool call would
// proceed (the fail-open contract per REQ-NS2-004 / REQ-NS2-012).

func TestDetectForChangedPath_FailOpenTable(t *testing.T) {
	// Serial: the timeout sub-case constructs a pre-cancelled context whose
	// machinery (runNavigatorDetectSafe) is fine in parallel, but keeping the
	// whole table serial makes the barrier-sharing atomic-read test (which
	// runs in this same -race invocation) easier to reason about.
	cases := []struct {
		name string
		// setup returns the graphPath to pass to detectForChangedPath and a
		// changedPath (or "" to trigger a traversal error).
		setup func(t *testing.T, tmpDir string) (graphPath, changedPath string)
		// expectNilResult: true when the branch MUST return nil; false when
		// an empty (non-nil) Result is also acceptable (advisory fail-open
		// at the edge level — malformed edges are skipped, not fatal).
		expectNilResult bool
	}{
		{
			name: "004a_graph_absent",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
				// Deliberately do NOT write the graph — absent.
				return graphPath, filepath.Join(tmpDir, "x.go")
			},
			expectNilResult: true,
		},
		{
			name: "004b_unparseable_json",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
				if err := os.MkdirAll(filepath.Dir(graphPath), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(graphPath, []byte("{not json"), 0o644); err != nil {
					t.Fatalf("write bad graph: %v", err)
				}
				return graphPath, filepath.Join(tmpDir, "x.go")
			},
			expectNilResult: true,
		},
		{
			name: "004c_schema_invalid_missing_edges_array",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
				if err := os.MkdirAll(filepath.Dir(graphPath), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				// Valid JSON, parses into sync.Graph, but the `edges` array is
				// ABSENT — schema-invalid per AC-NS2-004 row 004c. (An explicit
				// `"edges":[]` would be a valid empty graph — §F edge case.)
				if err := os.WriteFile(graphPath, []byte(`{"provenance":{"extract_commit_sha":"x","captured_at":"x"},"nodes":[]}`), 0o644); err != nil {
					t.Fatalf("write schema-invalid graph: %v", err)
				}
				return graphPath, filepath.Join(tmpDir, "x.go")
			},
			expectNilResult: true,
		},
		{
			name: "004d_traversal_error_empty_changed_path",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				// A well-formed graph but the changedPath is empty → detect.Traverse
				// returns an error ("empty changed path"), exercising the
				// traversal-error fail-open branch in detectForChangedPath.
				graphPath := writeWellFormedGraph(t, tmpDir)
				return graphPath, ""
			},
			expectNilResult: true,
		},
		{
			name: "004e_timeout_pre_cancelled_context",
			setup: func(t *testing.T, tmpDir string) (string, string) {
				// This row is verified separately by
				// TestRunNavigatorDetectSafe_PreCancelledContext (the timeout
				// machinery lives in the safe wrapper, not detectForChangedPath).
				// Placeholder entry keeps the table documentation complete.
				graphPath := writeWellFormedGraph(t, tmpDir)
				return graphPath, filepath.Join(tmpDir, "x.go")
			},
			expectNilResult: false, // detectForChangedPath itself is not timeout-aware
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			graphPath, changedPath := tc.setup(t, tmpDir)

			// AC-NS2-004: the branch returns cleanly — no panic.
			got := detectForChangedPath(graphPath, changedPath) //nolint:errcheck // purely advisory

			if tc.expectNilResult && got != nil {
				t.Errorf("expected nil Result for %s (fail-open); got %+v", tc.name, got)
			}
		})
	}
}

// writeWellFormedGraph writes a minimal but structurally-valid nav-graph.json
// (non-nil edges array with one well-formed edge) so the 004d/004e rows can
// isolate the fail-open trigger to their respective dimensions.
func writeWellFormedGraph(t *testing.T, tmpDir string) string {
	t.Helper()
	graph := navsync.Graph{
		Provenance: navsync.Provenance{ExtractCommitSHA: "fixture-sha", CapturedAt: "2026-08-06"},
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "pkg.Foo", DisplayName: "Foo"},
		},
		Edges: []navsync.Edge{
			{EdgeType: navsync.EdgeSym, SourceNode: "symbol:pkg.Foo", TargetNode: "spec:SPEC-X",
				SourcePath: filepath.Join(tmpDir, "foo.go"), LineNumber: 1},
		},
	}
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal graph: %v", err)
	}
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
	if err := os.MkdirAll(filepath.Dir(graphPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(graphPath, raw, 0o644); err != nil {
		t.Fatalf("write graph: %v", err)
	}
	return graphPath
}

// --- AC-NS2-004 row 004e — timeout / context cancellation (via the safe
// wrapper runNavigatorDetectSafe). ---

// TestRunNavigatorDetectSafe_PreCancelledContext verifies the bounded-deadline
// wrapper returns nil silently when the parent context is already cancelled
// (simulating mid-traversal cancellation per AC-NS2-004 row 004e's "or a test
// that cancels the context mid-traversal" alternative).
func TestRunNavigatorDetectSafe_PreCancelledContext(t *testing.T) {
	// NOT parallel — relies on t.TempDir isolation but kept serial for
	// determinism alongside the barrier-sharing atomic-read test in the same
	// -race invocation.
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel: dctx derived inside the wrapper is already Done.

	input := &HookInput{
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path":` + mustJSON(specPath) + `}`),
	}

	start := time.Now()
	got := runNavigatorDetectSafe(ctx, input)
	elapsed := time.Since(start)

	if got != nil {
		t.Errorf("expected nil Result on pre-cancelled context (AC-NS2-004 004e); got %+v", got)
	}
	// The wrapper MUST NOT block waiting for the goroutine when the context
	// is already cancelled — return should be near-instantaneous.
	if elapsed > navigatorDetectTimeout {
		t.Errorf("wrapper blocked %s past the %s budget on a pre-cancelled context (must return immediately)",
			elapsed, navigatorDetectTimeout)
	}
}

// TestRunNavigatorDetectSafe_RecoversFromPanic verifies the deferred recover
// swallows any panic inside the branch (REQ-NS2-012 — PostToolUse never
// blocks). A malicious HookInput that would panic the raw runNavigatorDetect
// is contained.
func TestRunNavigatorDetectSafe_RecoversFromPanic(t *testing.T) {
	// Construct a HookInput whose ToolInput is valid JSON but whose
	// file_path is a value that forces the branch deep enough to exercise
	// the recover path. The simplest reliable panic trigger is a nil ctx
	// passed to context.WithTimeout — but that panics inside the wrapper
	// itself, which is exactly the path the defer recover guards.
	tmpDir := t.TempDir()

	// We deliberately pass a nil context to force context.WithTimeout to
	// panic. The recover MUST swallow it and return nil.
	got := runNavigatorDetectSafe(nil, &HookInput{ //nolint:staticcheck // intentional nil context — verifies the recover path
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path":` + mustJSON(filepath.Join(tmpDir, "x.go")) + `}`),
	})
	if got != nil {
		t.Errorf("expected nil Result after panic recovery (REQ-NS2-012); got %+v", got)
	}
}

// TestRunNavigatorDetectSafe_PassesThroughNonCancel verifies the wrapper is a
// no-op pass-through when the context is healthy and the work completes within
// the budget — the advisory surfaces still fire.
func TestRunNavigatorDetectSafe_PassesThroughNonCancel(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_, specPath, _ := writeNavGraphFixture(t, tmpDir)

	got := runNavigatorDetectSafe(context.Background(), &HookInput{
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: json.RawMessage(`{"file_path":` + mustJSON(specPath) + `}`),
	})
	if got == nil {
		t.Fatal("expected non-nil Result for healthy-context pass-through; got nil")
	}
	if len(got.Edges) == 0 {
		t.Errorf("expected affected edges for matching path on pass-through; got 0")
	}
}

// --- AC-NS2-006 — atomic read during regen (concurrency test) ---

// TestDetectForChangedPath_AtomicReadDuringRegen verifies that a reader
// goroutine calling detectForChangedPath observes the PRIOR committed graph
// while the M0 writer (navsync.WriteGraph) is held at the
// NAVIGATOR_PRE_RENAME_BARRIER mid-rename. The reader never sees a partial
// file or the not-yet-renamed .tmp — it sees either the prior graph or, after
// the rename lands, the new graph.
//
// Run with `go test -race ./internal/hook/` — the race detector guards
// against any hidden shared-mutable-state defect in the read path.
func TestDetectForChangedPath_AtomicReadDuringRegen(t *testing.T) {
	// Serial: NAVIGATOR_PRE_RENAME_BARRIER is process-global (unset inside
	// atomicWrite at internal/navigator/sync/write.go:42).
	tmpDir := t.TempDir()
	graphPath := filepath.Join(tmpDir, ".moai", "project", "navigator", "nav-graph.json")
	if err := os.MkdirAll(filepath.Dir(graphPath), 0o755); err != nil {
		t.Fatalf("mkdir graph dir: %v", err)
	}

	priorEdgePath := filepath.Join(tmpDir, "a.go")
	newEdgePath := filepath.Join(tmpDir, "b.go")

	// Prior committed graph: one edge on priorEdgePath.
	priorGraph := navsync.Graph{
		Provenance: navsync.Provenance{ExtractCommitSHA: "prior-sha", CapturedAt: "2026-08-05"},
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "prior.Foo", DisplayName: "Foo"},
			{EntityType: navsync.EntitySpec, Identifier: "SPEC-PRIOR", DisplayName: "Prior"},
		},
		Edges: []navsync.Edge{{
			EdgeType: navsync.EdgeSym, SourceNode: "symbol:prior.Foo", TargetNode: "spec:SPEC-PRIOR",
			SourcePath: priorEdgePath, LineNumber: 1,
		}},
	}
	priorRaw, err := json.Marshal(priorGraph)
	if err != nil {
		t.Fatalf("marshal prior graph: %v", err)
	}
	if err := os.WriteFile(graphPath, priorRaw, 0o644); err != nil {
		t.Fatalf("write prior graph: %v", err)
	}

	// New graph the writer is trying to put in place: one edge on newEdgePath.
	newGraph := navsync.Graph{
		Provenance: navsync.Provenance{ExtractCommitSHA: "new-sha", CapturedAt: "2026-08-06"},
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "new.Bar", DisplayName: "Bar"},
			{EntityType: navsync.EntitySpec, Identifier: "SPEC-NEW", DisplayName: "New"},
		},
		Edges: []navsync.Edge{{
			EdgeType: navsync.EdgeSym, SourceNode: "symbol:new.Bar", TargetNode: "spec:SPEC-NEW",
			SourcePath: newEdgePath, LineNumber: 2,
		}},
	}

	// Arm the barrier. The writer will create .tmp, write "ready" to the
	// barrier file, then spin-wait until we remove it.
	barrier := filepath.Join(tmpDir, "barrier-marker")
	if err := os.Setenv("NAVIGATOR_PRE_RENAME_BARRIER", barrier); err != nil {
		t.Fatalf("setenv barrier: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("NAVIGATOR_PRE_RENAME_BARRIER") })

	writerDone := make(chan error, 1)
	go func() {
		writerDone <- navsync.WriteGraph(graphPath, newGraph)
	}()

	// Wait until the writer reaches the barrier (the barrier file appears
	// after .tmp is written and the writer is now spin-waiting pre-rename).
	if !waitForFile(barrier, 2*time.Second) {
		t.Fatalf("writer did not reach barrier within timeout; barrier file %s absent", barrier)
	}

	// AC-NS2-006 — the reader runs WHILE the writer holds at the rename
	// barrier. At this instant graphPath still holds the PRIOR graph and
	// <graphPath>.tmp holds the NEW graph (not yet renamed). The reader
	// MUST observe the prior graph — never a partial file, never the .tmp.
	result := detectForChangedPath(graphPath, priorEdgePath)

	// Release the barrier so the writer can rename and the goroutine exits.
	if err := os.Remove(barrier); err != nil {
		t.Fatalf("remove barrier to release writer: %v", err)
	}
	select {
	case err := <-writerDone:
		if err != nil {
			t.Fatalf("writer failed after barrier release: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("writer did not complete within 2s of barrier release")
	}

	if result == nil {
		t.Fatal("expected non-nil Result from reader (prior graph); got nil — " +
			"reader observed neither prior nor new graph correctly")
	}
	// The reader MUST have observed the prior graph (the prior.Foo edge).
	foundPrior := false
	for _, n := range result.Nodes {
		if n.EntityType == navsync.EntitySymbol && n.Identifier == "prior.Foo" {
			foundPrior = true
		}
	}
	if !foundPrior {
		t.Errorf("reader MUST observe the prior graph during held rename (AC-NS2-006); got nodes=%+v",
			result.Nodes)
	}
	// The reader MUST NOT have observed the new graph (no new.Bar) mid-rename.
	for _, n := range result.Nodes {
		if n.EntityType == navsync.EntitySymbol && n.Identifier == "new.Bar" {
			t.Errorf("reader observed the NEW graph mid-rename (atomic-rename violation); "+
				"nodes=%+v", result.Nodes)
			break
		}
	}
	// The reader MUST NOT have read a partial/zero-length file — at least
	// one edge is present.
	if len(result.Edges) == 0 {
		t.Errorf("reader observed an empty/partial graph (zero edges); expected the prior graph's edge")
	}
}

// waitForFile polls for path existence up to the deadline. Returns true if the
// file appears, false on timeout.
func waitForFile(path string, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(2 * time.Millisecond)
	}
	return false
}

// --- AC-NS2-012 — never blocks (grep guard) ---
//
// The Detect source MUST NOT contain a `Decision: "block"` or `os.Exit(2)`
// pattern. (This complements the M1.3 TestNavigatorDetect_NoWorkItemPromotion
// grep; M1.4 restates it for the hardening additions.)

func TestNavigatorDetect_NeverBlocks_GrepGuard(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("navigator_detect.go")
	if err != nil {
		t.Fatalf("read navigator_detect.go: %v", err)
	}
	body := string(raw)
	for _, pat := range []string{
		`Decision: "block"`,
		"os.Exit(2)",
		`"block"`,
	} {
		if strings.Contains(body, pat) {
			t.Errorf("forbidden blocking pattern %q found in navigator_detect.go (REQ-NS2-012)", pat)
		}
	}
}

// Compile-time assertion that the safe wrapper exists and has the expected
// signature. A rename/refactor of runNavigatorDetectSafe would break this at
// compile time, locking the M1.4 wiring contract.
var _ = runNavigatorDetectSafe

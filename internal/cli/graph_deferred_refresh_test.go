package cli

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/mx"
)

// SPEC-GRAPH-REPORT-001 REQ-GR-010/011/012 (AC-GR-015) — the cli-injected
// wrapper end-to-end: a SessionStart over a stale tree, run in
// synchronous-deferred mode with the DeferredEdgesRefresh seam wired to the
// wrapper around refreshEdgesArtifact, refreshes the default edges artifact
// (staleness predicate flips false) and stages nothing in git; a fresh tree
// performs zero writes; the duration is measured through the edgesRefreshClock
// seam and an over-budget duration produces the same warning-only signal as
// the query-time refresh.

// initFixtureGitRepo turns the fixture root into a git repository so the
// staged-entry assertion can read `git status --porcelain`.
func initFixtureGitRepo(t *testing.T, root string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}
}

// stagedPorcelainLines returns the porcelain lines representing STAGED entries
// (index status in column 1, excluding untracked '??').
func stagedPorcelainLines(t *testing.T, root string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git status --porcelain: %v (%s)", err, out)
	}
	var staged []string
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) >= 2 && line[0] != ' ' && line[0] != '?' {
			staged = append(staged, line)
		}
	}
	return staged
}

// AC-GR-015 stale leg: the wrapper (through the full Handle → deferred-step →
// injected-seam path) rebuilds the stale default artifact so both staleness
// predicates read false afterward, and nothing lands in the git index.
func TestDeferredEdgesRefresh_StaleRefreshesAndStagesNothing(t *testing.T) {
	root := graphFixtureProject(t)
	initFixtureGitRepo(t, root)
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	if _, err := os.Stat(edgesFile); !os.IsNotExist(err) {
		t.Fatalf("precondition: fixture must start without an edges artifact (err: %v)", err)
	}

	h := hook.NewSessionStartHandler(nil,
		hook.WithSynchronousDeferredScans(),
		hook.WithDeferredEdgesRefresh(deferredEdgesRefresh),
	)
	if _, err := h.Handle(context.Background(), &hook.HookInput{
		SessionID:     "deferred-edges-stale",
		CWD:           root,
		ProjectDir:    root,
		HookEventName: "SessionStart",
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if _, err := os.Stat(edgesFile); err != nil {
		t.Fatalf("stale tree: the deferred step must have written the edges artifact: %v", err)
	}
	if edgesRefreshNeeded(root, edgesFile, graph.DefaultThresholds().MXIndexChangedFiles) {
		t.Error("staleness predicate must read false after the deferred refresh (REQ-GR-010)")
	}
	if staged := stagedPorcelainLines(t, root); len(staged) != 0 {
		t.Errorf("REQ-GR-011: no entry may be staged by the deferred refresh, staged: %v", staged)
	}
}

// graphBuildFixture runs `graph build` on an EXISTING fixture root (the
// buildEdgesFixture helper above always creates its own).
func graphBuildFixture(t *testing.T, root string) error {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SetArgs([]string{"build", "--root", root})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

// AC-GR-015 fresh leg: a fresh edges layer is left untouched — the artifact
// bytes AND mtime are identical across the session start.
func TestDeferredEdgesRefresh_FreshNoRewrite(t *testing.T) {
	root := graphFixtureProject(t)
	// Fresh prep (same as the selected-artifact test): build, index the mx
	// sidecar (its appearance moves the mx-index fingerprint), then re-stamp
	// the meta over the now-current fingerprints.
	if err := graphBuildFixture(t, root); err != nil {
		t.Fatalf("graph build: %v", err)
	}
	if _, err := mx.RefreshIndex(filepath.Join(root, ".moai", "state"), root, nil); err != nil {
		t.Fatalf("refresh mx index: %v", err)
	}
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	if err := graph.WriteEdgesMeta(filepath.Join(filepath.Dir(edgesFile), graph.MetaFileName),
		root, graph.SourceFingerprintsForEdges(root), 0); err != nil {
		t.Fatal(err)
	}
	if edgesRefreshNeeded(root, edgesFile, graph.DefaultThresholds().MXIndexChangedFiles) {
		t.Fatal("precondition: fixture must probe fresh on both predicates")
	}

	before, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	beforeSum := fmt.Sprintf("%x", sha256.Sum256(before))
	beforeInfo, err := os.Stat(edgesFile)
	if err != nil {
		t.Fatal(err)
	}

	h := hook.NewSessionStartHandler(nil,
		hook.WithSynchronousDeferredScans(),
		hook.WithDeferredEdgesRefresh(deferredEdgesRefresh),
	)
	if _, err := h.Handle(context.Background(), &hook.HookInput{
		SessionID:     "deferred-edges-fresh",
		CWD:           root,
		ProjectDir:    root,
		HookEventName: "SessionStart",
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	after, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(after)); got != beforeSum {
		t.Error("fresh artifact must not be rewritten by the deferred step (SHA changed)")
	}
	afterInfo, err := os.Stat(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Error("fresh artifact must not be rewritten by the deferred step (mtime changed)")
	}
}

// REQ-GR-012: the deferred wrapper measures through the edgesRefreshClock seam
// and emits the same warning-only budget-overrun signal as the query-time
// refresh — deterministic via the injected duration, no real-timing dependence.
func TestDeferredEdgesRefresh_BudgetOverrunWarns(t *testing.T) {
	root := graphFixtureProject(t)

	origClock := edgesRefreshClock
	edgesRefreshClock = func() func() time.Duration {
		return func() time.Duration { return 50 * time.Millisecond }
	}
	defer func() { edgesRefreshClock = origClock }()

	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gateYAML := "gate:\n  graph_freshness:\n    enabled: true\n    update_budget_ms: 1\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "gate.yaml"), []byte(gateYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Capture stderr around the direct wrapper call (non-parallel test).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	callErr := deferredEdgesRefresh(root)
	if err := w.Close(); err != nil {
		t.Fatalf("close stderr pipe writer: %v", err)
	}
	os.Stderr = oldStderr
	captured, _ := io.ReadAll(r)

	if callErr != nil {
		t.Fatalf("overrun must not fail the deferred refresh: %v", callErr)
	}
	if !strings.Contains(string(captured), "update budget") {
		t.Errorf("overrun warning must name the budget, stderr: %q", captured)
	}
}

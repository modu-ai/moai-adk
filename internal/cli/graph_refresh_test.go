package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// buildEdgesFixture runs the build on the standard fixture and returns root.
func buildEdgesFixture(t *testing.T) string {
	t.Helper()
	root := graphFixtureProject(t)
	cmd := newGraphCmd()
	cmd.SetArgs([]string{"build", "--root", root})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	return root
}

// AC-GF-008 (graph side) / MUTANT B guard — a query after a source edit must
// answer from a REBUILT artifact whose content reflects the edit, not merely
// re-stamp the old one.
func TestGraphQuery_RefreshesStaleEdges(t *testing.T) {
	root := buildEdgesFixture(t)
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	before, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}

	// Move a SOURCE the edges build consumes: add a tagged @MX:SPEC sub-line
	// (a new mx-spec edge the old artifact cannot know about).
	src := filepath.Join(root, "internal", "demo", "extra.go")
	if err := os.WriteFile(src, []byte("package demo\n\n// @MX:NOTE: [AUTO] extra\n// @MX:SPEC:SPEC-GRAPH-CLI-001\nfunc Extra() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"query", "--callers", "SPEC-GRAPH-CLI-001", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("query after source edit: %v (stderr: %s)", err, errOut.String())
	}

	after, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Fatal("MUTANT B (graph side): edges artifact unchanged after a source moved — stamp-only refresh")
	}
	if !strings.Contains(string(after), "demo/extra.go") {
		t.Error("refreshed artifact must carry the new source file's mx-spec edge")
	}
	if !strings.Contains(out.String(), "demo/extra.go") {
		t.Error("the answer must reflect the refreshed artifact (extra.go as caller)")
	}
	// REQ-GF-008: the answer names the tree it read from.
	if !strings.Contains(errOut.String(), "provenance: tree=") {
		t.Errorf("answer must carry provenance naming, stderr was: %q", errOut.String())
	}
}

// A fresh artifact needs no rebuild: the query answers without touching the
// artifact bytes (mtime-independence is implicit; the artifact is byte-equal).
func TestGraphQuery_NoRefreshWhenFresh(t *testing.T) {
	root := buildEdgesFixture(t)
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	before, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"query", "--callers", "SPEC-GRAPH-CLI-001", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("fresh artifact must not be rewritten by a query")
	}
}

// AC-GF-010 — per-tree cache isolation: two copies of the fixture with
// DIFFERENT uncommitted edits answer differently, and each answer names its
// own tree. The substrate (mx sidecar + edges meta) is per-tree state, so a
// shared cache would have to be actively built to fail here — this pins the
// structural property.
func TestGraphQuery_PerTreeIsolation(t *testing.T) {
	treeA := buildEdgesFixture(t)
	treeB := buildEdgesFixture(t)

	// Distinct edits: A gains extra_a.go, B gains extra_b.go with a different
	// SPEC target.
	if err := os.WriteFile(filepath.Join(treeA, "internal", "demo", "extra_a.go"),
		[]byte("package demo\n\n// @MX:NOTE: [AUTO] a\n// @MX:SPEC:SPEC-GRAPH-CLI-001\nfunc ExtraA() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(treeB, "internal", "demo", "extra_b.go"),
		[]byte("package demo\n\n// @MX:NOTE: [AUTO] b\n// @MX:SPEC:SPEC-OTHER-999\nfunc ExtraB() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runQ := func(root, node string) (string, string) {
		cmd := newGraphCmd()
		cmd.SetArgs([]string{"query", "--callers", node, "--root", root})
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("query in %s: %v", root, err)
		}
		return out.String(), errOut.String()
	}

	outA, provA := runQ(treeA, "SPEC-GRAPH-CLI-001")
	outB, provB := runQ(treeB, "SPEC-OTHER-999")

	if !strings.Contains(outA, "extra_a.go") {
		t.Errorf("tree A's answer must reflect its own edit:\n%s", outA)
	}
	if !strings.Contains(outB, "extra_b.go") {
		t.Errorf("tree B's answer must reflect its own edit:\n%s", outB)
	}
	if !strings.Contains(provA, treeA) {
		t.Errorf("tree A's provenance must name tree A: %q", provA)
	}
	if !strings.Contains(provB, treeB) {
		t.Errorf("tree B's provenance must name tree B: %q", provB)
	}
}

// AC-GF-011 — update-cost budget warning: a budget configured below the
// measured refresh duration triggers a warning naming both, and the answer
// still arrives (exit reflects the query, not the overrun).
// CR round-2 3855149237: the duration arrives through the injection seam —
// a FIXED 50ms — so the overrun fires deterministically; the test no longer
// relies on "any real refresh exceeds the 1ms budget".
func TestGraphQuery_BudgetOverrunWarns(t *testing.T) {
	root := buildEdgesFixture(t)

	// Inject the deterministic duration BEFORE anything can read the seam.
	origClock := edgesRefreshClock
	edgesRefreshClock = func() func() time.Duration {
		return func() time.Duration { return 50 * time.Millisecond }
	}
	defer func() { edgesRefreshClock = origClock }()

	// A 1ms budget — the injected 50ms deterministically exceeds it.
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gateYAML := "gate:\n  graph_freshness:\n    enabled: true\n    update_budget_ms: 1\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "gate.yaml"), []byte(gateYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Force a refresh by moving a source.
	if err := os.WriteFile(filepath.Join(root, "internal", "demo", "budget.go"),
		[]byte("package demo\n\n// @MX:NOTE: [AUTO] budget\nfunc B() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"query", "--callers", "SPEC-GRAPH-CLI-001", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("overrun must not fail the query: %v", err)
	}
	if !strings.Contains(errOut.String(), "update budget") {
		t.Errorf("overrun warning must name the budget, stderr: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "callers of") {
		t.Errorf("the answer must still arrive, got: %q", out.String())
	}
}

// AC-GFR-007 default-construction check — the UN-injected seam constructs
// the wall-clock measurer, so production CLI behavior is unchanged by the
// seam's existence: the package var IS newEdgesRefreshClock, and a measurer
// it constructs reads durations that advance with real time.
func TestEdgesRefreshClockDefaultIsWallClock(t *testing.T) {
	if got, want := fmt.Sprintf("%p", edgesRefreshClock), fmt.Sprintf("%p", newEdgesRefreshClock); got != want {
		t.Errorf("edgesRefreshClock = %s, want the wall-clock constructor newEdgesRefreshClock (%s) — production must measure wall-clock", got, want)
	}
	elapsed := newEdgesRefreshClock()
	first := elapsed()
	time.Sleep(2 * time.Millisecond)
	second := elapsed()
	if second < first || second <= 0 {
		t.Errorf("wall-clock measurer must advance monotonically with real time: first=%s second=%s", first, second)
	}
}

// EdgesSourcesMoved sanity — the probe agrees with the meta it reads.
func TestEdgesSourcesMovedProbe(t *testing.T) {
	root := buildEdgesFixture(t)
	if graph.EdgesSourcesMoved(root) {
		t.Fatal("freshly built fixture must probe as not-moved")
	}
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "dependencies.md"),
		[]byte("# moved\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !graph.EdgesSourcesMoved(root) {
		t.Fatal("a moved codemaps source must probe as moved")
	}
}

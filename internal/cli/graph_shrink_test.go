//go:build cgo

package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// SPEC-GRAPH-REPORT-001 REQ-GR-008/009 (AC-GR-012/013/014) — the shrink guard
// on every automatic write path: the query-time refresh refuses the overwrite
// and answers from the existing artifact (exit 0, warning on stderr); the
// build path exits non-zero naming the removed edges; a genuine deletion
// proceeds; the deferred path inherits the refusal through the M4-wrapped
// refreshEdgesArtifact. Refusal means ZERO writes — edges.jsonl and its meta
// sidecar stay byte-identical (SHA-compared).

// shrinkCallsFixture extends graphFixtureProject with a real call-bearing Go
// file under a described root so extraction produces genuine code-call and
// code-import edges with real `file:function` Source shapes.
func shrinkCallsFixture(t *testing.T) string {
	t.Helper()
	root := graphFixtureProject(t)
	if err := os.WriteFile(filepath.Join(root, "internal", "demo", "calls.go"),
		[]byte("package demo\n\nimport \"fmt\"\n\nfunc Calls() { fmt.Println(Helper()) }\n\nfunc Helper() string { return \"x\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// injectUnscannedEdge models the partial-failure state the guard exists to
// catch: an edge in the EXISTING artifact whose source file exists on disk
// but lies outside the extraction walk's scanned set (rootlevel.go is at the
// project root — the walk covers internal/, cmd/, pkg/ only). The edge uses
// the real compound code-call Source shape; only its presence is injected.
func injectUnscannedEdge(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "rootlevel.go"),
		[]byte("package rootlevel\n\nfunc Fn() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	before, err := os.ReadFile(edgesFile)
	if err != nil {
		t.Fatal(err)
	}
	injected := map[string]string{"kind": "code-call", "source": "rootlevel.go:Fn", "target": "InjectedTarget"}
	line, err := json.Marshal(injected)
	if err != nil {
		t.Fatal(err)
	}
	with := append(append(bytes.Clone(before), line...), '\n')
	if err := os.WriteFile(edgesFile, with, 0o644); err != nil {
		t.Fatal(err)
	}
}

// fileSHA returns the hex sha256 of a file's current bytes.
func fileSHA(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// runGraphBuildOn executes 'moai graph build' on root and returns the error.
func runGraphBuildOn(t *testing.T, root string) error {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SetArgs([]string{"build", "--root", root})
	cmd.SetOut(io_discard())
	cmd.SetErr(io_discard())
	return cmd.Execute()
}

func io_discard() *bytes.Buffer { return &bytes.Buffer{} }

// loadEdgeIDs loads an edges.jsonl keyed by kind+source+target.
func loadEdgeIDs(t *testing.T, path string) map[string]graph.Edge {
	t.Helper()
	edges, err := graph.LoadJSONL(path)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]graph.Edge{}
	for _, e := range edges {
		out[e.Kind+"\x00"+e.Source+"\x00"+e.Target] = e
	}
	return out
}

// sourceFileOf decodes an edge's source file part (the guard's own decode).
func sourceFileOf(e graph.Edge) string {
	if e.Kind == graph.KindCodeCall {
		if i := strings.LastIndex(e.Source, ":"); i > 0 {
			return e.Source[:i]
		}
	}
	return e.Source
}

// AC-GR-012 second clause — the BUILD path: given the shrink condition, an
// explicit `moai graph build` exits non-zero naming the removed edges, with
// the prior artifact and meta sidecar SHA-identical (zero writes).
func TestShrinkGuard_BuildPathRefuses(t *testing.T) {
	root := shrinkCallsFixture(t)
	if err := runGraphBuildOn(t, root); err != nil {
		t.Fatalf("initial build: %v", err)
	}
	injectUnscannedEdge(t, root)

	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	metaFile := filepath.Join(root, ".moai", "project", "graph", graph.MetaFileName)
	edgesSHA, metaSHA := fileSHA(t, edgesFile), fileSHA(t, metaFile)

	err := runGraphBuildOn(t, root)
	if err == nil {
		t.Fatal("build over an unexplained shrink must exit non-zero")
	}
	for _, want := range []string{"rootlevel.go:Fn", "InjectedTarget", "rootlevel.go"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("build refusal must name %q, got: %v", want, err)
		}
	}
	if got := fileSHA(t, edgesFile); got != edgesSHA {
		t.Error("refused build must leave edges.jsonl byte-identical")
	}
	if got := fileSHA(t, metaFile); got != metaSHA {
		t.Error("refused build must leave the meta sidecar byte-identical")
	}
}

// AC-GR-012 first clause / AC-GR-014 — the QUERY-time refresh: the refusal
// skips the write (prior artifact byte-identical), a stated shrink-refusal
// warning lands on stderr naming the removed edge, the answer comes from the
// EXISTING artifact, and the command exits 0.
func TestShrinkGuard_QueryPathRefusedAnswersFromExisting(t *testing.T) {
	root := shrinkCallsFixture(t)
	if err := runGraphBuildOn(t, root); err != nil {
		t.Fatalf("initial build: %v", err)
	}
	injectUnscannedEdge(t, root)

	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	metaFile := filepath.Join(root, ".moai", "project", "graph", graph.MetaFileName)
	edgesSHA, metaSHA := fileSHA(t, edgesFile), fileSHA(t, metaFile)

	// Make the artifact stale so the query path attempts a refresh.
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "dependencies.md"),
		[]byte("# moved\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"query", "--callers", "InjectedTarget", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("a refused refresh must not fail the query (exit 0), got: %v", err)
	}
	for _, want := range []string{"unexplained shrink", "rootlevel.go:Fn"} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("stderr must carry the shrink-refusal warning naming %q, stderr: %q", want, errOut.String())
		}
	}
	if !strings.Contains(out.String(), "callers of InjectedTarget: 1") ||
		!strings.Contains(out.String(), "rootlevel.go:Fn") {
		t.Errorf("the answer must come from the EXISTING artifact (injected edge visible), stdout: %q", out.String())
	}
	if got := fileSHA(t, edgesFile); got != edgesSHA {
		t.Error("refused refresh must leave edges.jsonl byte-identical")
	}
	if got := fileSHA(t, metaFile); got != metaSHA {
		t.Error("refused refresh must leave the meta sidecar byte-identical")
	}
}

// AC-GR-013 — genuine deletion: a source file removed from disk drops its
// edges legitimately; the overwrite proceeds and the artifact shrinks by
// exactly the departed source's edges.
func TestShrinkGuard_GenuineDeletionProceeds(t *testing.T) {
	root := shrinkCallsFixture(t)
	if err := runGraphBuildOn(t, root); err != nil {
		t.Fatalf("initial build: %v", err)
	}
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	before := loadEdgeIDs(t, edgesFile)

	// The departed source's edges: every edge whose decoded source file is
	// internal/demo/calls.go (one code-call, one code-import).
	var departed int
	for _, e := range before {
		if sourceFileOf(e) == "internal/demo/calls.go" {
			departed++
		}
	}
	if departed == 0 {
		t.Fatal("fixture precondition: calls.go must contribute at least one edge")
	}

	if err := os.Remove(filepath.Join(root, "internal", "demo", "calls.go")); err != nil {
		t.Fatal(err)
	}
	if err := runGraphBuildOn(t, root); err != nil {
		t.Fatalf("a genuine deletion must proceed, got: %v (err must be nil)", err)
	}

	after := loadEdgeIDs(t, edgesFile)
	removed, added := 0, 0
	for id := range before {
		if _, ok := after[id]; !ok {
			removed++
			if sourceFileOf(before[id]) != "internal/demo/calls.go" {
				t.Errorf("removed edge not sourced from the deleted file: %+v", before[id])
			}
		}
	}
	for id := range after {
		if _, ok := before[id]; !ok {
			added++
			if sourceFileOf(after[id]) == "internal/demo/calls.go" {
				t.Errorf("edge added from a deleted file: %+v", after[id])
			}
		}
	}
	if removed != departed {
		t.Errorf("artifact must shrink by exactly the departed source's edges: removed %d, departed %d", removed, departed)
	}
	if added != 0 {
		t.Errorf("a pure deletion must add no edges, added %d", added)
	}
}

// REQ-GR-009 — the deferred path inherits the guard through the M4-wrapped
// refreshEdgesArtifact (wrapped, never forked): the wrapper returns the typed
// refusal and writes nothing.
func TestShrinkGuard_DeferredPathInheritsRefusal(t *testing.T) {
	root := shrinkCallsFixture(t)
	if err := runGraphBuildOn(t, root); err != nil {
		t.Fatalf("initial build: %v", err)
	}
	injectUnscannedEdge(t, root)

	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	metaFile := filepath.Join(root, ".moai", "project", "graph", graph.MetaFileName)
	edgesSHA, metaSHA := fileSHA(t, edgesFile), fileSHA(t, metaFile)

	err := deferredEdgesRefresh(root)
	var refuse *graph.ShrinkRefusalError
	if !errors.As(err, &refuse) {
		t.Fatalf("the deferred wrapper must surface the typed shrink refusal, got: %v", err)
	}
	if !strings.Contains(refuse.Error(), "rootlevel.go") {
		t.Errorf("refusal must name the unscanned source, got: %v", refuse)
	}
	if got := fileSHA(t, edgesFile); got != edgesSHA {
		t.Error("refused deferred refresh must leave edges.jsonl byte-identical")
	}
	if got := fileSHA(t, metaFile); got != metaSHA {
		t.Error("refused deferred refresh must leave the meta sidecar byte-identical")
	}
}

// REQ-GR-008 seam proof — the scanned set the guard consumes is the extraction
// walk's own processed-file list: described-root files present, a root-level Go
// file (outside the described roots) absent even though it exists on disk.
func TestShrinkGuard_ScannedSetIsExtractionWalk(t *testing.T) {
	root := shrinkCallsFixture(t)
	if err := os.WriteFile(filepath.Join(root, "rootlevel.go"),
		[]byte("package rootlevel\n\nfunc Fn() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, scanned, _, err := graph.BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	set := map[string]bool{}
	for _, f := range scanned {
		set[f] = true
	}
	for _, want := range []string{"internal/demo/calls.go", "internal/demo/demo.go"} {
		if !set[want] {
			t.Errorf("scanned set must contain the described-root file %q, got: %v", want, scanned)
		}
	}
	if set["rootlevel.go"] {
		t.Error("scanned set must NOT contain the root-level file outside the described roots")
	}
}

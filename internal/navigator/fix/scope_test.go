package fix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// --- test-graph builders ---

func symNode(id string) navsync.Node {
	return navsync.Node{EntityType: navsync.EntitySymbol, Identifier: id, DisplayName: id}
}

func specNode(id string) navsync.Node {
	return navsync.Node{EntityType: navsync.EntitySpec, Identifier: id, DisplayName: id}
}

// selfSymEdge creates a sym-edge whose source_path is sourcePath and whose
// both endpoints are the same symbol node. This models "this path binds this
// symbol's doc subtree" — when the path changes, the symbol's subtree is stale.
func selfSymEdge(nodeKey, sourcePath string) navsync.Edge {
	return navsync.Edge{
		EdgeType:   navsync.EdgeSym,
		SourceNode: nodeKey,
		TargetNode: nodeKey,
		SourcePath: sourcePath,
		LineNumber: 1,
	}
}

func nodeKey(n navsync.Node) string { return string(n.EntityType) + ":" + n.Identifier }

// entryKey is the dedup key for a DiffScopeEntry.
func entryKey(e DiffScopeEntry) string { return e.DocSurface + "|" + e.SubtreeID }

// =====================================================================
// ResolveBaseline — priority order (AC-NS5-003b)
// =====================================================================

func TestResolveBaseline_Priority(t *testing.T) {
	cases := []struct {
		name         string
		compareTo    string
		graphProvSHA string
		headTilde1   string
		wantBase     string
		wantDegr     bool
	}{
		{"flag wins over graph provenance", "flag-sha", "graph-sha", "head1-sha", "flag-sha", false},
		{"graph provenance when no flag", "", "graph-sha", "head1-sha", "graph-sha", false},
		{"HEAD~1 fallback is degraded", "", "", "head1-sha", "head1-sha", true},
		{"all empty → degraded empty baseline", "", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, degraded := ResolveBaseline(tc.compareTo, tc.graphProvSHA, tc.headTilde1)
			if got != tc.wantBase {
				t.Errorf("baseline = %q, want %q", got, tc.wantBase)
			}
			if degraded != tc.wantDegr {
				t.Errorf("degraded = %v, want %v", degraded, tc.wantDegr)
			}
		})
	}
}

// =====================================================================
// ComputeScope — UNION semantics (AC-NS5-003c, the SSOT formula)
// diff_scope = (gitDiffPaths ∪ M1ChangedPaths ∪ M2OwnerPaths) ∩ graphBoundPaths
// =====================================================================

// TestComputeScope_UnionSemantics validates that each of the three input sets
// contributes INDEPENDENTLY — a graph-bound path in git-diff alone, M1 alone,
// or M2 alone each seeds exactly one subtree. This is the UNION (OR) contract;
// an intersection reading would produce ZERO entries here (no path is in >1
// set).
func TestComputeScope_UnionSemantics(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA"), symNode("SymB"), symNode("SymC")},
		Edges: []navsync.Edge{
			selfSymEdge("symbol:SymA", "/proj/a.go"),
			selfSymEdge("symbol:SymB", "/proj/b.go"),
			selfSymEdge("symbol:SymC", "/proj/c.go"),
		},
	}
	gitDiff := []string{"/proj/a.go"} // git-diff-only graph-bound path
	m1 := []string{"/proj/b.go"}      // M1-only graph-bound path
	m2 := []WorkItemRef{              // M2-only graph-bound path
		{SourceKind: "detect", OwnerPath: "/proj/c.go", Action: "regen row"},
	}

	scope := ComputeScope(gitDiff, m1, m2, graph)

	if len(scope) != 3 {
		t.Fatalf("UNION semantics: got %d entries, want 3 (one per independently-contributing input set); entries=%+v", len(scope), scope)
	}

	byKey := make(map[string]DiffScopeEntry, len(scope))
	for _, e := range scope {
		byKey[entryKey(e)] = e
	}

	// Each input set contributed one distinct subtree.
	wantA := byKey["capability-symbols.json|SymA"]
	if wantA.SubtreeID == "" {
		t.Fatalf("missing entry for SymA (git-diff-only seed)")
	}
	if wantA.StaleReason != "git-diff" {
		t.Errorf("SymA stale_reason = %q, want \"git-diff\"", wantA.StaleReason)
	}

	wantB := byKey["capability-symbols.json|SymB"]
	if wantB.SubtreeID == "" {
		t.Fatalf("missing entry for SymB (M1-only seed)")
	}
	if wantB.StaleReason != "m1-detect" {
		t.Errorf("SymB stale_reason = %q, want \"m1-detect\"", wantB.StaleReason)
	}

	wantC := byKey["capability-symbols.json|SymC"]
	if wantC.SubtreeID == "" {
		t.Fatalf("missing entry for SymC (M2-only seed)")
	}
	if wantC.StaleReason != "m2-owner" {
		t.Errorf("SymC stale_reason = %q, want \"m2-owner\"", wantC.StaleReason)
	}
	// M2-seeded entry carries the work_item_ref.
	if wantC.WorkItemRef == nil {
		t.Error("SymC (M2-seeded) should carry a work_item_ref")
	}
	// Non-M2-seeded entries have nil work_item_ref.
	if wantA.WorkItemRef != nil {
		t.Error("SymA (git-diff-only) should have nil work_item_ref")
	}
}

// TestComputeScope_GraphBoundExclusion validates the ONE exclusion: a path NOT
// graph-bound does NOT seed a subtree regardless of which input set it is in.
// This is the `∩ graphBoundPaths` filter — distinct from "not in M1/M2".
func TestComputeScope_GraphBoundExclusion(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA")},
		Edges: []navsync.Edge{selfSymEdge("symbol:SymA", "/proj/a.go")},
	}
	// /proj/unbound.go is in ALL THREE input sets but has no graph edge → excluded.
	gitDiff := []string{"/proj/a.go", "/proj/unbound.go"}
	m1 := []string{"/proj/unbound.go"}
	m2 := []WorkItemRef{{SourceKind: "detect", OwnerPath: "/proj/unbound.go", Action: "x"}}

	scope := ComputeScope(gitDiff, m1, m2, graph)

	if len(scope) != 1 {
		t.Fatalf("graph-bound exclusion: got %d entries, want 1 (unbound.go excluded); entries=%+v", len(scope), scope)
	}
	if scope[0].SubtreeID != "SymA" {
		t.Errorf("expected SymA, got %q", scope[0].SubtreeID)
	}
}

// TestComputeScope_Incremental validates the incremental-not-full-regen
// contract (AC-NS5-003a): a fixture where only 2 of 10 bound subtrees are
// touched yields exactly 2 diff-scope entries, NOT 10.
func TestComputeScope_Incremental(t *testing.T) {
	const total = 10
	nodes := make([]navsync.Node, total)
	edges := make([]navsync.Edge, total)
	for i := 0; i < total; i++ {
		id := fmt.Sprintf("Sym%d", i)
		nodes[i] = symNode(id)
		edges[i] = selfSymEdge("symbol:"+id, fmt.Sprintf("/proj/file%d.go", i))
	}
	graph := &navsync.Graph{Nodes: nodes, Edges: edges}
	// Only file3 and file7 changed (2 of 10).
	gitDiff := []string{"/proj/file3.go", "/proj/file7.go"}

	scope := ComputeScope(gitDiff, nil, nil, graph)

	if len(scope) != 2 {
		t.Fatalf("incremental: got %d entries, want exactly 2 (NOT %d); entries=%+v", len(scope), total, scope)
	}
	got := make(map[string]bool, len(scope))
	for _, e := range scope {
		got[e.SubtreeID] = true
	}
	if !got["Sym3"] || !got["Sym7"] {
		t.Errorf("incremental: expected Sym3+Sym7, got %v", got)
	}
}

// TestComputeScope_Deterministic validates idempotence (AC-NS5-004b): two
// calls with identical inputs produce byte-identical output (sorted, dedup'd).
func TestComputeScope_Deterministic(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA"), symNode("SymB"), symNode("SymC")},
		Edges: []navsync.Edge{
			selfSymEdge("symbol:SymA", "/proj/a.go"),
			selfSymEdge("symbol:SymB", "/proj/b.go"),
			selfSymEdge("symbol:SymC", "/proj/c.go"),
		},
	}
	gitDiff := []string{"/proj/c.go", "/proj/a.go", "/proj/b.go"} // intentionally unsorted
	m1 := []string{"/proj/b.go"}
	m2 := []WorkItemRef{{SourceKind: "detect", OwnerPath: "/proj/c.go", Action: "regen"}}

	run1 := ComputeScope(gitDiff, m1, m2, graph)
	run2 := ComputeScope(gitDiff, m1, m2, graph)

	data1, err := json.Marshal(run1)
	if err != nil {
		t.Fatalf("marshal run1: %v", err)
	}
	data2, err := json.Marshal(run2)
	if err != nil {
		t.Fatalf("marshal run2: %v", err)
	}
	if !bytes.Equal(data1, data2) {
		t.Errorf("determinism: two identical-input runs differ\nrun1=%s\nrun2=%s", data1, data2)
	}

	// Output is sorted by (doc_surface, subtree_id) — SymA before SymB before SymC.
	if len(run1) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(run1))
	}
	if run1[0].SubtreeID != "SymA" || run1[1].SubtreeID != "SymB" || run1[2].SubtreeID != "SymC" {
		t.Errorf("output not sorted by subtree_id: %s %s %s", run1[0].SubtreeID, run1[1].SubtreeID, run1[2].SubtreeID)
	}
}

// TestComputeScope_EmptyUnion validates that an empty union (no path in any
// input set) yields an empty (non-nil) diff-scope — the success case where the
// doc map is already consistent (REQ-NS5-009 row 009g).
func TestComputeScope_EmptyUnion(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA")},
		Edges: []navsync.Edge{selfSymEdge("symbol:SymA", "/proj/a.go")},
	}
	scope := ComputeScope(nil, nil, nil, graph)
	if scope == nil {
		t.Fatal("empty union: got nil, want non-nil empty slice")
	}
	if len(scope) != 0 {
		t.Errorf("empty union: got %d entries, want 0", len(scope))
	}
}

// TestComputeScope_DedupSameSubtree validates that two graph-bound paths
// binding to the SAME node produce ONE entry (dedup by doc_surface+subtree_id),
// with a merged stale_reason capturing both paths' contributions.
func TestComputeScope_DedupSameSubtree(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA")},
		Edges: []navsync.Edge{
			selfSymEdge("symbol:SymA", "/proj/a1.go"),
			selfSymEdge("symbol:SymA", "/proj/a2.go"),
		},
	}
	gitDiff := []string{"/proj/a1.go"}     // a1.go in git-diff
	m1 := []string{"/proj/a2.go"}          // a2.go in M1
	scope := ComputeScope(gitDiff, m1, nil, graph)

	if len(scope) != 1 {
		t.Fatalf("dedup: got %d entries, want 1 (both paths bind SymA); entries=%+v", len(scope), scope)
	}
	if scope[0].SubtreeID != "SymA" {
		t.Errorf("dedup: expected SymA, got %q", scope[0].SubtreeID)
	}
	// Merged reason: a1.go is git-diff, a2.go is m1-detect → "git-diff+m1-detect".
	if scope[0].StaleReason != "git-diff+m1-detect" {
		t.Errorf("dedup merged stale_reason = %q, want \"git-diff+m1-detect\"", scope[0].StaleReason)
	}
}

// TestComputeScope_NilGraph validates fail-open on a nil graph: no graph means
// no graph-bound paths, so the intersection is empty → empty diff-scope.
func TestComputeScope_NilGraph(t *testing.T) {
	scope := ComputeScope([]string{"/proj/a.go"}, nil, nil, nil)
	if scope == nil {
		t.Fatal("nil graph: got nil, want non-nil empty slice")
	}
	if len(scope) != 0 {
		t.Errorf("nil graph: got %d entries, want 0 (no graph-bound paths)", len(scope))
	}
}

// TestComputeScope_DecisionDocSurface validates that a decision node maps to
// capability-map.md (the third entity-type → doc-surface branch that
// docSurfaceFor handles, complementing the symbol + spec cases above).
func TestComputeScope_DecisionDocSurface(t *testing.T) {
	decN := navsync.Node{EntityType: navsync.EntityDecision, Identifier: "AUTH-STRATEGY", DisplayName: "Auth Strategy"}
	graph := &navsync.Graph{
		Nodes: []navsync.Node{decN},
		Edges: []navsync.Edge{selfSymEdge(nodeKey(decN), "/proj/tech.md")},
	}
	scope := ComputeScope([]string{"/proj/tech.md"}, nil, nil, graph)
	if len(scope) != 1 {
		t.Fatalf("decision: got %d entries, want 1", len(scope))
	}
	if scope[0].DocSurface != "capability-map.md" {
		t.Errorf("decision doc_surface = %q, want capability-map.md", scope[0].DocSurface)
	}
}

// TestComputeScope_MultiM2RefDeterminism validates that when multiple M2
// owner-paths bind the SAME subtree, the work_item_ref is selected
// deterministically (lexicographically smallest by source_kind, owner_path,
// action) — so two runs on the same inputs produce byte-identical output.
func TestComputeScope_MultiM2RefDeterminism(t *testing.T) {
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symNode("SymA")},
		Edges: []navsync.Edge{
			selfSymEdge("symbol:SymA", "/proj/a1.go"),
			selfSymEdge("symbol:SymA", "/proj/a2.go"),
		},
	}
	m2 := []WorkItemRef{
		{SourceKind: "detect", OwnerPath: "/proj/a2.go", Action: "zzz"},
		{SourceKind: "audit-orphan", OwnerPath: "/proj/a1.go", Action: "aaa"},
	}
	run1 := ComputeScope(nil, nil, m2, graph)
	run2 := ComputeScope(nil, nil, m2, graph)

	if len(run1) != 1 {
		t.Fatalf("multi-m2: got %d entries, want 1", len(run1))
	}
	if run1[0].WorkItemRef == nil {
		t.Fatal("multi-m2: expected non-nil work_item_ref")
	}
	// Lexicographically smallest: "audit-orphan" < "detect".
	want := "audit-orphan"
	if run1[0].WorkItemRef.SourceKind != want {
		t.Errorf("multi-m2 ref source_kind = %q, want %q (smallest)", run1[0].WorkItemRef.SourceKind, want)
	}
	// Determinism: two runs identical.
	d1, _ := json.Marshal(run1)
	d2, _ := json.Marshal(run2)
	if !bytes.Equal(d1, d2) {
		t.Errorf("multi-m2: two runs differ despite deterministic ref selection")
	}
}

// TestComputeScope_DocSurfaceMapping validates that different entity types map
// to their respective doc surfaces (symbol→capability-symbols.json,
// spec→audit-report.json, decision→capability-map.md).
func TestComputeScope_DocSurfaceMapping(t *testing.T) {
	symN := symNode("ParseHeader")
	specN := specNode("SPEC-X-001")
	graph := &navsync.Graph{
		Nodes: []navsync.Node{symN, specN},
		Edges: []navsync.Edge{
			selfSymEdge(nodeKey(symN), "/proj/sym.go"),
			selfSymEdge(nodeKey(specN), "/proj/spec.go"),
		},
	}
	scope := ComputeScope([]string{"/proj/sym.go", "/proj/spec.go"}, nil, nil, graph)
	if len(scope) != 2 {
		t.Fatalf("doc-surface mapping: got %d entries, want 2", len(scope))
	}
	surfaces := make(map[string]string, len(scope))
	for _, e := range scope {
		surfaces[e.SubtreeID] = e.DocSurface
	}
	if surfaces["ParseHeader"] != "capability-symbols.json" {
		t.Errorf("symbol doc_surface = %q, want capability-symbols.json", surfaces["ParseHeader"])
	}
	if surfaces["SPEC-X-001"] != "audit-report.json" {
		t.Errorf("spec doc_surface = %q, want audit-report.json", surfaces["SPEC-X-001"])
	}
}

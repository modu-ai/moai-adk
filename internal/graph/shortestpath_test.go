package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// shortestPathFixture persists a hand-written edges artifact (acceptance.md
// §A convention) whose Sources use the REAL extraction shape `file:function`
// and whose Targets are bare callee names — the exact persisted shapes the
// BFS must consume. Returns the fixture root.
func shortestPathFixture(t *testing.T, edges []Edge) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONL(filepath.Join(dir, "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}
	return root
}

// chainEdges builds the a1→a2→…→aN chain (N-1 code-call hops) plus a
// terminal aN→fmt.Println edge so aN carries a node id — a leaf callee has
// none (its name persists only as a Target, REQ-GR-003 clause 3).
func chainEdges(n int) []Edge {
	var edges []Edge
	for i := 1; i < n; i++ {
		edges = append(edges, Edge{
			Kind:       KindCodeCall,
			Source:     fmt.Sprintf("internal/w/chain.go:a%d", i),
			Target:     fmt.Sprintf("a%d", i+1),
			Line:       10 * i,
			Grade:      "full",
			Resolution: ResolutionExtracted,
			Confidence: ConfidenceFor(ResolutionExtracted),
		})
	}
	edges = append(edges, Edge{
		Kind:       KindCodeCall,
		Source:     fmt.Sprintf("internal/w/chain.go:a%d", n),
		Target:     "fmt.Println",
		Line:       10 * n,
		Grade:      "full",
		Resolution: ResolutionInferred,
		Confidence: ConfidenceFor(ResolutionInferred),
	})
	return edges
}

// AC-GR-002 — exactly 8 hops: found, hop count ≤ 8, endpoints addressed both
// by node id and by bare name (the chain function names are unique).
func TestShortestPath_EightHopChainFound(t *testing.T) {
	root := shortestPathFixture(t, chainEdges(9)) // a1→…→a9 = 8 hops

	// Node-id addressing (the stored file:function shape).
	res, err := ShortestPath(root, "internal/w/chain.go:a1", "internal/w/chain.go:a9")
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if !res.Found {
		t.Fatalf("8-hop chain must be found: %+v", res)
	}
	if res.HopCount != 8 || len(res.Hops) != 8 {
		t.Errorf("hop count = %d (hops=%d), want 8", res.HopCount, len(res.Hops))
	}
	if res.Cap != maxTraceDepth {
		t.Errorf("cap = %d, want shared maxTraceDepth %d", res.Cap, maxTraceDepth)
	}
	// Hop content: the traversed edges in order, confidence riding along.
	for i, hop := range res.Hops {
		wantFrom := fmt.Sprintf("internal/w/chain.go:a%d", i+1)
		wantTo := fmt.Sprintf("internal/w/chain.go:a%d", i+2)
		if hop.From != wantFrom || hop.To != wantTo {
			t.Errorf("hop %d = %s→%s, want %s→%s", i, hop.From, hop.To, wantFrom, wantTo)
		}
		if hop.Line == 0 || hop.Confidence == 0 {
			t.Errorf("hop %d lacks line/confidence: %+v", i, hop)
		}
	}
	if !containsStr(res.Provenance, root) {
		t.Errorf("provenance must name the tree: %q", res.Provenance)
	}

	// Bare-name addressing: a1 and a9 each resolve to exactly one node.
	resBare, err := ShortestPath(root, "a1", "a9")
	if err != nil {
		t.Fatalf("ShortestPath bare names: %v", err)
	}
	if !resBare.Found || resBare.HopCount != 8 {
		t.Errorf("bare-name query: found=%t hops=%d, want found 8", resBare.Found, resBare.HopCount)
	}
}

// AC-GR-002 — 9 hops exceeds the cap: structured not-found naming both
// endpoints and the cap; a truncated non-reaching path is never returned.
func TestShortestPath_NineHopChainNotFound(t *testing.T) {
	root := shortestPathFixture(t, chainEdges(10)) // a1→…→a10 = 9 hops

	res, err := ShortestPath(root, "a1", "a10")
	if err != nil {
		t.Fatalf("over-cap must be a structured result, not an error: %v", err)
	}
	if res.Found {
		t.Fatalf("9-hop chain must not be found within the 8-hop cap: %+v", res)
	}
	if len(res.Hops) != 0 {
		t.Errorf("not-found must carry no path, got %d hops", len(res.Hops))
	}
	if !containsStr(res.Reason, "a1") || !containsStr(res.Reason, "a10") {
		t.Errorf("reason must name both endpoints: %q", res.Reason)
	}
	if !containsStr(res.Reason, fmt.Sprint(maxTraceDepth)) {
		t.Errorf("reason must name the cap: %q", res.Reason)
	}
	if !containsStr(res.Provenance, root) {
		t.Errorf("not-found must still carry provenance: %q", res.Provenance)
	}
}

// AC-GR-003 — disconnected endpoints: structured not-found, never an error.
// Q carries a node id (it is a Source), so the miss is genuinely "no path",
// not an unresolvable endpoint.
func TestShortestPath_DisconnectedNotFound(t *testing.T) {
	root := shortestPathFixture(t, []Edge{
		{Kind: KindCodeCall, Source: "internal/w/left.go:X", Target: "Y", Line: 5},
		{Kind: KindCodeCall, Source: "internal/w/right.go:Q", Target: "Y2", Line: 7},
	})

	res, err := ShortestPath(root, "internal/w/left.go:X", "internal/w/right.go:Q")
	if err != nil {
		t.Fatalf("disconnected must be a structured result, not an error: %v", err)
	}
	if res.Found || len(res.Hops) != 0 {
		t.Errorf("disconnected endpoints must yield no path: %+v", res)
	}
	if !containsStr(res.Reason, "X") || !containsStr(res.Reason, "Q") {
		t.Errorf("reason must name both endpoints: %q", res.Reason)
	}
}

// AC-GR-003 ambiguous-endpoint clause — a bare name resolving to two nodes
// (same function name in two files) yields the structured candidates list
// (name → matching node ids) and no path, never a guessed join.
func TestShortestPath_AmbiguousEndpointCandidates(t *testing.T) {
	root := shortestPathFixture(t, []Edge{
		{Kind: KindCodeCall, Source: "internal/w/one.go:Dup", Target: "leaf1", Line: 3},
		{Kind: KindCodeCall, Source: "internal/w/two.go:Dup", Target: "leaf2", Line: 4},
	})

	res, err := ShortestPath(root, "Dup", "leaf1")
	if err != nil {
		t.Fatalf("ambiguous endpoint must be a structured result, not an error: %v", err)
	}
	if res.Found || len(res.Hops) != 0 {
		t.Errorf("ambiguous endpoint must yield no path: %+v", res)
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("want one candidate entry, got %+v", res.Candidates)
	}
	cand := res.Candidates[0]
	if cand.Name != "Dup" {
		t.Errorf("candidate name = %q, want Dup", cand.Name)
	}
	wantNodes := []string{"internal/w/one.go:Dup", "internal/w/two.go:Dup"}
	if len(cand.Nodes) != 2 || cand.Nodes[0] != wantNodes[0] || cand.Nodes[1] != wantNodes[1] {
		t.Errorf("candidate nodes = %v, want sorted %v", cand.Nodes, wantNodes)
	}
}

// REQ-GR-003 clause 3 — a bare name matching NO node id (appearing only as
// a Target callee) yields the same structured not-found result: the honest
// degraded answer, never a guessed join.
func TestShortestPath_TargetOnlyNameNotFound(t *testing.T) {
	root := shortestPathFixture(t, chainEdges(3)) // fmt.Println appears only as a Target

	res, err := ShortestPath(root, "a1", "fmt.Println")
	if err != nil {
		t.Fatalf("target-only name must be a structured result, not an error: %v", err)
	}
	if res.Found || len(res.Hops) != 0 || len(res.Candidates) != 0 {
		t.Errorf("target-only endpoint must yield plain not-found: %+v", res)
	}
}

// REQ-GR-001 / AC-GR-003 ambiguous-INTERMEDIATE rule (plan §F M1): the bare
// callee M on the candidate path matches two nodes, so it is treated as no
// continuation — deterministic, never joined through the bare-name caller
// index into an unrelated node. S→E is unreachable without M, so the answer
// is not-found even though one.go:M does call E.
func TestShortestPath_AmbiguousIntermediateNoContinuation(t *testing.T) {
	root := shortestPathFixture(t, []Edge{
		// S → M (bare callee, ambiguous)
		{Kind: KindCodeCall, Source: "internal/w/start.go:S", Target: "M", Line: 2},
		// two node ids carry the function name M
		{Kind: KindCodeCall, Source: "internal/w/one.go:M", Target: "E", Line: 3},
		{Kind: KindCodeCall, Source: "internal/w/two.go:M", Target: "Z", Line: 4},
		// E carries a node id so the destination is resolvable
		{Kind: KindCodeCall, Source: "internal/w/end.go:E", Target: "leaf", Line: 5},
	})

	res, err := ShortestPath(root, "S", "E")
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if res.Found || len(res.Hops) != 0 {
		t.Errorf("ambiguous intermediate must not be joined through: %+v", res)
	}
	if !containsStr(res.Reason, "S") || !containsStr(res.Reason, "E") {
		t.Errorf("reason must name both endpoints: %q", res.Reason)
	}
}

// AC-GR-004 — two runs over the same tree produce byte-identical responses
// (total-order neighbor iteration: node id, then line). The exec-cmp subtest
// pins the same property at the shell level where cmp exists.
func TestShortestPath_Deterministic(t *testing.T) {
	// Two same-cost routes to D: via B (line 10) and via C (line 20) — the
	// tie must break identically on every run.
	root := shortestPathFixture(t, []Edge{
		{Kind: KindCodeCall, Source: "internal/w/x.go:A", Target: "B", Line: 10},
		{Kind: KindCodeCall, Source: "internal/w/x.go:A", Target: "C", Line: 20},
		{Kind: KindCodeCall, Source: "internal/w/b.go:B", Target: "D", Line: 30},
		{Kind: KindCodeCall, Source: "internal/w/c.go:C", Target: "D", Line: 40},
		{Kind: KindCodeCall, Source: "internal/w/d.go:D", Target: "leaf", Line: 50},
	})

	first, err := ShortestPath(root, "A", "D")
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if !first.Found || first.HopCount != 2 {
		t.Fatalf("want a 2-hop path, got %+v", first)
	}
	for i := 0; i < 10; i++ {
		again, err := ShortestPath(root, "A", "D")
		if err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		b1, _ := json.Marshal(first)
		b2, _ := json.Marshal(again)
		if !bytes.Equal(b1, b2) {
			t.Fatalf("run %d diverged:\n%s\n%s", i, b1, b2)
		}
	}

	// Shell-level cmp on the two serialized responses (AC-GR-004's literal
	// verification) — skipped where cmp(1) does not exist.
	if runtime.GOOS == "windows" {
		t.Skip("cmp(1) unavailable on windows")
	}
	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(func() PathResult {
		r, err := ShortestPath(root, "A", "D")
		if err != nil {
			t.Fatal(err)
		}
		return r
	}())
	f1 := filepath.Join(t.TempDir(), "run1.json")
	f2 := filepath.Join(t.TempDir(), "run2.json")
	if err := os.WriteFile(f1, b1, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, b2, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("cmp", f1, f2).Run(); err != nil {
		t.Fatalf("cmp of two consecutive runs must exit 0: %v", err)
	}
}

// The tie-break itself: A reaches D via B (B < C in node-id total order) —
// pinned so the determinism test's byte-equality is not vacuously satisfied
// by an arbitrary but unstable pick.
func TestShortestPath_TieBreakTotalOrder(t *testing.T) {
	root := shortestPathFixture(t, []Edge{
		{Kind: KindCodeCall, Source: "internal/w/x.go:A", Target: "B", Line: 10},
		{Kind: KindCodeCall, Source: "internal/w/x.go:A", Target: "C", Line: 20},
		{Kind: KindCodeCall, Source: "internal/w/b.go:B", Target: "D", Line: 30},
		{Kind: KindCodeCall, Source: "internal/w/c.go:C", Target: "D", Line: 40},
		{Kind: KindCodeCall, Source: "internal/w/d.go:D", Target: "leaf", Line: 50},
	})

	res, err := ShortestPath(root, "A", "D")
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if len(res.Hops) != 2 || res.Hops[0].To != "internal/w/b.go:B" {
		t.Errorf("tie must break through b.go:B (node-id total order): %+v", res.Hops)
	}
}

// Absent artifact (plan §F M1 tests): the actionable "run 'moai graph
// build'" error shape — the same loadCodeEdges contract as the other tools.
func TestShortestPath_AbsentArtifactActionableError(t *testing.T) {
	_, err := ShortestPath(t.TempDir(), "a", "b")
	if err == nil {
		t.Fatal("absent artifact must error")
	}
	if !containsStr(err.Error(), "moai graph build") {
		t.Errorf("error must be actionable: %v", err)
	}
}

func containsStr(s, sub string) bool { return strings.Contains(s, sub) }

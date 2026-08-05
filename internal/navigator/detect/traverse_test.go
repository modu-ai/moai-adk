package detect

import (
	"path/filepath"
	"strings"
	"testing"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// TestTraverse_ReverseTraversal is the table-driven AC-NS2-002 test covering
// all three M0 edge types (dec-edge / spec-edge / sym-edge) per
// acceptance.md §D.AC-NS2-002 rows 002a/002b/002c.
func TestTraverse_ReverseTraversal(t *testing.T) {
	t.Parallel()

	const projRoot = "/abs/project"
	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntityDecision, Identifier: "AUTH-STRATEGY", DisplayName: "OAuth2 client-credentials"},
			{EntityType: navsync.EntitySpec, Identifier: "SPEC-AUTH-001", DisplayName: "Auth SPEC"},
			{EntityType: navsync.EntitySymbol, Identifier: "auth.ParseBearer", DisplayName: "ParseBearer"},
		},
		Edges: []navsync.Edge{
			// 002a dec-edge — design doc carries the @NAV:DEC token.
			{
				EdgeType:   navsync.EdgeDec,
				SourceNode: "decision:AUTH-STRATEGY",
				TargetNode: "spec:SPEC-AUTH-001",
				SourcePath: filepath.Join(projRoot, ".moai/project/tech.md"),
				LineNumber: 42,
			},
			// 002b spec-edge — @MX:SPEC bridge back-pointer; source is the code file.
			{
				EdgeType:   navsync.EdgeSpec,
				SourceNode: "symbol:auth.ParseBearer",
				TargetNode: "spec:SPEC-AUTH-001",
				SourcePath: filepath.Join(projRoot, "internal/auth/login.go"),
				LineNumber: 17,
			},
			// 002c sym-edge — @NAV:SYM code-symbol binding.
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:auth.ParseBearer",
				TargetNode: "decision:AUTH-STRATEGY",
				SourcePath: filepath.Join(projRoot, "internal/auth/login.go"),
				LineNumber: 30,
			},
		},
	}

	tests := []struct {
		name         string
		changedPath  string
		wantEdgeType navsync.EdgeType
		wantNodeKey  string
		wantLine     int
	}{
		{
			name:         "002a dec-edge: design-doc change surfaces decision node",
			changedPath:  filepath.Join(projRoot, ".moai/project/tech.md"),
			wantEdgeType: navsync.EdgeDec,
			wantNodeKey:  "decision:AUTH-STRATEGY",
			wantLine:     42,
		},
		{
			name:         "002b spec-edge: code change surfaces SPEC back-pointer (highest-value case)",
			changedPath:  filepath.Join(projRoot, "internal/auth/login.go"),
			wantEdgeType: navsync.EdgeSpec,
			wantNodeKey:  "spec:SPEC-AUTH-001",
			wantLine:     17,
		},
		{
			name:         "002c sym-edge: code change surfaces symbol binding",
			changedPath:  filepath.Join(projRoot, "internal/auth/login.go"),
			wantEdgeType: navsync.EdgeSym,
			wantNodeKey:  "symbol:auth.ParseBearer",
			wantLine:     30,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := Traverse(graph, tc.changedPath)
			if err != nil {
				t.Fatalf("Traverse returned error: %v", err)
			}
			// Assert the expected originating edge is present.
			var foundEdge bool
			for _, e := range res.Edges {
				if e.EdgeType == tc.wantEdgeType && e.LineNumber == tc.wantLine {
					foundEdge = true
					break
				}
			}
			if !foundEdge {
				t.Errorf("expected edge (type=%s, line=%d) in affected edges; got %+v",
					tc.wantEdgeType, tc.wantLine, res.Edges)
			}
			// Assert the expected affected node is present.
			var foundNode bool
			for _, n := range res.Nodes {
				if n.Key == tc.wantNodeKey {
					foundNode = true
					if n.DisplayName == "" {
						t.Errorf("affected node %q has empty DisplayName", tc.wantNodeKey)
					}
					break
				}
			}
			if !foundNode {
				t.Errorf("expected node %q in affected nodes; got %+v", tc.wantNodeKey, res.Nodes)
			}
		})
	}
}

// TestTraverse_DirectoryPrefixFallback is the AC-NS2-010 test: a changed
// DIRECTORY prefix (path ending in a separator) maps to ≥1 edge whose
// source_path falls under that prefix. The fallback is INSPIRED by
// navigator-audit.sh heuristic_match() (cited in Traverse's doc comment);
// the primary engine remains absolute-path string equality.
func TestTraverse_DirectoryPrefixFallback(t *testing.T) {
	t.Parallel()

	const projRoot = "/abs/project"
	fooDir := filepath.Join(projRoot, "internal/foo")
	graph := &navsync.Graph{
		Nodes: []navsync.Node{
			{EntityType: navsync.EntitySymbol, Identifier: "foo.Bar", DisplayName: "Bar"},
			{EntityType: navsync.EntitySymbol, Identifier: "foo.Baz", DisplayName: "Baz"},
			{EntityType: navsync.EntitySymbol, Identifier: "other.Qux", DisplayName: "Qux"},
		},
		Edges: []navsync.Edge{
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:foo.Bar",
				TargetNode: "decision:X",
				SourcePath: filepath.Join(fooDir, "bar.go"),
				LineNumber: 10,
			},
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:foo.Baz",
				TargetNode: "decision:X",
				SourcePath: filepath.Join(fooDir, "baz.go"),
				LineNumber: 20,
			},
			// Unrelated edge OUTSIDE the prefix — MUST NOT match.
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:other.Qux",
				TargetNode: "decision:X",
				SourcePath: filepath.Join(projRoot, "internal/other/qux.go"),
				LineNumber: 5,
			},
			// Negative-prefix trap: a sibling directory whose name shares the
			// "foo" prefix string (foo-extra) MUST NOT leak through a naive
			// HasPrefix check on the un-slash-ed form.
			{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:foox.Trap",
				TargetNode: "decision:X",
				SourcePath: filepath.Join(projRoot, "internal/foo-extra/trap.go"),
				LineNumber: 99,
			},
		},
	}

	// changedPath is a directory prefix (trailing separator), per AC-NS2-010.
	changedPath := fooDir + string(filepath.Separator)

	res, err := Traverse(graph, changedPath)
	if err != nil {
		t.Fatalf("Traverse returned error: %v", err)
	}
	if len(res.Edges) < 1 {
		t.Fatalf("expected ≥1 affected edge under prefix %q; got 0", changedPath)
	}
	prefixWithSep := fooDir + string(filepath.Separator)
	for _, e := range res.Edges {
		if !strings.HasPrefix(e.SourcePath, prefixWithSep) {
			t.Errorf("edge source_path %q is not under prefix %q", e.SourcePath, changedPath)
		}
	}
	// The out-of-prefix sibling edge MUST NOT appear.
	for _, e := range res.Edges {
		if e.SourcePath == filepath.Join(projRoot, "internal/other/qux.go") {
			t.Errorf("prefix fallback leaked an out-of-prefix edge: %+v", e)
		}
	}
	// The negative-prefix trap (foo-extra) MUST NOT appear.
	for _, e := range res.Edges {
		if strings.Contains(e.SourcePath, "foo-extra") {
			t.Errorf("prefix fallback leaked the negative-prefix trap: %+v", e)
		}
	}
}

// TestTraverse_EdgeCases covers the §F edge-case bar: empty graph, non-match,
// deterministic ordering, and the input-validation error paths.
func TestTraverse_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty graph returns empty result no error", func(t *testing.T) {
		t.Parallel()
		res, err := Traverse(&navsync.Graph{}, "/abs/project/internal/foo/bar.go")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if len(res.Nodes) != 0 || len(res.Edges) != 0 {
			t.Errorf("expected empty result; got nodes=%d edges=%d", len(res.Nodes), len(res.Edges))
		}
	})

	t.Run("non-matching path returns empty result", func(t *testing.T) {
		t.Parallel()
		graph := &navsync.Graph{
			Nodes: []navsync.Node{{EntityType: navsync.EntitySymbol, Identifier: "x.Y", DisplayName: "Y"}},
			Edges: []navsync.Edge{{
				EdgeType:   navsync.EdgeSym,
				SourceNode: "symbol:x.Y",
				TargetNode: "decision:D",
				SourcePath: "/abs/project/internal/x/y.go",
				LineNumber: 1,
			}},
		}
		res, err := Traverse(graph, "/abs/project/internal/unrelated/z.go")
		if err != nil {
			t.Fatalf("expected nil error; got %v", err)
		}
		if len(res.Nodes) != 0 || len(res.Edges) != 0 {
			t.Errorf("expected empty result for non-matching path; got nodes=%d edges=%d",
				len(res.Nodes), len(res.Edges))
		}
	})

	t.Run("deterministic ordering across repeated calls", func(t *testing.T) {
		t.Parallel()
		graph := &navsync.Graph{
			Nodes: []navsync.Node{
				{EntityType: navsync.EntitySymbol, Identifier: "b.B", DisplayName: "B"},
				{EntityType: navsync.EntitySymbol, Identifier: "a.A", DisplayName: "A"},
				// decision:D is the shared target node; included so the
				// fixture graph is well-formed.
				{EntityType: navsync.EntityDecision, Identifier: "D", DisplayName: "D"},
			},
			// Edges deliberately inserted in non-sorted order.
			Edges: []navsync.Edge{
				{
					EdgeType: navsync.EdgeSym, SourceNode: "symbol:b.B", TargetNode: "decision:D",
					SourcePath: "/p/x.go", LineNumber: 5,
				},
				{
					EdgeType: navsync.EdgeSym, SourceNode: "symbol:a.A", TargetNode: "decision:D",
					SourcePath: "/p/x.go", LineNumber: 1,
				},
			},
		}
		first, err := Traverse(graph, "/p/x.go")
		if err != nil {
			t.Fatalf("Traverse returned error: %v", err)
		}
		second, err := Traverse(graph, "/p/x.go")
		if err != nil {
			t.Fatalf("Traverse returned error on second call: %v", err)
		}
		if len(first.Edges) != 2 {
			t.Fatalf("expected 2 edges; got %d", len(first.Edges))
		}
		// Edges sorted by (EdgeType, SourceNode, TargetNode, SourcePath, LineNumber).
		if first.Edges[0].LineNumber != 1 {
			t.Errorf("expected first edge LineNumber=1 (a.A); got %d", first.Edges[0].LineNumber)
		}
		if first.Edges[1].LineNumber != 5 {
			t.Errorf("expected second edge LineNumber=5 (b.B); got %d", first.Edges[1].LineNumber)
		}
		// Nodes sorted by (EntityType, Identifier) — decision sorts before
		// symbol, so decision:D is first and symbol:a.A precedes symbol:b.B.
		idxA, idxB, idxD := -1, -1, -1
		for i, n := range first.Nodes {
			switch n.Key {
			case "symbol:a.A":
				idxA = i
			case "symbol:b.B":
				idxB = i
			case "decision:D":
				idxD = i
			}
		}
		if idxA < 0 || idxB < 0 || idxD < 0 {
			t.Fatalf("expected all three nodes in result; got %+v", first.Nodes)
		}
		if idxD > idxA {
			t.Errorf("expected decision:D before symbol:a.A; got idxD=%d idxA=%d", idxD, idxA)
		}
		if idxA > idxB {
			t.Errorf("expected symbol:a.A before symbol:b.B; got idxA=%d idxB=%d", idxA, idxB)
		}
		// Byte-stable across calls.
		if len(first.Nodes) != len(second.Nodes) || len(first.Edges) != len(second.Edges) {
			t.Errorf("non-deterministic result sizes across calls")
		}
	})

	t.Run("nil graph returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Traverse(nil, "/abs/project/x.go")
		if err == nil {
			t.Error("expected error for nil graph; got nil")
		}
	})

	t.Run("empty changed path returns error", func(t *testing.T) {
		t.Parallel()
		_, err := Traverse(&navsync.Graph{}, "")
		if err == nil {
			t.Error("expected error for empty changed path; got nil")
		}
	})

	t.Run("both source and target nodes collected conservatively", func(t *testing.T) {
		t.Parallel()
		graph := &navsync.Graph{
			Nodes: []navsync.Node{
				{EntityType: navsync.EntitySymbol, Identifier: "s.Src", DisplayName: "Src"},
				{EntityType: navsync.EntitySpec, Identifier: "SPEC-TGT-001", DisplayName: "Tgt"},
			},
			Edges: []navsync.Edge{{
				EdgeType:   navsync.EdgeSpec,
				SourceNode: "symbol:s.Src",
				TargetNode: "spec:SPEC-TGT-001",
				SourcePath: "/abs/project/internal/s.go",
				LineNumber: 1,
			}},
		}
		res, err := Traverse(graph, "/abs/project/internal/s.go")
		if err != nil {
			t.Fatalf("Traverse returned error: %v", err)
		}
		keys := map[string]bool{}
		for _, n := range res.Nodes {
			keys[n.Key] = true
		}
		// Both endpoints collected (advisory over-inclusion per plan §C.2).
		if !keys["symbol:s.Src"] {
			t.Errorf("expected source node symbol:s.Src in affected set; got %v", keys)
		}
		if !keys["spec:SPEC-TGT-001"] {
			t.Errorf("expected target node spec:SPEC-TGT-001 in affected set; got %v", keys)
		}
	})
}

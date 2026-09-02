//go:build !cgo

package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-GEC-009 (SPEC-GRAPH-EDGE-CONFIDENCE-001) — under a CGO-disabled build
// the confidence layer is INERT (REQ-GEC-010): the extractor reports
// unsupported, zero code-derived edges are produced, no confidence-bearing
// line reaches the artifact, and the doc layers plus query paths behave
// exactly as before — no new error surface. Mirrors the astx nocgo_test.go
// pattern; runs only where tree-sitter is unavailable.
func TestNoCGOConfidenceInert(t *testing.T) {
	root := tierFixture(t)
	edges, matrix, err := CodeEdges(root)
	if err != nil {
		t.Fatalf("CodeEdges under !cgo must not error: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("CGO-disabled build must produce zero code-derived edges, got %d: %v", len(edges), edges)
	}
	if d := ValidateGradeMatrix(matrix); len(d) != 0 {
		t.Errorf("grade matrix must stay intact under !cgo: %v", d)
	}

	all, _, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers under !cgo: %v", err)
	}
	for _, e := range all {
		if e.Kind == KindCodeCall || e.Kind == KindCodeImport {
			t.Errorf("code-derived edge under !cgo: %+v", e)
		}
		if e.Resolution != "" || e.Confidence != 0 {
			t.Errorf("confidence state on a !cgo build edge: %+v", e)
		}
	}

	dir := filepath.Join(root, ".moai", "project", "graph")
	path := filepath.Join(dir, "edges.jsonl")
	if err := WriteJSONL(path, all); err != nil {
		t.Fatalf("WriteJSONL under !cgo: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.Contains(line, `"resolution"`) || strings.Contains(line, `"confidence"`) {
			t.Errorf("line %d carries a confidence key under !cgo: %s", i+1, line)
		}
	}

	// Query paths serve the artifact without a new error surface.
	matches, _, err := FindCode(root, "B")
	if err != nil {
		t.Fatalf("FindCode under !cgo: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("FindCode must observe zero code-call matches under !cgo, got %v", matches)
	}
	if _, _, err := TraceCalls(root, "B", 1); err != nil {
		t.Fatalf("TraceCalls under !cgo: %v", err)
	}
}

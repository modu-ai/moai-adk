//go:build !cgo

package graph

import (
	"os"
	"path/filepath"
	"testing"
)

// B5 nocgo semantics (plan §F M1): under a CGO-disabled build the code layer
// is absent — zero code-call edges reach the artifact — and graph_shortest_path
// serves that artifact with the structured not-found answer, never a new
// error surface. Mirrors nocgo_test.go's pattern.
func TestNoCGOShortestPathAbsentSemantics(t *testing.T) {
	root := tierFixture(t)
	all, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers under !cgo: %v", err)
	}
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSONL(filepath.Join(dir, "edges.jsonl"), all); err != nil {
		t.Fatal(err)
	}

	res, err := ShortestPath(root, "A", "B")
	if err != nil {
		t.Fatalf("ShortestPath under !cgo must not error: %v", err)
	}
	if res.Found || len(res.Hops) != 0 {
		t.Errorf("code layer absent under !cgo must yield the honest not-found: %+v", res)
	}
}

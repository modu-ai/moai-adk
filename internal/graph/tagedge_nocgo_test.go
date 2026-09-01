//go:build !cgo

package graph

import (
	"context"
	"testing"
)

// AC-MTE-014 (SPEC-MX-TAG-EDGES-001) — under a CGO-disabled build the tag
// layer still emits mx-* edges from the line-based scanner layer, with ALL
// targets resolving to the self-edge form (no range data); zero code-call
// edges exist; and the edge-backed fan-in source degrades with an error so
// the validator falls back to the textual source, labeled (REQ-MTE-015).
// Mirrors the nocgo_test.go pattern; runs only where tree-sitter is
// unavailable.
func TestNoCGOTagEdgesSelfEdge(t *testing.T) {
	root := tagFixture(t)
	edges, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers under !cgo: %v", err)
	}

	mxCount := 0
	tagKinds := map[string]bool{}
	for _, k := range mxTagEdgeKinds {
		tagKinds[k] = true
	}
	for _, e := range edges {
		if e.Kind == KindCodeCall || e.Kind == KindCodeImport {
			t.Errorf("code-derived edge under !cgo: %+v", e)
		}
		if tagKinds[e.Kind] {
			mxCount++
			if e.Target != e.Source {
				t.Errorf("mx-* edge under !cgo must self-edge (no range data): %+v", e)
			}
		}
	}
	if mxCount == 0 {
		t.Fatal("no mx-* edges under !cgo — the scanner layer must still emit the tag layer")
	}

	// The edge-backed source degrades: zero code-call evidence means it
	// cannot answer, so the validator's textual fallback takes over with
	// its label.
	src := NewEdgeFanInSource(root)
	ev, inf, label, err := src.EvidenceBacked(context.Background(), "S", root+"/internal/core/core.go")
	if err == nil {
		t.Fatalf("EvidenceBacked under !cgo must degrade with an error, got ev=%d inf=%d label=%q", ev, inf, label)
	}
	if ev != 0 || inf != 0 || label != "" {
		t.Errorf("degraded answer must be zero-valued, got ev=%d inf=%d label=%q", ev, inf, label)
	}
}

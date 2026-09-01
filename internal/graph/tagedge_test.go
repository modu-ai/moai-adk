package graph

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// M1 acceptance tests (SPEC-MX-TAG-EDGES-001): tag-kind edges in the doc
// layer. AC-MTE-002 (kind domain), AC-MTE-003 (endpoints + self-edge
// fallback), AC-MTE-006 (single scan per build).

// mxTagEdgeKinds is the closed mx-* kind domain (REQ-MTE-001): one kind per
// standalone tag kind, named by lowercasing the tag kind and prefixing mx-.
var mxTagEdgeKinds = []string{
	KindMXNote,
	KindMXWarn,
	KindMXAnchor,
	KindMXTodo,
	KindMXLegacy,
	KindMXDebt,
}

// mxEdgesOf returns the edges whose kind is in the mx-* STANDALONE-tag
// domain — mx-spec is a sub-line edge, not a tag-kind edge, and is excluded.
func mxEdgesOf(edges []Edge) []Edge {
	var out []Edge
	for _, e := range edges {
		if strings.HasPrefix(e.Kind, "mx-") && e.Kind != KindMXSpec {
			out = append(out, e)
		}
	}
	return out
}

// AC-MTE-002 — every standalone tag occurrence appears as exactly one edge,
// kinds drawn only from the six-kind mx-* domain, all six present.
// CGO-independent: the doc layer is scanner-derived.
func TestTagEdgeKindDomain(t *testing.T) {
	root := tagFixture(t)
	edges, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	mx := mxEdgesOf(edges)
	if len(mx) != tagFixtureTagCount {
		t.Fatalf("got %d mx-* edges, want %d (one per tag occurrence): %+v", len(mx), tagFixtureTagCount, mx)
	}

	kindSet := map[string]int{}
	seen := map[string]int{}
	for _, e := range mx {
		kindSet[e.Kind]++
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%d", e.Kind, e.Source, e.Target, e.Line)
		seen[key]++
		if seen[key] > 1 {
			t.Errorf("duplicate tag edge %s %s→%s:%d", e.Kind, e.Source, e.Target, e.Line)
		}
	}
	if len(kindSet) != len(mxTagEdgeKinds) {
		t.Fatalf("got %d distinct mx-* kinds (%v), want all %d", len(kindSet), kindSet, len(mxTagEdgeKinds))
	}
	for _, k := range mxTagEdgeKinds {
		if kindSet[k] == 0 {
			t.Errorf("kind %s absent from the artifact", k)
		}
	}

	// Artifact spot-check shape: every mx-* line is a bare doc edge.
	for _, e := range mx {
		if e.Resolution != "" || e.Confidence != 0 || e.DisagreesWith != "" || e.Grade != "" {
			t.Errorf("mx-* edge %s %s→%s carries non-occurrence state: %+v", e.Kind, e.Source, e.Target, e)
		}
	}
}

// AC-MTE-003 — a tag inside a function body joins to the enclosing symbol
// name (innermost range wins); a file-scope tag self-edges (file → file).
// Requires the extractor's ranges, so CGO must be available.
func TestTagEdgeEndpoints(t *testing.T) {
	requireCodeExtraction(t)
	root := tagFixture(t)
	edges, _, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}

	byKey := map[string]Edge{}
	for _, e := range mxEdgesOf(edges) {
		byKey[fmt.Sprintf("%s\x00%s\x00%d", e.Kind, e.Source, e.Line)] = e
	}

	// (a) body-anchored NOTE in internal/app/anchored.go inside Anchored().
	anchored, ok := byKey[fmt.Sprintf("%s\x00internal/app/anchored.go\x00%d", KindMXNote, anchoredNoteLine)]
	if !ok {
		t.Fatalf("body-anchored NOTE edge missing: %v", byKey)
	}
	if anchored.Target != "Anchored" {
		t.Errorf("body-anchored NOTE target = %q, want the enclosing symbol %q", anchored.Target, "Anchored")
	}

	// (a2) body-anchored ANCHOR in Contract() resolves the same way.
	contract, ok := byKey[fmt.Sprintf("%s\x00internal/app/anchored.go\x00%d", KindMXAnchor, anchoredAnchorLine)]
	if !ok {
		t.Fatalf("body-anchored ANCHOR edge missing: %v", byKey)
	}
	if contract.Target != "Contract" {
		t.Errorf("body-anchored ANCHOR target = %q, want %q", contract.Target, "Contract")
	}

	// (b) nested-closure TODO: innermost-containing range wins.
	nested, ok := byKey[fmt.Sprintf("%s\x00internal/app/anchored.go\x00%d", KindMXTodo, nestedTodoLine)]
	if !ok {
		t.Fatalf("nested-closure TODO edge missing: %v", byKey)
	}
	if nested.Target == "internal/app/anchored.go" {
		t.Errorf("nested-closure TODO self-edged — innermost range join did not fire: %+v", nested)
	}

	// (c) file-scope NOTE self-edges (never dropped, never misattributed).
	fs, ok := byKey[fmt.Sprintf("%s\x00internal/app/filescope.go\x00%d", KindMXNote, fileScopeNoteLine)]
	if !ok {
		t.Fatalf("file-scope NOTE edge missing: %v", byKey)
	}
	if fs.Target != fs.Source || fs.Target != "internal/app/filescope.go" {
		t.Errorf("file-scope NOTE = %+v, want the self-edge form (target == source == file)", fs)
	}
}

// AC-MTE-006 — exactly ONE mx.Scanner.ScanDir pass executes per build, and
// its output feeds BOTH mx-spec and mx-* edges (seam-counted).
func TestSingleScanPerBuild(t *testing.T) {
	root := tagFixture(t)

	orig := scanDirFn
	scans := 0
	scanDirFn = func(projectRoot string) ([]mx.Tag, error) {
		scans++
		return orig(projectRoot)
	}
	t.Cleanup(func() { scanDirFn = orig })

	edges, err := Build(root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if scans != 1 {
		t.Fatalf("ScanDir ran %d times per build, want exactly 1", scans)
	}

	// Both consumers draw from the one pass: mx-spec (LEGACY's SPEC sub-line)
	// and mx-legacy both present in the same artifact.
	hasSpec, hasLegacy := false, false
	for _, e := range edges {
		if e.Kind == KindMXSpec && e.Target == "SPEC-MTE-FIXTURE-001" {
			hasSpec = true
		}
		if e.Kind == KindMXLegacy {
			hasLegacy = true
		}
	}
	if !hasSpec || !hasLegacy {
		t.Errorf("single-scan feed broken: mx-spec=%v mx-legacy=%v", hasSpec, hasLegacy)
	}

	// The code-layer build is also a single pass (same doc layer inside).
	scans = 0
	if _, _, _, err := BuildWithCodeLayers(root); err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}
	if scans != 1 {
		t.Fatalf("ScanDir ran %d times per BuildWithCodeLayers, want exactly 1", scans)
	}
}

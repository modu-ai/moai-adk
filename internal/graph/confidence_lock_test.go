package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// M2 regression locks (SPEC-GRAPH-EDGE-CONFIDENCE-001): the artifact
// contract the confidence derivation must never break.

// AC-GEC-001 — builds over fresh copies of the same tree write
// byte-identical edges.jsonl (REQ-GEC-005 determinism). Five fresh copies:
// emission order is normalized by the canonical sort, so the hazard the
// extra copies exist for is a confidence VALUE that flips with map
// iteration — repeated fresh builds make such a flip observable. The
// fixture also carries a multi-dependency spec exercising the sorted
// spec-depends emission path.
func TestEdgesJSONLDeterministic(t *testing.T) {
	const builds = 5
	paths := make([]string, 0, builds)
	for i := 0; i < builds; i++ {
		// tierFixture carries a callee (Dup) declared in TWO directories,
		// only one of them caller-imported — the shape whose confidence
		// value would flip if any map-iteration order reached a label.
		root := tierFixture(t)
		path := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
		edges, _, err := BuildWithCodeLayers(root)
		if err != nil {
			t.Fatalf("build %d: %v", i, err)
		}
		if err := WriteJSONL(path, edges); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		paths = append(paths, path)
	}
	base, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(base) == 0 {
		t.Fatal("empty artifact — nothing was compared")
	}
	for i, path := range paths[1:] {
		other, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(base, other) {
			t.Fatalf("edges.jsonl not byte-identical: build 0 vs build %d\nA=%s\nB=%s", i+1, base, other)
		}
	}
}

// AC-GEC-006 — the committed golden (testdata/edges-doc-golden.jsonl,
// originally generated on base 5593e8cff, extended with mx-* lines on base
// 63435427c per plan.md §G) pins every doc-kind and code-import line
// byte-for-byte, with no resolution/confidence keys. The golden is NEVER
// hand-edited; regenerate only when goldenFixture itself changes, naming
// the new base SHA.
func TestDocEdgesByteIdentical(t *testing.T) {
	requireCodeExtraction(t)
	golden, err := os.ReadFile(filepath.Join("testdata", "edges-doc-golden.jsonl"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	root := goldenFixture(t)
	edges, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}

	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	for _, e := range edges {
		if e.Kind == KindCodeCall {
			continue
		}
		if e.Resolution != "" || e.Confidence != 0 {
			t.Errorf("doc-kind edge %s/%s carries confidence state: %q %v", e.Kind, e.Target, e.Resolution, e.Confidence)
		}
		if err := encoder.Encode(e); err != nil {
			t.Fatalf("encode edge: %v", err)
		}
	}
	if b.Len() != len(golden) || !bytes.Equal(b.Bytes(), golden) {
		t.Fatalf("doc+code-import lines diverged from the committed golden (%d vs %d bytes):\nGOLDEN=%s\nGOT=%s",
			len(golden), b.Len(), golden, b.String())
	}
}

// AC-GEC-008 — a pre-upgrade edges.jsonl (no resolution/confidence keys)
// loads and serves without error: consumers treat absent confidence as
// unknown (0/omitted), never a failure.
func TestLegacyArtifactLoad(t *testing.T) {
	requireCodeExtraction(t)
	root := tierFixture(t)
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"kind":"code-call","source":"internal/wire/wire.go:A","target":"B","line":6,"grade":"name-based"}
{"kind":"code-import","source":"internal/wire/wire.go","target":"internal/helper","line":3,"grade":"name-based"}
`
	path := filepath.Join(dir, "edges.jsonl")
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadJSONL(path)
	if err != nil {
		t.Fatalf("legacy artifact must load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("loaded %d edges, want 2", len(loaded))
	}
	for _, e := range loaded {
		if e.Resolution != "" || e.Confidence != 0 {
			t.Errorf("legacy edge %s→%s gained confidence state: %q %v", e.Source, e.Target, e.Resolution, e.Confidence)
		}
	}

	matches, _, err := FindCode(root, "B")
	if err != nil {
		t.Fatalf("FindCode over legacy artifact: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("FindCode returned no matches — nothing was swept")
	}
	for _, m := range matches {
		if m.Confidence != 0 {
			t.Errorf("legacy match confidence = %v, want 0 (unknown)", m.Confidence)
		}
	}
	callers, callees, err := TraceCalls(root, "B", 1)
	if err != nil {
		t.Fatalf("TraceCalls over legacy artifact: %v", err)
	}
	if len(callers) == 0 {
		t.Fatal("TraceCalls returned no callers — nothing was swept")
	}
	for _, e := range append(callers, callees...) {
		if e.Confidence != 0 {
			t.Errorf("legacy trace edge confidence = %v, want 0 (unknown)", e.Confidence)
		}
	}
}

// AC-GEC-011 — provenance shape unchanged: the meta sidecar carries the
// source fingerprints, and edges.jsonl itself stays wall-clock-free (the
// RFC3339 generated_at lives only in the sidecar, never in the JSONL).
func TestProvenanceShapeUnchanged(t *testing.T) {
	root := goldenFixture(t)
	edges, _, err := BuildWithCodeLayers(root)
	if err != nil {
		t.Fatalf("BuildWithCodeLayers: %v", err)
	}
	dir := filepath.Join(root, ".moai", "project", "graph")
	edgesPath := filepath.Join(dir, "edges.jsonl")
	if err := WriteJSONL(edgesPath, edges); err != nil {
		t.Fatalf("WriteJSONL: %v", err)
	}
	if err := WriteEdgesMeta(filepath.Join(dir, MetaFileName), root, SourceFingerprintsForEdges(root), len(edges)); err != nil {
		t.Fatalf("WriteEdgesMeta: %v", err)
	}

	pv, ok := ReadEdgesMeta(filepath.Join(dir, MetaFileName))
	if !ok {
		t.Fatal("edges.meta.json did not parse — provenance sidecar broken")
	}
	if pv.SchemaVersion == 0 {
		t.Error("provenance carries no schema version")
	}
	if pv.TreeRoot != root {
		t.Errorf("provenance tree root = %q, want %q", pv.TreeRoot, root)
	}
	if _, ok := pv.SourceFingerprints["codemaps"]; !ok {
		t.Errorf("source fingerprints missing the codemaps set: %v", pv.SourceFingerprints)
	}

	data, err := os.ReadFile(edgesPath)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		for _, clockKey := range []string{`"generated_at"`, `"captured_at"`, `"timestamp"`, `"built_at"`} {
			if strings.Contains(line, clockKey) {
				t.Errorf("line %d carries wall-clock key %s: %s", i+1, clockKey, line)
			}
		}
	}
	if len(data) == 0 {
		t.Fatal("empty artifact — nothing was swept")
	}
}

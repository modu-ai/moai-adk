package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// readerFixtureEdges builds a small edge set covering all three kinds and
// both direction rules:
//
//	import       cmd/moai -> internal/cli -> internal/config (chain)
//	mx-spec      internal/graph/reader.go -> SPEC-GRAPH-R-001
//	spec-depends SPEC-GRAPH-R-002 -> SPEC-GRAPH-R-001
func readerFixtureEdges() []Edge {
	return []Edge{
		{Kind: KindImport, Source: "cmd/moai", Target: "internal/cli"},
		{Kind: KindImport, Source: "internal/cli", Target: "internal/config"},
		{Kind: KindMXSpec, Source: "internal/graph/reader.go", Target: "SPEC-GRAPH-R-001", Line: 3},
		{Kind: KindSpecDepends, Source: "SPEC-GRAPH-R-002", Target: "SPEC-GRAPH-R-001"},
	}
}

func TestFindCallers_ReverseNeighborsAcrossKinds(t *testing.T) {
	edges := readerFixtureEdges()

	cases := []struct {
		node string
		want []string
	}{
		{"internal/config", []string{"internal/cli"}}, // importer
		// dependent SPEC + tagged file
		{"SPEC-GRAPH-R-001", []string{"SPEC-GRAPH-R-002", "internal/graph/reader.go"}},
		{"cmd/moai", []string{}}, // source-only node: no callers
		{"no-such-node", []string{}},
	}
	for _, tc := range cases {
		got := FindCallers(edges, tc.node)
		if len(got) == 0 && len(tc.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("FindCallers(%q) = %v, want %v", tc.node, got, tc.want)
		}
	}
}

func TestFindCallers_DeduplicatesAndSorts(t *testing.T) {
	edges := []Edge{
		{Kind: KindImport, Source: "b", Target: "x"},
		{Kind: KindImport, Source: "a", Target: "x"},
		{Kind: KindSpecDepends, Source: "a", Target: "x"}, // same source, different kind
	}
	got := FindCallers(edges, "x")
	if want := []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("FindCallers = %v, want %v", got, want)
	}
}

func TestBlastRadius_TransitiveOverReverseEdges(t *testing.T) {
	edges := readerFixtureEdges()

	// cmd/moai -> internal/cli -> internal/config: changing config reaches
	// both importers transitively.
	got := BlastRadius(edges, "internal/config")
	if want := []string{"cmd/moai", "internal/cli"}; !reflect.DeepEqual(got, want) {
		t.Errorf("BlastRadius(internal/config) = %v, want %v", got, want)
	}

	// mx-spec propagates both ways: the file reaches the SPEC it implements,
	// then the SPEC's dependents.
	got = BlastRadius(edges, "internal/graph/reader.go")
	if want := []string{"SPEC-GRAPH-R-001", "SPEC-GRAPH-R-002"}; !reflect.DeepEqual(got, want) {
		t.Errorf("BlastRadius(reader.go) = %v, want %v", got, want)
	}

	// Symmetrically from the SPEC: implementation file + dependent SPEC.
	got = BlastRadius(edges, "SPEC-GRAPH-R-001")
	if want := []string{"SPEC-GRAPH-R-002", "internal/graph/reader.go"}; !reflect.DeepEqual(got, want) {
		t.Errorf("BlastRadius(SPEC-GRAPH-R-001) = %v, want %v", got, want)
	}

	// The start node is never included; a source-only node has no radius.
	if got := BlastRadius(edges, "cmd/moai"); len(got) != 0 {
		t.Errorf("BlastRadius(cmd/moai) = %v, want empty", got)
	}
}

func TestBlastRadius_CycleTerminates(t *testing.T) {
	edges := []Edge{
		{Kind: KindImport, Source: "a", Target: "b"},
		{Kind: KindImport, Source: "b", Target: "a"},
	}
	if got := BlastRadius(edges, "a"); !reflect.DeepEqual(got, []string{"b"}) {
		t.Errorf("BlastRadius on a 2-cycle = %v, want [b]", got)
	}
}

func TestLoadJSONL_RoundTripsWithWriteJSONL(t *testing.T) {
	edges := readerFixtureEdges()
	path := filepath.Join(t.TempDir(), "edges.jsonl")
	if err := WriteJSONL(path, edges); err != nil {
		t.Fatalf("WriteJSONL: %v", err)
	}

	got, err := LoadJSONL(path)
	if err != nil {
		t.Fatalf("LoadJSONL: %v", err)
	}
	if !reflect.DeepEqual(got, edges) {
		t.Errorf("round-trip mismatch:\n got  %v\n want %v", got, edges)
	}
}

func TestLoadJSONL_MalformedLineFailsWithLineNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), "edges.jsonl")
	body := "{\"kind\":\"import\"}\nnot-json\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadJSONL(path)
	if err == nil {
		t.Fatal("malformed line must fail")
	}
	if !strings.Contains(err.Error(), ":2:") {
		t.Errorf("error must cite the 1-based line number, got: %v", err)
	}
}

func TestLoadJSONL_MissingFileFails(t *testing.T) {
	if _, err := LoadJSONL(filepath.Join(t.TempDir(), "absent.jsonl")); err == nil {
		t.Fatal("missing file must fail")
	}
}

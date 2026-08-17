package graph

import (
	"reflect"
	"testing"
)

func TestImportFanIn_RanksDistinctImporters(t *testing.T) {
	edges := []Edge{
		{Kind: KindImport, Source: "pkg/a", Target: "pkg/c"},
		{Kind: KindImport, Source: "pkg/b", Target: "pkg/c"},
		{Kind: KindImport, Source: "pkg/a", Target: "pkg/b"},
		{Kind: KindImport, Source: "pkg/a", Target: "pkg/c"}, // duplicate importer: counted once
		{Kind: KindMXSpec, Source: "f.go", Target: "pkg/c"},  // non-import kind: ignored
	}

	got := ImportFanIn(edges, 0)
	want := []FanInEntry{
		{Package: "pkg/c", FanIn: 2},
		{Package: "pkg/b", FanIn: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ImportFanIn(limit=0) = %v, want %v", got, want)
	}

	if got := ImportFanIn(edges, 1); len(got) != 1 || got[0].Package != "pkg/c" {
		t.Errorf("ImportFanIn(limit=1) = %v, want only pkg/c", got)
	}
}

func TestImportFanIn_TiesBreakByPackagePath(t *testing.T) {
	edges := []Edge{
		{Kind: KindImport, Source: "x", Target: "pkg/z"},
		{Kind: KindImport, Source: "x", Target: "pkg/a"},
	}
	got := ImportFanIn(edges, 0)
	if got[0].Package != "pkg/a" {
		t.Errorf("tie must break by package path asc, got %v", got)
	}
}

func TestUnreferencedSpecs(t *testing.T) {
	edges := []Edge{
		{Kind: KindMXSpec, Source: "a.go", Target: "SPEC-U-001"},
		{Kind: KindSpecDepends, Source: "SPEC-U-003", Target: "SPEC-U-002"}, // not a code reference
	}

	got := UnreferencedSpecs(edges, []string{"SPEC-U-001", "SPEC-U-002", "SPEC-U-003", ""})
	if want := []string{"SPEC-U-002", "SPEC-U-003"}; !reflect.DeepEqual(got, want) {
		t.Errorf("UnreferencedSpecs = %v, want %v", got, want)
	}
}

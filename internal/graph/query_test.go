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

func TestMilestoneClaims(t *testing.T) {
	edges := []Edge{
		{Kind: KindReportMilestone, Source: ".moai/reports/a.md", Target: "a#S1"},
		{Kind: KindMilestoneCard, Source: "a#S1", Target: "t108"},
		{Kind: KindReportMilestone, Source: ".moai/reports/a.md", Target: "a#S6"}, // no card claimed
		{Kind: KindMilestoneCard, Source: "a#S7", Target: "t59"},
		{Kind: KindMilestoneCard, Source: "a#S7", Target: "t58"},
		// A milestone-card edge without its declaring report-milestone edge is
		// dangling claim data, not a declared milestone: not in the output.
		{Kind: KindMilestoneCard, Source: "a#S9", Target: "t77"},
	}

	got := MilestoneClaims(edges)
	// Only S1 and S6 carry report-milestone edges; S7 and S9 appear in the
	// input WITHOUT one, so they must be absent — the declaration, not the
	// claim, admits a milestone into the output.
	if len(got) != 2 {
		t.Fatalf("MilestoneClaims = %+v, want 2 entries (declared milestones only)", got)
	}
	if got[0].Milestone != "a#S1" || !reflect.DeepEqual(got[0].Cards, []string{"t108"}) {
		t.Errorf("first claim = %+v, want a#S1 with [t108]", got[0])
	}
	if got[1].Milestone != "a#S6" || len(got[1].Cards) != 0 {
		t.Errorf("second claim = %+v, want a#S6 with no cards", got[1])
	}
}

func TestMilestoneClaims_SortsCardsWithinClaim(t *testing.T) {
	edges := []Edge{
		{Kind: KindReportMilestone, Source: ".moai/reports/a.md", Target: "a#S7"},
		{Kind: KindMilestoneCard, Source: "a#S7", Target: "t59"},
		{Kind: KindMilestoneCard, Source: "a#S7", Target: "t58"},
	}
	got := MilestoneClaims(edges)
	if want := []string{"t58", "t59"}; !reflect.DeepEqual(got[0].Cards, want) {
		t.Errorf("cards = %v, want sorted %v", got[0].Cards, want)
	}
}

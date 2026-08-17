package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// crossCheckDoc is a milestone-bearing report exercising the parser's
// contract: locale gloss on the header, bold milestone cells, a multi-card
// cell, a new-card marker, and a description cell that MENTIONS a card id
// (t21) that must NOT become a milestone-card edge.
const crossCheckDoc = `# Demo

## 7. Milestones

| # | scope | card |
|---|---|---|
| S3 | naming, absorbs t21 | **t56** |
| S7 | format + glossary | t58 + t59 |
| S6 | chief session | [신규 발행 필요] |

## Card Cross-Check (카드 대조표)

| milestone | scope | card |
|---|---|---|
| **S0** | measure touchpoints | [new card needed] |
| **S3** | session naming (absorbs t21) | **t56** |
| **S7** | format + glossary | t58 + t59 |
| **S6** | chief session | [신규 발행 필요] |
| | empty milestone cell | t99 |

## 8. Later section

Not a table.
`

func TestParseCardCrossCheck_HeaderWithLocaleGloss(t *testing.T) {
	rows := parseCardCrossCheck(crossCheckDoc)

	want := []crossCheckRow{ // table order
		{milestone: "S0", cards: nil},
		{milestone: "S3", cards: []string{"t56"}},
		{milestone: "S7", cards: []string{"t58", "t59"}},
		{milestone: "S6", cards: nil},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("rows = %+v, want %+v", rows, want)
	}
}

func TestParseCardCrossCheck_DescriptionCellCardIdNotClaimed(t *testing.T) {
	// The t21 mention lives in the SCOPE column, not the card column: the
	// mapping reads only the column headed "card".
	for _, row := range parseCardCrossCheck(crossCheckDoc) {
		for _, card := range row.cards {
			if card == "t21" {
				t.Errorf("row %s claimed t21 from a non-card column", row.milestone)
			}
		}
	}
}

func TestParseCardCrossCheck_NoSectionYieldsNothing(t *testing.T) {
	if rows := parseCardCrossCheck("# Plain report\n\nno cross-check anywhere\n"); len(rows) != 0 {
		t.Errorf("rows = %+v, want none", rows)
	}
}

func TestParseCardCrossCheck_NoCardColumnYieldsNothing(t *testing.T) {
	// A table without a card column is not a cross-check table — header
	// detection by name, never by position.
	doc := "## Card Cross-Check\n\n| milestone | scope |\n|---|---|\n| S1 | rename |\n"
	if rows := parseCardCrossCheck(doc); len(rows) != 0 {
		t.Errorf("rows = %+v, want none", rows)
	}
}

func TestReportEdges_FailOpenAndEmission(t *testing.T) {
	root := t.TempDir()

	// Absent reports directory: zero edges, no error (fail-open layer).
	edges, err := reportEdges(root)
	if err != nil {
		t.Fatalf("reportEdges on absent dir: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("edges = %v, want none", edges)
	}

	reports := filepath.Join(root, ".moai", "reports")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reports, "demo-design.md"), []byte(crossCheckDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	// A report without the section and a non-markdown file contribute nothing.
	if err := os.WriteFile(filepath.Join(reports, "plain.md"), []byte("# no section\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reports, "notes.txt"), []byte("## Card Cross-Check\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	edges, err = reportEdges(root)
	if err != nil {
		t.Fatalf("reportEdges: %v", err)
	}
	// Emission order is table order (like the sibling layers, reportEdges
	// does not sort — Build applies EdgeLess to the whole set).
	want := []Edge{
		{Kind: KindReportMilestone, Source: ".moai/reports/demo-design.md", Target: "demo-design#S0"},
		{Kind: KindReportMilestone, Source: ".moai/reports/demo-design.md", Target: "demo-design#S3"},
		{Kind: KindMilestoneCard, Source: "demo-design#S3", Target: "t56"},
		{Kind: KindReportMilestone, Source: ".moai/reports/demo-design.md", Target: "demo-design#S7"},
		{Kind: KindMilestoneCard, Source: "demo-design#S7", Target: "t58"},
		{Kind: KindMilestoneCard, Source: "demo-design#S7", Target: "t59"},
		{Kind: KindReportMilestone, Source: ".moai/reports/demo-design.md", Target: "demo-design#S6"},
	}
	if !reflect.DeepEqual(edges, want) {
		t.Errorf("edges:\n got  %+v\n want %+v", edges, want)
	}
}

func TestReportEdges_TwoReportsDoNotCollide(t *testing.T) {
	root := t.TempDir()
	reports := filepath.Join(root, ".moai", "reports")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := "## Card Cross-Check\n\n| milestone | card |\n|---|---|\n| S0 | t1 |\n"
	for _, name := range []string{"a-design.md", "b-design.md"} {
		if err := os.WriteFile(filepath.Join(reports, name), []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	edges, err := reportEdges(root)
	if err != nil {
		t.Fatalf("reportEdges: %v", err)
	}
	var milestones int
	for _, e := range edges {
		if e.Kind == KindReportMilestone {
			milestones++
		}
	}
	if milestones != 2 {
		t.Errorf("report-milestone edges = %d, want 2 (stem-qualified S0s must not dedup away)", milestones)
	}
}

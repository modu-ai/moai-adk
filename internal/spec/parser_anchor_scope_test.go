// parser_anchor_scope_test.go — SPEC-AC-ANCHOR-SCOPE-001 (card t747) unit
// fixtures for the findACSectionStart anchor-scope repair.
//
// Every fixture shape is derived from a measured corpus file (frozen lists:
// .moai/reports/t747/probe/), not invented:
//
//   - narrow axis, numeric-sub colon form  — SPEC-CC297-001 / SPEC-STATUS-AUTO-001
//     (`## Requirements` → `### REQ-N` → `- AC-1.1: ...`; no acceptance-vocabulary
//     heading anywhere, so the pre-repair anchor is -1)
//   - narrow axis, parseable-id colon form — SPEC-AC-COLLECTOR-ANCHOR-001
//   - loose axis, empty-vocabulary fallback — .moai/reports/t747/probe/empty-anchor.txt
//   - prose bound                           — the in=10/out=1 prose layer
//     (SPEC-CLAUDEMD-DIET-V2-001 / SPEC-DB-SYNC-HARDEN-001 shape: bullets that
//     mention an AC id WITHOUT the required separator never name a region)
//   - negative markers                      — acNegativeSectionMarkers headings
//     that mention acceptance criteria must never anchor (acceptance.md §D.1)
//
// The repair under test (plan §C decisions, binding): declaration-presence-based
// region selection (narrow) + declaration-aware terminal fallback (loose) — the
// region qualifier is the AC-shaped list form (bullet + AC-…: colon, separator
// REQUIRED), never a bare id mention, and the line grammar (parseSingleACLine)
// is untouched (t528 PRESERVE surface).
package spec

import (
	"strings"
	"testing"
)

// t747FixtureCase pairs an inline spec.md shape with the anchor the repaired
// selector must produce. wantStart indexes 0-based lines (findACSectionStart's
// return value: the line after the anchoring heading; -1 = no anchor).
type t747FixtureCase struct {
	name      string
	markdown  string
	wantStart int
}

func t747RunAnchorFixtures(t *testing.T, cases []t747FixtureCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(tc.markdown, "\n")
			if got := findACSectionStart(lines); got != tc.wantStart {
				t.Errorf("findACSectionStart = %d, want %d\nmarkdown:\n%s", got, tc.wantStart, tc.markdown)
			}
		})
	}
}

// TestFindACSectionStart_NarrowAxis_DeclarationNamedRegion — REQ-ACAS-001.
// A document with AC-declaration lines but no acceptance-vocabulary heading
// must anchor a region that includes them, not return no anchor.
func TestFindACSectionStart_NarrowAxis_DeclarationNamedRegion(t *testing.T) {
	t747RunAnchorFixtures(t, []t747FixtureCase{
		{
			// SPEC-CC297-001 shape: numeric-sub colon declarations under
			// level-3 REQ headings; no vocabulary heading anywhere.
			name: "numeric sub IDs under REQ headings",
			markdown: `# SPEC-FIXTURE-001: Title

## HISTORY

- 0.1.0 initial draft

## Requirements

### REQ-1: Feature (Priority: HIGH)

- AC-1.1: First behavior
- AC-1.2: Second behavior

## Scope

### Out of Scope

- unrelated work
`,
			wantStart: 7, // the line after "## Requirements"
		},
		{
			// Level-3 declarations nested inside a level-2 non-vocabulary
			// section (acceptance.md §D.1 edge): the level-2 heading's region
			// reads through its own deeper subheadings.
			name: "level-3 declarations inside level-2 section",
			markdown: `# SPEC-FIXTURE-008: Title

## Design

### REQ-1: detail

- AC-1.1: nested behavior

## Notes

- plain text
`,
			wantStart: 3, // the line after "## Design"
		},
	})
}

// TestFindACSectionStart_NarrowAxis_ParseableDeclarationsParse — REQ-ACAS-001
// end to end: when the narrowed anchor's declarations carry parseable ids, the
// parse recovers criteria instead of reporting the section missing
// (SPEC-AC-COLLECTOR-ANCHOR-001 shape).
func TestFindACSectionStart_NarrowAxis_ParseableDeclarationsParse(t *testing.T) {
	markdown := `# SPEC-FIXTURE-002: Title

## Verification

- AC-FIX-001: Given a parseable declaration, the parser recovers it

## Notes

- plain text
`
	criteria, errs := ParseAcceptanceCriteria(markdown, false)
	if len(criteria) != 1 {
		t.Fatalf("ParseAcceptanceCriteria criteria = %d, want 1 (errors: %v)", len(criteria), errs)
	}
	if criteria[0].ID != "AC-FIX-001" {
		t.Errorf("criteria[0].ID = %q, want AC-FIX-001", criteria[0].ID)
	}
	for _, err := range errs {
		if strings.Contains(err.Error(), "not found") {
			t.Errorf("parse reported the section missing despite a recoverable declaration: %v", err)
		}
	}
}

// TestFindACSectionStart_LooseAxis_DeclarationRegionOverEmptyVocabulary —
// REQ-ACAS-002. When every vocabulary section is empty of AC lines and
// declaration-bearing regions exist under other headings, the anchor must move
// to a declaration-bearing region, not sit on the empty first vocabulary
// section (the empty-anchor.txt defect shape).
func TestFindACSectionStart_LooseAxis_DeclarationRegionOverEmptyVocabulary(t *testing.T) {
	t747RunAnchorFixtures(t, []t747FixtureCase{
		{
			name: "empty acceptance section, declarations under later heading",
			markdown: `# SPEC-FIXTURE-003: Title

## Acceptance Criteria

## Requirements

- AC-LOOSE-001: real behavior
`,
			wantStart: 5, // the line after "## Requirements"
		},
		{
			name: "empty acceptance section, declarations under earlier heading",
			markdown: `# SPEC-FIXTURE-009: Title

## Requirements

- AC-LOOSE-002: real behavior

## Acceptance Criteria

## Follow-up
`,
			wantStart: 3, // the line after "## Requirements"
		},
	})
}

// TestFindACSectionStart_ProseBound_ColonlessMentionNeverNamesRegion —
// REQ-ACAS-004. A bullet that mentions an AC id without the required
// AC-…: separator is prose; it must never name a region, and the empty
// vocabulary section keeps the anchor (the pre-repair fallback).
func TestFindACSectionStart_ProseBound_ColonlessMentionNeverNamesRegion(t *testing.T) {
	t747RunAnchorFixtures(t, []t747FixtureCase{
		{
			name: "colon-less AC mentions stay outside region selection",
			markdown: `# SPEC-FIXTURE-004: Title

## Acceptance Criteria

See the plan for criteria.

## Notes

- AC-1 was discussed in review
- see AC-2 for the counterexample
`,
			wantStart: 3, // the empty acceptance section keeps the anchor
		},
	})
}

// TestFindACSectionStart_NegativeMarker_NeverNamesRegion — acceptance.md §D.1:
// acNegativeSectionMarkers headings that mention acceptance criteria must
// continue to never anchor, including via the new declaration-presence
// criteria.
func TestFindACSectionStart_NegativeMarker_NeverNamesRegion(t *testing.T) {
	t747RunAnchorFixtures(t, []t747FixtureCase{
		{
			name: "out-of-scope section holding declarations never anchors",
			markdown: `# SPEC-FIXTURE-005: Title

### Out of Scope

- AC-1.1: explicitly excluded behavior

## Real Section

- AC-2.1: included behavior
`,
			wantStart: 7, // "## Real Section", never "### Out of Scope"
		},
		{
			name: "out-of-scope-only document stays unanchored",
			markdown: `# SPEC-FIXTURE-010: Title

### Out of Scope

- AC-1.1: excluded behavior
`,
			wantStart: -1,
		},
	})
}

// TestFindACSectionStart_ControlPreservation — REQ-ACAS-003 at unit scale.
// Documents the repair must not move: a parseable vocabulary section (the
// dominant anchored shape), and a fallback vocabulary section whose own region
// already holds the colon-form declarations.
func TestFindACSectionStart_ControlPreservation(t *testing.T) {
	t747RunAnchorFixtures(t, []t747FixtureCase{
		{
			name: "parseable vocabulary section keeps the anchor",
			markdown: `# SPEC-FIXTURE-006: Title

## Acceptance Criteria

- AC-FIX-001: Given a normal AC, the parse is unchanged
`,
			wantStart: 3,
		},
		{
			// The control file whose fallback vocabulary section itself holds
			// colon-form (numeric-sub) declarations: the declaration-aware
			// fallback must re-select the SAME section, not move the anchor.
			name: "fallback vocabulary section holding colon-form declarations keeps the anchor",
			markdown: `# SPEC-FIXTURE-007: Title

## Acceptance Criteria

- AC-1.1: numeric sub form declared here
`,
			wantStart: 3,
		},
	})
}

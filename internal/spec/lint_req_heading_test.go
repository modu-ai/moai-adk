package spec

import (
	"strings"
	"testing"
)

// Heading-form collection tests — SPEC-HEADING-REQ-COLLECT-001 (card t894).
//
// The load-bearing criterion here is AC-HRC-002, and it needs BOTH directions
// asserted. Variant A (heading title as Text) and variant B (first body
// paragraph as Text) collect the SAME entries at the SAME lines; they differ
// only in Text. An assertion that an entry was collected therefore cannot tell
// the two apart, and a test written that way passes under the variant the SPEC
// rejects.

// headingPositiveBody is the AC-HRC-001 shape: a level-3 heading definition
// followed by a blank line and a SHALL-bearing statement.
const headingPositiveBody = `## §B. Requirements

### REQ-FIXH-001 — Ubiquitous — the collector shape

The system SHALL collect heading-form requirement definitions.

### REQ-FIXH-002 — Event-driven (When) — a section title

**When** the trigger fires, the system SHALL respond.
`

// headingNoParagraphBody is the AC-HRC-003 shape: a heading definition with no
// non-empty line before the next heading.
const headingNoParagraphBody = `### REQ-FIXH-004 — Ubiquitous — a title with no body

### Another section

The system SHALL do something unrelated.
`

// headingCrossLevelBoundaryBody is the AC-HRC-004 shape. The boundary heading
// is a DIFFERENT level (h2) than the one that produced the entry: a search
// bounded only by h3 walks straight past it and attributes the next section's
// sentence to REQ-FIXH-003, which a same-level-only fixture would never catch.
const headingCrossLevelBoundaryBody = `### REQ-FIXH-003 — Ubiquitous — a title whose section is empty

## Some Following Section

The system SHALL emit a distinctivetoken belonging to the next section.
`

func findEntry(t *testing.T, entries []REQEntry, id string) REQEntry {
	t.Helper()
	for _, e := range entries {
		if e.ID == id {
			return e
		}
	}
	t.Fatalf("no entry collected for %s (collected %d entries)", id, len(entries))
	return REQEntry{}
}

// --- AC-HRC-001 (REQ-HRC-001, 002, 003) — collection -----------------------

func TestHeadingCollection_CollectsLevel3Definitions(t *testing.T) {
	entries := parseREQsHeadingForm(headingPositiveBody)
	if len(entries) != 2 {
		t.Fatalf("collected %d entries, want 2: %+v", len(entries), entries)
	}
	if entries[0].ID != "REQ-FIXH-001" || entries[1].ID != "REQ-FIXH-002" {
		t.Errorf("entries out of document order: %q then %q", entries[0].ID, entries[1].ID)
	}
	if entries[0].Line >= entries[1].Line {
		t.Errorf("Line values not ascending: %d then %d", entries[0].Line, entries[1].Line)
	}
}

// TestHeadingCollection_IgnoresOtherHeadingLevels pins the `###`-only anchor
// (spec.md §D). `##` and `####` have no measured population, and widening the
// anchor for an unmeasured one is exactly the change §A.4 forbids.
func TestHeadingCollection_IgnoresOtherHeadingLevels(t *testing.T) {
	body := `## REQ-FIXH-010 — h2 title

#### REQ-FIXH-011 — h4 title

##### REQ-FIXH-012 — h5 title
`
	if entries := parseREQsHeadingForm(body); len(entries) != 0 {
		t.Fatalf("collected %d entries from non-h3 headings, want 0: %+v", len(entries), entries)
	}
}

// --- AC-HRC-002 (REQ-HRC-004) — the variant-A/B decision -------------------

func TestHeadingCollection_TextIsBodyParagraphNotHeadingTitle(t *testing.T) {
	entry := findEntry(t, parseREQsHeadingForm(headingPositiveBody), "REQ-FIXH-002")

	const wantParagraph = "**When** the trigger fires, the system SHALL respond."
	const headingTitle = "Event-driven (When) — a section title"

	// Direction 1 — the entry carries the statement.
	if entry.Text != wantParagraph {
		t.Errorf("Text = %q, want the body paragraph %q", entry.Text, wantParagraph)
	}
	// Direction 2 — the entry does NOT carry the title. Without this, the test
	// passes under variant A whenever direction 1 is loosened to a containment
	// check, which is the exact failure AC-HRC-002 names.
	if entry.Text == headingTitle {
		t.Errorf("Text equals the heading title %q — variant A shipped, not variant B", headingTitle)
	}
}

// --- AC-HRC-003 (REQ-HRC-005) — the fallback -------------------------------

func TestHeadingCollection_FallsBackToHeadingTextNeverEmpty(t *testing.T) {
	entry := findEntry(t, parseREQsHeadingForm(headingNoParagraphBody), "REQ-FIXH-004")

	if entry.Text == "" {
		t.Fatal("Text is empty — an empty Text fires ModalityUnjudged for a collector reason, re-manufacturing the constant signal variant A produces")
	}
	if want := "Ubiquitous — a title with no body"; entry.Text != want {
		t.Errorf("Text = %q, want the heading's own trailing text %q", entry.Text, want)
	}
}

// --- AC-HRC-004 (REQ-HRC-006) — the bounded search -------------------------

func TestHeadingCollection_SearchStopsAtNextHeadingOfAnyLevel(t *testing.T) {
	entry := findEntry(t, parseREQsHeadingForm(headingCrossLevelBoundaryBody), "REQ-FIXH-003")

	if want := "Ubiquitous — a title whose section is empty"; entry.Text != want {
		t.Errorf("Text = %q, want the AC-HRC-003 fallback %q", entry.Text, want)
	}
	if strings.Contains(entry.Text, "distinctivetoken") {
		t.Errorf("Text %q carries prose from the FOLLOWING section — the search crossed a `##` boundary, which is a manufactured judgment", entry.Text)
	}
}

// --- REQ-HRC-007 — provenance flag at entry level --------------------------

func TestHeadingCollection_EveryEntryIsWidened(t *testing.T) {
	entries := parseREQsHeadingForm(headingPositiveBody)
	if len(entries) == 0 {
		t.Fatal("no entries collected — the assertion below would be vacuous")
	}
	for _, e := range entries {
		if !e.Widened {
			t.Errorf("%s: Widened = false — its error-severity findings would arrive at error severity and the change would start gating", e.ID)
		}
		if e.Source != REQSourceHeading {
			t.Errorf("%s: Source = %v, want REQSourceHeading", e.ID, e.Source)
		}
	}
}

func TestREQSourceHeading_String(t *testing.T) {
	if got := REQSourceHeading.String(); got != "heading" {
		t.Errorf("REQSourceHeading.String() = %q, want %q", got, "heading")
	}
}

// --- REQ-HRC-003 — the merge keeps document order and existing entries -----

func TestHeadingCollection_MergedInDocumentOrder(t *testing.T) {
	body := `- **REQ-FIXH-020** — the list form SHALL stay first.

### REQ-FIXH-021 — Ubiquitous — the heading form

The system SHALL come second.

- **REQ-FIXH-022** — the list form SHALL come third.
`
	entries := parseREQsWithProvenance(body)
	var ids []string
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	want := []string{"REQ-FIXH-020", "REQ-FIXH-021", "REQ-FIXH-022"}
	if len(ids) != len(want) {
		t.Fatalf("collected %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("collected %v, want %v (document order broken)", ids, want)
		}
	}

	// The pre-existing list entries keep their own Text — the heading source
	// must not reach them.
	list := findEntry(t, entries, "REQ-FIXH-020")
	if list.Source != REQSourceList {
		t.Errorf("REQ-FIXH-020: Source = %v, want REQSourceList", list.Source)
	}
	if list.Text != "the list form SHALL stay first." {
		t.Errorf("REQ-FIXH-020: Text = %q — a list entry was perturbed by the heading merge", list.Text)
	}
}

// --- AC-HRC-006 (REQ-HRC-008) — Source never decides severity --------------
//
// Mirrors TestTableCollection_SourceDoesNotDecideSeverity, extended over the
// three-value Source set. The grep in AC-HRC-006 sees only one spelling of the
// prohibition; this flip is the load-bearing half.
func TestHeadingCollection_SourceDoesNotDecideSeverity(t *testing.T) {
	entries := append(parseREQsWithProvenance(headingPositiveBody), parseREQsWithProvenance(headingNoParagraphBody)...)
	if len(entries) == 0 {
		t.Fatal("no entries collected — the flip test would be vacuous")
	}

	distribution := func(rs []REQEntry) map[Severity]int {
		d := map[Severity]int{}
		for _, r := range rs {
			sev, _ := reqFindingSeverity(r, SeverityError)
			d[sev]++
		}
		return d
	}

	before := distribution(entries)

	// Rotate every Source through all three values rather than flipping a pair:
	// a two-value flip leaves the third value untested, and the third value is
	// the one this card adds.
	for _, next := range []REQSource{REQSourceList, REQSourceTable, REQSourceHeading} {
		rotated := make([]REQEntry, len(entries))
		copy(rotated, entries)
		for i := range rotated {
			rotated[i].Source = next
		}
		after := distribution(rotated)
		if len(before) != len(after) {
			t.Fatalf("severity distribution changed shape with Source=%v: %v -> %v", next, before, after)
		}
		for sev, n := range before {
			if after[sev] != n {
				t.Errorf("Source=%v, severity %v: %d before, %d after — Source must not reach reqFindingSeverity", next, sev, n, after[sev])
			}
		}
	}
}

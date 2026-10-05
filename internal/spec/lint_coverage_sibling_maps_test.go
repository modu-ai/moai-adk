package spec

import (
	"reflect"
	"sort"
	"testing"
)

// sortedMapsIDs returns a sorted copy so a test compares sets rather than the
// incidental order two different extraction paths happen to produce.
func sortedMapsIDs(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

// TestSiblingMapsNumericTailExpanded is the positive case: a bare numeric tail
// following a full id inside one `maps` section expands to the full id formed
// from that id's prefix (REQ-SMS-001).
func TestSiblingMapsNumericTailExpanded(t *testing.T) {
	text := "- AC-FIXA-001 (maps REQ-FIXA-001, 002): Given one, When two, Then three.\n" +
		"- AC-FIXA-002 (maps REQ-FIXA-003): Given one, When two, Then three.\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-FIXA-001", "REQ-FIXA-002", "REQ-FIXA-003"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v", got, want)
	}
}

// TestSiblingMapsBareTailWithoutPrecedingIDNotExpanded is the no-preceding-id
// negative control (REQ-SMS-006, fixture E): a tail with no full id before it
// inside its own capture yields nothing. A prefix is never inferred from the
// AC id, from a neighbouring line, or from the containing document.
func TestSiblingMapsBareTailWithoutPrecedingIDNotExpanded(t *testing.T) {
	text := "- AC-FIXE-001 (maps 002): Given one, When two, Then three.\n"

	if got := siblingMapsREQIDs(text); len(got) != 0 {
		t.Fatalf("siblingMapsREQIDs = %v, want no ids (no full id precedes the tail)", got)
	}
}

// TestSiblingMapsCaptureDoesNotOverReachTheLine is the over-reach positive
// control (REQ-SMS-004, fixture F). The textual unit is the capture of the
// widened `maps` locator, NOT the line: the capture ends at the first element
// that is neither a full id nor a bare numeric tail, so trailing prose naming
// further REQ ids stays outside it.
//
// A line-scoped implementation returns four ids here and fails this test. That
// is the whole point of the control: over-reach silences genuine
// CoverageIncomplete findings, which is this card's own defect class.
func TestSiblingMapsCaptureDoesNotOverReachTheLine(t *testing.T) {
	text := "- AC-FIXF-001 (maps REQ-FIXF-001, 002) — verifies REQ-FIXF-003, 004 are unreachable.\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-FIXF-001", "REQ-FIXF-002"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (003/004 lie outside the capture)", got, want)
	}
}

// TestSiblingMapsDoesNotCrossALineBoundary is the line-boundary control
// (REQ-SMS-005, fixture G). The locator's inter-element separator admits
// horizontal whitespace only, so a `maps` section ends at the end of its line
// and a section ending in a comma cannot absorb the next line.
func TestSiblingMapsDoesNotCrossALineBoundary(t *testing.T) {
	text := "- AC-FIXG-001 (maps REQ-FIXG-001,\n  002): Given one, When two, Then three.\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-FIXG-001"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (the run is broken by a newline)", got, want)
	}
}

// TestSiblingMapsDoesNotCrossIntoAnotherSection asserts each `maps` section is
// its own unit: a tail in one section never takes a prefix from a previous one
// (REQ-SMS-005, second sentence).
func TestSiblingMapsDoesNotCrossIntoAnotherSection(t *testing.T) {
	text := "- AC-X-001 (maps REQ-X-001): Given one, When two, Then three.\n" +
		"- AC-X-002 (maps 002): Given one, When two, Then three.\n"

	got := sortedMapsIDs(siblingMapsREQIDs(text))
	want := []string{"REQ-X-001"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("siblingMapsREQIDs = %v, want %v (the second section has no full id of its own)", got, want)
	}
}

// TestSiblingMapsDeclinedTailForms asserts the forms spec.md §D declines are in
// fact declined: a tail with a non-numeric body, a tail separated by something
// other than a comma, and a full id written without the `REQ-` prefix. Each case
// must contribute only the ids the author wrote in full.
func TestSiblingMapsDeclinedTailForms(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "non-numeric tail body",
			text: "- AC-D-001 (maps REQ-D-001, 00b): text.\n",
			want: []string{"REQ-D-001"},
		},
		{
			name: "separator other than a comma",
			text: "- AC-D-002 (maps REQ-D-001 002): text.\n",
			want: []string{"REQ-D-001"},
		},
		{
			name: "full id without the REQ- prefix",
			text: "- AC-D-003 (maps D-001, 002): text.\n",
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := siblingMapsREQIDs(tc.text)
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(sortedMapsIDs(got), sortedMapsIDs(tc.want)) {
				t.Fatalf("siblingMapsREQIDs = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSiblingMapsFullIDOnlyListUnchanged is the no-regression direction. The
// widened locator re-reads every `maps REQ-…` section in the corpus (1099
// occurrences measured in the working tree of 881aa4bb8 as of 2026-09-18), and
// every one of them must produce the same id set it produces today unless it
// carries a numeric tail. The assertion is made against the id set
// ExtractRequirementMappings — the pre-existing, unmodified extractor — produces
// for the same text, so this test compares the new path to the live behaviour
// rather than to a hand-copied expectation.
func TestSiblingMapsFullIDOnlyListUnchanged(t *testing.T) {
	cases := []string{
		"- AC-FIXB-001 (maps REQ-FIXB-001, REQ-FIXB-002): Given one, When two, Then three.\n" +
			"- AC-FIXB-002 (maps REQ-FIXB-003): Given one, When two, Then three.\n",
		"- AC-Q-001 (maps REQ-Q-001): text.\n",
		"- AC-Q-002 (maps REQ-Q-001, REQ-Q-002, REQ-Q-003) — trailing prose.\n",
		"Maps REQ-CASE-001, REQ-CASE-002 in a sentence.\n",
	}

	for _, text := range cases {
		var baseline []string
		for _, id := range ExtractRequirementMappings(text) {
			baseline = append(baseline, "REQ-"+id)
		}

		got := sortedMapsIDs(siblingMapsREQIDs(text))
		want := sortedMapsIDs(baseline)

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("full-id-only %q: siblingMapsREQIDs = %v, want %v (must match the existing extractor)", text, got, want)
		}
		if len(want) == 0 {
			t.Fatalf("full-id-only %q produced an empty baseline — the case asserts nothing", text)
		}
	}
}

// TestSiblingMapsExpansionSharesTheTableRule is the shared-rule test named by
// AC-SMS-010 command 3. It exercises the table path and the `maps` path on the
// SAME element run through the SAME helper (cellREQIDs) and asserts identical
// tail expansion, closing the "invokes the helper but discards its result" gap
// that a grep alone cannot see.
//
// REQ-SMS-002 forbids a second, independently-written expansion rule. If one
// were introduced on the `maps` path, this test is what notices the two rules
// disagreeing.
func TestSiblingMapsExpansionSharesTheTableRule(t *testing.T) {
	// One element run, three surfaces.
	const run = "REQ-SH-001, 002, 004"

	// Surface 1 — the shared helper, called directly.
	shared := sortedMapsIDs(cellREQIDs(run))

	// Surface 2 — the table path (card t561), reaching the helper through a
	// requirement-headed column.
	tableText := "| AC | REQ |\n|---|---|\n| AC-SH-001 | " + run + " |\n"
	viaTable := sortedMapsIDs(siblingTableREQIDs(tableText))

	// Surface 3 — the `maps` path (this card), reaching the helper through the
	// widened locator's capture.
	mapsText := "- AC-SH-001 (maps " + run + "): Given one, When two, Then three.\n"
	viaMaps := sortedMapsIDs(siblingMapsREQIDs(mapsText))

	want := []string{"REQ-SH-001", "REQ-SH-002", "REQ-SH-004"}

	if !reflect.DeepEqual(shared, want) {
		t.Fatalf("shared helper = %v, want %v", shared, want)
	}
	if !reflect.DeepEqual(viaTable, shared) {
		t.Fatalf("table path = %v, shared helper = %v — the two rules disagree", viaTable, shared)
	}
	if !reflect.DeepEqual(viaMaps, shared) {
		t.Fatalf("maps path = %v, shared helper = %v — the two rules disagree", viaMaps, shared)
	}
}

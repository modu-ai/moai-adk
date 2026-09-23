package spec

import "testing"

// Card t1104 — RED reproduction.
//
// A REQ definition written WITHOUT a leading list marker is not collected by
// any of the three existing collectors (list / table / heading), so every rule
// that consumes doc.REQs is silent on such a document. The lead measured the
// same asymmetry through the binary: the marker-less line produced 0 REQ
// coverage findings, and prefixing it with "- " produced 1.
const bareREQFixture = `## Requirements

**REQ-PRB-001** (Ubiquitous) — The system SHALL collect REQ definitions.

**REQ-PRB-002**: The system SHALL keep prose citations out of the REQ set.
`

func TestParseREQsBare_CollectsMarkerlessDefinition(t *testing.T) {
	got := parseREQsWithProvenance(bareREQFixture)

	want := map[string]string{
		"REQ-PRB-001": "The system SHALL collect REQ definitions.",
		"REQ-PRB-002": "The system SHALL keep prose citations out of the REQ set.",
	}
	if len(got) != len(want) {
		t.Fatalf("collected %d REQ entries, want %d: %+v", len(got), len(want), got)
	}
	for _, e := range got {
		expected, ok := want[e.ID]
		if !ok {
			t.Errorf("unexpected REQ %s collected", e.ID)
			continue
		}
		if e.Text != expected {
			t.Errorf("%s: Text = %q, want %q", e.ID, e.Text, expected)
		}
		if !e.Widened {
			t.Errorf("%s: Widened = false, want true (no narrow pattern reaches a bare line)", e.ID)
		}
		if e.Source != REQSourceBare {
			t.Errorf("%s: Source = %v, want REQSourceBare", e.ID, e.Source)
		}
	}
}

// Negative control: a REQ REFERENCE inside prose is not a definition. The
// anchor is what separates the two — a definition line OPENS with the ID, a
// citation does not.
func TestParseREQsBare_DoesNotCollectProseCitation(t *testing.T) {
	const body = `## Notes

See **REQ-PRB-001**: the definition lives in the section above.

The gate described by REQ-PRB-002 — which fires on every turn — is advisory.

Blank prefix indentation is also not a definition:
  **REQ-PRB-003** — indented continuation text.
`
	if got := parseREQsWithProvenance(body); len(got) != 0 {
		t.Errorf("collected %d entries from prose citations, want 0: %+v", len(got), got)
	}
}

// Negative control: the three existing shapes MUST NOT be double-collected by
// the bare collector — each line still yields exactly one entry, with the
// Source its own collector assigns.
func TestParseREQsBare_NoDoubleCollectionOfExistingShapes(t *testing.T) {
	const body = `## Requirements

- **REQ-PRB-001** — list shape SHALL stay a list entry.

| ID | Statement |
|----|-----------|
| REQ-PRB-002 | table shape SHALL stay a table entry. |

### REQ-PRB-003 — heading shape

The heading statement SHALL stay a heading entry.
`
	got := parseREQsWithProvenance(body)
	if len(got) != 3 {
		t.Fatalf("collected %d entries, want 3: %+v", len(got), got)
	}
	wantSource := map[string]REQSource{
		"REQ-PRB-001": REQSourceList,
		"REQ-PRB-002": REQSourceTable,
		"REQ-PRB-003": REQSourceHeading,
	}
	for _, e := range got {
		if want := wantSource[e.ID]; e.Source != want {
			t.Errorf("%s: Source = %v, want %v", e.ID, e.Source, want)
		}
	}
}

func TestREQSourceBare_String(t *testing.T) {
	if got := REQSourceBare.String(); got != "bare" {
		t.Errorf("REQSourceBare.String() = %q, want %q", got, "bare")
	}
}

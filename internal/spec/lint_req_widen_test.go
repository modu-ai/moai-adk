package spec

import (
	"strings"
	"testing"
)

// TestParseREQsWide_SixObservedShapes asserts that the widened REQ-line pattern
// recognizes every REQ ID shape observed in the corpus, plus the narrow shape the
// current reqLinePattern already accepts.
//
// AC-CRS-001-001 (maps REQ-CRS-001-001).
func TestParseREQsWide_SixObservedShapes(t *testing.T) {
	cases := []struct {
		name string
		id   string
		line string
	}{
		{"three-segment", "REQ-HOOK-001", "- REQ-HOOK-001: The system SHALL do X."},
		{"digits-in-domain", "REQ-WF001-001", "- REQ-WF001-001: The system SHALL do X."},
		{"five-segment", "REQ-VNRN-RT-001-001", "- REQ-VNRN-RT-001-001: The system SHALL do X."},
		{"two-part-domain", "REQ-HRN-FND-001", "- REQ-HRN-FND-001: The system SHALL do X."},
		{"trailing-digit-domain", "REQ-TUX1-001", "- REQ-TUX1-001: The system SHALL do X."},
		{"alnum-domain", "REQ-WC01-001", "- REQ-WC01-001: The system SHALL do X."},
		{"narrow-current", "REQ-SPC-003-001", "- REQ-SPC-003-001: The system SHALL do X."},
		{"bold-marker", "REQ-HOOK-002", "- **REQ-HOOK-002**: The system SHALL do X."},
		{"asterisk-bullet", "REQ-HOOK-003", "* REQ-HOOK-003: The system SHALL do X."},
		{"indented", "REQ-HOOK-004", "  - REQ-HOOK-004: The system SHALL do X."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqs := parseREQsWide(tc.line)
			if len(reqs) != 1 {
				t.Fatalf("expected 1 REQ from %q, got %d (%v)", tc.line, len(reqs), reqs)
			}
			if reqs[0].ID != tc.id {
				t.Errorf("expected ID %q, got %q", tc.id, reqs[0].ID)
			}
			if reqs[0].Text == "" {
				t.Errorf("expected non-empty text for %q", tc.line)
			}
		})
	}
}

// TestParseREQsWide_CountEqualsDefinitionLines asserts that a body containing
// exactly N recognizable REQ definition lines yields exactly N entries — no
// silent drops.
//
// AC-CRS-001-004 (maps REQ-CRS-001-005).
func TestParseREQsWide_CountEqualsDefinitionLines(t *testing.T) {
	body := strings.Join([]string{
		"## 2. Requirements",
		"",
		"- REQ-HOOK-001: first.",
		"- REQ-WF001-001: second.",
		"- REQ-VNRN-RT-001-001: third.",
		"- REQ-HRN-FND-001: fourth.",
		"- REQ-TUX1-001: fifth.",
		"- REQ-WC01-001: sixth.",
		"- REQ-SPC-003-001: seventh.",
		"",
	}, "\n")

	const want = 7
	reqs := parseREQsWide(body)
	if len(reqs) != want {
		t.Fatalf("expected %d REQ entries, got %d (%v)", want, len(reqs), reqs)
	}

	seen := make(map[string]bool, len(reqs))
	for _, r := range reqs {
		if seen[r.ID] {
			t.Errorf("duplicate ID collected: %s", r.ID)
		}
		seen[r.ID] = true
		if r.Line <= 0 {
			t.Errorf("REQ %s has non-positive line number %d", r.ID, r.Line)
		}
	}
}

// TestParseREQsWide_ProseMentionsNotCollected asserts that lines merely MENTIONING
// a REQ token — with no "- ID: text" definition shape — are not collected.
//
// AC-CRS-001-004 negative direction (maps REQ-CRS-001-005).
func TestParseREQsWide_ProseMentionsNotCollected(t *testing.T) {
	body := strings.Join([]string{
		"The rule REQ-HOOK-001 is enforced by the linter.",
		"See REQ-WF001-001 for details.",
		"| REQ-VNRN-RT-001-001 | covered | yes |",
		"maps REQ-HRN-FND-001",
		"REQ-TUX1-001: this is not a list item.",
		"- not a req line: REQ-WC01-001 appears after the colon.",
	}, "\n")

	reqs := parseREQsWide(body)
	if len(reqs) != 0 {
		t.Fatalf("expected 0 REQ entries from prose-only body, got %d (%v)", len(reqs), reqs)
	}
}

// TestParseREQsWide_NarrowSubsetOnFixture asserts that on a fixture containing
// both narrow-shape and wide-only-shape definition lines in canonical list form,
// every narrow match is also a wide match.
func TestParseREQsWide_NarrowSubsetOnFixture(t *testing.T) {
	body := strings.Join([]string{
		"- REQ-SPC-003-001: narrow shape.",
		"- REQ-HOOK-001: wide-only shape.",
		"- REQ-ABC-123-456: narrow shape two.",
	}, "\n")

	narrow := parseREQs(body)
	wide := parseREQsWide(body)

	wideSet := make(map[string]bool, len(wide))
	for _, r := range wide {
		wideSet[r.ID] = true
	}
	for _, r := range narrow {
		if !wideSet[r.ID] {
			t.Errorf("narrow match %q absent from wide set", r.ID)
		}
	}
	if len(wide) != 3 {
		t.Errorf("expected 3 wide matches, got %d", len(wide))
	}
}

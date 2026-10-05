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

// TestParseREQsWide_FormA_ColonBehaviorUnchanged pins the exact entries the
// pattern produces for the colon-separator forms (Form A in the t385 census):
// plain colon, bold-inside (`**REQ-X:**`) and bold-outside (`**REQ-X**:`). The
// separator/paren widening must be a strict superset — every entry here,
// including the Text segment and the line number, must be byte-identical before
// and after the widening. The stray `** ` prefix in bold-inside Text is the
// pattern's long-standing capture and is pinned as measured, not "cleaned up"
// here: cleaning it would be a behavior change riding a widening card.
func TestParseREQsWide_FormA_ColonBehaviorUnchanged(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantID   string
		wantText string
	}{
		{"plain-colon", "- REQ-SPC-003-001: The system SHALL do X.",
			"REQ-SPC-003-001", "The system SHALL do X."},
		{"bold-inside-colon", "- **REQ-HOOK-001:** The system SHALL do X.",
			"REQ-HOOK-001", "** The system SHALL do X."},
		{"bold-outside-colon", "- **REQ-HOOK-002**: The system SHALL do X.",
			"REQ-HOOK-002", "The system SHALL do X."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqs := parseREQsWide(tc.line)
			if len(reqs) != 1 {
				t.Fatalf("expected 1 REQ from %q, got %d (%v)", tc.line, len(reqs), reqs)
			}
			if reqs[0].ID != tc.wantID {
				t.Errorf("ID: expected %q, got %q", tc.wantID, reqs[0].ID)
			}
			if reqs[0].Text != tc.wantText {
				t.Errorf("Text: expected %q, got %q", tc.wantText, reqs[0].Text)
			}
			if reqs[0].Line != 1 {
				t.Errorf("Line: expected 1, got %d", reqs[0].Line)
			}
		})
	}
}

// TestParseREQsWide_EmDashSeparatorFormB covers Form B of the t385 census —
// the em-dash separator `**REQ-X** — text` (689 corpus lines measured, all
// silently dropped by the colon-only pattern). Asserts collection, the ID, and
// that Text captures the segment AFTER the em-dash, not the separator itself.
func TestParseREQsWide_EmDashSeparatorFormB(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantID   string
		wantText string
	}{
		{"bold-emdash", "- **REQ-BDR-001** — The system SHALL do X.",
			"REQ-BDR-001", "The system SHALL do X."},
		{"plain-emdash", "- REQ-BDR-002 — The system SHALL do X.",
			"REQ-BDR-002", "The system SHALL do X."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqs := parseREQsWide(tc.line)
			if len(reqs) != 1 {
				t.Fatalf("expected 1 REQ from %q, got %d (%v)", tc.line, len(reqs), reqs)
			}
			if reqs[0].ID != tc.wantID {
				t.Errorf("ID: expected %q, got %q", tc.wantID, reqs[0].ID)
			}
			if reqs[0].Text != tc.wantText {
				t.Errorf("Text: expected %q, got %q", tc.wantText, reqs[0].Text)
			}
		})
	}
}

// TestParseREQsWide_ParenClassifierFormC covers Form C of the t385 census —
// an EARS-style classifier in parentheses between the token and the separator
// (2,004 corpus lines measured, dropped in BOTH separator variants — including
// the paren+colon form, so the old defect was wider than "em-dash only").
// Asserts both variants are collected and that the classifier is excluded from
// the captured Text.
func TestParseREQsWide_ParenClassifierFormC(t *testing.T) {
	cases := []struct {
		name     string
		line     string
		wantID   string
		wantText string
	}{
		{"paren-emdash", "- **REQ-BDR-003** (Ubiquitous) — The system SHALL do X.",
			"REQ-BDR-003", "The system SHALL do X."},
		{"paren-colon", "- **REQ-BDR-004** (Where the system is running): The system SHALL do X.",
			"REQ-BDR-004", "The system SHALL do X."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqs := parseREQsWide(tc.line)
			if len(reqs) != 1 {
				t.Fatalf("expected 1 REQ from %q, got %d (%v)", tc.line, len(reqs), reqs)
			}
			if reqs[0].ID != tc.wantID {
				t.Errorf("ID: expected %q, got %q", tc.wantID, reqs[0].ID)
			}
			if reqs[0].Text != tc.wantText {
				t.Errorf("Text: expected %q, got %q", tc.wantText, reqs[0].Text)
			}
		})
	}
}

// TestParseREQsWide_SeparatorVariantsMultiLineCount asserts a mixed body yields
// exactly one entry per definition line with 1-based line numbers preserved —
// no silent drops across the separator axis.
func TestParseREQsWide_SeparatorVariantsMultiLineCount(t *testing.T) {
	body := strings.Join([]string{
		"## 2. Requirements",
		"",
		"- REQ-BDR-001 — em-dash body.",
		"- **REQ-BDR-002** (Ubiquitous) — classified em-dash body.",
		"- **REQ-BDR-003** (Where the system is idle): classified colon body.",
		"- **REQ-BDR-004:** bold-inside colon body.",
		"- REQ-ABC-123-456: narrow colon body.",
		"",
	}, "\n")

	reqs := parseREQsWide(body)
	if len(reqs) != 5 {
		t.Fatalf("expected 5 REQ entries, got %d (%v)", len(reqs), reqs)
	}
	wantLines := map[string]int{
		"REQ-BDR-001":     3,
		"REQ-BDR-002":     4,
		"REQ-BDR-003":     5,
		"REQ-BDR-004":     6,
		"REQ-ABC-123-456": 7,
	}
	for _, r := range reqs {
		if wantLines[r.ID] != r.Line {
			t.Errorf("REQ %s: expected line %d, got %d", r.ID, wantLines[r.ID], r.Line)
		}
	}
}

// TestParseREQsWithProvenance_SeparatorVariantsStayAdvisory asserts the
// widened-only advisory treatment survives the separator widening: entries the
// narrow reqLinePattern already collected keep Widened=false (their findings
// stay blocking), while em-dash and paren-classifier entries are marked
// Widened=true (their findings surface as advisory, never gate).
func TestParseREQsWithProvenance_SeparatorVariantsStayAdvisory(t *testing.T) {
	body := strings.Join([]string{
		"- REQ-ABC-123-456: narrow colon body.",
		"- **REQ-BDR-001** — em-dash body.",
		"- **REQ-BDR-002** (Ubiquitous) — classified em-dash body.",
		"- **REQ-BDR-003** (Where the system is idle): classified colon body.",
	}, "\n")

	reqs := parseREQsWithProvenance(body)
	if len(reqs) != 4 {
		t.Fatalf("expected 4 REQ entries, got %d (%v)", len(reqs), reqs)
	}

	wantWidened := map[string]bool{
		"REQ-ABC-123-456": false,
		"REQ-BDR-001":     true,
		"REQ-BDR-002":     true,
		"REQ-BDR-003":     true,
	}
	for _, r := range reqs {
		if r.Widened != wantWidened[r.ID] {
			t.Errorf("REQ %s: expected Widened=%v, got %v", r.ID, wantWidened[r.ID], r.Widened)
		}
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

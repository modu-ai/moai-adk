package spec

import "testing"

// Card t1120 — RED reproduction.
//
// A REQ definition split over two lines — a bold header line carrying only the
// ID (and optionally its classifier), then the statement on the next line — is
// collected by none of the four existing collectors. The bare collector needs a
// `—`/`:` separator on the ID line, and this shape has none. The corpus holds
// 565 such headers across 29 spec.md files (measured by t1104 §5 and re-measured
// for this card), so the REQ axis of lint is silent on every one of them.
//
// The fixture lines are copied verbatim from the corpus:
//   - SPEC-AGENCY-ABSORB-001 spec.md:107-108 (classifier inside the bold, trailing
//     two-space hard break)
//   - SPEC-AGENT-PROGRESS-PUSH-001 spec.md:89-90 (classifier after the bold)
const twoLineREQFixture = "## Requirements\n" +
	"\n" +
	"**REQ-ROUTE-001 (Event-Driven)**  \n" +
	"**When** 사용자가 `/moai design \"<brief>\"`를 호출하면, the 시스템 **shall** 자동으로 브리프를 분석한다.\n" +
	"\n" +
	"**REQ-APP-001** (Ubiquitous)\n" +
	"Each of the 9 retained MoAI agent definitions shall declare `SendMessage` in its frontmatter `tools:` CSV, in both the working tree and the template mirror.\n"

func TestParseREQsTwoLine_CollectsHeaderPlusStatement(t *testing.T) {
	got := parseREQsWithProvenance(twoLineREQFixture)

	want := map[string]struct {
		text string
		line int
	}{
		"REQ-ROUTE-001": {"**When** 사용자가 `/moai design \"<brief>\"`를 호출하면, the 시스템 **shall** 자동으로 브리프를 분석한다.", 3},
		"REQ-APP-001":   {"Each of the 9 retained MoAI agent definitions shall declare `SendMessage` in its frontmatter `tools:` CSV, in both the working tree and the template mirror.", 6},
	}
	if len(got) != len(want) {
		t.Fatalf("collected %d REQ entries, want %d: %+v", len(got), len(want), got)
	}
	for _, e := range got {
		w, ok := want[e.ID]
		if !ok {
			t.Errorf("unexpected REQ %s collected", e.ID)
			continue
		}
		if e.Text != w.text {
			t.Errorf("%s: Text = %q, want %q", e.ID, e.Text, w.text)
		}
		if e.Line != w.line {
			t.Errorf("%s: Line = %d, want %d (the header line, where the ID is)", e.ID, e.Line, w.line)
		}
		if !e.Widened {
			t.Errorf("%s: Widened = false, want true", e.ID)
		}
		if e.Source != REQSourceBare {
			t.Errorf("%s: Source = %v, want REQSourceBare", e.ID, e.Source)
		}
	}
}

// Negative controls. Each body is a shape that must NOT become a definition.
// The quoted and indented lines, and the column-zero prose citation, are
// verbatim corpus lines; the remaining cases pin the next-line conditions that
// separate a header from a label with no statement under it.
func TestParseREQsTwoLine_DoesNotCollectNonDefinitions(t *testing.T) {
	cases := map[string]string{
		// SPEC-BOARDLOCK-ERRNO-001 spec.md:122 — a blockquote citation.
		"quoted citation": "> REQ-BLE-001 과 짝을 이루는 반대 방향. 둘 중 하나만 잠그면 \"올바른 좁히기\"와 \"그냥 술어를 껐음\"이 구별되지 않는다.\n" +
			"The next line is prose.\n",
		// SPEC-AGENT-TEAM-RETIRE-001 spec.md:305 — an indented continuation citation.
		"indented citation": "  REQ-ATR-006.\n" +
			"The next line is prose.\n",
		// SPEC-ASTGREP-BREADTH-001 spec.md:115 — a column-zero prose citation.
		"column-zero prose citation": "REQ-A16-017 is singled out because it is the one inheritance that closes a hazard this SPEC would\n" +
			"open if left alone.\n",
		// A bare ID line with no bold is a line wrap inside prose, not a header.
		"unbolded id line":            "REQ-X-001\nThe next line is prose.\n",
		"indented header":             "  **REQ-X-001** (Ubiquitous)\nThe system shall do it.\n",
		"header then blank line":      "**REQ-X-001** (Ubiquitous)\n\nThe system shall do it.\n",
		"header then indented line":   "**REQ-X-001** (Ubiquitous)\n  The system shall do it.\n",
		"header then list item":       "**REQ-X-001** (Ubiquitous)\n- The system shall do it.\n",
		"header then star list item":  "**REQ-X-001** (Ubiquitous)\n* The system shall do it.\n",
		"header then table row":       "**REQ-X-001** (Ubiquitous)\n| a | b |\n",
		"header then heading":         "**REQ-X-001** (Ubiquitous)\n## Next section\n",
		"header then quote":           "**REQ-X-001** (Ubiquitous)\n> quoted\n",
		"header then another header":  "**REQ-X-001** (Ubiquitous)\n**REQ-X-002** (Ubiquitous)\n",
		"header at end of body":       "**REQ-X-001** (Ubiquitous)",
		"header then bare definition": "**REQ-X-001** (Ubiquitous)\n**REQ-X-002** — The system shall do it.\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			for _, e := range parseREQsWithProvenance(body) {
				if e.ID == "REQ-X-001" || e.ID == "REQ-BLE-001" || e.ID == "REQ-ATR-006" || e.ID == "REQ-A16-017" {
					t.Errorf("collected %s as a definition from %q", e.ID, body)
				}
			}
		})
	}
}

// The two-line shape must not disturb the single-line bare shape: a header
// followed by a bare definition collects the bare one only, and a single-line
// bare definition is still collected exactly once.
func TestParseREQsTwoLine_LeavesSingleLineBareUnchanged(t *testing.T) {
	got := parseREQsWithProvenance(bareREQFixture)
	if len(got) != 2 {
		t.Fatalf("single-line bare fixture collected %d entries, want 2: %+v", len(got), got)
	}
}

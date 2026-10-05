package spec

import (
	"strings"
	"testing"
)

// Card t1138 — RED reproduction (t1120 audit finding F1).
//
// The bare collector (both its single-line and its two-line form) and the list
// collector each took ONE physical line as REQEntry.Text. A statement wrapped
// over several lines was therefore judged on its first line only, and where the
// SHALL token sits on a later line the modality judge reported a defect the
// author never wrote. The heading collector already joins its whole paragraph
// (firstParagraphBelowHeading); these tests pin the same property for the
// other two.
//
// Every fixture paragraph is copied verbatim from the corpus:
//   - SPEC-TODO-LAND-AUTO-DONE-001 spec.md:155-158 (two-line bare, SHALL on line 3;
//     the source paragraph continues to line 160 and is cut after line 158 here)
//   - SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 spec.md:140-142 (single-line bare, SHALL on line 2)
//   - SPEC-AC-ANCHOR-SCOPE-001 spec.md:77-80 (list item, SHALL on line 4 — first line of the item kept, tail cut at the fixture end)
const multiLineREQFixture = "## Requirements\n" +
	"\n" +
	"**REQ-AD-014 (Ubiquitous)**\n" +
	"The workflow documentation (`.claude/skills/moai/workflows/todo.md` and its template\n" +
	"mirror) shall name the scan as the step the LEAD runs immediately after its post-push\n" +
	"remote-landing confirmation (`git fetch origin develop && git rev-parse origin/develop`),\n" +
	"\n" +
	"**REQ-CBR-004** — When a codex turn completes and its collected review text\n" +
	"contains no non-whitespace character, the turn shall return an `inconclusive`\n" +
	"review output and shall not invoke the verdict synthesizer.\n" +
	"\n" +
	"- **REQ-ACAS-001 (narrow axis)** — **When** a `spec.md` document carries at least one\n" +
	"  AC-declaration line (the discriminator-B colon form) but no heading matched by the\n" +
	"  acceptance-section vocabulary, **While** the document's non-AC prose remains outside the\n" +
	"  anchor region, the inline AC parser shall select an anchor that includes at least one of the\n"

func TestParseREQsMultiLine_JoinsWrappedStatement(t *testing.T) {
	got := parseREQsWithProvenance(multiLineREQFixture)

	want := map[string]struct {
		text string
		line int
	}{
		"REQ-AD-014": {"The workflow documentation (`.claude/skills/moai/workflows/todo.md` and its template " +
			"mirror) shall name the scan as the step the LEAD runs immediately after its post-push " +
			"remote-landing confirmation (`git fetch origin develop && git rev-parse origin/develop`),", 3},
		"REQ-CBR-004": {"When a codex turn completes and its collected review text " +
			"contains no non-whitespace character, the turn shall return an `inconclusive` " +
			"review output and shall not invoke the verdict synthesizer.", 8},
		"REQ-ACAS-001": {"**When** a `spec.md` document carries at least one " +
			"AC-declaration line (the discriminator-B colon form) but no heading matched by the " +
			"acceptance-section vocabulary, **While** the document's non-AC prose remains outside the " +
			"anchor region, the inline AC parser shall select an anchor that includes at least one of the", 12},
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
			t.Errorf("%s: Line = %d, want %d (the line carrying the ID)", e.ID, e.Line, w.line)
		}
	}
}

// The user-visible consequence: with the paragraph joined, none of the three
// corpus requirements is reported as missing its SHALL.
func TestEARSModalityRule_MultiLineStatementIsJudgedWhole(t *testing.T) {
	doc := &SPECDoc{Path: "spec.md", REQs: parseREQsWithProvenance(multiLineREQFixture)}
	for _, f := range (&EARSModalityRule{}).Check(doc, nil) {
		if f.Code == "ModalityMalformed" || f.Code == "ModalityUnjudged" {
			t.Errorf("line %d: %s — the SHALL on a continuation line was not seen: %s", f.Line, f.Code, f.Message)
		}
	}
}

// Negative controls: each continuation candidate below is a separate markdown
// block (or a separate definition), so it must NOT be appended to the Text of
// the entry above it. The want value is the entry's single-line Text.
func TestParseREQsMultiLine_StopsAtBlockBoundaries(t *testing.T) {
	cases := []struct {
		name, body, id, want string
	}{
		{"list item then blank line then paragraph",
			"- REQ-X-001: The system shall do A.\n\n  Further prose.\n", "REQ-X-001", "The system shall do A."},
		{"list item then indented sub-bullet",
			"- REQ-X-001: The system shall do A:\n  - first condition\n", "REQ-X-001", "The system shall do A:"},
		{"list item then indented numbered sub-item",
			"- REQ-X-001: The system shall do A:\n  1. first step\n", "REQ-X-001", "The system shall do A:"},
		{"list item then next list definition",
			"- REQ-X-001: The system shall do A.\n- REQ-X-002: The system shall do B.\n", "REQ-X-001", "The system shall do A."},
		{"list item then column-zero bullet",
			"- REQ-X-001: The system shall do A.\n- plain bullet\n", "REQ-X-001", "The system shall do A."},
		{"list item then indented blockquote",
			"- REQ-X-001: The system shall do A.\n  > superseded note\n", "REQ-X-001", "The system shall do A."},
		{"list item then indented fence",
			"- REQ-X-001: The system shall do A.\n  ```go\n", "REQ-X-001", "The system shall do A."},
		{"list item then heading",
			"- REQ-X-001: The system shall do A.\n### Next\n", "REQ-X-001", "The system shall do A."},
		{"bare definition then bullet",
			"**REQ-X-001** — The system shall do A:\n- first condition\n", "REQ-X-001", "The system shall do A:"},
		{"bare definition then next bare definition",
			"**REQ-X-001** — The system shall do A.\n**REQ-X-002** — The system shall do B.\n", "REQ-X-001", "The system shall do A."},
		{"bare definition then two-line header",
			"**REQ-X-001** — The system shall do A.\n**REQ-X-002 (Ubiquitous)**\nThe system shall do B.\n", "REQ-X-001", "The system shall do A."},
		{"bare definition then table",
			"**REQ-X-001** — The system shall do A.\n| a | b |\n", "REQ-X-001", "The system shall do A."},
		{"bare definition then indented line",
			"**REQ-X-001** — The system shall do A.\n  indented aside\n", "REQ-X-001", "The system shall do A."},
		{"two-line header statement then bullet",
			"**REQ-X-001 (Ubiquitous)**\nThe system shall do A:\n- first condition\n", "REQ-X-001", "The system shall do A:"},
		{"two-line header statement then fence",
			"**REQ-X-001 (Ubiquitous)**\nThe system shall do A:\n```yaml\n", "REQ-X-001", "The system shall do A:"},
		{"bare definition then tilde fence",
			"**REQ-X-001** — The system shall do A:\n~~~yaml\n", "REQ-X-001", "The system shall do A:"},
		{"bare definition then parenthesis-numbered item",
			"**REQ-X-001** — The system shall do A:\n1) first step\n", "REQ-X-001", "The system shall do A:"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var found bool
			for _, e := range parseREQsWithProvenance(c.body) {
				if e.ID != c.id {
					continue
				}
				found = true
				if e.Text != c.want {
					t.Errorf("Text = %q, want %q", e.Text, c.want)
				}
			}
			if !found {
				t.Fatalf("%s not collected from %q", c.id, c.body)
			}
		})
	}
}

// Widened-only shapes the corpus fixture above does not cover: an ID line whose
// separator carries no text, and an unindented lazy continuation of a list item.
func TestParseREQsMultiLine_EmptyFirstLineAndLazyContinuation(t *testing.T) {
	cases := []struct {
		name, body, id, want string
	}{
		{"bare definition with nothing after the separator",
			"**REQ-X-001** —\nThe system shall do A.\n", "REQ-X-001", "The system shall do A."},
		{"list item continued by an unindented line",
			"- REQ-X-001: The system\nshall do A.\n", "REQ-X-001", "The system shall do A."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var found bool
			for _, e := range parseREQsWithProvenance(c.body) {
				if e.ID != c.id {
					continue
				}
				found = true
				if e.Text != c.want {
					t.Errorf("Text = %q, want %q", e.Text, c.want)
				}
				if !e.Widened {
					t.Errorf("Widened = false, want true (the fixture must exercise the joined path)")
				}
			}
			if !found {
				t.Fatalf("%s not collected from %q", c.id, c.body)
			}
		})
	}
}

// Sync-audit F1: an entry the NARROW pattern also collects gates (Widened =
// false), so it must keep its pre-t1138 single-line Text. Joining its
// continuation would turn an advisory finding into a gating one (empty first
// line + WHEN without SHALL below → error ModalityMalformed) or introduce a
// non-advisory LegacyEARSKeyword (IF on the ID line, THEN below).
func TestParseREQsMultiLine_NarrowEntriesKeepSingleLineText(t *testing.T) {
	const body = "- REQ-FXA-001-001: When the operator acts,\n" +
		"  the system shall respond.\n" +
		"\n" +
		"- REQ-FXA-001-002: \n" +
		"  WHEN the operator acts the system responds.\n" +
		"\n" +
		"- REQ-FXA-001-003: IF the file is absent\n" +
		"  THEN the system shall report it.\n"

	want := map[string]string{
		"REQ-FXA-001-001": "When the operator acts,",
		"REQ-FXA-001-002": "",
		"REQ-FXA-001-003": "IF the file is absent",
	}
	reqs := parseREQsWithProvenance(body)
	if len(reqs) != len(want) {
		t.Fatalf("collected %d REQ entries, want %d: %+v", len(reqs), len(want), reqs)
	}
	for _, e := range reqs {
		if e.Widened {
			t.Errorf("%s: Widened = true, want false (the fixture must be narrow-collected)", e.ID)
		}
		if e.Text != want[e.ID] {
			t.Errorf("%s: Text = %q, want %q (the narrow single-line Text)", e.ID, e.Text, want[e.ID])
		}
	}

	// REQ-FXA-001-001 and -003 keep the gating ModalityMalformed their single
	// line already produced before t1138 — that finding is the preserved
	// behavior, not a regression. What joining would ADD is checked here: a
	// gating ModalityMalformed on -002 (line 4) and any LegacyEARSKeyword.
	doc := &SPECDoc{Path: "spec.md", REQs: reqs}
	for _, f := range (&EARSModalityRule{}).Check(doc, nil) {
		if f.Code == "LegacyEARSKeyword" || (f.Code == "ModalityMalformed" && f.Line == 4) {
			t.Errorf("line %d: %s (advisory=%v) — joining reached a gating entry: %s", f.Line, f.Code, f.Advisory, f.Message)
		}
	}
}

// Joining only ever ADDS text after the first line, so the prefix the modality
// judge keys on is unchanged; this pins that no leading or doubled whitespace
// is introduced at the seams.
func TestParseREQsMultiLine_SingleSpaceSeams(t *testing.T) {
	for _, e := range parseREQsWithProvenance(multiLineREQFixture) {
		if strings.Contains(e.Text, "  ") || strings.Contains(e.Text, "\n") || e.Text != strings.TrimSpace(e.Text) {
			t.Errorf("%s: Text has an unnormalised seam: %q", e.ID, e.Text)
		}
	}
}

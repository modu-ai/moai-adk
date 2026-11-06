package auditreceipt

import (
	"strings"
	"testing"
)

// Raw comparison of section E.2 (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-010,
// AC-AEC-003a/003b; card t2017, sync-audit 4 defects D16 and D13). Section E.2 is
// read from the raw text with no Markdown modelling, so a construct inside it (an
// inline-code line that only looks like a fence, an HTML <pre> block, a quoted
// heading) can no longer cut the section short and hide a later edit.

const (
	rawHead = "# P\n\n## §E.1 Plan\n\nplan one\n\n"
	rawE2   = "## §E.2 Run-phase Evidence\n\n| AC-1 | PASS |\n\n"
	rawTail = "## §E.3 Run\n\nrun one\n\n## §E.4 Sync\n\nsync_status: pending\n"
)

// rawDoc builds a progress.md whose E.2 holds `inside` followed by an AC-2 row.
func rawDoc(inside, ac2 string) string {
	return rawHead + rawE2 + inside + "| AC-2 | " + ac2 + " |\n\n" + rawTail
}

// Shapes placed inside E.2, above the AC-2 row, that made the fence parser end E.2
// at a "## " line inside them.
var rawInsideShapes = []struct{ name, inside string }{
	{"D16: inline-code line, then a bare fence holding a heading line", "```foo``` is inline code, not a fence\n\n```\n## quoted heading line\n```\n\n"},
	{"D16: two backtick info-string lines, then a bare fence holding a heading line", "```a`b\n```c`d\n\n```\n## quoted heading line\n```\n\n"},
	{"D13: HTML <pre> block holding a heading line", "<pre>\n## quoted heading line\n</pre>\n\n"},
	{"D13: HTML comment holding a heading line", "<!--\n## quoted heading line\n-->\n\n"},
	{"control: plain fence holding a heading line", "```text\n## quoted heading line\n```\n\n"},
	{"control: nothing", ""},
}

// An edit of an AC row below the shape must read as a change of E.2.
func TestE2DocsUnchanged_RawRegion_InsideShapesNeverHideAnEdit(t *testing.T) {
	for _, s := range rawInsideShapes {
		t.Run(s.name, func(t *testing.T) {
			oldDoc, newDoc := rawDoc(s.inside, "PASS"), rawDoc(s.inside, "FAIL")
			if e2DocsUnchanged(oldDoc, newDoc) {
				t.Error("an edit of the AC-2 row inside E.2 compared unchanged; the receipt would stay fresh")
			}
		})
	}
}

// Controls on the same shapes: identical documents, and an edit confined to E.3, stay
// unchanged; an edit above the shape (AC-1) is a change.
func TestE2DocsUnchanged_RawRegion_InsideShapeControls(t *testing.T) {
	for _, s := range rawInsideShapes {
		t.Run(s.name, func(t *testing.T) {
			doc := rawDoc(s.inside, "PASS")
			if !e2DocsUnchanged(doc, doc) {
				t.Error("identical documents compared changed")
			}
			if !e2DocsUnchanged(doc, strings.Replace(doc, "run one", "run two, rewritten wholesale", 1)) {
				t.Error("an edit confined to E.3 compared changed")
			}
			if !e2DocsUnchanged(doc, strings.Replace(doc, "sync_status: pending", "sync_status: complete", 1)) {
				t.Error("an edit confined to E.4 compared changed")
			}
			if e2DocsUnchanged(doc, strings.Replace(doc, "| AC-1 | PASS |", "| AC-1 | FAIL |", 1)) {
				t.Error("an edit of the AC-1 row inside E.2 compared unchanged")
			}
		})
	}
}

// Every ambiguous shape reads as changed, whether it appears on one side or both, and
// whether or not anything below it was edited.
func TestE2DocsUnchanged_RawRegion_AmbiguityFailsClosed(t *testing.T) {
	rows := []struct{ name, doc string }{
		{"two E.2 heading lines", rawHead + rawE2 + "| AC-2 | PASS |\n\n## §E.2 Again\n\nrow\n\n" + rawTail},
		{"near-miss heading ### §E.2", rawHead + "### §E.2 Run-phase Evidence\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"indented heading", rawHead + "  ## §E.2 Run-phase Evidence\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"tab between ## and §E.2", rawHead + "##\t§E.2 Run-phase Evidence\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"bare heading ending in a carriage return", rawHead + "## §E.2\r\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"sub-numbered heading ## §E.2.1", rawHead + "## §E.2.1 Run-phase Evidence\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"setext heading", rawHead + "§E.2 Run-phase Evidence\n=====\n\n| AC-2 | PASS |\n\n" + rawTail},
		{"boundary line above the E.2 heading", rawHead + "## §E.3 Run\n\nrun one\n\n" + rawE2 + "| AC-2 | PASS |\n\n"},
		{"E.4 above E.3 below the heading", rawHead + rawE2 + "| AC-2 | PASS |\n\n## §E.4 Sync\n\ns\n\n## §E.3 Run\n\nr\n"},
		{"quoted E.3 line inside E.2 and the real E.3", rawDoc("```text\n## §E.3 quoted\n```\n\n", "PASS")},
		{"quoted E.4 line inside E.2 and the real E.3, E.4", rawDoc("```text\n## §E.4 quoted\n```\n\n", "PASS")},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if _, _, ok := e2RawRegion(r.doc); ok {
				t.Error("an ambiguous document read as unambiguous")
			}
			good := rawDoc("", "PASS")
			if e2DocsUnchanged(r.doc, r.doc) {
				t.Error("an ambiguous document compared unchanged against itself")
			}
			if e2DocsUnchanged(r.doc, strings.Replace(r.doc, "| AC-2 | PASS |", "| AC-2 | FAIL |", 1)) {
				t.Error("an edit of a row inside an ambiguous document compared unchanged")
			}
			if e2DocsUnchanged(good, r.doc) || e2DocsUnchanged(r.doc, good) {
				t.Error("an ambiguous version against a well-formed one compared unchanged")
			}
		})
	}
}

// The quoted-heading corner with a row edit below the quote: the real E.3 repeats the
// quoted number, so the document is ambiguous and the edit cannot be hidden.
func TestE2DocsUnchanged_RawRegion_QuotedBoundaryCannotHideALaterEdit(t *testing.T) {
	inside := "```text\n## §E.3 Run-phase Audit-Ready Signal\n```\n\n"
	oldDoc, newDoc := rawDoc(inside, "PASS"), rawDoc(inside, "FAIL")
	if e2DocsUnchanged(oldDoc, newDoc) {
		t.Error("an AC-2 edit below a quoted E.3 line compared unchanged")
	}
}

// E.2 present on one side only, or moved, is a change; an absent E.2 on both sides is not.
func TestE2DocsUnchanged_RawRegion_PresenceAndPosition(t *testing.T) {
	withE2, withoutE2 := rawDoc("", "PASS"), rawHead+rawTail
	if e2DocsUnchanged(withoutE2, withE2) || e2DocsUnchanged(withE2, withoutE2) {
		t.Error("E.2 present on one side only compared unchanged")
	}
	if !e2DocsUnchanged(withoutE2, strings.Replace(withoutE2, "run one", "run two", 1)) {
		t.Error("two documents without E.2 and a different E.3 compared changed")
	}
	moved := strings.Replace(withE2, rawE2, "## §E.0 New\n\nz\n\n"+rawE2, 1)
	if e2DocsUnchanged(withE2, moved) {
		t.Error("a second-level heading inserted above E.2 compared unchanged")
	}
	renamed := strings.Replace(withE2, "## §E.2 Run-phase Evidence", "## §E.2 Run-phase Results", 1)
	if e2DocsUnchanged(withE2, renamed) {
		t.Error("a renamed E.2 heading line compared unchanged")
	}
	crlf := strings.ReplaceAll(withE2, "\n", "\r\n")
	if e2DocsUnchanged(withE2, crlf) {
		t.Error("an LF document against its CRLF conversion compared unchanged")
	}
}

package auditreceipt

import (
	"slices"
	"strings"
	"testing"
)

// Fail-closed comparison of section E.2 (SPEC-AUDIT-ENGINE-CONVERGENCE-001 REQ-AEC-010,
// AC-AEC-003b; card t2017, sync-audit 2 defect D10). When the fence parser misreads a
// document and misses a real "## §E.2" heading on both sides, both section lists are
// empty and equal, so an edit under that heading kept the receipt fresh. e2DocsUnchanged
// now also compares the number of raw heading lines against the number of sections the
// parser found, and reports "changed" whenever they disagree on either side.

const (
	d10Heading = "## §E.2 Run-phase Evidence\n"
	d10OldBody = "AC-1 PASS"
	d10NewBody = "AC-1 FAIL (rewritten)"
)

// d10Shape is a document prefix placed before the E.2 heading. parsedOK marks the
// prefixes the current section parser reads correctly (the heading is found), the
// others are misreads the raw-count check must catch.
type d10Shape struct {
	name     string
	prefix   string
	parsedOK bool
}

var d10Shapes = []d10Shape{
	{"control: no fence", "# P\n\n## §E.1 Plan\n\nplain\n\n", true},
	{"control: plain fence", "# P\n\n## §E.1 Plan\n```text\nplain\n```\n\n", true},
	{"D1: four-backtick fence holding three", "# P\n\n## §E.1 Plan\n````text\nA fence opens with a line like:\n```\n````\n\n", true},
	{"D1: four-tilde fence holding three", "# P\n\n## §E.1 Plan\n~~~~\n~~~\n~~~~\n\n", true},
	{"D1: tilde fence holding backticks", "# P\n\n## §E.1 Plan\n~~~\n```\n~~~\n\n", true},
	{"inline-code line x1", "# P\n\n## §E.1 Plan\n```foo``` is inline code, not a fence\n\n", false},
	{"inline-code lines x2", "# P\n\n## §E.1 Plan\n```a``` one\n```b``` two\n\n", false},
	{"info-string-with-backtick lines x2", "# P\n\n## §E.1 Plan\n```a`b\n```c`d\n\n", false},
	{"list fence deep-open shallow-close", "# P\n\n## §E.1 Plan\n- item\n\n    ```\n    code\n  ```\n\n", false},
	{"HTML comment holding a backtick line", "# P\n\n## §E.1 Plan\n<!--\n```\n-->\n\n", false},
	{"<pre> holding a tilde line", "# P\n\n## §E.1 Plan\n<pre>\n~~~\n</pre>\n\n", false},
	{"4-backtick open, 3-backtick close, unterminated", "# P\n\n## §E.1 Plan\n````\ncode\n```\n\n", false},
	{"info-string line as closer, unterminated", "# P\n\n## §E.1 Plan\n```text\ncode\n```text\n\n", false},
	{"4-space-indented line then unterminated fence", "# P\n\n## §E.1 Plan\n\n    ```\ncode\n```\n\n", false},
	{"plain unterminated fence", "# P\n\n## §E.1 Plan\n```\ncode\n\n", false},
}

// Every shape: an edit under the real E.2 heading must read as a change.
func TestE2DocsUnchanged_E2BodyEditIsAlwaysAChange(t *testing.T) {
	for _, s := range d10Shapes {
		t.Run(s.name, func(t *testing.T) {
			oldDoc := s.prefix + d10Heading + d10OldBody
			newDoc := s.prefix + d10Heading + d10NewBody
			if e2DocsUnchanged(oldDoc, newDoc) {
				_, oldHas, oldOK := e2RawRegion(oldDoc)
				_, newHas, newOK := e2RawRegion(newDoc)
				t.Errorf("an edit under the E.2 heading compared unchanged (present old=%v new=%v, ok old=%v new=%v); the receipt would stay fresh",
					oldHas, newHas, oldOK, newOK)
			}
		})
	}
}

// Positive controls: nothing differs, or only text outside E.2 differs.
func TestE2DocsUnchanged_PositiveControls(t *testing.T) {
	for _, s := range d10Shapes {
		doc := s.prefix + d10Heading + d10OldBody
		got := e2DocsUnchanged(doc, doc)
		if s.parsedOK {
			// Normal Markdown, parsed correctly: identical documents stay unchanged.
			if !got {
				t.Errorf("%s: identical documents compared changed", s.name)
			}
			continue
		}
		// A misread shape fails closed even when identical; recorded, not pinned.
		t.Logf("identical documents, misread shape %q: unchanged=%v", s.name, got)
	}

	t.Run("a change outside E.2 on a well-formed document stays unchanged", func(t *testing.T) {
		const base = "# P\n\n## §E.1 Plan\n\nplan one\n\n## §E.2 Run-phase Evidence\n\nAC-1 PASS\n\n## §E.3 Run\n\nrun\n\n## §E.4 Sync\n\nsync_status: pending\n"
		for _, edit := range []struct{ old, repl string }{
			{"plan one", "plan two"},
			{"sync_status: pending", "sync_status: complete"},
		} {
			if !e2DocsUnchanged(base, strings.Replace(base, edit.old, edit.repl, 1)) {
				t.Errorf("an edit of %q outside E.2 compared changed", edit.old)
			}
		}
		if e2DocsUnchanged(base, strings.Replace(base, "AC-1 PASS", "AC-1 FAIL", 1)) {
			t.Error("control: an edit inside E.2 compared unchanged")
		}
	})

	t.Run("a change outside E.2 beside a well-formed plain fence stays unchanged", func(t *testing.T) {
		const base = "# P\n\n## §E.1 Plan\n```text\nplain\n```\n\nplan one\n\n## §E.2 Run-phase Evidence\n\nAC-1 PASS\n\n## §E.3 Run\n\nrun\n"
		if !e2DocsUnchanged(base, strings.Replace(base, "plan one", "plan two", 1)) {
			t.Error("an edit outside E.2 compared changed")
		}
	})

	t.Run("a file added or deleted without E.2 is unchanged", func(t *testing.T) {
		if !e2DocsUnchanged("", "# P\n\n## §E.1 Plan\n\nx\n") {
			t.Error("an empty document against one without E.2 compared changed")
		}
		if e2DocsUnchanged("", d10Heading+d10OldBody) {
			t.Error("a file gaining E.2 compared unchanged")
		}
	})

	t.Run("E.2 quoted inside a fence only, identical documents", func(t *testing.T) {
		doc := "# P\n\n## §E.1 Plan\n```text\n## §E.2 Run-phase Evidence\nquoted\n```\n\n## §E.3 Run\n"
		_, present, ok := e2RawRegion(doc)
		t.Logf("present=%v ok=%v unchanged=%v", present, ok, e2DocsUnchanged(doc, doc))
	})
}

// e2Section and fenceRun are test-only copies of the fence-tracking parser removed from
// store.go by the D16/D13 repair (card t2017). They exist so the comparison table below
// can keep running the earlier parsers against the raw-region comparison; nothing in
// production uses them.
type e2Section struct {
	ordinal int
	heading string
	body    string
}

func fenceRun(line string) (ch byte, n int, rest string, ok bool) {
	t := strings.TrimLeft(line, " ")
	if len(line)-len(t) > 3 || t == "" || (t[0] != '`' && t[0] != '~') {
		return 0, 0, "", false
	}
	ch = t[0]
	for n < len(t) && t[n] == ch {
		n++
	}
	if n < 3 {
		return 0, 0, "", false
	}
	return ch, n, t[n:], true
}

// legacyE2Sections1008 is a verbatim copy of e2Sections at commit 1008b915b2 (the
// toggle-based fence tracking before the D1 repair). Test-only, for the comparison table.
func legacyE2Sections1008(doc string) []e2Section {
	var out []e2Section
	ordinal, cur, fenced := 0, -1, false
	for _, line := range strings.Split(doc, "\n") {
		if !fenced && strings.HasPrefix(line, "## ") {
			ordinal++
			cur = -1
			if e2HeadingPattern.MatchString(line) {
				out = append(out, e2Section{ordinal: ordinal, heading: line})
				cur = len(out) - 1
			}
			continue
		}
		if t := strings.TrimLeft(line, " "); strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			fenced = !fenced
		}
		if cur >= 0 {
			out[cur].body += line + "\n"
		}
	}
	return out
}

// e2SectionsD9665 is a verbatim copy of e2Sections at commit d9665c52c5 (CommonMark-style
// length- and character-aware tracking). It calls the package's fenceRun, which is
// unchanged between d9665c52c5 and the fix. Test-only, for the comparison table.
func e2SectionsD9665(doc string) []e2Section {
	var out []e2Section
	ordinal, cur := 0, -1
	var fenceCh byte
	fenceLen := 0 // 0 outside a fence, else the length of the run that opened it
	for _, line := range strings.Split(doc, "\n") {
		if fenceLen == 0 && strings.HasPrefix(line, "## ") {
			ordinal++
			cur = -1
			if e2HeadingPattern.MatchString(line) {
				out = append(out, e2Section{ordinal: ordinal, heading: line})
				cur = len(out) - 1
			}
			continue
		}
		if ch, n, rest, ok := fenceRun(line); ok {
			if fenceLen == 0 {
				fenceCh, fenceLen = ch, n
			} else if ch == fenceCh && n >= fenceLen && strings.TrimSpace(rest) == "" {
				fenceLen = 0
			}
		}
		if cur >= 0 {
			out[cur].body += line + "\n"
		}
	}
	return out
}

// Prints one row per shape: the unchanged verdict of the 1008b915b2 parser, the
// d9665c52c5 parser, and the raw-region comparison, for an edit under the real E.2 heading.
// Run with -v to read it. A row reading unchanged=true is a receipt that stays fresh.
func TestE2DocsUnchanged_ComparisonTable(t *testing.T) {
	t.Logf("%-50s | %-10s | %-10s | %-10s", "shape (edit under real E.2)", "1008b915b2", "d9665c52c5", "fixed")
	for _, s := range d10Shapes {
		oldDoc := s.prefix + d10Heading + d10OldBody
		newDoc := s.prefix + d10Heading + d10NewBody
		legacy := slices.Equal(legacyE2Sections1008(oldDoc), legacyE2Sections1008(newDoc))
		d9665 := slices.Equal(e2SectionsD9665(oldDoc), e2SectionsD9665(newDoc))
		fixed := e2DocsUnchanged(oldDoc, newDoc)
		t.Logf("%-50s | unch=%-5v | unch=%-5v | unch=%-5v", s.name, legacy, d9665, fixed)
	}
}

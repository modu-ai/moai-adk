package spec

import (
	"regexp"
	"strings"
)

// Bare-form REQ collection — card t1104.
//
// THE DEFECT. The three existing definition collectors are each anchored to a
// markdown marker: a list bullet (reqLineWidePattern), a table row
// (reqTableRowPattern), or a level-3 heading (reqHeadingWidePattern). A
// requirement written as a plain paragraph opening with its own ID —
//
//	**REQ-PRB-001** (Ubiquitous) — The system SHALL …
//
// — matches none of them, so it never entered doc.REQs at all. As with the
// heading axis, it did not fail a rule; it was never visited by one, and that
// silence was indistinguishable from a pass. The lead measured the asymmetry
// through the binary: the line above produced zero REQ-coverage findings, and
// prefixing it with `- ` produced one. Two lanes (cards t1020 and t1057)
// reproduced it independently before it was carded.
//
// THE ANCHOR IS COLUMN ZERO, and that is the whole safety argument. The other
// three collectors are kept honest by their marker — the marker is what
// separates a definition from a prose mention of the same ID. A bare line has
// no marker, so the only remaining discriminator is POSITION: a definition
// OPENS its line with the ID, a citation does not.
//
// [HARD] THERE IS NO `\s*` AFTER THE `^`, AND NONE MAY BE ADDED. An earlier
// revision of this pattern read `^\**\s*(REQ-…)`, which looks equivalent and is
// not: with no bold marker present, `\**` matches empty and the `\s*` then eats
// the indentation, so every indented line became eligible. Two corpus shapes
// were mis-collected as definitions before it was removed, both wrapped
// continuation lines of a markdown list item that merely CITE a REQ ID. Each
// begins with two spaces, written below as <SP><SP> because gofmt normalises
// leading whitespace inside a comment block and the indentation IS the point:
//
//	<SP><SP>REQ-A16-020: it is not a mirror source.
//	        — SPEC-ASTGREP-LANG16-001 spec.md:503
//	<SP><SP>REQ-BH-005: no card dropped, edited, closed, reordered, unpicked, or picked.
//	        — SPEC-BACKLOG-HYGIENE-001 spec.md:319
//
// The same `\s*` also let `\**` consume a `*` LIST BULLET and the space behind
// it, so `* REQ-X — …` reached this collector as well as the list one. Dropping
// `\s*` closes both holes with the same character, and both are measured by
// TestParseREQsBare_DoesNotCollectIndentedContinuationCitation and
// TestParseREQsBare_DoesNotReachListBullets.
//
// The unindented anchor is therefore load-bearing, not an oversight.
//
// The ID shape, the optional bold markers, the optional parenthesised
// classifier and the `—`/`:` separator are deliberately IDENTICAL to
// reqLineWidePattern, so no second ID lexicon is derived and this collector
// sees what an otherwise-unchanged collector would see.
//
// NO DOUBLE COLLECTION. The three existing shapes each begin their line with a
// character this pattern cannot match — `-`/`*` for a list item, `|` for a
// table row, `#` for a heading — so no line is collected twice. The merge is
// by line number regardless, but the disjointness is a property of the
// anchors rather than of the merge.
//
// SEVERITY AND PROVENANCE STAY SEPARATE, exactly as the table and heading
// sources left them. Every bare entry carries Widened = true — the narrow
// reqLinePattern requires a `-` followed by whitespace before the ID, which a
// bare line by construction does not have — and is demoted through the single
// existing axis (reqFindingSeverity). REQSourceBare records the shape for
// attribution ONLY and never reaches a severity decision.
var reqBareWidePattern = regexp.MustCompile(`^\**(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*(?:\([^)]*\)\s*\**\s*)?(?:—|:)\s*(.*)$`)

// Two-line bare form — card t1120.
//
// The separator requirement above left one corpus shape uncollected: a header
// line carrying ONLY the ID (and optionally its classifier), with the statement
// on the next line —
//
//	**REQ-ROUTE-001 (Event-Driven)**
//	**When** the user calls …, the system **shall** …
//
// — 565 headers across 29 spec.md files (t1104 §5, re-measured for t1120). Every
// one is bold, and every one is followed by a plain, unindented statement line.
//
// reqBareHeaderPattern is deliberately separate from reqBareWidePattern so that
// the single-line pattern keeps refusing separator-less lines
// (TestParseREQsBare_RequiresASeparator). It is narrower than the single-line
// pattern in one way: the opening `**` is REQUIRED. A lone unbolded ID line at
// column zero is what a wrapped prose paragraph produces, and the corpus holds
// no unbolded header, so the bold costs nothing and closes that hole.
//
// The next line is the statement only when it is a plain paragraph line:
// non-empty, unindented, not a list item (bulleted or numbered), table row,
// heading, blockquote, code fence, thematic break or HTML line, and not itself
// a REQ header or bare definition. Anything else means the header labels
// nothing collectable, and it is skipped rather than guessed at. None of the
// 565 corpus statement lines starts with any excluded shape. A statement
// that opens with bold (`**When** …`) is NOT a `* ` bullet and is accepted.
//
// A statement wrapped over several lines is joined into one Text by
// joinStatementContinuation (card t1138), as for the single-line form.
var reqBareHeaderPattern = regexp.MustCompile(`^\*\*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)(?:\*\*)?\s*(?:\([^)]*\))?\s*(?:\*\*)?\s*$`)

// reqOrderedListPattern matches a numbered list item ("1. ", "12. ", "1) ").
var reqOrderedListPattern = regexp.MustCompile(`^\d+[.)] `)

// isTwoLineStatement reports whether next can be the statement line under a
// two-line bare header.
func isTwoLineStatement(next string) bool {
	if strings.TrimSpace(next) == "" || next[0] == ' ' || next[0] == '\t' {
		return false
	}
	for _, p := range []string{"- ", "* ", "+ ", "|", "#", ">", "```", "~~~", "---", "<"} {
		if strings.HasPrefix(next, p) {
			return false
		}
	}
	if reqOrderedListPattern.MatchString(next) {
		return false
	}
	return !reqBareHeaderPattern.MatchString(next) && !reqBareWidePattern.MatchString(next)
}

// joinStatementContinuation returns first followed by every continuation line of
// the same paragraph, starting at lines[next], joined with single spaces — card
// t1138 (t1120 audit finding F1).
//
// Taking one physical line as Text truncated every wrapped statement, and where
// the SHALL token sat on a later line the modality judge reported a defect the
// author never wrote: 8 advisory false positives among the 565 two-line headers
// alone. The heading collector already joins its paragraph
// (firstParagraphBelowHeading); this gives the bare and list collectors the same
// property.
//
// A continuation line is a line isTwoLineStatement accepts: a plain paragraph
// line, never blank and never a list item, table row, heading, blockquote, code
// fence, thematic break, HTML line or another REQ definition. The first line that
// fails ends the paragraph. When indented is true (a list item) the line is
// judged after its indentation is stripped, so an indented wrapped line
// continues the item while an indented sub-bullet or fence still ends it; an
// unindented plain line continues it as a markdown lazy continuation. When
// indented is false (a bare definition) an indented line ends the paragraph.
//
// When first is non-empty, joining only appends after it, so the leading text
// the modality judge keys its prefix on is unchanged. When first is empty (an ID
// line whose separator carries no text), the first continuation line becomes the
// prefix instead — that is the statement the author wrote, and it is the point
// of the join. Entries the narrow pattern also collects are exempt from joining
// altogether (parseREQsWithProvenance), because they gate.
func joinStatementContinuation(first string, lines []string, next int, indented bool) string {
	var parts []string
	if t := strings.TrimSpace(first); t != "" {
		parts = append(parts, t)
	}
	for ; next < len(lines); next++ {
		line := lines[next]
		if indented {
			line = strings.TrimSpace(line)
		}
		if !isTwoLineStatement(line) {
			break
		}
		parts = append(parts, strings.TrimSpace(line))
	}
	return strings.Join(parts, " ")
}

// parseREQsBareForm returns one REQEntry per unindented, marker-less definition,
// in document order, with Line as a 1-based index into body. A single-line
// definition carries its statement after the separator; a two-line definition
// carries it on the line after the header, and Line points at the header.
func parseREQsBareForm(body string) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		id, text := "", ""
		if matches := reqBareWidePattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, text = matches[1], joinStatementContinuation(matches[2], lines, i+1, false)
		} else if m := reqBareHeaderPattern.FindStringSubmatch(line); m != nil && i+1 < len(lines) && isTwoLineStatement(lines[i+1]) {
			id, text = m[1], joinStatementContinuation(lines[i+1], lines, i+2, false)
		} else {
			continue
		}
		reqs = append(reqs, REQEntry{
			ID:      id,
			Text:    text,
			Line:    i + 1,
			Widened: true,
			Source:  REQSourceBare,
		})
	}
	return reqs
}

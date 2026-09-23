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
// non-empty, unindented, not a list item, table row, heading or blockquote, and
// not itself a REQ header or bare definition. Anything else means the header
// labels nothing collectable, and it is skipped rather than guessed at.
var reqBareHeaderPattern = regexp.MustCompile(`^\*\*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)(?:\*\*)?\s*(?:\([^)]*\))?\s*(?:\*\*)?\s*$`)

// isTwoLineStatement reports whether next can be the statement line under a
// two-line bare header.
func isTwoLineStatement(next string) bool {
	if strings.TrimSpace(next) == "" || next[0] == ' ' || next[0] == '\t' {
		return false
	}
	for _, p := range []string{"- ", "* ", "+ ", "|", "#", ">"} {
		if strings.HasPrefix(next, p) {
			return false
		}
	}
	return !reqBareHeaderPattern.MatchString(next) && !reqBareWidePattern.MatchString(next)
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
			id, text = matches[1], matches[2]
		} else if m := reqBareHeaderPattern.FindStringSubmatch(line); m != nil && i+1 < len(lines) && isTwoLineStatement(lines[i+1]) {
			id, text = m[1], lines[i+1]
		} else {
			continue
		}
		reqs = append(reqs, REQEntry{
			ID:      id,
			Text:    strings.TrimSpace(text),
			Line:    i + 1,
			Widened: true,
			Source:  REQSourceBare,
		})
	}
	return reqs
}

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
// OPENS its line with the ID, a citation does not. Allowing leading whitespace
// would give that discriminator away, because an indented continuation line or
// a line inside an indented block would then read as a definition. The
// unindented anchor is therefore load-bearing, not an oversight.
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
var reqBareWidePattern = regexp.MustCompile(`^\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*(?:\([^)]*\)\s*\**\s*)?(?:—|:)\s*(.*)$`)

// parseREQsBareForm returns one REQEntry per unindented, marker-less line that
// reqBareWidePattern admits as a definition, in document order, with Line as a
// 1-based index into body.
func parseREQsBareForm(body string) []REQEntry {
	var reqs []REQEntry
	for i, line := range strings.Split(body, "\n") {
		matches := reqBareWidePattern.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}
		reqs = append(reqs, REQEntry{
			ID:      matches[1],
			Text:    strings.TrimSpace(matches[2]),
			Line:    i + 1,
			Widened: true,
			Source:  REQSourceBare,
		})
	}
	return reqs
}

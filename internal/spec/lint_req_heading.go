package spec

import (
	"regexp"
	"strings"
)

// Heading-form REQ collection — SPEC-HEADING-REQ-COLLECT-001 (card t894).
//
// THE DEFECT. The two existing definition collectors are anchored to a markdown
// list bullet (reqLineWidePattern) and to a table row (reqTableRowPattern). A
// requirement written as a level-3 heading —
//
//	### REQ-ADV-001 — Event-driven (When) — advisor rung trigger and re-seed
//
// — matches neither anchor, so it never entered doc.REQs at all. It did not fail
// a rule; it was never visited by one. The six codes that consume doc.REQs
// (ModalityMalformed, ModalityUnjudged, LegacyEARSKeyword, InvalidREQID,
// DuplicateREQID, CoverageIncomplete) were all silent on such a document, and
// that silence was indistinguishable from a pass.
//
// THE DECISION COLLECTION FORCES — WHAT BECOMES REQEntry.Text. Modality judgment
// is defined over a requirement STATEMENT, and a heading line carries a section
// TITLE. Two variants were measured with the build-tagged probe
// (heading_req_volume_probe_test.go) against tree dcfad4805:
//
//	A — the heading's own trailing text:  ModalityUnjudged fires on 1023 of 1031
//	    newly collected entries (ratio 0.992), ModalityMalformed 8.
//	B — the first body paragraph below:   ModalityUnjudged 145 of 1031 (ratio
//	    0.141), ModalityMalformed 36.
//
// VARIANT B SHIPS, on that measurement rather than on preference. A signal that
// fires on 99.2% of its population is a constant, not a signal: under A the code
// reports that the collector fed the judge a title, and nothing about the
// requirement. B lowers the noise AND raises ModalityMalformed 8 → 36 — it finds
// MORE real defects while producing LESS noise, which is precisely what
// distinguishes it from suppression (suppression lowers both numbers).
//
// Note for the M2 reconciliation: the probe's extractor takes only the FIRST
// non-empty line, while this collector takes the whole run of consecutive
// non-empty lines per plan.md §B. The two therefore differ on multi-line
// paragraphs, and the per-code deltas are expected to diverge slightly from the
// figures above. That divergence is a named cause, not an unexplained one.
//
// WHY THE SEARCH IS BOUNDED BY THE NEXT HEADING OF ANY LEVEL. An unbounded (or
// h3-only) search walks past an intervening `##` and attributes a NEIGHBOURING
// section's sentence to this requirement. That is a manufactured judgment —
// strictly worse than the silence being repaired, because a wrong verdict reads
// exactly like a right one. REQ-HRC-006.
//
// WHY THE FALLBACK IS THE HEADING TEXT AND NEVER THE EMPTY STRING. An entry with
// empty Text is judged ModalityUnjudged for a reason that has nothing to do with
// the author's writing — re-manufacturing the constant signal variant A
// produces, one layer down. REQ-HRC-005.
//
// SEVERITY AND PROVENANCE STAY SEPARATE, exactly as the table source left them.
// Every heading entry carries Widened = true and is demoted through the single
// existing axis (reqFindingSeverity); REQSourceHeading records the shape for
// attribution ONLY and never reaches a severity decision. See REQEntry.Source.

// reqHeadingWidePattern mirrors reqLineWidePattern, differing ONLY in the
// anchor: the list bullet (`^\s*[-*]\s+`) is replaced by a level-3 heading
// (`^###\s+`). The ID shape, the optional bold markers, the optional
// parenthesised classifier and the `—`/`:` separator are deliberately identical,
// so no second ID lexicon is derived (REQ-HRC-002) and the collector sees what
// an otherwise-unchanged collector would see.
//
// `###` only (spec.md §D). `^###\s+` cannot match `####` — the fourth `#` is not
// whitespace — and `##` is likewise unreachable. Neither level has a measured
// population, and widening an anchor for an unmeasured one is the shape of
// change spec.md §A.4 forbids.
var reqHeadingWidePattern = regexp.MustCompile(`^###\s+\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*(?:\([^)]*\)\s*\**\s*)?(?:—|:)\s*(.*)$`)

// isMarkdownHeadingLine reports whether a line opens a markdown ATX heading of
// any level. It is the lower bound of the paragraph search: ANY level, not just
// the level that produced the entry.
func isMarkdownHeadingLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// firstParagraphBelowHeading returns the first paragraph strictly below the
// heading at idx, where "paragraph" is the run of consecutive non-empty lines
// beginning at the first non-empty line inside the bounds (plan.md §B). Lines
// are joined with a single space so the modality judge reads one statement
// rather than a fragment.
//
// The search is bounded above by its own heading (it starts at idx+1) and below
// by the next markdown heading of any level or by EOF. It returns "" when no
// such paragraph exists — the caller substitutes the heading's own trailing
// text, which is why "" here never becomes an empty REQEntry.Text.
func firstParagraphBelowHeading(lines []string, idx int) string {
	j := idx + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j >= len(lines) || isMarkdownHeadingLine(lines[j]) {
		return ""
	}

	var paragraph []string
	for ; j < len(lines); j++ {
		trimmed := strings.TrimSpace(lines[j])
		if trimmed == "" || isMarkdownHeadingLine(lines[j]) {
			break
		}
		paragraph = append(paragraph, trimmed)
	}
	return strings.Join(paragraph, " ")
}

// parseREQsHeadingForm returns one REQEntry per level-3 heading that
// reqHeadingWidePattern admits as a definition, in document order, with Line as
// a 1-based index into body.
//
// Every returned entry carries Widened = true: no heading was EVER reachable by
// the narrow pattern, so a finding against one is a newly-surfaced corpus fact
// rather than a regression an author introduced. It reports without gating, on
// exactly the terms t385 established for widened list entries and t518 for table
// entries.
func parseREQsHeadingForm(body string) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		matches := reqHeadingWidePattern.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		// Variant B: the statement below the heading, with the heading's own
		// trailing text as the fallback — never the empty string.
		text := firstParagraphBelowHeading(lines, i)
		if text == "" {
			text = strings.TrimSpace(matches[2])
		}

		reqs = append(reqs, REQEntry{
			ID:      matches[1],
			Text:    text,
			Line:    i + 1,
			Widened: true,
			Source:  REQSourceHeading,
		})
	}
	return reqs
}

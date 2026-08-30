package spec

import (
	"regexp"
	"strings"
)

// reqLineWidePattern is the widened REQ definition-line pattern.
//
// The live parser (parseREQs / reqLinePattern) accepts only the four-segment
// shape REQ-<2..5 uppercase letters>-<3 digits>-<3 digits>. Corpus measurement
// showed that shape covers 15 of 702 spec.md files; six further shapes appear in
// practice and are silently dropped:
//
//	REQ-HOOK-001          three-segment
//	REQ-WF001-001         digits inside the domain segment
//	REQ-VNRN-RT-001-001   five-segment
//	REQ-HRN-FND-001       two-part domain
//	REQ-TUX1-001          domain ending in a digit
//	REQ-WC01-001          alphanumeric domain
//
// The pattern below accepts all six plus the current narrow shape.
//
// Two deliberate differences from reqLinePattern:
//
//  1. It is ANCHORED to a markdown list item (`^\s*[-*]\s+`), where
//     reqLinePattern is unanchored and therefore also matches a mid-prose
//     hyphen. Anchoring is what keeps a prose mention from being read as a
//     definition; the containment consequence is measured by the corpus harness
//     rather than assumed.
//  2. It tolerates the bold marker (`**REQ-...**:`) commonly used in the corpus.
//
// M3 wired this into parseSPECDoc via parseREQsWithProvenance. The corpus
// behavior change it carries is absorbed by the widened-only advisory treatment
// documented on REQEntry.Widened and at each affected emission site.
var reqLineWidePattern = regexp.MustCompile(`^\s*[-*]\s+\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*:\s*(.*)$`)

// parseREQsWide mirrors parseREQs but uses reqLineWidePattern. It returns one
// REQEntry per recognized REQ definition line, in document order, with Line as a
// 1-based index into body.
func parseREQsWide(body string) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		matches := reqLineWidePattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			reqs = append(reqs, REQEntry{
				ID:   matches[1],
				Text: strings.TrimSpace(matches[2]),
				Line: i + 1,
			})
		}
	}
	return reqs
}

// parseREQsWithProvenance is the live-path collector. It returns the WIDE REQ
// set with each entry marked according to whether the NARROW reqLinePattern
// would also have collected it at the same line.
//
// The provenance is what lets the widening land without reddening the corpus.
// doc.REQs feeds four error-severity findings — ModalityMalformed,
// InvalidREQID, DuplicateREQID, CoverageIncomplete — and none of those codes is
// in eraDemotableCodes, so widening the collector turns them on across a corpus
// that was never linted against them. Measured live before the wiring: 25
// ModalityMalformed and 6 InvalidREQID errors appear that CoverageRule's own
// severity treatment does not touch. Marking the newly-reachable entries lets
// each emission site report the finding while declining to gate on it, and
// leaves every pre-existing narrow entry byte-identical in behavior.
func parseREQsWithProvenance(body string) []REQEntry {
	narrow := parseREQs(body)
	narrowAt := make(map[int]string, len(narrow))
	for _, r := range narrow {
		narrowAt[r.Line] = r.ID
	}

	wide := parseREQsWide(body)
	for i := range wide {
		wide[i].Widened = narrowAt[wide[i].Line] != wide[i].ID
	}
	return wide
}

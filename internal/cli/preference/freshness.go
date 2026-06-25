// Package preference — M5 data-freshness disclosure (freshness.go).
//
// This file implements the REQ-ADM-015 / AC-ADM-015 [S3 Major] data-freshness
// disclosure. When the orchestrator places a recommendation grounded in the
// preference memory layer, the recommendation's option description MUST include
// a freshness label so the user can see how stale the underlying data is
// (e.g. "based on 7-day-old data").
//
// FreshnessLabel is a PURE function: it depends only on the integer age in
// days. Callers compute the age via M4's ageInDays(now, entry.LastUsed) and
// pass the integer here. It does NOT call time.Now().

package preference

import "fmt"

// FreshnessLabel returns the data-freshness disclosure string for the given age
// in days (REQ-ADM-015, AC-ADM-015). The label takes the form
// "based on N-day-old data" so the acceptance grep
// `based on .*-day-old data` matches for every age, including 0.
//
// A negative age (clock skew, future last_used) is clamped to 0 so the label
// never reports a negative day count, which would break the acceptance grep
// and read as nonsense.
//
// The orchestrator appends this label to recommendation option descriptions.
// It is NOT a standalone user-facing message; it is a disclosure fragment.
//
// @MX:NOTE: [AUTO] FreshnessLabel — REQ-ADM-015 신선도 공개 (AC-ADM-015 grep 기반)
// @MX:SPEC: SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 REQ-ADM-015, AC-ADM-015
func FreshnessLabel(ageDays int) string {
	if ageDays < 0 {
		// Defensive clamp — mirror decayWeight's negative-age handling so the
		// disclosure and the weight computation stay consistent.
		ageDays = 0
	}
	return fmt.Sprintf("based on %d-day-old data", ageDays)
}

package guardliveness

import (
	"fmt"
	"strings"
	"time"
)

// advisoryPrefix opens every line this package renders, so a reader scanning a
// session-start block can attribute it without reading further.
const advisoryPrefix = "moai guard liveness"

// contractViolationMarker is the phrase a contract-violating result renders
// under. It is a constant rather than a literal because it is the only thing
// separating "the mechanism looked and found nothing" from "the mechanism could
// not read what it was given" — and reporting the second as the first is this
// card's own subject at the consumer's layer (spec.md §A.0).
const contractViolationMarker = "the evaluation result could not be read, so this is NOT an all-clear"

// RenderRecord is what the previous render announced: every subject it read,
// paired with the classification it carried at that moment.
//
// It exists because a channel that reprints an identical block every session
// trains the filter that removes it, and a new advisory rendered beside a
// permanently non-clean neighbour inherits that filter (spec.md §A.8). Leading
// with what CHANGED requires something to compare against, and the comparison
// subject is the previously RENDERED state rather than the previously persisted
// result: a verdict nobody was shown is not one a reader can be assumed to know.
//
// The whole map is recorded, clean entries included, so an entry that goes
// non-clean again after recovering is announced rather than counted.
type RenderRecord struct {
	RenderedAt      time.Time         `json:"rendered_at"`
	Classifications map[string]string `json:"classifications"`
}

// Advisory renders the operator-facing text for a persisted snapshot, together
// with the record the caller should persist for the next render to diff
// against. A nil record means there is nothing to record — the snapshot was not
// a measurement, or its result could not be read — and the previous record must
// be left alone.
//
// Three outcomes, and the difference between them is the whole point:
//
//   - a result the partition could not be computed over renders the contract
//     violation BY NAME and never an all-clear (REQ-GDL-013). The path of least
//     resistance here is to identify nothing as non-clean and therefore render
//     nothing, which is green about a set that was never read;
//   - a conforming result carrying any entry that is not the clean value
//     renders (REQ-GDL-004). The condition is the partition and nothing else —
//     never a list of particular classifications. What is rendered is shaped by
//     REQ-GDL-007: entries whose classification changed since the previous
//     render are named, and the rest are a count;
//   - a conforming result with nothing to report renders nothing. A session
//     start is not a diagnostic report.
//
// The age is always derived from the snapshot's own recorded timestamp and
// never from now, which is what makes a stale advisory declare its staleness
// (REQ-GDL-006). now is a parameter rather than a call to time.Now so the
// derivation is observable rather than asserted.
func Advisory(s Snapshot, prev RenderRecord, now time.Time) (string, *RenderRecord) {
	if s.TakenAt.IsZero() {
		// Not a measurement. Rendering one as though it were current is the
		// failure the age disclosure exists to prevent, and recording one would
		// let the next render treat entries nobody saw as already announced.
		return "", nil
	}
	age := formatAge(now.Sub(s.TakenAt))

	nonClean, err := s.Result.Partition()
	if err != nil {
		return fmt.Sprintf("%s (measured %s ago) — %s: %s", advisoryPrefix, age, contractViolationMarker, err.Error()), nil
	}

	// Recorded whatever the outcome, including the all-clean one: the record is
	// the reader's state, and a session where they were told nothing is still a
	// session where nothing was outstanding.
	next := &RenderRecord{RenderedAt: now, Classifications: classificationsOf(s.Result)}
	if len(nonClean) == 0 {
		return "", next
	}

	changed, standing := splitByChange(nonClean, prev)

	var b strings.Builder
	if len(changed) == 0 {
		fmt.Fprintf(&b, "%s (measured %s ago) — %d subject(s) still not reporting clean, unchanged since the last session.",
			advisoryPrefix, age, len(standing))
		return b.String(), next
	}

	fmt.Fprintf(&b, "%s (measured %s ago) — %d subject(s) changed:", advisoryPrefix, age, len(changed))
	for _, e := range changed {
		fmt.Fprintf(&b, "\n  - %s", e.Subject)
	}
	if len(standing) > 0 {
		// A count rather than the entries. Re-rendering them is what a reader
		// learns to skip, and the skip does not distinguish the standing block
		// from the line above it.
		fmt.Fprintf(&b, "\n  (%d further subject(s) standing, unchanged since the last session)", len(standing))
	}
	return b.String(), next
}

// splitByChange divides the non-clean set into what the reader has not been
// told and what they have.
//
// An entry is changed when its classification differs from the one the previous
// render announced, which covers three cases a membership diff would miss: a
// subject that is new, one that was clean before, and one that moved between
// two non-clean classifications. The last is the reason the record holds
// classifications rather than a set of subjects.
func splitByChange(nonClean []Entry, prev RenderRecord) (changed, standing []Entry) {
	for _, e := range nonClean {
		was, announced := prev.Classifications[e.Subject]
		if !announced || was != e.Classifications[0] {
			changed = append(changed, e)
			continue
		}
		standing = append(standing, e)
	}
	return changed, standing
}

// classificationsOf flattens a result into the subject-to-classification map a
// later render diffs against. It is called only after Partition has succeeded,
// which is what guarantees every entry carries exactly one classification.
func classificationsOf(r Result) map[string]string {
	m := make(map[string]string, len(r.Entries))
	for _, e := range r.Entries {
		m[e.Subject] = e.Classifications[0]
	}
	return m
}

// formatAge renders a duration at a resolution a reader can act on. A negative
// age — a clock that moved backwards between the refresh and the render — is
// reported as zero rather than as a future measurement.
func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		hours := int(d.Hours())
		return fmt.Sprintf("%dh%02dm", hours, int(d.Minutes())-hours*60)
	}
}

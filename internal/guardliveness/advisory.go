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

// Advisory renders the operator-facing text for a persisted snapshot, or the
// empty string when there is nothing to say.
//
// Three outcomes, and the difference between them is the whole point:
//
//   - a result the partition could not be computed over renders the contract
//     violation BY NAME and never an all-clear (REQ-GDL-013). The path of least
//     resistance here is to identify nothing as non-clean and therefore render
//     nothing, which is green about a set that was never read;
//   - a conforming result carrying any entry that is not the clean value
//     renders those entries (REQ-GDL-004). The condition is the partition and
//     nothing else — never a list of particular classifications;
//   - a conforming result with nothing to report renders nothing. A session
//     start is not a diagnostic report.
//
// The age is always derived from the snapshot's own recorded timestamp and
// never from now, which is what makes a stale advisory declare its staleness
// (REQ-GDL-006). now is a parameter rather than a call to time.Now so the
// derivation is observable rather than asserted.
func Advisory(s Snapshot, now time.Time) string {
	if s.TakenAt.IsZero() {
		// Not a measurement. Rendering one as though it were current is the
		// failure the age disclosure exists to prevent.
		return ""
	}
	age := formatAge(now.Sub(s.TakenAt))

	nonClean, err := s.Result.Partition()
	if err != nil {
		return fmt.Sprintf("%s (measured %s ago) — %s: %s", advisoryPrefix, age, contractViolationMarker, err.Error())
	}
	if len(nonClean) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s (measured %s ago) — %d subject(s) not reporting clean:", advisoryPrefix, age, len(nonClean))
	for _, e := range nonClean {
		fmt.Fprintf(&b, "\n  - %s", e.Subject)
	}
	return b.String()
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

package importer

import (
	"fmt"
	"sort"
	"strings"
)

// Check compares the source-state counts with what the operator expects (for example
// {"backlog:picked": 101, "backlog:queued": 64}) and returns one line per mismatch. A key the plan never saw
// counts as zero, so an expected state that is absent from the sources is a mismatch, not a pass.
func (p Plan) Check(expect map[string]int) []string {
	keys := make([]string, 0, len(expect))
	for k := range expect {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		if got := p.Counts[k]; got != expect[k] {
			out = append(out, fmt.Sprintf("%s: found %d, expected %d", k, got, expect[k]))
		}
	}
	return out
}

// Report renders the plan for a person: what was read, where each card would land, what was skipped, and the
// notes that explain a decision. It is plain text; the same plan is what an apply step would use.
func (p Plan) Report() string {
	var b strings.Builder
	b.WriteString("Import plan (dry-run; nothing was written to any source or to the ledger)\n")

	keys := make([]string, 0, len(p.Counts))
	for k := range p.Counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("Source rows by state:\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "  %-24s %d\n", k, p.Counts[k])
	}

	landing := map[string]int{}
	for _, pr := range p.Proposals {
		pos := pr.Card.Position
		key := string(pos.Stage)
		if pos.Park != nil {
			key += " (parked: " + string(pos.Park.Reason) + ")"
		}
		landing[key]++
	}
	keys = keys[:0]
	for k := range landing {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(&b, "Cards that would be created: %d\n", len(p.Proposals))
	for _, k := range keys {
		fmt.Fprintf(&b, "  %-32s %d\n", k, landing[k])
	}
	fmt.Fprintf(&b, "Source rows skipped: %d\n", len(p.Skips))
	for _, s := range p.Skips {
		fmt.Fprintf(&b, "  %s %s: %s\n", s.Source, s.SourceID, s.Reason)
	}
	withNotes := 0
	for _, pr := range p.Proposals {
		if len(pr.Notes) > 0 {
			withNotes++
		}
	}
	fmt.Fprintf(&b, "Cards with a note (a decision or something the ledger cannot hold): %d\n", withNotes)
	fmt.Fprintf(&b, "Every imported card gets class %s (the careful route) and no acceptance list: triage must set both before it is claimed.\n", DefaultClass)
	return b.String()
}

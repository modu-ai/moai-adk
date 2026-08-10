package delegationmap

import (
	"sort"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// routingLedgerFileName re-exports the producer's canonical ledger file name.
// The indirection is the point: this package declares no ledger path literal of
// its own, so it cannot drift from the file the producer writes (REQ-HLA-001).
const routingLedgerFileName = routing.LedgerFileName

// unattributedMarker re-exports the producer's absent-identity marker for the
// same reason. Roughly 30% of observed subagent stops carry no identity, and
// they must stay countable rather than being folded into an agent's count.
const unattributedMarker = routing.AgentUnattributed

// aggregate folds rows into per-subcommand statistics, retaining no per-row
// state beyond the running aggregate (REQ-HLA-001).
//
// A note on what "observation count" means here, because the choice changes
// what a support ratio is: a per-agent count is the number of QUALIFYING ROWS
// in which the agent appeared, deduplicated within a row. Counting entries
// instead would let a ratio exceed 1 and would make a single row that delegated
// twice look like two independent observations. The unattributed count is the
// opposite — entry count, not row-presence — because the share a reviewer needs
// is how much of the delegation population carried no identity.
//
// Row-splitting caveat (plan.md §B6, open as SPEC-HARNESS-LEARNING-EVO-001 R8):
// a session's delegations can land on a different row than its outcome, so a
// delegation may be absent from the row counted as qualifying. Both resulting
// errors — under-counted support ratios and over-counted empty-delegation rows —
// push toward suppressing real patterns rather than inventing false ones. That
// makes the degradation quiet, which is why EmptyDelegationRows is reported per
// subcommand: a subcommand whose qualifying rows are largely empty is the
// observable signature of the split, and a reviewer can see it in the result
// rather than having to infer it.
func aggregate(rows []routing.Row) map[string]*SubcommandStat {
	stats := make(map[string]*SubcommandStat)

	statFor := func(subcommand string) *SubcommandStat {
		s, ok := stats[subcommand]
		if !ok {
			s = &SubcommandStat{Subcommand: subcommand, AgentCounts: map[string]int{}}
			stats[subcommand] = s
		}
		return s
	}

	for _, row := range rows {
		s := statFor(row.MatchedSubcommand)

		switch row.Outcome {
		case routing.OutcomeReroute:
			s.RerouteRows++
			continue
		case routing.OutcomeAbort:
			s.AbortRows++
			continue
		}
		if !isQualifying(row.Outcome) {
			continue // an unrecognized outcome contributes nothing
		}

		s.QualifyingRows++
		if len(row.Delegations) == 0 {
			s.EmptyDelegationRows++
		}

		// Deduplicate within the row so a per-agent count stays row-presence.
		seen := make(map[string]struct{}, len(row.Delegations))
		for _, d := range row.Delegations {
			if d.Agent == unattributedMarker {
				s.UnattributedEntries++
				continue
			}
			if d.Agent == "" {
				continue // a value the producer never emits; ignore rather than count
			}
			if _, dup := seen[d.Agent]; dup {
				continue
			}
			seen[d.Agent] = struct{}{}
			s.AgentCounts[d.Agent]++
			if !IsRetainedAgent(d.Agent) {
				s.addNonCatalog(d.Agent)
			}
		}
	}

	for _, s := range stats {
		sort.Strings(s.NonCatalogAgents)
	}
	return stats
}

// addNonCatalog records an observed value as non-catalog exactly once.
func (s *SubcommandStat) addNonCatalog(agent string) {
	for _, existing := range s.NonCatalogAgents {
		if existing == agent {
			return
		}
	}
	s.NonCatalogAgents = append(s.NonCatalogAgents, agent)
}

// sortedStats flattens the aggregate into a deterministically ordered slice.
func sortedStats(stats map[string]*SubcommandStat) []SubcommandStat {
	out := make([]SubcommandStat, 0, len(stats))
	for _, s := range stats {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Subcommand < out[j].Subcommand })
	return out
}

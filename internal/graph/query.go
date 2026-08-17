package graph

import "sort"

// FanInEntry is one row of the import fan-in ranking.
type FanInEntry struct {
	Package string
	FanIn   int
}

// ImportFanIn ranks packages by the number of distinct importers (import
// edges targeting the package), highest first, ties broken by package path.
// limit <= 0 returns the full ranking.
//
// @MX:NOTE: [AUTO] ImportFanIn — import fan-in ranking stands in for an @MX:DEBT fan-in query until a tag-kind edge lands in edges.jsonl
func ImportFanIn(edges []Edge, limit int) []FanInEntry {
	counts := map[string]map[string]bool{}
	for _, e := range edges {
		if e.Kind != KindImport {
			continue
		}
		if counts[e.Target] == nil {
			counts[e.Target] = map[string]bool{}
		}
		counts[e.Target][e.Source] = true
	}

	entries := make([]FanInEntry, 0, len(counts))
	for pkg, importers := range counts {
		entries = append(entries, FanInEntry{Package: pkg, FanIn: len(importers)})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].FanIn != entries[j].FanIn {
			return entries[i].FanIn > entries[j].FanIn
		}
		return entries[i].Package < entries[j].Package
	})
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}

// UnreferencedSpecs returns the sorted universe SPEC ids that no mx-spec
// edge targets. The universe comes from the caller (the CLI passes spec.md
// frontmatter ids) so this stays a pure function over the artifact.
func UnreferencedSpecs(edges []Edge, universe []string) []string {
	referenced := map[string]bool{}
	for _, e := range edges {
		if e.Kind == KindMXSpec {
			referenced[e.Target] = true
		}
	}

	var out []string
	for _, id := range universe {
		if id != "" && !referenced[id] {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// MilestoneClaim is one cross-check row read back from the artifact: the
// milestone node (report-stem qualified) and the queue cards its
// milestone-card edges claim.
type MilestoneClaim struct {
	Milestone string
	Cards     []string
}

// MilestoneClaims returns every milestone a report-milestone edge declares,
// with the cards its milestone-card edges claim, sorted by milestone node.
// The live-queue cross-check stays with the caller: the artifact records
// what the report CLAIMS, the queue records what exists, and the gate is
// the comparison between the two.
//
// @MX:NOTE: [AUTO] MilestoneClaims — claim-side half of the milestone↔card gate; the queue-side half lives in the CLI where the live backlog is readable
func MilestoneClaims(edges []Edge) []MilestoneClaim {
	cards := map[string][]string{}
	declared := map[string]bool{}
	for _, e := range edges {
		switch e.Kind {
		case KindReportMilestone:
			declared[e.Target] = true
		case KindMilestoneCard:
			cards[e.Source] = append(cards[e.Source], e.Target)
		}
	}

	out := make([]MilestoneClaim, 0, len(declared))
	for milestone := range declared {
		list := append([]string(nil), cards[milestone]...)
		sort.Strings(list)
		out = append(out, MilestoneClaim{Milestone: milestone, Cards: list})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Milestone < out[j].Milestone })
	return out
}

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

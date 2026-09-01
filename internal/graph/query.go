package graph

import "sort"

// FanInEntry is one row of the import fan-in ranking.
type FanInEntry struct {
	Package string
	FanIn   int
}

// ImportFanIn ranks packages by the number of distinct importers (import
// edges targeting the package), highest first, ties broken by package path.
// limit <= 0 returns the full ranking. For @MX:DEBT fan-in use DebtFanIn /
// SymbolFanIn — the tag-kind edge layer answers that question directly.
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

// FanInResult is the evidence-backed fan-in of one symbol name (REQ-MTE-009):
// DISTINCT caller FILES of code-call edges targeting the name, split by
// resolution confidence. EvidenceFiles carry extracted or intra-package
// resolution; InferredFiles carry inferred resolution only — counted
// separately, never added to the blocking count (REQ-MTE-013). Both lists
// are sorted and exclude declaringFile.
type FanInResult struct {
	// EvidenceFiles are the evidence-backed caller files, sorted.
	EvidenceFiles []string
	// InferredFiles are the caller files whose calls are inferred-confidence
	// only, sorted.
	InferredFiles []string
}

// Evidence is the blocking fan-in: the number of distinct evidence-backed
// caller files.
func (r FanInResult) Evidence() int { return len(r.EvidenceFiles) }

// InferredOnly is the number of distinct inferred-only caller files.
func (r FanInResult) InferredOnly() int { return len(r.InferredFiles) }

// SymbolFanIn answers REQ-MTE-009's pure query over the artifact: the
// fan-in of symbolName is the set of DISTINCT caller files of code-call
// edges targeting that name, EXCLUDING the declaring files (ANCHOR's
// semantics — the tag records an invariant contract owed to EXTERNAL
// dependents; a same-file call is not external blast radius). A caller file
// with any evidence-backed call counts as evidence; only files whose every
// call is inferred-confidence land in InferredFiles. Callers whose
// resolution is unknown (pre-confidence artifacts) count nowhere.
//
// declaringFiles are repo-relative file paths, the same shape code-call
// edge sources carry.
func SymbolFanIn(edges []Edge, symbolName string, declaringFiles ...string) FanInResult {
	excluded := map[string]bool{}
	for _, f := range declaringFiles {
		excluded[f] = true
	}
	tier := map[string]int{} // 1 = inferred, 2 = evidence-backed
	for _, e := range edges {
		if e.Kind != KindCodeCall || e.Target != symbolName {
			continue
		}
		file, _ := splitCodeNode(e.Source)
		if file == "" || excluded[file] {
			continue
		}
		switch e.Resolution {
		case ResolutionExtracted, ResolutionIntraPackage:
			tier[file] = 2
		case ResolutionInferred:
			if tier[file] == 0 {
				tier[file] = 1
			}
		}
	}

	res := FanInResult{}
	for file, t := range tier {
		switch t {
		case 2:
			res.EvidenceFiles = append(res.EvidenceFiles, file)
		case 1:
			res.InferredFiles = append(res.InferredFiles, file)
		}
	}
	sort.Strings(res.EvidenceFiles)
	sort.Strings(res.InferredFiles)
	return res
}

// DebtFanInEntry is one row of the mx-debt fan-in ranking.
type DebtFanInEntry struct {
	// Target is the symbol name, or the file path for a self-edge.
	Target string
	// File is the declaring file (the mx-debt edge's source); for a
	// self-edge, the file itself.
	File string
	// FanIn is the evidence-backed graph fan-in (REQ-MTE-009). Self-edge
	// targets rank at 0 by definition — listed, never omitted
	// (REQ-MTE-014).
	FanIn int
	// Self marks a file-scope DEBT tag (self-edge, no enclosing symbol).
	Self bool
}

// DebtFanIn ranks mx-debt edge targets by the REQ-MTE-009 graph fan-in,
// descending, ties broken by target (REQ-MTE-014). The declaring file of a
// symbol target is the mx-debt edge's own source — the file the tag lives
// in IS the file that declares the tagged symbol. limit <= 0 returns the
// full ranking.
func DebtFanIn(edges []Edge, limit int) []DebtFanInEntry {
	type agg struct {
		declFiles map[string]bool
		self      bool
		file      string // smallest source, the deterministic representative
	}
	byTarget := map[string]*agg{}
	for _, e := range edges {
		if e.Kind != KindMXDebt {
			continue
		}
		a := byTarget[e.Target]
		if a == nil {
			a = &agg{declFiles: map[string]bool{}}
			byTarget[e.Target] = a
		}
		if e.Source == e.Target {
			a.self = true
			a.file = e.Source
			continue
		}
		a.declFiles[e.Source] = true
		if a.file == "" || e.Source < a.file {
			a.file = e.Source
		}
	}

	excluded := make([]string, 0, 4)
	entries := make([]DebtFanInEntry, 0, len(byTarget))
	for target, a := range byTarget {
		if a.self {
			entries = append(entries, DebtFanInEntry{Target: target, File: a.file, FanIn: 0, Self: true})
			continue
		}
		excluded = excluded[:0]
		for f := range a.declFiles {
			excluded = append(excluded, f)
		}
		sort.Strings(excluded)
		res := SymbolFanIn(edges, target, excluded...)
		entries = append(entries, DebtFanInEntry{Target: target, File: a.file, FanIn: res.Evidence()})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].FanIn != entries[j].FanIn {
			return entries[i].FanIn > entries[j].FanIn
		}
		return entries[i].Target < entries[j].Target
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

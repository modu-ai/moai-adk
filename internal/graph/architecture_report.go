package graph

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// architecture_report.go — SPEC-GRAPH-REPORT-001 M2 (REQ-GR-005..007): the
// three deterministic report sections computed as PURE functions over the
// persisted edge list. Named architecture_report.go to avoid the t107
// report.go collision (that file owns the report-milestone cross-check EDGE
// LAYER; this file owns the human-facing architecture REPORT).
//
// Package identity uses the DIRECTORY PROXY (REQ-GR-005): each endpoint's
// package is its source-file directory under the project root. A code-call
// edge's Target persists only a bare callee name, so the callee's package
// resolves only when exactly one node in the edge set carries that function
// name — ambiguous or unresolved callees are excluded, never guessed onto a
// package. This is a directory-level identity proxy, not Go package-name
// resolution.
//
// @MX:NOTE: [AUTO] determinism contract — stable sorts and a total order everywhere; no wall-clock in the body, so two runs over the same tree are byte-identical (REQ-GR-005)

// GraphReportRelPath is the FIXED rotating report location under the resolved
// project root (REQ-GR-005, D1 ADOPTED): a REGENERATING derived artifact —
// always the latest view, never committed. No --out flag exists to redirect
// it; a user-selectable path would let the report land on a committed
// location, defeating the derived-artifacts-never-committed contract.
const GraphReportRelPath = ".moai/reports/graph-report.md"

// codeLayerAbsentReason is the stated reason every code-dependent section
// carries when the artifact holds zero code-call edges (REQ-GR-006) — a
// nocgo build or a tree with no extraction, indistinguishable from the
// artifact alone, so the phrase covers both.
const codeLayerAbsentReason = "code layer absent: CGO disabled or no extraction"

// codeNodeIndex builds the fn-name → node-ids index over code-call Sources —
// the node universe the directory proxy resolves bare callee names against
// (same shape as ShortestPath's index; a name carried by exactly one node is
// unambiguous).
func codeNodeIndex(edges []Edge) map[string][]string {
	byName := map[string][]string{}
	for _, e := range edges {
		if e.Kind != KindCodeCall {
			continue
		}
		if _, fn := splitCodeNode(e.Source); fn != "" && !slices.Contains(byName[fn], e.Source) {
			byName[fn] = append(byName[fn], e.Source)
		}
	}
	for fn := range byName {
		sort.Strings(byName[fn])
	}
	return byName
}

// packageDir renders a repo-relative file path as its package directory in
// slash form — the directory proxy's unit of package identity.
func packageDir(file string) string {
	return filepath.ToSlash(filepath.Dir(file))
}

// ─── section 1: god nodes (REQ-GR-005 §1, REQ-GR-007) ───

// GodNode is one row of the god-nodes fan-in ranking.
type GodNode struct {
	Node  string
	FanIn int
}

// GodNodesResult is the god-nodes section payload: the ranked rows plus the
// edge kinds the aggregation counted — the provenance label REQ-GR-007 pins
// (report-only: the MX validator's grep-based fanInIndex is untouched).
type GodNodesResult struct {
	Nodes []GodNode
	Kinds []string
}

// GodNodes ranks targets by distinct-source fan-in over the import and
// code-call layers, highest first, ties by node id. Import targets are
// package paths; code-call targets are bare callee names normalized to their
// package directory via the directory proxy (ambiguous or unresolved callees
// excluded). limit > 0 truncates; limit <= 0 returns the full ranking. The
// fan-in counts DISTINCT raw Source values (a package id for import edges, a
// file:function node id for code-call edges) — normalization is defined on
// targets only.
func GodNodes(edges []Edge, limit int) GodNodesResult {
	byName := codeNodeIndex(edges)
	sources := map[string]map[string]bool{}
	kindSet := map[string]bool{}
	add := func(kind, node, src string) {
		if sources[node] == nil {
			sources[node] = map[string]bool{}
		}
		sources[node][src] = true
		kindSet[kind] = true
	}
	for _, e := range edges {
		switch e.Kind {
		case KindImport:
			add(e.Kind, e.Target, e.Source)
		case KindCodeCall:
			node := resolveName(e.Target, byName)
			if node == "" {
				continue // ambiguous or unresolved callee: excluded, never guessed
			}
			file, _ := splitCodeNode(node)
			add(e.Kind, packageDir(file), e.Source)
		}
	}

	res := GodNodesResult{Nodes: make([]GodNode, 0, len(sources))}
	for node, srcs := range sources {
		res.Nodes = append(res.Nodes, GodNode{Node: node, FanIn: len(srcs)})
	}
	sort.Slice(res.Nodes, func(i, j int) bool {
		if res.Nodes[i].FanIn != res.Nodes[j].FanIn {
			return res.Nodes[i].FanIn > res.Nodes[j].FanIn
		}
		return res.Nodes[i].Node < res.Nodes[j].Node
	})
	if limit > 0 && len(res.Nodes) > limit {
		res.Nodes = res.Nodes[:limit]
	}
	res.Kinds = sortedKindSet(kindSet)
	return res
}

// sortedKindSet renders a kind set as a sorted slice (the kinds line is
// itself part of the deterministic output).
func sortedKindSet(kindSet map[string]bool) []string {
	kinds := make([]string, 0, len(kindSet))
	for k := range kindSet {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}

// ─── section 2: surprising connections (REQ-GR-005 §2) ───

// SurprisingConnection is one INFERRED code-call edge whose endpoints'
// package directories differ: the boundary-crossing, name-only-resolved
// calls most likely to hide a real dependency the import layer never
// recorded. To is the RESOLVED callee node id — ambiguous and unresolved
// callees never reach the section.
type SurprisingConnection struct {
	From string // caller node id (file:function)
	To   string // resolved callee node id (file:function)
	// FromPkg/ToPkg are the endpoints' package directories (directory proxy).
	FromPkg string
	ToPkg   string
	Line    int
	Grade   string
	// Confidence is the source edge's INFERRED resolution confidence
	// (SPEC-GRAPH-EDGE-CONFIDENCE-001) — the section's ranking key.
	Confidence float64
}

// SurprisingConnections selects the INFERRED code-call edges whose endpoints
// sit in different package directories, ranked by confidence descending then
// the total order (from, to, line). The cross-package selector realizes
// REQ-GR-005's ranking clause — a boundary edge outranks any same-confidence
// intra-package edge, which the selector keeps out of the section entirely;
// edges without INFERRED resolution (extracted, intra-package, or legacy
// unlabeled artifacts) are never scored.
func SurprisingConnections(edges []Edge) []SurprisingConnection {
	byName := codeNodeIndex(edges)
	var out []SurprisingConnection
	for _, e := range edges {
		if e.Kind != KindCodeCall || e.Resolution != ResolutionInferred {
			continue
		}
		to := resolveName(e.Target, byName)
		if to == "" {
			continue // ambiguous or unresolved callee: excluded, never guessed
		}
		fromFile, _ := splitCodeNode(e.Source)
		toFile, _ := splitCodeNode(to)
		fromPkg, toPkg := packageDir(fromFile), packageDir(toFile)
		if fromPkg == toPkg {
			continue // intra-package: not a surprising connection
		}
		out = append(out, SurprisingConnection{
			From: e.Source, To: to, FromPkg: fromPkg, ToPkg: toPkg,
			Line: e.Line, Grade: e.Grade, Confidence: e.Confidence,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		return out[i].Line < out[j].Line
	})
	return out
}

// ─── section 3: import cycles (REQ-GR-005 §3, D26) ───

// ImportSCC is one import-layer strongly connected component. Members is the
// canonical member list, sorted ascending (smallest node id first). Rotation
// is non-empty exactly when the SCC is one simple cycle through ALL its
// members — the cycle walked from the smallest member following edge
// direction. A branched SCC (D26) can contain no simple cycle through all
// members, so membership, not a fabricated cycle, is what renders.
type ImportSCC struct {
	Members  []string
	Rotation []string
}

// SimpleCycle reports whether the SCC is exactly one simple cycle through
// all its members.
func (s ImportSCC) SimpleCycle() bool { return len(s.Rotation) > 0 }

// ImportCycles groups import edges into SCCs (Tarjan) and returns every
// non-trivial component — size >= 2, or size 1 with a self-edge — sorted by
// smallest member. The SCC count, never a simple-cycle enumeration, is the
// section's primary datum.
func ImportCycles(edges []Edge) []ImportSCC {
	adj := map[string]map[string]bool{}
	nodeSet := map[string]bool{}
	for _, e := range edges {
		if e.Kind != KindImport {
			continue
		}
		if adj[e.Source] == nil {
			adj[e.Source] = map[string]bool{}
		}
		adj[e.Source][e.Target] = true
		nodeSet[e.Source] = true
		nodeSet[e.Target] = true
	}
	nodes := make([]string, 0, len(nodeSet))
	for n := range nodeSet {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	// Tarjan's algorithm over the deterministically ordered node list, so
	// the raw component discovery order is itself stable.
	index := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	counter := 0
	var comps [][]string
	var strongconnect func(v string)
	strongconnect = func(v string) {
		index[v] = counter
		low[v] = counter
		counter++
		stack = append(stack, v)
		onStack[v] = true
		succs := make([]string, 0, len(adj[v]))
		for w := range adj[v] {
			succs = append(succs, w)
		}
		sort.Strings(succs)
		for _, w := range succs {
			if _, seen := index[w]; !seen {
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] && index[w] < low[v] {
				low[v] = index[w]
			}
		}
		if low[v] == index[v] {
			var comp []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			comps = append(comps, comp)
		}
	}
	for _, v := range nodes {
		if _, seen := index[v]; !seen {
			strongconnect(v)
		}
	}

	var out []ImportSCC
	for _, comp := range comps {
		sort.Strings(comp)
		inSCC := make(map[string]bool, len(comp))
		for _, m := range comp {
			inSCC[m] = true
		}
		if len(comp) == 1 && !adj[comp[0]][comp[0]] {
			continue // trivial SCC without a self-edge: not a cycle
		}
		// Simple-cycle test: every member has exactly one distinct successor
		// and one distinct predecessor INSIDE the SCC — with strong
		// connectivity that forces the component to be a single cycle.
		succ, pred := map[string]int{}, map[string]int{}
		for _, m := range comp {
			for w := range adj[m] {
				if inSCC[w] {
					succ[m]++
					pred[w]++
				}
			}
		}
		simple := true
		for _, m := range comp {
			if succ[m] != 1 || pred[m] != 1 {
				simple = false
				break
			}
		}
		entry := ImportSCC{Members: comp}
		if simple {
			rotation := []string{comp[0]}
			cur := comp[0]
			for {
				next := ""
				for w := range adj[cur] {
					if inSCC[w] {
						next = w // exactly one by the simple-cycle test
						break
					}
				}
				if next == comp[0] {
					break
				}
				rotation = append(rotation, next)
				cur = next
			}
			entry.Rotation = rotation
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Members[0] < out[j].Members[0]
	})
	return out
}

// ─── rendering (REQ-GR-005/006) ───

// RenderArchitectureReport renders the deterministic markdown body: the
// three sections in fixed order, each present even when empty with its
// reason stated (REQ-GR-006). No wall-clock, no provenance block — the body
// is a pure function of the edge list, so two runs over the same tree are
// byte-identical (REQ-GR-005).
func RenderArchitectureReport(edges []Edge, limit int) string {
	var b strings.Builder
	b.WriteString("# Graph Report\n\n")

	// Section 1 — god nodes.
	gn := GodNodes(edges, limit)
	b.WriteString("## God Nodes\n\n")
	if len(gn.Nodes) == 0 {
		if hasCodeCallEdges(edges) {
			b.WriteString("no fan-in evidence: no import edges and no unambiguous code-call targets\n")
		} else {
			b.WriteString("no fan-in evidence: no import edges and " + codeLayerAbsentReason + "\n")
		}
	} else {
		b.WriteString("fan-in over: " + strings.Join(gn.Kinds, ", ") + "\n\n")
		if !hasCodeCallEdges(edges) {
			b.WriteString("(" + codeLayerAbsentReason + " — ranking reflects the import layer only)\n\n")
		}
		b.WriteString("| fan-in | node |\n| -----: | ---- |\n")
		for _, n := range gn.Nodes {
			fmt.Fprintf(&b, "| %d | %s |\n", n.FanIn, n.Node)
		}
	}
	b.WriteString("\n")

	// Section 2 — surprising connections.
	b.WriteString("## Surprising Connections\n\n")
	conns := SurprisingConnections(edges)
	if len(conns) == 0 {
		if !hasCodeCallEdges(edges) {
			b.WriteString(codeLayerAbsentReason + "\n")
		} else {
			b.WriteString("none: no unambiguous INFERRED cross-package code-call edges\n")
		}
	} else {
		for _, c := range conns {
			line := ""
			if c.Line > 0 {
				line = ", line " + strconv.Itoa(c.Line)
			}
			fmt.Fprintf(&b, "- %s -> %s (%s -> %s, confidence %s%s)\n",
				c.From, c.To, c.FromPkg, c.ToPkg, formatConfidence(c.Confidence), line)
		}
	}
	b.WriteString("\n")

	// Section 3 — import cycles.
	b.WriteString("## Import Cycles\n\n")
	sccs := ImportCycles(edges)
	if len(sccs) == 0 {
		hasImports := false
		for _, e := range edges {
			if e.Kind == KindImport {
				hasImports = true
				break
			}
		}
		if hasImports {
			b.WriteString("none: the import layer carries no cycles\n")
		} else {
			b.WriteString("import layer empty: no import edges in the artifact\n")
		}
	} else {
		fmt.Fprintf(&b, "SCCs: %d\n\n", len(sccs))
		for _, s := range sccs {
			if s.SimpleCycle() {
				fmt.Fprintf(&b, "- cycle: %s -> %s\n", strings.Join(s.Rotation, " -> "), s.Rotation[0])
			} else {
				fmt.Fprintf(&b, "- members: %s (branched SCC — no simple cycle through all members)\n", strings.Join(s.Members, ", "))
			}
		}
	}
	return b.String()
}

// formatConfidence renders a confidence value without trailing zeros (0.85,
// 0.5) — stable across runs for the same input.
func formatConfidence(c float64) string {
	return strconv.FormatFloat(c, 'f', -1, 64)
}

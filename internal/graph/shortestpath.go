package graph

import (
	"fmt"
	"slices"
	"sort"
)

// shortestpath.go — the A→B reachability query over the code-derived edge
// layer (SPEC-GRAPH-REPORT-001 REQ-GR-001..004): BFS over KindCodeCall edges,
// bounded by the SAME maxTraceDepth const that bounds graph_trace_calls
// (one bound, two consumers — never a duplicated literal).

// PathHop is one traversed code-call edge on the shortest path. Endpoints
// are node ids in the stored `file:function` shape; the line rides the
// edge's Line field, never inside the node id (REQ-GR-001).
type PathHop struct {
	From string `json:"from"` // caller node id (file:function)
	To   string `json:"to"`   // callee node id (file:function)
	Line int    `json:"line"`
	// Grade is the per-language capability grade of the source edge;
	// Confidence is its resolution confidence (REQ-GEC-008), 0/omitted on
	// legacy artifacts (REQ-GEC-009).
	Grade      string  `json:"grade,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// PathCandidate disambiguates an ambiguous bare-name endpoint: the queried
// name plus every node id whose function part matches, sorted (REQ-GR-003).
type PathCandidate struct {
	Name  string   `json:"name"`
	Nodes []string `json:"nodes"`
}

// PathResult is the graph_shortest_path answer: the hop list when found, or
// the structured not-found / candidates shape otherwise — every field a
// valid answer, never a transport-level error (REQ-GR-003).
type PathResult struct {
	From string `json:"from"` // as supplied by the caller
	To   string `json:"to"`   // as supplied by the caller
	// FromNode/ToNode are the resolved node ids when each endpoint resolved.
	FromNode string    `json:"from_node,omitempty"`
	ToNode   string    `json:"to_node,omitempty"`
	Found    bool      `json:"found"`
	Hops     []PathHop `json:"hops,omitempty"`
	HopCount int       `json:"hop_count"`
	// Cap restates the traversal bound (maxTraceDepth) so the caller can
	// distinguish "no path" from "no path within the cap".
	Cap int `json:"cap"`
	// Reason names both endpoints and the cap on every not-found shape.
	Reason     string          `json:"reason,omitempty"`
	Candidates []PathCandidate `json:"candidates,omitempty"`
	// Provenance names the tree root + commit the answer was computed from.
	Provenance string `json:"provenance"`
}

// pathWalk records how BFS discovered a node: its parent node id plus the
// traversed edge, for shortest-path reconstruction.
type pathWalk struct {
	parent string
	hop    PathHop
}

// ShortestPath answers the A→B reachability query with a deterministic BFS
// over KindCodeCall edges indexed by caller node id. Endpoints are node ids
// in the stored `file:function` shape; a bare symbol name is accepted only
// when it resolves to exactly one node — ambiguous names yield the
// candidates list, name-only (Target-callee) names yield not-found
// (REQ-GR-001/003). Neighbor iteration follows the total order (node id,
// then line) so two runs over the same tree produce byte-identical results
// (REQ-GR-004).
func ShortestPath(projectRoot, from, to string) (PathResult, error) {
	edges, err := loadCodeEdges(projectRoot)
	if err != nil {
		return PathResult{}, err
	}

	res := PathResult{From: from, To: to, Cap: maxTraceDepth, Provenance: AnswerProvenance(projectRoot)}

	// Node universe: distinct code-call Sources are the only persisted node
	// ids — a function name appearing only as a Target callee has none.
	byName := map[string][]string{}
	adj := map[string][]Edge{}
	for _, e := range edges {
		if e.Kind != KindCodeCall {
			continue
		}
		if _, fn := splitCodeNode(e.Source); fn != "" && !slices.Contains(byName[fn], e.Source) {
			byName[fn] = append(byName[fn], e.Source)
		}
		adj[e.Source] = append(adj[e.Source], e)
	}
	for fn := range byName {
		sort.Strings(byName[fn])
	}

	// resolveEndpoint maps a supplied endpoint to its node id. A node-id
	// shape is valid only when the artifact actually carries it; a bare
	// symbol name resolves only on a unique match. Ambiguity is reported
	// through Candidates; a no-match leaves the node id empty (the honest
	// not-found, never a guessed join).
	resolveEndpoint := func(name string) (string, bool) {
		if _, ok := adj[name]; ok {
			return name, true // exact node id the artifact carries
		}
		ids := byName[name] // bare symbol name (node ids contain ':')
		if len(ids) == 1 {
			return ids[0], true
		}
		if len(ids) > 1 {
			res.Candidates = append(res.Candidates, PathCandidate{Name: name, Nodes: ids})
		}
		return "", false
	}

	fromNode, fromOK := resolveEndpoint(from)
	if fromOK {
		res.FromNode = fromNode
	}
	toNode, toOK := resolveEndpoint(to)
	if toOK {
		res.ToNode = toNode
	}
	if !fromOK || !toOK {
		if len(res.Candidates) > 0 {
			res.Reason = fmt.Sprintf("ambiguous endpoint %q matches %d nodes — disambiguate with a file:function node id",
				res.Candidates[0].Name, len(res.Candidates[0].Nodes))
		} else {
			res.Reason = fmt.Sprintf("no path from %s to %s within %d hops", from, to, maxTraceDepth)
		}
		return res, nil
	}
	if fromNode == toNode {
		res.Found = true
		return res, nil
	}

	// Deterministic BFS: each node's outgoing edges sorted by (resolved
	// neighbor node id, then line) — the total order REQ-GR-004 pins. A
	// bare callee resolving to 2+ nodes is NO continuation (REQ-GR-003):
	// never joined through into an unrelated node.
	walks := map[string]pathWalk{fromNode: {}}
	frontier := []string{fromNode}
	for hopCount := 1; hopCount <= maxTraceDepth; hopCount++ {
		var next []string
		for _, node := range frontier {
			for _, nb := range sortedNeighbors(adj[node], byName) {
				if _, seen := walks[nb.node]; seen {
					continue
				}
				walks[nb.node] = pathWalk{parent: node, hop: PathHop{
					From: node, To: nb.node, Line: nb.edge.Line,
					Grade: nb.edge.Grade, Confidence: nb.edge.Confidence,
				}}
				if nb.node == toNode {
					res.Found = true
					res.HopCount = hopCount
					res.Hops = reconstruct(walks, fromNode, toNode)
					return res, nil
				}
				next = append(next, nb.node)
			}
		}
		frontier = next
	}
	res.Reason = fmt.Sprintf("no path from %s to %s within %d hops", from, to, maxTraceDepth)
	return res, nil
}

// resolveName maps a persisted bare callee name to its single node id, ""
// when it is ambiguous or carries no node (REQ-GR-003).
func resolveName(name string, byName map[string][]string) string {
	if ids := byName[name]; len(ids) == 1 {
		return ids[0]
	}
	return ""
}

// neighborEdge pairs a resolved neighbor node id with its edge for sorting.
type neighborEdge struct {
	node string
	edge Edge
}

// sortedNeighbors yields a node's outgoing edges with single-resolution
// targets, in the total order (neighbor node id, then line) — the iteration
// order that makes two runs byte-identical (REQ-GR-004).
func sortedNeighbors(edges []Edge, byName map[string][]string) []neighborEdge {
	out := make([]neighborEdge, 0, len(edges))
	for _, e := range edges {
		if n := resolveName(e.Target, byName); n != "" {
			out = append(out, neighborEdge{node: n, edge: e})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].node != out[j].node {
			return out[i].node < out[j].node
		}
		return out[i].edge.Line < out[j].edge.Line
	})
	return out
}

// reconstruct walks the BFS parent chain back from toNode to fromNode,
// emitting hops in traversal order.
func reconstruct(walks map[string]pathWalk, fromNode, toNode string) []PathHop {
	var hops []PathHop
	for n := toNode; n != fromNode; n = walks[n].parent {
		hops = append(hops, walks[n].hop)
	}
	for i, j := 0, len(hops)-1; i < j; i, j = i+1, j-1 {
		hops[i], hops[j] = hops[j], hops[i]
	}
	return hops
}

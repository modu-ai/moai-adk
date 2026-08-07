// Package detect implements the M1.1 reverse-traversal engine of the BAS
// Falconer Detect layer (SPEC-NAVIGATOR-SYNC-002 M1.1, REQ-NS2-002).
//
// The package CONSUMES the M0 graph types from internal/navigator/sync
// (REQ-NS2-005 bridge-not-absorb): it imports sync.Graph / sync.Edge /
// sync.Node read-only and NEVER mutates the M0 producer surface. The
// traversal function is a pure Go function — no I/O, no side effects. The
// caller owns graph loading and fail-open policy (REQ-NS2-004); this package
// returns errors so the caller can decide the failure semantics.
package detect

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// Result is the affected-row set returned by Traverse: the deduplicated,
// deterministically-ordered list of affected nodes and the originating edges
// (REQ-NS2-002).
type Result struct {
	// Nodes is the deduplicated affected-node set, sorted by
	// (entity_type, identifier). Both endpoints of each matching edge are
	// collected — conservative over-inclusion is safe for an advisory layer.
	Nodes []AffectedNode
	// Edges is the deduplicated originating-edge set, sorted by
	// (edge_type, source_node, target_node, source_path, line_number),
	// mirroring M0's sortEdges (internal/navigator/sync/join.go:370).
	Edges []AffectedEdge
}

// AffectedNode is an affected graph node with its display name resolved
// against the M0 graph's Node table. The Key field is the canonical
// "<entity_type>:<identifier>" reference form used by Edge.SourceNode /
// Edge.TargetNode.
type AffectedNode struct {
	Key         string
	EntityType  navsync.EntityType
	Identifier  string
	DisplayName string
}

// AffectedEdge is the originating edge verbatim — the M0 Edge shape carried
// forward unchanged so downstream surfaces (systemMessage, JSONL record) see
// edge_type / source_node / target_node / source_path / line_number without
// re-mapping.
type AffectedEdge = navsync.Edge

// Traverse performs a reverse traversal of the M0 navigator graph from a
// changed file path (REQ-NS2-002).
//
// Mapping engine (primary): the changed path is normalized to absolute form
// via filepath.Abs, and for each edge whose SourcePath (also absolute per
// internal/navigator/sync/scan.go — both sides are filepath.Join(projectRoot,
// …)) matches the changed path, Traverse collects the originating edge plus
// BOTH the source_node and the target_node (conservative over-inclusion is
// safe for an advisory layer).
//
// Directory-prefix fallback (REQ-NS2-010): when changedPath ends with a path
// separator (e.g. "/abs/project/internal/foo/"), Traverse matches every edge
// whose SourcePath falls under that prefix. This fallback is INSPIRED by the
// last-segment resolution in navigator-audit.sh heuristic_match() at
// .claude/skills/moai-workflow-project/scripts/navigator-audit.sh:406-422
// (cited as inspiration only; the primary mapping engine is absolute-path
// string equality, NOT the shell script's commit-title/path → design-doc
// name matching).
//
// Per-edge normalization failures are skipped (advisory fail-open at the edge
// level); they do NOT abort the traversal. The function is pure: no I/O, no
// state mutation. Returns an error only for nil graph, empty changedPath, or
// changedPath normalization failure.
func Traverse(graph *navsync.Graph, changedPath string) (*Result, error) {
	if graph == nil {
		return nil, errors.New("navigator/detect: nil graph")
	}
	if changedPath == "" {
		return nil, errors.New("navigator/detect: empty changed path")
	}

	// Detect the directory-prefix signal BEFORE filepath.Abs cleans the
	// trailing separator (filepath.Clean strips it). The root "/" alone is
	// not treated as a prefix match — that would surface the entire graph
	// and is a degenerate case the caller should avoid.
	sep := string(filepath.Separator)
	isPrefix := strings.HasSuffix(changedPath, sep) && changedPath != sep

	changedAbs, err := filepath.Abs(changedPath)
	if err != nil {
		return nil, fmt.Errorf("navigator/detect: normalize changed path %q: %w", changedPath, err)
	}

	// Build a node-key → Node lookup for display-name resolution. The key
	// form is "<entity_type>:<identifier>" (the same form emitted by M0's
	// nodeKey helper at internal/navigator/sync/schema.go:97).
	nodeLookup := make(map[string]navsync.Node, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeLookup[nodeKey(n.EntityType, n.Identifier)] = n
	}

	// filepath.Abs cleaned off the trailing separator, so re-attach it for
	// the HasPrefix check that classifies edge source_paths.
	var prefix string
	if isPrefix {
		prefix = changedAbs + sep
	}

	var edges []AffectedEdge
	var nodes []AffectedNode
	seenEdge := make(map[string]bool, len(graph.Edges))
	seenNode := make(map[string]bool)

	addNode := func(key string) {
		if key == "" || seenNode[key] {
			return
		}
		seenNode[key] = true
		an := AffectedNode{Key: key, DisplayName: key}
		if resolved, ok := nodeLookup[key]; ok {
			an.EntityType = resolved.EntityType
			an.Identifier = resolved.Identifier
			an.DisplayName = resolved.DisplayName
		} else if et, id, ok := splitNodeKey(key); ok {
			// Edge references a node absent from the Node table (malformed
			// graph). Surface it anyway — advisory over-inclusion — with the
			// parsed entity_type/identifier and a display_name fallback.
			an.EntityType = et
			an.Identifier = id
			an.DisplayName = id
		}
		nodes = append(nodes, an)
	}

	for _, edge := range graph.Edges {
		edgeAbs, err := filepath.Abs(edge.SourcePath)
		if err != nil {
			// Per-edge normalization failure: skip this edge only. Do NOT
			// abort the traversal (advisory fail-open at the edge level).
			continue
		}
		var match bool
		if isPrefix {
			match = strings.HasPrefix(edgeAbs, prefix)
		} else {
			match = edgeAbs == changedAbs
		}
		if !match {
			continue
		}
		ek := edgeKey(edge)
		if !seenEdge[ek] {
			seenEdge[ek] = true
			edges = append(edges, edge)
		}
		addNode(edge.SourceNode)
		addNode(edge.TargetNode)
	}

	// Deterministic ordering, mirroring M0's sortEdges
	// (internal/navigator/sync/join.go:370) for edges and the M0 nodeSet
	// sort (entity_type then identifier) for nodes. Byte-stable across runs.
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].EntityType != nodes[j].EntityType {
			return nodes[i].EntityType < nodes[j].EntityType
		}
		return nodes[i].Identifier < nodes[j].Identifier
	})
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].EdgeType != edges[j].EdgeType {
			return edges[i].EdgeType < edges[j].EdgeType
		}
		if edges[i].SourceNode != edges[j].SourceNode {
			return edges[i].SourceNode < edges[j].SourceNode
		}
		if edges[i].TargetNode != edges[j].TargetNode {
			return edges[i].TargetNode < edges[j].TargetNode
		}
		if edges[i].SourcePath != edges[j].SourcePath {
			return edges[i].SourcePath < edges[j].SourcePath
		}
		return edges[i].LineNumber < edges[j].LineNumber
	})

	return &Result{Nodes: nodes, Edges: edges}, nil
}

// nodeKey builds the canonical "<entity_type>:<identifier>" reference form.
// Mirrors the unexported M0 helper at internal/navigator/sync/schema.go:97 —
// re-declared here because the M0 helper is package-private and the Detect
// layer must not mutate the M0 surface to export it.
func nodeKey(t navsync.EntityType, id string) string {
	return string(t) + ":" + id
}

// splitNodeKey parses a "<entity_type>:<identifier>" key back into its parts.
// The identifier MAY contain colons (e.g. a fully-qualified symbol), so split
// only on the FIRST colon.
func splitNodeKey(key string) (navsync.EntityType, string, bool) {
	before, after, found := strings.Cut(key, ":")
	if !found {
		return "", "", false
	}
	return navsync.EntityType(before), after, true
}

// edgeKey is a composite dedup key for an edge. Two edges with identical
// (edge_type, source_node, target_node, source_path, line_number) are the
// same edge for advisory purposes.
func edgeKey(e navsync.Edge) string {
	return fmt.Sprintf("%s|%s|%s|%s:%d", e.EdgeType, e.SourceNode, e.TargetNode, e.SourcePath, e.LineNumber)
}

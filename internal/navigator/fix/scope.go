package fix

import (
	"fmt"
	"sort"
	"strings"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// stale-reason contributor tokens. A path may be stale due to one or more of
// the three input sets; the tokens are OR'd into a composite reason (sorted,
// "+".joined) so the same subtree reached via multiple inputs reads as a single
// deterministic string (e.g. "git-diff+m1-detect").
const (
	reasonGitDiff = "git-diff"
	reasonM1      = "m1-detect"
	reasonM2      = "m2-owner"
)

// ResolveBaseline picks the baseline commit per the priority order
// (REQ-NS5-003 / design.md §C.2):
//  1. explicit compareTo flag (user override);
//  2. nav-graph.json provenance.extract_commit_sha (the default — M0's own
//     provenance, the last known-consistent doc-map state);
//  3. HEAD~1 (degenerate fallback for a fresh checkout with no nav-graph;
//     logged as degraded).
//
// Pure function — the caller provides all three candidate values (the I/O of
// reading nav-graph.json provenance and running `git rev-parse HEAD~1` lives
// in the M3.2 CLI wrapper). Returns (baseline, degraded) where degraded is
// true iff the HEAD~1 fallback path was taken (candidate 3), signalling the
// caller to log a degraded-baseline diagnostic line.
func ResolveBaseline(compareTo, graphProvenanceSHA, headTilde1Fallback string) (baseline string, degraded bool) {
	if compareTo != "" {
		return compareTo, false
	}
	if graphProvenanceSHA != "" {
		return graphProvenanceSHA, false
	}
	return headTilde1Fallback, true
}

// ComputeScope applies the diff-scope formula (REQ-NS5-003, the single source
// of truth — do NOT re-derive a divergent formula):
//
//	diff_scope = (gitDiffPaths ∪ m1ChangedPaths ∪ m2OwnerPaths) ∩ graphBoundPaths
//
// UNION semantics: the three input sets are OR'd, not AND'd. A graph-bound
// path in git-diff alone, M1 alone, or M2 alone each independently seeds a
// stale subtree. The ∩ graphBoundPaths filter is the ONE exclusion: a path
// NOT graph-bound does NOT seed (there is no doc row to fix).
//
// For each touched graph-bound path, the M0 graph is traversed to find the
// bound nodes (edges whose source_path matches); each distinct node defines a
// stale subtree identified by (doc_surface, subtree_id). The output is
// deduplicated by (doc_surface, subtree_id), sorted, and carries a
// stale_reason naming the contributing input set(s) + a work_item_ref when
// the M2 owner-path dimension seeded the entry.
//
// Pure function: no I/O, no side effects, deterministic. The caller owns
// input loading (git diff, detect JSONL, work-items.json, nav-graph.json)
// and fail-open policy (REQ-NS5-009).
//
// @MX:NOTE [AUTO], REQ-NS5-003: diff-scope = UNION(git-diff stale subtrees,
// M1 detect subtrees, M2 owner-path subtrees) ∩ graph-bound nodes. The UNION
// semantics unify three earlier-artifact scopes into one diff-scope; the
// graph-bound filter excludes paths the M0 nav-graph doesn't cover. The
// output feeds draft-request assembly (in scope.go's sibling Run) AND the
// REQ-NS5-013 scope-conformance partition (ConformDraftToScope below).
func ComputeScope(
	gitDiffPaths []string,
	m1ChangedPaths []string,
	m2OwnerPaths []WorkItemRef,
	graph *navsync.Graph,
) []DiffScopeEntry {
	// No graph (M0 not yet run) → no graph-bound paths → empty diff-scope.
	// This is the fail-open degraded case (REQ-NS5-009 row 009c), handled by
	// the caller; ComputeScope itself returns an empty (non-nil) slice.
	if graph == nil || len(graph.Edges) == 0 {
		return []DiffScopeEntry{}
	}

	// 1. Build the graph-bound path set + a path → bound-node-keys index.
	//    graph_bound_paths = { e.source_path : e ∈ graph.Edges } (REQ-NS5-003).
	//    For each bound path, collect BOTH edge endpoints (source_node +
	//    target_node) — a touched path may bind multiple subtrees (e.g. a
	//    doc binding a symbol to a decision). Conservative over-inclusion is
	//    safe: the advisory diff-scope identifies candidates, not certainties.
	boundPaths := make(map[string]bool, len(graph.Edges))
	pathToNodes := make(map[string]map[string]bool, len(graph.Edges))
	nodeTable := make(map[string]navsync.Node, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeTable[nodeKeyStr(n.EntityType, n.Identifier)] = n
	}
	for _, e := range graph.Edges {
		if e.SourcePath == "" {
			continue
		}
		boundPaths[e.SourcePath] = true
		if pathToNodes[e.SourcePath] == nil {
			pathToNodes[e.SourcePath] = make(map[string]bool)
		}
		pathToNodes[e.SourcePath][e.SourceNode] = true
		pathToNodes[e.SourcePath][e.TargetNode] = true
	}

	// 2. Build the UNION of input paths with per-path source tracking.
	//    The three sets are OR'd: each path records which input set(s) named it.
	type pathSrc struct {
		gitDiff bool
		m1      bool
		m2      bool
		m2Ref   WorkItemRef // valid when m2 == true
	}
	union := make(map[string]*pathSrc, len(gitDiffPaths)+len(m1ChangedPaths)+len(m2OwnerPaths))
	ensure := func(p string) *pathSrc {
		if union[p] == nil {
			union[p] = &pathSrc{}
		}
		return union[p]
	}
	for _, p := range gitDiffPaths {
		ensure(p).gitDiff = true
	}
	for _, p := range m1ChangedPaths {
		ensure(p).m1 = true
	}
	for _, ref := range m2OwnerPaths {
		ps := ensure(ref.OwnerPath)
		ps.m2 = true
		ps.m2Ref = ref
	}

	// 3. Apply ∩ graphBoundPaths: keep only graph-bound paths, and for each,
	//    accumulate subtree entries (deduplicated by doc_surface|subtree_id).
	type subtreeAcc struct {
		docSurface string
		subtreeID  string
		gitDiff    bool
		m1         bool
		m2         bool
		m2Refs     []WorkItemRef // all M2 refs touching this subtree (pick smallest for determinism)
	}
	subtrees := make(map[string]*subtreeAcc)
	for path, src := range union {
		if !boundPaths[path] {
			continue // ∩ graphBoundPaths: a non-graph-bound path does NOT seed.
		}
		for nk := range pathToNodes[path] {
			node, ok := nodeTable[nk]
			if !ok {
				// Edge references a node absent from the Node table (malformed
				// graph) — skip this node; advisory over-inclusion does not
				// extend to phantom nodes.
				continue
			}
			ds := docSurfaceFor(node.EntityType)
			key := ds + "|" + node.Identifier
			acc := subtrees[key]
			if acc == nil {
				acc = &subtreeAcc{docSurface: ds, subtreeID: node.Identifier}
				subtrees[key] = acc
			}
			// Merge sources (commutative OR — map iteration order is irrelevant).
			if src.gitDiff {
				acc.gitDiff = true
			}
			if src.m1 {
				acc.m1 = true
			}
			if src.m2 {
				acc.m2 = true
				acc.m2Refs = append(acc.m2Refs, src.m2Ref)
			}
		}
	}

	// 4. Materialize the deduplicated, sorted DiffScopeEntry slice.
	result := make([]DiffScopeEntry, 0, len(subtrees))
	for _, acc := range subtrees {
		entry := DiffScopeEntry{
			DocSurface:  acc.docSurface,
			SubtreeID:   acc.subtreeID,
			StaleReason: buildStaleReason(acc.gitDiff, acc.m1, acc.m2),
		}
		if acc.m2 && len(acc.m2Refs) > 0 {
			// Deterministic ref selection: pick the lexicographically smallest
			// (source_kind, owner_path, action) so two runs on the same inputs
			// produce byte-identical output even when multiple M2 paths bind
			// the same subtree.
			entry.WorkItemRef = smallestM2Ref(acc.m2Refs)
		}
		result = append(result, entry)
	}

	// Sort by (doc_surface, subtree_id) for byte-stable, deterministic output.
	sort.Slice(result, func(i, j int) bool {
		if result[i].DocSurface != result[j].DocSurface {
			return result[i].DocSurface < result[j].DocSurface
		}
		return result[i].SubtreeID < result[j].SubtreeID
	})

	return result
}

// buildStaleReason assembles the stale_reason string from the contributing
// source flags. The tokens are appended in canonical (alphabetical) order and
// "+".joined, so the same subtree reached via multiple inputs always reads as
// the same deterministic string regardless of merge order.
func buildStaleReason(gitDiff, m1, m2 bool) string {
	var parts []string
	if gitDiff {
		parts = append(parts, reasonGitDiff)
	}
	if m1 {
		parts = append(parts, reasonM1)
	}
	if m2 {
		parts = append(parts, reasonM2)
	}
	return strings.Join(parts, "+")
}

// smallestM2Ref returns the lexicographically smallest WorkItemRef by
// (source_kind, owner_path, action) — used for deterministic ref selection
// when multiple M2 paths bind the same subtree.
func smallestM2Ref(refs []WorkItemRef) *WorkItemRef {
	if len(refs) == 0 {
		return nil
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].SourceKind != refs[j].SourceKind {
			return refs[i].SourceKind < refs[j].SourceKind
		}
		if refs[i].OwnerPath != refs[j].OwnerPath {
			return refs[i].OwnerPath < refs[j].OwnerPath
		}
		return refs[i].Action < refs[j].Action
	})
	return &refs[0]
}

// nodeKeyStr builds the canonical "<entity_type>:<identifier>" node reference
// form used by navsync.Edge.SourceNode / TargetNode. Mirrors the unexported
// M0 helper at internal/navigator/sync/schema.go:97 — re-declared here because
// the M0 helper is package-private and the Fix layer must not mutate the M0
// surface to export it.
func nodeKeyStr(et navsync.EntityType, id string) string {
	return string(et) + ":" + id
}

// --- Scope-conformance validation (REQ-NS5-013 / AC-NS5-013) ---
//
// plan.md §C.6: before the gate preview AND before the apply, the draft's
// subtree IDs are validated against diff_scope[]. An over-produced subtree
// (drafted by manager-develop OUTSIDE the stale set — "produce ONLY diff_scope[]
// subtrees" violated) is excluded from the gate preview + the apply + logged
// as a warning. This validates the draft OUTPUT, closing the gap between the
// spawn-prompt instruction and mechanical enforcement.

// ConformDraftToScope partitions draft subtree IDs against the request's
// diff_scope[] allowed-subtree set (REQ-NS5-013). A draft subtree ID present in
// diff_scope[] is in-scope (returned in inScope); an ID NOT in diff_scope[]
// (simulating manager-develop over-production) is excluded (returned in
// excluded). Both outputs are sorted lexicographically and deduplicated for
// byte-stable, deterministic behavior across runs on identical inputs.
//
// Pure function — no I/O, no side effects. The caller owns the warning-logging
// (LogScopeExclusion) and the downstream gate-preview / apply filtering.
//
// @MX:ANCHOR: [AUTO] scope-conformance partition for draft-vs-diff_scope validation
// @MX:REASON: REQ-NS5-013 load-bearing invariant — the draft OUTPUT must conform to diff_scope[], not just the Go engine source; the approval gate + apply both depend on this partition
func ConformDraftToScope(draftSubtreeIDs []string, diffScope []DiffScopeEntry) (inScope, excluded []string) {
	allowed := make(map[string]bool, len(diffScope))
	for _, e := range diffScope {
		allowed[e.SubtreeID] = true
	}
	seenIn := make(map[string]bool, len(draftSubtreeIDs))
	seenOut := make(map[string]bool, len(draftSubtreeIDs))
	for _, id := range draftSubtreeIDs {
		if allowed[id] {
			if !seenIn[id] {
				seenIn[id] = true
				inScope = append(inScope, id)
			}
		} else {
			if !seenOut[id] {
				seenOut[id] = true
				excluded = append(excluded, id)
			}
		}
	}
	sort.Strings(inScope)
	sort.Strings(excluded)
	return inScope, excluded
}

// LogScopeExclusion writes one warning line per excluded subtree ID to the
// navigator-sync log (AC-NS5-013 — "A warning naming the excluded subtree ID +
// the diff_scope[] it is not in is logged to .moai/logs/navigator-sync.log").
// Each line carries the scope-conformance marker + the excluded ID + the
// diff_scope subtree-ID set, so a reader sees exactly what the excluded ID is
// NOT in. Fail-open: a write error is swallowed (reuses logFix). No-op when
// excluded is empty.
func LogScopeExclusion(root string, excludedIDs []string, diffScope []DiffScopeEntry) {
	if len(excludedIDs) == 0 {
		return
	}
	// The diff_scope representation: sorted subtree IDs the excluded ID is not
	// in. Sorted for byte-stable log output across runs.
	scopeIDs := make([]string, 0, len(diffScope))
	for _, e := range diffScope {
		scopeIDs = append(scopeIDs, e.SubtreeID)
	}
	sort.Strings(scopeIDs)
	scopeRepr := "[" + strings.Join(scopeIDs, ",") + "]"
	for _, id := range excludedIDs {
		msg := fmt.Sprintf("navigator-fix: scope-conformance excluded out-of-scope subtree %s (not in diff_scope: %s)", id, scopeRepr)
		logFix(root, msg)
	}
}

package graph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// LoadJSONL reads a persisted edges.jsonl artifact back into memory — the
// reader counterpart to WriteJSONL. Blank lines are skipped; a malformed
// line fails with its 1-based line number.
func LoadJSONL(path string) ([]Edge, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("graph: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }() // read-only: close error carries no signal

	var edges []Edge
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Edge
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("graph: %s:%d: parse edge: %w", path, lineNo, err)
		}
		edges = append(edges, e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("graph: read %s: %w", path, err)
	}
	return edges, nil
}

// FindCallers returns the sorted, deduplicated direct sources of every edge
// targeting node — the reverse neighbors across all kinds: importers of a
// package (import), SPECs depending on a SPEC (spec-depends), and the code
// files tagged @MX:SPEC with a SPEC (mx-spec).
func FindCallers(edges []Edge, node string) []string {
	seen := map[string]bool{}
	for _, e := range edges {
		if e.Target == node {
			seen[e.Source] = true
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// BlastRadius returns the sorted transitive closure of FindCallers: every
// node affected by a change at node, computed as a BFS over reverse edges.
// Direction rule: import and spec-depends edges propagate backwards only
// (the importer/dependent is affected by its dependency, never the
// converse); mx-spec edges propagate both ways — a code file and the SPEC
// it implements affect each other. The start node itself is excluded.
//
// @MX:NOTE: [AUTO] BlastRadius — mx-spec bidirectional rule lets a change at a code file reach the SPECs it implements and their dependents
func BlastRadius(edges []Edge, node string) []string {
	rev := map[string][]string{}
	for _, e := range edges {
		rev[e.Target] = append(rev[e.Target], e.Source)
		if e.Kind == KindMXSpec {
			rev[e.Source] = append(rev[e.Source], e.Target)
		}
	}

	seen := map[string]bool{node: true}
	queue := []string{node}
	var out []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range rev[cur] {
			if seen[next] {
				continue
			}
			seen[next] = true
			out = append(out, next)
			queue = append(queue, next)
		}
	}
	sort.Strings(out)
	return out
}

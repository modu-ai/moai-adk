package dtcg

import "fmt"

// DetectAliasCycle: Detects circular references in alias graph.
// [HARD] Must detect both A→B→C→A and A→A (self-reference).
//
// graph: Map of token path → alias target token path
// Example: {"color.primary": "color.blue-500", "color.blue-500": "color.blue-base"}
//
// Algorithm: DFS + visited set for cycle detection on each node.
// Uses DFS visited approach instead of Floyd's tortoise-and-hare (alias path reporting needed).
func DetectAliasCycle(graph map[string]string) error {
	// Visit state: 0=unvisited, 1=visiting, 2=completed
	state := make(map[string]int, len(graph))

	for node := range graph {
		if state[node] == 0 {
			if cycle := dfsCycle(graph, node, state, nil); cycle != nil {
				return fmt.Errorf("circular alias detected: %v", cycle)
			}
		}
	}

	return nil
}

// dfsCycle: Detects cycles via depth-first search. Returns cycle path (nil if none).
func dfsCycle(graph map[string]string, node string, state map[string]int, path []string) []string {
	// Encounter a node being visited → cycle
	if state[node] == 1 {
		// Find current node position in path and return cycle segment
		for i, n := range path {
			if n == node {
				cycle := make([]string, len(path)-i+1)
				copy(cycle, path[i:])
				cycle[len(cycle)-1] = node
				return cycle
			}
		}
		// Not in path, include only current node
		return append(path, node)
	}
	// Already fully explored → no cycle
	if state[node] == 2 {
		return nil
	}

	// Start visit
	state[node] = 1
	path = append(path, node)

	// Explore adjacent nodes
	if target, ok := graph[node]; ok {
		if cycle := dfsCycle(graph, target, state, path); cycle != nil {
			return cycle
		}
	}

	// Complete visit
	state[node] = 2
	return nil
}

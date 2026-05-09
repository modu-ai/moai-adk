package spec

// dag.go implements Tarjan SCC (Strongly Connected Components) algorithm
// Used to detect cycles in SPEC dependency DAG

// tarjanState is the execution state of Tarjan's SCC algorithm
type tarjanState struct {
	adj      [][]int
	n        int
	index    []int
	lowlink  []int
	onStack  []bool
	stack    []int
	counter  int
	sccs     [][]int // List of node indices for SCCs
}

// SCC returns strongly connected components (size > 1 or size 1 with self-loop)
//
// @MX:NOTE: [AUTO] Tarjan SCC detects cycles with O(V+E) time complexity
func findCyclesTarjan(adj [][]int, n int) [][]int {
	s := &tarjanState{
		adj:     adj,
		n:       n,
		index:   make([]int, n),
		lowlink: make([]int, n),
		onStack: make([]bool, n),
		counter: 1, // 0 is used to indicate "unvisited"
	}

	// -1 is unvisited (using 1 from start since index 0 makes visit status ambiguous)
	for i := range s.index {
		s.index[i] = 0 // 0 = unvisited
	}

	for v := 0; v < n; v++ {
		if s.index[v] == 0 {
			s.strongConnect(v)
		}
	}

	// key is loop (adj[v] contains v)
	var cycles [][]int
	for _, scc := range s.sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
			continue
		}
		// Verify self-loop in SCC
		if len(scc) == 1 {
			v := scc[0]
			for _, w := range adj[v] {
				if w == v {
					cycles = append(cycles, scc)
					break
				}
			}
		}
	}

	return cycles
}

// strongConnect is the core DFS function of Tarjan's SCC algorithm
func (s *tarjanState) strongConnect(v int) {
	s.index[v] = s.counter
	s.lowlink[v] = s.counter
	s.counter++

	s.stack = append(s.stack, v)
	s.onStack[v] = true

	for _, w := range s.adj[v] {
		if s.index[w] == 0 {
			s.strongConnect(w)
			if s.lowlink[w] < s.lowlink[v] {
				s.lowlink[v] = s.lowlink[w]
			}
		} else if s.onStack[w] {
			if s.index[w] < s.lowlink[v] {
				s.lowlink[v] = s.index[w]
			}
		}
	}

	// If v is the root of SCC, pop SCC from stack
	if s.lowlink[v] == s.index[v] {
		var scc []int
		for {
			w := s.stack[len(s.stack)-1]
			s.stack = s.stack[:len(s.stack)-1]
			s.onStack[w] = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		s.sccs = append(s.sccs, scc)
	}
}

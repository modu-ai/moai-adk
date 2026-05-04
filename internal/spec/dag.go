package spec

// dag.go는 Tarjan SCC(Strongly Connected Components) 알고리즘을 구현한다.
// SPEC 의존성 DAG에서 사이클을 탐지하는 데 사용된다.
// O(V+E) 시간 복잡도로 200개 SPEC 범위에서 충분히 빠르다.

// tarjanState는 Tarjan SCC 알고리즘의 실행 상태이다.
type tarjanState struct {
	adj      [][]int
	n        int
	index    []int
	lowlink  []int
	onStack  []bool
	stack    []int
	counter  int
	sccs     [][]int // 각 SCC의 노드 인덱스 목록
}

// findCyclesTarjan은 인접 리스트로 표현된 그래프에서 사이클을 포함하는
// SCC(크기 > 1 또는 자기 루프가 있는 크기 1)를 반환한다.
//
// @MX:NOTE: [AUTO] Tarjan SCC는 O(V+E) 시간 복잡도로 사이클 탐지에 사용된다.
// 반환값은 사이클을 구성하는 노드 인덱스 목록의 슬라이스이다.
func findCyclesTarjan(adj [][]int, n int) [][]int {
	s := &tarjanState{
		adj:     adj,
		n:       n,
		index:   make([]int, n),
		lowlink: make([]int, n),
		onStack: make([]bool, n),
		counter: 1, // 0은 "미방문" 표시로 사용
	}

	// -1은 미방문 (index 0은 방문 여부 판단이 애매하므로 1부터 시작)
	for i := range s.index {
		s.index[i] = 0 // 0 = 미방문
	}

	for v := 0; v < n; v++ {
		if s.index[v] == 0 {
			s.strongConnect(v)
		}
	}

	// 사이클이 있는 SCC만 반환
	// 크기 > 1: 여러 노드로 구성된 사이클
	// 크기 == 1: 자기 자신을 가리키는 루프 (adj[v] contains v)
	var cycles [][]int
	for _, scc := range s.sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
			continue
		}
		// 크기 1 SCC에서 자기 루프 확인
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

// strongConnect는 Tarjan SCC 알고리즘의 핵심 DFS 함수이다.
func (s *tarjanState) strongConnect(v int) {
	s.index[v] = s.counter
	s.lowlink[v] = s.counter
	s.counter++

	s.stack = append(s.stack, v)
	s.onStack[v] = true

	for _, w := range s.adj[v] {
		if s.index[w] == 0 {
			// w가 아직 방문되지 않음
			s.strongConnect(w)
			if s.lowlink[w] < s.lowlink[v] {
				s.lowlink[v] = s.lowlink[w]
			}
		} else if s.onStack[w] {
			// w가 스택에 있음 (현재 SCC의 일부)
			if s.index[w] < s.lowlink[v] {
				s.lowlink[v] = s.index[w]
			}
		}
	}

	// v가 SCC의 루트인 경우 SCC를 스택에서 꺼냄
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

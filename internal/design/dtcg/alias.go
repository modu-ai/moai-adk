package dtcg

import "fmt"

// DetectAliasCycle: 에일리어스 그래프에서 순환 참조를 검출한다.
// [HARD] A→B→C→A 및 A→A(자기 참조) 모두 감지해야 함.
//
// graph: 토큰 경로 → 에일리어스 대상 토큰 경로 맵
// 예: {"color.primary": "color.blue-500", "color.blue-500": "color.blue-base"}
//
// 알고리즘: 각 노드에 대해 DFS + visited 집합으로 순환 감지.
// Floyd's tortoise-and-hare 대신 DFS visited 방식 사용 (에일리어스 경로 보고 필요).
func DetectAliasCycle(graph map[string]string) error {
	// 방문 상태: 0=미방문, 1=방문 중, 2=완료
	state := make(map[string]int, len(graph))

	for node := range graph {
		if state[node] == 0 {
			if cycle := dfsCycle(graph, node, state, nil); cycle != nil {
				return fmt.Errorf("순환 에일리어스 감지: %v", cycle)
			}
		}
	}

	return nil
}

// dfsCycle: 깊이 우선 탐색으로 순환 감지. 순환 경로 반환 (없으면 nil).
func dfsCycle(graph map[string]string, node string, state map[string]int, path []string) []string {
	// 이미 방문 중인 노드를 다시 만남 → 순환
	if state[node] == 1 {
		// path에서 현재 노드 위치 찾아 순환 구간 반환
		for i, n := range path {
			if n == node {
				cycle := make([]string, len(path)-i+1)
				copy(cycle, path[i:])
				cycle[len(cycle)-1] = node
				return cycle
			}
		}
		// path에 없으면 현재 노드만 포함
		return append(path, node)
	}
	// 이미 완전 탐색 완료 → 순환 없음
	if state[node] == 2 {
		return nil
	}

	// 방문 시작
	state[node] = 1
	path = append(path, node)

	// 인접 노드 탐색
	if target, ok := graph[node]; ok {
		if cycle := dfsCycle(graph, target, state, path); cycle != nil {
			return cycle
		}
	}

	// 탐색 완료
	state[node] = 2
	return nil
}

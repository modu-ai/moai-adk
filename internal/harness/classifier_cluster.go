// Package harness — Stage-2 embedding-cluster classifier stub.
// REQ-HRN-CLS-001 / REQ-HRN-CLS-004: 기본 비활성 상태. Wave C에서 SimHash 클러스터링 구현 예정.
package harness

// ClassifierConfig holds configuration for the Stage-2 embedding-cluster classifier.
// Stage2Enabled defaults to false (Go zero value), preserving backward compatibility.
//
// Zero value (ClassifierConfig{}) is always "Stage-2 disabled" —
// Wave C가 새 필드를 추가할 때도 false/zero 기본값을 유지해야 한다 (EC-A5).
//
// @MX:NOTE: [AUTO] Wave A stub. Wave C에서 SimHash 클러스터링 필드 추가 예정.
// @MX:SPEC: REQ-HRN-CLS-001, REQ-HRN-CLS-004
type ClassifierConfig struct {
	// Stage2Enabled은 Stage-2 SimHash 클러스터 분류기를 활성화한다.
	// 기본값 false: Stage-1 byte-identical 경로 유지 (REQ-HRN-CLS-001).
	Stage2Enabled bool
}

// clusterSingletons is the Stage-2 seam. Wave A stub returns patterns unchanged.
// Wave C will implement the full Union-Find Hamming clustering algorithm.
//
// REQ-HRN-CLS-001: cfg.Stage2Enabled == false이면 입력 map을 그대로 반환하고
// auditLogPath에 아무것도 기록하지 않는다.
//
// @MX:NOTE: [AUTO] Wave A stub. Stage-2 클러스터링 로직은 Wave C에서 구현 예정.
// @MX:SPEC: REQ-HRN-CLS-001, REQ-HRN-CLS-004
func clusterSingletons(
	patterns map[string]*Pattern,
	cfg ClassifierConfig,
	auditLogPath string,
) (map[string]*Pattern, error) {
	// auditLogPath는 Wave C의 cluster-merges.jsonl 감사 로그 기록 경로 (REQ-HRN-CLS-013).
	// Wave A stub에서는 호출 시점에 미사용 — Wave C에서 활성화 예정.
	_ = auditLogPath

	if !cfg.Stage2Enabled {
		// Stage-2 비활성: 입력 map 그대로 반환, 감사 로그 미기록 (REQ-HRN-CLS-001).
		return patterns, nil
	}
	// Wave A: Stage-2 활성 분기는 Wave C에서 SimHash + Union-Find 클러스터링으로 확장됨.
	return patterns, nil
}

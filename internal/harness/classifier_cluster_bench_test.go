// Package harness — classifier_cluster_bench_test.go
// Stage-2 클러스터링 성능 벤치마크.
// REQ-HRN-CLS-015: 1000 이벤트 클러스터링이 단일 고루틴에서 1초 이내 완료.
package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// loadPatternsAndEvents는 JSONL 파일을 읽어 패턴 맵과 이벤트 슬라이스를 반환한다.
// testing.TB를 받아 *testing.T와 *testing.B 모두 지원한다.
func loadPatternsAndEvents(tb testing.TB, logPath string) (map[string]*Pattern, []Event, error) {
	tb.Helper()

	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, nil, err
	}

	var events []Event
	for _, line := range splitLines(string(data)) {
		if line == "" {
			continue
		}
		var evt Event
		if jsonErr := json.Unmarshal([]byte(line), &evt); jsonErr != nil {
			return nil, nil, jsonErr
		}
		events = append(events, evt)
	}

	patterns := buildPatternsFromEvents(events)
	return patterns, events, nil
}

// copyPatterns는 패턴 맵을 shallow copy한다. 벤치마크 반복 간 독립성을 위해 사용한다.
func copyPatterns(src map[string]*Pattern) map[string]*Pattern {
	dst := make(map[string]*Pattern, len(src))
	for k, v := range src {
		cp := *v
		dst[k] = &cp
	}
	return dst
}

// BenchmarkClusterSingletons1k는 1000개 이벤트로 Stage-2 클러스터링을 벤치마크한다.
// REQ-HRN-CLS-015: 1초 이내 완료 기준 (go test -bench=. 시 ns/op로 확인).
// 고루틴/채널 미사용 — [HARD] 단일 스레드 순차 처리.
//
// @MX:NOTE: [AUTO] Wave D T-D2: 성능 회귀 감지용 벤치마크. ns/op 상승 = 알고리즘 복잡도 증가 신호.
// @MX:SPEC: REQ-HRN-CLS-015
func BenchmarkClusterSingletons1k(b *testing.B) {
	// 1000 이벤트 fixture 로드 (한 번만)
	fixtureDir := "testdata"
	logPath := filepath.Join(fixtureDir, "stage2_perf_1k.jsonl")

	patterns, events, err := loadPatternsAndEvents(b, logPath)
	if err != nil {
		b.Fatalf("fixture 로드 실패: %v", err)
	}
	if len(patterns) == 0 {
		b.Fatal("빈 패턴 맵 — fixture 확인 필요")
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditPath := filepath.Join(b.TempDir(), "cluster-merges.jsonl")

	b.ResetTimer()
	for range b.N {
		// 각 반복마다 새 패턴 맵을 복사하여 순수 클러스터링 비용만 측정
		patCopy := copyPatterns(patterns)
		_, benchErr := clusterSingletons(patCopy, events, cfg, auditPath)
		if benchErr != nil {
			b.Fatalf("clusterSingletons 오류: %v", benchErr)
		}
	}
}

// BenchmarkClusterSingletons1k_Stage2Off는 Stage-2 비활성(기준선) 성능을 측정한다.
// Stage-2 On/Off 비교를 위한 baseline 벤치마크.
func BenchmarkClusterSingletons1k_Stage2Off(b *testing.B) {
	logPath := filepath.Join("testdata", "stage2_perf_1k.jsonl")

	patterns, events, err := loadPatternsAndEvents(b, logPath)
	if err != nil {
		b.Fatalf("fixture 로드 실패: %v", err)
	}

	cfg := ClassifierConfig{} // Stage2Enabled=false 제로값
	auditPath := filepath.Join(b.TempDir(), "cluster-merges.jsonl")

	b.ResetTimer()
	for range b.N {
		patCopy := copyPatterns(patterns)
		_, benchErr := clusterSingletons(patCopy, events, cfg, auditPath)
		if benchErr != nil {
			b.Fatalf("clusterSingletons 오류: %v", benchErr)
		}
	}
}

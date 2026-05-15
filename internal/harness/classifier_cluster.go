// Package harness — Stage-2 embedding-cluster classifier.
// REQ-HRN-CLS-001 / REQ-HRN-CLS-004: 기본 비활성 상태. Stage2Enabled=true 시 SimHash 클러스터링 실행.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ClassifierConfig는 Stage-2 embedding-cluster 분류기 설정을 담는다.
// Stage2Enabled 기본값은 false(Go 제로값): backward compatibility 보존 (REQ-HRN-CLS-001, EC-A5).
//
// @MX:NOTE: [AUTO] Wave A stub에서 Wave C 전체 구현으로 교체. 제로값 = Stage-2 비활성 유지.
// @MX:SPEC: REQ-HRN-CLS-001, REQ-HRN-CLS-004
type ClassifierConfig struct {
	// Stage2Enabled은 Stage-2 SimHash 클러스터 분류기를 활성화한다.
	// 기본값 false: Stage-1 byte-identical 경로 유지 (REQ-HRN-CLS-001).
	Stage2Enabled bool `yaml:"stage_2_enabled"`

	// SimilarityAlgorithm은 유사도 알고리즘 선택이다. "simhash" 또는 "none"만 허용.
	SimilarityAlgorithm string `yaml:"similarity_algorithm"`

	// HammingThreshold는 클러스터 병합 기준 Hamming 거리다. 기본값 3, 범위 [0, 64].
	HammingThreshold int `yaml:"hamming_threshold"`

	// ClusterMinSize는 병합 대상 최소 클러스터 크기다. 기본값 3, >= 2.
	ClusterMinSize int `yaml:"cluster_min_size"`
}

// Validate는 설정값이 유효 범위 내인지 검증한다.
// 무효 시 에러를 반환하고 clusterSingletons는 Stage-1 fallback으로 처리한다.
//
// @MX:NOTE: [AUTO] REQ-HRN-CLS-018 fallback 계약: Validate 실패 → 입력 map 그대로 반환.
func (c ClassifierConfig) Validate() error {
	if c.HammingThreshold < 0 || c.HammingThreshold > 64 {
		return fmt.Errorf("hamming_threshold %d out of range [0, 64]", c.HammingThreshold)
	}
	if c.ClusterMinSize < 2 {
		return fmt.Errorf("cluster_min_size %d must be >= 2", c.ClusterMinSize)
	}
	if c.SimilarityAlgorithm != "" && c.SimilarityAlgorithm != "simhash" && c.SimilarityAlgorithm != "none" {
		return fmt.Errorf("similarity_algorithm %q not supported (allowed: simhash, none)", c.SimilarityAlgorithm)
	}
	return nil
}

// WithDefaults는 제로값 필드에 기본값을 적용한 새 ClassifierConfig를 반환한다.
func (c ClassifierConfig) WithDefaults() ClassifierConfig {
	if c.SimilarityAlgorithm == "" {
		c.SimilarityAlgorithm = "simhash"
	}
	if c.HammingThreshold == 0 {
		c.HammingThreshold = 3
	}
	if c.ClusterMinSize == 0 {
		c.ClusterMinSize = 3
	}
	return c
}

// clusterAuditLog는 appendClusterMergeAudit에 전달할 감사 데이터 구조체다.
type clusterAuditLog struct {
	Ts               string  `json:"ts"`
	MemberKeys       []string `json:"member_keys"`
	MemberCounts     []int   `json:"member_counts"`
	HammingDistances []int   `json:"hamming_distances"`
	HammingPairCount int     `json:"hamming_pair_count"`
	Truncated        bool    `json:"truncated"`
	MergedKey        string  `json:"merged_key"`
	MergedCount      int     `json:"merged_count"`
	Confidence       float64 `json:"confidence"`
}

// unionFind는 Union-Find(Disjoint Set Union) 자료구조다.
// clusterSingletons에서 Hamming 거리 기반 클러스터 그룹화에 사용된다.
type unionFind struct {
	parent []int
	rank   []int
}

// newUnionFind는 n개 원소의 unionFind를 초기화한다.
func newUnionFind(n int) *unionFind {
	uf := &unionFind{parent: make([]int, n), rank: make([]int, n)}
	for i := range n {
		uf.parent[i] = i
	}
	return uf
}

// find는 경로 압축 find를 수행한다.
func (uf *unionFind) find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.find(uf.parent[x])
	}
	return uf.parent[x]
}

// union은 두 원소를 같은 집합으로 병합한다 (union by rank).
func (uf *unionFind) union(x, y int) {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
}

// singletonEntry는 clusterSingletons에서 singleton 패턴과 대응 이벤트, SimHash를 묶는 구조체다.
type singletonEntry struct {
	key     string
	pattern *Pattern
	event   Event
	hash    uint64
}

// clusterSingletons는 Stage-2 SimHash 클러스터링을 수행한다.
// cfg.Stage2Enabled == false이거나 Validate 실패 시 입력 map을 그대로 반환한다.
// EventType별로 파티셔닝하여 클러스터를 구성하고, ClusterMinSize 이상인 클러스터만 병합한다.
//
// @MX:ANCHOR: [AUTO] Stage-2 클러스터링 기본 진입점.
// @MX:REASON: [AUTO] fan_in >= 3: learner.go(AggregatePatterns), classifier_cluster_test.go, classifier_cluster_audit_test.go
// @MX:WARN: [AUTO] O(s²) Hamming 비교: s는 singleton 수. 대규모 로그에서 성능 병목 가능.
// @MX:REASON: [AUTO] Wave D 벤치마크(BenchmarkClusterSingletons1k)로 p99 ≤ 25ms 검증 예정.
func clusterSingletons(
	patterns map[string]*Pattern,
	events []Event,
	cfg ClassifierConfig,
	auditLogPath string,
) (map[string]*Pattern, error) {
	if !cfg.Stage2Enabled {
		// Stage-2 비활성: 입력 map 그대로 반환 (REQ-HRN-CLS-001)
		return patterns, nil
	}

	// 무효 설정 → stderr 경고 + Stage-1 fallback (REQ-HRN-CLS-018)
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "classifier: invalid config, falling back to stage-1: %v\n", err)
		return patterns, nil
	}

	// "none" 알고리즘 → 클러스터링 없이 반환
	if cfg.SimilarityAlgorithm == "none" {
		return patterns, nil
	}

	// 이벤트 키→Event 맵 구성 (O(n))
	evtByKey := make(map[string]Event, len(events))
	for _, evt := range events {
		k := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)
		evtByKey[k] = evt
	}

	var singletons []singletonEntry
	for key, p := range patterns {
		if p.Count != 1 {
			continue
		}
		evt, ok := evtByKey[key]
		if !ok {
			// 대응 이벤트 없는 singleton: 건너뜀
			continue
		}
		features := buildFeatureString(evt)
		h := SimHash64(features)
		singletons = append(singletons, singletonEntry{key: key, pattern: p, event: evt, hash: h})
	}

	if len(singletons) == 0 {
		return patterns, nil
	}

	// Step 2: EventType별 파티셔닝
	byType := make(map[EventType][]int) // eventType → singletons 인덱스 슬라이스
	for i, s := range singletons {
		byType[s.pattern.EventType] = append(byType[s.pattern.EventType], i)
	}

	// 클러스터 병합 결과 추적: 제거할 키 집합
	removedKeys := make(map[string]struct{})

	// Step 3: EventType 파티션 내 Union-Find 클러스터링
	for _, indices := range byType {
		n := len(indices)
		if n < cfg.ClusterMinSize {
			continue
		}

		uf := newUnionFind(n)

		// O(n²) Hamming 비교: HammingThreshold 이하 → 같은 집합
		for i := range n {
			for j := i + 1; j < n; j++ {
				si := singletons[indices[i]]
				sj := singletons[indices[j]]
				if Hamming(si.hash, sj.hash) <= cfg.HammingThreshold {
					uf.union(i, j)
				}
			}
		}

		// 집합별로 멤버 수집
		groups := make(map[int][]int) // root → 인덱스 집합
		for i := range n {
			root := uf.find(i)
			groups[root] = append(groups[root], i)
		}

		// Step 4: ClusterMinSize 이상인 클러스터만 병합
		for _, members := range groups {
			if len(members) < cfg.ClusterMinSize {
				continue
			}

			// lex-min subject 결정
			memberSingletons := make([]singletonEntry, len(members))
			for i, idx := range members {
				memberSingletons[i] = singletons[indices[idx]]
			}

			sort.Slice(memberSingletons, func(a, b int) bool {
				return memberSingletons[a].pattern.Subject < memberSingletons[b].pattern.Subject
			})
			lexMinSubject := memberSingletons[0].pattern.Subject
			et := memberSingletons[0].pattern.EventType

			// merged_key: "{event_type}:{lex-min-subject}" (REQ-HRN-CLS-005, 2-field)
			mergedKey := fmt.Sprintf("%s:%s", et, lexMinSubject)

			// merged 통계 계산
			totalCount := 0
			totalConf := 0.0
			memberKeys := make([]string, len(memberSingletons))
			memberCounts := make([]int, len(memberSingletons))
			for i, ms := range memberSingletons {
				totalCount += ms.pattern.Count
				totalConf += ms.pattern.Confidence
				memberKeys[i] = ms.key
				memberCounts[i] = ms.pattern.Count
			}
			meanConf := totalConf / float64(len(memberSingletons))

			// Hamming 거리 계산 (upper-triangle, row-major, CAP=20)
			hammingDists := computeHammingDistancesFromEntries(memberSingletons)

			// 감사 로그 기록
			if err := appendClusterMergeAudit(auditLogPath, clusterAuditLog{
				Ts:               time.Now().UTC().Format(time.RFC3339),
				MemberKeys:       memberKeys,
				MemberCounts:     memberCounts,
				HammingDistances: hammingDists,
				HammingPairCount: len(memberSingletons) * (len(memberSingletons) - 1) / 2,
				Truncated:        len(hammingDists) < len(memberSingletons)*(len(memberSingletons)-1)/2,
				MergedKey:        mergedKey,
				MergedCount:      totalCount,
				Confidence:       meanConf,
			}); err != nil {
				// 감사 로그 기록 실패는 non-fatal (경고만)
				fmt.Fprintf(os.Stderr, "classifier: audit log write failed: %v\n", err)
			}

			// pattern map에서 개별 singleton 제거 후 merged 패턴 삽입
			for _, ms := range memberSingletons {
				removedKeys[ms.key] = struct{}{}
			}
			patterns[mergedKey] = &Pattern{
				Key:        mergedKey,
				EventType:  et,
				Subject:    lexMinSubject,
				Count:      totalCount,
				Confidence: meanConf,
			}
		}
	}

	// 병합된 singleton 제거
	for key := range removedKeys {
		delete(patterns, key)
	}

	return patterns, nil
}

// computeHammingDistancesFromEntries는 멤버 슬라이스의 upper-triangle row-major Hamming 거리를 계산한다.
// 최대 20개 요소로 CAP한다 (REQ-HRN-CLS-013).
func computeHammingDistancesFromEntries(members []singletonEntry) []int {
	const maxDist = 20
	n := len(members)
	var dists []int
	for i := range n {
		for j := i + 1; j < n; j++ {
			dists = append(dists, Hamming(members[i].hash, members[j].hash))
			if len(dists) >= maxDist {
				return dists
			}
		}
	}
	return dists
}

// appendClusterMergeAudit는 cluster-merges.jsonl에 한 줄을 append한다.
// 부모 디렉토리가 없으면 자동 생성한다. 기존 파일에 append-only로 기록한다.
//
// @MX:NOTE: [AUTO] REQ-HRN-CLS-013: 감사 로그 대상 경로 및 스키마. cluster-merges.jsonl에 기록.
// @MX:SPEC: REQ-HRN-CLS-013
func appendClusterMergeAudit(path string, entry clusterAuditLog) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("classifier: audit dir 생성 실패: %w", err)
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("classifier: audit 직렬화 실패: %w", err)
	}
	data = append(data, '\n')

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("classifier: audit 파일 열기 실패: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("classifier: audit 파일 쓰기 실패: %w", err)
	}
	return nil
}

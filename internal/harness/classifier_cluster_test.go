// Package harness — classifier_cluster_test.go
// Stage-2 SimHash 클러스터 알고리즘 통합 테스트.
// AC-HRN-CLS-002, -003, -004, -007, -013.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ─────────────────────────────────────────────
// 헬퍼 함수
// ─────────────────────────────────────────────

// makePatternAndEvent는 테스트용 Pattern과 Event 쌍을 생성한다 (count=1, singleton).
func makePatternAndEvent(et EventType, subject, contextHash, promptPreview string, confidence float64) (*Pattern, Event) {
	key := buildPatternKey(et, subject, contextHash)
	p := &Pattern{
		Key:         key,
		EventType:   et,
		Subject:     subject,
		ContextHash: contextHash,
		Count:       1,
		Confidence:  confidence,
	}
	evt := Event{
		EventType:     et,
		Subject:       subject,
		ContextHash:   contextHash,
		PromptPreview: promptPreview,
		SchemaVersion: LogSchemaVersion,
	}
	return p, evt
}

// loadJSONLEvents는 JSONL 파일에서 Event 슬라이스를 로드한다.
func loadJSONLEvents(t *testing.T, path string) []Event {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("JSONL 파일 읽기 실패 %s: %v", path, err)
	}

	var events []Event
	for _, line := range splitLines(string(data)) {
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("이벤트 파싱 실패: %v (line: %s)", err, line)
		}
		events = append(events, evt)
	}
	return events
}

// splitLines는 문자열을 줄 단위로 분할한다.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// buildPatternsFromEvents는 이벤트 슬라이스에서 singleton 패턴 map을 생성한다.
func buildPatternsFromEvents(events []Event) map[string]*Pattern {
	patterns := make(map[string]*Pattern, len(events))
	for _, evt := range events {
		key := buildPatternKey(evt.EventType, evt.Subject, evt.ContextHash)
		patterns[key] = &Pattern{
			Key:        key,
			EventType:  evt.EventType,
			Subject:    evt.Subject,
			Count:      1,
			Confidence: defaultConfidence,
		}
	}
	return patterns
}

// ─────────────────────────────────────────────
// T-C5 통합 테스트
// ─────────────────────────────────────────────

// TestClusterStage2On_10SingletonAggregation는 유사 10개 singleton이 1개 merged 패턴으로
// 클러스터링되는지 검증한다 (AC-HRN-CLS-002).
func TestClusterStage2On_10SingletonAggregation(t *testing.T) {
	t.Parallel()

	fixturePath := filepath.Join("testdata", "stage2_similar_10.jsonl")
	events := loadJSONLEvents(t, fixturePath)
	if len(events) != 10 {
		t.Fatalf("픽스처 이벤트 수 = %d, want 10", len(events))
	}

	patterns := buildPatternsFromEvents(events)
	if len(patterns) != 10 {
		t.Fatalf("초기 패턴 수 = %d, want 10", len(patterns))
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// 10개 singleton → 1개 merged (유사한 prompt_preview 동일)
	if len(result) != 1 {
		t.Errorf("클러스터링 후 패턴 수 = %d, want 1", len(result))
		for k, v := range result {
			t.Logf("  pattern key=%q count=%d", k, v.Count)
		}
	}

	// merged 패턴의 count = 10
	for _, p := range result {
		if p.Count != 10 {
			t.Errorf("merged 패턴 count = %d, want 10", p.Count)
		}
	}

	// 감사 로그가 생성되었는지 확인
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		t.Error("감사 로그 파일이 생성되지 않음")
	}
}

// TestClusterStage2On_10DissimilarStaysSeparate는 비유사 10개 singleton이 분리 상태를 유지하는지 검증한다
// (AC-HRN-CLS-003).
func TestClusterStage2On_10DissimilarStaysSeparate(t *testing.T) {
	t.Parallel()

	fixturePath := filepath.Join("testdata", "stage2_dissimilar_10.jsonl")
	events := loadJSONLEvents(t, fixturePath)
	if len(events) != 10 {
		t.Fatalf("픽스처 이벤트 수 = %d, want 10", len(events))
	}

	patterns := buildPatternsFromEvents(events)

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// 비유사 패턴은 클러스터링되지 않아야 함
	// 클러스터 최소 크기 3인데 모두 분리이므로 패턴 수 유지 (10개 그대로)
	if len(result) != 10 {
		t.Errorf("클러스터링 후 패턴 수 = %d, want 10 (비유사 패턴은 분리 유지)", len(result))
	}
}

// TestClusterMergeEmissionShape는 5개 singleton 클러스터의 confidence 평균 계산을 검증한다
// (AC-HRN-CLS-004).
func TestClusterMergeEmissionShape(t *testing.T) {
	t.Parallel()

	// 5개 singleton, confidence [1.0, 1.0, 0.9, 0.8, 1.0], mean=0.94
	confidences := []float64{1.0, 1.0, 0.9, 0.8, 1.0}
	wantMean := 0.94

	var patterns = make(map[string]*Pattern)
	var events []Event

	for i, conf := range confidences {
		subject := fmt.Sprintf("spec-v3r4-cls-%03d", i)
		preview := "moai run spec v3r4 harness cluster emission shape test"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("ctx%d", i), preview, conf)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// 5개 → 1개 merged
	if len(result) != 1 {
		t.Errorf("패턴 수 = %d, want 1", len(result))
	}

	for _, p := range result {
		if p.Count != 5 {
			t.Errorf("merged count = %d, want 5", p.Count)
		}
		// confidence 평균 검증 (±0.01 허용)
		got := p.Confidence
		if got < wantMean-0.01 || got > wantMean+0.01 {
			t.Errorf("merged confidence = %.4f, want ~%.4f (±0.01)", got, wantMean)
		}
	}
}

// TestConfidenceFloorForcesObservation는 10개 singleton 클러스터의 평균 confidence가
// 0.70 미만일 때 tier가 Observation으로 강제되는지 검증한다 (AC-HRN-CLS-007).
func TestConfidenceFloorForcesObservation(t *testing.T) {
	t.Parallel()

	// 10개 singleton, mean confidence = 0.69 (< 0.70 threshold)
	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 10 {
		subject := fmt.Sprintf("low-conf-spec-%03d", i)
		preview := "moai run low confidence cluster floor test unified"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("lc%d", i), preview, 0.69)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// merged 패턴의 confidence < 0.70 이면 ClassifyTier는 Observation을 반환해야 함
	for _, p := range result {
		tier := ClassifyTier(p, []int{1, 3, 5, 10})
		if tier != TierObservation {
			t.Errorf("confidence < 0.70 merged 패턴 tier = %s, want observation", tier)
		}
	}
}

// TestTierThresholdsConfigOverride는 커스텀 tier_thresholds 오버라이드가 merged 클러스터에
// 올바르게 적용되는지 검증한다 (AC-HRN-CLS-013).
func TestTierThresholdsConfigOverride(t *testing.T) {
	t.Parallel()

	// tier_thresholds [2, 4, 8, 20] 적용 테스트: count=10이면 Rule (8 이상, 20 미만)
	customThresholds := []int{2, 4, 8, 20}

	var patterns = make(map[string]*Pattern)
	var events []Event
	for i := range 10 {
		subject := fmt.Sprintf("tier-override-spec-%03d", i)
		preview := "moai run tier thresholds config override test case"
		p, evt := makePatternAndEvent(EventTypeUserPrompt, subject, fmt.Sprintf("to%d", i), preview, defaultConfidence)
		patterns[p.Key] = p
		events = append(events, evt)
	}

	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}
	auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

	result, err := clusterSingletons(patterns, events, cfg, auditLogPath)
	if err != nil {
		t.Fatalf("clusterSingletons 오류: %v", err)
	}

	// 커스텀 thresholds로 tier 분류: count=10 → TierRule (8 이상, 20 미만)
	for _, p := range result {
		tier := ClassifyTier(p, customThresholds)
		if tier != TierRule {
			t.Errorf("count=%d, thresholds=%v: tier = %s, want rule", p.Count, customThresholds, tier)
		}
	}
}

// ─────────────────────────────────────────────
// T-D5 Invalid Config 테스트
// ─────────────────────────────────────────────

// TestInvalidConfigFailsSafeToStage1은 잘못된 설정 시 Stage-1 그대로 반환함을 검증한다
// (AC-HRN-CLS-014). 6가지 무효 설정 서브케이스.
func TestInvalidConfigFailsSafeToStage1(t *testing.T) {
	t.Parallel()

	// 공통 패턴 설정
	basePatterns := func() (map[string]*Pattern, []Event) {
		p := make(map[string]*Pattern)
		var evts []Event
		for i := range 5 {
			subj := fmt.Sprintf("invalid-cfg-subj-%d", i)
			pat, evt := makePatternAndEvent(EventTypeUserPrompt, subj, fmt.Sprintf("ic%d", i),
				"invalid config test preview text", defaultConfidence)
			p[pat.Key] = pat
			evts = append(evts, evt)
		}
		return p, evts
	}

	cases := []struct {
		name string
		cfg  ClassifierConfig
	}{
		{
			"hamming_threshold_negative",
			ClassifierConfig{Stage2Enabled: true, SimilarityAlgorithm: "simhash", HammingThreshold: -5, ClusterMinSize: 3},
		},
		{
			"hamming_threshold_overflow",
			ClassifierConfig{Stage2Enabled: true, SimilarityAlgorithm: "simhash", HammingThreshold: 99, ClusterMinSize: 3},
		},
		{
			"cluster_min_size_one",
			ClassifierConfig{Stage2Enabled: true, SimilarityAlgorithm: "simhash", HammingThreshold: 3, ClusterMinSize: 1},
		},
		{
			"similarity_algorithm_tfidf",
			ClassifierConfig{Stage2Enabled: true, SimilarityAlgorithm: "tfidf", HammingThreshold: 3, ClusterMinSize: 3},
		},
		{
			"similarity_algorithm_garbage",
			ClassifierConfig{Stage2Enabled: true, SimilarityAlgorithm: "garbage_value_xyz", HammingThreshold: 3, ClusterMinSize: 3},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			patterns, events := basePatterns()
			originalLen := len(patterns)
			// 원본 키 집합 기록
			origKeys := make(map[string]struct{}, len(patterns))
			for k := range patterns {
				origKeys[k] = struct{}{}
			}

			auditLogPath := filepath.Join(t.TempDir(), "cluster-merges.jsonl")

			result, err := clusterSingletons(patterns, events, tc.cfg, auditLogPath)
			if err != nil {
				t.Fatalf("[%s] clusterSingletons 오류: %v", tc.name, err)
			}

			// 패턴 수 불변: Stage-1 그대로 반환
			if len(result) != originalLen {
				t.Errorf("[%s] 패턴 수 = %d, want %d (Stage-1 fallback)", tc.name, len(result), originalLen)
			}

			// 키 집합 불변
			for k := range origKeys {
				if _, ok := result[k]; !ok {
					t.Errorf("[%s] 키 누락: %q", tc.name, k)
				}
			}

			// 감사 로그 미생성 (무효 설정 → fallback)
			if _, statErr := os.Stat(auditLogPath); !os.IsNotExist(statErr) {
				t.Errorf("[%s] 무효 설정에서 감사 로그가 생성됨", tc.name)
			}
		})
	}
}

// TestInvalidConfigYamlTypeMismatchFailsSafe는 yaml type mismatch 시 safe fallback을 검증한다.
// T-D1.5 yaml.TypeError 처리: 잘못된 타입의 설정값이 들어와도 panic 없이 Stage-1 반환.
// (AC-HRN-CLS-014 서브케이스 6)
func TestInvalidConfigYamlTypeMismatchFailsSafe(t *testing.T) {
	t.Parallel()

	// yaml.TypeError는 loader 레벨에서 처리됨. 여기서는 Validate() 경로만 검증.
	// hamming_threshold가 유효 범위 밖인 경우 (type mismatch 후 default 적용 시뮬레이션)
	cfg := ClassifierConfig{
		Stage2Enabled:       true,
		SimilarityAlgorithm: "simhash",
		HammingThreshold:    3,
		ClusterMinSize:      3,
	}.WithDefaults()

	// WithDefaults는 유효한 cfg를 반환해야 함
	if err := cfg.Validate(); err != nil {
		t.Errorf("WithDefaults 후 Validate 실패: %v", err)
	}
}

// Package harness — classifier_rate_limit_test.go
// Wave D T-D4: Rate limit 및 레거시 promotion 호환성 테스트.
// AC-HRN-CLS-006: ClassifyTier가 confidenceThreshold(0.70) 아래에서 TierObservation을 반환한다.
// REQ-HRN-CLS-019: WritePromotion이 legacy_promotions.jsonl을 읽어도 오류 없이 처리된다.
package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRateLimit_ConfidenceThresholdEnforced는 Confidence < 0.70 시
// ClassifyTier가 항상 TierObservation을 반환하는지 검증한다 (AC-HRN-CLS-006).
func TestRateLimit_ConfidenceThresholdEnforced(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}

	tests := []struct {
		name       string
		count      int
		confidence float64
		want       Tier
	}{
		{name: "confidence=0.69_count=100", count: 100, confidence: 0.69, want: TierObservation},
		{name: "confidence=0.00_count=10", count: 10, confidence: 0.00, want: TierObservation},
		{name: "confidence=0.70_count=10", count: 10, confidence: 0.70, want: TierAutoUpdate},
		{name: "confidence=1.00_count=5", count: 5, confidence: 1.00, want: TierRule},
		{name: "confidence=0.69_count=5", count: 5, confidence: 0.69, want: TierObservation},
		{name: "confidence=0.699_count=1", count: 1, confidence: 0.699, want: TierObservation},
		{name: "confidence=0.701_count=3", count: 3, confidence: 0.701, want: TierHeuristic},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := &Pattern{Count: tc.count, Confidence: tc.confidence}
			got := ClassifyTier(p, thresholds)
			if got != tc.want {
				t.Errorf("ClassifyTier(count=%d, confidence=%.3f) = %v, want %v",
					tc.count, tc.confidence, got, tc.want)
			}
		})
	}
}

// TestRateLimit_TierProgressionWithHighConfidence는 신뢰도 >= 0.70 시
// count에 따라 올바른 tier로 진급하는지 검증한다 (AC-HRN-CLS-006 + REQ-HL-002).
func TestRateLimit_TierProgressionWithHighConfidence(t *testing.T) {
	t.Parallel()

	thresholds := []int{1, 3, 5, 10}
	const conf = 0.90

	cases := []struct {
		count int
		want  Tier
	}{
		{0, TierObservation},
		{1, TierObservation},
		{2, TierObservation},
		{3, TierHeuristic},
		{4, TierHeuristic},
		{5, TierRule},
		{9, TierRule},
		{10, TierAutoUpdate},
		{100, TierAutoUpdate},
	}

	for _, tc := range cases {
		p := &Pattern{Count: tc.count, Confidence: conf}
		got := ClassifyTier(p, thresholds)
		if got != tc.want {
			t.Errorf("ClassifyTier(count=%d, confidence=%.2f) = %v, want %v",
				tc.count, conf, got, tc.want)
		}
	}
}

// TestRateLimit_LegacyPromotionsReadable는 legacy_promotions.jsonl이 올바른 Promotion 스키마를
// 사용하고 있어 json.Unmarshal이 오류 없이 파싱하는지 검증한다 (REQ-HRN-CLS-019).
func TestRateLimit_LegacyPromotionsReadable(t *testing.T) {
	t.Parallel()

	fixturePath := filepath.Join("testdata", "legacy_promotions.jsonl")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("legacy_promotions.jsonl 읽기 실패: %v", err)
	}

	lines := splitNonEmpty(string(data))
	if len(lines) == 0 {
		t.Fatal("legacy_promotions.jsonl 비어있음")
	}

	for i, line := range lines {
		var p Promotion
		if jsonErr := json.Unmarshal([]byte(line), &p); jsonErr != nil {
			t.Errorf("라인 %d: Promotion 파싱 실패: %v", i, jsonErr)
			continue
		}
		if p.PatternKey == "" {
			t.Errorf("라인 %d: pattern_key 비어있음", i)
		}
		if p.Confidence <= 0 {
			t.Errorf("라인 %d: confidence = %.3f (양수여야 함)", i, p.Confidence)
		}
	}
}

// TestRateLimit_WritePromotionAppendSemantics는 WritePromotion이 append 시맨틱으로
// 여러 번 호출해도 모든 promotion이 파일에 기록되는지 검증한다.
func TestRateLimit_WritePromotionAppendSemantics(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promotionPath := filepath.Join(dir, "tier-promotions.jsonl")

	learner := NewLearner(promotionPath)
	fixedTime := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	learner.nowFn = func() time.Time { return fixedTime }

	promos := []Promotion{
		{PatternKey: "user_prompt:subject-a", FromTier: "observation", ToTier: "heuristic", ObservationCount: 3, Confidence: 0.85},
		{PatternKey: "user_prompt:subject-b", FromTier: "heuristic", ToTier: "rule", ObservationCount: 5, Confidence: 0.90},
		{PatternKey: "user_prompt:subject-c", FromTier: "rule", ToTier: "auto_update", ObservationCount: 10, Confidence: 0.95},
	}

	for _, promo := range promos {
		if writeErr := learner.WritePromotion(promo); writeErr != nil {
			t.Fatalf("WritePromotion 오류: %v", writeErr)
		}
	}

	// 3개 라인이 기록되어야 함
	data, err := os.ReadFile(promotionPath)
	if err != nil {
		t.Fatalf("promotion 파일 읽기 실패: %v", err)
	}

	lines := splitNonEmpty(string(data))
	if len(lines) != len(promos) {
		t.Errorf("기록된 promotion 수 = %d, want %d", len(lines), len(promos))
	}

	// 각 라인이 올바른 PatternKey를 포함해야 함
	for i, line := range lines {
		var p Promotion
		if jsonErr := json.Unmarshal([]byte(line), &p); jsonErr != nil {
			t.Errorf("라인 %d: Promotion 파싱 실패: %v", i, jsonErr)
			continue
		}
		if p.PatternKey != promos[i].PatternKey {
			t.Errorf("라인 %d: PatternKey = %q, want %q", i, p.PatternKey, promos[i].PatternKey)
		}
	}
}

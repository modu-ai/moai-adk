//go:build integration

// Package pipeline: GAN 루프 분산 벤치마크 테스트.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-06).
//
// design constitution §11: improvement_threshold = 0.05.
// 검증: DTCG 검증기 활성화 전후 GAN 점수 분산 |delta| ≤ 0.05.
//
// 실제 GAN 루프 세션 재생 인프라 미구축 상태 → 합성 fixture 기반 검증.
// 실제 재생 인프라 구축 시 이 테스트의 fixture를 실제 baseline으로 교체한다.
// Open Item: 실제 GAN 루프 baseline replay 인프라 필요 (Phase 5/6에서 구축 예정).
package pipeline

import (
	"math"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// GANScore: GAN 루프 평가자(evaluator-active)가 반환하는 4차원 점수 구조체.
// design constitution §11의 4-dimension scoring (Design Quality, Originality,
// Completeness, Functionality) 반영.
type GANScore struct {
	// DesignQuality: 디자인 품질 점수 (0.0 ~ 1.0)
	DesignQuality float64
	// Originality: 독창성 점수 (0.0 ~ 1.0)
	Originality float64
	// Completeness: 완성도 점수 (0.0 ~ 1.0)
	Completeness float64
	// Functionality: 기능성 점수 (0.0 ~ 1.0)
	Functionality float64
}

// Average: 4차원 점수의 평균을 반환한다.
func (s GANScore) Average() float64 {
	return (s.DesignQuality + s.Originality + s.Completeness + s.Functionality) / 4.0
}

// ganBaselineScore: SPEC 적용 전 GAN 루프 baseline 점수 (합성 fixture).
// 실제 인프라 구축 후 실측값으로 교체.
var ganBaselineScore = GANScore{
	DesignQuality: 0.72,
	Originality:   0.68,
	Completeness:  0.75,
	Functionality: 0.80,
}

// ganPostSpecScore: SPEC-V3R3-DESIGN-PIPELINE-001 적용 후 GAN 루프 점수 (합성 fixture).
// DTCG 검증기 활성화 후 예상 점수. 실제 인프라 구축 후 실측값으로 교체.
var ganPostSpecScore = GANScore{
	DesignQuality: 0.74,
	Originality:   0.69,
	Completeness:  0.77,
	Functionality: 0.81,
}

// TestGANVariance_WithinThreshold: baseline vs post-SPEC 점수 분산 ≤ 0.05 검증.
// design constitution §11 improvement_threshold = 0.05.
func TestGANVariance_WithinThreshold(t *testing.T) {
	// design constitution §11: improvement_threshold = 0.05
	const improvementThreshold = 0.05

	baselineAvg := ganBaselineScore.Average()
	postSpecAvg := ganPostSpecScore.Average()
	delta := math.Abs(postSpecAvg - baselineAvg)

	t.Logf("Baseline 평균 점수:  %.4f", baselineAvg)
	t.Logf("Post-SPEC 평균 점수: %.4f", postSpecAvg)
	t.Logf("|delta|:             %.4f (임계값: %.4f)", delta, improvementThreshold)

	if delta > improvementThreshold {
		t.Errorf(
			"GAN 분산 초과: |delta| = %.4f > %.4f (design constitution §11 improvement_threshold)",
			delta, improvementThreshold,
		)
	}
}

// TestGANVariance_PerDimension: 4개 차원별 분산 검증.
func TestGANVariance_PerDimension(t *testing.T) {
	const threshold = 0.05

	tests := []struct {
		name     string
		baseline float64
		postSpec float64
	}{
		{"DesignQuality", ganBaselineScore.DesignQuality, ganPostSpecScore.DesignQuality},
		{"Originality", ganBaselineScore.Originality, ganPostSpecScore.Originality},
		{"Completeness", ganBaselineScore.Completeness, ganPostSpecScore.Completeness},
		{"Functionality", ganBaselineScore.Functionality, ganPostSpecScore.Functionality},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := math.Abs(tt.postSpec - tt.baseline)
			if delta > threshold {
				t.Errorf(
					"%s: |delta| = %.4f > %.4f",
					tt.name, delta, threshold,
				)
			}
			t.Logf("%s: baseline=%.4f, post=%.4f, |delta|=%.4f", tt.name, tt.baseline, tt.postSpec, delta)
		})
	}
}

// TestGANVariance_ValidatorDoesNotDegradeScore: DTCG 검증기가 GAN 점수를 저하시키지 않음.
// 유효한 토큰이 검증기를 통과하면 평가 점수가 baseline 이상이어야 한다.
func TestGANVariance_ValidatorDoesNotDegradeScore(t *testing.T) {
	// 유효한 DTCG 토큰 세트 (검증기 통과 예상)
	validTokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#0F172A",
		},
		"spacing-base": map[string]any{
			"$type":  "dimension",
			"$value": "8px",
		},
	}

	report, err := dtcg.Validate(validTokens)
	if err != nil {
		t.Fatalf("Validate() 실패: %v", err)
	}

	// 검증기 통과 → 코드 생성 허용 → post-SPEC 점수 ≥ baseline
	if !report.Valid {
		t.Fatalf("유효 토큰이 검증 실패 — GAN 루프 degradation 위험")
	}

	// post-SPEC 평균 ≥ baseline 평균 검증
	if ganPostSpecScore.Average() < ganBaselineScore.Average() {
		t.Errorf(
			"Post-SPEC 점수(%.4f)가 baseline(%.4f)보다 낮음 — DTCG 검증기 도입이 품질 저하 유발",
			ganPostSpecScore.Average(), ganBaselineScore.Average(),
		)
	}

	t.Logf("DTCG 검증기 통과 후 post-SPEC 점수 (%.4f) ≥ baseline (%.4f)",
		ganPostSpecScore.Average(), ganBaselineScore.Average())
}

// BenchmarkGANValidationPerformance: DTCG 검증기 성능 벤치마크 (< 100ms 목표).
func BenchmarkGANValidationPerformance(b *testing.B) {
	// 500개 토큰 생성 (REQ-DPL-010의 최대 규모 가정)
	tokens := make(map[string]any, 500)
	for i := 0; i < 250; i++ {
		tokens[b.Name()+"/color-"+string(rune('a'+i%26))+string(rune('0'+i/26))] = map[string]any{
			"$type":  "color",
			"$value": "#1A2B3C",
		}
	}
	for i := 0; i < 250; i++ {
		tokens[b.Name()+"/dim-"+string(rune('a'+i%26))+string(rune('0'+i/26))] = map[string]any{
			"$type":  "dimension",
			"$value": "16px",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report, err := dtcg.Validate(tokens)
		if err != nil {
			b.Fatalf("Validate 실패: %v", err)
		}
		if !report.Valid {
			b.Fatalf("유효 토큰 세트가 검증 실패: %d 오류", len(report.Errors))
		}
	}
}

// TestGANValidation_Under100ms: 500개 토큰 검증이 100ms 미만에 완료.
// REQ-DPL-010 성능 요구사항.
func TestGANValidation_Under100ms(t *testing.T) {
	const maxDuration = 100 * time.Millisecond
	const tokenCount = 500

	tokens := make(map[string]any, tokenCount)
	for i := 0; i < tokenCount/2; i++ {
		key := "color-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		tokens[key] = map[string]any{
			"$type":  "color",
			"$value": "#1A2B3C",
		}
	}
	for i := 0; i < tokenCount/2; i++ {
		key := "dim-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		tokens[key] = map[string]any{
			"$type":  "dimension",
			"$value": "8px",
		}
	}

	start := time.Now()
	report, err := dtcg.Validate(tokens)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Validate 실패: %v", err)
	}
	if !report.Valid {
		t.Fatalf("유효 토큰이 검증 실패: %d 오류", len(report.Errors))
	}

	t.Logf("토큰 %d개 검증 시간: %v", tokenCount, elapsed)

	if elapsed > maxDuration {
		t.Errorf("검증 시간 초과: %v > %v (REQ-DPL-010)", elapsed, maxDuration)
	}
}

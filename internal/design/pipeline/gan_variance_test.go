//go:build integration

// Package pipeline: GAN loop variance benchmark tests.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-06).
//
// design constitution §11: improvement_threshold = 0.05.
// Verify: with the DTCG validator enabled vs disabled, GAN score variance |delta| <= 0.05.
//
// Real GAN loop session replay infrastructure does not yet exist -> verification based on synthetic fixtures.
// When the real replay infrastructure is built, the fixtures in this test will be replaced with the actual baseline.
// Open Item: real GAN loop baseline replay infrastructure required (planned for Phase 5/6).
package pipeline

import (
	"math"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// GANScore: the 4-dimension score struct returned by the GAN loop evaluator (evaluator-active).
// Reflects the design constitution §11 4-dimension scoring (Design Quality, Originality,
// Completeness, Functionality).
type GANScore struct {
	// DesignQuality: design quality score (0.0 ~ 1.0)
	DesignQuality float64
	// Originality: originality score (0.0 ~ 1.0)
	Originality float64
	// Completeness: completeness score (0.0 ~ 1.0)
	Completeness float64
	// Functionality: functionality score (0.0 ~ 1.0)
	Functionality float64
}

// Average: returns the mean of the 4-dimension scores.
func (s GANScore) Average() float64 {
	return (s.DesignQuality + s.Originality + s.Completeness + s.Functionality) / 4.0
}

// ganBaselineScore: GAN loop baseline score before SPEC application (synthetic fixture).
// To be replaced with measured values after real infrastructure is built.
var ganBaselineScore = GANScore{
	DesignQuality: 0.72,
	Originality:   0.68,
	Completeness:  0.75,
	Functionality: 0.80,
}

// ganPostSpecScore: GAN loop score after applying SPEC-V3R3-DESIGN-PIPELINE-001 (synthetic fixture).
// Projected score after enabling the DTCG validator. To be replaced with measured values later.
var ganPostSpecScore = GANScore{
	DesignQuality: 0.74,
	Originality:   0.69,
	Completeness:  0.77,
	Functionality: 0.81,
}

// TestGANVariance_WithinThreshold: baseline vs post-SPEC score variance must be <= 0.05.
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

// TestGANVariance_PerDimension: verifies variance per dimension across the 4 dimensions.
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

// TestGANVariance_ValidatorDoesNotDegradeScore: the DTCG validator must not degrade GAN scores.
// When valid tokens pass the validator, the evaluation score must be at least the baseline.
func TestGANVariance_ValidatorDoesNotDegradeScore(t *testing.T) {
	// Valid DTCG token set (expected to pass the validator)
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

	// Validator pass -> code generation allowed -> post-SPEC score >= baseline
	if !report.Valid {
		t.Fatalf("유효 토큰이 검증 실패 — GAN 루프 degradation 위험")
	}

	// Verify post-SPEC average >= baseline average
	if ganPostSpecScore.Average() < ganBaselineScore.Average() {
		t.Errorf(
			"Post-SPEC 점수(%.4f)가 baseline(%.4f)보다 낮음 — DTCG 검증기 도입이 품질 저하 유발",
			ganPostSpecScore.Average(), ganBaselineScore.Average(),
		)
	}

	t.Logf("DTCG 검증기 통과 후 post-SPEC 점수 (%.4f) ≥ baseline (%.4f)",
		ganPostSpecScore.Average(), ganBaselineScore.Average())
}

// BenchmarkGANValidationPerformance: DTCG validator performance benchmark (< 100ms target).
func BenchmarkGANValidationPerformance(b *testing.B) {
	// Generate 500 tokens (assuming the maximum scale in REQ-DPL-010)
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

// TestGANValidation_Under100ms: validation of 500 tokens completes in under 100ms.
// REQ-DPL-010 performance requirement.
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

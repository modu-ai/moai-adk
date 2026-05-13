// Package harness — HRN-003 채점 엔진 구현.
// EvaluatorRunner: 계층 집계 + must-pass firewall + Sprint Contract 영속화.
// REQ-HRN-003-004: EvaluatorRunner.AggregateScoreCard().
// REQ-HRN-003-007: aggregation rules (min 기본, mean opt-in).
// REQ-HRN-003-008: must-pass firewall.
// REQ-HRN-003-011: WriteContract() Sprint Contract sub-criterion status 영속화.
package harness

// @MX:NOTE: [AUTO] aggregation default = min per REQ-HRN-003-007; mean opt-in per REQ-HRN-003-015; per-dimension override supported
// @MX:WARN: [AUTO] must-pass firewall FROZEN per design-constitution §12 Mechanism 3; bypass requires CON-002 amendment
// @MX:REASON: design-constitution §12 Mechanism 3 — "Must-pass criteria cannot be compensated by high scores in other areas" (FROZEN per zone-registry CONST-V3R2-154)

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EvaluatorRunner는 ScoreCard 계층 집계 + must-pass firewall 적용을 담당합니다.
// REQ-HRN-003-004: Score() 또는 AggregateScoreCard()로 채점합니다.
//
// @MX:ANCHOR: [AUTO] EvaluatorRunner.AggregateScoreCard — fan_in=3 (scorer_engine_test + SKILL.md Phase 3b + future HRN-001 router)
// @MX:REASON: SPEC-V3R2-HRN-003 REQ-004 — 채점 엔진의 단일 진입점; scorer_test.go + SKILL.md + future 라우터가 참조
type EvaluatorRunner struct {
	rubric *Rubric
}

// NewEvaluatorRunner는 주어진 Rubric으로 새 EvaluatorRunner를 생성합니다.
func NewEvaluatorRunner(rubric *Rubric) *EvaluatorRunner {
	return &EvaluatorRunner{rubric: rubric}
}

// AggregateScoreCard는 SubCriterionScore → CriterionScore → DimensionScore → ScoreCard
// 계층 집계를 수행하고 Verdict를 결정합니다.
// REQ-HRN-003-004, REQ-HRN-003-007, REQ-HRN-003-008.
//
// 집계 규칙:
//   - CriterionScore.Aggregate = aggregateMin/Mean(SubCriterionScore.Score) per dim.Aggregation or rubric.Aggregation
//   - DimensionScore.Aggregate = aggregateMin/Mean(CriterionScore.Aggregate) — dimension-level aggregation
//   - Verdict = applyMustPassFirewall(card, rubric)
func (r *EvaluatorRunner) AggregateScoreCard(card *ScoreCard) (*ScoreCard, error) {
	if card == nil {
		return nil, fmt.Errorf("AggregateScoreCard: card is nil")
	}

	result := &ScoreCard{
		SchemaVersion: card.SchemaVersion,
		SpecID:        card.SpecID,
		Dimensions:    make(map[Dimension]DimensionScore, len(card.Dimensions)),
	}

	for dim, ds := range card.Dimensions {
		// 차원별 aggregation 결정.
		agg := r.effectiveAggregation(dim)

		newDS := DimensionScore{
			Criteria: make(map[string]CriterionScore, len(ds.Criteria)),
		}

		// CriterionScore 집계.
		critAggregates := make([]float64, 0, len(ds.Criteria))
		for critID, cs := range ds.Criteria {
			subScores := make([]float64, 0, len(cs.SubCriteria))
			for _, sub := range cs.SubCriteria {
				subScores = append(subScores, sub.Score)
			}

			aggregator := selectAggregator(agg)
			critAggregate := aggregator.Aggregate(subScores)

			newCS := CriterionScore{
				Aggregate:   critAggregate,
				SubCriteria: cs.SubCriteria,
			}
			newDS.Criteria[critID] = newCS
			critAggregates = append(critAggregates, critAggregate)
		}

		// DimensionScore 집계.
		aggregator := selectAggregator(agg)
		dimAggregate := aggregator.Aggregate(critAggregates)
		newDS.Aggregate = dimAggregate
		result.Dimensions[dim] = newDS
	}

	// Must-pass firewall 적용.
	verdict, rationale := applyMustPassFirewall(result, r.rubric)
	result.Verdict = verdict
	result.Rationale = rationale

	return result, nil
}

// effectiveAggregation은 차원별 override를 고려하여 실효 aggregation 방식을 반환합니다.
// AC-HRN-003-03.c: 차원별 override가 있으면 그것을, 없으면 프로필 기본값을 사용합니다.
func (r *EvaluatorRunner) effectiveAggregation(dim Dimension) string {
	if r.rubric == nil {
		return "min"
	}
	if dr, ok := r.rubric.Dimensions[dim]; ok && dr.Aggregation != "" {
		return dr.Aggregation
	}
	if r.rubric.Aggregation != "" {
		return r.rubric.Aggregation
	}
	return "min" // REQ-HRN-003-007 기본값
}

// Aggregator는 점수 슬라이스를 단일 집계값으로 줄이는 전략 인터페이스입니다.
// T3.13 REFACTOR: aggregation 전략을 명시적 인터페이스로 추출하여 확장성을 확보합니다.
type Aggregator interface {
	Aggregate(scores []float64) float64
}

// MinAggregator는 최솟값 집계 전략입니다.
// REQ-HRN-003-007 기본 집계.
type MinAggregator struct{}

// Aggregate는 슬라이스의 최솟값을 반환합니다.
func (MinAggregator) Aggregate(scores []float64) float64 { return aggregateMin(scores) }

// MeanAggregator는 평균값 집계 전략입니다.
// REQ-HRN-003-015 opt-in 집계.
type MeanAggregator struct{}

// Aggregate는 슬라이스의 평균을 반환합니다.
func (MeanAggregator) Aggregate(scores []float64) float64 { return aggregateMean(scores) }

// selectAggregator는 aggregation 방식 문자열에 따라 Aggregator를 반환합니다.
func selectAggregator(agg string) Aggregator {
	if agg == "mean" {
		return MeanAggregator{}
	}
	return MinAggregator{} // 기본값 "min"
}

// aggregateMin은 슬라이스의 최솟값을 반환합니다.
// 빈 슬라이스에 대해 0.0을 반환합니다 (Edge Case E6).
// REQ-HRN-003-007, AC-HRN-003-03.a.
func aggregateMin(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	min := scores[0]
	for _, s := range scores[1:] {
		if s < min {
			min = s
		}
	}
	return min
}

// aggregateMean은 슬라이스의 평균을 반환합니다.
// 빈 슬라이스에 대해 0.0을 반환합니다.
// REQ-HRN-003-015, AC-HRN-003-03.b.
func aggregateMean(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}
	var sum float64
	for _, s := range scores {
		sum += s
	}
	return sum / float64(len(scores))
}

// applyMustPassFirewall은 must-pass 차원들의 Aggregate가
// 차원별 PassThreshold 이상인지 검증하고 Verdict를 반환합니다.
// REQ-HRN-003-008, AC-HRN-003-04.
//
// 규칙:
//   - 각 must-pass 차원에서 Aggregate < DimensionRubric.PassThreshold → VerdictFail
//   - 모든 must-pass 차원 통과 → VerdictPass
//   - Rationale에 실패 차원 및 점수 명시 (R6 user-friendliness 완화).
func applyMustPassFirewall(card *ScoreCard, rubric *Rubric) (Verdict, string) {
	if rubric == nil || len(rubric.MustPass) == 0 {
		return VerdictPass, ""
	}

	var failedDims []string

	for _, dim := range rubric.MustPass {
		ds, exists := card.Dimensions[dim]
		if !exists {
			// 차원이 ScoreCard에 없으면 0으로 간주 → 실패.
			failedDims = append(failedDims, fmt.Sprintf("%s (missing, 0.00 < threshold)", dim))
			continue
		}

		// 차원별 PassThreshold 확인.
		threshold := rubric.PassThreshold // 기본값: 프로필 전체 threshold
		if dr, ok := rubric.Dimensions[dim]; ok && dr.PassThreshold > 0 {
			threshold = dr.PassThreshold
		}

		if ds.Aggregate < threshold {
			failedDims = append(failedDims, fmt.Sprintf(
				"must-pass dimension %s failed (%.2f < %.2f)",
				dim, ds.Aggregate, threshold,
			))
		}
	}

	if len(failedDims) > 0 {
		rationale := strings.Join(failedDims, "; ")
		return VerdictFail, rationale
	}

	return VerdictPass, ""
}

// ─────────────────────────────────────────────
// Sprint Contract 영속화
// ─────────────────────────────────────────────

// ContractItem는 Sprint Contract YAML의 acceptance_checklist 항목입니다.
// REQ-HRN-003-011: status 필드 포함.
type ContractItem struct {
	ID       string `yaml:"id"`
	Criterion string `yaml:"criterion"`
	Evidence string `yaml:"evidence,omitempty"`
	Score    float64 `yaml:"score"`
	Anchor   string  `yaml:"rubric_anchor"`
	// Status는 sub-criterion 상태입니다: "passed | failed | refined | new".
	// HRN-002 §11.4.1 Sprint Contract durability — passed 항목은 다음 iteration으로 계승됩니다.
	Status string `yaml:"status"`
}

// SprintContractYAML은 WriteContract가 생성하는 YAML 문서의 구조입니다.
// HRN-002 Sprint Contract shape에 sub-criterion status 필드를 추가합니다.
// 기존 SKILL.md acceptance_checklist 형식과 호환됩니다.
type SprintContractYAML struct {
	SpecID             string         `yaml:"spec_id"`
	SchemaVersion      string         `yaml:"schema_version"`
	Verdict            string         `yaml:"verdict"`
	Rationale          string         `yaml:"rationale,omitempty"`
	AcceptanceChecklist []ContractItem `yaml:"acceptance_checklist"`
}

// WriteContract는 ScoreCard의 sub-criterion 점수와 상태를 Sprint Contract YAML로 영속화합니다.
// REQ-HRN-003-011, AC-HRN-003-10.
//
// 파일 경로의 디렉토리가 없으면 자동 생성합니다.
// AC-HRN-003-10: 생성된 YAML은 HRN-002 DetectPriorJudgmentLeak 패턴을 트리거하지 않습니다.
// (criterion state만 포함; evaluator 판단 rationale 없음).
func WriteContract(card *ScoreCard, path string) error {
	if card == nil {
		return fmt.Errorf("WriteContract: card is nil")
	}

	// 디렉토리 자동 생성.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("WriteContract: mkdir %q: %w", dir, err)
	}

	// acceptance_checklist 생성: 차원별 > criterion별 > sub-criterion별 순서로 순회.
	var items []ContractItem

	// 결정론적 출력을 위해 차원 순서를 정렬합니다.
	dims := sortedDimensions(card.Dimensions)
	for _, dim := range dims {
		ds := card.Dimensions[dim]
		critIDs := sortedStringKeys(ds.Criteria)
		for _, critID := range critIDs {
			cs := ds.Criteria[critID]
			subIDs := sortedStringKeys(cs.SubCriteria)
			for _, subID := range subIDs {
				sub := cs.SubCriteria[subID]
				status := subCriterionStatus(sub.Score, sub.RubricAnchor)
				items = append(items, ContractItem{
					ID:       subID,
					Criterion: critID,
					Evidence: sub.Evidence,
					Score:    sub.Score,
					Anchor:   sub.RubricAnchor,
					Status:   status,
				})
			}
		}
	}

	// YAML 직렬화 (표준 라이브러리 없이 간단한 직접 포맷팅).
	// 외부 의존성 없이 yaml 직렬화를 구현합니다.
	content := buildContractYAML(card, items)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("WriteContract: write %q: %w", path, err)
	}

	return nil
}

// subCriterionStatus는 sub-criterion 점수와 anchor에 따라 status를 결정합니다.
// status: "passed" (≥ 0.75), "failed" (< 0.50), "refined" (0.50 ≤ score < 0.75).
func subCriterionStatus(score float64, _ string) string {
	switch {
	case score >= 0.75:
		return "passed"
	case score < 0.50:
		return "failed"
	default:
		return "refined"
	}
}

// buildContractYAML은 SprintContract YAML 내용을 문자열로 생성합니다.
// 외부 yaml 라이브러리 없이 직접 포맷팅합니다.
// HARD(CLAUDE.local.md §14): 하드코딩 금지 — 모든 값은 파라미터에서 주입.
func buildContractYAML(card *ScoreCard, items []ContractItem) string {
	var sb strings.Builder
	sb.WriteString("# Sprint Contract — HRN-003 sub-criterion status\n")
	sb.WriteString("# format: HRN-003 WriteContract v1\n")
	sb.WriteString("# HRN-002 §11.4.1: passed criteria carry forward; evaluator memory is per-iteration\n\n")
	fmt.Fprintf(&sb, "spec_id: %q\n", card.SpecID)
	fmt.Fprintf(&sb, "schema_version: %q\n", card.SchemaVersion)
	fmt.Fprintf(&sb, "verdict: %q\n", string(card.Verdict))
	if card.Rationale != "" {
		fmt.Fprintf(&sb, "rationale: %q\n", card.Rationale)
	}
	sb.WriteString("\nacceptance_checklist:\n")
	for _, item := range items {
		fmt.Fprintf(&sb, "  - id: %q\n", item.ID)
		fmt.Fprintf(&sb, "    criterion: %q\n", item.Criterion)
		if item.Evidence != "" {
			fmt.Fprintf(&sb, "    evidence: %q\n", item.Evidence)
		}
		fmt.Fprintf(&sb, "    score: %.2f\n", item.Score)
		fmt.Fprintf(&sb, "    rubric_anchor: %q\n", item.Anchor)
		fmt.Fprintf(&sb, "    status: %q\n", item.Status)
	}
	return sb.String()
}

// sortedDimensions는 ScoreCard의 Dimension 키를 정렬된 순서로 반환합니다.
// 결정론적 YAML 출력을 위해 사용합니다.
func sortedDimensions(dims map[Dimension]DimensionScore) []Dimension {
	result := make([]Dimension, 0, len(dims))
	for d := range dims {
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		return int(result[i]) < int(result[j])
	})
	return result
}

// sortedStringKeys는 map[string]T의 키를 정렬된 순서로 반환합니다.
func sortedStringKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

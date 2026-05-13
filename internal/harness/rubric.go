// Package harness — HRN-003 Rubric struct, Markdown 파서, 인용 검증.
// REQ-HRN-003-003: 4 anchor level (0.25, 0.50, 0.75, 1.00) FROZEN.
// REQ-HRN-003-005: .md 프로필 파일 파서.
// REQ-HRN-003-009: rubric-anchor 인용 강제.
package harness

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// @MX:NOTE: [AUTO] FROZEN at {0.25, 0.50, 0.75, 1.00} per design-constitution §12 Mechanism 1; adding anchors requires CON-002 amendment
// @MX:WARN: [AUTO] FROZEN-zone constraint per CONST-V3R2-155; anchor set must remain {0.25, 0.50, 0.75, 1.00}
// @MX:REASON: design-constitution §12 Mechanism 1 — "Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0" (FROZEN)

// canonicalAnchors는 rubric anchor의 canonical 집합입니다 (FROZEN).
// REQ-HRN-003-003, REQ-HRN-003-013.
var canonicalAnchors = map[float64]bool{
	0.25: true,
	0.50: true,
	0.75: true,
	1.00: true,
}

// canonicalAnchorStrings는 rubric anchor의 canonical 문자열 집합입니다 (FROZEN).
// REQ-HRN-003-009: ValidateCitation에서 사용합니다.
var canonicalAnchorStrings = map[string]bool{
	"0.25": true,
	"0.50": true,
	"0.75": true,
	"1.00": true,
}

// DimensionRubric는 단일 Dimension에 대한 rubric 설정입니다.
type DimensionRubric struct {
	// Weight는 이 차원의 가중치 (0.0~1.0)입니다.
	Weight float64
	// PassThreshold는 이 차원의 최소 통과 임계값 (≥ 0.60 FROZEN floor)입니다.
	PassThreshold float64
	// Anchors는 anchor score → 설명 맵입니다.
	// canonical: {0.25, 0.50, 0.75, 1.00}
	Anchors map[float64]string
	// Aggregation은 이 차원의 집계 방식입니다 ("min" 또는 "mean").
	// 빈 값이면 프로필 수준의 Rubric.Aggregation을 상속합니다.
	Aggregation string
}

// Rubric는 evaluator profile의 채점 기준 구조체입니다.
// REQ-HRN-003-003: 4 anchor level 포함.
// REQ-HRN-003-005: .md 파일에서 파싱.
type Rubric struct {
	// ProfileName은 프로필 이름입니다 (예: "default", "strict").
	ProfileName string
	// Dimensions는 Dimension → DimensionRubric 맵입니다.
	// 정확히 4개의 차원 (Functionality, Security, Craft, Consistency) 포함.
	Dimensions map[Dimension]DimensionRubric
	// PassThreshold는 전체 통과 임계값 (≥ 0.60 FROZEN floor)입니다.
	PassThreshold float64
	// MustPass는 must-pass 차원 목록입니다.
	// [Security] floor — 이보다 좁은 집합은 허용되지 않습니다 (REQ-HRN-003-018).
	MustPass []Dimension
	// Aggregation은 기본 집계 방식입니다 ("min" 또는 "mean").
	Aggregation string
}

// Validate는 Rubric의 유효성을 검증합니다.
// 다음 조건을 모두 만족해야 합니다:
//   - 정확히 4개의 canonical 차원 (REQ-HRN-003-012)
//   - 각 차원에 정확히 4개의 canonical anchor (REQ-HRN-003-013)
//   - pass_threshold ≥ 0.60 (REQ-HRN-003-014)
//   - aggregation이 {"min", "mean"} 중 하나 (REQ-HRN-003-007)
//   - MustPass가 [Security] floor를 포함 (REQ-HRN-003-018)
func (r *Rubric) Validate() error {
	var errs []config.ValidationError

	// 정확히 4개의 canonical 차원을 가져야 합니다.
	if len(r.Dimensions) != 4 {
		errs = append(errs, config.ValidationError{
			Field:   "Dimensions",
			Message: fmt.Sprintf("must have exactly 4 dimensions, got %d (FROZEN per SPEC-V3R2-HRN-003 REQ-001)", len(r.Dimensions)),
			Value:   len(r.Dimensions),
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// 각 차원이 canonical인지 검증합니다.
	for dim, dr := range r.Dimensions {
		if !dim.IsValid() {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v]", dim),
				Message: fmt.Sprintf("unknown dimension %v; only {Functionality, Security, Craft, Consistency} are valid (FROZEN per REQ-HRN-003-001)", dim),
				Value:   dim,
				Wrapped: config.ErrUnknownDimension,
			})
			continue
		}

		// 각 차원에 정확히 4개의 canonical anchor가 있어야 합니다.
		if err := validateAnchors(dim, dr.Anchors); err != nil {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].Anchors", dim),
				Message: err.Error(),
				Wrapped: config.ErrInvalidConfig,
			})
		}

		// 차원별 pass_threshold ≥ 0.60 검증.
		if dr.PassThreshold < 0.60 {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].PassThreshold", dim),
				Message: fmt.Sprintf("pass_threshold %.2f is below FROZEN floor 0.60 (SPEC-V3R2-HRN-003 REQ-014)", dr.PassThreshold),
				Value:   dr.PassThreshold,
				Wrapped: config.ErrInvalidConfig,
			})
		}

		// 차원별 aggregation 검증.
		if dr.Aggregation != "" && dr.Aggregation != "min" && dr.Aggregation != "mean" {
			errs = append(errs, config.ValidationError{
				Field:   fmt.Sprintf("Dimensions[%v].Aggregation", dim),
				Message: fmt.Sprintf("aggregation %q is not valid; must be one of {min, mean}", dr.Aggregation),
				Value:   dr.Aggregation,
				Wrapped: config.ErrInvalidConfig,
			})
		}
	}

	// 전체 pass_threshold ≥ 0.60 검증.
	if r.PassThreshold < 0.60 {
		errs = append(errs, config.ValidationError{
			Field:   "PassThreshold",
			Message: fmt.Sprintf("pass_threshold %.2f is below FROZEN floor 0.60 (design-constitution §5, HRN-002 REQ-012)", r.PassThreshold),
			Value:   r.PassThreshold,
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// aggregation 검증.
	if r.Aggregation != "min" && r.Aggregation != "mean" {
		errs = append(errs, config.ValidationError{
			Field:   "Aggregation",
			Message: fmt.Sprintf("aggregation %q is not valid; must be one of {min, mean}", r.Aggregation),
			Value:   r.Aggregation,
			Wrapped: config.ErrInvalidConfig,
		})
	}

	// MustPass가 [Security] floor를 포함하는지 검증합니다.
	// REQ-HRN-003-018: [Security]가 floor이며 이보다 좁은 집합은 허용되지 않습니다.
	if err := validateMustPassFloor(r.MustPass); err != nil {
		errs = append(errs, config.ValidationError{
			Field:   "MustPass",
			Message: err.Error(),
			Wrapped: config.ErrMustPassBypassProhibited,
		})
	}

	if len(errs) > 0 {
		return &config.ValidationErrors{Errors: errs}
	}
	return nil
}

// validateAnchors는 anchor 집합이 정확히 canonical {0.25, 0.50, 0.75, 1.00}인지 검증합니다.
func validateAnchors(dim Dimension, anchors map[float64]string) error {
	if len(anchors) != 4 {
		return fmt.Errorf("dimension %v must have exactly 4 anchor levels, got %d (FROZEN per REQ-HRN-003-013)", dim, len(anchors))
	}
	for score := range anchors {
		if !canonicalAnchors[score] {
			return fmt.Errorf("dimension %v has non-canonical anchor %.2f; must be one of {0.25, 0.50, 0.75, 1.00}", dim, score)
		}
	}
	return nil
}

// validateMustPassFloor는 MustPass 집합이 [Security] floor를 포함하는지 검증합니다.
// REQ-HRN-003-018: profiles MAY widen but MAY NOT narrow below [Security].
func validateMustPassFloor(mustPass []Dimension) error {
	hasSecurityFloor := false
	for _, d := range mustPass {
		if d == Security {
			hasSecurityFloor = true
			break
		}
	}
	if !hasSecurityFloor {
		return fmt.Errorf("MustPass set must include Security (floor per OQ3 default + design-constitution §12 Mechanism 3 FROZEN); attempt to narrow below [Security] is prohibited")
	}
	return nil
}

// ValidateCitation는 SubCriterionScore의 RubricAnchor 필드가 canonical anchor 값 중 하나인지 검증합니다.
// REQ-HRN-003-009: empty 또는 non-canonical anchor는 ErrRubricCitationMissing을 반환합니다.
// AC-HRN-003-05.a: 빈 필드 → ErrRubricCitationMissing.
// AC-HRN-003-05.b: non-canonical 값 (예: "0.65") → ErrRubricCitationMissing.
// AC-HRN-003-05.c: canonical 값 → nil.
func (r *Rubric) ValidateCitation(score SubCriterionScore) error {
	if score.RubricAnchor == "" {
		return fmt.Errorf("%w: sub-criterion score missing rubric_anchor field (per design-constitution §12 Mechanism 1)", config.ErrRubricCitationMissing)
	}
	if !canonicalAnchorStrings[score.RubricAnchor] {
		return fmt.Errorf("%w: rubric_anchor %q is not a canonical anchor value; must be one of {\"0.25\", \"0.50\", \"0.75\", \"1.00\"}", config.ErrRubricCitationMissing, score.RubricAnchor)
	}
	return nil
}

// ─────────────────────────────────────────────
// Markdown 프로필 파서
// ─────────────────────────────────────────────

// dimensionNameMap은 프로필 파일의 차원 이름을 Dimension enum으로 매핑합니다.
var dimensionNameMap = map[string]Dimension{
	"Functionality": Functionality,
	"Security":      Security,
	"Craft":         Craft,
	"Consistency":   Consistency,
}

// ParseRubricMarkdown은 .md 형식의 evaluator profile 파일을 파싱하여 Rubric을 반환합니다.
// REQ-HRN-003-005: .moai/config/evaluator-profiles/{name}.md 파일 소비.
// 파서 구조:
//   - H2 "## Evaluation Dimensions" 테이블 → Weight + PassThreshold per dim
//   - H2 "## Must-Pass Criteria" → MustPass slice
//   - H2 "## Scoring Rubric" → H3 per dimension → 2-column score 테이블 → anchor map
//
// 톨러런트: 추가 공백 허용; anchor score 정규화 ("1.0" → "1.00" 변환 후 float64 파싱).
func ParseRubricMarkdown(path string) (*Rubric, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ParseRubricMarkdown: open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// 프로필 이름은 파일명(확장자 제외)에서 추출합니다.
	profileName := extractProfileName(path)

	rubric := &Rubric{
		ProfileName:   profileName,
		PassThreshold: 0.60, // 기본값 (FROZEN floor)
		Aggregation:   "min",
		MustPass:      []Dimension{Functionality, Security}, // 기본값
		Dimensions:    make(map[Dimension]DimensionRubric),
	}

	// 파서 상태 머신.
	type parseSection int
	const (
		sectionNone parseSection = iota
		sectionEvalDimensions
		sectionMustPass
		sectionScoringRubric
	)

	section := sectionNone
	currentRubricDim := Dimension(0)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")

		// H2 섹션 감지.
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")
			switch {
			case strings.Contains(heading, "Evaluation Dimensions"):
				section = sectionEvalDimensions
				currentRubricDim = 0
			case strings.Contains(heading, "Must-Pass Criteria"):
				section = sectionMustPass
				currentRubricDim = 0
			case strings.Contains(heading, "Scoring Rubric"):
				section = sectionScoringRubric
				currentRubricDim = 0
			default:
				section = sectionNone
				currentRubricDim = 0
			}
			continue
		}

		// H3 섹션 감지 (Scoring Rubric 내 차원별 섹션).
		if strings.HasPrefix(line, "### ") && section == sectionScoringRubric {
			heading := strings.TrimPrefix(line, "### ")
			// 차원 이름 추출 (예: "Functionality (40%)" → "Functionality")
			dimName := strings.Fields(heading)[0]
			if dim, ok := dimensionNameMap[dimName]; ok {
				currentRubricDim = dim
				// 차원이 없으면 초기화합니다.
				if _, exists := rubric.Dimensions[dim]; !exists {
					rubric.Dimensions[dim] = DimensionRubric{
						Weight:        0.25, // 기본값
						PassThreshold: 0.60, // FROZEN floor
						Anchors:       make(map[float64]string),
					}
				}
			} else {
				// Scoring Rubric에서 알 수 없는 차원 이름은 건너뜁니다.
				// Evaluation Dimensions 테이블에서의 알 수 없는 차원은 Validate()에서 검증합니다.
				// frontend.md는 "Craft & Functionality", "Design Quality", "Originality" 등의
				// 섹션을 가지는데, 파서는 이를 무시하고 Craft에 UI 관련 내용을 매핑합니다.
				currentRubricDim = 0
			}
			continue
		}

		// 테이블 행 파싱.
		if !strings.HasPrefix(line, "|") {
			continue
		}

		cells := parseTableRow(line)
		if len(cells) < 2 {
			continue
		}

		switch section {
		case sectionEvalDimensions:
			// "| Dimension | Weight | Pass Threshold |" 형식.
			if len(cells) < 3 {
				continue
			}
			dimName := cells[0]
			dim, ok := dimensionNameMap[dimName]
			if !ok {
				// 헤더 행 또는 구분선이면 건너뜁니다.
				continue
			}
			// Weight 파싱 (예: "40%" → 0.40).
			weight := parsePercentOrFloat(cells[1])
			// PassThreshold는 텍스트로 저장 — 숫자가 포함된 경우만 파싱.
			passThreshold := 0.60 // 기본값
			if pct := parsePercentOrFloat(cells[2]); pct > 0 {
				passThreshold = pct
			}

			dr := rubric.Dimensions[dim]
			if dr.Anchors == nil {
				dr.Anchors = make(map[float64]string)
			}
			dr.Weight = weight
			dr.PassThreshold = passThreshold
			rubric.Dimensions[dim] = dr

		case sectionMustPass:
			// "- Functionality: ..." 형식 (리스트 항목).
			// 테이블 행이 아니므로 이 case는 아래의 listItem 파싱에서 처리.

		case sectionScoringRubric:
			// "| Score | Description |" 형식.
			if currentRubricDim == 0 {
				continue
			}
			scoreStr := cells[0]
			description := cells[1]
			if scoreStr == "Score" || strings.Contains(scoreStr, "---") {
				continue // 헤더 또는 구분선
			}
			// anchor score 파싱 ("1.0" → 1.00, "0.25" → 0.25).
			anchor, err := parseAnchorScore(scoreStr)
			if err != nil {
				// 파싱 불가한 행은 건너뜁니다.
				continue
			}
			dr := rubric.Dimensions[currentRubricDim]
			if dr.Anchors == nil {
				dr.Anchors = make(map[float64]string)
			}
			dr.Anchors[anchor] = description
			rubric.Dimensions[currentRubricDim] = dr
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ParseRubricMarkdown: scan %q: %w", path, err)
	}

	// Must-Pass Criteria 재파싱 (리스트 형식).
	// 파일을 다시 열어 Must-Pass 섹션을 파싱합니다.
	mustPass, err := parseMustPassSection(path)
	if err != nil {
		return nil, err
	}
	if len(mustPass) > 0 {
		rubric.MustPass = mustPass
	}

	// 빈 Anchors를 가진 차원에 대해 기본값 초기화.
	// (EvalDimensions 파싱에서 생성된 차원이지만 Rubric 테이블이 없는 경우)
	for dim, dr := range rubric.Dimensions {
		if dr.Anchors == nil {
			dr.Anchors = make(map[float64]string)
			rubric.Dimensions[dim] = dr
		}
	}

	// strict.md 프로필의 경우 PassThreshold를 높게 설정합니다.
	// Rubric.Validate()에서 min 0.60 floor를 보장하므로 여기서는 0 방지만 합니다.
	if rubric.PassThreshold <= 0 {
		rubric.PassThreshold = 0.60
	}

	return rubric, nil
}

// parseMustPassSection은 파일에서 "## Must-Pass Criteria" 섹션의 차원 목록을 파싱합니다.
func parseMustPassSection(path string) ([]Dimension, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("parseMustPassSection: open %q: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var result []Dimension
	inSection := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")

		if strings.HasPrefix(line, "## ") {
			if strings.Contains(line, "Must-Pass Criteria") {
				inSection = true
			} else {
				inSection = false
			}
			continue
		}

		if !inSection {
			continue
		}

		// "- Functionality: ..." 또는 "- Security: ..." 형식의 리스트 항목.
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			content := strings.TrimLeft(line, "-* \t")
			// 첫 번째 단어가 dimension 이름인지 확인합니다.
			parts := strings.SplitN(content, ":", 2)
			dimName := strings.TrimSpace(parts[0])
			if dim, ok := dimensionNameMap[dimName]; ok {
				// 중복 추가 방지.
				found := false
				for _, d := range result {
					if d == dim {
						found = true
						break
					}
				}
				if !found {
					result = append(result, dim)
				}
			}
		}
	}

	return result, scanner.Err()
}

// parseTableRow는 Markdown 테이블 행을 파싱하여 셀 내용 슬라이스를 반환합니다.
// 선행/후행 파이프(|)를 제거하고 각 셀을 trim합니다.
func parseTableRow(line string) []string {
	// 선행/후행 파이프 제거.
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// parsePercentOrFloat은 "40%" 또는 "0.40" 형식의 문자열을 float64로 변환합니다.
func parsePercentOrFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return 0
		}
		return v / 100.0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseAnchorScore는 anchor score 문자열을 float64로 변환합니다.
// "1.0" → 1.00, "0.25" → 0.25 정규화를 포함합니다.
func parseAnchorScore(s string) (float64, error) {
	s = strings.TrimSpace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	// 정규화: float64로 변환 후 canonical 값 검증은 Validate()에서 수행합니다.
	return v, nil
}

// extractProfileName은 파일 경로에서 프로필 이름을 추출합니다.
// 예: ".moai/config/evaluator-profiles/default.md" → "default"
func extractProfileName(path string) string {
	// 마지막 슬래시 이후의 파일명 추출.
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	// 확장자 제거.
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		return filename[:idx]
	}
	return filename
}

// Unwrap is needed for errors.Is to work with ErrUnknownDimension wrapped inside ValidationError.
var _ error = (*config.ValidationErrors)(nil)

package dtcg

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// categoryValidator: 단일 DTCG 카테고리 검증 함수 시그니처.
type categoryValidator func(tokenPath string, value any) error

// categoryValidators: DTCG 2025.10 §8 카테고리 → 검증 함수 매핑.
// 소스: https://tr.designtokens.org/format/ (Editor's Draft 2025-10)
//
// @MX:ANCHOR: [AUTO] Validate 디스패치 핵심 맵 - fan_in 증가 예상
// @MX:REASON: 모든 카테고리 검증이 이 맵을 통해 디스패치됨 (DTCG §8 전체 커버)
var categoryValidators = map[string]categoryValidator{
	"color":        categories.ValidateColor,
	"dimension":    categories.ValidateDimension,
	"fontFamily":   categories.ValidateFontFamily,
	"fontWeight":   categories.ValidateFontWeight,
	"font":         categories.ValidateFont,
	"typography":   categories.ValidateTypography,
	"duration":     categories.ValidateDuration,
	"cubicBezier":  categories.ValidateCubicBezier,
	"number":       categories.ValidateNumber,
	"strokeStyle":  categories.ValidateStrokeStyle,
	"border":       categories.ValidateBorder,
	"transition":   categories.ValidateTransition,
	"shadow":       categories.ValidateShadow,
	"gradient":     categories.ValidateGradient,
}

// Validate: DTCG 2025.10 사양에 따라 토큰 맵 전체를 검증한다.
// REQ-DPL-004, REQ-DPL-010.
//
// tokens: 토큰 키 → 토큰 정의 맵 (각 정의는 $type, $value, $description 포함)
//
// 반환값:
//   - *Report: 검증 결과 (Valid 필드로 성공 여부 확인)
//   - error: 검증 자체를 실행할 수 없는 경우 (예: tokens가 nil이 아닌 잘못된 타입)
//
// 성능: [HARD] 500개 토큰 이하에서 100ms 미만 완료해야 함 (REQ-DPL-010).
func Validate(tokens map[string]any) (*Report, error) {
	report := &Report{
		Valid: true,
	}

	// nil 또는 빈 맵은 유효한 빈 토큰 집합으로 처리
	if len(tokens) == 0 {
		return report, nil
	}

	for key, rawToken := range tokens {
		// 토큰은 map[string]any 형식이어야 함
		tokenDef, ok := rawToken.(map[string]any)
		if !ok {
			// map이 아닌 값은 스킵 (DTCG 그룹 구조 일부일 수 있음)
			continue
		}

		// $type과 $value 모두 없으면 그룹 노드로 스킵
		_, hasType := tokenDef["$type"]
		_, hasValue := tokenDef["$value"]
		if !hasType && !hasValue {
			continue
		}

		// $type 필수 검증
		if !hasType {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  "(unknown)",
				Rule:      "$type 필드 누락 - DTCG 2025.10 §9: $type은 필수",
			})
			continue
		}

		// $value 필수 검증
		if !hasValue {
			category, _ := tokenDef["$type"].(string)
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  category,
				Rule:      "$value 필드 누락 - DTCG 2025.10 §9: $value는 필수",
			})
			continue
		}

		typeVal, ok := tokenDef["$type"].(string)
		if !ok {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  "(invalid)",
				Rule:      fmt.Sprintf("$type은 string이어야 함 (got %T)", tokenDef["$type"]),
			})
			continue
		}

		// 카테고리 유효성 검증
		validator, supported := categoryValidators[typeVal]
		if !supported {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  typeVal,
				Rule:      fmt.Sprintf("알 수 없는 카테고리 '%s' - DTCG 2025.10 §8 참조", typeVal),
				Value:     typeVal,
			})
			continue
		}

		// 카테고리별 값 검증
		if err := validator(key, tokenDef["$value"]); err != nil {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  typeVal,
				Rule:      err.Error(),
				Value:     tokenDef["$value"],
			})
		}
	}

	report.TokenCount = countTokens(tokens)
	return report, nil
}

// countTokens: 토큰 맵에서 실제 토큰 수를 계산한다.
// $type/$value가 있는 항목만 토큰으로 집계.
func countTokens(tokens map[string]any) int {
	count := 0
	for _, raw := range tokens {
		def, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		_, hasType := def["$type"]
		_, hasValue := def["$value"]
		if hasType || hasValue {
			count++
		}
	}
	return count
}

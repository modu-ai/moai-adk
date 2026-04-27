package categories

import (
	"fmt"
	"regexp"
	"strings"
)

// allowedDimensionUnits: DTCG 2025.10 §8.2 허용 단위.
var allowedDimensionUnits = map[string]bool{
	"px":  true,
	"rem": true,
	"em":  true,
	"%":   true,
}

// legacyDimensionPattern: "16px", "1.5rem", "0%" 형식의 레거시 string.
var legacyDimensionPattern = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+(px|rem|em|%)$`)

// ValidateDimension: DTCG 2025.10 §8.2 dimension 카테고리 검증.
// 구조화된 형식: {value: number, unit: "px"|"rem"|"em"|"%"}
// 레거시 string 형식: "16px", "1.5rem", "0%" (하위 호환)
func ValidateDimension(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return validateDimensionMap(tokenPath, v)
	case string:
		return validateDimensionString(tokenPath, v)
	default:
		return fmt.Errorf("토큰 '%s': dimension 값은 map 또는 string이어야 함 (got %T)", tokenPath, value)
	}
}

// validateDimensionMap: 구조화된 {value, unit} 형식 검증.
func validateDimensionMap(tokenPath string, m map[string]any) error {
	// value 필드 검증
	rawValue, hasValue := m["value"]
	if !hasValue {
		return fmt.Errorf("토큰 '%s': dimension map에 'value' 필드 누락", tokenPath)
	}
	if !isNumeric(rawValue) {
		return fmt.Errorf("토큰 '%s': dimension 'value'는 숫자여야 함 (got %T)", tokenPath, rawValue)
	}

	// unit 필드 검증
	unit, ok := m["unit"]
	if !ok {
		return fmt.Errorf("토큰 '%s': dimension map에 'unit' 필드 누락", tokenPath)
	}
	unitStr, ok := unit.(string)
	if !ok {
		return fmt.Errorf("토큰 '%s': dimension 'unit'는 문자열이어야 함", tokenPath)
	}
	if !allowedDimensionUnits[unitStr] {
		return fmt.Errorf("토큰 '%s': dimension unit '%s' 미지원 (허용: px, rem, em, %%)", tokenPath, unitStr)
	}

	return nil
}

// validateDimensionString: 레거시 "16px" 형식 검증.
func validateDimensionString(tokenPath, s string) error {
	if s == "" {
		return fmt.Errorf("토큰 '%s': dimension 값이 빈 문자열", tokenPath)
	}
	// "0" 같은 단위 없는 순수 숫자는 레거시 spec에서도 거부
	if !legacyDimensionPattern.MatchString(s) {
		// 단위만 있는 경우 (예: "px") 또는 잘못된 단위 처리
		return fmt.Errorf("토큰 '%s': dimension string '%s' 잘못된 형식 (허용: 숫자+단위, 예: 16px, 1.5rem, 100%%)", tokenPath, s)
	}

	// 단위 추출하여 허용 목록 확인
	unit := extractUnit(s)
	if !allowedDimensionUnits[unit] {
		return fmt.Errorf("토큰 '%s': dimension unit '%s' 미지원 (허용: px, rem, em, %%)", tokenPath, unit)
	}

	return nil
}

// extractUnit: "16px" → "px", "1.5rem" → "rem", "100%" → "%".
func extractUnit(s string) string {
	for _, u := range []string{"rem", "em", "px", "%"} {
		if strings.HasSuffix(s, u) {
			return u
		}
	}
	return ""
}

// isNumeric: 값이 숫자형(float64, int, float32 등)인지 확인.
func isNumeric(v any) bool {
	switch v.(type) {
	case float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

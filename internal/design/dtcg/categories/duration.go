package categories

import (
	"fmt"
	"regexp"
)

// allowedDurationUnits: DTCG 2025.10 duration 허용 단위.
var allowedDurationUnits = map[string]bool{
	"ms": true,
	"s":  true,
}

// legacyDurationPattern: "300ms", "0.3s" 형식의 레거시 string.
var legacyDurationPattern = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+(ms|s)$`)

// ValidateDuration: DTCG 2025.10 duration 카테고리 검증.
// 구조화된 형식: {value: number, unit: "ms"|"s"}
// 레거시 string: "300ms", "0.3s"
func ValidateDuration(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return validateDurationMap(tokenPath, v)
	case string:
		return validateDurationString(tokenPath, v)
	default:
		return fmt.Errorf("토큰 '%s': duration 값은 map 또는 string이어야 함 (got %T)", tokenPath, value)
	}
}

// validateDurationMap: 구조화된 {value, unit} 형식 검증.
func validateDurationMap(tokenPath string, m map[string]any) error {
	rawValue, hasValue := m["value"]
	if !hasValue {
		return fmt.Errorf("토큰 '%s': duration map에 'value' 필드 누락", tokenPath)
	}
	if !isNumeric(rawValue) {
		return fmt.Errorf("토큰 '%s': duration 'value'는 숫자여야 함 (got %T)", tokenPath, rawValue)
	}

	unit, ok := m["unit"]
	if !ok {
		return fmt.Errorf("토큰 '%s': duration map에 'unit' 필드 누락", tokenPath)
	}
	unitStr, ok := unit.(string)
	if !ok || !allowedDurationUnits[unitStr] {
		return fmt.Errorf("토큰 '%s': duration unit '%v' 미지원 (허용: ms, s)", tokenPath, unit)
	}

	return nil
}

// validateDurationString: 레거시 "300ms" 형식 검증.
func validateDurationString(tokenPath, s string) error {
	if !legacyDurationPattern.MatchString(s) {
		return fmt.Errorf("토큰 '%s': duration string '%s' 잘못된 형식 (허용: 숫자+단위, 예: 300ms, 0.3s)", tokenPath, s)
	}
	return nil
}

package categories

import "fmt"

// ValidateNumber: DTCG 2025.10 number 카테고리 검증.
// 허용: 숫자 원시값 (float64, int, float32 등).
// 문자열, bool, nil, map 등은 거부.
func ValidateNumber(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	if isNumeric(value) {
		return nil
	}
	return fmt.Errorf("토큰 '%s': number 값은 숫자여야 함 (got %T)", tokenPath, value)
}

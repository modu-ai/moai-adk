package categories

import "fmt"

// ValidateFontFamily: DTCG 2025.10 fontFamily 카테고리 검증.
// 허용 형식: 단일 string 또는 string 배열 (폴백 체인).
func ValidateFontFamily(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("토큰 '%s': fontFamily 값이 빈 문자열", tokenPath)
		}
		return nil
	case []any:
		if len(v) == 0 {
			return fmt.Errorf("토큰 '%s': fontFamily 배열이 비어있음", tokenPath)
		}
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				return fmt.Errorf("토큰 '%s': fontFamily 배열 %d번째 원소가 문자열이 아님 (got %T)", tokenPath, i, item)
			}
			if s == "" {
				return fmt.Errorf("토큰 '%s': fontFamily 배열 %d번째 원소가 빈 문자열", tokenPath, i)
			}
		}
		return nil
	default:
		return fmt.Errorf("토큰 '%s': fontFamily 값은 string 또는 string 배열이어야 함 (got %T)", tokenPath, value)
	}
}

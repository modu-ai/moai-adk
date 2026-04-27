package categories

import "fmt"

// namedFontWeights: DTCG 2025.10 fontWeight 허용 named 값.
var namedFontWeights = map[string]bool{
	"thin":    true,
	"light":   true,
	"normal":  true,
	"regular": true,
	"medium":  true,
	"bold":    true,
	"heavy":   true,
}

// ValidateFontWeight: DTCG 2025.10 fontWeight 카테고리 검증.
// 허용: 100-900 수치(100 단위) 또는 named("thin","light","normal","regular","medium","bold","heavy").
func ValidateFontWeight(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("토큰 '%s': fontWeight 값이 빈 문자열", tokenPath)
		}
		if !namedFontWeights[v] {
			return fmt.Errorf("토큰 '%s': fontWeight named 값 '%s' 미지원 (허용: thin, light, normal, regular, medium, bold, heavy)", tokenPath, v)
		}
		return nil
	case float64:
		return validateNumericFontWeight(tokenPath, v)
	case int:
		return validateNumericFontWeight(tokenPath, float64(v))
	case float32:
		return validateNumericFontWeight(tokenPath, float64(v))
	default:
		return fmt.Errorf("토큰 '%s': fontWeight 값은 숫자(100-900) 또는 named string이어야 함 (got %T)", tokenPath, value)
	}
}

// validateNumericFontWeight: 수치 폰트 두께 검증 (100-900, 100 단위).
func validateNumericFontWeight(tokenPath string, v float64) error {
	if v < 100 || v > 900 {
		return fmt.Errorf("토큰 '%s': fontWeight 수치 %g 범위 초과 (허용: 100-900)", tokenPath, v)
	}
	// 100 단위 검증
	if int(v)%100 != 0 {
		return fmt.Errorf("토큰 '%s': fontWeight 수치 %g는 100 단위여야 함", tokenPath, v)
	}
	return nil
}

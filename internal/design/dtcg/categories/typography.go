package categories

import "fmt"

// ValidateTypography: DTCG 2025.10 typography 복합 카테고리 검증.
// font 카테고리의 상위 집합: 추가로 letterSpacing, textDecoration, textTransform 지원.
// 필수: family, size, weight / 선택: style, lineHeight, letterSpacing, textDecoration, textTransform.
func ValidateTypography(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': typography 값은 map이어야 함 (got %T)", tokenPath, value)
	}

	// font 카테고리와 동일한 필수 필드 검증 (family, size, weight)
	if err := validateFontField(tokenPath, m, "family", func(v any) error {
		return ValidateFontFamily(tokenPath+".family", v)
	}); err != nil {
		return err
	}

	if err := validateFontField(tokenPath, m, "size", func(v any) error {
		return ValidateDimension(tokenPath+".size", v)
	}); err != nil {
		return err
	}

	if err := validateFontField(tokenPath, m, "weight", func(v any) error {
		return ValidateFontWeight(tokenPath+".weight", v)
	}); err != nil {
		return err
	}

	// typography 확장 선택 필드 - 존재하면 string이어야 함
	optionalStringFields := []string{"style", "lineHeight", "letterSpacing", "textDecoration", "textTransform"}
	for _, field := range optionalStringFields {
		if v, exists := m[field]; exists {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("토큰 '%s': typography '%s' 필드는 string이어야 함 (got %T)", tokenPath, field, v)
			}
		}
	}

	return nil
}

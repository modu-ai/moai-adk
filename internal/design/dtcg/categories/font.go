package categories

import "fmt"

// ValidateFont: DTCG 2025.10 font 복합 카테고리 검증.
// 필수: family, size, weight / 선택: style, lineHeight.
// 각 필드는 직접 값 또는 에일리어스 참조({token.path}) 허용.
func ValidateFont(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': font 값은 map이어야 함 (got %T)", tokenPath, value)
	}

	// family 필드 검증
	if err := validateFontField(tokenPath, m, "family", func(v any) error {
		return ValidateFontFamily(tokenPath+".family", v)
	}); err != nil {
		return err
	}

	// size 필드 검증 (dimension 타입)
	if err := validateFontField(tokenPath, m, "size", func(v any) error {
		return ValidateDimension(tokenPath+".size", v)
	}); err != nil {
		return err
	}

	// weight 필드 검증 (fontWeight 타입)
	if err := validateFontField(tokenPath, m, "weight", func(v any) error {
		return ValidateFontWeight(tokenPath+".weight", v)
	}); err != nil {
		return err
	}

	// 선택 필드 (style, lineHeight) - 존재하면 string 또는 에일리어스
	for _, optField := range []string{"style", "lineHeight"} {
		if v, exists := m[optField]; exists {
			if s, ok := v.(string); !ok || (s == "" && !IsAlias(s)) {
				if !ok {
					return fmt.Errorf("토큰 '%s': font '%s' 필드는 string이어야 함 (got %T)", tokenPath, optField, v)
				}
			}
		}
	}

	return nil
}

// validateFontField: font 복합 토큰의 단일 필드 검증 헬퍼.
// 에일리어스 참조인 경우 검증 통과; 없으면 필수 필드 누락 오류.
func validateFontField(tokenPath string, m map[string]any, field string, validate func(any) error) error {
	v, ok := m[field]
	if !ok {
		return fmt.Errorf("토큰 '%s': font 복합 토큰에 '%s' 필드 누락", tokenPath, field)
	}
	// 에일리어스 참조는 통과
	if s, isStr := v.(string); isStr && IsAlias(s) {
		return nil
	}
	return validate(v)
}

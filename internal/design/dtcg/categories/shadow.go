package categories

import "fmt"

// shadowRequiredFields: shadow 복합 토큰의 필수 필드 목록.
var shadowRequiredFields = []string{"color", "offsetX", "offsetY", "blur", "spread"}

// ValidateShadow: DTCG 2025.10 shadow 복합 카테고리 검증.
// 허용: 단일 {color, offsetX, offsetY, blur, spread, inset?} 또는 다층 배열.
func ValidateShadow(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		// 단일 그림자
		return validateShadowLayer(tokenPath, v)
	case []any:
		// 다층 그림자 배열
		if len(v) == 0 {
			return fmt.Errorf("토큰 '%s': shadow 배열이 비어있음", tokenPath)
		}
		for i, item := range v {
			layer, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("토큰 '%s': shadow 배열 %d번째 원소가 map이 아님 (got %T)", tokenPath, i, item)
			}
			if err := validateShadowLayer(fmt.Sprintf("%s[%d]", tokenPath, i), layer); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("토큰 '%s': shadow 값은 map 또는 배열이어야 함 (got %T)", tokenPath, value)
	}
}

// validateShadowLayer: 단일 그림자 레이어 {color, offsetX, offsetY, blur, spread, inset?} 검증.
func validateShadowLayer(tokenPath string, m map[string]any) error {
	// 필수 필드 존재 확인 및 검증
	for _, field := range shadowRequiredFields {
		v, ok := m[field]
		if !ok {
			return fmt.Errorf("토큰 '%s': shadow '%s' 필드 누락", tokenPath, field)
		}

		// 에일리어스 참조 통과
		if s, isStr := v.(string); isStr && IsAlias(s) {
			continue
		}

		// color 필드는 ValidateColor로 검증
		if field == "color" {
			if err := ValidateColor(tokenPath+".color", v); err != nil {
				return err
			}
			continue
		}

		// offsetX, offsetY, blur, spread는 dimension 검증
		if err := ValidateDimension(tokenPath+"."+field, v); err != nil {
			return err
		}
	}

	// inset 선택 필드 - bool이어야 함
	if inset, exists := m["inset"]; exists {
		if _, ok := inset.(bool); !ok {
			return fmt.Errorf("토큰 '%s': shadow 'inset' 필드는 bool이어야 함 (got %T)", tokenPath, inset)
		}
	}

	return nil
}

// ValidateBorder: DTCG 2025.10 border 복합 카테고리 검증.
// {color: color, width: dimension, style: strokeStyle}.
func ValidateBorder(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': border 값은 map이어야 함 (got %T)", tokenPath, value)
	}

	// color 필드 검증
	colorVal, ok := m["color"]
	if !ok {
		return fmt.Errorf("토큰 '%s': border 'color' 필드 누락", tokenPath)
	}
	if s, isStr := colorVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateColor(tokenPath+".color", colorVal); err != nil {
			return err
		}
	}

	// width 필드 검증 (dimension)
	widthVal, ok := m["width"]
	if !ok {
		return fmt.Errorf("토큰 '%s': border 'width' 필드 누락", tokenPath)
	}
	if s, isStr := widthVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateDimension(tokenPath+".width", widthVal); err != nil {
			return err
		}
	}

	// style 필드 검증 (strokeStyle)
	styleVal, ok := m["style"]
	if !ok {
		return fmt.Errorf("토큰 '%s': border 'style' 필드 누락", tokenPath)
	}
	if s, isStr := styleVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateStrokeStyle(tokenPath+".style", styleVal); err != nil {
			return err
		}
	}

	return nil
}

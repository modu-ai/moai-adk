package categories

import "fmt"

// ValidateGradient: DTCG 2025.10 gradient 카테고리 검증.
// 형식: 최소 2개 stop의 배열: [{color, position: 0..1}, ...]
func ValidateGradient(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': gradient 값은 배열이어야 함 (got %T)", tokenPath, value)
	}

	// 최소 2개 stop 필요
	if len(arr) < 2 {
		return fmt.Errorf("토큰 '%s': gradient는 최소 2개 color stop이 필요함 (got %d)", tokenPath, len(arr))
	}

	for i, item := range arr {
		stop, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("토큰 '%s': gradient stop[%d]이 map이 아님 (got %T)", tokenPath, i, item)
		}
		if err := validateGradientStop(tokenPath, i, stop); err != nil {
			return err
		}
	}

	return nil
}

// validateGradientStop: 그라디언트 stop {color, position} 검증.
func validateGradientStop(tokenPath string, idx int, stop map[string]any) error {
	stopPath := fmt.Sprintf("%s.stop[%d]", tokenPath, idx)

	// color 필드 검증
	colorVal, ok := stop["color"]
	if !ok {
		return fmt.Errorf("토큰 '%s': 'color' 필드 누락", stopPath)
	}
	if s, isStr := colorVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateColor(stopPath+".color", colorVal); err != nil {
			return err
		}
	}

	// position 필드 검증 (0..1 범위 숫자)
	posVal, ok := stop["position"]
	if !ok {
		return fmt.Errorf("토큰 '%s': 'position' 필드 누락", stopPath)
	}
	pos, err := toFloat64(posVal)
	if err != nil {
		return fmt.Errorf("토큰 '%s': 'position' 숫자여야 함 (got %T)", stopPath, posVal)
	}
	if pos < 0 || pos > 1 {
		return fmt.Errorf("토큰 '%s': 'position' %g ∈ [0, 1] 범위 초과", stopPath, pos)
	}

	return nil
}

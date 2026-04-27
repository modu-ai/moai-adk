package categories

import "fmt"

// ValidateCubicBezier: DTCG 2025.10 cubicBezier 카테고리 검증.
// 형식: [x1, y1, x2, y2] 4원소 배열
// x1, x2 ∈ [0, 1] (제약); y1, y2 제한 없음.
func ValidateCubicBezier(tokenPath string, value any) error {
	// 에일리어스 참조 통과
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("토큰 '%s': cubicBezier 값은 4원소 배열이어야 함 (got %T)", tokenPath, value)
	}

	if len(arr) != 4 {
		return fmt.Errorf("토큰 '%s': cubicBezier 배열은 정확히 4개 원소여야 함 (got %d)", tokenPath, len(arr))
	}

	// 각 원소를 float64로 변환
	coords := make([]float64, 4)
	for i, item := range arr {
		f, err := toFloat64(item)
		if err != nil {
			return fmt.Errorf("토큰 '%s': cubicBezier[%d] 숫자여야 함 (got %T)", tokenPath, i, item)
		}
		coords[i] = f
	}

	// x1, x2 (인덱스 0, 2) ∈ [0, 1] 검증
	if coords[0] < 0 || coords[0] > 1 {
		return fmt.Errorf("토큰 '%s': cubicBezier x1=%g ∈ [0,1] 범위 초과", tokenPath, coords[0])
	}
	if coords[2] < 0 || coords[2] > 1 {
		return fmt.Errorf("토큰 '%s': cubicBezier x2=%g ∈ [0,1] 범위 초과", tokenPath, coords[2])
	}

	return nil
}

// toFloat64: any → float64 변환 (숫자 타입만 허용).
func toFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case int32:
		return float64(n), nil
	}
	return 0, fmt.Errorf("숫자 타입이 아님: %T", v)
}

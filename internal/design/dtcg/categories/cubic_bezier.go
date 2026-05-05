package categories

import "fmt"

// ValidateCubicBezier: DTCG 2025.10 cubicBezier category validation.
// Format: [x1, y1, x2, y2] 4-element array
// x1, x2 ∈ [0, 1] (constrained); y1, y2 unlimited.
func ValidateCubicBezier(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("token '%s': cubicBezier value must be a 4-element array (got %T)", tokenPath, value)
	}

	if len(arr) != 4 {
		return fmt.Errorf("token '%s': cubicBezier array must have exactly 4 elements (got %d)", tokenPath, len(arr))
	}

	// Convert each element to float64
	coords := make([]float64, 4)
	for i, item := range arr {
		f, err := toFloat64(item)
		if err != nil {
			return fmt.Errorf("token '%s': cubicBezier[%d] must be a number (got %T)", tokenPath, i, item)
		}
		coords[i] = f
	}

	// x1, x2 (indices 0, 2) ∈ [0, 1] validation
	if coords[0] < 0 || coords[0] > 1 {
		return fmt.Errorf("token '%s': cubicBezier x1=%g out of [0,1] range", tokenPath, coords[0])
	}
	if coords[2] < 0 || coords[2] > 1 {
		return fmt.Errorf("token '%s': cubicBezier x2=%g out of [0,1] range", tokenPath, coords[2])
	}

	return nil
}

// toFloat64: any → float64 conversion (numeric types only).
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
	return 0, fmt.Errorf("not a numeric type: %T", v)
}

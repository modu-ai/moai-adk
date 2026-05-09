package categories

import "fmt"

// ValidateGradient: DTCG 2025.10 gradient category validation.
// Format: array of minimum 2 stops: [{color, position: 0..1}, ...]
func ValidateGradient(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	arr, ok := value.([]any)
	if !ok {
		return fmt.Errorf("token '%s': gradient value must be array (got %T)", tokenPath, value)
	}

	// Minimum 2 stops required
	if len(arr) < 2 {
		return fmt.Errorf("token '%s': gradient requires minimum 2 color stops (got %d)", tokenPath, len(arr))
	}

	for i, item := range arr {
		stop, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("token '%s': gradient stop[%d] is not map (got %T)", tokenPath, i, item)
		}
		if err := validateGradientStop(tokenPath, i, stop); err != nil {
			return err
		}
	}

	return nil
}

// validateGradientStop: Gradient stop {color, position} validation.
func validateGradientStop(tokenPath string, idx int, stop map[string]any) error {
	stopPath := fmt.Sprintf("%s.stop[%d]", tokenPath, idx)

	// color field validation
	colorVal, ok := stop["color"]
	if !ok {
		return fmt.Errorf("token '%s': missing 'color' field", stopPath)
	}
	if s, isStr := colorVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateColor(stopPath+".color", colorVal); err != nil {
			return err
		}
	}

	// position field validation (0..1 range number)
	posVal, ok := stop["position"]
	if !ok {
		return fmt.Errorf("token '%s': missing 'position' field", stopPath)
	}
	pos, err := toFloat64(posVal)
	if err != nil {
		return fmt.Errorf("token '%s': 'position' must be numeric (got %T)", stopPath, posVal)
	}
	if pos < 0 || pos > 1 {
		return fmt.Errorf("token '%s': 'position' %g out of [0, 1] range", stopPath, pos)
	}

	return nil
}

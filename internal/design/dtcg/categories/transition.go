package categories

import "fmt"

// ValidateTransition: DTCG 2025.10 transition composite category validation.
// Required: duration, timingFunction(cubicBezier) / Optional: delay(duration).
func ValidateTransition(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("token '%s': transition value must be map (got %T)", tokenPath, value)
	}

	// duration required field validation
	durationVal, ok := m["duration"]
	if !ok {
		return fmt.Errorf("token '%s': transition missing 'duration' field", tokenPath)
	}
	if s, isStr := durationVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateDuration(tokenPath+".duration", durationVal); err != nil {
			return err
		}
	}

	// timingFunction required field validation (cubicBezier)
	tfVal, ok := m["timingFunction"]
	if !ok {
		return fmt.Errorf("token '%s': transition missing 'timingFunction' field", tokenPath)
	}
	if s, isStr := tfVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateCubicBezier(tokenPath+".timingFunction", tfVal); err != nil {
			return err
		}
	}

	// delay optional field validation
	if delayVal, exists := m["delay"]; exists {
		if s, isStr := delayVal.(string); !isStr || !IsAlias(s) {
			if err := ValidateDuration(tokenPath+".delay", delayVal); err != nil {
				return err
			}
		}
	}

	return nil
}

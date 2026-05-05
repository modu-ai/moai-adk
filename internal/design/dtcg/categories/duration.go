package categories

import (
	"fmt"
	"regexp"
)

// allowedDurationUnits: DTCG 2025.10 duration allowed units.
var allowedDurationUnits = map[string]bool{
	"ms": true,
	"s":  true,
}

// legacyDurationPattern: Legacy string format like "300ms", "0.3s".
var legacyDurationPattern = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+(ms|s)$`)

// ValidateDuration: DTCG 2025.10 duration category validation.
// Structured format: {value: number, unit: "ms"|"s"}
// Legacy string: "300ms", "0.3s"
func ValidateDuration(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return validateDurationMap(tokenPath, v)
	case string:
		return validateDurationString(tokenPath, v)
	default:
		return fmt.Errorf("token '%s': duration value must be map or string (got %T)", tokenPath, value)
	}
}

// validateDurationMap: Structured {value, unit} format validation.
func validateDurationMap(tokenPath string, m map[string]any) error {
	rawValue, hasValue := m["value"]
	if !hasValue {
		return fmt.Errorf("token '%s': duration map missing 'value' field", tokenPath)
	}
	if !isNumeric(rawValue) {
		return fmt.Errorf("token '%s': duration 'value' must be numeric (got %T)", tokenPath, rawValue)
	}

	unit, ok := m["unit"]
	if !ok {
		return fmt.Errorf("token '%s': duration map missing 'unit' field", tokenPath)
	}
	unitStr, ok := unit.(string)
	if !ok || !allowedDurationUnits[unitStr] {
		return fmt.Errorf("token '%s': duration unit '%v' not supported (allowed: ms, s)", tokenPath, unit)
	}

	return nil
}

// validateDurationString: Legacy "300ms" format validation.
func validateDurationString(tokenPath, s string) error {
	if !legacyDurationPattern.MatchString(s) {
		return fmt.Errorf("token '%s': duration string '%s' invalid format (allowed: number+unit, e.g., 300ms, 0.3s)", tokenPath, s)
	}
	return nil
}

package categories

import (
	"fmt"
	"regexp"
	"strings"
)

// allowedDimensionUnits: DTCG 2025.10 §8.2 allowed units.
var allowedDimensionUnits = map[string]bool{
	"px":  true,
	"rem": true,
	"em":  true,
	"%":   true,
}

// legacyDimensionPattern: Legacy string format like "16px", "1.5rem", "0%".
var legacyDimensionPattern = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+(px|rem|em|%)$`)

// ValidateDimension: DTCG 2025.10 §8.2 dimension category validation.
// Structured format: {value: number, unit: "px"|"rem"|"em"|"%"}
// Legacy string format: "16px", "1.5rem", "0%" (backward compatibility)
func ValidateDimension(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return validateDimensionMap(tokenPath, v)
	case string:
		return validateDimensionString(tokenPath, v)
	default:
		return fmt.Errorf("token '%s': dimension value must be map or string (got %T)", tokenPath, value)
	}
}

// validateDimensionMap: Structured {value, unit} format validation.
func validateDimensionMap(tokenPath string, m map[string]any) error {
	// value field validation
	rawValue, hasValue := m["value"]
	if !hasValue {
		return fmt.Errorf("token '%s': dimension map missing 'value' field", tokenPath)
	}
	if !isNumeric(rawValue) {
		return fmt.Errorf("token '%s': dimension 'value' must be numeric (got %T)", tokenPath, rawValue)
	}

	// unit field validation
	unit, ok := m["unit"]
	if !ok {
		return fmt.Errorf("token '%s': dimension map missing 'unit' field", tokenPath)
	}
	unitStr, ok := unit.(string)
	if !ok {
		return fmt.Errorf("token '%s': dimension 'unit' must be string", tokenPath)
	}
	if !allowedDimensionUnits[unitStr] {
		return fmt.Errorf("token '%s': dimension unit '%s' not supported (allowed: px, rem, em, %%)", tokenPath, unitStr)
	}

	return nil
}

// validateDimensionString: Legacy "16px" format validation.
func validateDimensionString(tokenPath, s string) error {
	if s == "" {
		return fmt.Errorf("token '%s': dimension value is empty string", tokenPath)
	}
	// Pure numbers without units like "0" are rejected even in legacy spec
	if !legacyDimensionPattern.MatchString(s) {
		// Handle unit-only cases (e.g., "px") or invalid units
		return fmt.Errorf("token '%s': dimension string '%s' invalid format (allowed: number+unit, e.g., 16px, 1.5rem, 100%%)", tokenPath, s)
	}

	// Extract unit and verify against allowed list
	unit := extractUnit(s)
	if !allowedDimensionUnits[unit] {
		return fmt.Errorf("token '%s': dimension unit '%s' not supported (allowed: px, rem, em, %%)", tokenPath, unit)
	}

	return nil
}

// extractUnit: "16px" → "px", "1.5rem" → "rem", "100%" → "%".
func extractUnit(s string) string {
	for _, u := range []string{"rem", "em", "px", "%"} {
		if strings.HasSuffix(s, u) {
			return u
		}
	}
	return ""
}

// isNumeric: Check if value is numeric type (float64, int, float32, etc.).
func isNumeric(v any) bool {
	switch v.(type) {
	case float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

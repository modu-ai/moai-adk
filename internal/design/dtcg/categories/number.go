package categories

import "fmt"

// ValidateNumber: DTCG 2025.10 number category validation.
// Allowed: numeric primitive values (float64, int, float32, etc.).
// Rejected: strings, bool, nil, map, etc.
func ValidateNumber(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	if isNumeric(value) {
		return nil
	}
	return fmt.Errorf("token '%s': number value must be numeric (got %T)", tokenPath, value)
}

package categories

import "fmt"

// ValidateFontFamily: DTCG 2025.10 fontFamily category validation.
// Allowed formats: single string or string array (fallback chain).
func ValidateFontFamily(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("token '%s': fontFamily value is empty string", tokenPath)
		}
		return nil
	case []any:
		if len(v) == 0 {
			return fmt.Errorf("token '%s': fontFamily array is empty", tokenPath)
		}
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				return fmt.Errorf("token '%s': fontFamily array element %d is not string (got %T)", tokenPath, i, item)
			}
			if s == "" {
				return fmt.Errorf("token '%s': fontFamily array element %d is empty string", tokenPath, i)
			}
		}
		return nil
	default:
		return fmt.Errorf("token '%s': fontFamily value must be string or string array (got %T)", tokenPath, value)
	}
}

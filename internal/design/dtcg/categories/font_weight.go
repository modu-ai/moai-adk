package categories

import "fmt"

// namedFontWeights: DTCG 2025.10 fontWeight allowed named values.
var namedFontWeights = map[string]bool{
	"thin":    true,
	"light":   true,
	"normal":  true,
	"regular": true,
	"medium":  true,
	"bold":    true,
	"heavy":   true,
}

// ValidateFontWeight: DTCG 2025.10 fontWeight category validation.
// Allowed: 100-900 numeric (100 increments) or named ("thin","light","normal","regular","medium","bold","heavy").
func ValidateFontWeight(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("token '%s': fontWeight value is empty string", tokenPath)
		}
		if !namedFontWeights[v] {
			return fmt.Errorf("token '%s': fontWeight named value '%s' not supported (allowed: thin, light, normal, regular, medium, bold, heavy)", tokenPath, v)
		}
		return nil
	case float64:
		return validateNumericFontWeight(tokenPath, v)
	case int:
		return validateNumericFontWeight(tokenPath, float64(v))
	case float32:
		return validateNumericFontWeight(tokenPath, float64(v))
	default:
		return fmt.Errorf("token '%s': fontWeight value must be numeric (100-900) or named string (got %T)", tokenPath, value)
	}
}

// validateNumericFontWeight: Numeric font weight validation (100-900, 100 increments).
func validateNumericFontWeight(tokenPath string, v float64) error {
	if v < 100 || v > 900 {
		return fmt.Errorf("token '%s': fontWeight numeric value %g out of range (allowed: 100-900)", tokenPath, v)
	}
	// 100 increment validation
	if int(v)%100 != 0 {
		return fmt.Errorf("token '%s': fontWeight numeric value %g must be in 100 increments", tokenPath, v)
	}
	return nil
}

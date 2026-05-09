package categories

import "fmt"

// ValidateFont: DTCG 2025.10 font composite category validation.
// Required: family, size, weight / Optional: style, lineHeight.
// Each field allows direct value or alias reference ({token.path}).
func ValidateFont(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("token '%s': font value must be map (got %T)", tokenPath, value)
	}

	// family field validation
	if err := validateFontField(tokenPath, m, "family", func(v any) error {
		return ValidateFontFamily(tokenPath+".family", v)
	}); err != nil {
		return err
	}

	// size field validation (dimension type)
	if err := validateFontField(tokenPath, m, "size", func(v any) error {
		return ValidateDimension(tokenPath+".size", v)
	}); err != nil {
		return err
	}

	// weight field validation (fontWeight type)
	if err := validateFontField(tokenPath, m, "weight", func(v any) error {
		return ValidateFontWeight(tokenPath+".weight", v)
	}); err != nil {
		return err
	}

	// Optional fields (style, lineHeight) - if present, must be string or alias
	for _, optField := range []string{"style", "lineHeight"} {
		if v, exists := m[optField]; exists {
			if s, ok := v.(string); !ok || (s == "" && !IsAlias(s)) {
				if !ok {
					return fmt.Errorf("token '%s': font '%s' field must be string (got %T)", tokenPath, optField, v)
				}
			}
		}
	}

	return nil
}

// validateFontField: Single field validation helper for font composite token.
// Alias references pass validation; missing required fields error.
func validateFontField(tokenPath string, m map[string]any, field string, validate func(any) error) error {
	v, ok := m[field]
	if !ok {
		return fmt.Errorf("token '%s': font composite token missing '%s' field", tokenPath, field)
	}
	// Alias reference passes
	if s, isStr := v.(string); isStr && IsAlias(s) {
		return nil
	}
	return validate(v)
}

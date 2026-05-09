package categories

import "fmt"

// ValidateTypography: DTCG 2025.10 typography composite category validation.
// Superset of font category: additionally supports letterSpacing, textDecoration, textTransform.
// Required: family, size, weight / Optional: style, lineHeight, letterSpacing, textDecoration, textTransform.
func ValidateTypography(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("token '%s': typography value must be map (got %T)", tokenPath, value)
	}

	// Required field validation same as font category (family, size, weight)
	if err := validateFontField(tokenPath, m, "family", func(v any) error {
		return ValidateFontFamily(tokenPath+".family", v)
	}); err != nil {
		return err
	}

	if err := validateFontField(tokenPath, m, "size", func(v any) error {
		return ValidateDimension(tokenPath+".size", v)
	}); err != nil {
		return err
	}

	if err := validateFontField(tokenPath, m, "weight", func(v any) error {
		return ValidateFontWeight(tokenPath+".weight", v)
	}); err != nil {
		return err
	}

	// Typography extension optional fields - if present, must be string
	optionalStringFields := []string{"style", "lineHeight", "letterSpacing", "textDecoration", "textTransform"}
	for _, field := range optionalStringFields {
		if v, exists := m[field]; exists {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("token '%s': typography '%s' field must be string (got %T)", tokenPath, field, v)
			}
		}
	}

	return nil
}

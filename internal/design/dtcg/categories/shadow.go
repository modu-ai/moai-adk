package categories

import "fmt"

// shadowRequiredFields: Required field list for shadow composite token.
var shadowRequiredFields = []string{"color", "offsetX", "offsetY", "blur", "spread"}

// ValidateShadow: DTCG 2025.10 shadow composite category validation.
// Allowed: single {color, offsetX, offsetY, blur, spread, inset?} or multi-layer array.
func ValidateShadow(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		// Single shadow
		return validateShadowLayer(tokenPath, v)
	case []any:
		// Multi-layer shadow array
		if len(v) == 0 {
			return fmt.Errorf("token '%s': shadow array is empty", tokenPath)
		}
		for i, item := range v {
			layer, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("token '%s': shadow array element %d is not map (got %T)", tokenPath, i, item)
			}
			if err := validateShadowLayer(fmt.Sprintf("%s[%d]", tokenPath, i), layer); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("token '%s': shadow value must be map or array (got %T)", tokenPath, value)
	}
}

// validateShadowLayer: Single shadow layer {color, offsetX, offsetY, blur, spread, inset?} validation.
func validateShadowLayer(tokenPath string, m map[string]any) error {
	// Required field existence check and validation
	for _, field := range shadowRequiredFields {
		v, ok := m[field]
		if !ok {
			return fmt.Errorf("token '%s': shadow missing '%s' field", tokenPath, field)
		}

		// Alias reference passes
		if s, isStr := v.(string); isStr && IsAlias(s) {
			continue
		}

		// color field validated with ValidateColor
		if field == "color" {
			if err := ValidateColor(tokenPath+".color", v); err != nil {
				return err
			}
			continue
		}

		// offsetX, offsetY, blur, spread validated with dimension
		if err := ValidateDimension(tokenPath+"."+field, v); err != nil {
			return err
		}
	}

	// inset optional field - must be bool
	if inset, exists := m["inset"]; exists {
		if _, ok := inset.(bool); !ok {
			return fmt.Errorf("token '%s': shadow 'inset' field must be bool (got %T)", tokenPath, inset)
		}
	}

	return nil
}

// ValidateBorder: DTCG 2025.10 border composite category validation.
// {color: color, width: dimension, style: strokeStyle}.
func ValidateBorder(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	m, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("token '%s': border value must be map (got %T)", tokenPath, value)
	}

	// color field validation
	colorVal, ok := m["color"]
	if !ok {
		return fmt.Errorf("token '%s': border missing 'color' field", tokenPath)
	}
	if s, isStr := colorVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateColor(tokenPath+".color", colorVal); err != nil {
			return err
		}
	}

	// width field validation (dimension)
	widthVal, ok := m["width"]
	if !ok {
		return fmt.Errorf("token '%s': border missing 'width' field", tokenPath)
	}
	if s, isStr := widthVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateDimension(tokenPath+".width", widthVal); err != nil {
			return err
		}
	}

	// style field validation (strokeStyle)
	styleVal, ok := m["style"]
	if !ok {
		return fmt.Errorf("token '%s': border missing 'style' field", tokenPath)
	}
	if s, isStr := styleVal.(string); !isStr || !IsAlias(s) {
		if err := ValidateStrokeStyle(tokenPath+".style", styleVal); err != nil {
			return err
		}
	}

	return nil
}

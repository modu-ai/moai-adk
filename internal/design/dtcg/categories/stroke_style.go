package categories

import "fmt"

// validStrokeStyleEnums: DTCG 2025.10 strokeStyle allowed enum values.
var validStrokeStyleEnums = map[string]bool{
	"solid":  true,
	"dashed": true,
	"dotted": true,
	"double": true,
	"groove": true,
	"ridge":  true,
	"outset": true,
	"inset":  true,
}

// validLineCaps: strokeStyle composite format lineCap allowed values.
var validLineCaps = map[string]bool{
	"butt":   true,
	"round":  true,
	"square": true,
}

// ValidateStrokeStyle: DTCG 2025.10 strokeStyle category validation.
// Allowed: enum string or {dashArray: [<dimension>...], lineCap: "butt"|"round"|"square"}.
func ValidateStrokeStyle(tokenPath string, value any) error {
	// Alias reference passes
	if s, ok := value.(string); ok && IsAlias(s) {
		return nil
	}

	switch v := value.(type) {
	case string:
		if !validStrokeStyleEnums[v] {
			return fmt.Errorf("token '%s': strokeStyle enum '%s' not supported (allowed: solid, dashed, dotted, double, groove, ridge, outset, inset)", tokenPath, v)
		}
		return nil
	case map[string]any:
		return validateStrokeStyleMap(tokenPath, v)
	default:
		return fmt.Errorf("token '%s': strokeStyle value must be string enum or map (got %T)", tokenPath, value)
	}
}

// validateStrokeStyleMap: Composite strokeStyle {dashArray, lineCap} validation.
func validateStrokeStyleMap(tokenPath string, m map[string]any) error {
	// dashArray field validation
	rawDA, ok := m["dashArray"]
	if !ok {
		return fmt.Errorf("token '%s': strokeStyle map missing 'dashArray' field", tokenPath)
	}
	dashArray, ok := rawDA.([]any)
	if !ok {
		return fmt.Errorf("token '%s': strokeStyle 'dashArray' must be array (got %T)", tokenPath, rawDA)
	}
	if len(dashArray) == 0 {
		return fmt.Errorf("token '%s': strokeStyle 'dashArray' array is empty", tokenPath)
	}
	// Each dashArray element must be dimension format
	for i, item := range dashArray {
		if err := ValidateDimension(fmt.Sprintf("%s.dashArray[%d]", tokenPath, i), item); err != nil {
			return err
		}
	}

	// lineCap field validation
	rawLC, ok := m["lineCap"]
	if !ok {
		return fmt.Errorf("token '%s': strokeStyle map missing 'lineCap' field", tokenPath)
	}
	lineCap, ok := rawLC.(string)
	if !ok {
		return fmt.Errorf("token '%s': strokeStyle 'lineCap' must be string (got %T)", tokenPath, rawLC)
	}
	if !validLineCaps[lineCap] {
		return fmt.Errorf("token '%s': strokeStyle 'lineCap' value '%s' not supported (allowed: butt, round, square)", tokenPath, lineCap)
	}

	return nil
}

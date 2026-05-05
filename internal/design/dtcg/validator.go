package dtcg

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/design/dtcg/categories"
)

// categoryValidator: Single DTCG category validation function signature.
type categoryValidator func(tokenPath string, value any) error

// categoryValidators: DTCG 2025.10 §8 category → validation function mapping.
// Source: https://tr.designtokens.org/format/ (Editor's Draft 2025-10)
//
// @MX:ANCHOR: [AUTO] Validate dispatch core map - fan_in increase expected
// @MX:REASON: All category validations dispatch through this map (covers entire DTCG §8)
var categoryValidators = map[string]categoryValidator{
	"color":        categories.ValidateColor,
	"dimension":    categories.ValidateDimension,
	"fontFamily":   categories.ValidateFontFamily,
	"fontWeight":   categories.ValidateFontWeight,
	"font":         categories.ValidateFont,
	"typography":   categories.ValidateTypography,
	"duration":     categories.ValidateDuration,
	"cubicBezier":  categories.ValidateCubicBezier,
	"number":       categories.ValidateNumber,
	"strokeStyle":  categories.ValidateStrokeStyle,
	"border":       categories.ValidateBorder,
	"transition":   categories.ValidateTransition,
	"shadow":       categories.ValidateShadow,
	"gradient":     categories.ValidateGradient,
}

// Validate: Validates entire token map according to DTCG 2025.10 spec.
// REQ-DPL-004, REQ-DPL-010.
//
// tokens: Token key → token definition map (each definition includes $type, $value, $description)
//
// Returns:
//   - *Report: Validation result (check Valid field for success)
//   - error: When validation itself cannot be executed (e.g., tokens is not nil but wrong type)
//
// Performance: [HARD] Must complete in <100ms for under 500 tokens (REQ-DPL-010).
func Validate(tokens map[string]any) (*Report, error) {
	report := &Report{
		Valid: true,
	}

	// Treat nil or empty map as valid empty token set
	if len(tokens) == 0 {
		return report, nil
	}

	for key, rawToken := range tokens {
		// Token must be in map[string]any format
		tokenDef, ok := rawToken.(map[string]any)
		if !ok {
			// Skip non-map values (may be part of DTCG group structure)
			continue
		}

		// Skip group node if both $type and $value are absent
		_, hasType := tokenDef["$type"]
		_, hasValue := tokenDef["$value"]
		if !hasType && !hasValue {
			continue
		}

		// $type required validation
		if !hasType {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  "(unknown)",
				Rule:      "$type field missing - DTCG 2025.10 §9: $type is required",
			})
			continue
		}

		// $value required validation
		if !hasValue {
			category, _ := tokenDef["$type"].(string)
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  category,
				Rule:      "$value field missing - DTCG 2025.10 §9: $value is required",
			})
			continue
		}

		typeVal, ok := tokenDef["$type"].(string)
		if !ok {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  "(invalid)",
				Rule:      fmt.Sprintf("$type must be string (got %T)", tokenDef["$type"]),
			})
			continue
		}

		// Category validity validation
		validator, supported := categoryValidators[typeVal]
		if !supported {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  typeVal,
				Rule:      fmt.Sprintf("unknown category '%s' - see DTCG 2025.10 §8", typeVal),
				Value:     typeVal,
			})
			continue
		}

		// Category-specific value validation
		if err := validator(key, tokenDef["$value"]); err != nil {
			report.AddError(&ValidationError{
				TokenPath: key,
				Category:  typeVal,
				Rule:      err.Error(),
				Value:     tokenDef["$value"],
			})
		}
	}

	report.TokenCount = countTokens(tokens)
	return report, nil
}

// countTokens: Counts actual tokens in token map.
// Only aggregates items with $type/$value as tokens.
func countTokens(tokens map[string]any) int {
	count := 0
	for _, raw := range tokens {
		def, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		_, hasType := def["$type"]
		_, hasValue := def["$value"]
		if hasType || hasValue {
			count++
		}
	}
	return count
}

// Package pipeline: /moai design workflow brand conflict checker.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-05) implementation.
//
// BrandConflict extracts colors from `.moai/project/brand/visual-identity.md` and
// compares them with color tokens in tokens.json, returning ValidationWarning on conflicts.
//
// Subagent boundary principle: This package only returns warning data,
// does not call AskUserQuestion directly (orchestrator's responsibility).
package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// hexColorPattern: Regex to extract hex color codes (#RGB, #RRGGBB) from markdown.
var hexColorPattern = regexp.MustCompile(`#(?:[0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})\b`)

// BrandColor: Brand color entry extracted from visual-identity.md.
type BrandColor struct {
	// HexValue: Normalized lowercase hex code (e.g., "#1d4ed8")
	HexValue string
	// SourceLine: Original markdown line (for debugging)
	SourceLine string
}

// ExtractBrandColors: Extracts hex color codes from visual-identity.md file.
//
// Returns empty slice if file doesn't exist or only has _TBD_ placeholder.
// Returns error for file read errors.
func ExtractBrandColors(visualIdentityPath string) ([]BrandColor, error) {
	data, err := os.ReadFile(visualIdentityPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No file: brand context uninitialized — return empty result
			return nil, nil
		}
		return nil, fmt.Errorf("visual-identity.md read failed (%s): %w", visualIdentityPath, err)
	}

	content := string(data)

	// Case with only _TBD_ placeholder (brand interview incomplete)
	if strings.Contains(content, "_TBD_") && !hexColorPattern.MatchString(content) {
		return nil, nil
	}

	var colors []BrandColor
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		matches := hexColorPattern.FindAllString(line, -1)
		for _, hex := range matches {
			colors = append(colors, BrandColor{
				HexValue:   strings.ToLower(hex),
				SourceLine: strings.TrimSpace(line),
			})
		}
	}

	return colors, scanner.Err()
}

// CheckBrandConflicts: Compares tokens.json color tokens with brand colors and returns conflicts.
//
// tokens: Same map[string]any passed to Validate() (tokens.json parse result)
// brandColors: List of brand colors extracted by ExtractBrandColors()
//
// Returns: ValidationWarning slice if conflicts exist (empty slice if none)
//
// [HARD] Preserve subagent boundary: This function only returns warning data,
// does not request AskUserQuestion or user input.
//
// @MX:ANCHOR: [AUTO] Brand conflict check core entry point — called by orchestrator
// @MX:REASON: REQ-DPL-009 brand context priority implementation; fan_in increase expected
func CheckBrandConflicts(tokens map[string]any, brandColors []BrandColor) []*dtcg.ValidationWarning {
	if len(brandColors) == 0 {
		return nil
	}

	// Build Set of brand hex values (O(1) lookup)
	brandHexSet := make(map[string]struct{}, len(brandColors))
	for _, bc := range brandColors {
		brandHexSet[bc.HexValue] = struct{}{}
	}

	var warnings []*dtcg.ValidationWarning

	for key, rawToken := range tokens {
		tokenDef, ok := rawToken.(map[string]any)
		if !ok {
			continue
		}

		typeVal, _ := tokenDef["$type"].(string)
		if typeVal != "color" {
			continue
		}

		rawValue, hasValue := tokenDef["$value"]
		if !hasValue {
			continue
		}

		tokenHex, ok := rawValue.(string)
		if !ok {
			continue
		}

		normalizedTokenHex := strings.ToLower(strings.TrimSpace(tokenHex))

		// Conflict if token color value not in brand color set
		if _, inBrand := brandHexSet[normalizedTokenHex]; !inBrand {
			warnings = append(warnings, &dtcg.ValidationWarning{
				TokenPath: key,
				Category:  "brand-conflict",
				Message: fmt.Sprintf(
					"color token value '%s' not in brand visual-identity.md — use brand color or add via interview",
					normalizedTokenHex,
				),
			})
		}
	}

	return warnings
}

// RunBrandConflictCheck: One-step conflict check given visual-identity.md path and tokens.
//
// Returns empty result if visualIdentityPath is empty or file doesn't exist.
// This is a convenience function that composes ExtractBrandColors + CheckBrandConflicts.
func RunBrandConflictCheck(visualIdentityPath string, tokens map[string]any) ([]*dtcg.ValidationWarning, error) {
	brandColors, err := ExtractBrandColors(visualIdentityPath)
	if err != nil {
		return nil, err
	}
	return CheckBrandConflicts(tokens, brandColors), nil
}

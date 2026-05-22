package router

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// defaultEffortMapping is the default map used by EffortForLevel.
// Used when cfg.EffortMapping is empty or missing the level key.
// REQ-HRN-001-005: minimal->medium, standard->high, thorough->xhigh.
var defaultEffortMapping = map[Level]string{
	LevelMinimal:  "medium",
	LevelStandard: "high",
	LevelThorough: "xhigh",
}

// EffortForLevel returns the effort-level string for the given harness level.
// Reads the level key from cfg.EffortMapping, falling back to the default if absent.
// REQ-HRN-001-005, AC-HRN-001-10.
//
// @MX:ANCHOR: [AUTO] EffortForLevel — source of the effort field for CLI route --json
// @MX:REASON: fan_in >= 3: called by CLI route command, effort_test.go, and the harness router integration
func EffortForLevel(level Level, cfg *config.HarnessConfig) string {
	if cfg != nil && len(cfg.EffortMapping) > 0 {
		if v, ok := cfg.EffortMapping[string(level)]; ok && v != "" {
			return v
		}
	}
	// Default fallback.
	if v, ok := defaultEffortMapping[level]; ok {
		return v
	}
	return "medium" // Final fallback.
}

// EffortForLevelFromProxy returns the effort level via a ConfigProxy.
// Used in tests that rely on a lightweight config wrapper.
func EffortForLevelFromProxy(level Level, proxy *ConfigProxy) string {
	if proxy != nil && len(proxy.EffortMapping) > 0 {
		if v, ok := proxy.EffortMapping[string(level)]; ok && v != "" {
			return v
		}
	}
	// Default fallback.
	if v, ok := defaultEffortMapping[level]; ok {
		return v
	}
	return "medium"
}

// ParseProfileFloor parses the minimum pass_threshold from an evaluator profile .md file.
// REQ-HRN-001-012: used to verify the pass_threshold >= 0.60 FROZEN floor.
// Searches the file for "pass_threshold" or the PassThreshold value in a table and returns the minimum.
func ParseProfileFloor(profilePath string) (float64, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return 0, fmt.Errorf("ParseProfileFloor: read %q: %w", profilePath, err)
	}

	content := string(data)
	minThreshold := 1.0
	found := false

	// Extract values from the "Pass Threshold" table column.
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}

		cells := parseTableRowSimple(line)
		if len(cells) < 3 {
			continue
		}

		// Skip header / separator rows.
		header := strings.ToLower(cells[0])
		if strings.Contains(header, "dimension") || strings.Contains(cells[0], "---") {
			continue
		}

		// Parse the third column (Pass Threshold).
		thresholdStr := strings.TrimSpace(cells[2])
		if strings.HasSuffix(thresholdStr, "%") {
			thresholdStr = strings.TrimSuffix(thresholdStr, "%")
			v, err := strconv.ParseFloat(strings.TrimSpace(thresholdStr), 64)
			if err == nil {
				v = v / 100.0
				if v < minThreshold {
					minThreshold = v
					found = true
				}
			}
		} else {
			v, err := strconv.ParseFloat(thresholdStr, 64)
			if err == nil && v > 0 {
				if v < minThreshold {
					minThreshold = v
					found = true
				}
			}
		}
	}

	if !found {
		// Fall back to the default (FROZEN floor) when no pass_threshold table is present.
		return 0.60, nil
	}
	return minThreshold, nil
}

// parseTableRowSimple performs a simple parse of a Markdown table row.
func parseTableRowSimple(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// Ensure config import is used.
var _ *config.HarnessConfig

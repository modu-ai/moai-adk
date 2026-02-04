package statusline

import (
	"fmt"
	"strconv"
	"strings"
)

// CollectMetrics extracts session cost and model information from stdin data.
// Returns a MetricsData with Available=false if input is nil.
func CollectMetrics(input *StdinData) *MetricsData {
	if input == nil {
		return &MetricsData{Available: false}
	}

	// Extract model name from nested structure
	// Priority: display_name (use directly) > id/name (shorten)
	// Per https://code.claude.com/docs/en/statusline documentation
	var modelName string
	if input.Model != nil {
		if input.Model.DisplayName != "" {
			modelName = input.Model.DisplayName
		} else if input.Model.ID != "" {
			modelName = ShortenModelName(input.Model.ID)
		} else if input.Model.Name != "" {
			modelName = ShortenModelName(input.Model.Name)
		}
	}

	data := &MetricsData{
		Model:     modelName,
		Available: modelName != "",
	}

	if input.Cost != nil {
		// Support both field names
		if input.Cost.TotalCostUSD > 0 {
			data.CostUSD = input.Cost.TotalCostUSD
		} else {
			data.CostUSD = input.Cost.TotalUSD
		}
	}

	return data
}

// ShortenModelName abbreviates a Claude model name to match Python's format.
// Converts to capitalized name with spaces instead of hyphens.
// Examples:
//
//	"claude-opus-4-5-20250514"  -> "Opus 4.5"
//	"claude-sonnet-4-20250514"  -> "Sonnet 4"
//	"claude-3-5-sonnet-20241022" -> "Sonnet 3.5"
//	"claude-3-5-haiku-20241022"  -> "Haiku 3.5"
//	"gpt-4"                       -> "gpt-4" (non-Claude names unchanged)
func ShortenModelName(model string) string {
	if model == "" {
		return ""
	}

	// Handle non-Claude models
	if !strings.HasPrefix(model, "claude-") {
		return model
	}

	// Remove "claude-" prefix
	name := strings.TrimPrefix(model, "claude-")

	// Remove trailing date suffix (e.g., "-20250514")
	parts := strings.Split(name, "-")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if len(last) == 8 {
			if _, err := strconv.Atoi(last); err == nil {
				name = strings.Join(parts[:len(parts)-1], "-")
				parts = strings.Split(name, "-")
			}
		}
	}

	// Parse the model components to extract name and version
	// Expected patterns:
	// - "opus-4-5" -> Opus 4.5
	// - "sonnet-4" -> Sonnet 4
	// - "3-5-sonnet" -> Sonnet 3.5
	// - "3-5-haiku" -> Haiku 3.5

	var modelName string
	var versionParts []string

	// Check if the model name (sonnet, opus, haiku) is at the end or beginning
	for i, part := range parts {
		lower := strings.ToLower(part)
		if lower == "sonnet" || lower == "opus" || lower == "haiku" {
			// Capitalize the model name
			modelName = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			// Everything before is version
			versionParts = parts[:i]
			// Everything after is additional version info
			if i < len(parts)-1 {
				versionParts = append(versionParts, parts[i+1:]...)
			}
			break
		}
	}

	// If no known model name found, use the first part as name
	if modelName == "" && len(parts) > 0 {
		if len(parts[0]) > 0 {
			modelName = strings.ToUpper(parts[0][:1]) + strings.ToLower(parts[0][1:])
		}
		if len(parts) > 1 {
			versionParts = parts[1:]
		}
	}

	// Format version parts with dots for numeric version
	// e.g., ["4", "5"] -> "4.5", ["3", "5"] -> "3.5"
	var versionStr string
	if len(versionParts) > 0 {
		versionStr = strings.Join(versionParts, ".")
	}

	if versionStr != "" {
		return modelName + " " + versionStr
	}
	return modelName
}

// formatCost formats a USD cost value as a string with two decimal places.
func formatCost(usd float64) string {
	return fmt.Sprintf("$%.2f", usd)
}

// formatTokens formats a token count with K suffix for thousands.
// Examples: 50000 -> "50K", 200000 -> "200K", 500 -> "500"
func formatTokens(tokens int) string {
	if tokens >= 1000 {
		return fmt.Sprintf("%dK", tokens/1000)
	}
	return fmt.Sprintf("%d", tokens)
}

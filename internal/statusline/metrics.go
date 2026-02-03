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

	data := &MetricsData{
		Model:     ShortenModelName(input.Model),
		Available: true,
	}

	if input.Cost != nil {
		data.CostUSD = input.Cost.TotalUSD
	}

	return data
}

// ShortenModelName abbreviates a Claude model name by removing the
// "claude-" prefix and any trailing date suffix (YYYYMMDD format).
// Examples:
//
//	"claude-sonnet-4"           -> "sonnet-4"
//	"claude-opus-4-5-20251101"  -> "opus-4-5"
//	"gpt-4"                    -> "gpt-4" (non-Claude names unchanged)
func ShortenModelName(model string) string {
	if model == "" {
		return ""
	}

	name := strings.TrimPrefix(model, "claude-")

	// Remove trailing date suffix (e.g., "-20251101")
	parts := strings.Split(name, "-")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if len(last) == 8 {
			if _, err := strconv.Atoi(last); err == nil {
				name = strings.Join(parts[:len(parts)-1], "-")
			}
		}
	}

	return name
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

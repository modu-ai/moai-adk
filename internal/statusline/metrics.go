package statusline

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CollectMetrics extracts session cost and model information from stdin data.
// Returns a MetricsData with Available=false if input is nil.
// Model name fallback chain (AC-SF-002/003/004):
// 1. stdin model field
// 2. MOAI_LAST_MODEL env var
// 3. Cache file (~/.moai/state/last-model.txt)
// 4. Unavailable
//
// AC-SF-005: When model is available, writes to cache file for future fallback.
//
// @MX:NOTE: [AUTO] Fallback chain: stdin → MOAI_LAST_MODEL → cache file → unavailable
// @MX:SPEC: SPEC-V3R3-STATUSLINE-FALLBACK-001
func CollectMetrics(input *StdinData, homeDir string) *MetricsData {
	if input == nil {
		// Try fallback chain for nil input (AC-SF-001/003/004)
		return collectMetricsFromFallback(homeDir)
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

	// If no model in stdin, try fallback chain (AC-SF-002/003/004)
	if modelName == "" {
		modelName = getModelNameFromFallback(homeDir)
	}

	// Override display name with actual GLM model when running in GLM mode.
	// Claude Code reports "Opus"/"Sonnet"/"Haiku" even when env vars route to GLM models.
	modelName = resolveGLMModelName(modelName)

	// AC-SF-005: Write model name to cache for future fallback
	if modelName != "" && homeDir != "" {
		_ = WriteModelCache(homeDir, modelName) //nolint:errcheck // EC-SF-003: silent ignore
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
		// Extract session duration (REQ-V3-TIME-001)
		data.SessionDurationMS = input.Cost.TotalDurationMS
	}

	return data
}

// collectMetricsFromFallback attempts to get model name from fallback sources.
// Called when stdin is nil (AC-SF-001).
func collectMetricsFromFallback(homeDir string) *MetricsData {
	modelName := getModelNameFromFallback(homeDir)

	return &MetricsData{
		Model:     modelName,
		Available: modelName != "",
	}
}

// getModelNameFromFallback implements the fallback chain:
// 1. MOAI_LAST_MODEL env var (AC-SF-003)
// 2. Cache file (AC-SF-004)
// 3. Empty string (unavailable)
func getModelNameFromFallback(homeDir string) string {
	// Priority 1: MOAI_LAST_MODEL env var (AC-SF-003)
	if envModel := os.Getenv("MOAI_LAST_MODEL"); envModel != "" {
		return ShortenModelName(envModel)
	}

	// Priority 2: Cache file (AC-SF-004)
	if homeDir != "" {
		if cachedModel, err := ReadModelCache(homeDir); err == nil && cachedModel != "" {
			return cachedModel
		}
	}

	// Priority 3: No fallback available
	return ""
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

	// Handle non-Claude models: strip [1m] suffix (GLM doesn't support 1M context)
	if !strings.HasPrefix(model, "claude-") {
		return strings.TrimSuffix(model, "[1m]")
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

// resolveGLMModelName checks if GLM mode is active via environment variables
// and returns the actual GLM model name instead of the Claude display name.
// When ANTHROPIC_DEFAULT_*_MODEL env vars contain non-Claude model names,
// the display name from Claude Code (e.g., "Opus") is replaced with the actual model.
// Also strips [1m] suffix from non-Claude models (GLM doesn't support 1M context).
func resolveGLMModelName(displayName string) string {
	if displayName == "" {
		return displayName
	}

	// Strip [1m] suffix before matching — Claude Code may append it
	// to any model in the Opus/Sonnet slot regardless of provider.
	cleaned := strings.TrimSuffix(displayName, "[1m]")
	cleaned = strings.TrimSpace(cleaned)
	lower := strings.ToLower(cleaned)

	// Map Claude display names to their corresponding env vars
	var envKey string
	switch {
	case strings.Contains(lower, "opus"):
		envKey = "ANTHROPIC_DEFAULT_OPUS_MODEL"
	case strings.Contains(lower, "sonnet"):
		envKey = "ANTHROPIC_DEFAULT_SONNET_MODEL"
	case strings.Contains(lower, "haiku"):
		envKey = "ANTHROPIC_DEFAULT_HAIKU_MODEL"
	default:
		// Not a known Claude display name — might be GLM model name passed directly.
		// Strip [1m] for non-Claude models and return.
		if !strings.HasPrefix(lower, "claude") {
			return cleaned
		}
		return displayName
	}

	glmModel := os.Getenv(envKey)
	if glmModel == "" || strings.HasPrefix(glmModel, "claude-") {
		return displayName
	}

	// Non-Claude model detected (e.g., "glm-5.1", "gpt-4o")
	return glmModel
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

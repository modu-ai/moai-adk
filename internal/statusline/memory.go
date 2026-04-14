package statusline

import (
	"os"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// defaultAutoCompactPct is the default auto-compact trigger threshold percentage.
// Claude Code triggers auto-compact at approximately 80-85% of the total context window.
// This can be overridden by the CLAUDE_AUTOCOMPACT_PCT_OVERRIDE environment variable.
const defaultAutoCompactPct = 85

// glmContextWindows maps known non-Claude (typically GLM-family) model names
// served via the Anthropic-compatible endpoint to their actual context window
// sizes in tokens. Claude Code reports `context_window_size` based on the
// Claude model slot (Opus = 1M, Sonnet/Haiku = 200K), but the underlying GLM
// model has a different limit. Issue #653.
//
// Entries use lowercase substring matching against the resolved model name.
// Add entries here when a new GLM model ships; users can also override per
// invocation via MOAI_STATUSLINE_CONTEXT_SIZE.
var glmContextWindows = map[string]int{
	"glm-5.1":     200_000, // GLM-5.1 (z.ai) — actual ~230K, leave headroom
	"glm-5":       128_000,
	"glm-4.7":     128_000,
	"glm-4.6":     128_000,
	"glm-4.5":     128_000,
	"glm-4.5-air": 128_000,
}

// getAutoCompactThreshold returns the auto-compact trigger percentage.
// Reads CLAUDE_AUTOCOMPACT_PCT_OVERRIDE env var if set, otherwise uses default.
func getAutoCompactThreshold() int {
	if override := os.Getenv(config.EnvClaudeAutoCompactPct); override != "" {
		if v, err := strconv.Atoi(override); err == nil && v > 0 && v <= 100 {
			return v
		}
	}
	return defaultAutoCompactPct
}

// resolveContextWindowOverride returns a non-zero context window override when
// the user is running through GLM (or another non-Claude provider) so that the
// memory gauge reflects the actual provider limit rather than the Claude slot's
// nominal size. Issue #653 — GLM "opus" slot reported as 1M but real limit is
// 128-230K depending on the GLM model.
//
// Resolution priority:
//  1. MOAI_STATUSLINE_CONTEXT_SIZE env (explicit user override)
//  2. ANTHROPIC_DEFAULT_*_MODEL env contains a known GLM model → glmContextWindows
//  3. Return 0 (caller falls back to stdin's context_window_size)
func resolveContextWindowOverride() int {
	if v := os.Getenv(config.EnvStatuslineContextSize); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	// Check the three Anthropic model env slots; first match wins (any of them
	// being non-Claude implies a non-Claude provider is active).
	envSlots := []string{
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
	}
	for _, key := range envSlots {
		val := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
		if val == "" || strings.HasPrefix(val, "claude") {
			continue
		}
		// Match longest known model name first to avoid "glm-4.5" matching
		// before "glm-4.5-air".
		var matched int
		var matchedSize int
		for name, size := range glmContextWindows {
			if strings.Contains(val, name) && len(name) > matched {
				matched = len(name)
				matchedSize = size
			}
		}
		if matchedSize > 0 {
			return matchedSize
		}
	}
	return 0
}

// CollectMemory extracts context window token usage from stdin data.
// Returns a MemoryData with Available=false if input or context_window is nil.
// Priority (following Claude Code documentation):
// 1. Use pre-calculated used_percentage from Claude Code (most accurate)
// 2. Calculate from current_usage tokens
// 3. Fall back to legacy used/total fields
//
// The token budget is scaled to the auto-compact threshold so that the CW bar
// reaches 100% at the point where auto-compact triggers, not at the full context window.
func CollectMemory(input *StdinData) *MemoryData {
	if input == nil || input.ContextWindow == nil {
		return &MemoryData{Available: false}
	}

	ctx := input.ContextWindow

	// Get context window size (default 200K).
	// GLM/non-Claude providers may have a smaller real limit than the Claude
	// slot's reported context_window_size — apply override when detected
	// (issue #653).
	contextSize := ctx.ContextWindowSize
	if contextSize <= 0 {
		contextSize = ctx.Total
	}
	if contextSize <= 0 {
		contextSize = 200000 // Default context window size
	}
	if override := resolveContextWindowOverride(); override > 0 {
		contextSize = override
	}

	// Scale budget to auto-compact threshold so bar shows 100% at compact point
	threshold := getAutoCompactThreshold()
	effectiveBudget := contextSize * threshold / 100

	var tokensUsed int

	// Priority 1: Use pre-calculated percentage from Claude Code
	if ctx.UsedPercentage != nil {
		// Calculate tokens from percentage (of full context window)
		tokensUsed = int(float64(contextSize) * (*ctx.UsedPercentage) / 100.0)
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: effectiveBudget,
			Available:   true,
		}
	}

	// Priority 2: Calculate from current_usage tokens
	if ctx.CurrentUsage != nil {
		cu := ctx.CurrentUsage
		tokensUsed = cu.InputTokens + cu.CacheCreationTokens + cu.CacheReadTokens
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: effectiveBudget,
			Available:   true,
		}
	}

	// Priority 3: Fall back to legacy used/total fields
	if ctx.Used > 0 || ctx.Total > 0 {
		return &MemoryData{
			TokensUsed:  ctx.Used,
			TokenBudget: ctx.Total * threshold / 100,
			Available:   true,
		}
	}

	// No data available - return 0% (session start state)
	return &MemoryData{
		TokensUsed:  0,
		TokenBudget: effectiveBudget,
		Available:   true,
	}
}

// contextUsageLevel determines the color severity level based on
// context window usage percentage.
// Returns levelOk for <50%, levelWarn for 50-80%, levelError for >=80%.
func contextUsageLevel(used, total int) contextLevel {
	if total <= 0 {
		return levelOk
	}

	pct := used * 100 / total

	switch {
	case pct >= 80:
		return levelError
	case pct >= 50:
		return levelWarn
	default:
		return levelOk
	}
}

// usagePercent calculates the percentage of context window used.
// Returns 0 if total is zero to avoid division by zero. Capped at 100.
func usagePercent(used, total int) int {
	if total <= 0 {
		return 0
	}
	pct := used * 100 / total
	if pct > 100 {
		return 100
	}
	return pct
}

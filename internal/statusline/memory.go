package statusline

// CollectMemory extracts context window token usage from stdin data.
// Returns a MemoryData with Available=false if input or context_window is nil.
// Priority (following Claude Code documentation):
// 1. Use pre-calculated used_percentage from Claude Code (most accurate)
// 2. Calculate from current_usage tokens
// 3. Fall back to legacy used/total fields
func CollectMemory(input *StdinData) *MemoryData {
	if input == nil || input.ContextWindow == nil {
		return &MemoryData{Available: false}
	}

	ctx := input.ContextWindow

	// Get context window size (default 200K)
	contextSize := ctx.ContextWindowSize
	if contextSize <= 0 {
		contextSize = ctx.Total
	}
	if contextSize <= 0 {
		contextSize = 200000 // Default context window size
	}

	var tokensUsed int

	// Priority 1: Use pre-calculated percentage from Claude Code
	if ctx.UsedPercentage != nil {
		// Calculate tokens from percentage
		tokensUsed = int(float64(contextSize) * (*ctx.UsedPercentage) / 100.0)
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: contextSize,
			Available:   true,
		}
	}

	// Priority 2: Calculate from current_usage tokens
	if ctx.CurrentUsage != nil {
		cu := ctx.CurrentUsage
		tokensUsed = cu.InputTokens + cu.CacheCreationTokens + cu.CacheReadTokens
		return &MemoryData{
			TokensUsed:  tokensUsed,
			TokenBudget: contextSize,
			Available:   true,
		}
	}

	// Priority 3: Fall back to legacy used/total fields
	if ctx.Used > 0 || ctx.Total > 0 {
		return &MemoryData{
			TokensUsed:  ctx.Used,
			TokenBudget: ctx.Total,
			Available:   true,
		}
	}

	// No data available - return 0% (session start state)
	return &MemoryData{
		TokensUsed:  0,
		TokenBudget: contextSize,
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
// Returns 0 if total is zero to avoid division by zero.
func usagePercent(used, total int) int {
	if total <= 0 {
		return 0
	}
	return used * 100 / total
}

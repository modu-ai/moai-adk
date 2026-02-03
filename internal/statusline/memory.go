package statusline

// CollectMemory extracts context window token usage from stdin data.
// Returns a MemoryData with Available=false if input or context_window is nil.
func CollectMemory(input *StdinData) *MemoryData {
	if input == nil || input.ContextWindow == nil {
		return &MemoryData{Available: false}
	}

	return &MemoryData{
		TokensUsed:  input.ContextWindow.Used,
		TokenBudget: input.ContextWindow.Total,
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

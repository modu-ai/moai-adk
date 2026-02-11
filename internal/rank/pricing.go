// Package rank provides model pricing for MoAI Rank cost calculation.
package rank

// ModelPricing holds pricing information for a Claude model.
// Prices are in USD per million tokens.
type ModelPricing struct {
	Input         float64 `json:"input"`
	Output        float64 `json:"output"`
	CacheCreation float64 `json:"cache_creation"`
	CacheRead     float64 `json:"cache_read"`
}

// modelPricingDB holds the pricing database for all Claude models.
// Source: https://platform.claude.com/docs/en/about-claude/pricing
// Last updated: 2026-02
var modelPricingDB = map[string]ModelPricing{
	// Claude Opus 4.6 - Latest flagship model (same pricing tier as Opus 4/4.1)
	"claude-opus-4-6-20260203": {
		Input:         15.00,
		Output:        75.00,
		CacheCreation: 18.75,
		CacheRead:     1.50,
	},
	// Claude Opus 4.5
	"claude-opus-4-5-20251101": {
		Input:         5.00,
		Output:        25.00,
		CacheCreation: 6.25,
		CacheRead:     0.50,
	},
	// Claude Opus 4.1
	"claude-opus-4-1-20250414": {
		Input:         15.00,
		Output:        75.00,
		CacheCreation: 18.75,
		CacheRead:     1.50,
	},
	// Claude Opus 4
	"claude-opus-4-20250514": {
		Input:         15.00,
		Output:        75.00,
		CacheCreation: 18.75,
		CacheRead:     1.50,
	},
	// Claude Sonnet 4.5
	"claude-sonnet-4-5-20251022": {
		Input:         3.00,
		Output:        15.00,
		CacheCreation: 3.75,
		CacheRead:     0.30,
	},
	// Claude Sonnet 4
	"claude-sonnet-4-20250514": {
		Input:         3.00,
		Output:        15.00,
		CacheCreation: 3.75,
		CacheRead:     0.30,
	},
	// Claude Sonnet 3.7 (deprecated but still supported)
	"claude-3-7-sonnet-20250219": {
		Input:         3.00,
		Output:        15.00,
		CacheCreation: 3.75,
		CacheRead:     0.30,
	},
	// Claude Haiku 4.5
	"claude-haiku-4-5-20251022": {
		Input:         1.00,
		Output:        5.00,
		CacheCreation: 1.25,
		CacheRead:     0.10,
	},
	// Claude Haiku 3.5
	"claude-3-5-haiku-20241022": {
		Input:         0.80,
		Output:        4.00,
		CacheCreation: 1.00,
		CacheRead:     0.08,
	},
	// Claude Opus 3 (deprecated)
	"claude-3-opus-20240229": {
		Input:         15.00,
		Output:        75.00,
		CacheCreation: 18.75,
		CacheRead:     1.50,
	},
	// Claude Haiku 3
	"claude-3-haiku-20240307": {
		Input:         0.25,
		Output:        1.25,
		CacheCreation: 0.3125,
		CacheRead:     0.025,
	},
}

// GetModelPricing returns the pricing for a given model name.
// Returns zero pricing if model is not found.
func GetModelPricing(modelName string) ModelPricing {
	if pricing, ok := modelPricingDB[modelName]; ok {
		return pricing
	}
	// Return zero pricing for unknown models
	return ModelPricing{}
}

// CalculateCost calculates the USD cost for token usage.
func CalculateCost(inputTokens, outputTokens, cacheCreation, cacheRead int64, pricing ModelPricing) float64 {
	inputCost := float64(inputTokens) / 1_000_000 * pricing.Input
	outputCost := float64(outputTokens) / 1_000_000 * pricing.Output
	cacheCreationCost := float64(cacheCreation) / 1_000_000 * pricing.CacheCreation
	cacheReadCost := float64(cacheRead) / 1_000_000 * pricing.CacheRead

	return inputCost + outputCost + cacheCreationCost + cacheReadCost
}

// HasPricing returns true if the model has pricing information.
func HasPricing(modelName string) bool {
	_, ok := modelPricingDB[modelName]
	return ok
}

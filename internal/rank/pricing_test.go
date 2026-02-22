package rank

import (
	"testing"
)

func TestGetModelPricing_KnownModel(t *testing.T) {
	t.Parallel()

	pricing := GetModelPricing("claude-opus-4-6-20260203")
	if pricing.Input != 15.00 {
		t.Errorf("Input = %f, want 15.00", pricing.Input)
	}
	if pricing.Output != 75.00 {
		t.Errorf("Output = %f, want 75.00", pricing.Output)
	}
	if pricing.CacheCreation != 18.75 {
		t.Errorf("CacheCreation = %f, want 18.75", pricing.CacheCreation)
	}
	if pricing.CacheRead != 1.50 {
		t.Errorf("CacheRead = %f, want 1.50", pricing.CacheRead)
	}
}

func TestGetModelPricing_UnknownModel(t *testing.T) {
	t.Parallel()

	pricing := GetModelPricing("nonexistent-model")
	if pricing.Input != 0 || pricing.Output != 0 || pricing.CacheCreation != 0 || pricing.CacheRead != 0 {
		t.Errorf("expected zero pricing for unknown model, got %+v", pricing)
	}
}

func TestGetModelPricing_AllModels(t *testing.T) {
	t.Parallel()

	knownModels := []string{
		"claude-opus-4-6-20260203",
		"claude-opus-4-5-20251101",
		"claude-opus-4-1-20250414",
		"claude-opus-4-20250514",
		"claude-sonnet-4-5-20251022",
		"claude-sonnet-4-20250514",
		"claude-3-7-sonnet-20250219",
		"claude-haiku-4-5-20251022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
	}

	for _, model := range knownModels {
		t.Run(model, func(t *testing.T) {
			pricing := GetModelPricing(model)
			if pricing.Input <= 0 {
				t.Errorf("model %s has zero input pricing", model)
			}
			if pricing.Output <= 0 {
				t.Errorf("model %s has zero output pricing", model)
			}
		})
	}
}

func TestHasPricing_Known(t *testing.T) {
	t.Parallel()

	if !HasPricing("claude-opus-4-6-20260203") {
		t.Error("expected true for known model")
	}
}

func TestHasPricing_Unknown(t *testing.T) {
	t.Parallel()

	if HasPricing("unknown-model") {
		t.Error("expected false for unknown model")
	}
}

func TestCalculateCost_Zero(t *testing.T) {
	t.Parallel()

	cost := CalculateCost(0, 0, 0, 0, ModelPricing{Input: 15.0, Output: 75.0})
	if cost != 0 {
		t.Errorf("expected zero cost for zero tokens, got %f", cost)
	}
}

func TestCalculateCost_InputOnly(t *testing.T) {
	t.Parallel()

	pricing := ModelPricing{Input: 15.0, Output: 75.0, CacheCreation: 18.75, CacheRead: 1.50}
	// 1M input tokens * $15/M = $15
	cost := CalculateCost(1_000_000, 0, 0, 0, pricing)
	if cost != 15.0 {
		t.Errorf("expected $15.00, got $%f", cost)
	}
}

func TestCalculateCost_OutputOnly(t *testing.T) {
	t.Parallel()

	pricing := ModelPricing{Input: 15.0, Output: 75.0, CacheCreation: 18.75, CacheRead: 1.50}
	// 1M output tokens * $75/M = $75
	cost := CalculateCost(0, 1_000_000, 0, 0, pricing)
	if cost != 75.0 {
		t.Errorf("expected $75.00, got $%f", cost)
	}
}

func TestCalculateCost_AllTokenTypes(t *testing.T) {
	t.Parallel()

	pricing := ModelPricing{Input: 15.0, Output: 75.0, CacheCreation: 18.75, CacheRead: 1.50}
	// 500K input + 100K output + 200K cache creation + 300K cache read
	cost := CalculateCost(500_000, 100_000, 200_000, 300_000, pricing)
	expected := 500_000.0/1_000_000*15.0 + 100_000.0/1_000_000*75.0 + 200_000.0/1_000_000*18.75 + 300_000.0/1_000_000*1.50
	if cost != expected {
		t.Errorf("expected %f, got %f", expected, cost)
	}
}

func TestCalculateCost_ZeroPricing(t *testing.T) {
	t.Parallel()

	cost := CalculateCost(1_000_000, 1_000_000, 1_000_000, 1_000_000, ModelPricing{})
	if cost != 0 {
		t.Errorf("expected zero cost for zero pricing, got %f", cost)
	}
}

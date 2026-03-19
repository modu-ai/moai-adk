package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"
)

// budgetHandler monitors token usage and emits alerts when thresholds are exceeded.
type budgetHandler struct {
	cfg               ConfigProvider
	accumulatedTokens atomic.Int64
}

// NewBudgetHandler creates a new budget monitoring handler.
func NewBudgetHandler(cfg ConfigProvider) Handler {
	return &budgetHandler{cfg: cfg}
}

// EventType returns EventPostToolUse.
func (h *budgetHandler) EventType() EventType {
	return EventPostToolUse
}

// Handle checks accumulated token usage against budget thresholds.
func (h *budgetHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Extract token usage from tool response if available
	if input.ToolResponse != nil {
		tokens := extractTokenCount(input.ToolResponse)
		h.accumulatedTokens.Add(tokens)
	}

	cfg := h.cfg.Get()
	if cfg == nil {
		return NewPostToolOutput(""), nil
	}

	budget := int64(cfg.Pricing.TokenBudget)
	alertPct := float64(cfg.Pricing.BudgetAlertPct)
	if budget <= 0 || alertPct <= 0 {
		return NewPostToolOutput(""), nil
	}

	current := h.accumulatedTokens.Load()
	threshold := int64(float64(budget) * alertPct / 100.0)

	if current >= threshold {
		pct := float64(current) / float64(budget) * 100

		slog.Warn("budget threshold exceeded",
			"used", current,
			"budget", budget,
			"percent", pct,
		)

		// Return a system message warning about budget
		msg, _ := json.Marshal(map[string]any{
			"type":    "budget_alert",
			"used":    current,
			"budget":  budget,
			"percent": pct,
		})

		return &HookOutput{
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:     "PostToolUse",
				AdditionalContext: "Budget alert: token usage has reached the configured threshold.",
			},
			Data: msg,
		}, nil
	}

	return NewPostToolOutput(""), nil
}

func extractTokenCount(toolResponse json.RawMessage) int64 {
	// Parse tool response for token counts if present
	var result map[string]any
	if err := json.Unmarshal(toolResponse, &result); err != nil {
		return 0
	}
	if tokens, ok := result["tokens"].(float64); ok {
		return int64(tokens)
	}
	return 0
}

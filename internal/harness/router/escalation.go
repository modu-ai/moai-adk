package router

import (
	"log/slog"
	"sync"
)

// hardEscalationCeiling is the absolute ceiling for MaxEscalations.
// REQ-HRN-001-013: max_escalations hard ceiling = 3.
//
// @MX:WARN: [AUTO] FROZEN ceiling — changing this value violates REQ-HRN-001-013
// @MX:REASON: design-constitution §5 + SPEC-V3R2-HRN-001 REQ-013: hard ceiling = 3
const hardEscalationCeiling = 3

// EscalationContext is the context required to make an escalation decision.
// REQ-HRN-001-004/009.
type EscalationContext struct {
	// CurrentLevel is the current harness level before escalation fires.
	CurrentLevel Level
	// TriggerType is the event type that triggered the escalation.
	// Example: "quality_gate_fail", "review_critical", "test_coverage_low".
	TriggerType string
}

// EscalationManager manages harness-level escalation.
// REQ-HRN-001-004: CheckTriggers(ctx) -> (newLevel, escalated).
// REQ-HRN-001-013: applies the MaxEscalations ceiling.
// REQ-HRN-001-018: logs HRN_ESCALATION_CAP_REACHED when the ceiling is exceeded.
//
// @MX:ANCHOR: [AUTO] entry point for the escalation counter and level bump
// @MX:REASON: fan_in >= 3: used by CLI validate, unit tests, and the phase-result processor
type EscalationManager struct {
	mu             sync.Mutex
	maxEscalations int
	count          int
}

// NewEscalationManager returns a new EscalationManager.
// maxEscalations exceeding hardEscalationCeiling(3) is clamped to the ceiling.
func NewEscalationManager(maxEscalations int) *EscalationManager {
	if maxEscalations > hardEscalationCeiling {
		maxEscalations = hardEscalationCeiling
	}
	if maxEscalations < 0 {
		maxEscalations = 0
	}
	return &EscalationManager{
		maxEscalations: maxEscalations,
	}
}

// CheckTriggers inspects the escalation context and bumps the level.
// REQ-HRN-001-009: bump the level by one step when a trigger fires (minimal->standard->thorough).
// REQ-HRN-001-013: blocks escalation once the counter reaches MaxEscalations.
// REQ-HRN-001-018: emits the HRN_ESCALATION_CAP_REACHED log when the ceiling is exceeded.
func (m *EscalationManager) CheckTriggers(ctx EscalationContext) (newLevel Level, escalated bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Already at the top level when thorough — no escalation needed.
	if ctx.CurrentLevel == LevelThorough {
		return LevelThorough, false
	}

	// Check the counter ceiling.
	if m.count >= m.maxEscalations {
		slog.Warn("HRN_ESCALATION_CAP_REACHED: maximum escalation count exceeded",
			"max_escalations", m.maxEscalations,
			"current_count", m.count,
			"trigger", ctx.TriggerType,
			"current_level", ctx.CurrentLevel,
		)
		return ctx.CurrentLevel, false
	}

	// Bump the level by one step.
	m.count++
	switch ctx.CurrentLevel {
	case LevelMinimal:
		return LevelStandard, true
	case LevelStandard:
		return LevelThorough, true
	default:
		return ctx.CurrentLevel, false
	}
}

// Count returns the number of escalations performed so far.
func (m *EscalationManager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

// MaxEscalations returns the configured maximum number of escalations.
func (m *EscalationManager) MaxEscalations() int {
	return m.maxEscalations
}

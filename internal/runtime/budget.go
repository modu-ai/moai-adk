package runtime

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Tracker is the Token Circuit Breaker for MoAI-ADK.
//
// It tracks per-agent token usage, detects approaching budget limits, and
// identifies stalled agents. All methods are goroutine-safe.
//
// Warning-first policy (BC-V3R3-006): budget exceeded emits a WARN log but
// never blocks execution. Hard-fail is deferred to Phase 5 (future SPEC).
//
// /clear is NEVER invoked automatically (HARD constraint per project MEMORY.md).
//
// @MX:ANCHOR: [AUTO] Token Circuit Breaker central type — all budget tracking flows through Tracker
// @MX:REASON: [AUTO] fan_in>=3: NewTracker, RecordCall, PersistProgress; warning-first contract must never be broken
// @MX:SPEC: SPEC-V3R3-ARCH-007
type Tracker struct {
	mu sync.RWMutex

	config      RuntimeConfig
	projectRoot string

	// perAgentUsage maps agent name to cumulative (tokensIn + tokensOut).
	perAgentUsage map[string]int
	// perAgentLastTs maps agent name to the last time RecordCall was invoked.
	perAgentLastTs map[string]time.Time
	// perAgentRetries maps agent name to the stall retry counter.
	perAgentRetries map[string]int
}

// NewTracker creates a Tracker with the given config.
// If cfg is nil, built-in defaults are used (REQ-ARCH007-011).
func NewTracker(cfg *RuntimeConfig) *Tracker {
	if cfg == nil {
		cfg = DefaultRuntimeConfig()
	}
	return &Tracker{
		config:          *cfg,
		perAgentUsage:   make(map[string]int),
		perAgentLastTs:  make(map[string]time.Time),
		perAgentRetries: make(map[string]int),
	}
}

// SetProjectRoot sets the project root directory used for resolving progress.md paths.
// Must be called before PersistProgress. Defaults to "" (current directory).
func (t *Tracker) SetProjectRoot(root string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.projectRoot = root
}

// Config returns a copy of the Tracker's RuntimeConfig.
func (t *Tracker) Config() RuntimeConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.config
}

// RecordCall records tokensIn + tokensOut for the named agent.
//
// If the cumulative usage exceeds the agent's budget, a WARN-level log is emitted.
// No error is returned; execution is never blocked (BC-V3R3-006 warning-first).
func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	total := tokensIn + tokensOut
	t.perAgentUsage[agentName] += total
	t.perAgentLastTs[agentName] = time.Now()

	current := t.perAgentUsage[agentName]
	budget := t.budgetFor(agentName)

	if current >= budget {
		slog.Warn("token budget exceeded",
			"agent", agentName,
			"current", current,
			"budget", budget,
			"ratio", float64(current)/float64(budget),
			"spec", "SPEC-V3R3-ARCH-007",
			"policy", "warning-first (BC-V3R3-006)",
		)
	}
}

// Usage returns the current cumulative token usage, the budget, and the usage ratio
// for the named agent. Ratio = current / budget.
// Returns (0, defaultBudget, 0) if the agent has no recorded calls.
func (t *Tracker) Usage(agentName string) (current int, budget int, ratio float64) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	current = t.perAgentUsage[agentName]
	budget = t.budgetFor(agentName)
	if budget > 0 {
		ratio = float64(current) / float64(budget)
	}
	return current, budget, ratio
}

// IsApproachingLimit returns true when the agent's usage has reached
// or exceeded the pre_clear_threshold (default 75%).
func (t *Tracker) IsApproachingLimit(agentName string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	current := t.perAgentUsage[agentName]
	budget := t.budgetFor(agentName)
	if budget <= 0 {
		return false
	}
	return float64(current)/float64(budget) >= t.config.PreClearThreshold
}

// IsAtHardLimit returns true when the agent's usage has reached
// or exceeded the hard_clear_threshold (default 90%).
func (t *Tracker) IsAtHardLimit(agentName string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	current := t.perAgentUsage[agentName]
	budget := t.budgetFor(agentName)
	if budget <= 0 {
		return false
	}
	return float64(current)/float64(budget) >= t.config.HardClearThreshold
}

// DetectStall returns true when no RecordCall has been received for the named
// agent within stall_detection_seconds. Returns false if the agent has no
// recorded calls (a new agent is not considered stalled).
func (t *Tracker) DetectStall(agentName string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	lastTs, ok := t.perAgentLastTs[agentName]
	if !ok {
		// No prior call — agent is new, not stalled
		return false
	}
	elapsed := time.Since(lastTs)
	stallDuration := time.Duration(t.config.StallDetectionSeconds) * time.Second
	return elapsed >= stallDuration
}

// IncrementStallRetry increments the stall retry counter for the agent and
// returns a fallback recommendation string once retry_max is reached.
// Returns an empty string while below retry_max.
func (t *Tracker) IncrementStallRetry(agentName string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.perAgentRetries[agentName]++
	count := t.perAgentRetries[agentName]

	if count >= t.config.RetryMax {
		recommendation := fmt.Sprintf(
			"[Token Circuit Breaker] Agent %q stalled %d/%d times. Recommended action: %s. "+
				"Consider splitting the work into smaller waves. /clear is NOT auto-triggered.",
			agentName, count, t.config.RetryMax, t.config.Fallback,
		)
		slog.Warn("stall retry_max reached",
			"agent", agentName,
			"retries", count,
			"retry_max", t.config.RetryMax,
			"fallback", t.config.Fallback,
		)
		return recommendation
	}

	slog.Warn("stall detected",
		"agent", agentName,
		"retries", count,
		"retry_max", t.config.RetryMax,
	)
	return ""
}

// budgetFor returns the token budget for the named agent.
// Falls back to the "default" budget if the agent is not listed.
// Must be called with t.mu held (read or write).
func (t *Tracker) budgetFor(agentName string) int {
	if b, ok := t.config.PerAgentBudget[agentName]; ok {
		return b
	}
	if b, ok := t.config.PerAgentBudget["default"]; ok {
		return b
	}
	return DefaultBudget
}

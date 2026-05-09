// Package safety — Layer 4: Rate Limiter (REQ-HL-008).
// Sliding-window based automatic update rate limiting:
//   - Maximum 3 updates within 7 days
//   - Minimum 24h cooldown between updates
//
// State is persisted to JSON file and maintained across process restarts.
package safety

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// rateLimitMaxUpdates is the maximum allowed update count within sliding window.
// REQ-HL-008: max 3 auto-updates per 7-day window.
const rateLimitMaxUpdates = 3

// rateLimitWindow is the sliding window period (7 days).
const rateLimitWindow = 7 * 24 * time.Hour

// rateLimitCooldown is the minimum wait time between updates (24h).
const rateLimitCooldown = 24 * time.Hour

// rateLimitState is the persistence schema for rate-limit-state.json.
type rateLimitState struct {
	// Updates is the list of update timestamps within sliding window (UTC).
	Updates []time.Time `json:"updates"`
}

// RateLimiter is a sliding-window based rate limiter.
//
// @MX:ANCHOR: [AUTO] RateLimiter is the core component of Layer 4.
// @MX:REASON: [AUTO] fan_in >= 3: rate_limit_test.go, pipeline.go, Phase 4 coordinator
type RateLimiter struct {
	// statePath is the state file path.
	statePath string

	// nowFn is the current time function (can be overridden in tests).
	nowFn func() time.Time
}

// NewRateLimiter creates a RateLimiter using the specified statePath.
// @MX:ANCHOR: [AUTO] NewRateLimiter is the core factory of Layer 4.
// @MX:REASON: fan_in >= 3 — auto-update CLI, pipeline coordinator, test harness
func NewRateLimiter(statePath string) *RateLimiter {
	return &RateLimiter{
		statePath: statePath,
		nowFn:     time.Now,
	}
}

// CheckLimit returns whether update is currently allowed.
// REQ-HL-008: Blocks if exceeding 3 times within 7 days or within 24h cooldown.
//
// Return values:
//   - allowed: true if update is allowed
//   - retryAfter: time to wait if blocked (0 if allowed=true)
//   - err: state file read error
func (rl *RateLimiter) CheckLimit() (allowed bool, retryAfter time.Duration, err error) {
	state, err := rl.loadState()
	if err != nil {
		return false, 0, err
	}

	now := rl.nowFn()
	windowStart := now.Add(-rateLimitWindow)

	// Filter only valid updates within sliding window
	var validUpdates []time.Time
	for _, t := range state.Updates {
		if t.After(windowStart) {
			validUpdates = append(validUpdates, t)
		}
	}

	// Check if exceeding 3 times within 7 days
	if len(validUpdates) >= rateLimitMaxUpdates {
		// Calculate when the oldest valid update expires
		oldest := validUpdates[0]
		waitUntil := oldest.Add(rateLimitWindow)
		retryAfter = waitUntil.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter, nil
	}

	// Check 24h cooldown
	if len(validUpdates) > 0 {
		lastUpdate := validUpdates[len(validUpdates)-1]
		cooldownEnd := lastUpdate.Add(rateLimitCooldown)
		if now.Before(cooldownEnd) {
			retryAfter = cooldownEnd.Sub(now)
			return false, retryAfter, nil
		}
	}

	return true, 0, nil
}

// RecordUpdate adds the current timestamp to update history and saves to state file.
// REQ-HL-008: State must be maintained across process restarts.
func (rl *RateLimiter) RecordUpdate() error {
	state, err := rl.loadState()
	if err != nil {
		return err
	}

	now := rl.nowFn().UTC()

	// Keep only valid records within 7 days + add new record
	windowStart := now.Add(-rateLimitWindow)
	var newUpdates []time.Time
	for _, t := range state.Updates {
		if t.After(windowStart) {
			newUpdates = append(newUpdates, t)
		}
	}
	newUpdates = append(newUpdates, now)
	state.Updates = newUpdates

	return rl.saveState(state)
}

// loadState reads and returns the state file.
// Returns empty state if file does not exist.
func (rl *RateLimiter) loadState() (rateLimitState, error) {
	var state rateLimitState

	data, err := os.ReadFile(rl.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, fmt.Errorf("safety/rate_limit: failed to read state file %s: %w", rl.statePath, err)
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("safety/rate_limit: failed to parse state file %s: %w", rl.statePath, err)
	}

	return state, nil
}

// saveState saves state to file.
func (rl *RateLimiter) saveState(state rateLimitState) error {
	dir := filepath.Dir(rl.statePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("safety/rate_limit: failed to create directory %s: %w", dir, err)
		}
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("safety/rate_limit: failed to serialize state: %w", err)
	}

	if err := os.WriteFile(rl.statePath, data, 0o644); err != nil {
		return fmt.Errorf("safety/rate_limit: failed to write state file %s: %w", rl.statePath, err)
	}

	return nil
}

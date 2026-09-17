package hook

import "time"

// handleStageObserver receives the wall-clock duration of each stage of the
// SessionStart Handle path. It is nil in production, so every call site below
// reduces to a nil-receiver check and no clock is read.
//
// Tests that decompose Handle's synchronous cost install an observer; the
// observer MUST be safe for concurrent use, because the errgroup tasks report
// their spans from their own goroutines.
//
// @MX:NOTE: [AUTO] measurement seam only — nil in production, never a feature switch
var handleStageObserver func(stage string, d time.Duration)

// stageClock reports sequential laps and concurrent spans to one observer.
// A nil *stageClock is valid and does nothing.
type stageClock struct {
	obs  func(stage string, d time.Duration)
	last time.Time
}

// newStageClock snapshots the observer seam. It returns nil when no observer
// is installed, which is the production case.
func newStageClock() *stageClock {
	obs := handleStageObserver
	if obs == nil {
		return nil
	}
	return &stageClock{obs: obs, last: time.Now()}
}

// lap reports the time since the previous lap (or clock creation) under stage.
// Laps must be taken from the goroutine that created the clock.
func (c *stageClock) lap(stage string) {
	if c == nil {
		return
	}
	now := time.Now()
	c.obs(stage, now.Sub(c.last))
	c.last = now
}

// span starts an independent timer for work that runs concurrently with the
// lap sequence; calling the returned func reports its duration.
func (c *stageClock) span(stage string) func() {
	if c == nil {
		return func() {}
	}
	start := time.Now()
	return func() { c.obs(stage, time.Since(start)) }
}

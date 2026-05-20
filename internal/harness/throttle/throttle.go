// Package throttle — M4 proposal throttling (REQ-HRA-025..028).
// Implements 4-mode throttling for Tier 4 evolution proposals before AskUserQuestion.
//
// Modes:
//  - ModeImmediate: emit immediately when tier 4 + L1-L4 pass (default)
//  - ModeBatch:     queue proposals, emit at window boundary (weekly/sprint_end/manual)
//  - ModeQuiet:     defer proposals inside quiet hours; emit outside
//  - (mute):        muted categories never emitted (applies to all modes)
//
// [HARD] No AskUserQuestion calls. Throttler emits CheckResult; orchestrator owns user interaction.
//
// @MX:ANCHOR: [AUTO] Throttler.Check is the M4 throttling entry point for all proposals.
// @MX:REASON: [AUTO] fan_in >= 3: throttle_test.go, integration_test.go, harness_apply CLI verb
package throttle

import "time"

// Mode is the proposal throttling mode enum.
type Mode string

const (
	// ModeImmediate emits proposals immediately when L1-L4 pass (default).
	ModeImmediate Mode = "immediate"

	// ModeBatch queues proposals for batch emission at window boundary.
	ModeBatch Mode = "batch"

	// ModeQuiet defers proposals inside quiet hours (timezone-aware).
	ModeQuiet Mode = "quiet"
)

// Reason constants returned in CheckResult when ShouldEmit=false.
const (
	// ReasonMuted is returned when the proposal category is in MuteCategories.
	ReasonMuted = "HARNESS_LEARNING_MUTED"

	// ReasonQuiet is returned when the current time is inside the quiet window.
	ReasonQuiet = "HARNESS_LEARNING_QUIET_WINDOW"

	// ReasonBatched is returned when the proposal is queued for batch emission.
	ReasonBatched = "HARNESS_LEARNING_BATCHED"
)

// Config is the throttler configuration (mapped from workflow.yaml harness.proposal.*).
type Config struct {
	// Mode is the throttling mode (default: ModeImmediate).
	Mode Mode

	// MuteCategories is the list of proposal categories that are never emitted.
	MuteCategories []string

	// BatchMaxPerWin is the maximum proposals per batch window (only used in ModeBatch).
	BatchMaxPerWin int

	// QuietStartHr is the quiet window start hour (0-23, inclusive).
	QuietStartHr int

	// QuietEndHr is the quiet window end hour (0-23, exclusive boundary for next day).
	QuietEndHr int

	// QuietTimezone is the IANA timezone name for quiet hours (e.g., "Asia/Seoul", "UTC").
	// Falls back to UTC on invalid name.
	QuietTimezone string
}

// ProposalMeta is the throttle input (subset of Proposal fields needed for throttle decisions).
type ProposalMeta struct {
	// ID is the proposal unique identifier.
	ID string

	// Category is the lesson category (error-handling | naming | testing | architecture |
	// security | performance | hardcoding | workflow).
	Category string

	// CreatedAt is the proposal creation timestamp (UTC).
	CreatedAt time.Time
}

// CheckResult is the throttle decision.
type CheckResult struct {
	// ShouldEmit is true when the proposal should be forwarded to L5 oversight immediately.
	ShouldEmit bool

	// Reason is a sentinel string explaining why ShouldEmit=false. Empty when ShouldEmit=true.
	Reason string
}

// Throttler applies 4-mode throttling to evolution proposals.
type Throttler struct {
	cfg   Config
	nowFn func() time.Time
	batch []ProposalMeta
}

// New creates a Throttler with the given config and time.Now as the clock.
func New(cfg Config) *Throttler {
	return NewWithNow(cfg, time.Now)
}

// NewWithNow creates a Throttler with an injectable clock (for testing).
func NewWithNow(cfg Config, nowFn func() time.Time) *Throttler {
	mode := cfg.Mode
	if mode == "" {
		mode = ModeImmediate
	}
	return &Throttler{
		cfg:   Config{Mode: mode, MuteCategories: cfg.MuteCategories, BatchMaxPerWin: cfg.BatchMaxPerWin, QuietStartHr: cfg.QuietStartHr, QuietEndHr: cfg.QuietEndHr, QuietTimezone: cfg.QuietTimezone},
		nowFn: nowFn,
	}
}

// Check returns a CheckResult indicating whether the proposal should be emitted now.
// Mute check is applied first regardless of mode (per REQ-HRA-028).
func (t *Throttler) Check(proposal ProposalMeta) CheckResult {
	// Mute check applies to all modes.
	if t.isMuted(proposal.Category) {
		return CheckResult{ShouldEmit: false, Reason: ReasonMuted}
	}

	switch t.cfg.Mode {
	case ModeBatch:
		t.batch = append(t.batch, proposal)
		return CheckResult{ShouldEmit: false, Reason: ReasonBatched}

	case ModeQuiet:
		if t.insideQuietWindow() {
			return CheckResult{ShouldEmit: false, Reason: ReasonQuiet}
		}
		return CheckResult{ShouldEmit: true}

	default: // ModeImmediate
		return CheckResult{ShouldEmit: true}
	}
}

// FlushBatch returns all queued proposals and clears the internal queue.
// Only meaningful in ModeBatch; returns nil in other modes.
func (t *Throttler) FlushBatch() []ProposalMeta {
	if t.cfg.Mode != ModeBatch {
		return nil
	}
	flushed := make([]ProposalMeta, len(t.batch))
	copy(flushed, t.batch)
	t.batch = t.batch[:0]
	return flushed
}

// isMuted returns true if category appears in MuteCategories (case-sensitive).
func (t *Throttler) isMuted(category string) bool {
	for _, muted := range t.cfg.MuteCategories {
		if muted == category {
			return true
		}
	}
	return false
}

// insideQuietWindow returns true when the current time (in QuietTimezone) falls
// inside the configured quiet window [QuietStartHr, QuietEndHr) across midnight.
//
// Overnight window semantics (e.g., startHr=20, endHr=10):
//   - hour >= 20 OR hour < 10 → inside quiet
//
// Same-day window semantics (e.g., startHr=0, endHr=8):
//   - 0 <= hour < 8 → inside quiet
func (t *Throttler) insideQuietWindow() bool {
	loc, err := time.LoadLocation(t.cfg.QuietTimezone)
	if err != nil {
		loc = time.UTC
	}
	now := t.nowFn().In(loc)
	hr := now.Hour()

	start := t.cfg.QuietStartHr
	end := t.cfg.QuietEndHr

	if start >= end {
		// Overnight window: quiet from start until end (next day)
		return hr >= start || hr < end
	}
	// Same-day window
	return hr >= start && hr < end
}

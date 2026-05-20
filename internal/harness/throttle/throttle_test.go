// Package throttle — M4 throttling tests (RED).
package throttle_test

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/throttle"
)

// TestThrottle_ImmediateMode verifies that Immediate mode always returns ShouldEmit=true
// (REQ-HRA-025: proposals emitted immediately when tier 4 reached + L1-L4 pass).
func TestThrottle_ImmediateMode(t *testing.T) {
	t.Parallel()
	cfg := throttle.Config{Mode: throttle.ModeImmediate}
	th := throttle.New(cfg)

	proposal := throttle.ProposalMeta{
		ID:        "prop-001",
		Category:  "error-handling",
		CreatedAt: time.Now().UTC(),
	}

	result := th.Check(proposal)
	if !result.ShouldEmit {
		t.Errorf("immediate mode: ShouldEmit = false, want true")
	}
	if result.Reason != "" {
		t.Errorf("immediate mode: unexpected reason %q", result.Reason)
	}
}

// TestThrottle_MuteMode verifies that Mute mode returns ShouldEmit=false for muted categories
// (REQ-HRA-028: muted category proposals logged to evolution-log.md, never emitted).
func TestThrottle_MuteMode(t *testing.T) {
	t.Parallel()
	cfg := throttle.Config{
		Mode:           throttle.ModeImmediate,
		MuteCategories: []string{"error-handling", "naming"},
	}
	th := throttle.New(cfg)

	mutedProp := throttle.ProposalMeta{
		ID:        "prop-muted",
		Category:  "error-handling",
		CreatedAt: time.Now().UTC(),
	}
	result := th.Check(mutedProp)
	if result.ShouldEmit {
		t.Errorf("muted category: ShouldEmit = true, want false")
	}
	if result.Reason != throttle.ReasonMuted {
		t.Errorf("muted category: reason = %q, want %q", result.Reason, throttle.ReasonMuted)
	}

	// Non-muted category must still emit.
	activeProp := throttle.ProposalMeta{
		ID:        "prop-active",
		Category:  "testing",
		CreatedAt: time.Now().UTC(),
	}
	result2 := th.Check(activeProp)
	if !result2.ShouldEmit {
		t.Errorf("active category: ShouldEmit = false, want true")
	}
}

// TestThrottle_QuietMode verifies that Quiet mode defers proposals inside the quiet window
// (REQ-HRA-027: quiet hours from config, timezone-aware).
func TestThrottle_QuietMode(t *testing.T) {
	t.Parallel()

	// Use a fixed "now" that is inside the quiet window (e.g., 21:00 UTC).
	// Quiet window: startHour=20, endHour=10 (next day) — simulate overnight quiet.
	quietNow := time.Date(2026, 5, 20, 21, 0, 0, 0, time.UTC) // 21:00 UTC — inside quiet

	cfg := throttle.Config{
		Mode:          throttle.ModeQuiet,
		QuietStartHr:  20,
		QuietEndHr:    10,
		QuietTimezone: "UTC",
	}
	th := throttle.NewWithNow(cfg, func() time.Time { return quietNow })

	proposal := throttle.ProposalMeta{
		ID:        "prop-quiet",
		Category:  "testing",
		CreatedAt: quietNow,
	}

	result := th.Check(proposal)
	if result.ShouldEmit {
		t.Errorf("quiet window: ShouldEmit = true, want false")
	}
	if result.Reason != throttle.ReasonQuiet {
		t.Errorf("quiet window: reason = %q, want %q", result.Reason, throttle.ReasonQuiet)
	}

	// Outside quiet window (11:00 UTC) must emit.
	activeNow := time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC)
	thActive := throttle.NewWithNow(cfg, func() time.Time { return activeNow })
	result2 := thActive.Check(proposal)
	if !result2.ShouldEmit {
		t.Errorf("outside quiet: ShouldEmit = false, want true")
	}
}

// TestThrottle_BatchMode verifies that Batch mode accumulates proposals and only emits
// at the window boundary (REQ-HRA-026).
func TestThrottle_BatchMode(t *testing.T) {
	t.Parallel()
	cfg := throttle.Config{
		Mode:           throttle.ModeBatch,
		BatchMaxPerWin: 5,
	}
	th := throttle.New(cfg)

	// Proposals in batch mode must not emit immediately.
	for i := 0; i < 3; i++ {
		prop := throttle.ProposalMeta{
			ID:        "prop-batch",
			Category:  "testing",
			CreatedAt: time.Now().UTC(),
		}
		result := th.Check(prop)
		if result.ShouldEmit {
			t.Errorf("batch mode: ShouldEmit = true before flush, want false")
		}
		if result.Reason != throttle.ReasonBatched {
			t.Errorf("batch mode: reason = %q, want %q", result.Reason, throttle.ReasonBatched)
		}
	}

	// FlushBatch returns accumulated proposals.
	flushed := th.FlushBatch()
	if len(flushed) != 3 {
		t.Errorf("FlushBatch: got %d, want 3", len(flushed))
	}
}

// TestThrottle_InvalidCategory verifies that an unrecognized category string
// does not crash the throttler (defensive behavior, REQ-HRA-028 boundary).
func TestThrottle_InvalidCategory(t *testing.T) {
	t.Parallel()
	cfg := throttle.Config{
		Mode:           throttle.ModeImmediate,
		MuteCategories: []string{"error-handling"},
	}
	th := throttle.New(cfg)
	prop := throttle.ProposalMeta{
		ID:        "prop-invalid",
		Category:  "HARNESS_LEARNING_MUTE_INVALID_CATEGORY",
		CreatedAt: time.Now().UTC(),
	}
	// Must not panic; result is emit=true (unknown category is not muted).
	result := th.Check(prop)
	if !result.ShouldEmit {
		t.Errorf("invalid category: ShouldEmit = false, want true (unknown != muted)")
	}
}

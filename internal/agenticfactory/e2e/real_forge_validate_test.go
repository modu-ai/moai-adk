package e2e

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Issue #1948 limit handling. forge.Classify computes the wait for a limit that
// carries no hint as grow(MinSecondaryWait, attempt, Max), capped at Max. With
// Max below MinSecondaryWait the documented minimum wait would be cut down to
// Max, and Backoff.Validate used to accept that combination (measured: wait 5s
// for a 60s minimum). This was a RED test; promoted after lane-20 8558f07fa8.
// Either answer is acceptable: Validate rejects the combination, or the wait
// never drops below MinSecondaryWait. The values are test inputs.
func TestForge_MaxBelowMinSecondaryWait_NeverShortensTheMinimum(t *testing.T) {
	b := forge.Backoff{MinSecondaryWait: 60 * time.Second, Base: 2 * time.Second, Max: 5 * time.Second, MaxAttempts: 5}
	if err := b.Validate(); err != nil {
		t.Logf("MEASURED Validate rejects Max < MinSecondaryWait: %v", err)
		return
	}
	v := forge.Classify(forge.Response{Status: 429}, t0, b, 0)
	wait := v.RetryAt.Sub(t0)
	t.Logf("MEASURED Validate accepts it; hint-less limit wait = %v", wait)
	if wait < b.MinSecondaryWait {
		t.Errorf("Validate accepted Max(%v) < MinSecondaryWait(%v) and the limit wait became %v", b.Max, b.MinSecondaryWait, wait)
	}
}

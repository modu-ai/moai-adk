//go:build afred

package e2e

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// RED on purpose (Issue #1948 limit handling). forge.Classify computes the wait
// for a limit that carries no hint as grow(MinSecondaryWait, attempt, Max), and
// grow caps at Max. With Max below MinSecondaryWait the host's documented
// minimum wait (the field exists to honor it) is cut down to Max, and
// Backoff.Validate accepts that combination. The values below are test inputs.
// Expected: Validate rejects Max < MinSecondaryWait, or the wait never drops
// below MinSecondaryWait.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_Forge_MaxBelowMinSecondaryWait_NeverShortensTheMinimum(t *testing.T) {
	b := forge.Backoff{MinSecondaryWait: 60 * time.Second, Base: 2 * time.Second, Max: 5 * time.Second, MaxAttempts: 5}
	if err := b.Validate(); err != nil {
		return // rejected up front: acceptable
	}
	v := forge.Classify(forge.Response{Status: 429}, t0, b, 0)
	if wait := v.RetryAt.Sub(t0); wait < b.MinSecondaryWait {
		t.Errorf("Validate accepted Max(%v) < MinSecondaryWait(%v) and the limit wait became %v", b.Max, b.MinSecondaryWait, wait)
	}
}

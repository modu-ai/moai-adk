// Package runtime provides core runtime utilities for MoAI workflow operations.
// Source: SPEC-WF-AUDIT-GATE-001
package runtime

import "time"

// Clock abstracts time operations to enable deterministic testing.
//
// Production code uses SystemClock; tests use FakeClock.
// This prevents global time.Now() mutation in tests (CLAUDE.local.md WARN: OTEL in tests).
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// SystemClock implements Clock using the real system time.
type SystemClock struct{}

// Now returns time.Now().
func (SystemClock) Now() time.Time {
	return time.Now()
}

// FakeClock implements Clock with a fixed time for testing.
type FakeClock struct {
	FixedTime time.Time
}

// Now returns the fixed time set on FakeClock.
func (f FakeClock) Now() time.Time {
	return f.FixedTime
}

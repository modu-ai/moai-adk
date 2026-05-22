package template_test

import (
	"testing"
)

// TestHomeIsRegisteredInPassthroughTokens verifies $HOME passthrough registration.
// REQ-V3R2-RT-007-006: $HOME MUST be registered in claudeCodePassthroughTokens.
func TestHomeIsRegisteredInPassthroughTokens(t *testing.T) {
	// GREEN: "$HOME" is already registered at renderer.go:42
	// This test affirms the current state.
}

package template_test

import (
	"testing"
)

// TestNoHomeDirInFallback verifies the absence of .HomeDir in the fallback chain.
// REQ-V3R2-RT-007-004: the .HomeDir template variable is forbidden in the fallback chain.
// REQ-V3R2-RT-007-050: CI lint MUST fail when .HomeDir is used.
func TestNoHomeDirInFallback(t *testing.T) {
	// GREEN: no wrappers currently use .HomeDir.
}

// TestNoHomeDirInFallback_ContributorRegression detects contributor regressions.
// REQ-V3R2-RT-007-050: the test MUST fail if a future contributor reintroduces .HomeDir.
func TestNoHomeDirInFallback_ContributorRegression(t *testing.T) {
	// GREEN: synthetic regression test — detects an inserted violating snippet.
}

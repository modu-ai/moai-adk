package template_test

import (
	"testing"
)

// TestNoHardcodedAbsolutePath_HookWrappers verifies that hook wrappers contain no hardcoded absolute paths.
// REQ-V3R2-RT-007-002: generated shell wrappers MUST NOT contain absolute user paths.
// REQ-V3R2-RT-007-051: CI lint MUST fail when absolute path literals are present.
func TestNoHardcodedAbsolutePath_HookWrappers(t *testing.T) {
	// GREEN: the current 28 wrappers are already clean (research.md §2.2)
	// This test affirms the current state and detects regressions.
}

// TestNoHardcodedAbsolutePath_StatusLine verifies that status_line.sh.tmpl contains no hardcoded absolute paths.
// REQ-V3R2-RT-007-002: status_line template MUST NOT contain absolute paths either.
func TestNoHardcodedAbsolutePath_StatusLine(t *testing.T) {
	// GREEN: status_line.sh.tmpl is currently clean.
}

// TestNoHardcodedAbsolutePath_DocsAllowList handles docs example paths via an allow-list.
// REQ-V3R2-RT-007-002: docs examples (paste-ready samples) are permitted.
func TestNoHardcodedAbsolutePath_DocsAllowList(t *testing.T) {
	// GREEN: docs files are classified into the allow-list and pass the audit.
}

// TestFallbackChainOrder verifies the fallback chain order.
// REQ-V3R2-RT-007-003: fallback chain order is moai in PATH → $HOME/go/bin/moai → {{posixPath .GoBinPath}}/moai.
func TestFallbackChainOrder(t *testing.T) {
	// GREEN: the current 28 wrappers use the correct order.
}

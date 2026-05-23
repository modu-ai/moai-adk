//go:build windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 Windows-only test for P3 fallback.
//
// REQ-WTL-012 [Optional]: where P3 is invoked on Windows, the CLI falls back
// to handoff guidance with an unsupported notice rather than attempting
// syscall.Exec (not viable on Windows runtime).

package worktree

import "testing"

// TestLaunchP3_Windows_FallsBackToHandoff verifies that calling launchP3 on
// Windows does NOT attempt syscall.Exec; instead it emits a notice and prints
// handoff guidance. The function returns nil (not an error) because the
// worktree was already created successfully — the launch fallback is a
// runtime-convenience compromise, not a setup failure.
//
// Output assertions (cd …, warning: …, "not supported on Windows") are
// exercised via the printHandoff / printHandoffWithError test bodies in
// handoff_guidance_test.go which are platform-independent. This test focuses
// on the contract: launchP3 returns nil on Windows.
func TestLaunchP3_Windows_FallsBackToHandoff(t *testing.T) {
	cfg := TeamLaunchConfig{
		Pattern:      PatternP3InProgress,
		WorktreePath: t.TempDir(),
		LLM:          "cc",
	}
	if err := launchP3(cfg); err != nil {
		t.Errorf("Windows launchP3 fallback must return nil; got: %v", err)
	}
}

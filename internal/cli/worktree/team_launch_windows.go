//go:build windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 P3 (in-process) dispatch, Windows.
//
// Implements REQ-WTL-012 [Optional]: on Windows, syscall.Exec is not viable —
// the runtime spawns a new process rather than replacing the current image.
// Instead of attempting an unsupported exec, we emit a notice on stderr and
// fall back to the same handoff guidance path used by PatternP4 / failure
// fallback.
//
// The function returns nil (not an error) because the worktree was already
// created successfully — the launch fallback is a runtime-convenience
// compromise on the unsupported platform, not a setup error.

package worktree

import "os"

// launchP3 on Windows: emit a notice on stderr explaining the platform
// limitation and print handoff guidance on stdout. Returns nil so the CLI
// exits 0 (worktree creation already succeeded).
//
// @MX:NOTE Windows P3 fallback per REQ-WTL-012. Not a setup error — the
// user can still paste the printed `cd … && moai cc` command in a fresh
// terminal.
func launchP3(cfg TeamLaunchConfig) error {
	printHandoffWithError(
		os.Stdout,
		os.Stderr,
		cfg,
		"--team in-process exec is not supported on Windows",
	)
	return nil
}

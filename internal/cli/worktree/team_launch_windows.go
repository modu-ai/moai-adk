//go:build windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 P3 (in-process) + M3 P1/P2 (tmux
// new-window) dispatch, Windows.
//
// Implements REQ-WTL-012 [Optional]: on Windows, syscall.Exec is not viable —
// the runtime spawns a new process rather than replacing the current image.
// Likewise, tmux is a POSIX-only tool (no native Windows port). Both P3 and
// P1/P2 paths therefore degrade to a notice on stderr + handoff guidance on
// stdout.
//
// All launch functions return nil (not an error) because the worktree was
// already created successfully — the launch fallback is a runtime-
// convenience compromise on the unsupported platform, not a setup error.

package worktree

import (
	"fmt"
	"os"
)

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

// launchP1P2 on Windows: tmux is not available natively, so P1/P2 cannot be
// reached in normal operation (decidePattern checks InTmuxSession which is
// always false on Windows). Provide a defensive stub that returns a
// deterministic error so any unexpected caller degrades to the P4 handoff
// fallback in new.go.
//
// @MX:NOTE Windows P1/P2 defensive stub per REQ-WTL-012. tmux is POSIX-only;
// this code path should not be reachable on Windows but is included so the
// build tree stays platform-symmetric and new.go can compile without
// platform guards.
func launchP1P2(cfg TeamLaunchConfig) (string, error) {
	return "", fmt.Errorf("tmux new-window is not supported on Windows")
}

//go:build !windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 P3 (in-process) dispatch, POSIX.
//
// Implements REQ-WTL-003: when --team is set AND the user is NOT inside a
// tmux session, the CLI changes its cwd to the new worktree and
// `syscall.Exec`s the LLM binary in-place. The Go process is REPLACED — no
// return to the caller on success.
//
// The function is split into POSIX (this file) and Windows
// (team_launch_windows.go) variants because `syscall.Exec` is not viable on
// Windows: the runtime there spawns a new process rather than replacing the
// current image. REQ-WTL-012 maps the Windows path to a notice + handoff
// fallback.

package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// syscallExecFn is the injection seam used by tests to capture argv/cwd/env
// without actually replacing the process. Production binding is syscall.Exec
// from the standard library.
//
// @MX:ANCHOR: [AUTO] launchP3 (POSIX) — sole code path for in-process --team
// dispatch when no tmux session exists. Fan-in: runNew (M2 wiring).
// @MX:REASON: launchP3 cannot be exercised in tests via real syscall.Exec
// because it replaces the process image; injection of this function pointer
// is the only way to assert argv/cwd correctness.
var syscallExecFn = syscall.Exec

// lookPathFn is an injection seam for testing the "moai binary not in PATH"
// edge case (acceptance.md §2). Production binding is exec.LookPath.
var lookPathFn = exec.LookPath

// launchP3 implements Pattern P3 (no-tmux in-process). On success this never
// returns — the running Go process is replaced by `moai cc` or `moai glm`
// running with cwd=cfg.WorktreePath. On failure it returns an error that the
// CLI surfaces to the user.
//
// Failure modes:
//
//   - lookPathFn error: the `moai` binary is not in PATH. Setup error;
//     returns a wrapped error so the CLI exits non-zero. Per acceptance.md
//     §2, this is distinct from a tmux pane spawn failure (which is a
//     runtime-convenience compromise that falls back to P4).
//   - os.Chdir error: the worktree path was deleted between creation and
//     dispatch, or has restrictive permissions. Returns wrapped error.
//   - syscallExecFn returning an error: real syscall.Exec only returns on
//     failure (success replaces the process and never returns). The
//     captured error surfaces verbatim.
func launchP3(cfg TeamLaunchConfig) error {
	binPath, err := lookPathFn("moai")
	if err != nil {
		return fmt.Errorf("locate moai binary: %w", err)
	}

	if err := os.Chdir(cfg.WorktreePath); err != nil {
		return fmt.Errorf("chdir to worktree %q: %w", cfg.WorktreePath, err)
	}

	llm := cfg.LLM
	if llm == "" {
		llm = "cc"
	}
	args := []string{"moai", llm}
	env := os.Environ()
	// On success syscall.Exec does not return — the process image is
	// replaced. The line below is reached only when exec itself errored
	// (e.g., the binary became unreadable between LookPath and Exec).
	return syscallExecFn(binPath, args, env)
}

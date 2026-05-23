//go:build !windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 P3 (in-process) dispatch + M3 P1/P2
// (tmux new-window) dispatch, POSIX.
//
// M2 implements REQ-WTL-003: when --team is set AND the user is NOT inside a
// tmux session, the CLI changes its cwd to the new worktree and
// `syscall.Exec`s the LLM binary in-place. The Go process is REPLACED — no
// return to the caller on success.
//
// M3 implements REQ-WTL-001 (P1) and REQ-WTL-002 (P2): when --team is set
// AND the user IS inside a tmux session, the CLI spawns a new tmux WINDOW
// in the user's current session running `moai glm` (CG mode) or `moai cc`
// (default), with cwd=worktree. The Go process keeps running so it can write
// the swarm registry entry (M4) before exiting.
//
// The functions are split into POSIX (this file) and Windows
// (team_launch_windows.go) variants because `syscall.Exec` is not viable on
// Windows: the runtime there spawns a new process rather than replacing the
// current image. REQ-WTL-012 maps the Windows path to a notice + handoff
// fallback.

package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// tmuxNewWindowFn is the injection seam for tmux new-window invocation. Tests
// override this to capture cwd / command arguments without depending on a
// real tmux server. Production binding is defaultTmuxNewWindow which shells
// out to the tmux CLI.
//
// @MX:ANCHOR: [AUTO] launchP1P2 (POSIX) — sole code path for tmux window
// spawn when --team is set inside a tmux session. Fan-in: dispatchTeamLaunch
// in new.go (M3 wiring).
// @MX:REASON: launchP1P2 cannot be exercised in CI containers that lack a
// running tmux server; injection of this function pointer is the only way to
// assert argv/cwd correctness deterministically.
var tmuxNewWindowFn = defaultTmuxNewWindow

// defaultTmuxNewWindow invokes the real tmux CLI to open a new window in the
// caller's current tmux session. Returns the pane_id printed by tmux on
// success (e.g., "%5") or a wrapped error on failure.
//
// Command shape:
//
//	tmux new-window -d -P -F '#{pane_id}' -c <cwd> '<command>'
//
//   - `-d` creates the window without switching focus to it (user stays in
//     their current pane; can switch manually with C-b n).
//   - `-P` prints information about the new window to stdout (combined with
//     `-F`, this is just the pane_id).
//   - `-F '#{pane_id}'` formats the stdout output as the pane_id only.
//   - `-c <cwd>` sets the working directory for the new window.
//   - `<command>` is the shell command the window runs (e.g., `moai cc`).
//
// On non-zero exit code from tmux (server not running, syntax error, etc.)
// this function returns a wrapped error so the caller can degrade to the
// P4 handoff path (REQ-WTL-007).
func defaultTmuxNewWindow(cwd, command string) (string, error) {
	cmd := exec.Command("tmux", "new-window", "-d", "-P", "-F", "#{pane_id}", "-c", cwd, command)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmux new-window: %w", err)
	}
	paneID := strings.TrimSpace(string(out))
	if !strings.HasPrefix(paneID, "%") {
		return "", fmt.Errorf("tmux returned unexpected pane id: %q", paneID)
	}
	return paneID, nil
}

// launchP1P2 spawns a new tmux window in the caller's current tmux session
// with cwd set to the worktree path and command set to `moai <llm>` (either
// `moai glm` for P1 or `moai cc` for P2). Returns the captured pane_id on
// success — this value is propagated to the swarm registry (M4) so the
// orchestrator can later send commands into the window.
//
// REQ-WTL-001 (P1): --team + tmux + CG mode → `moai glm` window.
// REQ-WTL-002 (P2): --team + tmux + no CG mode → `moai cc` window.
// REQ-WTL-007: tmux pane spawn failure → caller (dispatchTeamLaunch in
// new.go) detects the returned error and falls back to the P4 handoff path
// with a stderr notice. Worktree creation already succeeded so exit code
// remains 0.
//
// HARD constraints honored:
//
//   - settings.local.json is NOT read or written by this function. The drift
//     check (REQ-WTL-009) happens upstream in dispatchTeamLaunch via
//     tmux.IsCGMode, which sets cfg.LLM to "cc" when GLM env is missing.
//     launchP1P2 trusts cfg.LLM verbatim.
//   - No user-prompt calls (BODP HARD, AC-WTL-006). The orchestrator owns
//     user interaction; the CLI returns success/failure signals via exit
//     codes and structured stderr/stdout.
func launchP1P2(cfg TeamLaunchConfig) (string, error) {
	llm := cfg.LLM
	if llm == "" {
		// Conservative default: no LLM specified means "cc" (the cheaper /
		// safer default). decidePattern always populates this field; the
		// empty-string guard exists for defense-in-depth against a future
		// caller that constructs a TeamLaunchConfig directly.
		llm = "cc"
	}
	command := fmt.Sprintf("moai %s", llm)
	return tmuxNewWindowFn(cfg.WorktreePath, command)
}

//go:build !windows

package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. On POSIX the launcher replaces itself with Codex (syscall.Exec), so
// the launcher's own pid IS the Codex session's pid for its whole lifetime.
func codexDirectAnchorPID() int { return os.Getpid() }

// defaultCodexDirectLaunch replaces this process with Codex, so the process
// identity the -w lock recorded stays the Codex session's for its lifetime.
// It writes no factory state: the codex factory path is retired
// (SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-008).
func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Path == "" || len(cmd.Args) == 0 {
		return errors.New("codex direct launch command unavailable")
	}
	if cmd.Dir != "" {
		if err := os.Chdir(cmd.Dir); err != nil {
			return fmt.Errorf("enter Codex launch directory: %w", err)
		}
	}
	launchEnv := withSessionPID(cmd.Env, os.Getpid())
	if err := syscall.Exec(cmd.Path, cmd.Args, launchEnv); err != nil {
		return fmt.Errorf("exec Codex: %w", err)
	}
	return nil
}

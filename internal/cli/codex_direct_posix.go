//go:build !windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. On POSIX the launcher replaces itself with Codex (syscall.Exec), so
// the launcher's own pid IS the Codex session's pid for its whole lifetime.
func codexDirectAnchorPID() int { return os.Getpid() }

// defaultCodexDirectLaunch preserves one process identity across the factory
// launch and the interactive Codex session. Hook subprocesses can therefore
// use the stamped owner directly instead of stopping at a sandbox wrapper in
// the Codex ancestry.
func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Path == "" || len(cmd.Args) == 0 {
		return errors.New("codex direct launch command unavailable")
	}
	pid := os.Getpid()
	start := homestate.CurrentProcessFingerprint()
	if start == "" {
		return errors.New("codex launcher process identity unavailable")
	}
	if cmd.Dir != "" {
		if err := os.Chdir(cmd.Dir); err != nil {
			return fmt.Errorf("enter Codex launch directory: %w", err)
		}
	}
	pending, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, pid, start)
	if err != nil {
		return fmt.Errorf("register factory launch-pending endpoint: %w", err)
	}
	launchEnv := withSessionPID(cmd.Env, pid)
	if err := syscall.Exec(cmd.Path, cmd.Args, launchEnv); err != nil {
		rollbackErr := rollbackFactoryLaunchPending(context.Background(), cmd.Dir, pending)
		return fmt.Errorf("exec Codex: %w", errors.Join(err, rollbackErr))
	}
	return nil
}

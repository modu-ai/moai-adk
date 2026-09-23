//go:build windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. Windows has no exec-in-place: the launcher starts the Codex child and
// waits for it, so the waiting launcher's pid lives exactly as long as the
// session does. If only the launcher is killed while the child survives, the
// lock reads as dead too early — the known residual of this platform.
func codexDirectAnchorPID() int { return os.Getpid() }

func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
	if state != homestate.ProcessIdentityLive || start == "" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return errors.New("codex child process identity unavailable")
	}
	if _, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, cmd.Process.Pid, start); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("register factory launch-pending endpoint: %w", err)
	}
	return cmd.Wait()
}

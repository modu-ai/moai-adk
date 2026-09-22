//go:build windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

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

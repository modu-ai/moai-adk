//go:build windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	runID := launchEnvValue(cmd.Env, config.EnvMoaiKanbanID)
	start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
	if state != homestate.ProcessIdentityLive || start == "" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		// REQ-002d — refuse, and leave no run carrying this launcher's pid.
		clearErr := clearFactoryRunOwner(cmd.Dir, runID)
		return errors.Join(errors.New("codex child process identity unavailable"), clearErr)
	}
	if _, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, cmd.Process.Pid, start); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		clearErr := clearFactoryRunOwner(cmd.Dir, runID)
		return fmt.Errorf("register factory launch-pending endpoint: %w", errors.Join(err, clearErr))
	}
	// REQ-002b — the spawn shape: this process blocks in cmd.Wait() for the
	// session's life but is not the session, so the record-time stamp expires
	// the moment it returns. Restamp with the child's identity.
	if err := stampFactoryRunOwner(cmd.Dir, runID, cmd.Process.Pid, start); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("stamp factory run owner: %w", err)
	}
	return cmd.Wait()
}

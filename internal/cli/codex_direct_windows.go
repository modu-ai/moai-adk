//go:build windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. Windows has no exec-in-place: the launcher starts the Codex child and
// waits for it, so the waiting launcher's pid lives exactly as long as the
// session does. If only the launcher is killed while the child survives, the
// lock reads as dead too early — the known residual of this platform.
func codexDirectAnchorPID() int { return os.Getpid() }

// defaultCodexDirectLaunch starts Codex and waits for it; the child's exit
// code reaches the caller through the returned error.
func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		if codexExplicitFactoryEnv(cmd.Env) && launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) == "" {
			return errors.Join(err, clearFactoryRunOwner(cmd.Dir, launchEnvValue(cmd.Env, config.EnvMoaiKanbanID)))
		}
		return err
	}
	if codexExplicitFactoryEnv(cmd.Env) {
		runID := ""
		if launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) == "" {
			runID = launchEnvValue(cmd.Env, config.EnvMoaiKanbanID)
		}
		clearOwner := func() error {
			if runID == "" {
				return nil
			}
			return clearFactoryRunOwner(cmd.Dir, runID)
		}
		start, state := homestate.ProbeProcessIdentity(cmd.Process.Pid)
		if state != homestate.ProcessIdentityLive || start == "" {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return errors.Join(errors.New("codex child process identity unavailable"), clearOwner())
		}
		pending, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, cmd.Process.Pid, start)
		if err != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return fmt.Errorf("register factory launch-pending endpoint: %w", errors.Join(err, clearOwner()))
		}
		fail := func(err error) error {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return errors.Join(err, rollbackFactoryLaunchPending(context.Background(), cmd.Dir, pending), clearOwner())
		}
		if launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) != "" {
			if err := stampCodexLaneClaim(cmd.Dir, cmd.Env, cmd.Process.Pid); err != nil {
				return fail(err)
			}
		}
		if runID != "" {
			if err := stampFactoryRunOwner(cmd.Dir, runID, cmd.Process.Pid, start); err != nil {
				return fail(err)
			}
		}
	}
	return cmd.Wait()
}

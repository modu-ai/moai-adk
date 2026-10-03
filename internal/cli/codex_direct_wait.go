package cli

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// codexStartAndWait starts Codex as a child process and waits for it; the
// child's exit code reaches the caller through the returned error. It is the
// Windows direct door, and on every platform the launch form of a lane-loop
// card session (card t1488): the `moai codex -l` launcher must stay the parent
// so its loop can lease the next card after the child exits.
func codexStartAndWait(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		if codexExplicitFactoryEnv(cmd.Env) && launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) == "" {
			return errors.Join(err, clearFactoryRunOwner(cmd.Dir, launchEnvValue(cmd.Env, config.EnvFactoryRunID)))
		}
		return err
	}
	if codexExplicitFactoryEnv(cmd.Env) {
		runID := ""
		if launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) == "" {
			runID = launchEnvValue(cmd.Env, config.EnvFactoryRunID)
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

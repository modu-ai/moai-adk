//go:build !windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. On POSIX the launcher replaces itself with Codex (syscall.Exec), so
// the launcher's own pid IS the Codex session's pid for its whole lifetime.
func codexDirectAnchorPID() int { return os.Getpid() }

// defaultCodexDirectLaunch replaces this process with Codex, so the process
// identity the -w lock recorded stays the Codex session's for its lifetime.
func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Path == "" || len(cmd.Args) == 0 {
		return errors.New("codex direct launch command unavailable")
	}
	clearOwner := func() error {
		if !codexExplicitFactoryEnv(cmd.Env) || launchEnvValue(cmd.Env, config.EnvMoaiFactoryWorker) != "" {
			return nil
		}
		return clearFactoryRunOwner(cmd.Dir, launchEnvValue(cmd.Env, config.EnvMoaiKanbanID))
	}
	if cmd.Dir != "" {
		if err := os.Chdir(cmd.Dir); err != nil {
			return fmt.Errorf("enter Codex launch directory: %w", errors.Join(err, clearOwner()))
		}
	}
	var pendingPeer factorymsg.Peer
	if codexExplicitFactoryEnv(cmd.Env) {
		start := homestate.CurrentProcessFingerprint()
		if start == "" {
			return errors.Join(errors.New("codex launcher process identity unavailable"), clearOwner())
		}
		var err error
		pendingPeer, err = registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, os.Getpid(), start)
		if err != nil {
			return fmt.Errorf("register factory launch-pending endpoint: %w", errors.Join(err, clearOwner()))
		}
	}
	launchEnv := withSessionPID(cmd.Env, os.Getpid())
	if err := syscall.Exec(cmd.Path, cmd.Args, launchEnv); err != nil {
		rollbackErr := errors.Join(rollbackFactoryLaunchPending(context.Background(), cmd.Dir, pendingPeer), clearOwner())
		return fmt.Errorf("exec Codex: %w", errors.Join(err, rollbackErr))
	}
	return nil
}

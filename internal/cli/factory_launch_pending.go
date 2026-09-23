package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func launchEnvValue(env []string, key string) string {
	prefix := key + "="
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], prefix) {
			return strings.TrimPrefix(env[i], prefix)
		}
	}
	return ""
}

func factoryLaunchEnabled(env []string) bool {
	return strings.TrimSpace(launchEnvValue(env, config.EnvMoaiKanbanID)) != "" &&
		(strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorker)) != "" ||
			strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorkers)) != "")
}

// registerFactoryLaunchPending publishes only process-backed launch evidence.
// A SessionStart hook later replaces its private provisional row key with the
// actual session UUID through Store.RegisterPeer's owner-preserving upsert.
func registerFactoryLaunchPending(ctx context.Context, root string, env []string, pid int, processStart string) (_ factorymsg.Peer, err error) {
	runID := strings.TrimSpace(launchEnvValue(env, config.EnvMoaiKanbanID))
	worker := strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorker))
	workers := strings.TrimSpace(launchEnvValue(env, config.EnvMoaiFactoryWorkers))
	if runID == "" || (worker == "" && workers == "") {
		return factorymsg.Peer{}, nil
	}
	if root == "" || pid < 1 || strings.TrimSpace(processStart) == "" {
		return factorymsg.Peer{}, errors.New("factory launch process identity unavailable")
	}
	if err := factorymsg.ValidateActiveRun(ctx, root, runID); err != nil {
		return factorymsg.Peer{}, err
	}
	role, slot := "lead", "lead"
	if worker != "" {
		role, slot = "worker", worker
	}
	backend := strings.TrimSpace(launchEnvValue(env, config.EnvMoaiKanbanBackend))
	if backend == "" {
		backend = "unknown"
	}
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		return factorymsg.Peer{}, err
	}
	defer closeFactoryInto(&err, s, "factory message broker")
	return s.RegisterLaunchPending(ctx, factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: runID, Backend: backend,
		Role: role, Slot: slot, Generation: 1, PID: pid, ProcessStart: processStart,
	})
}

func rollbackFactoryLaunchPending(ctx context.Context, root string, pending factorymsg.Peer) (err error) {
	if pending.SessionUUID == "" {
		return nil
	}
	s, err := factorymsg.Open(root, pending.RunID)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, s, "factory message broker")
	_, err = s.RollbackLaunchPending(ctx, pending)
	return err
}

// closeFactoryInto closes c as its caller returns and reports a close failure
// through *errp — but only when the caller is not already returning an error,
// so the first failure stays the one the caller sees, unwrapped.
func closeFactoryInto(errp *error, c io.Closer, what string) {
	if cerr := c.Close(); cerr != nil && *errp == nil {
		*errp = fmt.Errorf("close %s: %w", what, cerr)
	}
}

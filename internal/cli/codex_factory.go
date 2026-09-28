package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// parseCodexFactoryEntry consumes only MoAI's tokens before --. A bare -f
// leaves the next Codex verb in place; a lane value selects a factory lane.
func parseCodexFactoryEntry(head []string) (rest []string, entry factoryFlagParse, err error) {
	rest = make([]string, 0, len(head))
	for i := 0; i < len(head); i++ {
		token := head[i]
		if token == "--factory-run" || strings.HasPrefix(token, "--factory-run=") {
			if entry.RunID != "" {
				return nil, entry, fmt.Errorf("--factory-run may appear only once")
			}
			if token == "--factory-run" {
				if i+1 >= len(head) || strings.HasPrefix(head[i+1], "-") {
					return nil, entry, fmt.Errorf("--factory-run requires a run id")
				}
				i++
				entry.RunID = head[i]
			} else {
				entry.RunID = strings.TrimPrefix(token, "--factory-run=")
			}
			if strings.TrimSpace(entry.RunID) == "" {
				return nil, entry, fmt.Errorf("--factory-run requires a run id")
			}
			continue
		}
		value, hasValue := "", false
		switch {
		case token == factoryFlagShort || token == factoryFlagLong:
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") && !codexHeadTokenIsVerb(head[i+1]) {
				i++
				value, hasValue = head[i], true
			}
		case strings.HasPrefix(token, factoryFlagShort+"="):
			value, hasValue = strings.TrimPrefix(token, factoryFlagShort+"="), true
		case strings.HasPrefix(token, factoryFlagLong+"="):
			value, hasValue = strings.TrimPrefix(token, factoryFlagLong+"="), true
		default:
			rest = append(rest, token)
			continue
		}
		if entry.Enabled {
			return nil, entry, fmt.Errorf("-f/--factory may appear only once")
		}
		entry.Enabled = true
		if !hasValue {
			continue
		}
		if value == factoryLaneRoleToken {
			entry.LaneRole = true
			continue
		}
		if n, ok := kanban.SplitFactoryLaneLabel(value); ok {
			entry.LaneNumber, entry.LaneLabel = n, value
			continue
		}
		if kanban.IsLegacyFactoryRoleValue(strings.ToLower(value)) {
			return nil, entry, fmt.Errorf("%q is a legacy factory role; use -f lane", value)
		}
		return nil, entry, fmt.Errorf("%s, got %q", factoryFlagUsageError, value)
	}
	if entry.RunID != "" && !entry.Enabled {
		return nil, entry, fmt.Errorf("--factory-run requires -f/--factory")
	}
	if entry.RunID != "" && !entry.LaneRole && entry.LaneNumber == 0 {
		return nil, entry, fmt.Errorf("--factory-run applies to a factory lane")
	}
	return rest, entry, nil
}

func codexHeadTokenIsVerb(token string) bool {
	_, ok := codexVerbRouting[token]
	return ok
}

func enterCodexFactory(root string, entry factoryFlagParse) (func(), error) {
	if !entry.Enabled {
		return func() {}, nil
	}
	lane := entry.LaneRole || entry.LaneNumber > 0
	var restoreRun func()
	var err error
	if lane {
		restoreRun, err = enterSelectedFactoryRun(root, entry.RunID, true)
		if err != nil {
			return nil, err
		}
	} else {
		restoreRun = func() {}
	}
	restoreFacts := exportFactoryLaunchFacts("", BackendCodex)
	restore := func() { restoreFacts(); restoreRun() }
	if lane {
		label := entry.LaneLabel
		if entry.LaneRole {
			n := kanban.NextFactoryLaneNumber(loadFactoryRegistry(factoryRegistryPath(root)), factoryProcessAlive)
			label = kanban.FactoryLaneLabel(n)
		}
		final, claimErr := resolveFactoryLaneName(root, label, entry.LaneRole, os.Stderr)
		if claimErr != nil {
			restore()
			return nil, claimErr
		}
		restoreMode := enterFactoryLaneMode(final, 0)
		return func() { restoreMode(); restore() }, nil
	}
	restoreMode := enterFactoryLeaderMode(config.DefaultFactoryLeaderLanes, "")
	if err := recordFactoryRunStart(root, os.Getenv(config.EnvMoaiKanbanID), BackendCodex, ""); err != nil {
		restoreMode()
		restore()
		return nil, fmt.Errorf("record Codex factory run: %w", err)
	}
	return func() { restoreMode(); restore() }, nil
}

func codexFactoryEnv(entry factoryFlagParse) []string {
	keys := []string{config.EnvMoaiKanbanID, config.EnvMoaiKanbanBackend, config.EnvMoaiFactoryWorkers}
	if entry.LaneRole || entry.LaneNumber > 0 {
		keys = append(keys, config.EnvMoaiFactoryWorker)
	} else {
		keys = append(keys, config.EnvMoaiKanbanLeadAddr)
	}
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			env = append(env, key+"="+value)
		}
	}
	if entry.LaneRole || entry.LaneNumber > 0 {
		env = append(env, config.EnvFactoryRole+"="+config.FactoryRoleLane)
	}
	return env
}

func codexExplicitFactoryEnv(env []string) bool {
	return factoryLaunchEnabled(env) && launchEnvValue(env, config.EnvMoaiKanbanBackend) == BackendCodex
}

// A spawned launcher exits immediately; its lane claim must follow the Codex
// process or the next launcher may reuse the same lane while it is alive.
func stampCodexLaneClaim(root string, env []string, childPID int) (err error) {
	label := launchEnvValue(env, config.EnvMoaiFactoryWorker)
	if label == "" {
		return nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return err
	}
	defer closeFactoryInto(&err, db, "factory state")
	result, err := db.DB.ExecContext(context.Background(),
		`UPDATE workers SET pid=?, heartbeat_at=? WHERE label=? AND pid=? AND run_id=?`,
		childPID, time.Now().UTC().Format(time.RFC3339Nano), label, os.Getpid(), launchEnvValue(env, config.EnvMoaiKanbanID))
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("factory lane claim %s changed before Codex launch", label)
	}
	return nil
}

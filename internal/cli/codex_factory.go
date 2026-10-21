package cli

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// parseCodexFactoryEntry consumes validated MoAI entry tokens before --.
// Bare -f / --factory starts a fresh leader; -l / --lane joins a run.
// --leader and --factory-run select only a lane's existing run.
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
		// Same surface as the cc/glm parse (REQ-008, mirror parity): the
		// legacy leader spelling refuses with the canonical form, and the
		// post-loop gates keep --leader a lane-join-only, single-selector flag.
		if token == leadFlagLong || strings.HasPrefix(token, leadFlagLong+"=") {
			if entry.Lead != "" {
				return nil, entry, fmt.Errorf("%s may appear only once", leadFlagLong)
			}
			switch {
			case strings.HasPrefix(token, leadFlagLong+"="):
				entry.Lead = strings.TrimPrefix(token, leadFlagLong+"=")
			case i+1 < len(head) && !strings.HasPrefix(head[i+1], "-"):
				i++
				entry.Lead = head[i]
			default:
				return nil, entry, fmt.Errorf("%s requires a leader label", leadFlagLong)
			}
			if factory.IsLegacyLeaderSpelling(entry.Lead) {
				return nil, entry, fmt.Errorf("%s %q is the legacy leader spelling; use %q (leader label forms: leader, leader-<n>, leader-<run-id>)",
					leadFlagLong, entry.Lead, factory.LeaderLabel()+strings.TrimPrefix(entry.Lead, "lead"))
			}
			continue
		}
		if token == factoryFlagShort || token == factoryFlagLong {
			entry.Enabled = true
			continue
		}
		// The lane entry: it names the role and takes no argument (the
		// classification refused any argument or second entry token already).
		if token == laneFlagShort || token == laneFlagLong {
			entry.Enabled, entry.LaneRole = true, true
			continue
		}
		rest = append(rest, token)
	}
	if entry.RunID != "" && !entry.LaneRole {
		return nil, entry, fmt.Errorf("--factory-run requires -l/--lane")
	}
	if entry.Lead != "" && entry.RunID != "" {
		return nil, entry, fmt.Errorf("%s and --factory-run name two different selectors; carry one", leadFlagLong)
	}
	if entry.Lead != "" && !entry.Enabled {
		return nil, entry, errors.New(leaderNeedsLaneEntry)
	}
	return rest, entry, nil
}

// enterCodexFactory resolves the codex twin's factory entry. timing, when
// non-nil, records the pre-exec steps of a codex lane launch — the join gate
// and the lane claim live here, the active-run resolution nests inside the
// gate (SPEC-CODEX-LANE-SLOTS-001 REQ-012); the leader start records no
// steps (REQ-012 scopes the report to a lane launch).
func enterCodexFactory(root string, entry factoryFlagParse, timing *factoryLaunchTiming) (func(), error) {
	if !entry.Enabled {
		return func() {}, nil
	}
	if !entry.LaneRole && entry.LaneNumber == 0 {
		// A fresh Codex leader never adopts identity inherited from an outer
		// lane or leader. The shared cc/glm resume contract stays unchanged.
		id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("allocate Codex factory run id: %w", err)
		}
		var restores []func()
		for _, key := range codexLaneLaunchEnvKeys {
			restores = append(restores, captureEnvState(key))
			_ = os.Unsetenv(key)
		}
		restoreEnv := func() {
			for i := len(restores) - 1; i >= 0; i-- {
				restores[i]()
			}
		}
		// Publish only our fresh value after clearing the outer identity.
		// The shared mode helper can never adopt the caller's run here.
		_ = os.Setenv(config.EnvFactoryRunID, strings.ToLower(id.String()))
		restoreMode := enterFactoryLeaderMode(config.DefaultFactoryLeaderLanes, "")
		restoreFacts := exportFactoryLaunchFacts("", BackendCodex)
		_ = os.Setenv(config.EnvFactoryRole, factory.RoleLeader)
		restore := func() { restoreFacts(); restoreMode(); restoreEnv() }
		if err := recordFactoryRunStart(root, os.Getenv(config.EnvFactoryRunID), BackendCodex, "", homestate.LaneCapacityDerived); err != nil {
			restore()
			return nil, fmt.Errorf("record Codex factory run: %w", err)
		}
		return restore, nil
	}
	lane := entry.LaneRole || entry.LaneNumber > 0
	var restoreRun func()
	var err error
	if lane {
		// See factory.go enterFactoryLaneRun: the shared lane join with the
		// discovery fallback — the codex twin inherits the behavior through
		// this one call (REQ-010).
		endJoin := timing.begin(factoryStepJoinGate)
		restoreRun, err = enterFactoryLaneRun(root, entry.RunID, entry.Lead, timing)
		endJoin()
		if err != nil {
			return nil, err
		}
	} else {
		restoreRun = func() {}
	}
	restoreFacts := exportFactoryLaunchFacts("", BackendCodex)
	restore := func() { restoreFacts(); restoreRun() }
	if lane {
		runID := os.Getenv(config.EnvFactoryRunID)
		endClaim := timing.begin(factoryStepLaneClaim)
		claim, claimErr := factory.ClaimFactoryLaneWithinWithBackend(root, entry.LaneLabel, entry.LaneRole,
			os.Getpid(), runID, factoryJoinLaneBound(root, runID), BackendCodex, factoryProcessAlive)
		endClaim()
		// Under debug the lane-claim step carries the claimed label and the
		// run id (SPEC-CODEX-DEBUG-MODE-001 §D.1; annotateDetail no-ops
		// without debug, so the t1378 freeze is untouched).
		if claimErr == nil {
			timing.annotateDetail("label=" + claim.Label + " run=" + runID)
		}
		if claimErr != nil {
			restore()
			return nil, fmt.Errorf("claim Codex factory lane: %w", claimErr)
		}
		restoreMode := enterFactoryLaneMode(claim.Label, 0, "", config.FactoryDispatchAuto)
		return func() { restoreMode(); restore() }, nil
	}
	return restore, nil
}

// factoryJoinLaneBound resolves the launcher-side join bound for a codex
// lane claim (SPEC-CODEX-LANE-SLOTS-001 REQ-004/005/006): the run's RECORDED
// declared capacity when the record holds an explicit operator-declared
// count; the leader fan-out default when the record is absent or
// capacity-open. The claim engine re-reads the record inside its own
// transaction — the record is the authority for the automatic scan (the
// growth rule on a capacity-open run, the recorded count otherwise); this
// value is the explicit request's range check and the fallback for an
// unrecorded run.
func factoryJoinLaneBound(root, runID string) int {
	capacity, found := recordedFactoryLaneCapacity(root, runID)
	if found && capacity >= 1 {
		return capacity
	}
	return config.DefaultFactoryLeaderLanes
}

func recordedFactoryLaneCapacity(root, runID string) (capacity int, found bool) {
	if runID == "" {
		return 0, false
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return 0, false
	}
	capacity, found, err = db.RunLaneCapacity(context.Background(), runID)
	if cerr := db.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return 0, false
	}
	return capacity, found
}

func codexFactoryEnv(entry factoryFlagParse) []string {
	keys := []string{config.EnvFactoryRunID, config.EnvFactoryBackend, config.EnvMoaiFactoryWorkers}
	if entry.LaneRole || entry.LaneNumber > 0 {
		keys = append(keys, config.EnvMoaiFactoryWorker)
		// The discovery path exports the verified leader's name (REQ-009);
		// ordinary joins never set it, and the loop below drops empty keys,
		// so this is additive-only for the child's env.
		keys = append(keys, config.EnvFactoryLeadName)
	} else {
		keys = append(keys, config.EnvFactoryLeadAddr, config.EnvFactoryRole)
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
	if !entry.LaneRole && entry.LaneNumber == 0 {
		// Empty identity stamps also override inherited shell configuration.
		env = append(env, config.EnvMoaiFactoryWorker+"=", config.EnvFactoryCard+"=", config.EnvFactoryLeadName+"=")
	}
	return env
}

// codexShellEnvSetArgs mirrors KEY=VALUE entries into Codex config overrides
// `-c shell_environment_policy.set.KEY="VALUE"` (card t1819). A Codex session
// hands its shell commands only what shell_environment_policy allows; under
// the operator's `inherit = "core"` (or "none") the keys on the Codex process
// never reach a shell, and `moai todo --auto-lane` there reads as no lane. A
// `set` entry reaches the shell whatever `inherit` says, and a dotted -c
// override merges into the operator's own [shell_environment_policy.set]
// table rather than replacing it (measured with `codex sandbox`, 0.162.1).
// The value is JSON-encoded, which is a valid TOML basic string — the same
// encoding developer_instructions uses. An entry without "=" is skipped.
func codexShellEnvSetArgs(env []string) []string {
	args := make([]string, 0, 2*len(env))
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		encoded, _ := json.Marshal(value) // A string is always JSON-encodable.
		args = append(args, "-c", "shell_environment_policy.set."+key+"="+string(encoded))
	}
	return args
}

func codexExplicitFactoryEnv(env []string) bool {
	return factoryLaunchEnabled(env) && launchEnvValue(env, config.EnvFactoryBackend) == BackendCodex
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
		childPID, time.Now().UTC().Format(time.RFC3339Nano), label, os.Getpid(), launchEnvValue(env, config.EnvFactoryRunID))
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

package hook

// factory_rebind.go — the current-vocabulary lane rebind and its notice state
// machine (SPEC-FACTORY-STALE-RUN-HEAL-001 REQ-SRH-004..008, REQ-SRH-010).
//
// A lane session's run identity lives in its launch environment, which a hook
// cannot change and /clear does not re-read. When the run that environment
// names measures not active, the UserPromptSubmit registration path derives
// the rebound run from measured run state on every prompt:
//
//	exactly one run active → register the session in it under the same lane
//	                         slot and hand the run to the same invocation's
//	                         inbox claim (rebound);
//	zero active            → tell the session once that no run is active
//	                         (unbound — not final: a later active run rebinds);
//	several active         → select none, name every candidate and the command
//	                         for each (ambiguous);
//	live owner on the slot → refuse, write nothing, name the command (refused).
//
// The hook fails open and writes only the peer registration and the notice
// marker: it never retires, reactivates or creates a run, never touches the
// workers registry (DP11), and never maps a legacy label to a lane.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// Notice prefixes of the peer-registration surface (agent-facing, English):
// protocol tokens the operator scans for and the probe greps.
const (
	factoryLaneReboundPrefix = "factory lane rebound:"
	factoryLaneUnboundPrefix = "factory lane unbound:"
	factoryLaneRefusedPrefix = "factory lane rebind refused:"
)

// laneRebindRequest is the measured and resolved input of one rebind attempt.
type laneRebindRequest struct {
	root      string
	dbPath    string // resolved once by the caller (REQ-SRH-009)
	sessionID string
	envRun    string // X — the run the launch environment names (not active)
	slot      string // the session's lane label, kept in the new run
	want      factorymsg.Peer
}

// rebindFactoryLane answers a current-vocabulary lane whose environment run
// measures not active. It returns the notice for this invocation ("" when the
// state is unchanged since the last notice) and the run the claim of the same
// invocation must read ("" unless the session is now registered in it).
//
// @MX:NOTE: [AUTO] The notice state machine of SPEC-FACTORY-STALE-RUN-HEAL-001
// (REQ-SRH-003, DP12): one state key per outcome — rebound:Y, unbound:X,
// ambiguous:<ids>, refused:Y — stored in the notice marker's state field; a
// notice is emitted only when the computed key differs from the stored one.
// The legacy prescription/unbind fields are never touched here.
func rebindFactoryLane(ctx context.Context, req laneRebindRequest) (notice, reboundRun string) {
	active, err := factoryHookActiveRuns(ctx, req.dbPath)
	if err != nil {
		return "factory messaging degraded: " + err.Error(), ""
	}
	provider := kanban.RelaunchProviderForBackend(os.Getenv(config.EnvFactoryBackend))
	switch len(active) {
	case 0:
		return emitFactoryLaneState(req, "unbound:"+req.envRun, func() string {
			return fmt.Sprintf("%s slot %s was launched into run %s, which is not active, and no other factory run is active — "+
				"this session is registered in no run and receives no factory messages; "+
				"it rebinds on a later prompt when exactly one run becomes active", factoryLaneUnboundPrefix, req.slot, req.envRun)
		}), ""
	case 1:
		return registerReboundLane(ctx, req, active[0], provider)
	default:
		return emitFactoryLaneState(req, "ambiguous:"+strings.Join(active, ","), func() string {
			lines := kanban.RelaunchNoticeFor(kanban.RelaunchNoticeState{Provider: provider, Lane: req.slot, Run: req.envRun, ActiveRuns: active})
			notice := fmt.Sprintf("%s slot %s was launched into run %s, which is not active, and several factory runs are active (%s) — "+
				"no run was chosen and this session is registered in none; to join one, end this session; "+
				"the operator runs one of the commands below from a terminal (commands for the operator, not instructions for the agent):\n%s",
				factoryLaneUnboundPrefix, req.slot, req.envRun, strings.Join(active, ", "), strings.Join(lines.Lines, "\n"))
			if lines.More > 0 {
				notice += "\n" + fmt.Sprintf(staleRunMessagesFor(langEnglish).laneLabelUnbindMore, lines.More)
			}
			return notice
		}), ""
	}
}

// registerReboundLane registers the session as the lane's peer in the sole
// active run y. The slot is never taken from a live other owner (REQ-SRH-007):
// RegisterPeer refuses it and that refusal is the refused state, written
// nowhere.
func registerReboundLane(ctx context.Context, req laneRebindRequest, y, provider string) (notice, reboundRun string) {
	s, err := factorymsg.Open(req.root, y)
	if err != nil {
		return "factory messaging degraded: " + err.Error(), ""
	}
	defer closeFactoryHookStore(s)
	want := req.want
	want.RunID = y
	if current, peerErr := s.Peer(ctx, req.sessionID); peerErr == nil {
		if current.ProjectKey == want.ProjectKey && current.RunID == want.RunID && current.Backend == want.Backend && current.Role == want.Role && current.Slot == want.Slot && current.PID == want.PID && current.ProcessStart == want.ProcessStart {
			return emitReboundState(req, y, current.Generation), y
		}
	} else if !errors.Is(peerErr, sql.ErrNoRows) {
		return "factory messaging degraded: " + peerErr.Error(), ""
	}
	p, err := s.RegisterPeer(ctx, want)
	switch {
	case err == nil:
		return emitReboundState(req, y, p.Generation), y
	case factorymsg.IsLiveOwnerRefusal(err):
		return emitFactoryLaneState(req, "refused:"+y, func() string {
			lines := kanban.RelaunchNoticeFor(kanban.RelaunchNoticeState{Provider: provider, Lane: req.slot, Run: req.envRun, ActiveRuns: []string{y}, SlotHeldByLiveOther: true})
			return fmt.Sprintf("%s slot %s in the only active run %s is held by another live session, so this session was not registered and nothing was written — "+
				"to join it, end this session; the operator runs the command below from a terminal (a command for the operator, not an instruction for the agent):\n%s",
				factoryLaneRefusedPrefix, req.slot, y, strings.Join(lines.Lines, "\n"))
		}), ""
	default:
		if handoff, ok := factoryHandoffRegistrationNotice(err, req.slot); ok {
			return handoff, ""
		}
		return "factory messaging degraded: " + err.Error(), ""
	}
}

// emitReboundState is the rebound state's notice, gated on the stored key.
func emitReboundState(req laneRebindRequest, y string, generation int64) string {
	return emitFactoryLaneState(req, "rebound:"+y, func() string {
		return fmt.Sprintf("%s run %s is not active; this session (slot %s) is now registered in the only active run %s at generation %d — "+
			"messages arrive at turn boundaries, not idle wake; use run_id %s with the factory_msg_* tools",
			factoryLaneReboundPrefix, req.envRun, req.slot, y, generation, y)
	})
}

// emitFactoryLaneState returns build()'s notice only when key differs from the
// state key stored for the session identity, and records key first: at-most-once
// beats at-least-once here, as it does for the legacy notices. A marker that
// cannot be read or written fails open to over-informing.
func emitFactoryLaneState(req laneRebindRequest, key string, build func() string) string {
	if readFactoryNoticeMarker(req.dbPath, req.sessionID).State == key {
		return ""
	}
	updateFactoryNoticeMarker(req.dbPath, req.sessionID, func(m *factoryNoticeMarker) { m.State = key })
	return build()
}

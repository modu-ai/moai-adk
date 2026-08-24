// factory_lanes.go — per-lane factory progress (SPEC-WEB-CONSOLE-015 M1).
//
// Lanes are NOT chain roles, so they live beside ChainRoles rather than
// widening it: the four-role chain is a fixed dispatch vocabulary, and a
// variable-length role list would make every chain consumer defend against it.
//
// The join is `workers.json[lane-N].PID → active-sessions entry → kanban
// record`, and it is non-unique on BOTH sides. The factory registry's loader is
// fail-open and pruning dead claims is a separate call this console does not
// make, so one pid can sit on two lanes; and session.Registry.Register
// deduplicates by session id alone, so two entries can carry one pid. Either
// case is resolved to "unresolved" rather than to a winner — a completed but
// wrong lookup renders a confident wrong row, which is worse than an empty one.
//
// Read-only, like the rest of the console: nothing here writes, and the
// registry is consumed through kanban.LoadFactoryRegistry, whose failure mode
// is an empty map rather than a mutation.
package web

import (
	"sort"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// Unresolved-lane reasons. They are why a row carries no join values, and the
// view renders the marker rather than a plausible substitute.
const (
	LaneUnresolvedNoSession = "no-session" // the pid is in no session entry
	LaneUnresolvedAmbiguous = "ambiguous"  // the pid does not identify one session
	LaneUnresolvedNoRecord  = "no-record"  // the session resolved but wrote no record
	LaneUnresolvedLaneClash = "lane-clash" // the record names a different lane than the registry
)

// LaneVM is one registered factory lane. A lane is always PRESENT once
// registered — an unresolvable join marks the row rather than dropping it,
// because a missing row is indistinguishable from a lane that never ran.
type LaneVM struct {
	Lane    int    // the lane number; always set, and the row's identity
	Session string // short session id, empty while unresolved
	CardID  string
	SpecID  string
	Backend string
	State   string
	Stage   string

	// StageEstimated mirrors RoleVM's flag: a stage derived from the heartbeat
	// is an estimate, never a recorded transition.
	StageEstimated bool

	Unresolved bool
	// UnresolvedReason is one of the Lane* constants above, empty when resolved.
	UnresolvedReason string
}

// loadFactoryLanes builds the lane rows from the factory registry, the session
// registry index (keyed by full session id) and the kanban records already read
// for this render — no additional read of active-sessions.json, and no write.
//
// An absent or malformed registry yields zero lanes and no error: the section
// then renders as carrying no registered lanes (REQ-WC15-046).
func loadFactoryLanes(root string, sessions map[string]SessionVM, records []KanbanRecord) []LaneVM {
	reg := kanban.LoadFactoryRegistry(kanban.FactoryRegistryPath(root))
	if len(reg) == 0 {
		return nil
	}

	// Count-then-resolve on the factory side: a pid claimed by two lanes
	// identifies neither, so it must not collapse to a last-write-wins winner.
	laneClaims := map[int]int{}
	for label, entry := range reg {
		if _, ok := kanban.SplitFactoryLaneLabel(label); !ok {
			continue // not a lane label — not this section's row
		}
		laneClaims[entry.PID]++
	}

	// Count-then-resolve on the session side, for the same reason: the session
	// registry does not enforce pid uniqueness.
	sessionsByPID := map[int][]string{}
	for id, vm := range sessions {
		if vm.PID > 0 {
			sessionsByPID[vm.PID] = append(sessionsByPID[vm.PID], id)
		}
	}

	recordBySession := make(map[string]KanbanRecord, len(records))
	for _, rec := range records {
		if rec.SessionID != "" {
			recordBySession[rec.SessionID] = rec
		}
	}

	out := make([]LaneVM, 0, len(reg))
	for label, entry := range reg {
		n, ok := kanban.SplitFactoryLaneLabel(label)
		if !ok {
			continue
		}
		out = append(out, resolveLane(n, entry.PID, laneClaims, sessionsByPID, sessions, recordBySession))
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Lane < out[b].Lane })
	return out
}

// resolveLane resolves one lane's row. Every failure direction returns a
// present row carrying its lane number and an unresolved marker.
func resolveLane(
	lane, pid int,
	laneClaims map[int]int,
	sessionsByPID map[int][]string,
	sessions map[string]SessionVM,
	recordBySession map[string]KanbanRecord,
) LaneVM {
	row := LaneVM{Lane: lane}

	ids := sessionsByPID[pid]
	if pid <= 0 || len(ids) == 0 {
		return unresolvedLane(row, LaneUnresolvedNoSession)
	}
	if laneClaims[pid] > 1 || len(ids) > 1 {
		return unresolvedLane(row, LaneUnresolvedAmbiguous)
	}

	sessionID := ids[0]
	rec, ok := recordBySession[sessionID]
	if !ok {
		return unresolvedLane(row, LaneUnresolvedNoRecord)
	}

	vm := sessions[sessionID]
	if vm.State == "" {
		vm.State = StateStale
	}
	stage, estimated := estimateStage(vm)

	// The registry label's number IS the row identity and is never overwritten:
	// a record whose own lane number disagrees with it is evidence the join
	// landed on the wrong session, which REQ-WC15-047 renders unresolved rather
	// than presenting. Overwriting instead would let one bad join relabel a row
	// — two lanes could then render the same number while the registered lane
	// vanished from the page, which is the misattribution this section exists to
	// keep off the screen.
	if rec.Lane > 0 && rec.Lane != lane {
		return unresolvedLane(row, LaneUnresolvedLaneClash)
	}
	row.Session = vm.ID
	row.CardID = rec.CardID
	row.SpecID = rec.SpecID
	row.Backend = rec.Backend
	row.State = vm.State
	row.Stage = stage
	row.StageEstimated = estimated
	return row
}

// unresolvedLane marks a row unresolved. It carries no state and no stage: an
// unresolved lane is one whose session is not known, so asserting a stage for
// it would be exactly the plausible-substitute this console refuses to render.
func unresolvedLane(row LaneVM, reason string) LaneVM {
	row.Unresolved = true
	row.UnresolvedReason = reason
	return row
}

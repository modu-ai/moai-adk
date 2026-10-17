package homestate

// card_age.go — SPEC-CARD-AGE-ALARM-001 (card t1838) M2: the read-only age
// measurement behind the display-only work-item-age alarm. Every function
// here reads the store (or an injected source) and writes nothing
// (REQ-CAA-007): no card-row write, no event append, no lease change.

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The alarm class names of REQ-CAA-008.
const (
	AgeClassGateWait   = "gate-wait"
	AgeClassNoProgress = "no-progress"
)

// gateWaitingStates is class (a) of REQ-CAA-008: the cards waiting on a human
// gate. Renewal and expiry touch lease-holding states only, so a card in one
// of these states holds no lease and its wait is the leader's to see.
var gateWaitingStates = []string{CardKickoff, CardNeedsDecision, CardBlocked}

// GateWaitingCardState reports whether state is in class (a).
func GateWaitingCardState(state string) bool { return slices.Contains(gateWaitingStates, state) }

// pipelineStageStates is the stage set of the REQ-CAA-004 anchor: the working
// states plus the delivery states, minus the pre-pipeline pool states
// (picked, assigned, leased) a card passes through but never works in.
var pipelineStageStates = []string{
	CardPlan, CardPlanAudit, CardRun, CardSync, CardSyncAudit, CardMergeReady, CardMerging,
	CardMergedLocal, CardPushed, CardCIGreen, CardPROpen, CardMergedPR,
}

// PipelineCheckpoint reports whether state is a pipeline stage of the
// REQ-CAA-004 stage-change anchor.
func PipelineCheckpoint(state string) bool { return slices.Contains(pipelineStageStates, state) }

// AgeClassOf returns the alarm class of a card row (REQ-CAA-008): gate-wait
// for kickoff, needs-decision, and blocked; no-progress for every other
// non-terminal F1 state (the delivery states included — a landing that waits
// past its batch is stuck work); "" for terminal states and legacy rows,
// which no class covers.
func AgeClassOf(c Card) string {
	if c.Legacy() || IsTerminalCardState(c.State) {
		return ""
	}
	if GateWaitingCardState(c.State) {
		return AgeClassGateWait
	}
	return AgeClassNoProgress
}

// CardEvent is one parsed events row the age anchors derive from: a
// card.transition or lease.expired row, with the payload's card identity and
// from/to pair. CreatedAt is carried raw; the timestamp-validity rule of
// REQ-CAA-004 is applied where it is read.
type CardEvent struct {
	RunID     string
	CardID    string
	Kind      string
	From      string
	To        string
	CreatedAt string
	Seq       int64
}

// LoadCardEvents reads every card.transition and lease.expired event of runID
// (every run when runID is empty), oldest first — one bounded query per
// status render (REQ-CAA-003/004/005). A row whose payload does not parse is
// skipped: the anchors fall back rather than fail the read. It never writes.
func (f *FactoryDB) LoadCardEvents(ctx context.Context, runID string) ([]CardEvent, error) {
	// SQL: the fragment is a compile-time constant; runID goes through a ? placeholder.
	query := `SELECT run_id, kind, payload_json, created_at, seq FROM events WHERE kind IN ('card.transition','lease.expired')`
	var args []any
	if strings.TrimSpace(runID) != "" {
		query += ` AND run_id=?`
		args = append(args, runID)
	}
	rows, err := f.DB.QueryContext(ctx, query+` ORDER BY seq`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []CardEvent
	for rows.Next() {
		var runID, kind, payload, createdAt string
		var seq int64
		if err := rows.Scan(&runID, &kind, &payload, &createdAt, &seq); err != nil {
			return nil, err
		}
		var p struct {
			CardID string `json:"card_id"`
			From   string `json:"from"`
			To     string `json:"to"`
		}
		if json.Unmarshal([]byte(payload), &p) != nil {
			continue
		}
		out = append(out, CardEvent{
			RunID: runID, CardID: p.CardID, Kind: kind,
			From: p.From, To: p.To, CreatedAt: createdAt, Seq: seq,
		})
	}
	return out, rows.Err()
}

// AgeSignal is one displayed age. Known=false renders "unknown" — an absent,
// unreadable, or invalid source, never a fabricated value (REQ-CAA-004).
type AgeSignal struct {
	Known bool
	At    time.Time
	Age   time.Duration
}

// CardAges carries every derived age of one card: the class age the alarm
// bands (GateWait for class (a), NoProgress for class (b)) and the three
// signals of REQ-CAA-004 displayed beside it.
type CardAges struct {
	GateWait   AgeSignal
	Stage      AgeSignal
	Commit     AgeSignal
	Activity   AgeSignal
	NoProgress AgeSignal
}

// CommitSource supplies the head-commit anchor of a card's worktree
// (REQ-CAA-004): the CLI's bounded git reader, injected so this package reads
// only the store. (zero, false) is unknown — a missing or unreadable
// worktree, or a failed git invocation, never a fabricated timestamp.
type CommitSource func(worktreePath string) (time.Time, bool)

// validAnchorAt applies the timestamp-validity rule of REQ-CAA-004 to a
// parsed timestamp: a source dated in the future beyond
// config.CardAgeAlarmClockSkewTolerance (a rolled-back reading clock, or a DB
// read from another machine) is unknown — never an age of zero, never a
// negative age. A skew within the tolerance stays valid and reads age zero.
func validAnchorAt(at, now time.Time) bool {
	return !at.After(now.Add(config.CardAgeAlarmClockSkewTolerance))
}

// validAgeAnchor parses raw as an RFC3339 timestamp and applies the validity
// rule. Empty and unparseable values are unknown.
func validAgeAnchor(raw string, now time.Time) (time.Time, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, false
	}
	at, err := time.Parse(time.RFC3339Nano, trimmed)
	if err != nil || !validAnchorAt(at, now) {
		return time.Time{}, false
	}
	return at, true
}

// knownAge builds the signal of a valid anchor; no derived age is negative.
func knownAge(at, now time.Time) AgeSignal {
	age := now.Sub(at)
	if age < 0 {
		age = 0
	}
	return AgeSignal{Known: true, At: at, Age: age}
}

// ComputeCardAges derives every age this SPEC displays for one card from its
// row, its card.transition/lease.expired events, and the injected commit
// source. A terminal or legacy row yields all-unknown signals (no class
// covers it); the gate-wait age is derived only for class (a) and the
// no-progress age only for class (b). It reads its inputs and writes nothing.
func ComputeCardAges(c Card, events []CardEvent, commitAt CommitSource, now time.Time) CardAges {
	var out CardAges
	class := AgeClassOf(c)
	if class == "" {
		return out
	}

	// Stage checkpoint (REQ-CAA-004): the most recent stage CHANGE — an
	// event whose `to` is a pipeline stage and differs from the prior such
	// event's `to`. A same-stage re-entry (the T4 resume shape) never
	// refreshes it.
	if ev, ok := lastStageChange(events); ok {
		if at, valid := validAgeAnchor(ev.CreatedAt, now); valid {
			out.Stage = knownAge(at, now)
		}
	}

	// Commit (REQ-CAA-004): the injected head-commit anchor of the recorded
	// worktree — the one signal that crosses machine boundaries, so the
	// validity rule applies here too. An empty worktree path never invokes
	// the source.
	if commitAt != nil && strings.TrimSpace(c.WorktreePath) != "" {
		if at, ok := commitAt(c.WorktreePath); ok && validAnchorAt(at, now) {
			out.Commit = knownAge(at, now)
		}
	}

	// Activity (REQ-CAA-004): the lease heartbeat, displayed but never
	// admitted to the no-progress basis (activity is not progress).
	if at, ok := validAgeAnchor(c.HeartbeatAt, now); ok {
		out.Activity = knownAge(at, now)
	}

	// Gate-wait age (REQ-CAA-003): the most recent event whose payload `to`
	// equals the current state, falling back to updated_at when no such
	// event exists or its timestamp is invalid — and rendering unknown when
	// no valid basis remains.
	if class == AgeClassGateWait {
		raw := ""
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			if e.To == c.State {
				raw = e.CreatedAt
				break
			}
		}
		if at, ok := validAgeAnchor(raw, now); !ok {
			if fb, ok2 := validAgeAnchor(c.UpdatedAt, now); ok2 {
				out.GateWait = knownAge(fb, now)
			}
		} else {
			out.GateWait = knownAge(at, now)
		}
	}

	// No-progress age (REQ-CAA-005): the newest of the stage checkpoint, the
	// commit, and the card's creation event — over the known members only, so
	// a future-dated member can neither yield a negative age nor suppress a
	// band. The creation member bounds the age in both directions (iter2
	// findings B1/B2): pool-state cycling never resets it, and a fresh card
	// never inherits its base commit's age. The updated_at fallback binds
	// only rows with no events at all.
	if class == AgeClassNoProgress {
		var basis []time.Time
		if out.Stage.Known {
			basis = append(basis, out.Stage.At)
		}
		if out.Commit.Known {
			basis = append(basis, out.Commit.At)
		}
		if ev, ok := creationEvent(events); ok {
			if at, valid := validAgeAnchor(ev.CreatedAt, now); valid {
				basis = append(basis, at)
			}
		}
		if len(basis) > 0 {
			newest := basis[0]
			for _, at := range basis[1:] {
				if at.After(newest) {
					newest = at
				}
			}
			out.NoProgress = knownAge(newest, now)
		} else if len(events) == 0 {
			if at, ok := validAgeAnchor(c.UpdatedAt, now); ok {
				out.NoProgress = knownAge(at, now)
			}
		}
	}
	return out
}

// lastStageChange returns the most recent card.transition event that records
// a stage change (REQ-CAA-004, plan-audit iter1 D1): its `to` is a pipeline
// stage and differs from the `to` of the card's most recent prior
// pipeline-stage transition — a first entry, with no such prior event, is a
// stage change. Stage changes in either direction qualify; only a re-entry
// into the stage already held is excluded, so the expiry → re-acquire → T4
// crash-loop cycle never resets the checkpoint.
func lastStageChange(events []CardEvent) (CardEvent, bool) {
	var out CardEvent
	sawStage, sawChange := false, false
	var prevTo string
	for _, e := range events {
		if e.Kind != "card.transition" || !PipelineCheckpoint(e.To) {
			continue
		}
		if !sawStage || e.To != prevTo {
			out, sawChange = e, true
		}
		prevTo, sawStage = e.To, true
	}
	return out, sawChange
}

// creationEvent returns the card's creation event — the card.transition whose
// payload `from` is empty (card_picked.go:173), exactly one per card that has
// events and by construction that card's earliest event.
func creationEvent(events []CardEvent) (CardEvent, bool) {
	for _, e := range events {
		if e.Kind == "card.transition" && e.From == "" {
			return e, true
		}
	}
	return CardEvent{}, false
}

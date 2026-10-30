package core

import (
	"fmt"
	"strings"
)

// Fact keys a record can show about itself. A write carries them (derived from the
// record it writes) as the evidence of the edges it absorbs; they are facts, so the
// writer cannot assert one the record does not hold.
const (
	FactReceipt         = "receipt"          // a PR receipt is on the record
	FactReceiptAccepted = "receipt-accepted" // the leader accepted that receipt
	FactMerged          = "merged"           // the remote merge was seen
	FactClosed          = "closed"           // the Issue was closed
	FactSettled         = "settled"          // no obligation (local sync, deploy, close) is pending
)

// The evidence a write carries for each of the named in-place parks.
const (
	GuardOperatorHold = "operator-hold"
	GuardFixLimit     = "fix-limit"
	GuardReclaimCap   = "reclaim-cap"
)

// inPlaceParks are the parks S46 has no edge for: the operator's hold (OD-12,
// unconfirmed) and the two the local ledger makes. Each is taken in place, from a
// position that is neither finished nor parked, with its own guard.
var inPlaceParks = map[BlockReason]string{
	ReasonOperatorHold: GuardOperatorHold,
	ReasonFixLimit:     GuardFixLimit,
	ReasonReclaimCap:   GuardReclaimCap,
}

// edgeFacts are the facts a single edge needs on the record on top of its guard keys,
// for the edges whose keys the record can show.
var edgeFacts = map[string][]string{
	"E03": {FactReceipt, FactReceiptAccepted},
	"E12": {FactMerged},
	"E17": {FactMerged}, // the same position as E12, so the same fact
	"E13": {FactMerged, FactClosed, FactSettled},
}

// composition is a move the old State makes in one write that stands for several
// edges. The list below is closed: a composition not in it is refused. Each one
// needs the facts of the edges it absorbed on the record it writes.
type composition struct {
	name     string
	from, to func(Position) bool
	needs    []string
}

func atStage(p Position, ss ...Stage) bool {
	for _, s := range ss {
		if p.Stage == s {
			return true
		}
	}
	return false
}

// noHandoff is BUILDING before any receipt reads as FIX: a FIX position holds a receipt
// the leader has not accepted, and finishing from it would skip that acceptance.
func noHandoff(p Position) bool { return atStage(p, StageBuilding) && !isFix(p) }

func isFix(p Position) bool     { return p.Stage == StageBuilding && p.Sub == SubFix }
func isDone(p Position) bool    { return p.Stage == StageDone }
func isMergedP(p Position) bool { return p.Stage == StageMerging && p.RemoteIntegrated }

// compositions are the legacy compositions, named:
//   - an unaccepted receipt reads as BUILDING/FIX (E03, E06);
//   - an accepted receipt finished in one write (E08, E12, E13);
//   - an accepted receipt with the merge seen (E08, E12);
//   - a record finished or merged with no handoff on it (E03, E08, E12, E13 / E03, E08,
//     E12): the runner's local path, which skips the leader's acceptance. These two are
//     a named gap, kept so the runner can finish an Issue without a PR until its steps
//     carry that evidence.
var compositions = []composition{
	{"published-receipt", func(p Position) bool { return atStage(p, StageBuilding) && !isFix(p) }, isFix, []string{FactReceipt}},
	{"accepted-receipt-finished", func(p Position) bool { return atStage(p, StageReviewing) }, isDone,
		[]string{FactReceipt, FactReceiptAccepted, FactMerged, FactClosed, FactSettled}},
	{"accepted-receipt-merge-seen", func(p Position) bool { return atStage(p, StageReviewing) }, isMergedP,
		[]string{FactReceipt, FactReceiptAccepted, FactMerged}},
	{"no-handoff-finished", noHandoff, isDone,
		[]string{FactMerged, FactClosed, FactSettled}},
	{"no-handoff-merge-seen", noHandoff, isMergedP,
		[]string{FactMerged}},
}

// CheckMove says whether a record may go from one S46 position to another in a
// single write. It may stay where it is; or one table edge leads there (E01-E27); or
// the move is one of the named legacy compositions. Evidence:
//
//   - E20 (resume) is the one edge whose guard the writer must carry: have holds
//     "resume-guard:<reason>" for the reason being resolved, and none is invented here,
//     the unknown family included.
//   - E03, E12 and E13 and every composition need facts the record shows (the Fact
//     constants); the other edges' guard keys are supplied here, because the record
//     holds nothing to test them against and the table only checks they are present.
//
// A parked position leaves by a single edge only: resumed in place, cancelled, or an
// E18 adopt to BUILDING/SPEC. A write that moves the stage and parks in the same
// write has no edge. What this does not say is which of the absorbed edges' other
// guards held; that stays with the runner step that makes the write.
//
// The runner keeps the old State as its progress record (step 2 of the model
// migration) and derives the position from it. This is the point where a derived
// position that the table cannot reach is refused, instead of being stored.
func CheckMove(from, to Position, have []string) error {
	if sameSpot(from, to) {
		return nil
	}
	has := map[string]bool{}
	for _, h := range have {
		has[h] = true
	}
	var missingErr error
	lacks := func(what string, needs []string) bool {
		var m []string
		for _, n := range needs {
			if !has[n] {
				m = append(m, n)
			}
		}
		if len(m) == 0 {
			return false
		}
		if missingErr == nil {
			missingErr = fmt.Errorf("%w: %s needs %s", ErrGuardMissing, what, strings.Join(m, ", "))
		}
		return true
	}
	// The named in-place parks have no S46 edge. sameSpot with the park stripped from to
	// also requires that from carries none.
	if to.Park != nil && !from.Stage.terminal() && sameSpot(from, withoutPark(to)) {
		if guard, ok := inPlaceParks[to.Park.Reason]; ok {
			if lacks(string(to.Park.Reason), []string{guard}) {
				return missingErr
			}
			return nil
		}
	}
	// E20 resolves a park in place and nothing else.
	if from.Park != nil && sameSpot(withoutPark(from), to) {
		_, err := Transition(from, Event{ID: "E20", Evidence: have})
		return err
	}
	for _, e := range Edges() {
		if e.ID == "E20" {
			continue
		}
		for _, o := range moveOutcomes(e) {
			ev := Event{ID: e.ID, Outcome: o, Evidence: e.needsFor(o)}
			if to.Park != nil {
				ev.Park = *to.Park
			}
			next, err := Transition(from, ev)
			if err != nil || !sameSpot(next, to) {
				continue
			}
			if lacks(e.ID, edgeFacts[e.ID]) {
				continue
			}
			return nil
		}
	}
	if from.Park == nil && to.Park == nil {
		for _, c := range compositions {
			if c.from(from) && c.to(to) && !lacks(c.name, c.needs) {
				return nil
			}
		}
	}
	if missingErr != nil {
		return missingErr
	}
	return fmt.Errorf("%w: no S46 edge or named composition moves %s to %s", ErrTransitionInvalid, spot(from), spot(to))
}

func withoutPark(p Position) Position { p.Park = nil; return p }

func moveOutcomes(e Edge) []string {
	if len(e.Outcomes) == 0 {
		return []string{""}
	}
	out := make([]string, 0, len(e.Outcomes))
	for _, o := range e.Outcomes {
		out = append(out, o.Name)
	}
	return out
}

func sameSpot(a, b Position) bool {
	if a.Stage != b.Stage || a.Sub != b.Sub || a.RemoteIntegrated != b.RemoteIntegrated {
		return false
	}
	if (a.Park == nil) != (b.Park == nil) {
		return false
	}
	return a.Park == nil || a.Park.Reason == b.Park.Reason
}

func spot(p Position) string {
	s := string(p.Stage)
	if p.Sub != "" {
		s += "/" + string(p.Sub)
	}
	if p.RemoteIntegrated {
		s += " (integrated)"
	}
	if p.Park != nil {
		s += " parked(" + string(p.Park.Reason) + ")"
	}
	return s
}

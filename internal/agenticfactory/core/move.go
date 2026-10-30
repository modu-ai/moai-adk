package core

import "fmt"

// maxChain is the most table edges one write may stand for. The old State collapses
// whole stretches of the lifecycle into one write (landing to done is E08, E12 and
// E13; a claimed record that already holds a receipt is E03 and E06), and the
// longest stretch the runner writes at once is claimed to done: E03, E08, E12, E13
// (E03 leaves BUILDING from any sub stage). A queued record never reaches done in
// one write (E01 comes first), which five edges would allow.
const maxChain = 4

// CheckMove says whether a record may go from one S46 position to another in a
// single write: it may stay where it is, or a chain of at most maxChain table edges
// (E01-E27, E20 resume included) must lead from the first to the second. It is a
// table-membership test and nothing more: it supplies every guard key itself, so
// it never judges the evidence an edge needs, only that the way exists. The park
// text may change while the reason stays (a refreshed evidence or resume line).
//
// What it refuses is a write the table has no way to make: a finished position
// opening again, a sub stage going backwards, a parked position moving on, a queued
// record finishing in one write. It does not say WHICH edges a write stands for;
// a runner step that must be exact about that checks its own edge.
//
// The runner keeps the old State as its progress record (step 2 of the model
// migration) and derives the position from it. This is the point where a derived
// position that the table cannot reach is refused, instead of being stored.
func CheckMove(from, to Position) error {
	if sameSpot(from, to) {
		return nil
	}
	seen := map[string]bool{spot(from): true}
	frontier := []Position{from}
	for depth := 0; depth < maxChain && len(frontier) > 0; depth++ {
		var next []Position
		for _, p := range frontier {
			for _, m := range successors(p, to) {
				n := m.pos
				if sameSpot(n, to) {
					return nil
				}
				// Resuming a parked position is a write of its own: E20 as a step
				// on the way to another place would be a way round the rule that a
				// parked position does not move on (Transition's fail-closed rule).
				if m.edge == "E20" {
					continue
				}
				if k := spot(n); !seen[k] {
					seen[k] = true
					next = append(next, n)
				}
			}
		}
		frontier = next
	}
	return fmt.Errorf("%w: no chain of S46 edges (at most %d) moves %s to %s", ErrTransitionInvalid, maxChain, spot(from), spot(to))
}

type step struct {
	pos  Position
	edge string
}

// successors are the positions one table edge reaches from p. A park outcome parks
// with the target's park, since that is the only one a chain could end on.
func successors(p, to Position) []step {
	var out []step
	for _, e := range Edges() {
		for _, o := range moveOutcomes(e) {
			ev := Event{ID: e.ID, Outcome: o, Evidence: moveEvidence(e, o, p)}
			if to.Park != nil {
				ev.Park = *to.Park
			}
			if n, err := Transition(p, ev); err == nil {
				out = append(out, step{n, e.ID})
			}
		}
	}
	return out
}

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

// moveEvidence names every guard key of the edge and outcome, so the table's own
// existence and position rules are the only thing left to fail.
func moveEvidence(e Edge, outcome string, from Position) []string {
	if e.ID == "E20" {
		if from.Park == nil {
			return nil
		}
		return []string{"resume-guard:" + string(from.Park.Reason)}
	}
	return e.needsFor(outcome)
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

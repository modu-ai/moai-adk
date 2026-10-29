package core

import (
	"fmt"
	"sort"
	"strings"
)

// Outcome is one result an edge can have. ReasonFromCaller outcomes leave the
// overlay reason to the caller (canon-unspecified: S46 does not name it for E15
// cleanup-unknown and E23 inconclusive). ReasonClassified outcomes (E19) accept
// only a reason from the classification set (canon line 65, guard "실제오류분류"). S46 gives some edges two results
// (E10, E15, E16, E23); the one marked Default is taken when the caller names none.
type Outcome struct {
	Name    string
	Default bool
	// To is the new stage; "" keeps the current one. Sub applies to BUILDING.
	To  Stage
	Sub SubStage
	// Park, when set, parks the position with that reason (ReasonFromCaller
	// leaves the reason to the caller's classification).
	Park BlockReason
	// Adopt marks the E18 outcome that applies an approved revision: it is taken
	// only from a contract-change parked BUILDING/FIX, moves to To/Sub and
	// releases the park, the one stage change a parked position may make.
	Adopt bool
	// Integrated sets remote_integrated on a MERGING position (E12, E17).
	Integrated bool
	// Needs are guard keys this outcome needs on top of the edge's own. A parking
	// outcome that records an unclear result needs less than the one that claims
	// the result is known (S46 E15: an unclear cleanup stays parked).
	Needs []string
}

// Edge is one S46 transition. Needs are guard evidence keys the caller must
// present; the table checks that they are present, never what they mean.
type Edge struct {
	ID       string
	From     []Stage
	FromSub  []SubStage // when set, the BUILDING position must be in one of these
	Needs    []string
	Outcomes []Outcome
	// RequiresIntegrated demands a MERGING position whose remote merge was seen.
	RequiresIntegrated bool
}

func (e Edge) allows(p Position) bool {
	ok := false
	for _, s := range e.From {
		if s == p.Stage {
			ok = true
		}
	}
	if !ok {
		return false
	}
	if e.FromSub != nil {
		sub := false
		for _, s := range e.FromSub {
			if s == p.Sub {
				sub = true
			}
		}
		if !sub {
			return false
		}
	}
	return !e.RequiresIntegrated || p.RemoteIntegrated
}

// needsFor is the edge's guard plus the needs of the named outcome ("" is the
// default one).
func (e Edge) needsFor(outcome string) []string {
	out := append([]string(nil), e.Needs...)
	for _, o := range e.Outcomes {
		if (outcome == "" && o.Default) || (outcome != "" && o.Name == outcome) {
			out = append(out, o.Needs...)
		}
	}
	return out
}

func (e Edge) defaultOutcome() *Outcome {
	for i := range e.Outcomes {
		if e.Outcomes[i].Default {
			return &e.Outcomes[i]
		}
	}
	return nil
}

var nonterminal = []Stage{StageReady, StageBuilding, StageReviewing, StageMerging}
var anyStage = []Stage{StageReady, StageBuilding, StageReviewing, StageMerging, StageDone, StageCancelled}

func stay(name string) Outcome       { return Outcome{Name: name, Default: true} }
func parkWith(r BlockReason) Outcome { return Outcome{Name: "park", Default: true, Park: r} }

// edges is S46's transition table (s46.md E01-E27). E20 (resume) is handled in
// Transition because its guard depends on the parked reason.
var edges = []Edge{
	{ID: "E01", From: []Stage{StageReady}, Needs: []string{"contract", "dependencies", "scope", "budget", "no-active-owner", "ready-ack"},
		Outcomes: []Outcome{{Name: "claim", Default: true, To: StageBuilding, Sub: SubSpec}}},
	{ID: "E02", From: []Stage{StageReady}, Needs: []string{"dependency-evidence"}, Outcomes: []Outcome{parkWith(ReasonDependency)}},
	{ID: "E03", From: []Stage{StageBuilding}, Needs: []string{"pr-remote-head-base", "contract", "spec", "task-results", "check-status", "original-writer", "leader-accept-receipt", "quiescence", "preserved-worktree"},
		Outcomes: []Outcome{{Name: "handoff", Default: true, To: StageReviewing}}},
	{ID: "E04", From: []Stage{StageBuilding}, Needs: []string{"incarnation-check"}, Outcomes: []Outcome{parkWith(ReasonStopUnknown)}},
	// E05: S46 does not name the sub stage after a safe handoff from REVIEWING;
	// IMPLEMENT is this code's choice (canon-unspecified).
	{ID: "E05", From: []Stage{StageBuilding, StageReviewing}, Needs: []string{"quiescence", "remaining-effects-cleared", "original-preserved"},
		Outcomes: []Outcome{{Name: "reassign", Default: true, To: StageBuilding}}},
	{ID: "E06", From: []Stage{StageReviewing}, Needs: []string{"repair-decision", "exclusive-owner", "scope-reserved", "old-writer-ended"},
		Outcomes: []Outcome{{Name: "repair", Default: true, To: StageBuilding, Sub: SubFix}}},
	{ID: "E07", From: []Stage{StageReviewing}, Needs: []string{"advisory-fallback-allowed"}, Outcomes: []Outcome{stay("fallback")}},
	{ID: "E08", From: []Stage{StageReviewing}, Needs: []string{"same-head-base-contract", "gates", "handoff", "protection", "approval"},
		Outcomes: []Outcome{{Name: "merge", Default: true, To: StageMerging}}},
	{ID: "E09", From: []Stage{StageReviewing}, Needs: []string{"provider-fact-diff"}, Outcomes: []Outcome{stay("invalidate")}},
	{ID: "E10", From: []Stage{StageMerging},
		Outcomes: []Outcome{{Name: "back", Default: true, To: StageReviewing, Needs: []string{"remote-not-merged", "sender-cleared"}}, {Name: "unknown", Park: ReasonUnknownEffect}}},
	{ID: "E11", From: []Stage{StageMerging}, Needs: []string{"effect-intent", "sender-check"}, Outcomes: []Outcome{parkWith(ReasonUnknownEffect)}},
	{ID: "E12", From: []Stage{StageMerging}, Needs: []string{"pr-head-merge-target"},
		Outcomes: []Outcome{{Name: "integrated", Default: true, Integrated: true}}},
	{ID: "E13", From: []Stage{StageMerging}, RequiresIntegrated: true, Needs: []string{"acceptance-evidence", "merge-commit", "local-sync", "close", "cleanup"},
		Outcomes: []Outcome{{Name: "done", Default: true, To: StageDone}}},
	{ID: "E14", From: []Stage{StageMerging}, Needs: []string{"merge-fact"}, Outcomes: []Outcome{parkWith(ReasonPendingEffect)}},
	// E15 (canon line 61): an unclear cleanup stays at the current stage, parked.
	{ID: "E15", From: []Stage{StageReady, StageBuilding, StageReviewing}, Needs: []string{"new-grants-blocked"},
		Outcomes: []Outcome{{Name: "cancel", Default: true, To: StageCancelled, Needs: []string{"writers-effects-cleaned"}}, {Name: "cleanup-unknown", Park: ReasonFromCaller}}},
	{ID: "E16", From: []Stage{StageMerging},
		Outcomes: []Outcome{{Name: "cancel", Default: true, To: StageCancelled, Needs: []string{"sender-residue-removed", "queue-effects-checked"}}, {Name: "unknown", Park: ReasonUnknownEffect}}},
	// E17 is a fact: the integration is kept, nothing is reverted.
	{ID: "E17", From: []Stage{StageMerging, StageDone}, Needs: []string{"provider-merge-fact"}, Outcomes: []Outcome{{Name: "keep", Default: true, Integrated: true}}},
	// E18: S46 says "PARKED"; contract-change as its reason comes from E24 and
	// SPEC 0.6.1 (canon-unspecified for E18 itself).
	// "adopt" (leader decision on canon rows 192/2760, "E18 후 SPEC단계"): the
	// approved revision is applied for a new attempt, which takes a contract-change
	// parked BUILDING/FIX back to BUILDING/SPEC. E20 alone never does that.
	{ID: "E18", From: nonterminal, Needs: []string{"contract-diff"}, Outcomes: []Outcome{parkWith(ReasonContractChange),
		{Name: "adopt", To: StageBuilding, Sub: SubSpec, Adopt: true, Needs: []string{"approved-revision"}}}},
	// E19 (canon line 65): quota, permission or budget block; the guard is the
	// classified error, so the reason must come from the classification set.
	{ID: "E19", From: nonterminal, Needs: []string{"error-classification"}, Outcomes: []Outcome{parkWith(ReasonClassified)}},
	{ID: "E21", From: []Stage{StageDone, StageCancelled}, Needs: []string{"id-epoch-receipt"}, Outcomes: []Outcome{stay("reject-late")}},
	{ID: "E22", From: []Stage{StageDone, StageCancelled}, Needs: []string{"new-intent"}, Outcomes: []Outcome{stay("separate-lifecycle")}},
	{ID: "E23", From: []Stage{StageBuilding}, FromSub: []SubStage{SubSpec}, Needs: []string{"spec-revision", "selected-auditor", "required-coverage"},
		Outcomes: []Outcome{{Name: "pass", Default: true, To: StageBuilding, Sub: SubImplement}, {Name: "fail", To: StageBuilding, Sub: SubSpec}, {Name: "inconclusive", Park: ReasonFromCaller}}},
	// E24 (canon line 70): the current position is "BUILDING/FIX".
	{ID: "E24", From: []Stage{StageBuilding}, FromSub: []SubStage{SubFix}, Needs: []string{"contract-impact"}, Outcomes: []Outcome{parkWith(ReasonContractChange)}},
	{ID: "E25", From: []Stage{StageBuilding}, Needs: []string{"checkpoint", "quiescence", "scope-returned", "full-reservation"}, Outcomes: []Outcome{stay("rereserve")}},
	{ID: "E26", From: nonterminal, Needs: []string{"launch-key-unknown"}, Outcomes: []Outcome{parkWith(ReasonStartUnknown)}},
	{ID: "E27", From: anyStage, Needs: []string{"verified-producer"}, Outcomes: []Outcome{stay("ingest")}},
}

// resume is E20: it clears the overlay and nothing else.
var resumeEdge = Edge{ID: "E20", From: anyStage, Outcomes: []Outcome{stay("resume")}}

// Edges returns the full table, E20 included, in ID order.
func Edges() []Edge {
	out := append([]Edge{resumeEdge}, edges...)
	sort.Slice(out, func(i, k int) bool { return out[i].ID < out[k].ID })
	return out
}

func EdgeFor(id string) (Edge, bool) {
	if id == "E20" {
		return resumeEdge, true
	}
	for _, e := range edges {
		if e.ID == id {
			return e, true
		}
	}
	return Edge{}, false
}

// Event asks for one transition. Park is read only by an outcome that parks.
type Event struct {
	ID       string
	Outcome  string
	Evidence []string
	Park     Park
}

func missing(needs, have []string) []string {
	set := map[string]bool{}
	for _, h := range have {
		set[h] = true
	}
	var out []string
	for _, n := range needs {
		if !set[n] {
			out = append(out, n)
		}
	}
	return out
}

// Transition applies one S46 edge. It checks the edge exists, the position may
// take it, every guard key has evidence, and the result is a valid position.
func Transition(pos Position, ev Event) (Position, error) {
	e, ok := EdgeFor(ev.ID)
	if !ok {
		return pos, fmt.Errorf("%w: %q", ErrUnknownEdge, ev.ID)
	}
	if err := pos.Validate(); err != nil {
		return pos, err
	}
	if e.ID == "E20" {
		if pos.Park == nil {
			return pos, ErrNotParked
		}
		if m := missing([]string{"resume-guard:" + string(pos.Park.Reason)}, ev.Evidence); len(m) > 0 {
			return pos, fmt.Errorf("%w: %s", ErrGuardMissing, strings.Join(m, ", "))
		}
		pos.Park = nil
		return pos, nil
	}
	if !e.allows(pos) {
		return pos, fmt.Errorf("%w: %s is not allowed from %s/%s", ErrTransitionInvalid, e.ID, pos.Stage, pos.Sub)
	}
	var out *Outcome
	if ev.Outcome == "" {
		out = e.defaultOutcome()
	}
	for i := range e.Outcomes {
		if ev.Outcome != "" && e.Outcomes[i].Name == ev.Outcome {
			out = &e.Outcomes[i]
		}
	}
	if out == nil {
		return pos, fmt.Errorf("%w: %s has no outcome %q", ErrUnknownOutcome, e.ID, ev.Outcome)
	}
	if m := missing(e.needsFor(out.Name), ev.Evidence); len(m) > 0 {
		return pos, fmt.Errorf("%w: %s needs %s", ErrGuardMissing, e.ID, strings.Join(m, ", "))
	}
	if out.Adopt && (pos.Stage != StageBuilding || pos.Sub != SubFix || pos.Park == nil || pos.Park.Reason != ReasonContractChange) {
		return pos, fmt.Errorf("%w: %s %s needs a contract-change parked BUILDING/FIX, not %s/%s",
			ErrTransitionInvalid, e.ID, out.Name, pos.Stage, pos.Sub)
	}
	next := pos
	if out.Adopt {
		next.Park = nil
	}
	if out.To != "" {
		next.Stage = out.To
		if out.To != StageBuilding {
			next.Sub = ""
		} else if out.Sub != "" {
			next.Sub = out.Sub
		} else if next.Sub == "" {
			next.Sub = SubImplement
		}
	}
	if out.Integrated && next.Stage == StageMerging {
		next.RemoteIntegrated = true
	}
	if next.Stage != StageMerging {
		next.RemoteIntegrated = false
	}
	// Fail closed (leader decision): a parked position may not move to another
	// stage, so a run stopped for unknown-effect cannot slide into MERGING. The
	// way out is E20, cancellation (E15/E16), or an edge that leaves the stage as
	// it is (facts, and a park update). E13 (DONE) is refused too, because DONE
	// needs the parked obligation resolved first; S46 is silent here (canon-unspecified).
	if pos.Park != nil && !out.Adopt && (next.Stage != pos.Stage || next.Sub != pos.Sub) && next.Stage != StageCancelled {
		return pos, fmt.Errorf("%w: %s moves %s/%s on while parked (%s); resume with E20 first",
			ErrTransitionInvalid, e.ID, pos.Stage, pos.Sub, pos.Park.Reason)
	}
	if next.Stage.terminal() {
		next.Park = nil // canon-unspecified: finishing or cancelling resolves the overlay
	}
	if out.Park != "" {
		// A park edge on a parked position replaces the overlay: one reason at a time.
		reason := out.Park
		if out.Park == ReasonFromCaller {
			reason = ev.Park.Reason
		} else if out.Park == ReasonClassified {
			reason = ev.Park.Reason
			if !isClassified(reason) {
				return pos, fmt.Errorf("%w: %s takes a classified error reason, not %q", ErrReasonNotAllowed, e.ID, reason)
			}
		} else if ev.Park.Reason != "" && ev.Park.Reason != out.Park {
			return pos, fmt.Errorf("%w: %s parks with %s, not %q", ErrReasonNotAllowed, e.ID, out.Park, ev.Park.Reason)
		}
		p := Park{Reason: reason, Evidence: ev.Park.Evidence, Resume: ev.Park.Resume}
		if err := p.Validate(); err != nil {
			return pos, err
		}
		// A reason that says "an outcome or state is unknown" is settled only by
		// its own resume (E20). Another reason must not overwrite the flag; the
		// same reason may refresh its evidence (leader decision).
		if pos.Park != nil && unknownFamily(pos.Park.Reason) && reason != pos.Park.Reason {
			return pos, fmt.Errorf("%w: %s cannot replace the parked %s with %s; resume with E20 first",
				ErrTransitionInvalid, e.ID, pos.Park.Reason, reason)
		}
		next.Park = &p
	}
	if err := next.Validate(); err != nil {
		return pos, err
	}
	return next, nil
}

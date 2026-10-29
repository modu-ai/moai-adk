package core

import (
	"errors"
	"sort"
	"strings"
	"testing"
)

var allStages = []Stage{StageReady, StageBuilding, StageReviewing, StageMerging, StageDone, StageCancelled}

func park(r BlockReason) Park { return Park{Reason: r, Evidence: "ev", Resume: "when fixed"} }

func TestParkReasonsCanonSixAreExactStrings(t *testing.T) {
	want := map[BlockReason]string{
		ReasonDependency: "dependency", ReasonStopUnknown: "stop_unknown", ReasonUnknownEffect: "unknown-effect",
		ReasonPendingEffect: "pending-effect", ReasonContractChange: "contract-change", ReasonStartUnknown: "start_unknown",
	}
	for r, s := range want {
		if string(r) != s {
			t.Errorf("%q != %q (S46 spells stop_unknown/start_unknown with underscores)", r, s)
		}
		if !IsCanonParkReason(r) {
			t.Errorf("%s must be a canon park reason", r)
		}
	}
	if len(CanonParkReasons()) != 6 {
		t.Errorf("canon set = %v", CanonParkReasons())
	}
}

func TestOperatorHoldIsParkableButMarkedCanonUnconfirmed(t *testing.T) {
	if IsCanonParkReason(ReasonOperatorHold) {
		t.Error("operator-hold is not in S46 (OD-12); it must not read as canon")
	}
	if got := CanonUnconfirmedParkReasons(); len(got) != 1 || got[0] != ReasonOperatorHold {
		t.Errorf("unconfirmed set = %v", got)
	}
	if err := park(ReasonOperatorHold).Validate(); err != nil {
		t.Errorf("operator-hold is still a valid overlay reason until OD-12: %v", err)
	}
}

func TestParkValidate(t *testing.T) {
	for _, r := range append(CanonParkReasons(), ReasonRateLimited, ReasonStoreUnavailable, ReasonUserControl, ReasonCapacityShort) {
		if err := park(r).Validate(); err != nil {
			t.Errorf("%s: %v", r, err)
		}
	}
	// refusal (R) and merge-gate (G) reasons never become an overlay
	for _, r := range []BlockReason{ReasonNotIssueAuthor, ReasonIdentityUnknown, ReasonRepoNotAllowed, ReasonPermissionNotGranted,
		ReasonOwnerUnconfirmed, ReasonForeignClosing, ReasonOpNotDefined, ReasonCheckNotPassed, ReasonDebtPresent,
		ReasonVerdictNotPass, ReasonProducerNotAllowed, ReasonOwnershipUnverified, "made-up"} {
		if err := park(r).Validate(); !errors.Is(err, ErrReasonNotAllowed) {
			t.Errorf("%s: %v", r, err)
		}
	}
	if err := (Park{Reason: ReasonDependency, Resume: "x"}).Validate(); !errors.Is(err, ErrEvidenceIncomplete) {
		t.Errorf("no evidence: %v", err)
	}
	if err := (Park{Reason: ReasonDependency, Evidence: "x"}).Validate(); !errors.Is(err, ErrEvidenceIncomplete) {
		t.Errorf("no resume condition: %v", err)
	}
}

func TestPositionValidate(t *testing.T) {
	good := []Position{
		{Stage: StageReady}, {Stage: StageBuilding, Sub: SubSpec}, {Stage: StageReviewing}, {Stage: StageMerging, RemoteIntegrated: true},
		{Stage: StageDone}, {Stage: StageCancelled}, {Stage: StageBuilding, Sub: SubFix, Park: ptr(park(ReasonStopUnknown))},
	}
	for _, p := range good {
		if err := p.Validate(); err != nil {
			t.Errorf("%+v: %v", p, err)
		}
	}
	bad := map[string]Position{
		"building without sub":   {Stage: StageBuilding},
		"sub outside building":   {Stage: StageReviewing, Sub: SubFix},
		"integrated outside mrg": {Stage: StageReviewing, RemoteIntegrated: true},
		"park on done":           {Stage: StageDone, Park: ptr(park(ReasonDependency))},
		"park on cancelled":      {Stage: StageCancelled, Park: ptr(park(ReasonDependency))},
		"invalid park":           {Stage: StageReady, Park: &Park{Reason: "made-up"}},
		"unknown stage":          {Stage: "BLOCKED"},
	}
	for name, p := range bad {
		if err := p.Validate(); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func ptr(p Park) *Park { return &p }

func TestEdgeTableIsExactlyE01ToE27(t *testing.T) {
	var ids []string
	for _, e := range Edges() {
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	var want []string
	for i := 1; i <= 27; i++ {
		want = append(want, "E"+pad(i))
	}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Errorf("edges = %v", ids)
	}
}

func pad(i int) string {
	if i < 10 {
		return "0" + string(rune('0'+i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

// evidenceFor returns every guard key an edge needs.
func evidenceFor(e Edge) []string { return append([]string(nil), e.Needs...) }

func posFor(st Stage, sub SubStage) Position {
	p := Position{Stage: st, Sub: sub}
	if st == StageBuilding && sub == "" {
		p.Sub = SubImplement
	}
	return p
}

func TestEveryEdgeAcceptsExactlyItsFromStages(t *testing.T) {
	for _, e := range Edges() {
		if e.ID == "E20" {
			continue // resume needs a parked position; covered below
		}
		for _, st := range allStages {
			subs := []SubStage{""}
			if st == StageBuilding {
				subs = []SubStage{SubSpec, SubImplement, SubFix}
			}
			for _, sub := range subs {
				pos := posFor(st, sub)
				pos.Sub = sub
				if st == StageBuilding && sub == "" {
					continue
				}
				if st == StageMerging && e.ID == "E13" {
					pos.RemoteIntegrated = true
				}
				ev := Event{ID: e.ID, Evidence: evidenceFor(e)}
				if o := e.defaultOutcome(); o != nil && o.Park != "" {
					r := o.Park
					if r == ReasonFromCaller {
						r = ReasonRateLimited
					}
					ev.Park = park(r)
				}
				_, err := Transition(pos, ev)
				inFrom := e.allows(pos)
				if inFrom && err != nil {
					t.Errorf("%s from %s/%s refused: %v", e.ID, st, sub, err)
				}
				if !inFrom && !errors.Is(err, ErrTransitionInvalid) {
					t.Errorf("%s from %s/%s: want ErrTransitionInvalid, got %v", e.ID, st, sub, err)
				}
			}
		}
	}
}

func TestGuardEvidenceIsRequiredAndNamed(t *testing.T) {
	e, _ := EdgeFor("E01")
	if len(e.Needs) == 0 {
		t.Fatal("E01 has no guard")
	}
	for i := range e.Needs {
		ev := append(append([]string(nil), e.Needs[:i]...), e.Needs[i+1:]...)
		_, err := Transition(Position{Stage: StageReady}, Event{ID: "E01", Evidence: ev})
		if !errors.Is(err, ErrGuardMissing) || !strings.Contains(err.Error(), e.Needs[i]) {
			t.Errorf("without %q: %v", e.Needs[i], err)
		}
	}
	if _, err := Transition(Position{Stage: StageReady}, Event{ID: "E99"}); !errors.Is(err, ErrUnknownEdge) {
		t.Errorf("unknown edge: %v", err)
	}
}

func run(t *testing.T, pos Position, id, outcome string, p Park) (Position, error) {
	t.Helper()
	e, ok := EdgeFor(id)
	if !ok {
		t.Fatalf("no edge %s", id)
	}
	return Transition(pos, Event{ID: id, Outcome: outcome, Evidence: evidenceFor(e), Park: p})
}

func TestNamedEdgeResults(t *testing.T) {
	type c struct {
		name    string
		from    Position
		id, out string
		park    Park
		want    Position
		reason  BlockReason
	}
	cases := []c{
		{"E01 claim", Position{Stage: StageReady}, "E01", "", Park{}, Position{Stage: StageBuilding, Sub: SubSpec}, ""},
		{"E02 dependency", Position{Stage: StageReady}, "E02", "", park(ReasonDependency), Position{Stage: StageReady}, ReasonDependency},
		{"E03 handoff", Position{Stage: StageBuilding, Sub: SubImplement}, "E03", "", Park{}, Position{Stage: StageReviewing}, ""},
		{"E04 stop unknown", Position{Stage: StageBuilding, Sub: SubImplement}, "E04", "", park(ReasonStopUnknown), Position{Stage: StageBuilding, Sub: SubImplement}, ReasonStopUnknown},
		{"E06 repair", Position{Stage: StageReviewing}, "E06", "", Park{}, Position{Stage: StageBuilding, Sub: SubFix}, ""},
		{"E08 merge", Position{Stage: StageReviewing}, "E08", "", Park{}, Position{Stage: StageMerging}, ""},
		{"E10 back to review", Position{Stage: StageMerging}, "E10", "", Park{}, Position{Stage: StageReviewing}, ""},
		{"E10 unknown", Position{Stage: StageMerging}, "E10", "unknown", park(ReasonUnknownEffect), Position{Stage: StageMerging}, ReasonUnknownEffect},
		{"E11 send unknown", Position{Stage: StageMerging}, "E11", "", park(ReasonUnknownEffect), Position{Stage: StageMerging}, ReasonUnknownEffect},
		{"E12 integrated", Position{Stage: StageMerging}, "E12", "", Park{}, Position{Stage: StageMerging, RemoteIntegrated: true}, ""},
		{"E13 done", Position{Stage: StageMerging, RemoteIntegrated: true}, "E13", "", Park{}, Position{Stage: StageDone}, ""},
		{"E14 pending effect", Position{Stage: StageMerging, RemoteIntegrated: true}, "E14", "", park(ReasonPendingEffect), Position{Stage: StageMerging, RemoteIntegrated: true}, ReasonPendingEffect},
		{"E15 cancel", Position{Stage: StageBuilding, Sub: SubFix}, "E15", "", Park{}, Position{Stage: StageCancelled}, ""},
		{"E16 cancel clean", Position{Stage: StageMerging}, "E16", "", Park{}, Position{Stage: StageCancelled}, ""},
		{"E16 cancel unknown", Position{Stage: StageMerging}, "E16", "unknown", park(ReasonUnknownEffect), Position{Stage: StageMerging}, ReasonUnknownEffect},
		{"E18 contract change", Position{Stage: StageReviewing}, "E18", "", park(ReasonContractChange), Position{Stage: StageReviewing}, ReasonContractChange},
		{"E23 pass", Position{Stage: StageBuilding, Sub: SubSpec}, "E23", "pass", Park{}, Position{Stage: StageBuilding, Sub: SubImplement}, ""},
		{"E23 fail stays in spec", Position{Stage: StageBuilding, Sub: SubSpec}, "E23", "fail", Park{}, Position{Stage: StageBuilding, Sub: SubSpec}, ""},
		{"E24 spec change proposal", Position{Stage: StageBuilding, Sub: SubFix}, "E24", "", park(ReasonContractChange), Position{Stage: StageBuilding, Sub: SubFix}, ReasonContractChange},
		{"E26 start unknown", Position{Stage: StageBuilding, Sub: SubSpec}, "E26", "", park(ReasonStartUnknown), Position{Stage: StageBuilding, Sub: SubSpec}, ReasonStartUnknown},
	}
	for _, c := range cases {
		got, err := run(t, c.from, c.id, c.out, c.park)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		parked := got.Park
		got.Park = nil
		if got != c.want {
			t.Errorf("%s: got %+v want %+v", c.name, got, c.want)
		}
		switch {
		case c.reason == "" && parked != nil:
			t.Errorf("%s: unexpected overlay %+v", c.name, parked)
		case c.reason != "" && (parked == nil || parked.Reason != c.reason):
			t.Errorf("%s: overlay = %+v want %s", c.name, parked, c.reason)
		}
	}
}

func TestE13NeedsTheRemoteIntegratedMark(t *testing.T) {
	if _, err := run(t, Position{Stage: StageMerging}, "E13", "", Park{}); !errors.Is(err, ErrTransitionInvalid) {
		t.Errorf("done before the merge was observed: %v", err)
	}
}

func TestE23OnlyFromTheSpecSubStage(t *testing.T) {
	if _, err := run(t, Position{Stage: StageBuilding, Sub: SubImplement}, "E23", "pass", Park{}); !errors.Is(err, ErrTransitionInvalid) {
		t.Errorf("spec audit result outside SPEC: %v", err)
	}
	if _, err := run(t, Position{Stage: StageBuilding, Sub: SubSpec}, "E23", "maybe", Park{}); !errors.Is(err, ErrUnknownOutcome) {
		t.Errorf("unknown outcome: %v", err)
	}
}

func TestResumeClearsOnlyTheOverlay(t *testing.T) {
	from := Position{Stage: StageBuilding, Sub: SubFix, Park: ptr(park(ReasonStopUnknown))}
	if _, err := Transition(from, Event{ID: "E20"}); !errors.Is(err, ErrGuardMissing) {
		t.Errorf("resume without the reason's guard: %v", err)
	}
	got, err := Transition(from, Event{ID: "E20", Evidence: []string{"resume-guard:stop_unknown"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Park != nil || got.Stage != StageBuilding || got.Sub != SubFix {
		t.Errorf("resume changed more than the overlay: %+v", got)
	}
	if _, err := Transition(Position{Stage: StageReady}, Event{ID: "E20"}); !errors.Is(err, ErrNotParked) {
		t.Errorf("resume of a position that is not parked: %v", err)
	}
	if _, err := Transition(from, Event{ID: "E20", Evidence: []string{"resume-guard:dependency"}}); !errors.Is(err, ErrGuardMissing) {
		t.Errorf("another reason's guard must not resume: %v", err)
	}
}

func TestAParkedPositionHasExactlyOneReasonAndTerminalsAreNeverParked(t *testing.T) {
	parked := Position{Stage: StageReady, Park: ptr(park(ReasonRateLimited))}
	got, err := run(t, parked, "E19", "", park(ReasonCapacityShort))
	if err != nil || got.Park == nil || got.Park.Reason != ReasonCapacityShort || got.Stage != StageReady {
		t.Errorf("a park update replaces the reason, never stacks a second one: %+v %v", got, err)
	}
	for _, id := range []string{"E18", "E19", "E26"} {
		for _, st := range []Stage{StageDone, StageCancelled} {
			if _, err := run(t, Position{Stage: st}, id, "", park(ReasonRateLimited)); !errors.Is(err, ErrTransitionInvalid) {
				t.Errorf("%s on %s: %v", id, st, err)
			}
		}
	}
}

// An unknown outcome or state is only ever settled by its own resume: another
// reason must not overwrite the flag, the same reason may refresh its evidence.
func TestAnUnknownFamilyReasonIsNotReplacedByAnotherReason(t *testing.T) {
	for _, sticky := range []BlockReason{ReasonUnknownEffect, ReasonPendingEffect, ReasonStopUnknown, ReasonStartUnknown} {
		base := Position{Stage: StageBuilding, Sub: SubImplement, Park: &Park{Reason: sticky, Evidence: "first", Resume: "resume when known"}}
		_, err := run(t, base, "E19", "", park(ReasonRateLimited))
		if !errors.Is(err, ErrTransitionInvalid) || !strings.Contains(err.Error(), string(sticky)) {
			t.Errorf("E19 over %s: %v", sticky, err)
		}
		if base.Park.Reason != sticky || base.Park.Evidence != "first" {
			t.Errorf("%s: the refused edge changed the overlay", sticky)
		}
	}
	// the same reason may refresh its evidence
	base := Position{Stage: StageMerging, Park: &Park{Reason: ReasonUnknownEffect, Evidence: "first", Resume: "r"}}
	got, err := run(t, base, "E11", "", Park{Reason: ReasonUnknownEffect, Evidence: "second", Resume: "r2"})
	if err != nil || got.Park == nil || got.Park.Evidence != "second" || got.Park.Reason != ReasonUnknownEffect {
		t.Errorf("same-reason evidence update: %+v %v", got, err)
	}
	// reasons outside that family may be replaced by any parkable reason
	for _, plain := range []BlockReason{ReasonRateLimited, ReasonDependency, ReasonContractChange, ReasonStoreUnavailable} {
		base := Position{Stage: StageReady, Park: ptr(park(plain))}
		got, err := run(t, base, "E19", "", park(ReasonCapacityShort))
		if err != nil || got.Park == nil || got.Park.Reason != ReasonCapacityShort {
			t.Errorf("E19 over %s: %+v %v", plain, got, err)
		}
	}
	// rate-limited may become an unknown-family reason
	base = Position{Stage: StageBuilding, Sub: SubImplement, Park: ptr(park(ReasonRateLimited))}
	if got, err := run(t, base, "E26", "", park(ReasonStartUnknown)); err != nil || got.Park.Reason != ReasonStartUnknown {
		t.Errorf("rate-limited -> start_unknown: %+v %v", got, err)
	}
}

// Fail closed: an Issue stopped for unknown-effect must not slide into MERGING.
func TestAParkedPositionRefusesEdgesThatChangeTheStage(t *testing.T) {
	parkedAt := func(st Stage, sub SubStage, integrated bool) Position {
		return Position{Stage: st, Sub: sub, RemoteIntegrated: integrated, Park: ptr(park(ReasonUnknownEffect))}
	}
	refused := []struct {
		id, outcome string
		from        Position
	}{
		{"E01", "", parkedAt(StageReady, "", false)},
		{"E03", "", parkedAt(StageBuilding, SubImplement, false)},
		{"E05", "", parkedAt(StageReviewing, "", false)},
		{"E06", "", parkedAt(StageReviewing, "", false)},
		{"E08", "", parkedAt(StageReviewing, "", false)},
		{"E10", "", parkedAt(StageMerging, "", false)},
		{"E13", "", parkedAt(StageMerging, "", true)},
		{"E23", "pass", parkedAt(StageBuilding, SubSpec, false)},
	}
	for _, c := range refused {
		_, err := run(t, c.from, c.id, c.outcome, park(ReasonUnknownEffect))
		if !errors.Is(err, ErrTransitionInvalid) || !strings.Contains(err.Error(), "parked") {
			t.Errorf("%s while parked: %v", c.id, err)
		}
	}
	allowed := []struct {
		id, outcome string
		from        Position
		check       func(Position) bool
	}{
		{"E15", "", parkedAt(StageBuilding, SubFix, false), func(p Position) bool { return p.Stage == StageCancelled && p.Park == nil }},
		{"E16", "", parkedAt(StageMerging, "", false), func(p Position) bool { return p.Stage == StageCancelled && p.Park == nil }},
		{"E07", "", parkedAt(StageReviewing, "", false), func(p Position) bool { return p.Park != nil && p.Stage == StageReviewing }},
		{"E09", "", parkedAt(StageReviewing, "", false), func(p Position) bool { return p.Park != nil }},
		{"E25", "", parkedAt(StageBuilding, SubImplement, false), func(p Position) bool { return p.Park != nil }},
		{"E27", "", parkedAt(StageReady, "", false), func(p Position) bool { return p.Park != nil }},
		{"E12", "", parkedAt(StageMerging, "", false), func(p Position) bool { return p.RemoteIntegrated && p.Park != nil }},
		{"E05", "", parkedAt(StageBuilding, SubFix, false), func(p Position) bool { return p.Stage == StageBuilding && p.Sub == SubFix && p.Park != nil }},
		{"E11", "", parkedAt(StageMerging, "", false), func(p Position) bool { return p.Park != nil && p.Park.Reason == ReasonUnknownEffect }},
	}
	for _, c := range allowed {
		got, err := run(t, c.from, c.id, c.outcome, park(ReasonUnknownEffect))
		if err != nil || !c.check(got) {
			t.Errorf("%s while parked: %+v %v", c.id, got, err)
		}
	}
	// resume is always allowed, then the stage edge works
	resumed, err := Transition(parkedAt(StageReviewing, "", false), Event{ID: "E20", Evidence: []string{"resume-guard:unknown-effect"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, resumed, "E08", "", Park{}); err != nil {
		t.Errorf("after the resume: %v", err)
	}
}

func TestCallerSuppliedAndFixedReasons(t *testing.T) {
	// E19 takes the classified reason from the caller, but only a parkable one
	if _, err := run(t, Position{Stage: StageReady}, "E19", "", park(ReasonRateLimited)); err != nil {
		t.Errorf("E19 rate-limited: %v", err)
	}
	if _, err := run(t, Position{Stage: StageReady}, "E19", "", park(ReasonNotIssueAuthor)); !errors.Is(err, ErrReasonNotAllowed) {
		t.Errorf("E19 with a refusal reason: %v", err)
	}
	if _, err := run(t, Position{Stage: StageReady}, "E19", "", Park{}); err == nil {
		t.Error("E19 without a park accepted")
	}
	// a fixed-reason edge refuses a different reason from the caller
	if _, err := run(t, Position{Stage: StageReady}, "E02", "", park(ReasonRateLimited)); !errors.Is(err, ErrReasonNotAllowed) {
		t.Errorf("E02 with another reason: %v", err)
	}
	// E15 cleanup-unknown and E23 inconclusive leave the reason to the caller
	if got, err := run(t, Position{Stage: StageReviewing}, "E15", "cleanup-unknown", park(ReasonStopUnknown)); err != nil || got.Park == nil || got.Stage != StageReviewing {
		t.Errorf("E15 cleanup unknown: %+v %v", got, err)
	}
	if got, err := run(t, Position{Stage: StageBuilding, Sub: SubSpec}, "E23", "inconclusive", park(ReasonStopUnknown)); err != nil || got.Park == nil {
		t.Errorf("E23 inconclusive: %+v %v", got, err)
	}
}

func TestFactsOnlyEdgesChangeNothing(t *testing.T) {
	for _, id := range []string{"E07", "E09", "E21", "E22", "E27", "E25"} {
		e, _ := EdgeFor(id)
		pos := Position{Stage: e.From[0]}
		if pos.Stage == StageBuilding {
			pos.Sub = SubImplement
		}
		if e.FromSub != nil {
			pos.Sub = e.FromSub[0]
		}
		got, err := run(t, pos, id, "", Park{})
		if err != nil || got != pos {
			t.Errorf("%s: %+v %v", id, got, err)
		}
	}
}

func TestFromLegacyMapping(t *testing.T) {
	cases := []struct {
		name string
		in   Legacy
		want Position
		park BlockReason
	}{
		{"queued", Legacy{State: StateQueued}, Position{Stage: StageReady}, ""},
		{"claimed", Legacy{State: StateClaimed}, Position{Stage: StageBuilding, Sub: SubSpec}, ""},
		{"claimed repair claim", Legacy{State: StateClaimed, HasReceipt: true}, Position{Stage: StageBuilding, Sub: SubFix}, ""},
		{"implementing", Legacy{State: StateImplementing}, Position{Stage: StageBuilding, Sub: SubImplement}, ""},
		{"implementing after reopen", Legacy{State: StateImplementing, HasReceipt: true}, Position{Stage: StageBuilding, Sub: SubFix}, ""},
		{"verifying has no stage of its own", Legacy{State: StateVerifying}, Position{Stage: StageBuilding, Sub: SubImplement}, ""},
		{"ready has no stage of its own", Legacy{State: StateReady}, Position{Stage: StageBuilding, Sub: SubImplement}, ""},
		{"ready with an accepted receipt is REVIEWING", Legacy{State: StateReady, HasReceipt: true, LeaderAccepted: true}, Position{Stage: StageReviewing}, ""},
		{"ready with a receipt only is still BUILDING", Legacy{State: StateReady, HasReceipt: true}, Position{Stage: StageBuilding, Sub: SubFix}, ""},
		{"verifying with an accepted receipt is REVIEWING", Legacy{State: StateVerifying, HasReceipt: true, LeaderAccepted: true}, Position{Stage: StageReviewing}, ""},
		{"implementing with an accepted receipt is REVIEWING", Legacy{State: StateImplementing, HasReceipt: true, LeaderAccepted: true}, Position{Stage: StageReviewing}, ""},
		{"verifying with a merge seen is MERGING", Legacy{State: StateVerifying, HasReceipt: true, LeaderAccepted: true, Merged: true}, Position{Stage: StageMerging, RemoteIntegrated: true}, ""},
		{"landing before the leader accepted", Legacy{State: StateLanding}, Position{Stage: StageBuilding, Sub: SubImplement}, ""},
		{"landing with accepted receipt", Legacy{State: StateLanding, HasReceipt: true, LeaderAccepted: true}, Position{Stage: StageReviewing}, ""},
		{"landing merged", Legacy{State: StateLanding, HasReceipt: true, LeaderAccepted: true, Merged: true}, Position{Stage: StageMerging, RemoteIntegrated: true}, ""},
		{"done", Legacy{State: StateDone}, Position{Stage: StageDone}, ""},
		{"blocked keeps its stage", Legacy{State: StateBlocked, Return: StateImplementing, Reason: ReasonRateLimited, Resume: "after reset"}, Position{Stage: StageBuilding, Sub: SubImplement}, ReasonRateLimited},
		{"blocked at landing", Legacy{State: StateBlocked, Return: StateLanding, Reason: ReasonStoreUnavailable, Resume: "cas ok", HasReceipt: true, LeaderAccepted: true}, Position{Stage: StageReviewing}, ReasonStoreUnavailable},
	}
	for _, c := range cases {
		got, err := FromLegacy(c.in)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		parked := got.Park
		got.Park = nil
		if got != c.want {
			t.Errorf("%s: got %+v want %+v", c.name, got, c.want)
		}
		if c.park == "" && parked != nil || c.park != "" && (parked == nil || parked.Reason != c.park || parked.Resume == "") {
			t.Errorf("%s: overlay %+v want %s", c.name, parked, c.park)
		}
		if err := got.Validate(); err != nil {
			t.Errorf("%s: mapped position invalid: %v", c.name, err)
		}
	}
}

func TestFromLegacyOnHoldMapsToOperatorHoldCanonUnconfirmed(t *testing.T) {
	got, err := FromLegacy(Legacy{State: StateOnHold, Return: StateVerifying, Reason: ReasonOperatorHold, Resume: "operator resumes"})
	if err != nil || got.Park == nil || got.Park.Reason != ReasonOperatorHold || got.Stage != StageBuilding {
		t.Fatalf("on-hold: %+v %v", got, err)
	}
	if IsCanonParkReason(got.Park.Reason) {
		t.Error("the mapped reason must stay marked as not canon (OD-12)")
	}
}

func TestFromLegacyRefusesWhatItCannotMap(t *testing.T) {
	bad := []Legacy{
		{State: "mystery"},
		{State: StateBlocked, Reason: ReasonRateLimited, Resume: "x"},                               // no return stage
		{State: StateBlocked, Return: StateImplementing, Reason: ReasonNotIssueAuthor, Resume: "x"}, // refusal reason cannot be an overlay
		{State: StateBlocked, Return: StateBlocked, Reason: ReasonRateLimited, Resume: "x"},         // pause inside a pause
		{State: StateBlocked, Return: StateImplementing, Reason: ReasonRateLimited},                 // no resume condition
		{State: StateBlocked, Return: StateDone, Reason: ReasonRateLimited, Resume: "x"},            // nothing parks a finished record
	}
	for i, in := range bad {
		if _, err := FromLegacy(in); !errors.Is(err, ErrLegacyUnmappable) {
			t.Errorf("case %d: %v", i, err)
		}
	}
}

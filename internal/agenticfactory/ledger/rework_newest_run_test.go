package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, the newest re-measurement: a card whose landed commit c1 is measured again after
// the move back finishes on that measurement (decision in rework_receipt_test.go). A measurement
// that failed is not a measurement that passed: when the NEWEST run of c1 recorded after the move
// back, of a kind that was run, is a failure, the card does not finish, whichever order the runs
// came in. The rule is the ledger's own "what counts": for each kind (green, gate) the newest run
// of the commit decides (factsTx for the receipt, latestResult at Enqueue); at least one passing
// run must exist after the move back, and none of the kinds that was run again may end failing.

type rerun struct{ kind, result string }

var (
	gp = rerun{KindGreen, ResultPass}
	gf = rerun{KindGreen, ResultFail}
	hp = rerun{KindGate, ResultPass}
	hf = rerun{KindGate, ResultFail}
)

func addRuns(t *testing.T, s *Store, e int64, runs []rerun) {
	t.Helper()
	for _, r := range runs {
		if err := s.AddEvidence(bg, "a", e, Evidence{Kind: r.kind, TreeSHA: "tree-c1", SubjectSHA: "c1", Command: "re-run", Result: r.result}); err != nil {
			t.Fatalf("evidence %s %s: %v", r.kind, r.result, err)
		}
	}
}

var newestRunCases = []struct {
	name    string
	runs    []rerun
	finish  bool // true: the card may finish; false: DONE is refused on every path
	comment string
}{
	// Control and the allowed cases.
	{"control: green pass and gate pass", []rerun{gp, hp}, true, ""},
	{"control: gate pass alone", []rerun{hp}, true, ""},
	{"green fail, then green pass", []rerun{gf, gp}, true, "a failing run followed by a passing one"},
	{"gate fail, then gate pass", []rerun{hf, hp}, true, "a failing run followed by a passing one"},
	// The four orders from the leader's reproduction.
	{"green pass, then green fail", []rerun{gp, gf}, false, ""},
	{"gate pass, then gate fail", []rerun{hp, hf}, false, ""},
	{"green pass, then gate fail", []rerun{gp, hf}, false, ""},
	{"gate pass, then green fail", []rerun{hp, gf}, false, ""},
	// A failing kind is not undone by a pass of the other kind.
	{"green fail, then gate pass", []rerun{gf, hp}, false, "the newest green run is a failure"},
	{"gate fail, then green pass", []rerun{hf, gp}, false, "the newest gate run is a failure"},
	// Fail first, both kinds repaired.
	{"green fail and gate fail, then both pass", []rerun{gf, hf, gp, hp}, true, ""},
	// A failure only, no passing run at all (already refused; kept as a guard).
	{"green fail alone", []rerun{gf}, false, ""},
}

// The three ways a card reaches the DONE attempt after the move back.
const (
	pathBuilding        = "straight from BUILDING"
	pathRunsThenHandOff = "runs, then the hand-off, then REVIEWING"
	pathHandOffThenRun  = "hand-off on the old receipt, then the runs in REVIEWING"
)

func TestMove_DoneRefusedWhenTheNewestRemeasurementOfTheLandedCommitFailed(t *testing.T) {
	done := core.Position{Stage: core.StageDone}
	for _, tc := range newestRunCases {
		for _, path := range []string{pathBuilding, pathRunsThenHandOff, pathHandOffThenRun} {
			for _, restart := range []bool{false, true} {
				name := tc.name + " / " + path
				if restart {
					name += " / after a process restart"
				}
				t.Run(name, func(t *testing.T) {
					s, file, e := landedBackedOutAt(t)
					walk(t, s, e, []core.Position{posImplement})
					handedOff := false
					switch path {
					case pathBuilding:
						addRuns(t, s, e, tc.runs)
					case pathRunsThenHandOff:
						addRuns(t, s, e, tc.runs)
						err := s.Move(bg, "a", e, posReviewing, "lane", "handoff")
						handedOff = err == nil
						if !handedOff && tc.finish {
							t.Fatalf("hand-off refused for a card that may finish: %v", err)
						}
					case pathHandOffThenRun:
						// The old receipt of c1 satisfies the hand-off; the card is in REVIEWING when
						// the re-measurement is recorded.
						if err := s.Move(bg, "a", e, posReviewing, "lane", "handoff"); err != nil {
							t.Fatalf("hand-off on the old receipt: %v", err)
						}
						handedOff = true
						addRuns(t, s, e, tc.runs)
					}
					if restart {
						s = reopen(t, s, file)
					}
					err := s.Move(bg, "a", e, done, "lane", "done")
					if tc.finish {
						if err != nil {
							t.Fatalf("DONE refused after c1 was measured again (%s): %v", tc.comment, err)
						}
						return
					}
					if !errors.Is(err, core.ErrGuardMissing) {
						t.Fatalf("DONE although the newest re-measurement of c1 failed (handed off=%v): %v", handedOff, err)
					}
					if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
						t.Error("the card is DONE")
					}
				})
			}
		}
	}
}

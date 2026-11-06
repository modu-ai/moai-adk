package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 22a0afd247 N16: rework_newest_run_test.go reads the newest run of each kind from the
// runs recorded AFTER the move back. A run that failed BEFORE the move back (the very reason the
// holder took the card back) was out of that window, so a later pass of the other kind finished the
// card although the newest green of the landed commit was a failure. The newest result of each
// kind (green, gate) about the commit decides, whenever it was recorded; a kind that was never
// run for the commit is not a failure; and at least one passing run must still follow the move back.

type kindCase struct {
	name   string
	before []rerun // recorded after the landing, before the move back into BUILDING
	after  []rerun // recorded after the move back
	finish bool
}

var newestKindCases = []kindCase{
	// The finding.
	{"green failed before the move back, then a gate pass", []rerun{gf}, []rerun{hp}, false},
	{"gate failed before the move back, then a green pass", []rerun{hf}, []rerun{gp}, false},
	{"green and gate failed before the move back, then a green pass", []rerun{gf, hf}, []rerun{gp}, false},
	// The same kind repaired after the move back.
	{"green failed before, green passes after", []rerun{gf}, []rerun{gp}, true},
	{"gate failed before, gate passes after", []rerun{hf}, []rerun{hp}, true},
	{"green failed before, green and gate pass after", []rerun{gf}, []rerun{gp, hp}, true},
	{"green failed before, a gate pass, then a green pass", []rerun{gf}, []rerun{hp, gp}, true},
	// A kind that was never run again is not a failure: the earlier passes stand.
	{"nothing before, a gate pass after", nil, []rerun{hp}, true},
	{"nothing before, a green pass after", nil, []rerun{gp}, true},
	// A pass before the move back is still not a measurement after it.
	{"a pass before the move back and nothing after", []rerun{gp, hp}, nil, false},
}

func TestMove_DoneReadsTheNewestRunOfEachKindWhateverItsPlaceAroundTheMoveBack(t *testing.T) {
	done := core.Position{Stage: core.StageDone}
	for _, tc := range newestKindCases {
		for _, path := range []string{pathBuilding, pathRunsThenHandOff} {
			for _, restart := range []bool{false, true} {
				name := tc.name + " / " + path
				if restart {
					name += " / after a process restart"
				}
				t.Run(name, func(t *testing.T) {
					s, file, e := landedBackedOutAt(t) // the card is in REVIEWING, c1 landed
					addRuns(t, s, e, tc.before)
					walk(t, s, e, []core.Position{posImplement})
					addRuns(t, s, e, tc.after)
					if path == pathRunsThenHandOff {
						err := s.Move(bg, "a", e, posReviewing, "lane", "handoff")
						if err != nil && tc.finish {
							t.Fatalf("hand-off refused for a card that may finish: %v", err)
						}
					}
					if restart {
						s = reopen(t, s, file)
					}
					err := s.Move(bg, "a", e, done, "lane", "done")
					if tc.finish {
						if err != nil {
							t.Fatalf("DONE refused: %v", err)
						}
						return
					}
					if !errors.Is(err, core.ErrGuardMissing) {
						t.Fatalf("DONE although the newest run of a kind about the landed commit failed, or nothing was measured after the move back: %v", err)
					}
					if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
						t.Error("the card is DONE")
					}
				})
			}
		}
	}
}

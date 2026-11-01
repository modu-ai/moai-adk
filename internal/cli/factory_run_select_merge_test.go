package cli

// factory_run_select_merge_test.go — AC-FCR-022 (e) and the CLI half of
// AC-FCR-021 (d)(e) (REQ-FCR-021, card t1963): the lane-join chooser's `m<n>`
// merge option. A dead-owner run that still holds in-progress cards is offered
// `m<n>` only while exactly one active run has a live owner; choosing it moves
// the cards under that run and retires the dead one, then the join goes on.
// `r<n>` is not touched (its own tests are unchanged).

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func mergeFacts(entries ...homestate.RunFacts) []laneRunRow { return buildLaneRows(entries, nil) }

func liveFact(id string) homestate.RunFacts {
	return homestate.RunFacts{RunOwner: homestate.RunOwner{RunID: id, Classification: homestate.OwnerLive}}
}

func indeterminateFact(id string) homestate.RunFacts {
	return homestate.RunFacts{RunOwner: homestate.RunOwner{RunID: id, Classification: homestate.OwnerIndeterminate}}
}

func deadFact(id string, inProgress int) homestate.RunFacts {
	return homestate.RunFacts{RunOwner: homestate.RunOwner{RunID: id, Classification: homestate.OwnerDead}, InProgress: inProgress}
}

func TestLaneChoiceMergeOffer(t *testing.T) {
	t.Run("eligibility is dead, in progress, and exactly one live run", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			rows []laneRunRow
			want []bool
		}{
			{"one live, one dead with cards", mergeFacts(liveFact("r1"), deadFact("r2", 2)), []bool{false, true}},
			{"two live runs", mergeFacts(liveFact("r1"), liveFact("r2"), deadFact("r3", 2)), []bool{false, false, false}},
			{"no live run", mergeFacts(deadFact("r1", 2), deadFact("r2", 3)), []bool{false, false}},
			{"an indeterminate run does not count as live", mergeFacts(liveFact("r1"), indeterminateFact("r2"), deadFact("r3", 2)), []bool{false, false, true}},
			{"a dead run with no cards is cleanup, not merge", mergeFacts(liveFact("r1"), deadFact("r2", 0)), []bool{false, false}},
			{"a single row never offers it", mergeFacts(deadFact("r1", 2)), []bool{false}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for i, want := range tc.want {
					if got := laneMergeEligible(tc.rows, i); got != want {
						t.Errorf("row %d (%s) merge eligible = %v, want %v", i+1, tc.rows[i].RunID, got, want)
					}
				}
			})
		}
	})

	t.Run("m<n> parses only for an eligible row; r<n> and plain numbers are unchanged", func(t *testing.T) {
		rows := mergeFacts(liveFact("r1"), deadFact("r2", 2))
		merge := runChoice{Kind: choiceMergeThenJoin, RunID: "r2"}
		for _, tc := range []struct {
			name, input string
			want        runChoice
			wantReason  string
		}{
			{"m2", "m2\n", merge, ""},
			{"M2 (upper case)", "M2\n", merge, ""},
			{"m with padding", "  m2 \n", merge, ""},
			{"m1 is the live row", "m1\n2\n", runChoice{Kind: choiceJoin, RunID: "r2"}, fmt.Sprintf(laneMsgIneligibleMergeFmt, 1)},
			{"m9 is out of range", "m9\n1\n", runChoice{Kind: choiceJoin, RunID: "r1"}, laneMsgInvalid},
			{"bare m", "m\n1\n", runChoice{Kind: choiceJoin, RunID: "r1"}, laneMsgInvalid},
			{"r2 still needs zero cards", "r2\n1\n", runChoice{Kind: choiceJoin, RunID: "r1"}, fmt.Sprintf(laneMsgIneligibleFmt, 2)},
			{"plain 2 joins", "2\n", runChoice{Kind: choiceJoin, RunID: "r2"}, ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, out := chooseFrom(t, rows, strings.NewReader(tc.input))
				if got != tc.want {
					t.Errorf("choice = %+v, want %+v", got, tc.want)
				}
				if tc.wantReason != "" && !strings.Contains(out, tc.wantReason) {
					t.Errorf("output lacks the reason %q:\n%s", tc.wantReason, out)
				}
			})
		}
	})

	t.Run("the merge legend appears only when some row offers it", func(t *testing.T) {
		_, out := chooseFrom(t, mergeFacts(liveFact("r1"), deadFact("r2", 2)), strings.NewReader("q\n"))
		if !strings.Contains(out, laneMsgLegendMerge) || !strings.Contains(out, laneMsgLegendMany) {
			t.Errorf("prompt lacks the merge legend or the existing legend:\n%s", out)
		}
		_, out = chooseFrom(t, mergeFacts(liveFact("r1"), liveFact("r2"), deadFact("r3", 2)), strings.NewReader("q\n"))
		if strings.Contains(out, laneMsgLegendMerge) || !strings.Contains(out, laneMsgLegendMany) {
			t.Errorf("prompt shows the merge legend with no eligible row, or lost the existing legend:\n%s", out)
		}
	})

	t.Run("choosing m<n> moves the cards, retires the source, and joins the live run", func(t *testing.T) {
		root := laneGateRoot(t)
		seedLiveRun(t, root, "rl-live")
		selSeedDead(t, root, "rd-dead", false, 2)
		spy := newSpy(root, choose(choiceMergeThenJoin, "rd-dead"))
		notices := installLaneSeams(t, spy)
		restore, err := enterFactoryLaneRun(root, "", "", nil)
		if err != nil {
			t.Fatalf("gate: %v", err)
		}
		defer restore()
		requireEnv(t, "rl-live")
		if got := selRunStatus(t, root, "rd-dead"); got != "retired" {
			t.Errorf("rd-dead status = %q, want retired", got)
		}
		if n := selQueryInt(t, root, `SELECT COUNT(*) FROM cards WHERE run_id='rl-live'`); n != 2 {
			t.Errorf("cards under rl-live = %d, want the two moved cards", n)
		}
		if n := selRetiredEvents(t, root, "rd-dead"); n != 1 {
			t.Errorf("rd-dead has %d run.retired events, want 1", n)
		}
		for _, want := range []string{"rd-dead", "rl-live", "card-rd-dead-0", "card-rd-dead-1"} {
			if !strings.Contains(notices.String(), want) {
				t.Errorf("notices %q do not name %q", notices.String(), want)
			}
		}
		if len(spy.entries) != 1 {
			t.Fatalf("spy entered %d times, want 1", len(spy.entries))
		}
		offered := false
		for i, row := range spy.entries[0].Rows {
			if row.RunID == "rd-dead" {
				offered = laneMergeEligible(spy.entries[0].Rows, i)
			}
		}
		if !offered {
			t.Errorf("the dead row was not offered the merge: rows=%+v", spy.entries[0].Rows)
		}
	})

	t.Run("a chooser cannot merge a source that is live, or merge beside two live runs", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			seed   func(t *testing.T, root string)
			source string
		}{
			{"live source", func(t *testing.T, root string) {
				seedLiveRun(t, root, "rl-live")
				seedLiveRun(t, root, "rm-live")
			}, "rl-live"},
			{"dead source beside two live runs", func(t *testing.T, root string) {
				seedLiveRun(t, root, "rl-live")
				seedLiveRun(t, root, "rm-live")
				selSeedDead(t, root, "rd-dead", false, 2)
			}, "rd-dead"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				root := laneGateRoot(t)
				tc.seed(t, root)
				// A scripted chooser bypasses the parser's eligibility check.
				spy := newSpy(root, choose(choiceMergeThenJoin, tc.source), choose(choiceCancel, ""))
				notices := installLaneSeams(t, spy)
				before := selSnapshot(t, root)
				_, err := enterFactoryLaneRun(root, "", "", nil)
				if err == nil || !strings.Contains(err.Error(), joinCancelledSentinel) {
					t.Fatalf("err = %v, want %s", err, joinCancelledSentinel)
				}
				if !strings.Contains(notices.String(), fmt.Sprintf("run %s 합치기에 실패했습니다", tc.source)) {
					t.Errorf("notices %q lack the printed refusal", notices.String())
				}
				if after := selSnapshot(t, root); fmt.Sprint(after) != fmt.Sprint(before) {
					t.Errorf("the refused merge changed runs/events:\n before %v\n after  %v", before, after)
				}
				if len(spy.entries) != 2 {
					t.Errorf("spy entered %d times, want 2 (presented again after the refusal)", len(spy.entries))
				}
			})
		}
	})
}

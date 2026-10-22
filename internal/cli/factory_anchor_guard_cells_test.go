package cli

// factory_anchor_guard_cells_test.go — the SPEC-LAUNCHER-ARGV-001 M2 cells. The
// harness (fixture, leader driver, placement indicators) is in
// factory_anchor_guard_test.go.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// refusalNames asserts a refusal error names the detected tree and branch and
// offers the documented override (REQ-LA-004).
func refusalNames(t *testing.T, label string, obs leaderObs, tree, branch string) {
	t.Helper()
	if obs.err == nil {
		return
	}
	msg := obs.err.Error()
	for _, want := range []string{tree, branch, allowCardAnchorFlag} {
		if !strings.Contains(msg, want) {
			t.Errorf("%s: refusal %q does not name %q", label, msg, want)
		}
	}
}

// TestLauncherAnchorFinalArgv holds the guard-input cells of the iter8 failure
// families (a) consumed-option shadowing, (b) multi-option last-wins, (c)
// bundled short options, and the §D.4 view-exec equality table (AC-LA-001..003,
// AC-LA-005). Each family cell runs on both arms through the real leader entry.
func TestLauncherAnchorFinalArgv(t *testing.T) {
	pinSnapshotModel(t)

	// AC-LA-001, family (a): `moai cc -f -w -b -- <card-tree>`. The launcher
	// consumes -b and the separator, so claude receives `--worktree <card-tree>`;
	// the guard reads that final argv.
	t.Run("consumed_option_shadowing", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm, func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "-b", "--", "WT-card"})
				assertRefusedBeforeMutation(t, arm, obs)
				refusalNames(t, arm, obs, fx.card, "WT-card")
			})
		}
	})

	// AC-LA-002, family (b): the final argv carries two worktree options and the
	// backend takes the LAST. The control proves the verdict follows the last
	// option, not "any option names a card tree" and not the first.
	t.Run("multi_option_last_wins", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm, func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "plain-tree", "--worktree", "--branch", "main", "WT-card"})
				assertRefusedBeforeMutation(t, arm, obs)
				refusalNames(t, arm, obs, fx.card, "WT-card")

				fx = newLeaderFixture(t)
				control := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card", "--worktree", "plain-tree"})
				if control.err != nil || !control.launched {
					t.Errorf("%s control: a card tree FIRST and a plain tree LAST must launch (the backend takes the last), err=%v launched=%v", arm, control.err, control.launched)
				}
			})
		}
	})

	// AC-LA-003, family (c): `-pw<tree>` in the claude segment is a boolean -p
	// followed by -w with the attached value.
	t.Run("bundled_short_cluster", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm, func(t *testing.T) {
				for _, args := range [][]string{
					{"-f", "--", "-pwWT-card", "PROMPT"},
					{"-f", "--", "-pw", "WT-card", "PROMPT"},
				} {
					fx := newLeaderFixture(t)
					obs := driveLeaderLaunch(t, fx, arm, args)
					assertRefusedBeforeMutation(t, arm+" "+strings.Join(args, " "), obs)
					refusalNames(t, arm, obs, fx.card, "WT-card")
				}
				fx := newLeaderFixture(t)
				control := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-pwplain", "PROMPT"})
				if control.err != nil || !control.launched {
					t.Errorf("%s control: a cluster naming a plain tree must launch, err=%v launched=%v", arm, control.err, control.launched)
				}
			})
		}
	})

	// Controls for §D.4 rows 15-17: with a SINGLE `--` the launcher eats it as its
	// own separator, so the backend sees `-n -w WT-card`: `-n` takes `-w` as its
	// value (a required-value option takes a flag-shaped token), and `WT-card` is
	// a positional — no worktree selection, the leader launch is allowed. Without
	// them rows 15-17 would also pass a resolver that simply stopped honouring `--`.
	t.Run("required_value_separator_controls", func(t *testing.T) {
		for _, arm := range leaderArms {
			for _, args := range [][]string{
				{"-n", "--", "-w", "WT-card"},
				{"-cn", "--", "-w", "WT-card"},
				{"--append-system-prompt", "--", "--worktree", "WT-card"},
			} {
				t.Run(arm+"/"+strings.Join(args, "_"), func(t *testing.T) {
					view := resolveLaunchView(args)
					if view.Selection.Present {
						t.Errorf("view selection = %+v, want none (the value option took the worktree flag as its value)", view.Selection)
					}
					fx := newLeaderFixture(t)
					obs := driveLeaderLaunch(t, fx, arm, append([]string{"-f"}, args...))
					if obs.err != nil || !obs.launched {
						t.Errorf("leader launch of a single-separator form: err=%v launched=%v, want allowed", obs.err, obs.launched)
					}
				})
			}
		}
	})

	// AC-LA-005: for every §D.4 spelling the view's FinalArgv equals, token by
	// token, the argv the exec boundary receives (minus the flags the launcher
	// injects from profile state and, for a leader, the tail flags it appends
	// after the guard has judged), the view carries the expected backend
	// selection, the selection resolves to the expected tree and effective
	// branch, and the verdict is the expected one. Equality alone would pass a
	// resolver that misparses every value identically on both sides, so the
	// selection, tree, branch and verdict are asserted too.
	t.Run("view_exec_byte_equality", func(t *testing.T) {
		elsewhere := filepath.Join(t.TempDir(), "elsewhere")
		for _, arm := range leaderArms {
			for _, row := range viewEqualityRows(elsewhere) {
				t.Run(arm+"/"+row.name, func(t *testing.T) {
					view := resolveLaunchView(row.args)

					// Non-factory launch: the equality is exact.
					obs := launchWithStageSeams(t, arm, row.args, "bare")
					if obs.Err != "" {
						t.Fatalf("non-factory launch refused: %q", obs.Err)
					}
					if !reflect.DeepEqual(view.FinalArgv, obs.ExecTail) {
						t.Errorf("view FinalArgv = %q, exec boundary received %q", view.FinalArgv, obs.ExecTail)
					}
					sel := view.Selection
					if !sel.Present || sel.Spelling != row.spelling || sel.Value != row.value || sel.HasValue != row.hasValue {
						t.Errorf("selection = %+v, want spelling %q value %q hasValue %v", sel, row.spelling, row.value, row.hasValue)
					}
					if !row.leader {
						return
					}

					// Leader launch in the fixture: the same spelling, judged.
					fx := newLeaderFixture(t)
					lead := driveLeaderLaunch(t, fx, arm, append([]string{"-f"}, row.args...))
					wantBranch := row.branch
					if wantBranch == "<base>" {
						wantBranch = fx.base
					}
					gotTree := ""
					gotBranch := anchorBranchOfTree(fx.root)
					if view.Selection.HasValue {
						target := resolveWorktreeLaunchTarget(fx.root, view.Selection.Value)
						gotTree = relToStageRoot(fx.root, target)
						gotBranch = anchorBranchOfTree(target)
					}
					if gotTree != row.tree || gotBranch != wantBranch {
						t.Errorf("resolved tree %q branch %q, want tree %q branch %q", gotTree, gotBranch, row.tree, wantBranch)
					}
					if row.refuse {
						assertRefusedBeforeMutation(t, row.name, lead)
						refusalNames(t, row.name, lead, filepath.Join(fx.root, filepath.FromSlash(row.tree)), wantBranch)
						return
					}
					if lead.err != nil || !lead.launched {
						t.Fatalf("leader launch of an allowed spelling: err=%v launched=%v", lead.err, lead.launched)
					}
					if got := stripLauncherAppended(stageExecTail(lead.execArgs)); !reflect.DeepEqual(got, view.FinalArgv) {
						t.Errorf("leader exec argv (appended flags stripped) = %q, view FinalArgv = %q", got, view.FinalArgv)
					}
				})
			}
		}
	})
}

func TestLauncherAnchorBranchAnnotation(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		// (i) an existing tree is trusted for its real branch: a --branch value
		// that is not a card branch does not launder a card tree, and a card
		// --branch value does not condemn a plain tree.
		t.Run(arm+"/existing_tree_real_branch", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card", "--branch", "main"})
			assertRefusedBeforeMutation(t, "card tree with a plain --branch", obs)
			refusalNames(t, arm, obs, fx.card, "WT-card")

			fx = newLeaderFixture(t)
			plain := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "plain", "--branch", "WT-elsewhere"})
			if plain.err != nil || !plain.launched {
				t.Errorf("a plain tree with a card --branch value is reused unchanged and must launch: err=%v launched=%v", plain.err, plain.launched)
			}
		})
		// (ii) a tree that does not exist yet takes the annotation's branch, and
		// the refusal precedes the materialization.
		t.Run(arm+"/created_tree_annotation", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "fresh", "--branch", "WT-fresh"})
			assertRefusedBeforeMutation(t, "created tree at a card branch", obs)
			refusalNames(t, arm, obs, filepath.Join(fx.root, ".claude", "worktrees", "fresh"), "WT-fresh")

			fx = newLeaderFixture(t)
			control := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "fresh", "--branch", "release"})
			if control.err != nil || !control.launched {
				t.Fatalf("a created tree at a plain branch must launch: err=%v launched=%v", control.err, control.launched)
			}
			if want := []string{"fresh@release"}; !reflect.DeepEqual(control.materialized, want) {
				t.Errorf("materializations = %q, want %q", control.materialized, want)
			}
		})
	}
}

// TestLauncherAnchorRunCountInstrument is family (d): the counter a refusal
// test reads must be able to move. Asserting "the refused launch left no run
// record" with a counter that always reads zero proves nothing — the iter8
// counter looked for a layout-specific file the current storage never writes.
// The positive control launches an allowed leader and watches the counter move;
// the absence assertion is only meaningful next to it.
func TestLauncherAnchorRunCountInstrument(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		t.Run(arm+"/positive_control_counter_moves", func(t *testing.T) {
			fx := newLeaderFixture(t)
			before := launcherRunCount(t, fx.root)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "plain"})
			if obs.err != nil || !obs.launched {
				t.Fatalf("allowed leader launch: err=%v launched=%v", obs.err, obs.launched)
			}
			if after := launcherRunCount(t, fx.root); after != before+1 {
				t.Errorf("run records %d -> %d after a recorded leader start, want +1: the counter cannot see a record", before, after)
			}
			// The legacy single-layout probe (is the queue's backlog.json file
			// there?) reads this project as record-less even though a run was
			// recorded — the divergence that made the iter8 counter vacuous.
			if _, err := os.Stat(factory.BacklogPathForRoot(fx.root)); err == nil {
				t.Errorf("a legacy backlog.json exists: the fixture no longer separates the canonical counter from the legacy probe")
			}
		})
		t.Run(arm+"/refusal_leaves_the_counter_unmoved", func(t *testing.T) {
			fx := newLeaderFixture(t)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card"})
			if obs.err == nil {
				t.Fatal("anchored leader launch was not refused")
			}
			if obs.runsAfter != obs.runsBefore {
				t.Errorf("run records %d -> %d, want unchanged", obs.runsBefore, obs.runsAfter)
			}
		})
	}
}

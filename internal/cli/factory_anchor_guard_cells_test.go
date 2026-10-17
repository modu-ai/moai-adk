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

	"github.com/modu-ai/moai-adk/internal/config"
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

// TestLauncherAnchorGuardSemantics holds the preserved non-argv semantics of
// the guard (AC-LA-006..008, 010, 011).
func TestLauncherAnchorGuardSemantics(t *testing.T) {
	pinSnapshotModel(t)

	// AC-LA-006: every session mutation waits behind the guard.
	t.Run("refusal_before_mutation", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm+"/named_tree", func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card"})
				assertRefusedBeforeMutation(t, arm, obs)
				refusalNames(t, arm, obs, fx.card, "WT-card")
			})
			t.Run(arm+"/launch_directory", func(t *testing.T) {
				fx := newLeaderFixture(t)
				t.Chdir(fx.card)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f"})
				assertRefusedBeforeMutation(t, arm, obs)
				refusalNames(t, arm, obs, fx.card, "WT-card")
			})
		}
	})

	// AC-LA-007: a lane join, or a launch with no factory entry, anchors in a
	// card tree on purpose — the guard does not fire.
	t.Run("lane_and_nonfactory_exempt", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm+"/nonfactory_named_tree", func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-w", "WT-card"})
				if obs.err != nil || !obs.launched || obs.probed {
					t.Errorf("a launch with no factory entry: err=%v launched=%v guard ran=%v, want a launch the guard never saw", obs.err, obs.launched, obs.probed)
				}
			})
			t.Run(arm+"/nonfactory_from_card_tree", func(t *testing.T) {
				fx := newLeaderFixture(t)
				t.Chdir(fx.card)
				obs := driveLeaderLaunch(t, fx, arm, nil)
				if obs.err != nil || !obs.launched || obs.probed {
					t.Errorf("a plain launch from a card tree: err=%v launched=%v guard ran=%v", obs.err, obs.launched, obs.probed)
				}
			})
		}
		// A lane join into the card tree: the run fixture of the lane-join tests.
		root := codexLedRun(t, "rc", "claude")
		anchorGit(t, root, "worktree", "add", "-q", "-b", "WT-card", filepath.Join(root, ".claude", "worktrees", "WT-card"))
		t.Chdir(root)
		fx := &leaderFixture{root: root, base: anchorGit(t, root, "branch", "--show-current")}
		obs := driveLeaderLaunch(t, fx, "cc", []string{"-l", "-w", "WT-card"})
		if obs.err != nil || !obs.launched || obs.probed {
			t.Errorf("a lane join into a card tree: err=%v launched=%v guard ran=%v, want a launch the guard never saw", obs.err, obs.launched, obs.probed)
		}
	})

	// AC-LA-008: the documented override proceeds with a warning and an audit
	// row; a bypass that cannot be recorded is refused; the flag counts only
	// where it stands.
	t.Run("override_audit_row", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm+"/recorded", func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", allowCardAnchorFlag, "-w", "WT-card"})
				if obs.err != nil || !obs.launched {
					t.Fatalf("override launch: err=%v launched=%v", obs.err, obs.launched)
				}
				for _, a := range obs.execArgs {
					if a == allowCardAnchorFlag {
						t.Errorf("the override flag reached the backend argv: %q", obs.execArgs)
					}
				}
				if !strings.Contains(obs.stderr, "override") {
					t.Errorf("no override warning on stderr: %q", obs.stderr)
				}
				raw, err := os.ReadFile(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath))
				if err != nil {
					t.Fatalf("read the audit log: %v", err)
				}
				backend := map[string]string{"cc": "claude", "glm": "glm"}[arm]
				for _, want := range []string{"backend=" + backend, fx.card, "WT-card", allowCardAnchorFlag} {
					if !strings.Contains(string(raw), want) {
						t.Errorf("audit row %q lacks %q", raw, want)
					}
				}
			})
			t.Run(arm+"/unrecordable_is_refused", func(t *testing.T) {
				fx := newLeaderFixture(t)
				if err := os.MkdirAll(filepath.Join(fx.root, config.DefaultLeaderAnchorAuditRelPath), 0o755); err != nil {
					t.Fatal(err)
				}
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", allowCardAnchorFlag, "-w", "WT-card"})
				if obs.err == nil || obs.launched {
					t.Fatalf("a bypass that cannot be recorded proceeded: err=%v launched=%v", obs.err, obs.launched)
				}
				if !strings.Contains(obs.err.Error(), "could not be recorded") {
					t.Errorf("refusal %q does not say the bypass could not be recorded", obs.err)
				}
				if obs.runsAfter != obs.runsBefore {
					t.Errorf("run records %d -> %d, want unchanged", obs.runsBefore, obs.runsAfter)
				}
			})
			t.Run(arm+"/flag_as_an_option_value_is_not_an_override", func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--model", allowCardAnchorFlag, "-w", "WT-card"})
				if obs.err == nil || obs.launched {
					t.Errorf("the flag standing as --model's value was read as an override: err=%v launched=%v", obs.err, obs.launched)
				}
				fx = newLeaderFixture(t)
				plain := driveLeaderLaunch(t, fx, arm, []string{"-f", "--model", allowCardAnchorFlag, "-w", "plain"})
				if plain.err != nil || !plain.launched {
					t.Fatalf("plain-tree launch: err=%v launched=%v", plain.err, plain.launched)
				}
				found := false
				for _, a := range plain.execArgs {
					found = found || a == allowCardAnchorFlag
				}
				if !found {
					t.Errorf("the option value %s was stripped from the backend argv: %q", allowCardAnchorFlag, plain.execArgs)
				}
			})
		}
	})

	// AC-LA-010: refusal rests on positive evidence of a card prefix only.
	t.Run("fail_open_on_unverifiable", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm+"/absent_tree", func(t *testing.T) {
				fx := newLeaderFixture(t)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "no-such-tree"})
				if obs.err != nil || !obs.launched {
					t.Errorf("a tree that does not exist and has no --branch annotation must launch: err=%v launched=%v", obs.err, obs.launched)
				}
			})
			t.Run(arm+"/branch_unreadable", func(t *testing.T) {
				fx := newLeaderFixture(t)
				detached := filepath.Join(fx.root, ".claude", "worktrees", "detached-card")
				anchorGit(t, fx.root, "worktree", "add", "-q", "--detach", detached)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "detached-card"})
				if obs.err != nil || !obs.launched {
					t.Errorf("a tree whose branch cannot be read must launch: err=%v launched=%v", obs.err, obs.launched)
				}
			})
		}
	})

	// AC-LA-011: the discriminator is the branch prefix only; stale env labels
	// that hint at a card anchor decide nothing.
	t.Run("discriminator_ignores_env", func(t *testing.T) {
		for _, arm := range leaderArms {
			t.Run(arm, func(t *testing.T) {
				fx := newLeaderFixture(t)
				t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
				t.Setenv(config.EnvFactoryRunID, "stale-card-run")
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f"})
				if obs.err != nil || !obs.launched {
					t.Errorf("a primary-checkout leader launch with stale card labels in the env must launch: err=%v launched=%v", obs.err, obs.launched)
				}
				fx = newLeaderFixture(t)
				t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
				t.Setenv(config.EnvFactoryRunID, "stale-card-run")
				contrast := driveLeaderLaunch(t, fx, arm, []string{"-f", "-w", "WT-card"})
				if contrast.err == nil || contrast.launched {
					t.Errorf("the same env with a WT- anchor must be refused: err=%v launched=%v", contrast.err, contrast.launched)
				}
			})
		}
	})
}

// TestLauncherAnchorBranchAnnotation holds AC-LA-012: the effective branch of
// a tree the --branch path touches.
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

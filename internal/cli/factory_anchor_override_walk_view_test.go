package cli

// factory_anchor_override_walk_view_test.go — SPEC-LAUNCHER-ARGV-001 REQ-LA-005 /
// REQ-LA-001, the white-box half of the override walk tests (it names the walk's
// entry point, so it arrives with the implementation).
//
// The leader entry decides the override with resolveOwning — the launch-view walk
// run with the flag in play — and then re-resolves the argv with the flag
// removed for the launch readers. The two must be the same interpretation: the
// view the decision implies (the argv the guard would judge) is the view the
// launch will act on, over argv that carry the launcher's -w and --branch
// tokens as well.

import (
	"testing"
)

func TestLauncherAnchorOverrideWalkViewAgreement(t *testing.T) {
	pinSnapshotModel(t)

	alphabet := []string{
		"-w", "WT-card", "--worktree=", "--branch", "main", allowCardAnchorFlag,
		"--name", "-cn", "-b", "--model", "opus", argSeparator,
	}
	alphabet = append(alphabet, flagBearingWorktreeForms()...)
	total, bad := 0, 0
	eachArgv(alphabet, 4, func(g []string) {
		total++
		decided, ordinals := resolveOwning(g, allowCardAnchorFlag)
		rest, allowed := consumeAllowCardAnchor(g)
		launched := resolveLaunchView(rest)

		var problems []string
		if allowed != (len(ordinals) > 0) {
			problems = append(problems, "decision and ordinals disagree")
		}
		if !sameArgv(decided.FinalArgv, launched.FinalArgv) {
			problems = append(problems, "the argv the decision implies is not the argv the launch gets")
		}
		if decided.Selection != launched.Selection {
			problems = append(problems, "selection differs")
		}
		if decided.Consumed.ChainTarget != launched.Consumed.ChainTarget {
			problems = append(problems, "chain target differs")
		}
		if decided.Branch.Present != launched.Branch.Present || decided.Branch.Branch != launched.Branch.Branch ||
			(decided.Branch.Err == nil) != (launched.Branch.Err == nil) {
			problems = append(problems, "branch annotation differs")
		}
		if len(problems) == 0 {
			return
		}
		bad++
		if bad <= 8 {
			t.Errorf("argv %q: %v\n  decided: final=%q selection=%+v\n  launch:  final=%q selection=%+v (rest=%q)",
				g, problems, decided.FinalArgv, decided.Selection, launched.FinalArgv, launched.Selection, rest)
		}
	})
	if want := argvSpace(len(alphabet), 4); total != want {
		t.Errorf("enumerated %d argv, want the full space %d", total, want)
	}
	if bad > 0 {
		t.Errorf("%d of %d argv: the override decision and the launch disagree", bad, total)
	}
}

package cli

// factory_anchor_backend_worktree_name_test.go — card t1958 (issue #1847): two
// worktree-name spellings on which the leader anchor guard and the backend
// disagreed about the tree the launch enters.
//
//   - `-w -`          the backend (claude 2.1.303) takes a lone `-` as the optional
//                     VALUE of -w/--worktree; the launch view used to read it as
//                     "no value", so the guard saw no selection at all.
//   - `-w feature/foo` the backend maps a slash in the name to `+` for the native
//                     directory (.claude/worktrees/feature+foo); the guard looked
//                     at .claude/worktrees/feature/foo, which does not exist.
//
// The cells drive the real runCC / runGLM leader entry through the harness of
// factory_anchor_guard_test.go; only the engine launch, the claude binary and
// the --branch materializer are substituted.

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// addBackendNamedTree links a tree at .claude/worktrees/<dir> on branch.
func addBackendNamedTree(t *testing.T, fx *leaderFixture, dir, branch string) string {
	t.Helper()
	path := filepath.Join(fx.root, ".claude", "worktrees", dir)
	anchorGit(t, fx.root, "worktree", "add", "-q", "-b", branch, path, fx.base)
	return path
}

func TestLauncherAnchorBackendWorktreeNames(t *testing.T) {
	pinSnapshotModel(t)
	cases := []struct {
		name   string
		args   []string
		dir    string // directory under .claude/worktrees the backend enters ("" = the fixture's WT-card)
		branch string
		refuse bool
	}{
		{"lone dash, short, separator", []string{"-f", "--", "-w", "-"}, "-", "WT-dash", true},
		{"lone dash, short, no separator", []string{"-f", "-w", "-"}, "-", "WT-dash", true},
		{"lone dash, long", []string{"-f", "--worktree", "-"}, "-", "WT-dash", true},
		{"slash, short, separator", []string{"-f", "--", "-w", "feature/foo"}, "feature+foo", "WT-slash", true},
		{"slash, long equals", []string{"-f", "--worktree=feature/foo"}, "feature+foo", "WT-slash", true},
		{"slash, nested", []string{"-f", "-w", "a/b/c"}, "a+b+c", "WT-nested", true},
		{"lone dash, plain branch", []string{"-f", "--", "-w", "-"}, "-", "plain-dash", false},
		{"slash, plain branch", []string{"-f", "--", "-w", "feature/foo"}, "feature+foo", "plain-slash", false},
		{"short name, card tree (existing control)", []string{"-f", "-w", "WT-card"}, "", "", true},
	}
	for _, arm := range leaderArms {
		for _, tc := range cases {
			t.Run(arm+"/"+tc.name, func(t *testing.T) {
				fx := newLeaderFixture(t)
				tree, branch := fx.card, "WT-card"
				if tc.dir != "" {
					tree, branch = addBackendNamedTree(t, fx, tc.dir, tc.branch), tc.branch
				}
				obs := driveLeaderLaunch(t, fx, arm, tc.args)
				if tc.refuse {
					assertRefusedBeforeMutation(t, tc.name, obs)
					refusalNames(t, tc.name, obs, tree, branch)
					return
				}
				if obs.err != nil || !obs.launched {
					t.Errorf("%s: a tree on a plain branch must launch: err=%v launched=%v", tc.name, obs.err, obs.launched)
				}
			})
		}
	}
}

// TestLauncherAnchorBackendWorktreeNameSelection pins the unified meaning of a
// lone `-` in the launch view: the backend model reads it as the worktree
// option's value, and only there (a lone `-` standing alone is still a
// positional, and the consumer-scoped stage reads are untouched).
func TestLauncherAnchorBackendWorktreeNameSelection(t *testing.T) {
	pinSnapshotModel(t)
	cases := []struct {
		name string
		argv []string
		want worktreeSelection
	}{
		{"short", []string{"-w", "-"}, worktreeSelection{Present: true, Spelling: "-w", Value: "-", HasValue: true}},
		{"long", []string{"--worktree", "-"}, worktreeSelection{Present: true, Spelling: "--worktree", Value: "-", HasValue: true}},
		{"cluster", []string{"-pw", "-"}, worktreeSelection{Present: true, Spelling: "-<cluster>w<value>", Value: "-", HasValue: true}},
		{"attached dash", []string{"-w-"}, worktreeSelection{Present: true, Spelling: "-w<value>", Value: "-", HasValue: true}},
		{"a flag-shaped token is still not a value", []string{"-w", "-x"}, worktreeSelection{Present: true, Spelling: "-w"}},
		{"the separator is still not a value", []string{"-w", "--", "-"}, worktreeSelection{Present: true, Spelling: "-w"}},
		{"a lone dash before the option stays positional", []string{"-", "-w", "x"}, worktreeSelection{Present: true, Spelling: "-w", Value: "x", HasValue: true, Index: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := selectBackendWorktree(tc.argv); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("selectBackendWorktree(%q) = %+v, want %+v", tc.argv, got, tc.want)
			}
		})
	}

	// The consumer-scoped reads keep reading `-w -` as a bare option (SPEC-LAUNCHER-ARGV-001 §2).
	view := resolveLaunchView([]string{"-w", "-"})
	if view.StageA.Has || view.StageB.Has || view.Consumed.ChainTarget != "" {
		t.Errorf("stage reads of `-w -` = A %+v B %+v chain %q, want no value on all three", view.StageA, view.StageB, view.Consumed.ChainTarget)
	}
	if want := []string{"--worktree", "-"}; !reflect.DeepEqual(view.FinalArgv, want) {
		t.Errorf("FinalArgv = %q, want %q", view.FinalArgv, want)
	}
}

// TestLauncherAnchorBackendWorktreeNameTarget pins the name-to-directory
// mapping the guard inspects: the slash of a short name becomes `+`, an
// absolute value names its tree directly.
func TestLauncherAnchorBackendWorktreeNameTarget(t *testing.T) {
	pinSnapshotModel(t)
	fx := newLeaderFixture(t)
	for value, wantDir := range map[string]string{
		"plain":       "plain",
		"WT-card":     "WT-card",
		"-":           "-",
		"feature/foo": "feature+foo",
		"a/b/c":       "a+b+c",
		"feature+foo": "feature+foo",
	} {
		got := filepath.ToSlash(resolveWorktreeLaunchTarget(fx.root, value))
		if want := "/.claude/worktrees/" + wantDir; !strings.HasSuffix(got, want) {
			t.Errorf("resolveWorktreeLaunchTarget(%q) = %q, want a path ending in %q", value, got, want)
		}
	}
	abs := filepath.Join(t.TempDir(), "a", "b")
	if got := resolveWorktreeLaunchTarget(fx.root, abs); got != abs {
		t.Errorf("resolveWorktreeLaunchTarget(%q) = %q, want the absolute path unchanged", abs, got)
	}
}

// TestLauncherAnchorBackendWorktreeNameNonLeader proves the non-leader launch
// of `-w -` is untouched: the guard is a leader-only reader of the selection,
// so the launch proceeds even where the tree the backend would enter is a card
// tree.
func TestLauncherAnchorBackendWorktreeNameNonLeader(t *testing.T) {
	pinSnapshotModel(t)
	for _, arm := range leaderArms {
		t.Run(arm, func(t *testing.T) {
			fx := newLeaderFixture(t)
			addBackendNamedTree(t, fx, "-", "WT-dash")
			obs := driveLeaderLaunch(t, fx, arm, []string{"--", "-w", "-"})
			if obs.err != nil || !obs.launched {
				t.Errorf("non-leader `-w -` launch: err=%v launched=%v, want allowed", obs.err, obs.launched)
			}
		})
	}
}
